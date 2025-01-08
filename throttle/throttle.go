package throttle

import (
	"sync/atomic"
	"time"
)

type Throttle struct {
	last        time.Time
	interval    time.Duration
	max         uint64
	accumulated atomic.Uint64
}

func (t *Throttle) Doable() bool {
	elapsed := time.Now().Sub(t.last)
	if elapsed <= t.interval {
		return t.accumulated.Add(1) <= t.max
	}
	t.last = time.Now()
	t.accumulated.Store(1)
	return true
}

func NewThrottle(max uint64, interval time.Duration) *Throttle {
	return &Throttle{
		last:        time.Now(),
		interval:    interval,
		max:         max,
		accumulated: atomic.Uint64{},
	}
}
