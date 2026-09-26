package _interface

import "time"

//go:generate go tool mockgen -source=i_clock.go -destination=mocks/mock_i_clock.go -package=mocks

type IClock interface {
	Now() time.Time
}
