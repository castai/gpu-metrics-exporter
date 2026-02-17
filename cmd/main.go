package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"

	"github.com/castai/gpu-metrics-exporter/internal/castai"
	"github.com/castai/gpu-metrics-exporter/internal/config"
	"github.com/castai/gpu-metrics-exporter/internal/exporter"
	"github.com/castai/gpu-metrics-exporter/internal/server"
	"github.com/castai/gpu-metrics-exporter/internal/workload"
	"github.com/castai/logging"
	"github.com/castai/metrics"
)

var (
	GitCommit = "undefined"
	GitRef    = "no-ref"
	Version   = "local"
)

const (
	workloadCacheSize = 512
	workloadsLabelKey = "workloads.cast.ai/custom-workload"
)

func main() {
	log := logrus.New()

	cfg, err := config.GetFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	logLevel, err := parseLogLevel(cfg.LogLevel)
	if err != nil {
		log.Warnf("failed to parse log level, defaulting to 'info': %v", err)
		logLevel = slog.LevelInfo
	}

	castaiLogger := logging.New(logging.NewTextHandler(logging.TextHandlerConfig{
		Output: os.Stdout,
		Level:  logLevel,
	}))

	if err := run(cfg, castaiLogger); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}

func parseLogLevel(level string) (slog.Level, error) {
	var lvl slog.Level
	err := lvl.UnmarshalText([]byte(level))
	return lvl, err
}

func run(cfg *config.Config, log *logging.Logger) error {
	mux := server.NewServerMux()

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPListenPort),
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		stopper := make(chan os.Signal, 1)
		signal.Notify(stopper, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		<-stopper

		ctx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Errorf("http server shutdown: %v", err)
		}

		cancel()
	}()

	dynClient, err := newDynamicClient(cfg)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("failed to create kubernetes dynamic client")
	}

	labelSelector, err := selectorFromMap(cfg.DCGMLabels)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("failed to create get label selector")
	}

	metricClient, err := metrics.NewMetricClient(
		metrics.Config{
			APIAddr:   cfg.TelemetryURL,
			APIToken:  cfg.APIKey,
			ClusterID: cfg.ClusterID,
		}, log)
	if err != nil {
		log.WithField("error", err.Error()).Warn("failed to create metrics client")
	}

	if metricClient != nil {
		go func() {
			if err := metricClient.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
				log.WithField("error", err.Error()).Error("error in metrics client")
			}
		}()
	}

	client := setupCastAIClient(log, cfg)
	scraper := exporter.NewScraper(&http.Client{}, log)
	workloadResolver, err := workload.NewResolver(dynClient, workload.Config{
		LabelKeys: []string{workloadsLabelKey},
		CacheSize: workloadCacheSize,
	})
	if err != nil {
		log.WithField("error", err.Error()).Fatal("failed to create workload resolver")
	}

	mapper := exporter.NewMapper(cfg.NodeName, workloadResolver, log)
	ex := exporter.NewExporter(exporter.Config{
		ExportInterval:   cfg.ExportInterval,
		Selector:         labelSelector.String(),
		DCGMExporterPort: cfg.DCGMPort,
		DCGMExporterPath: cfg.DCGMMetricsEndpoint,
		DCGMExporterHost: cfg.DCGMHost,
		Enabled:          true,
		NodeName:         cfg.NodeName,
	}, dynClient, log, scraper, mapper, client, metricClient)

	go func() {
		if err := ex.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Errorf("exporter stopped with error %v", err)
			cancel()
		}
	}()

	return srv.ListenAndServe()
}

func newDynamicClient(cfg *config.Config) (dynamic.Interface, error) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", cfg.KubeConfigPath)
	if err != nil {
		return nil, err
	}
	restConfig.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(float32(10), 25)

	return dynamic.NewForConfig(restConfig)
}

func selectorFromMap(labelMap map[string]string) (labels.Selector, error) {
	selector := labels.NewSelector()
	var requirements labels.Requirements

	for label, value := range labelMap {
		req, err := labels.NewRequirement(label, selection.Equals, []string{value})
		if err != nil {
			return nil, err
		}
		requirements = append(requirements, *req)
	}

	return selector.Add(requirements...), nil
}

func setupCastAIClient(log *logging.Logger, cfg *config.Config) castai.Client {
	clientConfig := castai.Config{
		ClusterID: cfg.ClusterID,
		APIKey:    cfg.APIKey,
		URL:       cfg.CastAPI,
	}
	restyClient := resty.NewWithClient(&http.Client{
		Timeout: 2 * time.Minute,
		Transport: &http.Transport{
			Proxy:             http.ProxyFromEnvironment,
			ForceAttemptHTTP2: true,
		},
	})

	return castai.NewClient(clientConfig, log, restyClient, Version)
}
