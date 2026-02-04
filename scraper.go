package domainreceiver

import (
	"context"
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

// Probably not the best practice
type capturedFields captureFields

func whoisLookup(domain string, timeout uint64, fields captureFields) capturedFields {

	return capturedFields{
		Expiration:  "",
		LastChanged: "",
	}
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

					expiryTime, err := time.Parse(time.RFC3339, event.Date)
					if err != nil {
						ds.settings.Logger.Error(
							"Failed to parse expiry date",
							zap.String("domain", d.Name),
							zap.String("date", event.Date),
						)
						continue
					}

					mu.Lock()
					ds.mb.RecordDomainExpiryTimeDataPoint(now, int64(expiryTime.Unix()), d.Name)
					mu.Unlock()
				}
			}
		}(domain)
	}

	wg.Wait()

	return ds.mb.Emit(), nil
}
