package main

import (
	"fmt"

	"github.com/RedAFD/natunnel/internal/client"
	"github.com/RedAFD/natunnel/internal/config"
	"github.com/RedAFD/natunnel/internal/handler"
)

func main() {
STARTOVER:
	fmt.Printf("Please enter your natunnel server address(e.g. 40.100.70.1:80):\n")
	fmt.Scanln(&config.ServerPublicAddr)
	fmt.Printf("Please enter your local server address that need to be exposed to the internet(e.g. 127.0.0.1:8080):\n")
	fmt.Scanln(&config.LocalTargetAddr)

	sig := make(chan struct{}, 0)

	h := handler.NewCliHandler(config.LocalTargetAddr)
	cli := client.New(config.ServerPublicAddr, h)
	go func() {
		cli.DialAndServe()
		sig <- struct{}{}
	}()

	acquireHostPacket := h.MakeAcquireHostPacket()

	select {
	case h.SendingQueue <- acquireHostPacket:
		select {
		case host := <-h.ReceivingQueue:
			fmt.Printf("Successfully running, your public host is http://%s. Enjoy :)\n", host)
			<-sig
			fmt.Printf("An error occurred. Please re-enter valid information!\n\n")
			goto STARTOVER
		case <-sig:
			fmt.Printf("An error occurred. Please re-enter valid information!\n\n")
			goto STARTOVER
		}
	case <-sig:
		fmt.Printf("An error occurred. Please re-enter valid information!\n\n")
		goto STARTOVER
	}
}
