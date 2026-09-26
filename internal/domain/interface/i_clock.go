package _interface

import "time"

//go:generate go tool mockgen -source=i_clock.go -destination=mocks/mock_i_clock.go -package=mocks

// IClock is what time it is, so that how long a judgement is remembered can be
// checked rather than waited for.
type IClock interface {
	Now() time.Time
}
