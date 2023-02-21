package connection

import (
	"sync"
	"time"

	"github.com/NanoRed/natunnel/internal/config"
)

var DP *Dispatcher

func init() {
	DP = &Dispatcher{
		PriCtrlConns: make(map[string]*Connection),
		PriRevsChans: make(map[uint64]chan *Connection),
	}
	// shrink the pool
	go func() {
		for range time.Tick(config.CtrlConnCleanInterval) {
			DP.PCCMutex.Lock()
			for key, val := range DP.PriCtrlConns {
				if !val.Alive() {
					delete(DP.PriCtrlConns, key)
				}
			}
			DP.PCCMutex.Unlock()
		}
	}()
}

type Dispatcher struct {
	PriCtrlConns map[string]*Connection
	PriRevsChans map[uint64]chan *Connection

	PCCMutex sync.RWMutex // PriCtrlConns Mutex
	PRCMutex sync.Mutex   // PriRevsConns Mutex
}

func (d *Dispatcher) RegPriCtrlConn(host string, c *Connection) (ok bool) {
	d.PCCMutex.Lock()
	defer d.PCCMutex.Unlock()
	if _, has := d.PriCtrlConns[host]; has {
		return false
	}
	d.PriCtrlConns[host] = c
	return true
}

func (d *Dispatcher) GetPriCtrlConn(host string) (c *Connection, has bool) {
	d.PCCMutex.RLock()
	defer d.PCCMutex.RUnlock()
	c, has = d.PriCtrlConns[host]
	return
}

func (d *Dispatcher) NewPriRevsChan(key uint64) chan *Connection {
	d.PRCMutex.Lock()
	defer d.PRCMutex.Unlock()
	ch := make(chan *Connection, 0)
	d.PriRevsChans[key] = ch
	return ch
}

func (d *Dispatcher) GetPriRevsChan(key uint64) (ch chan *Connection, has bool) {
	d.PRCMutex.Lock()
	defer d.PRCMutex.Unlock()
	ch, has = d.PriRevsChans[key]
	return
}

func (d *Dispatcher) RemPriRevsChan(key uint64) {
	d.PRCMutex.Lock()
	defer d.PRCMutex.Unlock()
	delete(d.PriRevsChans, key)
}
