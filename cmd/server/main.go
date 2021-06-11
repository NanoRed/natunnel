package main

import (
	"flag"

	"github.com/RedAFD/natunnel/internal/config"
	"github.com/RedAFD/natunnel/internal/handler"
	"github.com/RedAFD/natunnel/internal/logger"
	"github.com/RedAFD/natunnel/internal/server"
)

func init() {
	flag.StringVar(&config.HostDomain,
		"HostDomain",
		"",
		"Input Your Domain.",
	)
	flag.StringVar(&config.ServerListeningAddr,
		"ServerAddr",
		"0.0.0.0:80",
		"Input the listening address of the natunnel server.",
	)
	flag.StringVar(&config.HTTPParserAddr,
		"HTTPParserAddr",
		"127.0.0.1:7714",
		"Input the listening address of the http parser service.",
	)
}

func main() {
	flag.Parse()
	if config.HostDomain == "" {
		logger.Fatal("Please input your domain")
	}
	handler := handler.NewSrvHandler(config.CtrlConnKeepAlive, config.HTTPParserAddr)
	server := server.New(config.ServerListeningAddr, handler)
	err := server.ListenAndServe()
	if err != nil {
		logger.Error("Server running error: %v", err)
	}
}
