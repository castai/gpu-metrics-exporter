package exporter

type MetricName = string

const (
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
)

var (
	EnabledMetrics = map[MetricName]struct{}{
		MetricStreamingMultiProcessorActive:       {},
		MetricStreamingMultiProcessorOccupancy:    {},
		MetricStreamingMultiProcessorTensorActive: {},
		MetricDRAMActive:                          {},
		MetricPCIeTXBytes:                         {},
		MetricPCIeRXBytes:                         {},
		MetricGraphicsEngineActive:                {},
		MetricFrameBufferTotal:                    {},
		MetricFrameBufferFree:                     {},
		MetricFrameBufferUsed:                     {},
		MetricPCIeLinkGen:                         {},
		MetricPCIeLinkWidth:                       {},
		MetricGPUTemperature:                      {},
		MetricMemoryTemperature:                   {},
		MetricPowerUsage:                          {},
	}
)
