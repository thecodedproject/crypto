package util

import (
  "github.com/stretchr/testify/mock"
	"time"
)

type MockTime struct{
	mock.Mock
}

func (m *MockTime) Now() time.Time {

	args := m.Called()
	return args.Get(0).(time.Time)
}
