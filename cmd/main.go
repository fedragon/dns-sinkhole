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

	"github.com/fedragon/sinkhole/audit"
	"github.com/fedragon/sinkhole/internal/config"
	"github.com/fedragon/sinkhole/internal/dns"
	"github.com/fedragon/sinkhole/internal/hosts"
	"github.com/fedragon/sinkhole/internal/metrics"
	"github.com/fedragon/sinkhole/internal/upstream"
)

var Version = "dev"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	var cfg config.Config
	if err := envconfig.Process("", &cfg); err != nil {
		logger.Error("Unable to parse configuration", "error", err)
		return
	}

	auditLogger, err := audit.New(cfg.AuditLogEnabled)
	if err != nil {
		logger.Error("Unable to create audit logger", "error", err)
		return
	}
	defer auditLogger.Close()

	upstream, err := upstream.NewClient(cfg.UpstreamServerAddr)
	if err != nil {
		logger.Error("Unable to connect to upstream DNS resolver", "address", cfg.UpstreamServerAddr, "error", err)
		return
	}
	defer upstream.Close()

	logger.Debug("Reading non-routable domains from hosts file", "path", cfg.HostsPath)

	file, err := os.Open(cfg.HostsPath)
	if err != nil {
		logger.Error("Unable to open hosts file", "path", cfg.HostsPath, "error", err)
		return
	}
	defer file.Close()

	sinkhole := dns.NewSinkhole(logger)

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var count int
	for line := range hosts.Parse(scanner) {
		if line.Err != nil {
			logger.Error("Unable to parse hosts file", "error", line.Err)
			return
		}

		if err := sinkhole.Register(line.Domain); err != nil {
			logger.Error("Unable to register domain", "domain", line.Domain, "error", err)
			return
		}
		count++
	}

	metrics.NonRoutableDomains.Set(float64(count))
	logger.Debug("Finished registering non-routable domains", "count", count)

	group, gCtx := errgroup.WithContext(ctx)
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

		httpHandler.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(Version))
		})

		httpServer := &http.Server{
			Addr:    cfg.HttpServerAddr,
			Handler: &httpHandler,
		}

		group.Go(func() error {
			logger.Debug("Starting HTTP server", "address", cfg.HttpServerAddr)
			return httpServer.ListenAndServe()
		})
		group.Go(func() error {
			<-gCtx.Done()
			logger.Debug("Shutting down HTTP server")
			sctx, scancel := context.WithTimeout(context.Background(), cfg.HttpShutdownTimeout)
			defer scancel()
			return httpServer.Shutdown(sctx)
		})
	}

	group.Go(func() error {
		return dns.NewServer(sinkhole, upstream, logger, auditLogger).Serve(gCtx, cfg.LocalServerAddr)
	})

	if err := group.Wait(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Fatal error", "error", err)
		}

		return
	}
}
