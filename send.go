package ping

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

var sequence uint32

func (p *Ping) Ping(ip string, timeout time.Duration) (time.Duration, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(timeout))
	defer cancel()

	dest, err := net.ResolveIPAddr("ip", ip)
	if err != nil {
		return 0, err
	}

	var (
		seq = int(atomic.AddUint32(&sequence, 1))
		msg = icmp.Message{
			Code: 0,
			Body: &icmp.Echo{
				ID:   p.id,
				Seq:  seq,
				Data: p.payload,
			},
		}
		conn *icmp.PacketConn
		lock *sync.Mutex
	)

	if dest.IP.To4() != nil {
		msg.Type = ipv4.ICMPTypeEcho
		conn = p.conn4
		lock = &p.lock4
	} else {
		msg.Type = ipv6.ICMPTypeEchoRequest
		conn = p.conn6
		lock = &p.lock6
	}

	bs, err := msg.Marshal(nil)
	if err != nil {
		return 0, err
	}

	req := newRequest()
	p.addReq(seq, req)

	lock.Lock()
	_, err = conn.WriteTo(bs, dest)
	lock.Unlock()

	if err != nil {
		req.close()
		p.delReq(seq)
		return 0, err
	}

	select {
	case <-req.wait:
	case <-ctx.Done():
		p.delReq(seq)
		return 0, errors.New("wait echo reply timeout")
	}

	return req.rtt()
}
