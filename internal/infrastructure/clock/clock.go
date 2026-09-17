package clock

import "time"

// Clock is what time it really is.
type Clock struct{}

func NewClock() *Clock {
	return &Clock{}
}

func (clock *Clock) Now() time.Time {
	return time.Now()
}
