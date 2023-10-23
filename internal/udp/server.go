package udp

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
)

const (
	maxDnsPacketSize = 512
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
		logger:   logger.With("source", "server"),
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

	s.logger.Debug("listening on address", "address", address)

	if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Debug("shutting down server", "address", address)
			return nil
		default:
			if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
				return err
			}

			in := make([]byte, maxDnsPacketSize)
			_, addr, err := conn.ReadFromUDP(in)
			if err != nil {
				if !errors.Is(err, os.ErrDeadlineExceeded) {
					return err
				}
				continue
			}

			query, err := message.ParseQuery(in)
			if err != nil {
				s.logger.Error("unable to parse query", "error", err)
				continue
			}

			s.logger.Debug("handling query", "query", query, "raw_query", in)

			response, handled := s.sinkhole.Handle(query)
			var out []byte
			if handled {
				s.logger.Debug("the query has been handled by the sinkhole", "query", query)

				out, err = response.Marshal()
				if err != nil {
					s.logger.Error("unable to marshal response", "error", err)
					continue
				}
			} else {
				out, err = s.queryFallbackDNS(in)
				if err != nil {
					s.logger.Error("unable to query fallback DNS", "error", err)
					continue
				}
				s.logger.Debug("the query has been handled by the fallback", "query", query)
			}

			s.logger.Debug("sending response", "response", response, "raw_response", out)

			if _, err := conn.WriteToUDP(out, addr); err != nil {
				return err
			}
		}
	}
}

func (s *Server) queryFallbackDNS(buffer []byte) ([]byte, error) {
	if _, err := s.fallback.Write(buffer); err != nil {
		return nil, err
	}

	response := make([]byte, maxDnsPacketSize)
	_, err := s.fallback.Read(response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
