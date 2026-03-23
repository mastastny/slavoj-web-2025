package service

import (
	"time"

	"github.com/mastastny/slavoj-web-2025/internal/config"
)

type LockerService struct {
	conf config.Config
}

func NewLockerService(conf config.Config) *LockerService {
	return &LockerService{conf: conf}
}

func (l LockerService) GetLockerCode(targetTime time.Time) string {
	if l.conf.Areal.CabinetLockCode != "" {
		return l.conf.Areal.CabinetLockCode
	}

	tz, err := time.LoadLocation("Europe/Prague")
	if err != nil { // Always check errors even if they should not happen.
		panic(err)
	}

	return targetTime.In(tz).Format("0106")
}
