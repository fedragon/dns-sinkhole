package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"

	"github.com/fedragon/sinkhole/internal/config"
	"github.com/fedragon/sinkhole/internal/dns"
	"github.com/fedragon/sinkhole/internal/udp"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	var cfg config.Config
	if err := envconfig.Process("", &cfg); err != nil {
		logger.Error("unable to parse configuration", "error", err)
		os.Exit(1)
	}

	fallback, err := udp.NewClient(cfg.FallbackAddr)
	if err != nil {
		logger.Error("unable to connect to fallback DNS", "address", cfg.FallbackAddr, "error", err)
		os.Exit(1)
	}
	defer fallback.Close()

	udpServer := udp.NewServer(dns.NewSinkhole(logger), fallback, logger)
	if err := udpServer.Serve(ctx, cfg.Addr); err != nil {
		logger.Error("unable to serve UDP", "address", cfg.Addr, "error", err)
		os.Exit(1)
	}
}
