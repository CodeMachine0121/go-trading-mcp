package _interface

import "time"

//go:generate go tool mockgen -source=i_clock.go -destination=mocks/mock_i_clock.go -package=mocks

// IClock is what time it is.
//
// It exists so that "this signing-in has expired" is a rule that can be checked
// rather than waited for. Without it, the only honest test of the renewal path takes
// fifteen minutes.
type IClock interface {
	Now() time.Time
}
