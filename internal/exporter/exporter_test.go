package exporter_test

import (
	"context"
	"errors"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/castai/gpu-metrics-exporter/internal/exporter"
	castai_mock "github.com/castai/gpu-metrics-exporter/mock/castai"
	mocks "github.com/castai/gpu-metrics-exporter/mock/exporter"
)

func TestExporter_Running(t *testing.T) {
	log := logrus.New()

	t.Run("discovers pods with labels and scrapes them", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		kubeClient := fake.NewSimpleClientset(&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dcgm-exporter",
				Namespace: "default",
				Labels:    map[string]string{"app": "dcgm-exporter"},
			},
			Status: corev1.PodStatus{
				PodIP: "192.168.1.1",
				Phase: corev1.PodRunning,
			},
		})

		config := exporter.Config{
			ExportInterval:   2 * time.Second,
			ScrapeInterval:   2 * time.Second,
			DCGMExporterPort: 9400,
			DCGMExporterPath: "/metrics",
			Selector:         "app=dcgm-exporter",
			Enabled:          true,
		}

		scraper := mocks.NewMockScraper(t)
		mapper := mocks.NewMockMetricMapper(t)
		client := castai_mock.NewMockClient(t)

		ex := exporter.NewExporter(config, kubeClient, log, scraper, mapper, client)
		ex.Enable()

		metricFamilies := []exporter.MetricFamilyMap{
			{
				"test_gauge": {
					Type: dto.MetricType_GAUGE.Enum(),
					Metric: []*dto.Metric{
						{
							Label: []*dto.LabelPair{
								newLabelPair("label1", "value1"),
							},
							Gauge: newGauge(1.0),
						},
					},
				},
				exporter.MetricGraphicsEngineActive: {
					Type: dto.MetricType_GAUGE.Enum(),
					Metric: []*dto.Metric{
						{
							Label: []*dto.LabelPair{
								newLabelPair("label1", "value1"),
							},
							Gauge: newGauge(1.0),
						},
					},
				},
			},
		}

		batch := &exporter.MetricsBatch{
			Metrics: map[string]exporter.MeasurementsByLabelKey{
				exporter.MetricGraphicsEngineActive: {
					"label1=value1": {
						Value:      1,
						NumSamples: 1,
						Labels: map[string]string{
							"label1": "value1",
						},
						LabelsKey: "label1=value1",
					},
				},
			},
		}
		pbBatch := batch.ToProto()

		scraper.EXPECT().Scrape(ctx, []string{"http://192.168.1.1:9400/metrics"}).Times(1).Return(metricFamilies, nil)
		mapper.EXPECT().Map(metricFamilies).Times(1).Return(batch, nil)
		client.EXPECT().UploadBatch(mock.Anything, pbBatch).Times(1).Return(nil, nil)

		go func() {
			err := ex.Start(ctx)
			if err != nil && !errors.Is(err, context.Canceled) {
				t.Errorf("unexpected error: %v", err)
			}
		}()

		time.Sleep(2400 * time.Millisecond)

		r := require.New(t)
		r.True(ex.Enabled())
	})

	t.Run("scrapes single host provided in config", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		kubeClient := fake.NewSimpleClientset()

		config := exporter.Config{
			ExportInterval:   2 * time.Second,
			ScrapeInterval:   2 * time.Second,
			DCGMExporterPort: 9400,
			DCGMExporterPath: "/metrics",
			DCGMExporterHost: "localhost",
			Enabled:          true,
		}

		scraper := mocks.NewMockScraper(t)
		mapper := mocks.NewMockMetricMapper(t)
		client := castai_mock.NewMockClient(t)

		ex := exporter.NewExporter(config, kubeClient, log, scraper, mapper, client)
		ex.Enable()

		metricFamilies := []exporter.MetricFamilyMap{
			{
				"test_gauge": {
					Type: dto.MetricType_GAUGE.Enum(),
					Metric: []*dto.Metric{
						{
							Label: []*dto.LabelPair{
								newLabelPair("label1", "value1"),
							},
							Gauge: newGauge(1.0),
						},
					},
				},
				exporter.MetricGraphicsEngineActive: {
					Type: dto.MetricType_GAUGE.Enum(),
					Metric: []*dto.Metric{
						{
							Label: []*dto.LabelPair{
								newLabelPair("label1", "value1"),
							},
							Gauge: newGauge(1.0),
						},
					},
				},
			},
		}

		batch := &exporter.MetricsBatch{
			Metrics: map[string]exporter.MeasurementsByLabelKey{
				exporter.MetricGraphicsEngineActive: {
					"label1=value1": {
						Value:      1,
						NumSamples: 1,
						Labels: map[string]string{
							"label1": "value1",
						},
						LabelsKey: "label1=value1",
					},
				},
			},
		}
		pbBatch := batch.ToProto()

		scraper.EXPECT().Scrape(ctx, []string{"http://localhost:9400/metrics"}).Times(1).Return(metricFamilies, nil)
		mapper.EXPECT().Map(metricFamilies).Times(1).Return(batch, nil)
		client.EXPECT().UploadBatch(mock.Anything, pbBatch).Times(1).Return(nil, nil)

		go func() {
			err := ex.Start(ctx)
			if err != nil && !errors.Is(err, context.Canceled) {
				t.Errorf("unexpected error: %v", err)
			}
		}()

		time.Sleep(2400 * time.Millisecond)

		r := require.New(t)
		r.True(ex.Enabled())
	})

	t.Run("don't send empty batch of metrics", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		kubeClient := fake.NewSimpleClientset()

		config := exporter.Config{
			ExportInterval:   2 * time.Second,
			ScrapeInterval:   2 * time.Second,
			DCGMExporterPort: 9400,
			DCGMExporterPath: "/metrics",
			DCGMExporterHost: "localhost",
			Enabled:          true,
		}

		scraper := mocks.NewMockScraper(t)
		mapper := mocks.NewMockMetricMapper(t)
		client := castai_mock.NewMockClient(t)

		ex := exporter.NewExporter(config, kubeClient, log, scraper, mapper, client)
		ex.Enable()

		metricFamilies := []exporter.MetricFamilyMap{
			{
				"test_gauge": {
					Type: dto.MetricType_GAUGE.Enum(),
					Metric: []*dto.Metric{
						{
							Label: []*dto.LabelPair{
								newLabelPair("label1", "value1"),
							},
							Gauge: newGauge(1.0),
						},
					},
				},
				exporter.MetricGraphicsEngineActive: {
					Type: dto.MetricType_GAUGE.Enum(),
					Metric: []*dto.Metric{
						{
							Label: []*dto.LabelPair{
								newLabelPair("label1", "value1"),
							},
							Gauge: newGauge(1.0),
						},
					},
				},
			},
		}

		batch := exporter.NewMetricsBatch()

		scraper.EXPECT().Scrape(ctx, []string{"http://localhost:9400/metrics"}).Times(1).Return(metricFamilies, nil)
		mapper.EXPECT().Map(metricFamilies).Times(1).Return(batch, nil)

		go func() {
			err := ex.Start(ctx)
			if err != nil && !errors.Is(err, context.Canceled) {
				t.Errorf("unexpected error: %v", err)
			}
		}()

		time.Sleep(2400 * time.Millisecond)

		r := require.New(t)
		r.True(ex.Enabled())
	})

	t.Run("varying scrape and export intervals", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		kubeClient := fake.NewSimpleClientset()

		config := exporter.Config{
			ExportInterval:   2 * time.Second,
			ScrapeInterval:   1 * time.Second,
			DCGMExporterPort: 9400,
			DCGMExporterPath: "/metrics",
			DCGMExporterHost: "localhost",
			Enabled:          true,
		}

		scraper := mocks.NewMockScraper(t)
		mapper := mocks.NewMockMetricMapper(t)
		client := castai_mock.NewMockClient(t)

		ex := exporter.NewExporter(config, kubeClient, log, scraper, mapper, client)
		ex.Enable()

		metricFamilies := []exporter.MetricFamilyMap{
			{
				exporter.MetricGraphicsEngineActive: {
					Type: dto.MetricType_GAUGE.Enum(),
					Metric: []*dto.Metric{
						{
							Label: []*dto.LabelPair{
								newLabelPair("label1", "value1"),
							},
							Gauge: newGauge(1.0),
						},
					},
				},
			},
		}

		batch := &exporter.MetricsBatch{
			Metrics: map[string]exporter.MeasurementsByLabelKey{
				exporter.MetricGraphicsEngineActive: {
					"label1=value1": {
						Value:      1,
						NumSamples: 2,
						Labels: map[string]string{
							"label1": "value1",
						},
						LabelsKey: "label1=value1",
					},
				},
			}}
		pbBatch := batch.ToProto()
		scraper.EXPECT().Scrape(ctx, []string{"http://localhost:9400/metrics"}).Times(2).Return(metricFamilies, nil)
		mapper.EXPECT().Map(metricFamilies).Times(2).Return(batch, nil)
		client.EXPECT().UploadBatch(ctx, pbBatch).Times(1).Return(nil)
		go func() {
			err := ex.Start(ctx)
			if err != nil && !errors.Is(err, context.Canceled) {
				t.Errorf("unexpected error: %v", err)
			}
		}()

		time.Sleep(2400 * time.Millisecond)

		r := require.New(t)
		r.True(ex.Enabled())
	})
}
