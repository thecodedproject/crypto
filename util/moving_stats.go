package util

import(
	"fmt"
	"time"
)

type MovingStats interface {
	Add(t time.Time, v float64)

	Mean(since time.Time) (float64, error)
	Sum(since time.Time) (float64, error)
}

type movingStats struct {

	values map[time.Time]float64
	maxCacheDuration time.Duration
}

// TODO: Change this struct to be a MovingStats struct with various stats methods:
// SumSince
// AverageSince
// TWMASince
// MaxSince
// MinSince
// VariationSince ( this is MaxSince - MinSince)
// others??
//
// THen remove the Float64MovingAverage struct
func NewMovingStats(maxCacheDuration time.Duration) MovingStats {

	return &movingStats{
		values: make(map[time.Time]float64),
		maxCacheDuration: maxCacheDuration,
	}
}

func (ma *movingStats) Add(t time.Time, v float64) {

	ma.values[t] = v

	for valueTime := range ma.values {
		if valueTime.Before(t.Add(-ma.maxCacheDuration)) {
			delete(ma.values, valueTime)
		}
	}
}

func (ma *movingStats) Mean(since time.Time) (float64, error) {

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

func (ma *movingStats) Sum(since time.Time) (float64, error) {

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
	for t, v := range ma.values {
		if t.After(since) {
			sum += v
		}
	}

	return sum, nil
}
