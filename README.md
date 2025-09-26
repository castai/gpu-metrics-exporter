# GPU Metrics Exporter

A tool to collect GPU metrics [from DCGM Exporter](https://github.com/NVIDIA/dcgm-exporter) instasnces, and forward them to cast.ai.

## How it works

The exporter can run as a sidecar to the DCGM DaemonSet, or as a single instance service in the cluster.
When it runs as a sidecar, the `DCGM_HOST` should be set. In this case it will only scrape metrics from that particular 
instance of DCGM and send them to cast.ai

If it is deployed as a single instance in the cluster, it will automatically discover the `DCGM` instances and scrape 
the metrics from them. If the `DCGM` instances have some custom labels, make sure to properly set the `DCGM_LABELS` 
environment variable.

It is also possible to deploy the DCGM exporter but have it configured to read the metrics from an existing 
nv-hostengine.

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

## Installation

### Helm

#### Cloning this repository

You can clone this repository and install the chart with the following commands:
```bash
$ cd charts/gpu-metrics-exporter
$ helm install --generate-name <deployment-name> -f values.yaml -f values-<k8s-provider>.yaml .
```
Where:
* `<deployment-name>` is a name of your choice
* `<k8s-provider>` is the name of the k8s provider you are using (e.g. `eks`, `gke`, `aks`, 'omni')
   * this sets the proper node affinity so the Daemon Set only runs on nodes with GPUs
#### Adding the cast.ai repository

You can add the cast.ai repository and install the chart with the following commands:

```bash
$ helm repo add castai https://castai.github.io/charts
$ helm repo update
$ helm pull castai/gpu-metrics-exporter --untar
$ cd gpu-metrics-exporter
$ helm install --generate-name castai/gpu-metrics-exporter -f values.yaml
```
#### Configuring the installation

By default, it will be deployed as a sidecar to the DCGM exporter. 
If you don't want to deploy it as a sidecar, in the values.yaml file you can:
1. Set `dcgmExporter.enabled` to false
2. Set the `DCGM_HOST` and `DCGM_LABELS` environment variables in `gpuMetricsExporter.config` of 
   the values.yaml file
   1. `DCGM_HOST` is the address of the DCGM exporter instance
   2. `DCGM_LABELS` is a comma-separated list of labels that the DCGM instances have
3. If you want to deploy the DCGM exporter but have it configured to read the metrics from an existing nv-hostengine,
you can:
   1. set the `dcgmExporter.useExternalHostEngine` to true in the values.yaml file
   2. it will try to connect to the 5555 port of the node.

