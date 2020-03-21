package time

import (
	"testing"
	gotime "time"
)

var nowFunc = gotime.Now

func Now() gotime.Time {

	return nowFunc()
}

func SetTimeNowFuncForTesting(_ *testing.T, now func() gotime.Time) func() {

	oldNowFunc := nowFunc
	nowFunc = now
	return func() {
		nowFunc = oldNowFunc
	}
}

func SetTimeNowForTesting(t *testing.T, time gotime.Time) func() {

	return SetTimeNowFuncForTesting(t, func() gotime.Time {
		return time
	})
}
