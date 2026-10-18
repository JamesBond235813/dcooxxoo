package fivesim

import (
	"context"

	"github.com/dujiao-next/internal/modules/smsactivation/domain"
)

type Provider struct {
	client *Client
}

func NewProvider(client *Client) *Provider { return &Provider{client: client} }

func (p *Provider) AutoBuyFromPool(ctx context.Context, countries []string, product string, maxPrice float64, priorFailures map[string]int) (*domain.ProviderOrder, error) {
	order, err := p.client.AutoBuyFromPool(ctx, countries, product, maxPrice, priorFailures)
	if err != nil || order == nil {
		return nil, err
	}
	return &domain.ProviderOrder{ID: order.ID, Phone: order.Phone, Country: order.Country, Operator: order.Operator, Product: order.Product, Price: order.Price, Expires: order.Expires}, nil
}

func (p *Provider) CheckOrder(ctx context.Context, orderID int64) (*domain.ProviderStatus, error) {
	status, err := p.client.CheckOrder(ctx, orderID)
	if err != nil || status == nil {
		return nil, err
	}
	result := &domain.ProviderStatus{Status: status.Status}
	for _, sms := range status.SMS {
		result.SMS = append(result.SMS, domain.ProviderSMS{Code: sms.Code, Text: sms.Text})
	}
	return result, nil
}

func (p *Provider) FinishOrder(ctx context.Context, orderID int64) error {
	return p.client.FinishOrder(ctx, orderID)
}
func (p *Provider) CancelOrder(ctx context.Context, orderID int64) error {
	return p.client.CancelOrder(ctx, orderID)
}
