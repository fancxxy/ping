package ping

import (
	"net"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

func (p *Ping) receive(proto int, conn *icmp.PacketConn) {
	buf := make([]byte, 1500)
	defer p.wg.Done()

	for {
		// conn.IPv4PacketConn().ReadFrom(buf)
		// conn.IPv6PacketConn().ReadFrom(buf)
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); !ok || !netErr.Temporary() {
				break
			}
		}

		msg, err := icmp.ParseMessage(proto, buf[:n])
		if err != nil {
			continue
		}

		switch msg.Type {
		case ipv4.ICMPTypeEchoReply, ipv6.ICMPTypeEchoReply:
			echo, ok := msg.Body.(*icmp.Echo)
			if !ok || echo == nil {
				continue
			}

			if echo.ID != p.id {
				continue
			}

			req := p.delReq(echo.Seq)
			if req != nil {
				req.close()
			}

		default:
		}
	}
}
