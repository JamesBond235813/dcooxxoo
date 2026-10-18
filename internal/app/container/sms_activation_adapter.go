package container

import "time"

type smsSystemFulfillment interface {
	CreateSystemManual(orderID uint, payload string, deliveredAt time.Time) error
}

type smsFulfillmentAdapter struct{ service smsSystemFulfillment }

func (a smsFulfillmentAdapter) DeliverSmsOrder(orderID uint, payload string, deliveredAt time.Time) error {
	return a.service.CreateSystemManual(orderID, payload, deliveredAt)
}
