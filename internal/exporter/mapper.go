package exporter

type MetricMapper interface {
	Map(metrics []MetricFamilyMap) *MetricsBatch
}

type metricMapper struct{}

func NewMapper() MetricMapper {
	return &metricMapper{}
}

func (p metricMapper) Map(metricFamilyMaps []MetricFamilyMap) *MetricsBatch {
	newBatch := NewMetricsBatch()
	for _, familyMap := range metricFamilyMaps {
		for name, family := range familyMap {
			if _, found := EnabledMetrics[name]; !found {
				continue
			}

			metric, found := newBatch.Metrics[name]
			if !found {
				metric = MeasurementsByLabelKey{}
				newBatch.Metrics[name] = metric
			}

			t := family.Type.String()

			for _, m := range family.Metric {
				labels := make(map[string]string, len(m.Label))
				for _, label := range m.Label {
					labels[*label.Name] = *label.Value
				}
				var newValue float64
				switch t {
				case "COUNTER":
					newValue = *m.GetCounter().Value
				case "GAUGE":
					newValue = *m.GetGauge().Value
				}

				msrmt := newMeasurement(newValue, labels)
				metric[msrmt.LabelsKey] = msrmt
			}
		}
	}

	return newBatch
}
