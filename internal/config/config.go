package config

import "time"

type Config struct {
	LocalServerAddr    string `envconfig:"LOCAL_SERVER_ADDR" default:"0.0.0.0:1153"`
	UpstreamServerAddr string `envconfig:"UPSTREAM_SERVER_ADDR" default:"1.1.1.1:53"`
	HostsPath          string `envconfig:"HOSTS_PATH" default:"./hosts"`

	// HTTP server config: it will only be started if either DebugEndpointEnabled or MetricsEnabled is true
	HttpServerAddr       string        `envconfig:"HTTP_SERVER_ADDR" default:"0.0.0.0:8000"`
	HttpShutdownTimeout  time.Duration `envconfig:"HTTP_SHUTDOWN_TIMEOUT" default:"5s"`
	DebugEndpointEnabled bool          `envconfig:"DEBUG_ENDPOINT_ENABLED" default:"false"`
	MetricsEnabled       bool          `envconfig:"METRICS_ENABLED" default:"false"`
}
