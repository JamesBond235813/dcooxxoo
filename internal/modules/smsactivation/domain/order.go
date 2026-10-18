package domain

import (
	"context"
	"time"
)

// Order is the minimum order information required by the SMS activation use case.
type Order struct {
	ID         uint
	OrderNo    string
	UserID     uint
	GuestEmail string
	Status     string
	Items      []OrderItem
}

type OrderItem struct {
	TitleJSON map[string]interface{}
	Schema    map[string]interface{}
}

type OrderReader interface {
	GetSmsOrderByOrderNo(orderNo string) (*Order, error)
}

type Deliverer interface {
	DeliverSmsOrder(orderID uint, payload string, deliveredAt time.Time) error
}

type ProviderOrder struct {
	ID       int64
	Phone    string
	Country  string
	Operator string
	Product  string
	Price    float64
	Expires  string
}

type ProviderSMS struct {
	Code string
	Text string
}

type ProviderStatus struct {
	Status string
	SMS    []ProviderSMS
}

type Provider interface {
	AutoBuyFromPool(ctx context.Context, countries []string, product string, maxPrice float64, priorFailures map[string]int) (*ProviderOrder, error)
	CheckOrder(ctx context.Context, orderID int64) (*ProviderStatus, error)
	FinishOrder(ctx context.Context, orderID int64) error
	CancelOrder(ctx context.Context, orderID int64) error
}
