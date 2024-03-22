package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	HTTPListenPort      int               `envconfig:"HTTP_LISTEN_PORT" default:"6061"`
	LogLevel            string            `envconfig:"LOG_LEVEL" default:"info"`
	KubeConfigPath      string            `envconfig:"KUBE_CONFIG_PATH" default:""`
	DCGMLabels          map[string]string `envconfig:"DCGM_LABELS" default:"app.kubernetes.io/name:dcgm-exporter"`
	DCGMPort            int               `envconfig:"DCGM_PORT" default:"9400"`
	DCGMMetricsEndpoint string            `envconfig:"DCGM_METRICS_ENDPOINT" default:"/metrics"`
	ExportInterval      time.Duration     `envconfig:"EXPORT_INTERVAL" default:"15s"`
	CastAPI             string            `envconfig:"CAST_API" default:"https://api.cast.ai"`
	ClusterID           string            `envconfig:"CLUSTER_ID"`
	APIKey              string            `envconfig:"API_KEY"`
}

func GetFromEnvironment() (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
