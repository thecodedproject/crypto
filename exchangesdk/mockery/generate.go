package mockery

import "github.com/thecodedproject/crypto/exchangesdk"

//go:generate mockery --dir=.. --outpkg=mockery --output=. --case=snake --name=Client

var _ exchangesdk.Client = (*Client)(nil)
