package domainreceiver

import (
	"errors"
	"fmt"

	"github.com/solidassassin/domainreceiver/internal/metadata"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/scraper/scraperhelper"
)

type Protocol string

const (
	ProtocolEmpty Protocol = ""
	ProtocolRDAP  Protocol = "rdap"
	ProtocolWhoIs Protocol = "whois"
)

type Config struct {
	confighttp.ClientConfig        `mapstructure:",squash"`
	scraperhelper.ControllerConfig `mapstructure:",squash"`
	metadata.MetricsBuilderConfig  `mapstructure:",squash"`
	Domains                        []*domainConfig `mapstructure:"domains"`
	// RateLimit is the maximum number of RDAP/WHOIS queries per second.
	// Set to 0 to disable rate limiting.
	RateLimit float64 `mapstructure:"rate_limit"`
}

type domainConfig struct {
	Name          string         `mapstructure:"name"`
	Protocol      string         `mapstructure:"protocol"`
	CaptureFields *captureFields `mapstructure:"capture_fields"`
}

type captureFields struct {
	Expiration  string `mapstructure:"expiration"`
	LastChanged string `mapstructure:"last_changed"`
}

// Domain name validation is quite hard to do, so we do not validate it and just inform about the
// validity when making a request to RDAP or WHOIS
func (cfg *domainConfig) Validate() error {
	protocol := Protocol(cfg.Protocol)

	switch protocol {
	case ProtocolEmpty, ProtocolRDAP:
		return nil
	case ProtocolWhoIs:
		return errors.New("the WhoIs protocol support is currently in progress.")
		//if cfg.CaptureFields == nil {
		//	return fmt.Errorf("capture fields must be defined when the %s protocol is used.", protocol)
		//}
	default:
		return fmt.Errorf("invalid protocol was provided: %s", protocol)
	}
}

func (cfg *Config) Validate() error {
	var combinedErr error

	if len(cfg.Domains) == 0 {
		combinedErr = errors.Join(errors.New("no domains configured"))
	}

	for _, domain := range cfg.Domains {
		combinedErr = errors.Join(combinedErr, domain.Validate())
	}

	return combinedErr
}
