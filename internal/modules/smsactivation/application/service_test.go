package application

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/modules/smsactivation/domain"
	"github.com/dujiao-next/internal/modules/smsactivation/infrastructure/fivesim"
)

type memoryRepo struct {
	items map[string]*domain.SmsActivation
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{items: make(map[string]*domain.SmsActivation)}
}

func (r *memoryRepo) Create(a *domain.SmsActivation) error {
	a.ID = uint(len(r.items) + 1)
	r.items[a.OrderNo] = a
	return nil
}

func (r *memoryRepo) Update(a *domain.SmsActivation) error {
	r.items[a.OrderNo] = a
	return nil
}

func (r *memoryRepo) GetLatestByOrderNo(orderNo string) (*domain.SmsActivation, error) {
	return r.items[orderNo], nil
}

func (r *memoryRepo) ListByOrderNo(orderNo string) ([]domain.SmsActivation, error) {
	if item, ok := r.items[orderNo]; ok {
		return []domain.SmsActivation{*item}, nil
	}
	return nil, nil
}

type testOrderReader struct{ order *domain.Order }

func (r testOrderReader) GetSmsOrderByOrderNo(orderNo string) (*domain.Order, error) {
	if r.order.OrderNo != orderNo {
		return nil, nil
	}
	return r.order, nil
}

type testDeliverer struct {
	orderID uint
	payload string
}

func (d *testDeliverer) DeliverSmsOrder(orderID uint, payload string, _ time.Time) error {
	d.orderID, d.payload = orderID, payload
	return nil
}

func TestSmsActivationService_Lifecycle(t *testing.T) {
	// 1. Mock 5sim server
	callCount := 0
	expires := time.Now().UTC().Add(15 * time.Minute).Format(time.RFC3339)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/user/buy/activation/vietnam/any/openai" {
			callCount++
			if callCount == 1 {
				_, _ = fmt.Fprintf(w, `{"id":1001,"phone":"+84911111111","status":"PENDING","expires":%q,"country":"vietnam"}`, expires)
			} else {
				_, _ = fmt.Fprintf(w, `{"id":1002,"phone":"+84922222222","status":"PENDING","expires":%q,"country":"vietnam"}`, expires)
			}
			return
		}
		if r.URL.Path == "/user/cancel/1001" {
			_, _ = w.Write([]byte(`{"status": "CANCELED"}`))
			return
		}
		if r.URL.Path == "/user/check/1001" {
			_, _ = w.Write([]byte(`{
				"id": 1001,
				"phone": "+84911111111",
				"status": "PENDING"
			}`))
			return
		}
		if r.URL.Path == "/user/check/1002" {
			_, _ = w.Write([]byte(`{
				"id": 1002,
				"phone": "+84922222222",
				"status": "RECEIVED",
				"sms": [{"code": "888999", "text": "Your OpenAI code is 888999"}]
			}`))
			return
		}
		if r.URL.Path == "/user/finish/1002" {
			_, _ = w.Write([]byte(`{"status": "FINISHED"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	repo := newMemoryRepo()
	client := fivesim.NewClient(ts.URL, "mock-token")
	order := &domain.Order{ID: 42, OrderNo: "ORD-SMS-TEST-001", UserID: 10, Status: constants.OrderStatusPaid, Items: []domain.OrderItem{{
		TitleJSON: map[string]interface{}{"zh-CN": "OpenAI接码台"},
		Schema:    map[string]interface{}{"type": "sms_activation", "service": "openai"},
	}}}
	deliverer := &testDeliverer{}
	svc := NewService(testOrderReader{order: order}, repo, fivesim.NewProvider(client), deliverer, Config{
		Enabled:   true,
		Countries: []string{"vietnam"},
		Product:   "openai",
	})

	ctx := context.Background()

	// Step 1: Acquire number
	act1, err := svc.AcquireNumber(ctx, "ORD-SMS-TEST-001", 10, "")
	if err != nil {
		t.Fatalf("AcquireNumber failed: %v", err)
	}
	if act1.Phone != "+84911111111" {
		t.Errorf("expected phone +84911111111, got %s", act1.Phone)
	}
	if act1.Status != domain.StatusPending {
		t.Errorf("expected status pending, got %s", act1.Status)
	}

	// Step 2: Acquire again within expiry -> should return same without buying new
	act1Again, err := svc.AcquireNumber(ctx, "ORD-SMS-TEST-001", 10, "")
	if err != nil {
		t.Fatalf("AcquireNumber again failed: %v", err)
	}
	if act1Again.ProviderOrderID != 1001 {
		t.Errorf("expected dedup same order id 1001, got %d", act1Again.ProviderOrderID)
	}

	// Step 3: Check status -> still pending
	act1Checked, err := svc.CheckStatus(ctx, "ORD-SMS-TEST-001", 10, "")
	if err != nil {
		t.Fatalf("CheckStatus failed: %v", err)
	}
	if act1Checked.Status != domain.StatusPending {
		t.Errorf("expected still pending, got %s", act1Checked.Status)
	}

	// Step 4: Retry / Change Number
	act2, err := svc.RetryNumber(ctx, "ORD-SMS-TEST-001", 10, "")
	if err != nil {
		t.Fatalf("RetryNumber failed: %v", err)
	}
	if act2.Phone != "+84922222222" {
		t.Errorf("expected new phone +84922222222, got %s", act2.Phone)
	}
	if act2.RetryCount != 1 {
		t.Errorf("expected retry_count 1, got %d", act2.RetryCount)
	}

	// Step 5: Check status on second number -> receives code and completes
	act2Checked, err := svc.CheckStatus(ctx, "ORD-SMS-TEST-001", 10, "")
	if err != nil {
		t.Fatalf("CheckStatus failed: %v", err)
	}
	if act2Checked.Code != "888999" {
		t.Errorf("expected code 888999, got %s", act2Checked.Code)
	}
	if act2Checked.Status != domain.StatusFinished {
		t.Errorf("expected status finished, got %s", act2Checked.Status)
	}

	// Step 6: Verify delivery is routed through the shared fulfillment service port.
	if deliverer.orderID != 42 || !strings.Contains(deliverer.payload, "验证码: 888999") {
		t.Errorf("expected shared delivery call for order 42 with received code, got order=%d payload=%q", deliverer.orderID, deliverer.payload)
	}
}
