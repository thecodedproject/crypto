package util_test

import(
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thecodedproject/crypto/util"
	"testing"
	"time"
)

func TestFloa64MovingSumSince(t *testing.T) {

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

			ma := util.NewFloat64MovingSum(test.MaxDuration)

			for _, v := range test.Values {
				ma.Add(v.Time, v.Value)
			}

			av, err := ma.Since(test.SinceTime)
			if test.ExpectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.ExpectedSum, av)
		})
	}
}
