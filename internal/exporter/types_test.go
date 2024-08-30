package exporter

import (
	"github.com/castai/gpu-metrics-exporter/pb"
	"github.com/stretchr/testify/require"
	"slices"
	"strings"
	"testing"
)

func TestMetricsBatchAggregate(t *testing.T) {
	firstBatch := &MetricsBatch{
		Metrics: map[string]MeasurementsByLabelKey{
			"metric1": {
				"label1:val1": {
					Value:      1,
					NumSamples: 1,
					Labels: map[string]string{
						"label1": "val1",
					},
					LabelsKey: "label1:val1",
				},
				"label1:val2": {
					Value:      2,
					NumSamples: 1,
					Labels: map[string]string{
						"label1": "val2",
					},
					LabelsKey: "label1:val2",
				},
			},
			"metric2": {
				"label2:val1": {
					Value:      3,
					NumSamples: 1,
					Labels: map[string]string{
						"label2": "val1",
					},
					LabelsKey: "label2:val1",
				},
			},
		},
	}

	secondBatch := &MetricsBatch{
		Metrics: map[string]MeasurementsByLabelKey{
			"metric1": { // should be aggregated with existing
				"label1:val1": { // should be aggregated with existing
					Value:      1,
					NumSamples: 1,
					Labels: map[string]string{
						"label1": "val1",
					},
					LabelsKey: "label1:val1",
				},
				"label1:val3": { // should be added, it doesn't exist in first batch
					Value:      2,
					NumSamples: 1,
					Labels: map[string]string{
						"label1": "val2",
					},
					LabelsKey: "label1:val2",
				},
			},
			"metric3": { // should be added, it doesn't exist in first batch
				"label2:val1": {
					Value:      3,
					NumSamples: 1,
					Labels: map[string]string{
						"label2": "val1",
					},
					LabelsKey: "label2:val1",
				},
			},
		}}
	expectedAfterMerge := &MetricsBatch{
		Metrics: map[string]MeasurementsByLabelKey{
			"metric1": {
				"label1:val1": {
					Value:      2,
					NumSamples: 2,
					Labels: map[string]string{
						"label1": "val1",
					},
					LabelsKey: "label1:val1",
				},
				"label1:val2": {
					Value:      2,
					NumSamples: 1,
					Labels: map[string]string{
						"label1": "val2",
					},
					LabelsKey: "label1:val2",
				},
				"label1:val3": {
					Value:      2,
					NumSamples: 1,
					Labels: map[string]string{
						"label1": "val2",
					},
					LabelsKey: "label1:val2",
				},
			},
			"metric2": {
				"label2:val1": {
					Value:      3,
					NumSamples: 1,
					Labels: map[string]string{
						"label2": "val1",
					},
					LabelsKey: "label2:val1",
				},
			},
			"metric3": {
				"label2:val1": {
					Value:      3,
					NumSamples: 1,
					Labels: map[string]string{
						"label2": "val1",
					},
					LabelsKey: "label2:val1",
				},
			},
		},
	}

	firstBatch.aggregate(secondBatch)
	r := require.New(t)
	r.Equal(expectedAfterMerge, firstBatch)
}

func TestMetricBatchToProto(t *testing.T) {
	batch := &MetricsBatch{
		Metrics: map[string]MeasurementsByLabelKey{
			"metric1": {
				"label1=val1;label2=val2": {
					Value:      1,
					NumSamples: 1,
					Labels: map[string]string{
						"label1": "val1",
						"label2": "val2",
					},
				},
				"label1=val2": {
					Value:      2,
					NumSamples: 1,
					Labels: map[string]string{
						"label1": "val2",
					},
				},
			},
			"metric2": {
				"label2=val1": {
					Value:      1,
					NumSamples: 4,
					Labels: map[string]string{
						"label2": "val1",
					},
				},
			},
		},
	}
	expectedProto := &pb.MetricsBatch{
		Metrics: []*pb.Metric{
			{
				Name: "metric1",
				Measurements: []*pb.Metric_Measurement{
					{
						Value: 1,
						Labels: []*pb.Metric_Label{
							{
								Name:  "label1",
								Value: "val1",
							},
							{
								Name:  "label2",
								Value: "val2",
							},
						},
					},
					{
						Value: 2,
						Labels: []*pb.Metric_Label{
							{
								Name:  "label1",
								Value: "val2",
							},
						},
					},
				},
			},
			{
				Name: "metric2",
				Measurements: []*pb.Metric_Measurement{
					{
						Value: 0.25,
						Labels: []*pb.Metric_Label{
							{
								Name:  "label2",
								Value: "val1",
							},
						},
					},
				},
			},
		},
	}
	r := require.New(t)
	gotProto := batch.ToProto()
	// gotProto is generated from maps, so we need to sort it to compare
	slices.SortFunc(gotProto.Metrics, func(i, j *pb.Metric) int {
		return strings.Compare(i.Name, j.Name)
	})
	for _, metric := range gotProto.Metrics {
		slices.SortFunc(metric.Measurements, func(i, j *pb.Metric_Measurement) int {
			if i.Value < j.Value {
				return -1
			} else if i.Value == j.Value {
				return 0
			}
			return 1
		})
	}
	r.Equal(expectedProto, gotProto)
}
