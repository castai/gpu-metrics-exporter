package exporter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	maxConcurrentScrapes = 15
)

type MetricFamilyMap map[string]*dto.MetricFamily

type Scraper interface {
	Scrape(ctx context.Context, urls []string) ([]MetricFamiliyMap, error)
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type result struct {
	metricFamilyMap MetricFamiliyMap
	err             error
	ts              time.Time
}

type scraper struct {
	httpClient HTTPClient
	parser     expfmt.TextParser
	log        logrus.FieldLogger
}

func NewScraper(httpClient HTTPClient, log logrus.FieldLogger) Scraper {
	return &scraper{
		httpClient: httpClient,
		log:        log,
	}
}

func (s scraper) Scrape(ctx context.Context, urls []string) ([]MetricFamiliyMap, error) {
	var g errgroup.Group
	g.SetLimit(maxConcurrentScrapes)

	resultsChan := make(chan result, maxConcurrentScrapes)

	now := time.Now().UTC()

	for i := range urls {
		url := urls[i]
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				metrics, err := s.scrapeURL(ctx, url)
				if err != nil {
					err = fmt.Errorf("error while fetching metrics from '%s' %w", url, err)
				}
				resultsChan <- result{metricFamilyMap: metrics, err: err, ts: now}
			}
			return nil
		})
	}

	go func() {
		if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
			s.log.Errorf("error while scraping metrics %v", err)
		}
		close(resultsChan)
	}()

	metrics := make([]MetricFamiliyMap, 0, len(urls))
	for result := range resultsChan {
		if result.err != nil {
			s.log.Error(result.err)
			continue
		}
		metrics = append(metrics, result.metricFamilyMap)
	}

	return metrics, nil
}

func (s scraper) scrapeURL(ctx context.Context, url string) (map[string]*dto.MetricFamily, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxWithTimeout, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create request %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while making http request %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	metrics, err := s.parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot parse metrics %w", err)
	}

	return metrics, nil
}
