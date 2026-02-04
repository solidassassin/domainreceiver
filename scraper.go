package domainreceiver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/openrdap/rdap"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"

	"github.com/solidassassin/domainreceiver/internal/metadata"
)

// In some cases we can encounter different time formats
func parseExpiryTime(date string) (*time.Time, error) {
	var err error
	timeFormats := []string{
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, timeFormat := range timeFormats {
		expiryTime, err := time.Parse(timeFormat, date)
		if err == nil {
			return &expiryTime, nil
		}
	}
	return nil, fmt.Errorf("unable to parse `%s`, last parsing error: %w", date, err)
}

type domainScraper struct {
	rdapClient *rdap.Client
	cfg        *Config
	settings   component.TelemetrySettings
	mb         *metadata.MetricsBuilder
}

func newScraper(cfg *Config, settings receiver.Settings) *domainScraper {
	return &domainScraper{
		cfg:      cfg,
		settings: settings.TelemetrySettings,
		mb:       metadata.NewMetricsBuilder(cfg.MetricsBuilderConfig, settings),
	}
}

func (ds *domainScraper) start(ctx context.Context, host component.Host) error {
	httpClient, err := ds.cfg.ClientConfig.ToClient(ctx, host.GetExtensions(), ds.settings)
	if err != nil {
		return fmt.Errorf("failed to create the HTTP client: %w", err)
	}

	ds.rdapClient = &rdap.Client{
		HTTP: httpClient,
	}

	return nil
}

func (ds *domainScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	now := pcommon.NewTimestampFromTime(time.Now())
	wg.Add(len(ds.cfg.Domains))

	for _, domain := range ds.cfg.Domains {
		go func(d *domainConfig) {
			defer wg.Done()

			domainData, err := ds.rdapClient.QueryDomain(d.Name)
			if err != nil {
				ds.settings.Logger.Error(
					"Failed to fetch RDAP data",
					zap.String("domain", d.Name),
					zap.Error(err),
				)
				return
			}

			for _, event := range domainData.Events {
				if event.Action == "expiration" {

					expiryTime, err := parseExpiryTime(event.Date)
					if err != nil {
						ds.settings.Logger.Error(
							"Failed to parse expiry date",
							zap.String("domain", d.Name),
							zap.String("date", event.Date),
						)
						return
					}

					mu.Lock()
					ds.mb.RecordDomainExpiryTimeDataPoint(now, int64(expiryTime.Unix()), d.Name)
					mu.Unlock()

					return
				}
			}
		}(domain)
	}

	wg.Wait()
	return ds.mb.Emit(), nil
}
