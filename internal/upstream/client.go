package upstream

import (
	"io"
	"net"
	"time"
)

type Client struct {
	conn *net.UDPConn
}

func NewClient(addr string) (io.ReadWriteCloser, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		return nil, err
	}

	return &Client{conn: conn}, nil
}

func (c *Client) Write(data []byte) (int, error) {
	return c.conn.Write(data)
}

func (c *Client) Read(buffer []byte) (int, error) {
	if err := c.conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		return 0, err
	}

	n, _, err := c.conn.ReadFromUDP(buffer)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
