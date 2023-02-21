package server

import (
	"net"

	"github.com/NanoRed/natunnel/internal/connection"
	"github.com/NanoRed/natunnel/internal/handler"
	"github.com/NanoRed/natunnel/internal/logger"
)

type Server struct {
	Addr    string
	Handler *handler.SrvHandler
}

func New(addr string, h *handler.SrvHandler) *Server {
	return &Server{
		Addr:    addr,
		Handler: h,
	}
}

func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	return s.Serve(ln)
}

func (s *Server) Serve(l net.Listener) error {
	defer l.Close()
	logger.Info("Server now is in progress :)")
	for {
		conn, err := l.Accept()
		if err != nil {
			logger.Error("Accept error: %v", err)
			continue
		}
		go s.Handler.Handle(connection.New(conn))
	}
}
