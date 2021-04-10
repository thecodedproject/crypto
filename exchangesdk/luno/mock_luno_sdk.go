package luno

import (
	"context"

	luno_sdk "github.com/luno/luno-go"
	"github.com/stretchr/testify/mock"
)

// MockLunoSdk provides a testify mock of the LunoSDK API
type MockLunoSdk struct {
	mock.Mock
}

func (m *MockLunoSdk) GetTicker(ctx context.Context, req *luno_sdk.GetTickerRequest) (*luno_sdk.GetTickerResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*luno_sdk.GetTickerResponse), args.Error(1)
}

func (m *MockLunoSdk) PostLimitOrder(ctx context.Context, req *luno_sdk.PostLimitOrderRequest) (*luno_sdk.PostLimitOrderResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*luno_sdk.PostLimitOrderResponse), args.Error(1)
}

func (m *MockLunoSdk) StopOrder(ctx context.Context, req *luno_sdk.StopOrderRequest) (*luno_sdk.StopOrderResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*luno_sdk.StopOrderResponse), args.Error(1)
}

func (m *MockLunoSdk) GetOrder(ctx context.Context, req *luno_sdk.GetOrderRequest) (*luno_sdk.GetOrderResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*luno_sdk.GetOrderResponse), args.Error(1)
}

func (m *MockLunoSdk) ListUserTrades(ctx context.Context, req *luno_sdk.ListUserTradesRequest) (*luno_sdk.ListUserTradesResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*luno_sdk.ListUserTradesResponse), args.Error(1)
}
