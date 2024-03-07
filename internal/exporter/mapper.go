package exporter

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/castai/gpu-metrics-exporter/pb"
)

type MetricMapper interface {
	Map(metrics []MetricFamiliyMap, ts time.Time) *pb.MetricsBatch
}

type metricMapper struct{}

func NewMapper() MetricMapper {
	return &metricMapper{}
}

func (p metricMapper) Map(metricFamilyMaps []MetricFamiliyMap, ts time.Time) *pb.MetricsBatch {
	metrics := &pb.MetricsBatch{}
	metricsMap := make(map[string]*pb.Metric)
	for _, familyMap := range metricFamilyMaps {
		for name, family := range familyMap {
			if _, found := EnabledMetrics[name]; !found {
				continue
			}

			metric, found := metricsMap[name]
			if !found {
				metric = &pb.Metric{
					Name: name,
				}
				metricsMap[name] = metric
				metrics.Metrics = append(metrics.Metrics, metric)
			}
			t := family.Type.String()
			for _, m := range family.Metric {
				labels := []*pb.Metric_Label{}
				for _, l := range m.Label {
					labels = append(labels, &pb.Metric_Label{
						Name:  *l.Name,
						Value: *l.Value,
					})
				}
				switch t {
				case "COUNTER":
					metric.Measurements = append(metric.Measurements, &pb.Metric_Measurement{
						Value:  *m.GetCounter().Value,
						Ts:     timestamppb.New(ts),
						Labels: labels,
					})
				case "GAUGE":
					metric.Measurements = append(metric.Measurements, &pb.Metric_Measurement{
						Value:  *m.GetGauge().Value,
						Ts:     timestamppb.New(ts),
						Labels: labels,
					})
				}
			}
		}
	}

	return metrics
}