package util_test

import(
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thecodedproject/crypto/util"
	"testing"
	"time"
)

func secondsAgo(i int) time.Time {

	return time.Now().Add(time.Duration(-i)*time.Second)
}

func TestMovingStatsLatest(t * testing.T) {
	type TimeValue struct {
		Time time.Time
		Value float64
	}

	testCases := []struct{
		Name string
		Values []TimeValue
		Expected float64
	}{
		{
			Name: "No values returns zero",
		},
		{
			Name: "Some values returns the latest",
			Values: []TimeValue{
				{secondsAgo(100), 20.0},
				{secondsAgo(10), 22.0},
				{secondsAgo(1), 24.0},
			},
			Expected: 24.0,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {

			ms := util.NewMovingStats(time.Minute)

			for _, v := range test.Values {
				ms.Add(v.Time, v.Value)
			}

			val := ms.Latest()
			assert.Equal(t, test.Expected, val)
		})
	}

}

func TestMovingStatsMean(t *testing.T) {

	type TimeValue struct {
		Time time.Time
		Value float64
	}

	testCases := []struct{
		Name string
		Values []TimeValue
		MaxDuration time.Duration
		SinceTime time.Time
		ExpectedAverage float64
		ExpectErr bool
	}{
		{
			Name: "No values returns average of zero",
			MaxDuration: time.Minute,
			SinceTime: secondsAgo(10),
			ExpectedAverage: 0.0,
		},
		{
			Name: "Single value within time since returns average of value",
			Values: []TimeValue{
				{secondsAgo(1), 20.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			ExpectedAverage: 20.0,
		},
		{
			Name: "Multiple values within time since returns average of values",
			Values: []TimeValue{
				{secondsAgo(1), 20.0},
				{secondsAgo(2), 30.0},
				{secondsAgo(3), 40.0},
				{secondsAgo(4), 50.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			ExpectedAverage: 35.0,
		},
		{
			Name: "Multiple values returns average of values since TimeSince",
			Values: []TimeValue{
				{secondsAgo(1), 20.0},
				{secondsAgo(2), 30.0},
				{secondsAgo(11), 40.0},
				{secondsAgo(12), 50.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			ExpectedAverage: 25.0,
		},
		{
			Name: "Request average since time which is older than max duration returns error",
			MaxDuration: time.Second,
			SinceTime: secondsAgo(10),
			ExpectErr: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {

			ma := util.NewMovingStats(test.MaxDuration)

			for _, v := range test.Values {
				ma.Add(v.Time, v.Value)
			}

			av, err := ma.Mean(test.SinceTime)
			if test.ExpectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.ExpectedAverage, av)
		})
	}
}

func TestMovingStatsSum(t *testing.T) {

	type TimeValue struct {
		Time time.Time
		Value float64
	}

	testCases := []struct{
		Name string
		Values []TimeValue
		MaxDuration time.Duration
		SinceTime time.Time
		ExpectedSum float64
		ExpectErr bool
	}{
		{
			Name: "No values returns sum of zero",
			MaxDuration: time.Minute,
			SinceTime: secondsAgo(10),
			ExpectedSum: 0.0,
		},
		{
			Name: "Single value within time since returns value",
			Values: []TimeValue{
				{secondsAgo(1), 20.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			ExpectedSum: 20.0,
		},
		{
			Name: "Multiple values within time since returns sum of values",
			Values: []TimeValue{
				{secondsAgo(1), 20.0},
				{secondsAgo(2), 30.0},
				{secondsAgo(3), 40.0},
				{secondsAgo(4), 50.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			ExpectedSum: 140.0,
		},
		{
			Name: "Multiple values returns sum of values since TimeSince",
			Values: []TimeValue{
				{secondsAgo(1), 20.0},
				{secondsAgo(2), 30.0},
				{secondsAgo(11), 40.0},
				{secondsAgo(12), 50.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			ExpectedSum: 50.0,
		},
		{
			Name: "Request sum since time which is older than max duration returns error",
			MaxDuration: time.Second,
			SinceTime: secondsAgo(10),
			ExpectErr: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {

			ma := util.NewMovingStats(test.MaxDuration)

			for _, v := range test.Values {
				ma.Add(v.Time, v.Value)
			}

			av, err := ma.Sum(test.SinceTime)
			if test.ExpectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.ExpectedSum, av)
		})
	}
}

func TestMovingStatsMax(t *testing.T) {

	type TimeValue struct {
		Time time.Time
		Value float64
	}

	testCases := []struct{
		Name string
		Values []TimeValue
		MaxDuration time.Duration
		SinceTime time.Time
		ExpectedSum float64
		ExpectErr bool
	}{
		{
			Name: "No values",
			MaxDuration: time.Minute,
			SinceTime: secondsAgo(10),
			ExpectedSum: 0.0,
		},
		{
			Name: "Single value within time since",
			Values: []TimeValue{
				{secondsAgo(1), -20.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			ExpectedSum: -20.0,
		},
		{
			Name: "Multiple values all within time since",
			Values: []TimeValue{
				{secondsAgo(1), 20.0},
				{secondsAgo(2), 30.0},
				{secondsAgo(3), 40.0},
				{secondsAgo(4), 50.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			ExpectedSum: 50.0,
		},
		{
			Name: "All negative values",
			Values: []TimeValue{
				{secondsAgo(1), -20.0},
				{secondsAgo(2), -30.0},
				{secondsAgo(3), -40.0},
				{secondsAgo(4), -50.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			ExpectedSum: -20.0,
		},
		{
			Name: "Multiple values with only some within since TimeSince",
			Values: []TimeValue{
				{secondsAgo(1), 20.0},
				{secondsAgo(2), 30.0},
				{secondsAgo(11), 40.0},
				{secondsAgo(12), 50.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			ExpectedSum: 30.0,
		},
		{
			Name: "Request since time which is older than max duration returns error",
			MaxDuration: time.Second,
			SinceTime: secondsAgo(10),
			ExpectErr: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {

			ma := util.NewMovingStats(test.MaxDuration)

			for _, v := range test.Values {
				ma.Add(v.Time, v.Value)
			}

			av, err := ma.Max(test.SinceTime)
			if test.ExpectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.ExpectedSum, av)
		})
	}
}

func TestMovingStatsMin(t *testing.T) {

	type TimeValue struct {
		Time time.Time
		Value float64
	}

	testCases := []struct{
		Name string
		Values []TimeValue
		MaxDuration time.Duration
		SinceTime time.Time
		ExpectedSum float64
		ExpectErr bool
	}{
		{
			Name: "No values",
			MaxDuration: time.Minute,
			SinceTime: secondsAgo(10),
			ExpectedSum: 0.0,
		},
		{
			Name: "Single value within time since",
			Values: []TimeValue{
				{secondsAgo(1), 20.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			ExpectedSum: 20.0,
		},
		{
			Name: "Multiple values all within time since",
			Values: []TimeValue{
				{secondsAgo(1), 20.0},
				{secondsAgo(2), 30.0},
				{secondsAgo(3), 40.0},
				{secondsAgo(4), 50.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			ExpectedSum: 20.0,
		},
		{
			Name: "All negative values",
			Values: []TimeValue{
				{secondsAgo(1), -20.0},
				{secondsAgo(2), -30.0},
				{secondsAgo(3), -40.0},
				{secondsAgo(4), -50.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			ExpectedSum: -50.0,
		},
		{
			Name: "Multiple values with only some within since TimeSince",
			Values: []TimeValue{
				{secondsAgo(1), 20.0},
				{secondsAgo(2), 30.0},
				{secondsAgo(11), 10.0},
				{secondsAgo(12), 50.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			ExpectedSum: 20.0,
		},
		{
			Name: "Request since time which is older than max duration returns error",
			MaxDuration: time.Second,
			SinceTime: secondsAgo(10),
			ExpectErr: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {

			ma := util.NewMovingStats(test.MaxDuration)

			for _, v := range test.Values {
				ma.Add(v.Time, v.Value)
			}

			av, err := ma.Min(test.SinceTime)
			if test.ExpectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.ExpectedSum, av)
		})
	}
}

func TestMovingStatsVariation(t *testing.T) {

	type TimeValue struct {
		Time time.Time
		Value float64
	}

	testCases := []struct{
		Name string
		Values []TimeValue
		MaxDuration time.Duration
		SinceTime time.Time
		ExpectedSum float64
		ExpectErr bool
	}{
		{
			Name: "No values",
			MaxDuration: time.Minute,
			SinceTime: secondsAgo(10),
			ExpectedSum: 0.0,
		},
		{
			Name: "Single value within time since",
			Values: []TimeValue{
				{secondsAgo(1), 20.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			ExpectedSum: 0.0,
		},
		{
			Name: "Multiple values all within time since",
			Values: []TimeValue{
				{secondsAgo(1), 20.0},
				{secondsAgo(2), 30.0},
				{secondsAgo(3), 40.0},
				{secondsAgo(4), 50.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			ExpectedSum: 30.0,
		},
		{
			Name: "All negative values",
			Values: []TimeValue{
				{secondsAgo(1), -20.0},
				{secondsAgo(2), -30.0},
				{secondsAgo(3), -40.0},
				{secondsAgo(4), -50.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			ExpectedSum: 30.0,
		},
		{
			Name: "Multiple values with only some within since TimeSince",
			Values: []TimeValue{
				{secondsAgo(1), 20.0},
				{secondsAgo(2), 30.0},
				{secondsAgo(11), 10.0},
				{secondsAgo(12), 50.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			ExpectedSum: 10.0,
		},
		{
			Name: "Request since time which is older than max duration returns error",
			MaxDuration: time.Second,
			SinceTime: secondsAgo(10),
			ExpectErr: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {

			ma := util.NewMovingStats(test.MaxDuration)

			for _, v := range test.Values {
				ma.Add(v.Time, v.Value)
			}

			av, err := ma.Variation(test.SinceTime)
			if test.ExpectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.ExpectedSum, av)
		})
	}
}

func TestMovingStatsGradient(t *testing.T) {

	type TimeValue struct {
		Time time.Time
		Value float64
	}

	testCases := []struct{
		Name string
		Values []TimeValue
		MaxDuration time.Duration
		SinceTime time.Time
		Expected float64
		ExpectErr bool
	}{
		{
			Name: "No values",
			MaxDuration: time.Minute,
			SinceTime: secondsAgo(10),
			Expected: 0.0,
		},
		{
			Name: "Single value within time since",
			Values: []TimeValue{
				{secondsAgo(1), 20.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			Expected: 0.0,
		},
		{
			Name: "Multiple values all within time since",
			Values: []TimeValue{
				{secondsAgo(1), 40.0},
				{secondsAgo(2), 20.0},
				{secondsAgo(3), 60.0},
				{secondsAgo(4), 50.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			Expected: -10.0,
		},
		{
			Name: "All negative values",
			Values: []TimeValue{
				{secondsAgo(1), -30.0},
				{secondsAgo(2), -10.0},
				{secondsAgo(3), -60.0},
				{secondsAgo(4), -50.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			Expected: 20.0,
		},
		{
			Name: "Multiple values with only some within since TimeSince",
			Values: []TimeValue{
				{secondsAgo(1), 20.0},
				{secondsAgo(2), 70.0},
				{secondsAgo(3), 40.0},
				{secondsAgo(11), 10.0},
				{secondsAgo(12), 50.0},
			},
			MaxDuration: time.Hour,
			SinceTime: secondsAgo(10),
			Expected: -20.0,
		},
		{
			Name: "Request since time which is older than max duration returns error",
			MaxDuration: time.Second,
			SinceTime: secondsAgo(10),
			ExpectErr: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {

			ma := util.NewMovingStats(test.MaxDuration)

			for _, v := range test.Values {
				ma.Add(v.Time, v.Value)
			}

			av, err := ma.Gradient(test.SinceTime)
			if test.ExpectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.Expected, av)
		})
	}
}
