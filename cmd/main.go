package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"

	"github.com/fedragon/sinkhole/internal/config"
	"github.com/fedragon/sinkhole/internal/dns"
	"github.com/fedragon/sinkhole/internal/udp"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	var cfg config.Config
	if err := envconfig.Process("", &cfg); err != nil {
		logger.Error("unable to parse configuration", "error", err)
		os.Exit(1)
	}

	sinkhole := dns.NewSinkhole(logger)
	fallback, err := udp.NewClient(cfg.FallbackAddr.String())
	if err != nil {
		logger.Error("unable to connect to fallback DNS", "address", cfg.FallbackAddr.String(), "error", err)
		os.Exit(1)
	}
	defer fallback.Close()

	udpServer := udp.NewServer(sinkhole, fallback, logger)
	if err := udpServer.Serve(ctx, cfg.Addr.String()); err != nil {
		logger.Error("unable to serve UDP", "address", cfg.Addr.String(), "error", err)
		os.Exit(1)
	}
}
