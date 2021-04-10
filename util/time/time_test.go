package time_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	utiltime "github.com/thecodedproject/crypto/util/time"
)

func TestSetTimeNowFuncAndReset(t *testing.T) {

	someTime := time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	reset := utiltime.SetTimeNowFuncForTesting(t, func() time.Time {
		return someTime
	})
	assert.Equal(t, someTime, utiltime.Now())
	assert.Equal(t, someTime, utiltime.Now())

	reset()

	now := utiltime.Now()
	diff := time.Now().Sub(now)
	assert.True(t, diff < 5*time.Millisecond)
}

func TestSetTimeNowValueAndReset(t *testing.T) {

	someTime := time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	reset := utiltime.SetTimeNowForTesting(t, someTime)

	assert.Equal(t, someTime, utiltime.Now())
	assert.Equal(t, someTime, utiltime.Now())

	reset()

	now := utiltime.Now()
	diff := time.Now().Sub(now)
	assert.True(t, diff < 5*time.Millisecond)
}

func TestTimeNowWithoutMocking(t *testing.T) {

	now := utiltime.Now()
	diff := time.Now().Sub(now)
	assert.True(t, diff < 5*time.Millisecond)
}
