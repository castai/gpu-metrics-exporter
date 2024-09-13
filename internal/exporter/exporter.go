package exporter

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/castai/gpu-metrics-exporter/internal/castai"
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
	cfg     Config
	kube    kubernetes.Interface
	log     logrus.FieldLogger
	scraper Scraper
	mapper  MetricMapper
	enabled *atomic.Bool
	client  castai.Client
}

func NewExporter(
	cfg Config,
	kube kubernetes.Interface,
	log logrus.FieldLogger,
	scraper Scraper,
	mapper MetricMapper,
	castaiClient castai.Client,
) Exporter {
	enabled := atomic.Bool{}
	enabled.Store(cfg.Enabled)

	return &exporter{
		cfg:     cfg,
		kube:    kube,
		log:     log,
		scraper: scraper,
		mapper:  mapper,
		enabled: &enabled,
		client:  castaiClient,
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
				e.log.Errorf("export error: %v", err)
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
	dcgmExporterList, err := e.kube.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		LabelSelector: e.cfg.Selector,
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return []string{}, fmt.Errorf("error getting DCGM exporter pods %w", err)
	}

	urls := make([]string, len(dcgmExporterList.Items))
	for i := range dcgmExporterList.Items {
		dcgmExporter := dcgmExporterList.Items[i]
		urls[i] = fmt.Sprintf("http://%s:%d%s", dcgmExporter.Status.PodIP, e.cfg.DCGMExporterPort, e.cfg.DCGMExporterPath)
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

	batch := e.mapper.Map(metricFamilies)
	if len(batch.Metrics) == 0 {
		e.log.Warn("no metrics to export from activated metrics, scraped %d metrics from dcgm-exporter", len(metricFamilies))
		return nil
	}

	if err := e.client.UploadBatch(ctx, batch); err != nil {
		return fmt.Errorf("error whlie sending %d metrics to castai %w", len(batch.Metrics), err)
	}

	e.log.Infof("successfully exported %d metrics", len(batch.Metrics))

	return nil
}
