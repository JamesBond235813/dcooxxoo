package upstream

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dujiao-next/internal/constants"
	siteconnectiondomain "github.com/dujiao-next/internal/modules/siteconnection/domain"
	"github.com/dujiao-next/internal/shared/jsonmap"
)

func TestZovoCardAdapter_Ping(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/openapi/v1/balance" {
			if r.Header.Get("X-API-Key") != "test-key" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"msg":  "ok",
				"data": map[string]interface{}{
					"balance":           100.5,
					"spendable_balance": 80.5,
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	conn := &siteconnectiondomain.Connection{
		BaseURL:  ts.URL,
		ApiKey:   "test-key",
		Protocol: constants.ConnectionProtocolZovoCard,
	}

	adapter := NewZovoCardAdapter(conn, t.TempDir())
	res, err := adapter.Ping(context.Background())
	if err != nil {
		t.Fatalf("unexpected error on ping: %v", err)
	}
	if res.Balance != "80.50" {
		t.Errorf("expected balance 80.50, got %s", res.Balance)
	}
}

func TestZovoCardAdapter_ListProducts(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/openapi/v1/gpt-direct/plans" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"msg":  "ok",
				"data": map[string]interface{}{
					"version": "1.0",
					"registry": []map[string]interface{}{
						{
							"key":                   "plus",
							"name":                  "Plus",
							"purchasable":           true,
							"service_fee_usd_minor": 15,
						},
						{
							"key":                   "pro_50x",
							"name":                  "Pro (50x)",
							"purchasable":           true,
							"service_fee_usd_minor": 15,
						},
					},
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	conn := &siteconnectiondomain.Connection{
		BaseURL:  ts.URL,
		ApiKey:   "test-key",
		Protocol: constants.ConnectionProtocolZovoCard,
	}

	adapter := NewZovoCardAdapter(conn, t.TempDir())
	res, err := adapter.ListProducts(context.Background(), ListProductsOpts{})
	if err != nil {
		t.Fatalf("unexpected error listing products: %v", err)
	}
	if len(res.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(res.Items))
	}
	if res.Items[0].SKUs[0].SKUCode != "plus" {
		t.Errorf("expected sku code plus, got %s", res.Items[0].SKUs[0].SKUCode)
	}
	if res.Items[0].FulfillmentType != constants.FulfillmentTypeAuto {
		t.Errorf("expected provider auto fulfillment, got %s", res.Items[0].FulfillmentType)
	}
	if res.Items[0].SKUs[0].StockQuantity != -1 {
		t.Errorf("expected untracked/on-demand stock, got %d", res.Items[0].SKUs[0].StockQuantity)
	}
}

func TestZovoCardAdapter_ListProducts_PlusHalfOnlyPhilippines(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"payment_regions": []map[string]string{{"country": "US", "currency": "USD"}},
				"registry": []map[string]interface{}{{
					"key": "plus_half", "name": "Plus Half", "purchasable": true,
					"payment_regions": []string{"US", "JP"},
				}},
			},
		})
	}))
	defer ts.Close()
	adapter := NewZovoCardAdapter(&siteconnectiondomain.Connection{BaseURL: ts.URL, ApiKey: "test-key"}, t.TempDir())
	result, err := adapter.ListProducts(context.Background(), ListProductsOpts{})
	if err != nil || len(result.Items) != 1 {
		t.Fatalf("unexpected result: %#v, %v", result, err)
	}
	regions, ok := result.Items[0].SKUs[0].SpecValues["payment_regions"].([]map[string]string)
	if !ok || len(regions) != 1 || regions[0]["country"] != "PH" || regions[0]["currency"] != "PHP" {
		t.Fatalf("half-price regions must be PH/PHP, got %#v", result.Items[0].SKUs[0].SpecValues["payment_regions"])
	}
}

func TestZovoCardAdapter_CheckPlusHalfEligibility(t *testing.T) {
	for _, tc := range []struct {
		name     string
		eligible interface{}
		want     bool
	}{
		{name: "eligible", eligible: true, want: true},
		{name: "ineligible", eligible: false},
		{name: "unknown", eligible: nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Fatal(err)
				}
				if request["plan"] != "plus_half" || request["payment_country"] != "PH" || request["payment_currency"] != "PHP" {
					t.Fatalf("wrong half-price preflight parameters: %#v", request)
				}
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": map[string]interface{}{"email": "buyer@example.com", "promotions": map[string]interface{}{"plus_half": map[string]interface{}{"eligible": tc.eligible}}}})
			}))
			defer ts.Close()
			adapter := NewZovoCardAdapter(&siteconnectiondomain.Connection{BaseURL: ts.URL, ApiKey: "test-key"}, t.TempDir())
			got, email, err := adapter.CheckPlusHalfEligibility(context.Background(), "session")
			if err != nil || got != tc.want || email != "buyer@example.com" {
				t.Fatalf("eligible=%v email=%q err=%v", got, email, err)
			}
		})
	}
}

func TestZovoCardAdapter_CreateOrder_AutoRedeem(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/openapi/v1/gpt-direct/plans":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": map[string]interface{}{"registry": []map[string]interface{}{{"key": "plus", "name": "Plus", "purchasable": true, "service_fee_usd_minor": 15, "payment_regions": []string{"PH"}}}},
			})
		case "/openapi/v1/gpt-direct/preflight":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": map[string]interface{}{
					"quotes": map[string]interface{}{
						"plus": map[string]interface{}{
							"country":           "PH",
							"currency":          "PHP",
							"amountMinor":       100000,
							"minorUnitExponent": 2,
						},
					},
				},
			})
		case "/openapi/v1/fx/rates":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": map[string]interface{}{
					"rates": []map[string]interface{}{
						{
							"currency":     "PHP",
							"usd_per_unit": 0.017,
						},
					},
				},
			})
		case "/openapi/v1/gpt-direct/cdks":
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["plan"] != "plus" {
				t.Fatalf("expected resolved Plus plan, got %#v (err=%v)", body, err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"msg":  "ok",
				"data": map[string]interface{}{
					"requested": 1,
					"issued": []map[string]interface{}{
						{
							"id":               101,
							"code":             "UUUPLUS-TEST-1234",
							"code_prefix":      "UUUPLUS-TEST",
							"plan":             "plus",
							"fee_amount_minor": 15,
						},
					},
				},
			})
		case "/api/v1/cdk/preview":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"msg":  "ok",
				"data": map[string]interface{}{
					"redemption_token": "token_prev_123",
					"plan":             "plus",
				},
			})
		case "/api/v1/cdk/preflight":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"msg":  "ok",
				"data": map[string]interface{}{
					"preflight_token": "token_pre_456",
				},
			})
		case "/openapi/v1/gpt-direct/cdks/redeem":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"msg":  "ok",
				"data": map[string]interface{}{
					"order": map[string]interface{}{
						"order_id": 888,
						"status":   "awaiting_card",
					},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	conn := &siteconnectiondomain.Connection{
		BaseURL:  ts.URL,
		ApiKey:   "test-key",
		Protocol: constants.ConnectionProtocolZovoCard,
	}

	adapter := NewZovoCardAdapter(conn, t.TempDir())
	res, err := adapter.CreateOrder(context.Background(), CreateUpstreamOrderReq{
		SKUID:             stablePlanID("plus"),
		Quantity:          1,
		DownstreamOrderNo: "DJ-2026-TEST-001",
		ManualFormData: jsonmap.JSON{
			"plan":    "plus",
			"session": "session_sample_token",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error creating order: %v", err)
	}
	if !res.OK {
		t.Fatalf("expected order response OK, got error: %s", res.ErrorMessage)
	}
	if res.OrderID != 888 {
		t.Errorf("expected OrderID 888, got %d", res.OrderID)
	}
	if res.Status != "fulfilling" {
		t.Errorf("expected status fulfilling, got %s", res.Status)
	}
}

func TestZovoCardAdapter_CreateOrder_ResolvesPlanAndDeliversEveryCDK(t *testing.T) {
	issueCalls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/openapi/v1/gpt-direct/plans":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": map[string]interface{}{"registry": []map[string]interface{}{{"key": "plus", "purchasable": true}, {"key": "pro_20x", "purchasable": true}}}})
		case "/openapi/v1/gpt-direct/cdks":
			issueCalls++
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode issuance request: %v", err)
			}
			if body["plan"] != "pro_20x" || body["count"] != float64(2) {
				t.Fatalf("wrong upstream plan/quantity: %#v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": map[string]interface{}{"issued": []map[string]interface{}{{"id": 1, "code": "CDK-A", "fee_amount_minor": 100}, {"id": 2, "code": "CDK-B", "fee_amount_minor": 100}}}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	adapter := NewZovoCardAdapter(&siteconnectiondomain.Connection{BaseURL: ts.URL, ApiKey: "test-key", Protocol: constants.ConnectionProtocolZovoCard}, t.TempDir())
	res, err := adapter.CreateOrder(context.Background(), CreateUpstreamOrderReq{SKUID: stablePlanID("pro_20x"), Quantity: 2, DownstreamOrderNo: "DJ-MULTI-1"})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if !res.OK || res.Fulfillment == nil || res.Fulfillment.Payload != "CDK-A\nCDK-B" {
		t.Fatalf("expected both CDKs in fulfillment, got %+v", res)
	}
	if res.OrderNo == "CDK-A" || res.OrderNo == "CDK-B" || issueCalls != 1 {
		t.Fatalf("supplier reference must not expose a CDK; calls=%d order_no=%q", issueCalls, res.OrderNo)
	}
}

func TestZovoCardAdapter_CreateOrder_UnknownSKUFailsClosed(t *testing.T) {
	issueCalled := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/openapi/v1/gpt-direct/plans" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": map[string]interface{}{"registry": []map[string]interface{}{{"key": "plus", "purchasable": true}}}})
			return
		}
		if r.URL.Path == "/openapi/v1/gpt-direct/cdks" {
			issueCalled = true
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	adapter := NewZovoCardAdapter(&siteconnectiondomain.Connection{BaseURL: ts.URL, ApiKey: "test-key", Protocol: constants.ConnectionProtocolZovoCard}, t.TempDir())
	if _, err := adapter.CreateOrder(context.Background(), CreateUpstreamOrderReq{SKUID: stablePlanID("not-a-plan"), Quantity: 1}); err == nil {
		t.Fatal("unknown SKU must be rejected")
	}
	if issueCalled {
		t.Fatal("must not issue a CDK for an unknown SKU")
	}
}

func TestZovoCardAdapter_CreateOrder_PlusHalfUsesPhilippines(t *testing.T) {
	issued := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/openapi/v1/gpt-direct/plans":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": map[string]interface{}{"registry": []map[string]interface{}{{"key": "plus_half", "purchasable": true}}}})
		case "/openapi/v1/gpt-direct/preflight":
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["plan"] != "plus_half" || body["payment_country"] != "PH" || body["payment_currency"] != "PHP" {
				t.Errorf("wrong preflight routing: %#v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": map[string]interface{}{"email": "customer@example.com", "promotions": map[string]interface{}{"plus_half": map[string]interface{}{"eligible": true}}}})
		case "/openapi/v1/gpt-direct/cdks":
			issued++
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["plan"] != "plus_half" || body["payment_country"] != "PH" || body["payment_currency"] != "PHP" {
				t.Errorf("wrong issuance routing: %#v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": map[string]interface{}{"issued": []map[string]interface{}{{"id": 321, "code": "HALF-CDK", "fee_amount_minor": 15}}}})
		case "/api/v1/cdk/preview":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": map[string]interface{}{"redemption_token": "preview"}})
		case "/api/v1/cdk/preflight":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": map[string]interface{}{"preflight_token": "preflight"}})
		case "/openapi/v1/gpt-direct/cdks/redeem":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": map[string]interface{}{"order": map[string]interface{}{"order_id": 654, "status": "awaiting_card"}}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()
	a := NewZovoCardAdapter(&siteconnectiondomain.Connection{BaseURL: ts.URL, ApiKey: "test-key"}, t.TempDir())
	result, err := a.CreateOrder(context.Background(), CreateUpstreamOrderReq{SKUID: stablePlanID("plus_half"), Quantity: 1, DownstreamOrderNo: "HALF-TEST-1", ManualFormData: jsonmap.JSON{"session": "test-session"}})
	if err != nil || result == nil || !result.OK || result.OrderID != 654 || issued != 1 {
		t.Fatalf("half-price auto recharge failed: result=%+v err=%v issued=%d", result, err, issued)
	}
}
