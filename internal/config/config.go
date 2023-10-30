package config

type Config struct {
	DnsServerAddr string `envconfig:"DNS_SERVER_ADDR" default:"0.0.0.0:1153"`
	FallbackAddr  string `envconfig:"DNS_FALLBACK_ADDR" default:"1.1.1.1:53"`
	HostsPath     string `envconfig:"HOSTS_PATH" default:"./hosts"`

	// HTTP server config: it will only be started if either DebugEndpointEnabled or MetricsEnabled is true
	HttpServerAddr       string `envconfig:"HTTP_SERVER_ADDR" default:"0.0.0.0:8000"`
	DebugEndpointEnabled bool   `envconfig:"DEBUG_ENDPOINT_ENABLED" default:"false"`
	MetricsEnabled       bool   `envconfig:"METRICS_ENABLED" default:"true"`
}
