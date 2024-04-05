# GPU Metrics Exporter

A tool to collect GPU metrics [from DCGM Exporter](https://github.com/NVIDIA/dcgm-exporter) instasnces, and forward them to cast.ai.

## How it works

The exporter can run as a sidecar to the DCGM DaemonSet, or as a single instance service in the cluster.
When it runs as a sidecar, the `DCGM_HOST` should be set. In this case it will only scrape metrics from that particular instance of DCGM and send them to cast.ai

If it is deployed as a single instance in the cluster, it will automatically discover the `DCGM` instances and scrape the metrics from them. If the `DCGM` instances have some custom labels, make sure to properly set the `DCGM_LABELS` environment variable.

## Scraped metrics

Make sure that these fields are exposed by DCGM exporter as metrics:

```
DCGM_FI_PROF_SM_ACTIVE
DCGM_FI_PROF_SM_OCCUPANCY
DCGM_FI_PROF_PIPE_TENSOR_ACTIVE
DCGM_FI_PROF_DRAM_ACTIVE
DCGM_FI_PROF_PCIE_TX_BYTES
DCGM_FI_PROF_PCIE_RX_BYTES
DCGM_FI_PROF_GR_ENGINE_ACTIVE
DCGM_FI_DEV_FB_TOTAL
DCGM_FI_DEV_FB_FREE
DCGM_FI_DEV_FB_USED
DCGM_FI_DEV_PCIE_LINK_GEN
DCGM_FI_DEV_PCIE_LINK_WIDTH
DCGM_FI_DEV_GPU_TEMP
DCGM_FI_DEV_MEMORY_TEMP
DCGM_FI_DEV_POWER_USAGE
```