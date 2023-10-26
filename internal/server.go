package internal

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/fedragon/sinkhole/internal/dns"
	"github.com/fedragon/sinkhole/internal/dns/message"
	"github.com/fedragon/sinkhole/internal/metrics"
)

const (
	maxPacketSize = 512
)

type Server struct {
	sinkhole *dns.Sinkhole
	fallback io.ReadWriteCloser
	logger   *slog.Logger
}

func NewServer(sinkhole *dns.Sinkhole, fallback io.ReadWriteCloser, logger *slog.Logger) *Server {
	return &Server{
		sinkhole: sinkhole,
		fallback: fallback,
		logger:   logger.With("source", "udp_server"),
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

			query, err := message.ParseQuery(rawQuery)
			if err != nil {
				metrics.QueryParsingErrors.Inc()
				s.logger.Error("Unable to parse query", "error", err)
				continue
			}

			response, handled := s.sinkhole.Resolve(query)
			var rawResponse []byte
			if handled {
				metrics.BlockedQueries.Inc()

				rawResponse, err = response.Marshal()
				if err != nil {
					metrics.ResponseMarshallingErrors.Inc()
					s.logger.Error("Unable to marshal response", "response", response, "error", err)
					continue
				}
			} else {
				metrics.FallbackQueries.Inc()
				rawResponse, err = s.queryFallbackDNS(rawQuery)
				if err != nil {
					metrics.FallbackErrors.Inc()
					s.logger.Error("Unable to query fallback DNS", "raw_query", rawQuery, "error", err)
					continue
				}
			}

			if _, err := conn.WriteToUDP(rawResponse, addr); err != nil {
				return err
			}
		}
	}
}

func (s *Server) queryFallbackDNS(buffer []byte) ([]byte, error) {
	if _, err := s.fallback.Write(buffer); err != nil {
		return nil, err
	}

	response := make([]byte, maxPacketSize)
	_, err := s.fallback.Read(response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
