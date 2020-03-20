package util_test

import (
  "github.com/stretchr/testify/assert"
	"github.com/thecodedproject/crypto/util"
	"testing"
	"time"
)

func TestMockTime(t *testing.T) {

	someTime := time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	m := new(util.MockTime)
	m.On("Now").Return(someTime)
	assert.Equal(t, someTime, m.Now())
}
