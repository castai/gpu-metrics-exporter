package exporter_test

import (
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/castai/gpu-metrics-exporter/internal/exporter"
	"github.com/castai/gpu-metrics-exporter/pb"
)

func newGauge(value float64) *dto.Gauge {
	return &dto.Gauge{
		Value: &value,
	}
}

func newLabelPair(name, value string) *dto.LabelPair {
	return &dto.LabelPair{
		Name:  &name,
		Value: &value,
	}
}

func TestMetricMapper_Map(t *testing.T) {
	mapper := exporter.NewMapper()

	t.Run("empty input yields empty MetricsBatch", func(t *testing.T) {
		ts := time.Now()
		metricFamilyMaps := []exporter.MetricFamiliyMap{}

		got := mapper.Map(metricFamilyMaps, ts)
		expected := &pb.MetricsBatch{}

		r := require.New(t)
		r.Equal(expected, got)
	})

	t.Run("metric familiy which is not enabled is skipped", func(t *testing.T) {
		ts := time.Now()
		metricFamilyMaps := []exporter.MetricFamiliyMap{
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
			},
		}

		got := mapper.Map(metricFamilyMaps, ts)
		expected := &pb.MetricsBatch{}

		r := require.New(t)
		r.Equal(expected, got)
	})

	t.Run("enabled metric family is included", func(t *testing.T) {
		ts := time.Now()
		metricFamilyMaps := []exporter.MetricFamiliyMap{
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

		got := mapper.Map(metricFamilyMaps, ts)
		expected := &pb.MetricsBatch{
			Metrics: []*pb.Metric{
				{
					Name: exporter.MetricGraphicsEngineActive,
					Measurements: []*pb.Metric_Measurement{
						{
							Value: 1.0,
							Ts:    timestamppb.New(ts),
							Labels: []*pb.Metric_Label{
								{Name: "label1", Value: "value1"},
							},
						},
					},
				},
			},
		}

		r := require.New(t)
		r.Equal(expected, got)
	})
}
