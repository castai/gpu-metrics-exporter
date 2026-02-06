package exporter

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/castai/gpu-metrics-exporter/internal/castai"
	"github.com/castai/logging"
	"github.com/castai/metrics"
)

type Exporter interface {
	Start(ctx context.Context) error
	Enable()
	Disable()
	Enabled() bool
}

type Config struct {
	ExportInterval   time.Duration
	DCGMExporterPort int
	DCGMExporterPath string
	DCGMExporterHost string
	Selector         string
	Enabled          bool
	NodeName         string
}

type exporter struct {
	cfg          Config
	dynamic      dynamic.Interface
	log          *logging.Logger
	scraper      Scraper
	mapper       MetricMapper
	enabled      *atomic.Bool
	client       castai.Client
	metricClient metrics.MetricClient
	metricWriter metrics.Metric[GPUMetric]
}

func NewExporter(
	cfg Config,
	dynClient dynamic.Interface,
	log *logging.Logger,
	scraper Scraper,
	mapper MetricMapper,
	castaiClient castai.Client,
	metricClient metrics.MetricClient,
) Exporter {
	enabled := atomic.Bool{}
	enabled.Store(cfg.Enabled)

	var m metrics.Metric[GPUMetric]
	var err error
	if metricClient != nil {
		m, err = metrics.NewMetric[GPUMetric](
			metricClient,
			metrics.WithCollectionName[GPUMetric]("gpu_metrics"),
			metrics.WithSkipTimestamp[GPUMetric](),
		)

		if err != nil {
			log.WithField("error", err.Error()).Warn("failed to create metric")
		}
	}

	return &exporter{
		cfg:          cfg,
		dynamic:      dynClient,
		log:          log,
		scraper:      scraper,
		mapper:       mapper,
		enabled:      &enabled,
		client:       castaiClient,
		metricClient: metricClient,
		metricWriter: m,
	}
}

func (e *exporter) Start(ctx context.Context) error {
	exportTicker := time.NewTicker(e.cfg.ExportInterval)
	defer exportTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-exportTicker.C:
			if !e.enabled.Load() {
				continue
			}
			if err := e.export(ctx); err != nil {
				e.log.WithField("error", err.Error()).Errorf("error while exporting metrics")
			}
		}
	}
}

func (e *exporter) Enable() {
	e.enabled.Store(true)
}

func (e *exporter) Disable() {
	e.enabled.Store(false)
}

func (e *exporter) Enabled() bool {
	return e.enabled.Load()
}

var podGVR = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

func (e *exporter) getDCGMUrls(ctx context.Context) ([]string, error) {
	if e.cfg.DCGMExporterHost != "" {
		// we are scraping a single host, no need to check for other pods
		return []string{
			fmt.Sprintf("http://%s:%d%s", e.cfg.DCGMExporterHost, e.cfg.DCGMExporterPort, e.cfg.DCGMExporterPath),
		}, nil
	}

	fieldSelector := "status.phase=Running"
	if e.cfg.NodeName != "" {
		fieldSelector = fmt.Sprintf("%s,spec.nodeName=%s", fieldSelector, e.cfg.NodeName)
	}

	// TODO: consider using an informer and keeping a list of pods which match the selector
	// at the moment seems like an overkill
	dcgmExporterList, err := e.dynamic.Resource(podGVR).Namespace("").List(ctx, metav1.ListOptions{
		LabelSelector: e.cfg.Selector,
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return []string{}, fmt.Errorf("error getting DCGM exporter pods %w", err)
	}

	urls := make([]string, len(dcgmExporterList.Items))
	for i := range dcgmExporterList.Items {
		var pod corev1.Pod
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(dcgmExporterList.Items[i].Object, &pod); err != nil {
			return nil, fmt.Errorf("converting unstructured to pod: %w", err)
		}
		urls[i] = fmt.Sprintf("http://%s:%d%s", pod.Status.PodIP, e.cfg.DCGMExporterPort, e.cfg.DCGMExporterPath)
	}

	return urls, nil
}

func (e *exporter) export(ctx context.Context) error {
	urls, err := e.getDCGMUrls(ctx)
	if err != nil {
		return err
	}

	if len(urls) == 0 {
		e.log.Info("no dcgm-exporter instances to scrape")
		return nil
	}

	metricFamilies, err := e.scraper.Scrape(ctx, urls)
	if err != nil {
		return fmt.Errorf("couldn't scrape DCGM exporters %w", err)
	}
	if len(metricFamilies) == 0 {
		e.log.Warnf("no metrics collected from %d dcgm-exporters", len(urls))
		return nil
	}
	now := time.Now()

	batch := e.mapper.Map(metricFamilies)
	if len(batch.Metrics) == 0 {
		e.log.Warnf("no metrics to export from activated metrics, scraped %d metrics from dcgm-exporter", len(urls))
		return nil
	}

	if err := e.client.UploadBatch(ctx, batch); err != nil {
		return fmt.Errorf("error while sending %d metrics to backend %w", len(batch.Metrics), err)
	}

	e.log.Infof("successfully exported %d metrics", len(batch.Metrics))

	// Export metrics to Custom Metrics API
	// Right now optionally, so any errors are logged and ignored
	if e.metricWriter != nil {
		gpuMetrics := e.mapper.MapToAvro(ctx, metricFamilies)
		for _, metric := range gpuMetrics {
			metric.Timestamp = now
			err = e.metricWriter.Write(metric)
			if err != nil {
				e.log.WithField("error", err.Error()).Warn("error while writing metrics to custom metrics api")
				break
			}
		}
	}

	return nil
}
