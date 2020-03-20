package util_test

import (
  "github.com/stretchr/testify/assert"
	"github.com/thecodedproject/crypto/util"
	"testing"
	"time"
)

func TestMockTimeNow(t *testing.T) {

	someTime := time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	m := new(util.MockTime)
	m.On("Now").Return(someTime)
	assert.Equal(t, someTime, m.Now())
}

func TestTimeNow(t *testing.T) {

	utilTime := util.NewTime()
	now := utilTime.Now()

	diff := time.Now().Sub(now)
	assert.True(t, diff < 5*time.Millisecond)
}
