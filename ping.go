package ping

import (
	"errors"
	"os"
	"sync"

	"golang.org/x/net/icmp"
)

const (
	ProtocolICMP     = 1
	ProtocolIPv6ICMP = 58
)

type Ping struct {
	id       int
	payload  []byte
	requests map[int]*request
	reqLock  sync.Mutex
	conn4    *icmp.PacketConn
	lock4    sync.Mutex
	conn6    *icmp.PacketConn
	lock6    sync.Mutex
	wg       sync.WaitGroup
}

func (p *Ping) Close() {
	if p.conn4 != nil {
		p.conn4.Close()
	}

	if p.conn6 != nil {
		p.conn6.Close()
	}

	p.wg.Wait()
}

func (p *Ping) addReq(seq int, req *request) {
	p.reqLock.Lock()
	defer p.reqLock.Unlock()

	p.requests[seq] = req
}

func (p *Ping) delReq(seq int) *request {
	p.reqLock.Lock()
	defer p.reqLock.Unlock()

	req := p.requests[seq]
	delete(p.requests, seq)
	return req
}

func New(ip4, ip6 string) (*Ping, error) {
	ping := &Ping{
		id:       os.Getpid(),
		requests: make(map[int]*request),
		payload:  make([]byte, 56, 56),
	}

	var err error
	if ip4 != "" {
		ping.conn4, err = icmp.ListenPacket("ip4:icmp", ip4)
		if err != nil {
			return nil, err
		}
		ping.wg.Add(1)
		go ping.receive(ProtocolICMP, ping.conn4)
	}

	if ip6 != "" {
		ping.conn6, err = icmp.ListenPacket("ip6:ipv6-icmp", ip6)
		if err != nil {
			if ping.conn4 != nil {
				ping.conn4.Close()
			}
			return nil, err
		}
		ping.wg.Add(1)
		go ping.receive(ProtocolIPv6ICMP, ping.conn6)
	}

	if ping.conn4 == nil && ping.conn6 == nil {
		return nil, errors.New("no listen")
	}

	return ping, nil
}
