package ping

import "time"

type request struct {
	wait  chan struct{}
	start time.Time
	end   time.Time
}

func newRequest() *request {
	return &request{
		wait:  make(chan struct{}),
		start: time.Now(),
	}
}

func (r *request) close() {
	r.end = time.Now()
	close(r.wait)
}

func (r *request) rtt() (time.Duration, error) {
	return r.end.Sub(r.start), nil
}
