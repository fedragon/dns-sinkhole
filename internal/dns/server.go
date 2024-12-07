package dns

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"time"

	p "github.com/prometheus/client_golang/prometheus"

	"github.com/fedragon/sinkhole/audit"
	"github.com/fedragon/sinkhole/internal/dns/message"
	"github.com/fedragon/sinkhole/internal/metrics"
)

const (
	maxPacketSize = 512
)

type Server struct {
	sinkhole *Sinkhole
	upstream io.ReadWriteCloser
	logger   *slog.Logger
	audit    *audit.Logger
}

func NewServer(sinkhole *Sinkhole, upstream io.ReadWriteCloser, logger *slog.Logger, audit *audit.Logger) *Server {
	return &Server{
		sinkhole: sinkhole,
		upstream: upstream,
		logger:   logger.With("source", "dns_server"),
		audit:    audit,
	}
}

func (s *Server) Serve(ctx context.Context, address string) error {
	udpAddr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	s.logger.Debug("Starting UDP server", "address", address)

	if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Debug("Shutting down UDP server")
			return nil
		default:
			if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
				return err
			}

			rawQuery := make([]byte, maxPacketSize)
			_, addr, err := conn.ReadFromUDP(rawQuery)
			if err != nil {
				if !errors.Is(err, os.ErrDeadlineExceeded) {
					return err
				}
				continue
			}

			if err := s.process(rawQuery, conn, addr); err != nil {
				s.logger.Error("Error processing query", "error", err)
				continue
			}
		}
	}
}

func (s *Server) process(rawQuery []byte, conn *net.UDPConn, addr *net.UDPAddr) error {
	totalTimer := p.NewTimer(metrics.ResponseTimesTotal)
	defer totalTimer.ObserveDuration()

	query, err := message.UnmarshalQuery(rawQuery)
	if err != nil {
		metrics.QueryParsingErrors.Inc()
		return fmt.Errorf("unable to unmarshal query: %w, query: %v", err, rawQuery)
	}

	response, handled := s.sinkhole.Resolve(query)
	var rawResponse []byte
	if handled {
		metrics.BlockedQueries.Inc()

		rawResponse, err = message.MarshalResponse(response)
		if err != nil {
			metrics.ResponseMarshallingErrors.Inc()
			return fmt.Errorf("unable to marshal response: %w, response: %v", err, rawResponse)
		}
	} else {
		metrics.UpstreamQueries.Inc()

		rawResponse, err = s.queryUpstreamServer(rawQuery)
		if err != nil {
			metrics.UpstreamErrors.Inc()
			return fmt.Errorf("unable to query upstream DNS: %w", err)
		}

		s.audit.Log(query.ID, uint16(query.Question.Type), rawQuery, rawResponse)
	}

	writeTimer := p.NewTimer(metrics.ResponseTimesWriteResponse)
	defer writeTimer.ObserveDuration()
	if _, err := conn.WriteToUDP(rawResponse, addr); err != nil {
		metrics.WriteResponseErrors.Inc()
		return fmt.Errorf("unable to write response: %w", err)
	}

	return nil
}

func (s *Server) queryUpstreamServer(buffer []byte) ([]byte, error) {
	timer := p.NewTimer(metrics.ResponseTimesUpstreamResolve)
	defer timer.ObserveDuration()

	if _, err := s.upstream.Write(buffer); err != nil {
		return nil, err
	}

	response := make([]byte, maxPacketSize)
	if _, err := s.upstream.Read(response); err != nil {
		return nil, err
	}

	return response, nil
}
