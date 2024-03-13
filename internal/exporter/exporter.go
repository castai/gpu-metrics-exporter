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
	Selector         string
	Enabled          bool
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

func (e *exporter) export(ctx context.Context) error {
	// TODO: consider using an informer and keeping a list of pods which match the selector
	// at the moment seems like an overkill
	dcgmExporterList, err := e.kube.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		LabelSelector: e.cfg.Selector,
		FieldSelector: "status.phase=Running",
	})
	if err != nil {
		return fmt.Errorf("error getting DCGM exporter pods %w", err)
	}

	urls := make([]string, len(dcgmExporterList.Items))
	for i := range dcgmExporterList.Items {
		dcgmExporter := dcgmExporterList.Items[i]
		urls[i] = fmt.Sprintf("http://%s:%d%s", dcgmExporter.Status.PodIP, e.cfg.DCGMExporterPort, e.cfg.DCGMExporterPath)
	}

	metricFamilies, err := e.scraper.Scrape(ctx, urls)
	if err != nil {
		return fmt.Errorf("couldn't scrape DCGM exporters %w", err)
	}

	batch := e.mapper.Map(metricFamilies, time.Now())
	if err := e.client.UploadBatch(ctx, batch); err != nil {
		return fmt.Errorf("error whlie sending metrics %d to castai %w", len(batch.Metrics), err)
	}

	e.log.Infof("successfully exported %d metrics", len(batch.Metrics))

	return nil
}
