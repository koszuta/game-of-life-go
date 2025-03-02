package life

import (
	"math"
	"time"
)

type Ticker interface {
	// Ticked updates the internal state and returns whether a tick happened
	Ticked() bool
	// Progress returns how much of the tick is complete in the range [0,1]
	Progress() float64
}

type Tick struct {
	// deltaTime is the number of nanoseconds per tick, i.e. 1,000,000,000/TPS
	deltaTime time.Duration
	// accumulator is the number of nanoseconds which have elapsed since the last tick.
	// It's guaranteed to be non-negative
	accumulator time.Duration
	previous    time.Time
}

func NewTick(tps float64) *Tick {
	if tps <= 0 {
		panic("ticks per second must be positive")
	}
	if tps > float64(time.Second.Nanoseconds()) {
		panic("less than one tick per nanosecond")
	}
	return &Tick{
		deltaTime:   time.Duration(float64(time.Second.Nanoseconds()) / tps),
		accumulator: 0,
	}
}

func (t *Tick) Ticked() bool {
	diff := time.Since(t.previous)
	if t.accumulator+diff >= t.deltaTime {
		t.accumulator += diff - t.deltaTime
		t.previous = time.Now()
		return true
	}
	return false
}

func (t *Tick) Progress() float64 {
	// Limit progress to 100%; accumulator will always be >= 0 so no need to clamp the lower end
	return math.Min(float64(t.accumulator)/float64(t.deltaTime), 1.0)
}
