package config

import "net"

type Config struct {
	Addr         net.Addr `envconfig:"SINKHOLE_ADDR" default:"0.0.0.0:53"`
	FallbackAddr net.Addr `envconfig:"FALLBACK_ADDR" default:"8.8.8.8:53"`
}
