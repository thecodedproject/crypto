package exchangesdk

import (
  "context"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}


func (m *MockClient) LatestPrice(ctx context.Context) (decimal.Decimal, error) {

	args := m.Called(ctx)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockClient) GetOrderStatus(ctx context.Context, orderId string) (OrderStatus, error) {

	args := m.Called(ctx, orderId)
	return args.Get(0).(OrderStatus), args.Error(1)
}

func (m *MockClient) GetTrades(ctx context.Context, page int64) ([]Trade, error) {

	args := m.Called(ctx, page)
	return args.Get(0).([]Trade), args.Error(1)
}

func (m *MockClient) PostLimitOrder(ctx context.Context, order Order) (string, error) {

	args := m.Called(ctx, order)
	return args.String(0), args.Error(1)
}

func (m *MockClient) StopOrder(ctx context.Context, orderId string) error {

	args := m.Called(ctx, orderId)
	return args.Error(0)
}

func (m *MockClient) MakerFee() decimal.Decimal {

	args := m.Called()
	return args.Get(0).(decimal.Decimal)
}
