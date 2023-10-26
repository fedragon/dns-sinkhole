package config

type Config struct {
	Addr                 string `envconfig:"SINKHOLE_ADDR" default:"0.0.0.0:1153"`
	FallbackAddr         string `envconfig:"FALLBACK_ADDR" default:"1.1.1.1:53"`
	HostsPath            string `envconfig:"HOSTS_PATH" default:"./hosts"`
	DebugEndpointEnabled bool   `envconfig:"DEBUG_ENDPOINT_ENABLED" default:"true"`
	MetricsEnabled       bool   `envconfig:"METRICS_ENABLED" default:"true"`
	HttpAddr             string `envconfig:"HTTP_ADDR" default:"0.0.0.0:8000"`
}
