package exporter

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	cleanupInterval = 3 * time.Minute
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
}

func NewExporter(cfg Config, kube kubernetes.Interface, log logrus.FieldLogger, scraper Scraper, mapper MetricMapper) Exporter {
	enabled := atomic.Bool{}
	enabled.Store(cfg.Enabled)

	return &exporter{
		cfg:     cfg,
		kube:    kube,
		log:     log,
		scraper: scraper,
		mapper:  mapper,
		enabled: &enabled,
	}
}

func (e *exporter) Start(ctx context.Context) error {
	exportTicker := time.NewTicker(e.cfg.ExportInterval)
	defer exportTicker.Stop()

	cleanupTicker := time.NewTicker(cleanupInterval)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-exportTicker.C:
			if !e.enabled.Load() {
				continue
			}
			if err := e.collect(ctx); err != nil {
				e.log.Errorf("collect error: %v", err)
			}
		case <-cleanupTicker.C:
			// TODO: call cleanup procedure
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

func (e *exporter) collect(ctx context.Context) error {
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
	metrics := e.mapper.Map(metricFamilies, time.Now())
	_ = metrics

	return nil
}
