package util

import(
	"fmt"
	"math"
	"time"
)

type MovingStats interface {
	Add(t time.Time, v float64)

	Latest() float64
	Mean(since time.Time) (float64, error)
	MeanLatestOrNan(l time.Duration) float64
	Sum(since time.Time) (float64, error)
	Max(since time.Time) (float64, error)
	Min(since time.Time) (float64, error)
	Variation(since time.Time) (float64, error)
	Gradient(since time.Time) (float64, error)
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

func (ma *movingStats) Latest() float64 {

	if len(ma.values) == 0 {
		return 0.0
	}

	var latestT time.Time
	for t := range ma.values {
		if t.After(latestT) {
			latestT = t
		}
	}

	return ma.values[latestT]
}

func (ma *movingStats) MeanLatest(d time.Duration) (float64, error) {

	return ma.Mean(timeAgo(d))
}

func (ma *movingStats) MeanLatestOrNan(d time.Duration) float64 {

	return ma.MeanOrNan(timeAgo(d))
}

func (ma *movingStats) Mean(since time.Time) (float64, error) {

	v := ma.MeanOrNan(since)
	if math.IsNaN(v) {
		return 0.0, fmt.Errorf(
			"time since (%s) excceeds maxCacheDuration (%s)",
			time.Now().Sub(since),
			ma.maxCacheDuration,
		)
	}
	return v, nil
}

func (ma *movingStats) MeanOrNan(since time.Time) float64 {

	if timeOutsideOfCache(since, ma.maxCacheDuration) {
		return math.NaN()
	}

	if len(ma.values) == 0 {
		return 0.0
	}

	var sum float64
	var count int64
	for t, v := range ma.values {
		if t.After(since) {
			sum += v
			count++
		}
	}

	return sum/float64(count)
}

func (ma *movingStats) SumLatest(d time.Duration) (float64, error) {

	return ma.Sum(timeAgo(d))
}

func (ma *movingStats) SumLatestOrNan(d time.Duration) float64 {

	return ma.SumOrNan(timeAgo(d))
}

func (ma *movingStats) Sum(since time.Time) (float64, error) {

	v := ma.SumOrNan(since)
	if math.IsNaN(v) {
		return 0.0, fmt.Errorf(
			"time since (%s) excceeds maxCacheDuration (%s)",
			time.Now().Sub(since),
			ma.maxCacheDuration,
		)
	}
	return v, nil
}

func (ma *movingStats) SumOrNan(since time.Time) float64 {

	if timeOutsideOfCache(since, ma.maxCacheDuration) {
		return math.NaN()
	}

	if len(ma.values) == 0 {
		return 0.0
	}

	var sum float64
	for t, v := range ma.values {
		if t.After(since) {
			sum += v
		}
	}

	return sum
}

func (ma *movingStats) MaxLatest(d time.Duration) (float64, error) {

	return ma.Max(timeAgo(d))
}

func (ma *movingStats) MaxLatestOrNan(d time.Duration) float64 {

	return ma.MaxOrNan(timeAgo(d))
}

func (ma *movingStats) Max(since time.Time) (float64, error) {

	v := ma.MaxOrNan(since)
	if math.IsNaN(v) {
		return 0.0, fmt.Errorf(
			"time since (%s) excceeds maxCacheDuration (%s)",
			time.Now().Sub(since),
			ma.maxCacheDuration,
		)
	}
	return v, nil
}

func (ma *movingStats) MaxOrNan(since time.Time) float64 {

	if timeOutsideOfCache(since, ma.maxCacheDuration) {
		return math.NaN()
	}

	if len(ma.values) == 0 {
		return 0.0
	}

	var max float64
	var maxSet bool
	for t, v := range ma.values {
		if !t.After(since) {
			continue
		}
		if !maxSet {
			max = v
			maxSet = true
		} else if t.After(since) && v > max{
			max = v
		}
	}

	return max
}

func (ma *movingStats) MinLatest(d time.Duration) (float64, error) {

	return ma.Min(timeAgo(d))
}

func (ma *movingStats) MinLatestOrNan(d time.Duration) float64 {

	return ma.MinOrNan(timeAgo(d))
}

func (ma *movingStats) Min(since time.Time) (float64, error) {

	v := ma.MinOrNan(since)
	if math.IsNaN(v) {
		return 0.0, fmt.Errorf(
			"time since (%s) excceeds maxCacheDuration (%s)",
			time.Now().Sub(since),
			ma.maxCacheDuration,
		)
	}
	return v, nil
}

func (ma *movingStats) MinOrNan(since time.Time) float64 {

	if timeOutsideOfCache(since, ma.maxCacheDuration) {
		return math.NaN()
	}

	if len(ma.values) == 0 {
		return 0.0
	}

	var min float64
	var minSet bool
	for t, v := range ma.values {
		if !t.After(since) {
			continue
		}
		if !minSet {
			min = v
			minSet = true
		} else if t.After(since) && v < min{
			min = v
		}
	}

	return min
}

func (ma *movingStats) VariationLatest(d time.Duration) (float64, error) {

	return ma.Variation(timeAgo(d))
}

func (ma *movingStats) VariationLatestOrNan(d time.Duration) float64 {

	return ma.VariationOrNan(timeAgo(d))
}

func (ma *movingStats) Variation(since time.Time) (float64, error) {

	v := ma.VariationOrNan(since)
	if math.IsNaN(v) {
		return 0.0, fmt.Errorf(
			"time since (%s) excceeds maxCacheDuration (%s)",
			time.Now().Sub(since),
			ma.maxCacheDuration,
		)
	}
	return v, nil
}

func (ma *movingStats) VariationOrNan(since time.Time) float64 {

	if timeOutsideOfCache(since, ma.maxCacheDuration) {
		return math.NaN()
	}

	min := ma.MinOrNan(since)
	if min == math.NaN() {
		return math.NaN()
	}
	max := ma.MaxOrNan(since)
	if max == math.NaN() {
		return math.NaN()
	}
	return max - min
}

func (ma *movingStats) GradientLatest(d time.Duration) (float64, error) {

	return ma.Gradient(timeAgo(d))
}

func (ma *movingStats) GradientLatestOrNan(d time.Duration) float64 {

	return ma.GradientOrNan(timeAgo(d))
}

func (ma *movingStats) Gradient(since time.Time) (float64, error) {

	v := ma.GradientOrNan(since)
	if math.IsNaN(v) {
		return 0.0, fmt.Errorf(
			"time since (%s) excceeds maxCacheDuration (%s)",
			time.Now().Sub(since),
			ma.maxCacheDuration,
		)
	}
	return v, nil
}

func (ma *movingStats) GradientOrNan(since time.Time) float64 {

	if timeOutsideOfCache(since, ma.maxCacheDuration) {
		return math.NaN()
	}

	if len(ma.values) == 0 {
		return 0.0
	}

	var firstValue float64
	var firstValueTime time.Time
	var lastValue float64
	var lastValueTime time.Time
	var initalValuesSet bool
	for t, v := range ma.values {

		if !t.After(since) {
			continue
		}
		if !initalValuesSet {
			firstValue = v
			firstValueTime = t
			lastValue = v
			lastValueTime = t
			initalValuesSet = true
			continue
		}
		if t.Before(firstValueTime) {
			firstValue = v
			firstValueTime = t
			continue
		}
		if t.After(lastValueTime) {
			lastValue = v
			lastValueTime = t
			continue
		}
	}

	return lastValue - firstValue
}

func timeOutsideOfCache(t time.Time, maxCacheDuration time.Duration) bool {

	return time.Now().Sub(t) > maxCacheDuration
}

func timeAgo(d time.Duration) time.Time {

	return time.Now().Add(-d)
}
