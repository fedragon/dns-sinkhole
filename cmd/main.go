package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"

	"github.com/fedragon/sinkhole/internal/blacklist"
	"github.com/fedragon/sinkhole/internal/config"
	"github.com/fedragon/sinkhole/internal/dns"
	"github.com/fedragon/sinkhole/internal/metrics"
	"github.com/fedragon/sinkhole/internal/udp"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	var cfg config.Config
	if err := envconfig.Process("", &cfg); err != nil {
		logger.Error("unable to parse configuration", "error", err)
		return
	}

	fallback, err := udp.NewClient(cfg.FallbackAddr)
	if err != nil {
		logger.Error("unable to connect to fallback DNS", "address", cfg.FallbackAddr, "error", err)
		return
	}
	defer fallback.Close()

	logger.Debug("Reading blacklisted domains from hosts file", "path", cfg.HostsPath)

	file, err := os.Open(cfg.HostsPath)
	if err != nil {
		logger.Error("unable to open hosts file", "path", cfg.HostsPath, "error", err)
		return
	}
	defer file.Close()

	sinkhole := dns.NewSinkhole(logger)

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var count int
	for domain := range blacklist.Parse(scanner) {
		if err := sinkhole.Register(domain); err != nil {
			logger.Error("unable to register domain", "domain", domain, "error", err)
			return
		}
		count++
	}

	group, gCtx := errgroup.WithContext(ctx)

	metrics.BlacklistedDomains.Set(float64(count))
	logger.Debug("Finished registering blacklisted domains", "count", count)

	if cfg.MetricsEnabled || cfg.DebugEndpointEnabled {
		httpHandler := http.ServeMux{}

		if cfg.DebugEndpointEnabled {
			httpHandler.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
				domain := r.URL.Query().Get("domain")
				contains := sinkhole.Contains(domain)
				_, _ = w.Write([]byte(fmt.Sprintf("%t", contains)))
			})
		}

		if cfg.MetricsEnabled {
			httpHandler.Handle("/metrics", promhttp.Handler())
		}

		httpServer := &http.Server{
			Addr:    cfg.HttpAddr,
			Handler: &httpHandler,
		}

		group.Go(func() error {
			logger.Debug("Starting metrics server", "address", cfg.HttpAddr)
			return httpServer.ListenAndServe()
		})
		group.Go(func() error {
			<-gCtx.Done()
			return httpServer.Shutdown(context.Background())
		})
	}

	group.Go(func() error {
		udpServer := udp.NewServer(sinkhole, fallback, logger)
		return udpServer.Serve(gCtx, cfg.Addr)
	})

	if err := group.Wait(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("fatal error", "error", err)
		}

		return
	}
}
