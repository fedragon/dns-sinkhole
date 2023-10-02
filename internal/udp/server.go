package udp

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/fedragon/sinkhole/internal/dns"
	"github.com/fedragon/sinkhole/internal/dns/message"
)

const (
	maxDnsPacketSize = 512
)

type Server struct {
	sinkhole *dns.Sinkhole
	fallback io.ReadWriteCloser
}

func NewServer(sinkhole *dns.Sinkhole, fallback io.ReadWriteCloser) *Server {
	return &Server{
		sinkhole: sinkhole,
		fallback: fallback,
	}
}

func (s *Server) Serve(ctx context.Context, port string) error {
	udpAddr, err := net.ResolveUDPAddr("udp4", port)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			buffer := make([]byte, maxDnsPacketSize)
			_, addr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println(err)
				return err
			}

			query, err := message.ParseQuery(buffer)
			if err != nil {
				fmt.Println(err)
				continue
			}

			response, handled := s.sinkhole.Handle(query)
			if handled {
				data, err := response.Marshal()
				if err != nil {
					return err
				}

				if _, err := conn.WriteToUDP(data, addr); err != nil {
					return err
				}
				continue
			}

			if err := s.respondWithFallback(conn, addr, buffer); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (s *Server) respondWithFallback(conn *net.UDPConn, addr *net.UDPAddr, buffer []byte) error {
	if _, err := s.fallback.Write(buffer); err != nil {
		return err
	}

	response := make([]byte, maxDnsPacketSize)
	_, err := s.fallback.Read(response)
	if err != nil {
		return err
	}

	if _, err := conn.WriteToUDP(response, addr); err != nil {
		return err
	}

	return nil
}
