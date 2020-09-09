package ping

import (
	"testing"
	"time"
)

func TestPing(t *testing.T) {
	p, err := New("0.0.0.0", "")
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	dest := "www.baidu.com"
	timeout := time.Second * 2

	rtt, err := p.Ping(dest, timeout)
	if err != nil {
		t.Error(err)
	}
	t.Logf("ping %s, rtt: %.3fms\n", dest, float64(rtt/time.Microsecond)/1000.0)
}
