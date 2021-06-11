package handler

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"time"
	"unsafe"

	"github.com/RedAFD/natunnel/internal/config"
	"github.com/RedAFD/natunnel/internal/connection"
	"github.com/RedAFD/natunnel/internal/logger"
)

const ConnSign string = "NATUN"

// packet type
const (
	Heartbeat byte = iota
	AcquireHost
	ReverseDial
)

type SrvHandler struct {
	KeepAlive      time.Duration
	HTTPParserAddr string
}

func NewSrvHandler(t time.Duration, parserAddr string) *SrvHandler {
	newHandler := &SrvHandler{
		KeepAlive:      t,
		HTTPParserAddr: parserAddr,
	}
	go http.ListenAndServe(parserAddr, newHandler)
	return newHandler
}

func (h *SrvHandler) Handle(c *connection.Connection) {
	defer c.Close()
	reader := bufio.NewReader(c.Conn)
	pkType, _ := reader.Peek(len(ConnSign))
	switch string(pkType) {
	case ConnSign:
		h.HandleCtlPkt(c, reader)
	default:
		h.HandleHTTPPkt(c, reader)
	}
}

func (h *SrvHandler) HandleCtlPkt(c *connection.Connection, r *bufio.Reader) {
	r.Discard(len(ConnSign))
	for {
		err := c.Conn.SetReadDeadline(time.Now().Add(h.KeepAlive))
		if err != nil {
			logger.Error("Set read deadline error: %v", err)
			return
		}
		t, err := r.ReadByte()
		if err != nil {
			logger.Error("Get packet type error: %v", err)
			return
		}
		switch t {
		case Heartbeat:
			continue
		case AcquireHost:
			go func() {
				host := h.GenerateHost(c)
				sizeb := make([]byte, 2)
				binary.BigEndian.PutUint16(sizeb, uint16(len(host)))

				packet := &bytes.Buffer{}
				packet.WriteByte(AcquireHost)
				packet.Write(sizeb)
				packet.WriteString(host)

				_, err := c.Write(packet.Bytes())
				if err != nil {
					logger.Error("Send packet error: %v", err)
					c.Close()
					return
				}
			}()
		case ReverseDial:
			err := c.Conn.SetReadDeadline(time.Time{})
			if err != nil {
				logger.Error("Reset read deadline error: %v", err)
				return
			}
			keyb := make([]byte, 8)
			_, err = io.ReadFull(r, keyb)
			if err != nil {
				logger.Error("Get key error: %v", err)
				return
			}
			key := binary.BigEndian.Uint64(keyb)
			sizeb := make([]byte, 2)
			_, err = io.ReadFull(r, sizeb)
			if err != nil {
				logger.Error("Get size error: %v", err)
				return
			}
			size := binary.BigEndian.Uint16(sizeb)
			laddrb := make([]byte, size)
			_, err = io.ReadFull(r, laddrb)
			if err != nil {
				logger.Error("Unexpected size packet: %v", err)
				return
			}
			ch, has := connection.DP.GetPriRevsChan(key)
			if !has {
				logger.Error("Can not find the reverse connection channel")
				return
			}
			c.Store(laddrb)
			select {
			case ch <- c:
				c.ImmunityNextClose()
				return
			case <-time.After(config.CtrlConnMountReverseConnectTimeout):
				logger.Error("Mount reverse connection timeout")
				return
			}
		default:
			logger.Error("Unknown packet type")
			return
		}
	}
}

func (h *SrvHandler) GenerateHost(c *connection.Connection) string {
	dict := "abcdefghijklmnopqrstuvwxyz0123456789"
	len := len(dict)
AGAIN:
	var randb [6]byte
	for i := 0; i < 6; i++ {
		rand.Seed(time.Now().UnixNano())
		randb[i] = dict[rand.Intn(len)]
	}
	newhost := fmt.Sprintf("%s.%s", randb, config.HostDomain)
	if connection.DP.RegPriCtrlConn(newhost, c) {
		return newhost
	} else {
		goto AGAIN
	}
}

func (h *SrvHandler) HandleHTTPPkt(c *connection.Connection, r *bufio.Reader) {
	conn, err := net.Dial("tcp", h.HTTPParserAddr)
	if err != nil {
		logger.Error("Dial http parser error: %v", err)
		return
	}
	defer conn.Close()
	go io.Copy(conn, r)
	io.Copy(c, conn)
}

func (h *SrvHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := uint64(uintptr(unsafe.Pointer(r)))
	ch := connection.DP.NewPriRevsChan(key)
	defer connection.DP.RemPriRevsChan(key)
	cconn, has := connection.DP.GetPriCtrlConn(r.Host)
	if !has || !cconn.Alive() {
		logger.Error("Failed to get controlling tunnel")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	keyb := make([]byte, 8)
	binary.BigEndian.PutUint64(keyb, key)
	packet := &bytes.Buffer{}
	packet.WriteByte(ReverseDial)
	packet.Write(keyb)
	_, err := cconn.Write(packet.Bytes())
	if err != nil {
		logger.Error("Failed to get reverse dialing: %v", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	var rconn *connection.Connection
	select {
	case rconn = <-ch:
		if !rconn.Alive() {
			logger.Error("Broken reverse connection")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
	case <-time.After(config.HTTPServeTimeout):
		logger.Error("Reverse dial timeout")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	url := fmt.Sprintf("http://%s%s", bytes.TrimRight(rconn.Load().([]byte), ":80"), r.RequestURI)
	newreq, err := http.NewRequest(r.Method, url, r.Body)
	if err != nil {
		logger.Error("New request error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for hk, hv := range r.Header {
		for _, v := range hv {
			newreq.Header.Add(hk, v)
		}
	}
	resp, err := (&http.Client{Transport: &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return rconn.Conn, nil
		},
	}}).Do(newreq)
	if err != nil {
		logger.Error("Roundtrip error: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	for hk, hv := range resp.Header {
		for _, v := range hv {
			w.Header().Add(hk, v)
		}
	}
	body, _ := ioutil.ReadAll(resp.Body)
	w.Write(body)
}

type CliHandler struct {
	SendingQueue    chan []byte
	ReceivingQueue  chan []byte
	LocalServerAddr string
}

func NewCliHandler(laddr string) *CliHandler {
	return &CliHandler{
		SendingQueue:    make(chan []byte, 0),
		ReceivingQueue:  make(chan []byte, 0),
		LocalServerAddr: laddr,
	}
}

func (h *CliHandler) Handle(c *connection.Connection) {
	// receiving loop
	go func() {
		defer c.Close()
		r := bufio.NewReader(c.Conn)
		for {
			err := c.Conn.SetReadDeadline(time.Time{})
			if err != nil {
				logger.Error("Set read deadline error: %v", err)
				return
			}
			t, err := r.ReadByte()
			if err != nil {
				logger.Error("Get packet type error: %v", err)
				return
			}
			switch t {
			case AcquireHost:
				sizeb := make([]byte, 2)
				_, err = io.ReadFull(r, sizeb)
				if err != nil {
					logger.Error("Get size error: %v", err)
					return
				}
				size := binary.BigEndian.Uint16(sizeb)
				body := make([]byte, size)
				_, err = io.ReadFull(r, body)
				if err != nil {
					logger.Error("Unexpected size packet: %v", err)
					return
				}
				go func() {
					h.ReceivingQueue <- body
				}()
			case ReverseDial:
				key := make([]byte, 8)
				_, err = io.ReadFull(r, key)
				if err != nil {
					logger.Error("Get key error: %v", err)
					return
				}
				go func() {
					rc, err := net.Dial("tcp", c.Conn.RemoteAddr().String())
					if err != nil {
						logger.Error("Failed to dial server: %v", err)
						return
					}
					rconn := connection.New(rc)
					defer rconn.Close()

					sizeb := make([]byte, 2)
					binary.BigEndian.PutUint16(sizeb, uint16(len(h.LocalServerAddr)))
					packet := &bytes.Buffer{}
					packet.WriteString(ConnSign)
					packet.WriteByte(ReverseDial)
					packet.Write(key)
					packet.Write(sizeb)
					packet.WriteString(h.LocalServerAddr)
					_, err = rconn.Write(packet.Bytes())
					if err != nil {
						logger.Error("Send identity error: %v", err)
						return
					}

					lc, err := net.Dial("tcp", h.LocalServerAddr)
					if err != nil {
						logger.Error("Failed to dial local server: %v", err)
						return
					}
					defer lc.Close()

					go io.Copy(lc, rc)
					io.Copy(rc, lc)
				}()
			default:
				logger.Error("Unknown packet type")
				return
			}
		}
	}()
	// sending loop
	defer c.Close()
	packet := &bytes.Buffer{}
	packet.WriteString(ConnSign)
	_, err := c.Write(packet.Bytes())
	if err != nil {
		logger.Error("Send type error: %v", err)
		return
	}
	for {
		select {
		case data := <-h.SendingQueue:
			_, err := c.Write(data)
			if err != nil {
				logger.Error("Send packet error: %v", err)
				return
			}
		case <-time.After(config.CtrlConnHeartbeatInterval):
			packet := &bytes.Buffer{}
			packet.WriteByte(Heartbeat)
			_, err := c.Write(packet.Bytes())
			if err != nil {
				logger.Error("Send heartbeat error: %v", err)
				return
			}
		}
	}
}

func (h *CliHandler) MakeAcquireHostPacket() []byte {
	packet := &bytes.Buffer{}
	packet.WriteByte(AcquireHost)
	return packet.Bytes()
}
