package connection

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/NanoRed/natunnel/internal/config"
)

type Connection struct {
	OK        bool
	Protected bool
	Conn      net.Conn
	Mu        sync.Mutex
	Payload   interface{}
}

func New(c net.Conn) *Connection {
	return &Connection{
		OK:   true,
		Conn: c,
	}
}

func (c *Connection) Write(b []byte) (n int, err error) {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	if c.Conn == nil {
		err = errors.New("Conn is nil")
		return
	}
	if err = c.Conn.SetWriteDeadline(time.Now().Add(config.CtrlConnWritingTimeout)); err != nil {
		return
	}
	return c.Conn.Write(b)
}

func (c *Connection) Close() {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	if c.OK {
		if c.Protected {
			c.Protected = false
		} else {
			c.OK = false
			c.Conn.Close()
		}
	}
}

func (c *Connection) Alive() bool {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	return c.OK
}

func (c *Connection) ImmunityNextClose() {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	c.Protected = true
}

func (c *Connection) Store(data interface{}) {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	c.Payload = data
}

func (c *Connection) Load() interface{} {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	return c.Payload
}
