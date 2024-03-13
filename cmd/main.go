package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"

	"github.com/castai/gpu-metrics-exporter/internal/castai"
	"github.com/castai/gpu-metrics-exporter/internal/config"
	"github.com/castai/gpu-metrics-exporter/internal/exporter"
	"github.com/castai/gpu-metrics-exporter/internal/server"
)

var (
	GitCommit = "undefined"
	GitRef    = "no-ref"
	Version   = "local"
)

func main() {
	log := logrus.New()

	cfg, err := config.GetFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	logLevel, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(logLevel)

	if err := run(cfg, log); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}

func run(cfg *config.Config, log logrus.FieldLogger) error {
	mux := server.NewServerMux(log)

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

	clientset, err := newKubernetesClientset(cfg)
	if err != nil {
		log.Fatal(err)
	}

	labelSelector, err := selectorFromMap(cfg.DCGMLabels)
	if err != nil {
		log.Fatal(err)
	}

	client := setupCastAIClient(log, cfg)
	scraper := exporter.NewScraper(&http.Client{}, log)
	mapper := exporter.NewMapper()
	ex := exporter.NewExporter(exporter.Config{
		ExportInterval:   cfg.ExportInterval,
		Selector:         labelSelector.String(),
		DCGMExporterPort: cfg.DCGMPort,
		DCGMExporterPath: cfg.DCGMMetricsEndpoint,
		Enabled:          true,
	}, clientset, log, scraper, mapper, client)

	go func() {
		if err := ex.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Errorf("exporter stopped with error %v", err)
			cancel()
		}
	}()

	return srv.ListenAndServe()
}

func newKubernetesClientset(cfg *config.Config) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", cfg.KubeConfigPath)
	if err != nil {
		return nil, err
	}
	config.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(float32(10), 25)

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
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

func setupCastAIClient(log logrus.FieldLogger, cfg *config.Config) castai.Client {
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
