package util

import(
	"fmt"
	"time"
)

type Float64MovingAverage interface {
	Add(t time.Time, v float64)
	Since(t time.Time) (float64, error)
}

type float64MovingAverage struct {

	values map[time.Time]float64
	maxCacheDuration time.Duration
}

func NewFloat64MovingAverage(maxCacheDuration time.Duration) Float64MovingAverage {

	return &float64MovingAverage{
		values: make(map[time.Time]float64),
		maxCacheDuration: maxCacheDuration,
	}
}

func (ma *float64MovingAverage) Add(t time.Time, v float64) {

	ma.values[t] = v

	for valueTime := range ma.values {
		if valueTime.Before(t.Add(-ma.maxCacheDuration)) {
			delete(ma.values, valueTime)
		}
	}
}

func (ma *float64MovingAverage) Since(since time.Time) (float64, error) {

	if time.Now().Sub(since) > ma.maxCacheDuration {
		return 0.0, fmt.Errorf(
			"time since (%s) excceeds maxCacheDuration (%s)",
			time.Now().Sub(since),
			ma.maxCacheDuration,
		)
	}

	if len(ma.values) == 0 {
		return 0.0, nil
	}

	var sum float64
	var count int64
	for t, v := range ma.values {
		if t.After(since) {
			sum += v
			count++
		}
	}

	return sum/float64(count), nil
}
