package config

import (
	"fmt"
	"net/url"
	"strings"
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
	DCGMHost            string            `envconfig:"DCGM_HOST"`
	NodeName            string            `envconfig:"NODE_NAME"`
	ExportInterval      time.Duration     `envconfig:"EXPORT_INTERVAL" default:"15s"`
	CastAPI             string            `envconfig:"CAST_API" default:"https://api.cast.ai"`
	ClusterID           string            `envconfig:"CLUSTER_ID"`
	APIKey              string            `envconfig:"API_KEY"`
	TelemetryURL        string            `envconfig:"TELEMETRY_URL" default:""`
}

func deriveTelemetryURL(apiURL string) string {
	if apiURL == "" {
		return ""
	}

	parsed, err := url.Parse(apiURL)
	if err != nil {
		return ""
	}

	host := parsed.Host
	if host == "" {
		return ""
	}

	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// handle local dev URLs
	if strings.HasSuffix(host, ".local.cast.ai") {
		parts := strings.Split(host, ".")
		if len(parts) == 4 && strings.HasPrefix(parts[0], "api--") {
			name := strings.TrimPrefix(parts[0], "api--")
			return fmt.Sprintf("api-grpc--%s.local.cast.ai", name)
		}
		return ""
	}

	// api.cast.ai → [api, cast, ai] → 3 parts → prod-master
	// api.dev-master.cast.ai → [api, dev-master, cast, ai] → 4 parts → dev-master
	// api.eu.cast.ai → [api, eu, cast, ai] → 4 parts → prod-eu
	parts := strings.Split(host, ".")
	if len(parts) < 3 || parts[0] != "api" || parts[len(parts)-2] != "cast" || parts[len(parts)-1] != "ai" {
		return ""
	}

	var env string
	if len(parts) == 3 {
		env = "prod-master"
	} else if len(parts) == 4 {
		env = parts[1]
		if env == "eu" {
			env = "prod-eu"
		}
	} else {
		return ""
	}

	return fmt.Sprintf("telemetry.%s.cast.ai", env)
}

func GetFromEnvironment() (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}

	if cfg.TelemetryURL == "" {
		cfg.TelemetryURL = deriveTelemetryURL(cfg.CastAPI)
	}

	return cfg, nil
}
