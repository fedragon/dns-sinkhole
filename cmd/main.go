package main

import (
	"bufio"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"

	"github.com/fedragon/sinkhole/internal/blacklist"
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

	sinkhole := dns.NewSinkhole(logger)

	logger.Debug("Reading blacklisted domains from hosts file", "path", cfg.HostsPath)

	file, err := os.Open(cfg.HostsPath)
	if err != nil {
		logger.Error("unable to open hosts file", "path", cfg.HostsPath, "error", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var count int
	for domain := range blacklist.Parse(scanner) {
		if err := sinkhole.Register(domain); err != nil {
			logger.Error("unable to register domain", "domain", domain, "error", err)
			os.Exit(1)
		}
		count++
	}

	logger.Debug("Finished registering blacklisted domains", "count", count)

	udpServer := udp.NewServer(sinkhole, fallback, logger)
	if err := udpServer.Serve(ctx, cfg.Addr); err != nil {
		logger.Error("unable to serve UDP", "address", cfg.Addr, "error", err)
		os.Exit(1)
	}
}
