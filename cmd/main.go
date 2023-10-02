package main

import (
	"context"
	"log"

	"github.com/kelseyhightower/envconfig"

	"github.com/fedragon/sinkhole/internal/config"
	"github.com/fedragon/sinkhole/internal/dns"
	"github.com/fedragon/sinkhole/internal/udp"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var cfg config.Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}

	hole := dns.NewSinkhole()

	fallback, err := udp.NewClient(cfg.FallbackAddr.String())
	if err != nil {
		log.Fatal(err)
	}
	udpServer := udp.NewServer(hole, fallback)
	if err := udpServer.Serve(ctx, cfg.Addr.String()); err != nil {
		log.Fatal(err)
	}
}
