.PHONY: server-linux
server-linux:
	GOOS=linux go build github.com/NanoRed/natunnel/cmd/server

.PHONY: client-windows
client-windows:
	GOOS=windows go build github.com/NanoRed/natunnel/cmd/client

.PHONY: client-darwin
client-darwin:
	GOOS=darwin go build github.com/NanoRed/natunnel/cmd/client