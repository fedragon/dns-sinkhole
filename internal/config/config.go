package config

type Config struct {
	Addr           string `envconfig:"SINKHOLE_ADDR" default:"0.0.0.0:1153"`
	FallbackAddr   string `envconfig:"FALLBACK_ADDR" default:"1.1.1.1:53"`
	HostsPath      string `envconfig:"HOSTS_PATH" default:"./hosts"`
	MetricsEnabled bool   `envconfig:"METRICS_ENABLED" default:"false"`
	MetricsAddr    string `envconfig:"METRICS_ADDR" default:"127.0.0.1:8000"`
}
