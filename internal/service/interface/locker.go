package _interface

import "time"

type Locker interface {
	GetLockerCode(targetTime time.Time) string
}
