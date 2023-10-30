package dns

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/fedragon/sinkhole/internal/dns/message"
	"github.com/fedragon/sinkhole/internal/metrics"
)

const (
	maxPacketSize = 512
)

type Server struct {
	sinkhole *Sinkhole
	fallback io.ReadWriteCloser
	logger   *slog.Logger
}

func NewServer(sinkhole *Sinkhole, fallback io.ReadWriteCloser, logger *slog.Logger) *Server {
	return &Server{
		sinkhole: sinkhole,
		fallback: fallback,
		logger:   logger.With("source", "dns_server"),
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
			if handled == ResolveSuccess {
				metrics.BlockedQueries.Inc()

				rawResponse, err = response.Marshal()
				if err != nil {
					metrics.ResponseMarshallingErrors.Inc()
					s.logger.Error("Unable to marshal response", "response", response, "error", err)
					continue
				}
			} else {
				metrics.FallbackQueries.Inc()

				switch handled {
				case UnresolvedNonStandard:
					metrics.NonStandardQueries.Inc()
					s.logger.Debug("Passing non-standard query to fallback DNS resolver", "query", rawQuery)
				case UnresolvedNonRecursive:
					metrics.NonRecursiveQueries.Inc()
					s.logger.Debug("Passing non-recursive query to fallback DNS resolver", "query", rawQuery)
				case UnresolvedUnsupportedClass:
					metrics.UnsupportedClassQueries.Inc()
					s.logger.Debug("Passing unsupported class query to fallback DNS resolver", "query", rawQuery)
				case UnresolvedUnsupportedType:
					metrics.UnsupportedTypeQueries.Inc()
					s.logger.Debug("Passing unsupported type query to fallback DNS resolver", "query", rawQuery)
				case UnresolvedNotFound:
					// nothing to do
				}

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
	if _, err := s.fallback.Read(response); err != nil {
		return nil, err
	}

	return response, nil
}
