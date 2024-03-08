package exporter_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/castai/gpu-metrics-exporter/internal/exporter"
	mocks "github.com/castai/gpu-metrics-exporter/mock/exporter"
)

var (
	metricsString = `
	# HELP DCGM_FI_DEV_GPU_TEMP Current temperature readings for the device in degrees C.
	# TYPE DCGM_FI_DEV_GPU_TEMP gauge
	DCGM_FI_DEV_GPU_TEMP{gpu="0",UUID="GPU-93461651-6be6-8fb7-a69a-c9eedc6984db",device="nvidia0",modelName="Tesla T4",Hostname="gke-gpu-default-pool",container="",namespace="",pod=""} 40
	`
)

func TestScraper_Scrape(t *testing.T) {
	log := logrus.New()

	t.Run("scrapes metrics without error", func(t *testing.T) {
		httpClient := mocks.NewMockHTTPClient(t)
		scraper := exporter.NewScraper(httpClient, log)

		response1 := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(metricsString)),
		}

		response2 := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(metricsString)),
		}

		response3 := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(metricsString)),
		}

		httpClient.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
			return req.URL.Host == "localhost:9400" && req.URL.Path == "/metrics"
		})).Times(1).Return(response1, nil)

		httpClient.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
			return req.URL.Host == "localhost:9410" && req.URL.Path == "/metrics"
		})).Times(1).Return(response2, nil)

		httpClient.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
			return req.URL.Host == "localhost:9420" && req.URL.Path == "/metrics"
		})).Times(1).Return(response3, nil)

		metricsFamily, err := scraper.Scrape(
			context.Background(),
			[]string{
				"http://localhost:9400/metrics",
				"http://localhost:9410/metrics",
				"http://localhost:9420/metrics",
			})

		r := require.New(t)
		r.NoError(err)
		r.NotNil(metricsFamily)
		r.Len(metricsFamily, 3)
		r.NotEmpty(metricsFamily[0])
		r.NotEmpty(metricsFamily[1])
		r.NotEmpty(metricsFamily[2])
	})

	t.Run("partially scrapes metrics when some exporter returns non-200 code", func(t *testing.T) {
		httpClient := mocks.NewMockHTTPClient(t)
		scraper := exporter.NewScraper(httpClient, log)

		response := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(metricsString)),
		}

		response1 := &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("")),
		}

		response2 := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(metricsString)),
		}

		httpClient.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
			return req.URL.Host == "localhost:9400" && req.URL.Path == "/metrics"
		})).Times(1).Return(response, nil)

		httpClient.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
			return req.URL.Host == "localhost:9410" && req.URL.Path == "/metrics"
		})).Times(1).Return(response1, nil)

		httpClient.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
			return req.URL.Host == "localhost:9420" && req.URL.Path == "/metrics"
		})).Times(1).Return(response2, nil)

		metricsFamily, err := scraper.Scrape(
			context.Background(),
			[]string{
				"http://localhost:9400/metrics",
				"http://localhost:9410/metrics",
				"http://localhost:9420/metrics",
			})

		r := require.New(t)
		r.NoError(err)
		r.NotNil(metricsFamily)
		r.Len(metricsFamily, 2)
		r.NotEmpty(metricsFamily[0])
		r.NotEmpty(metricsFamily[1])
	})

	t.Run("partially scrapes metrics when some exporter cannot be scraped", func(t *testing.T) {
		httpClient := mocks.NewMockHTTPClient(t)
		scraper := exporter.NewScraper(httpClient, log)

		response := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(metricsString)),
		}

		response1 := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(metricsString)),
		}

		httpClient.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
			return req.URL.Host == "localhost:9400" && req.URL.Path == "/metrics"
		})).Times(1).Return(response, nil)

		httpClient.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
			return req.URL.Host == "localhost:9410" && req.URL.Path == "/metrics"
		})).Times(1).Return(nil, errors.New("network error"))

		httpClient.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
			return req.URL.Host == "localhost:9420" && req.URL.Path == "/metrics"
		})).Times(1).Return(response1, nil)

		metricsFamily, err := scraper.Scrape(
			context.Background(),
			[]string{
				"http://localhost:9400/metrics",
				"http://localhost:9410/metrics",
				"http://localhost:9420/metrics",
			})

		r := require.New(t)
		r.NoError(err)
		r.NotNil(metricsFamily)
		r.Len(metricsFamily, 2)
		r.NotEmpty(metricsFamily[0])
		r.NotEmpty(metricsFamily[1])
	})
}
