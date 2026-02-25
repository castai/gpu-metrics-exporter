package castai

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/castai/gpu-metrics-exporter/pb"
	"github.com/castai/logging"
)

const (
	tokenHeader = "X-API-Key" // #nosec G101
	retryCount  = 5
)

var (
	backoff = wait.Backoff{
		Steps:    retryCount,
		Duration: 50 * time.Millisecond,
		Factor:   8,
		Jitter:   0.15,
	}

	contentTypeHeader = http.CanonicalHeaderKey("Content-Type")
	contentType       = "application/protobuf"

	contentEncodingHeader = http.CanonicalHeaderKey("Content-Encoding")
	contentEncoding       = "gzip"

	userAgentHeader = http.CanonicalHeaderKey("User-Agent")
	userAgent       = "castai-gpu-metrics-exporter/"
)

type Config struct {
	URL       string
	APIKey    string // nolint:gosec // G117: false positive
	ClusterID string
}

type Client interface {
	UploadBatch(ctx context.Context, batch *pb.MetricsBatch) error
}

type client struct {
	restyClient *resty.Client
	cfg         Config
	log         *logging.Logger
}

func NewClient(cfg Config, log *logging.Logger, restyClient *resty.Client, version string) Client {
	restyClient.BaseURL = cfg.URL
	restyClient.SetHeaders(map[string]string{
		tokenHeader:           cfg.APIKey,
		contentTypeHeader:     contentType,
		contentEncodingHeader: contentEncoding,
		userAgentHeader:       fmt.Sprintf("%s%s", userAgent, version),
	})

	return &client{
		restyClient: restyClient,
		cfg:         cfg,
		log:         log,
	}
}

func (c client) UploadBatch(ctx context.Context, batch *pb.MetricsBatch) error {
	buffer, err := c.toBuffer(batch)
	if err != nil {
		return err
	}

	return wait.ExponentialBackoffWithContext(ctx, backoff, func(ctx context.Context) (done bool, err error) {
		resp, err := c.restyClient.R().
			SetContext(ctx).
			SetBody(buffer).
			Post(fmt.Sprintf("/v1/kubernetes/clusters/%s/gpu-metrics", c.cfg.ClusterID))

		if err != nil {
			c.log.WithField("error", err.Error()).Error("error making http request")
			return false, nil
		}

		statusCode := resp.StatusCode()
		switch {
		case statusCode >= 200 && statusCode < 300:
			return true, nil
		case statusCode >= 400 && statusCode < 500:
			return true, fmt.Errorf("status code: %d, status: %s", statusCode, resp.Status())
		default:
			c.log.Errorf("server error or unexpected status code: %d, status: %s", statusCode, resp.Status())
			return false, nil
		}
	})
}

func (c client) toBuffer(batch *pb.MetricsBatch) (*bytes.Buffer, error) {
	payload := new(bytes.Buffer)

	protoBytes, err := proto.Marshal(batch)
	if err != nil {
		return nil, fmt.Errorf("error marshaling batch %w", err)
	}

	writer := gzip.NewWriter(payload)
	if _, err := writer.Write(protoBytes); err != nil {
		return nil, fmt.Errorf("error compressing payload %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("error closing gzip writer %w", err)
	}

	return payload, nil
}
