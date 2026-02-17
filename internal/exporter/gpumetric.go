package exporter

import "time"

type GPUMetric struct {
	NodeName      string `avro:"node_name"`
	ModelName     string `avro:"model_name"`
	Device        string `avro:"device"`
	DeviceID      string `avro:"device_id"`
	DeviceUUID    string `avro:"device_uuid"`
	MIGProfile    string `avro:"mig_profile"`
	MIGInstanceID string `avro:"mig_instance_id"`

	Pod          string `avro:"pod"`
	Container    string `avro:"container"`
	Namespace    string `avro:"namespace"`
	WorkloadName string `avro:"workload_name"`
	WorkloadKind string `avro:"workload_kind"`

	SMActive             float64 `avro:"sm_active"`
	SMOccupancy          float64 `avro:"sm_occupancy"`
	TensorActive         float64 `avro:"tensor_active"`
	DRAMActive           float64 `avro:"dram_active"`
	PCIeTXBytes          float64 `avro:"pcie_tx_bytes"`
	PCIeRXBytes          float64 `avro:"pcie_rx_bytes"`
	NVLinkTXBytes        float64 `avro:"nvlink_tx_bytes"`
	NVLinkRXBytes        float64 `avro:"nvlink_rx_bytes"`
	GraphicsEngineActive float64 `avro:"graphics_engine_active"`
	FramebufferTotal     float64 `avro:"framebuffer_total"`
	FramebufferUsed      float64 `avro:"framebuffer_used"`
	FramebufferFree      float64 `avro:"framebuffer_free"`
	PCIeLinkGen          float64 `avro:"pcie_link_gen"`
	PCIeLinkWidth        float64 `avro:"pcie_link_width"`
	Temperature          float64 `avro:"temperature"`
	MemoryTemperature    float64 `avro:"memory_temperature"`
	PowerUsage           float64 `avro:"power_usage"`
	GPUUtilization       float64 `avro:"gpu_utilization"`
	IntPipeActive        float64 `avro:"int_pipe_active"`
	FP16PipeActive       float64 `avro:"fp16_pipe_active"`
	FP32PipeActive       float64 `avro:"fp32_pipe_active"`
	FP64PipeActive       float64 `avro:"fp64_pipe_active"`
	ClocksEventReasons   float64 `avro:"clocks_event_reasons"`
	XIDErrors            float64 `avro:"xid_errors"`
	PowerViolation       float64 `avro:"power_violation"`
	ThermalViolation     float64 `avro:"thermal_violation"`

	Timestamp time.Time `avro:"ts"`
}
