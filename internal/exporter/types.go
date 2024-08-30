package exporter

import (
	"fmt"
	"strings"

	"github.com/castai/gpu-metrics-exporter/pb"
)

type MetricName = string

const (
	MetricNameGPUGraphicsEngineActive  = "gpu_graphic_engine_active"
	MetricNameGPUFrameBufferUsedBytes  = "gpu_framebuffer_used_bytes"
	MetricNameGPUFrameBufferFreeBytes  = "gpu_framebuffer_free_bytes"
	MetricNameGPUFrameBufferTotalBytes = "gpu_framebuffer_total_bytes"

	MetricStreamingMultiProcessorActive       = MetricName("DCGM_FI_PROF_SM_ACTIVE")
	MetricStreamingMultiProcessorOccupancy    = MetricName("DCGM_FI_PROF_SM_OCCUPANCY")
	MetricStreamingMultiProcessorTensorActive = MetricName("DCGM_FI_PROF_PIPE_TENSOR_ACTIVE")
	MetricDRAMActive                          = MetricName("DCGM_FI_PROF_DRAM_ACTIVE")
	MetricPCIeTXBytes                         = MetricName("DCGM_FI_PROF_PCIE_TX_BYTES")
	MetricPCIeRXBytes                         = MetricName("DCGM_FI_PROF_PCIE_RX_BYTES")
	MetricGraphicsEngineActive                = MetricName("DCGM_FI_PROF_GR_ENGINE_ACTIVE")
	MetricFrameBufferTotal                    = MetricName("DCGM_FI_DEV_FB_TOTAL")
	MetricFrameBufferFree                     = MetricName("DCGM_FI_DEV_FB_FREE")
	MetricFrameBufferUsed                     = MetricName("DCGM_FI_DEV_FB_USED")
	MetricPCIeLinkGen                         = MetricName("DCGM_FI_DEV_PCIE_LINK_GEN")
	MetricPCIeLinkWidth                       = MetricName("DCGM_FI_DEV_PCIE_LINK_WIDTH")
	MetricGPUTemperature                      = MetricName("DCGM_FI_DEV_GPU_TEMP")
	MetricMemoryTemperature                   = MetricName("DCGM_FI_DEV_MEMORY_TEMP")
	MetricPowerUsage                          = MetricName("DCGM_FI_DEV_POWER_USAGE")
	MetricIntPipeActive                       = MetricName("DCGM_FI_PROF_PIPE_INT_ACTIVE")
	MetricFloat16PipeActive                   = MetricName("DCGM_FI_PROF_PIPE_FP16_ACTIVE")
	MetricFloat32PipeActive                   = MetricName("DCGM_FI_PROF_PIPE_FP32_ACTIVE")
	MetricFloat64PipeActive                   = MetricName("DCGM_FI_PROF_PIPE_FP64_ACTIVE")

	float64ZeroThreshold = float64(1e-6)
)

var (
	EnabledMetrics = map[MetricName]struct{}{
		MetricGraphicsEngineActive:             {},
		MetricFrameBufferTotal:                 {},
		MetricFrameBufferFree:                  {},
		MetricFrameBufferUsed:                  {},
		MetricStreamingMultiProcessorActive:    {},
		MetricStreamingMultiProcessorOccupancy: {},
		MetricDRAMActive:                       {},
		MetricIntPipeActive:                    {},
		MetricFloat16PipeActive:                {},
		MetricFloat32PipeActive:                {},
		MetricFloat64PipeActive:                {},
	}
)

type Measurement struct {
	Value      float64
	NumSamples int
	Labels     map[string]string
	LabelsKey  string
}

func newMeasurement(value float64, labels map[string]string) *Measurement {
	newM := &Measurement{
		Value:      value,
		NumSamples: 1,
		Labels:     labels,
	}
	newM.generateLabelsKey()
	return newM
}

func (m *Measurement) toLabelsProto() []*pb.Metric_Label {
	labelsArray := make([]*pb.Metric_Label, len(m.Labels))
	currentLbl := 0
	for labelName, labelValue := range m.Labels {
		labelsArray[currentLbl] = &pb.Metric_Label{
			Name:  labelName,
			Value: labelValue,
		}
		currentLbl++
	}
	return labelsArray
}

func (m *Measurement) toProto() *pb.Metric_Measurement {
	// Optimizations for better compression
	var val float64
	if m.NumSamples == 0 || m.Value < float64ZeroThreshold {
		val = 0
	} else if m.NumSamples == 1 {
		val = m.Value
	} else {
		val = m.Value / float64(m.NumSamples)
	}
	return &pb.Metric_Measurement{
		Value:  val,
		Labels: m.toLabelsProto(),
	}
}

func (m *Measurement) generateLabelsKey() {
	lblPairs := make([]string, len(m.Labels))
	currLbl := 0
	for labelName, labelValue := range m.Labels {
		lblPairs[currLbl] = fmt.Sprintf("%s=%s", labelName, labelValue)
		currLbl++
	}
	m.LabelsKey = strings.Join(lblPairs, ";")
}

type MeasurementsByLabelKey map[string]*Measurement

func (m MeasurementsByLabelKey) toProto() []*pb.Metric_Measurement {
	mappedMeasurements := make([]*pb.Metric_Measurement, len(m))
	currentMsrm := 0
	for _, msrm := range (map[string]*Measurement)(m) {
		mappedMeasurements[currentMsrm] = msrm.toProto()
		currentMsrm++
	}
	return mappedMeasurements
}

type MetricsBatch struct {
	Metrics map[string]MeasurementsByLabelKey
}

func NewMetricsBatch() *MetricsBatch {
	return &MetricsBatch{
		Metrics: make(map[string]MeasurementsByLabelKey),
	}
}

func (b *MetricsBatch) aggregate(o *MetricsBatch) {
	if o == nil {
		return
	}
	if b.Metrics == nil {
		b.Metrics = make(map[string]MeasurementsByLabelKey)
	}
	for metricName, measurements := range o.Metrics {
		var existingMeasurements MeasurementsByLabelKey
		var found bool
		if existingMeasurements, found = b.Metrics[metricName]; !found {
			b.Metrics[metricName] = measurements
			continue
		}
		for labelsKey, msrm := range measurements {
			var existing *Measurement
			var foundM bool
			if existing, foundM = existingMeasurements[labelsKey]; !foundM {
				existingMeasurements[labelsKey] = msrm
				continue
			}
			existing.Value += msrm.Value
			existing.NumSamples += msrm.NumSamples
		}
	}
}

func (b *MetricsBatch) ToProto() *pb.MetricsBatch {
	metrics := make([]*pb.Metric, len(b.Metrics))
	currMetric := 0
	for metricName, measurements := range b.Metrics {
		metrics[currMetric] = &pb.Metric{
			Name:         metricName,
			Measurements: measurements.toProto(),
		}
		currMetric++
	}

	return &pb.MetricsBatch{
		Metrics: metrics,
	}
}
