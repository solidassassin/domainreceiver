package domainreceiver

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/scraper"
	"go.opentelemetry.io/collector/scraper/scraperhelper"

	"github.com/solidassassin/domainreceiver/internal/metadata"
)

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		metadata.Type,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, metadata.MetricsStability),
	)
}

func createDefaultConfig() component.Config {
	cfg := scraperhelper.NewDefaultControllerConfig()
	cfg.CollectionInterval = 15 * time.Minute

	return &Config{
		ControllerConfig:     cfg,
		MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
		Domains:              []*domainConfig{},
		RateLimit:            1,
	}
}

func createMetricsReceiver(_ context.Context, params receiver.Settings, baseConf component.Config, consumer consumer.Metrics) (receiver.Metrics, error) {
	cfg := baseConf.(*Config)

	domainScraper := newScraper(cfg, params)
	s, err := scraper.NewMetrics(domainScraper.scrape, scraper.WithStart(domainScraper.start))
	if err != nil {
		return nil, err
	}

	return scraperhelper.NewMetricsController(&cfg.ControllerConfig, params, consumer, scraperhelper.AddMetricsScraper(metadata.Type, s))
}
