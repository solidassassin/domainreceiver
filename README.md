# OpenTelemetry Domain Receiver

An OpenTelemetry receiver for checking the expiration date of the provided domains.

> [!WARNING]
> This receiver is in early stages of development. No proper testing or validation has been done yet. If you decide to use this regardless and encounter an issue then feel free to raise an issue.

## Compilation

To compile an OpenTelemetry Collector binary with this receiver it's recommended to use OpenTelemetry builder. Here's an example builder configuration:

```yaml
dist:
  name: custom-otelcol
  description: OpenTelemetry Collector with the domain receiver
  output_path: ./dist
  version: "0.1.0"
  otelcol_version: "0.144.0"

receivers:
  - gomod: github.com/solidassassin/domainreceiver <version_here>

processors:
  - gomod: go.opentelemetry.io/collector/processor/batchprocessor v0.144.0

exporters:
  - gomod: go.opentelemetry.io/collector/exporter/otlpexporter v0.144.0
  - gomod: go.opentelemetry.io/collector/exporter/debugexporter v0.144.0

providers:
  - gomod: go.opentelemetry.io/collector/confmap/provider/envprovider v1.50.0
  - gomod: go.opentelemetry.io/collector/confmap/provider/fileprovider v1.50.0
  - gomod: go.opentelemetry.io/collector/confmap/provider/httpprovider v1.50.0
  - gomod: go.opentelemetry.io/collector/confmap/provider/httpsprovider v1.50.0
  - gomod: go.opentelemetry.io/collector/confmap/provider/yamlprovider v1.50.0

```

## Configuration

The following options are available:

- `domains` (required): A list of domain configurations. Each item has the following options:
    - `name` (required): The domain name (like `google.com`).
    - `protocol` (optional, default = `rdap`): The protocol to use for fetching the expiration date. Currently only RDAP is supported (WhoIs support is in progress).
- `collection_interval` (optional, default = `15m`): The collection interval. If you need to monitor a lot of domains it's recommended to increase this value to avoid rate limits (`15 minutes` is already overkill for such a metric).

> [!NOTE]
> This receiver also exposes HTTP Client options (`confighttp.ClientConfig`) on the top level. These options can be found in the `confighttp` package [documentation](https://pkg.go.dev/go.opentelemetry.io/collector/config/confighttp#readme-client-configuration).

## Example Configuration

```yaml
receivers:
  domain:
    domains:
      - name: example.com
      - name: google.com
        protocol: rdap
    collection_internal: 1h

exporters:
  debug:
    verbosity: detailed

service:
  pipelines:
    metrics:
      receivers: [domain]
      exporters: [debug]
```

## Metrics

Details about the produced metrics can be found in [documentation.md](./documentation.md)
