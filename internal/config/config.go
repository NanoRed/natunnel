package config

import "time"

var (
	CtrlConnKeepAlive                  time.Duration = time.Second * 5
	CtrlConnCleanInterval              time.Duration = time.Minute
	CtrlConnWritingTimeout             time.Duration = time.Second * 2
	CtrlConnHeartbeatInterval          time.Duration = time.Second
	CtrlConnMountReverseConnectTimeout time.Duration = time.Second * 3

	HTTPParserAddr   string        = "127.0.0.1:7714"
	HTTPServeTimeout time.Duration = time.Second * 3

	HostDomain string = "example.com"

	IsDebugMode bool = true

	ServerListeningAddr string = "0.0.0.0:80"
	ServerPublicAddr    string = "example.com:80"
	LocalTargetAddr     string = "127.0.0.1:80"
)
