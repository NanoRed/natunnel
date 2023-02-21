package client

import (
	"net"

	"github.com/NanoRed/natunnel/internal/connection"
	"github.com/NanoRed/natunnel/internal/handler"
	"github.com/NanoRed/natunnel/internal/logger"
)

type Client struct {
	Addr    string
	Handler *handler.CliHandler
}

func New(addr string, h *handler.CliHandler) *Client {
	return &Client{
		Addr:    addr,
		Handler: h,
	}
}

func (c *Client) DialAndServe() {
	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		logger.Error("Dial error: %v", err)
		return
	}
	c.Handler.Handle(connection.New(conn))
}
