package exporter

import (
	"strings"

	client_model "github.com/prometheus/client_model/go"

	"github.com/castai/gpu-metrics-exporter/pb"
)

const (
	nodeNameLabel = "Hostname"
)

type MetricMapper interface {
	Map(metrics []MetricFamilyMap) *pb.MetricsBatch
}

type metricMapper struct {
	nodeName string
}

func NewMapper(nodeName string) MetricMapper {
	return &metricMapper{nodeName: nodeName}
}

func (p metricMapper) Map(metricFamilyMaps []MetricFamilyMap) *pb.MetricsBatch {
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
				labels := p.mapLabels(m.Label)
				var newValue float64
				switch t {
				case "COUNTER":
					newValue = *m.GetCounter().Value
				case "GAUGE":
					newValue = *m.GetGauge().Value
				}

				metric.Measurements = append(metric.Measurements, &pb.Metric_Measurement{
					Value:  newValue,
					Labels: labels,
				})
			}
		}
	}

	return metrics
}

func (p metricMapper) mapLabels(labelPairs []*client_model.LabelPair) []*pb.Metric_Label {
	labels := make([]*pb.Metric_Label, len(labelPairs))
	for i, label := range labelPairs {
		value := *label.Value
		if p.nodeName != "" && strings.EqualFold(*label.Name, nodeNameLabel) {
			value = p.nodeName
		}
		labels[i] = &pb.Metric_Label{
			Name:  *label.Name,
			Value: value,
		}
	}

	return labels
}
