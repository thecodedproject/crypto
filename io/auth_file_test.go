package io_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/thecodedproject/crypto/io"
	"testing"
)

func TestReadAuthFile(t *testing.T) {

	authFile := "example_api_auth.json"

	_, err := io.ReadAuthFile(authFile)
	assert.NoError(t, err)
}
