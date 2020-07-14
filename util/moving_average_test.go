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

func TestFloa64MovingAverageSince(t *testing.T) {

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

			ma := util.NewFloat64MovingAverage(test.MaxDuration)

			for _, v := range test.Values {
				ma.Add(v.Time, v.Value)
			}

			av, err := ma.Since(test.SinceTime)
			if test.ExpectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.ExpectedAverage, av)
		})
	}
}
