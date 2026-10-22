package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	procurementcontract "github.com/dujiao-next/internal/modules/procurement/contract"
	procurementdomain "github.com/dujiao-next/internal/modules/procurement/domain"
)

// PreflightSessionForOrder asks the configured upstream to validate the
// submitted credential without saving it or creating a procurement order.
func (s *Service) PreflightSessionForOrder(orderID, productID, skuID uint, session string) (map[string]interface{}, error) {
	session = strings.TrimSpace(session)
	if orderID == 0 || productID == 0 || skuID == 0 || session == "" || len(session) > 1<<20 {
		return nil, errors.New("invalid session preflight request")
	}
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("load order for session preflight: %w", err)
	}
	if order == nil {
		return nil, procurementcontract.ErrOrderNotFound
	}
	orders := []procurementdomain.LocalOrder{*order}
	orders = append(orders, order.Children...)
	matched := false
	for _, candidate := range orders {
		for _, item := range candidate.Items {
			if item.ProductID == productID && item.SKUID == skuID &&
				item.FulfillmentType == constants.FulfillmentTypeUpstream && schemaRequiresSession(item.ManualFormSchemaSnapshot) {
				if item.Quantity != 1 {
					return nil, errors.New("session preflight requires a single-quantity order")
				}
				matched = true
			}
		}
	}
	if !matched {
		return nil, errors.New("order item does not require upstream session preflight")
	}
	connectionID, found, err := s.mappingRepo.FindConnectionID(productID)
	if err != nil {
		return nil, fmt.Errorf("resolve upstream mapping: %w", err)
	}
	if !found {
		return nil, errors.New("order product has no active upstream mapping")
	}
	connection, err := s.connections.Open(connectionID)
	if err != nil {
		return nil, fmt.Errorf("open upstream connection: %w", err)
	}
	preflighter, ok := connection.(procurementcontract.SessionPreflighter)
	if !ok {
		return nil, errors.New("upstream connection does not support session preflight")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	return preflighter.PreflightSession(ctx, session)
}

// CheckPlusHalfEligibilityForProduct validates a credential before a customer
// is allowed to create a Plus half-price order. No order or CDK is created.
func (s *Service) CheckPlusHalfEligibilityForProduct(productID uint, session string) (bool, string, error) {
	session = strings.TrimSpace(session)
	if productID == 0 || session == "" || len(session) > 1<<20 {
		return false, "", errors.New("invalid Plus half-price eligibility request")
	}
	connectionID, found, err := s.mappingRepo.FindConnectionID(productID)
	if err != nil {
		return false, "", err
	}
	if !found {
		return false, "", errors.New("Plus half-price product has no upstream mapping")
	}
	connection, err := s.connections.Open(connectionID)
	if err != nil {
		return false, "", err
	}
	checker, ok := connection.(interface {
		CheckPlusHalfEligibility(context.Context, string) (bool, string, error)
	})
	if !ok {
		return false, "", errors.New("upstream does not support Plus half-price eligibility")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	return checker.CheckPlusHalfEligibility(ctx, session)
}
