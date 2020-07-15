package util

import(
	"fmt"
	"time"
)

type MovingStats interface {
	Add(t time.Time, v float64)

	Mean(since time.Time) (float64, error)
	Sum(since time.Time) (float64, error)
	Max(since time.Time) (float64, error)
	Min(since time.Time) (float64, error)
	Variation(since time.Time) (float64, error)
}

type movingStats struct {

	values map[time.Time]float64
	maxCacheDuration time.Duration
}

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

	if timeOutsideOfCache(since, ma.maxCacheDuration) {
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

	if timeOutsideOfCache(since, ma.maxCacheDuration) {
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

func (ma *movingStats) Max(since time.Time) (float64, error) {

	if timeOutsideOfCache(since, ma.maxCacheDuration) {
		return 0.0, fmt.Errorf(
			"time since (%s) excceeds maxCacheDuration (%s)",
			time.Now().Sub(since),
			ma.maxCacheDuration,
		)
	}

	if len(ma.values) == 0 {
		return 0.0, nil
	}

	var max float64
	var maxSet bool
	for t, v := range ma.values {
		if !maxSet {
			max = v
			maxSet = true
		} else if t.After(since) && v > max{
			max = v
		}
	}

	return max, nil
}

func (ma *movingStats) Min(since time.Time) (float64, error) {

	if timeOutsideOfCache(since, ma.maxCacheDuration) {
		return 0.0, fmt.Errorf(
			"time since (%s) excceeds maxCacheDuration (%s)",
			time.Now().Sub(since),
			ma.maxCacheDuration,
		)
	}

	if len(ma.values) == 0 {
		return 0.0, nil
	}

	var min float64
	var minSet bool
	for t, v := range ma.values {
		if !minSet {
			min = v
			minSet = true
		} else if t.After(since) && v < min{
			min = v
		}
	}

	return min, nil
}

func (ma *movingStats) Variation(since time.Time) (float64, error) {

	min, err := ma.Min(since)
	if err != nil {
		return 0.0, err
	}
	max, err := ma.Max(since)
	if err != nil {
		return 0.0, err
	}
	return max - min, nil
}

func timeOutsideOfCache(t time.Time, maxCacheDuration time.Duration) bool {

	return time.Now().Sub(t) > maxCacheDuration
}
