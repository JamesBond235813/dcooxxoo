package gormstore

import (
	"errors"
	"strings"

	orderdomain "github.com/dujiao-next/internal/modules/order/domain"
	"github.com/dujiao-next/internal/modules/smsactivation/domain"
	"gorm.io/gorm"
)

type OrderReader struct{ db *gorm.DB }

func NewOrderReader(db *gorm.DB) *OrderReader { return &OrderReader{db: db} }

func (r *OrderReader) GetSmsOrderByOrderNo(orderNo string) (*domain.Order, error) {
	var order orderdomain.Order
	err := r.db.Preload("Items").Preload("Children.Items").Where("order_no = ?", strings.TrimSpace(orderNo)).First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	result := &domain.Order{ID: order.ID, OrderNo: order.OrderNo, UserID: order.UserID, GuestEmail: order.GuestEmail, Status: order.Status, Items: mapItems(order.Items)}
	if containsSmsItem(result.Items) {
		return result, nil
	}
	for _, child := range order.Children {
		items := mapItems(child.Items)
		if !containsSmsItem(items) {
			continue
		}
		result.ID = child.ID
		result.Status = child.Status
		result.Items = items
		if child.UserID > 0 {
			result.UserID = child.UserID
		}
		if child.GuestEmail != "" {
			result.GuestEmail = child.GuestEmail
		}
		return result, nil
	}
	return result, nil
}

func mapItems(items []orderdomain.OrderItem) []domain.OrderItem {
	result := make([]domain.OrderItem, 0, len(items))
	for _, item := range items {
		result = append(result, domain.OrderItem{TitleJSON: item.TitleJSON, Schema: item.ManualFormSchemaSnapshotJSON})
	}
	return result
}

func containsSmsItem(items []domain.OrderItem) bool {
	for _, item := range items {
		if item.Schema["type"] == "sms_activation" || item.Schema["service"] == "openai" {
			return true
		}
		for _, title := range item.TitleJSON {
			if value, ok := title.(string); ok && (strings.Contains(value, "接码") || strings.Contains(value, "收码") || strings.Contains(strings.ToLower(value), "sms")) {
				return true
			}
		}
	}
	return false
}
