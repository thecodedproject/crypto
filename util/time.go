package util

import (
	"time"
)

type Time interface {
	Now() time.Time
}

type timeImpl struct {}

func (t timeImpl) Now() time.Time {

	return time.Now()
}
