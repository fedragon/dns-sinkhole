package dns

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"time"

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

			query, err := message.UnmarshalQuery(rawQuery)
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
				metrics.UpstreamQueries.Inc()

				rawResponse, err = s.queryUpstreamServer(rawQuery)
				if err != nil {
					metrics.UpstreamErrors.Inc()
					s.logger.Error("Unable to query upstream DNS", "raw_query", rawQuery, "error", err)
					continue
				}

				s.audit.Log(query.ID(), uint16(query.Type()), rawQuery, rawResponse)
			}

			if _, err := conn.WriteToUDP(rawResponse, addr); err != nil {
				metrics.WriteResponseErrors.Inc()
				return err
			}
		}
	}
}

func (s *Server) queryUpstreamServer(buffer []byte) ([]byte, error) {
	if _, err := s.upstream.Write(buffer); err != nil {
		return nil, err
	}

	response := make([]byte, maxPacketSize)
	if _, err := s.upstream.Read(response); err != nil {
		return nil, err
	}

	return response, nil
}
