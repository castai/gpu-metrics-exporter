package castai_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/castai/gpu-metrics-exporter/internal/castai"
	"github.com/castai/gpu-metrics-exporter/pb"
)

func Test_UploadBatch(t *testing.T) {
	log := logrus.New()
	restyClient := resty.New()
	client := castai.NewClient(castai.Config{
		URL:       "http://localhost",
		APIKey:    "my-fake-token",
		ClusterID: "cluster-id-1",
	}, log, restyClient, "test")

	httpmock.ActivateNonDefault(restyClient.GetClient())
	t.Run("calls gpu-metrics endpoint with proper headers", func(t *testing.T) {
		r := require.New(t)

		httpmock.RegisterResponder(
			"POST",
			"http://localhost/v1/kubernetes/clusters/cluster-id-1/gpu-metrics",
			func(req *http.Request) (*http.Response, error) {
				contentType := req.Header.Get("Content-type")
				r.Equal("application/protobuf", contentType)

				contentEncoding := req.Header.Get("Content-encoding")
				r.Equal("gzip", contentEncoding)

				apiKey := req.Header.Get("X-API-Key")
				r.Equal("my-fake-token", apiKey)

				userAgent := req.Header.Get("User-Agent")
				r.Equal("castai-gpu-metrics-exporter/test", userAgent)

				return &http.Response{StatusCode: 200}, nil
			},
		)

		_ = client.UploadBatch(context.Background(), &pb.MetricsBatch{})
	})
}
