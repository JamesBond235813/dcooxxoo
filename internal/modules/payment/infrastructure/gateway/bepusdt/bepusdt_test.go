package bepusdt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dujiao-next/internal/constants"
)

func TestCreatePaymentSignsAddressAndOrderSpecificRate(t *testing.T) {
	const address = "TXbPPq88sTWxirGWqxXmvfrmLERhwFCXpK"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != bepusdtCreateTransactionPath {
			t.Errorf("request path = %s", r.URL.Path)
		}
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		signature, _ := payload["signature"].(string)
		delete(payload, "signature")
		if signature != Sign(payload, "test-secret") {
			t.Error("order request signature does not cover the complete payload")
		}
		if payload["rate"] != "6.67" || payload["address"] != address {
			t.Errorf("rate/address = %v/%v", payload["rate"], payload["address"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status_code":200,"message":"success","data":{"trade_id":"trade-1","order_id":"order-1","amount":"168.00","actual_amount":"25.19","token":"` + address + `","payment_url":"https://pay.example.test/order-1"}}`))
	}))
	defer server.Close()
	cfg := &Config{GatewayURL: server.URL, AuthToken: "test-secret", TradeType: "usdt.trc20", Fiat: "CNY", Address: address}
	result, err := CreatePayment(context.Background(), cfg, CreateInput{OrderNo: "order-1", Amount: "168.00", Rate: "6.67", NotifyURL: "https://shop.example.test/callback", ReturnURL: "https://shop.example.test/pay"})
	if err != nil {
		t.Fatalf("create BEpusdt payment: %v", err)
	}
	if result.Token != address || result.ActualAmount != "25.19" {
		t.Fatalf("unexpected payment result: %+v", result)
	}
}

func TestParseConfigAndNormalizeDefaults(t *testing.T) {
	cfg, err := ParseConfig(map[string]interface{}{
		"gateway_url": " https://pay.example.com/ ",
		"auth_token":  " token ",
		"currencies":  "USDT,USDC",
		"notify_url":  " https://example.com/notify ",
		"return_url":  " https://example.com/return ",
	})
	if err != nil {
		t.Fatalf("parse config failed: %v", err)
	}
	if cfg.TradeType != bepusdtTradeTypeUSDTTRC20 {
		t.Fatalf("unexpected default trade type: %s", cfg.TradeType)
	}
	if cfg.Fiat != constants.SiteCurrencyDefault {
		t.Fatalf("unexpected default fiat: %s", cfg.Fiat)
	}
	if cfg.GatewayURL != "https://pay.example.com" {
		t.Fatalf("unexpected normalized gateway url: %s", cfg.GatewayURL)
	}
	if cfg.Currencies != "" {
		t.Fatalf("transaction mode currencies should be empty, got %q", cfg.Currencies)
	}
}

func TestParseConfig_CashierModeKeepsTradeTypeEmpty(t *testing.T) {
	cfg, err := ParseConfig(map[string]interface{}{
		"gateway_url": " https://pay.example.com/ ",
		"auth_token":  " token ",
		"order_mode":  constants.PaymentBepusdtOrderModeCashier,
		"trade_type":  " usdt.trc20 ",
		"currencies":  " usdt, usdc ",
		"notify_url":  " https://example.com/notify ",
		"return_url":  " https://example.com/return ",
	})
	if err != nil {
		t.Fatalf("parse config failed: %v", err)
	}
	if cfg.TradeType != "" {
		t.Fatalf("cashier mode trade type should stay empty, got %s", cfg.TradeType)
	}
	if cfg.Currencies != "USDT,USDC" {
		t.Fatalf("cashier currencies = %q, want USDT,USDC", cfg.Currencies)
	}
}

func TestResolveTradeType(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{name: "USDT", input: bepusdtChannelTypeUSDT, expect: bepusdtTradeTypeUSDTTRC20},
		{name: "USDTTRC20", input: bepusdtChannelTypeUSDTTRC20, expect: bepusdtTradeTypeUSDTTRC20},
		{name: "USDCTRC20", input: bepusdtChannelTypeUSDCTRC20, expect: bepusdtTradeTypeUSDCTRC20},
		{name: "TRX", input: bepusdtChannelTypeTRX, expect: bepusdtTradeTypeTRX},
		{name: "Unknown", input: "unknown", expect: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolveTradeType(tc.input); got != tc.expect {
				t.Fatalf("unexpected trade type: got %s, want %s", got, tc.expect)
			}
		})
	}
}

func TestToPaymentStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
		expect string
	}{
		{name: "Success", status: StatusSuccess, expect: constants.PaymentStatusSuccess},
		{name: "Expired", status: StatusExpired, expect: constants.PaymentStatusExpired},
		{name: "Waiting", status: StatusWaiting, expect: constants.PaymentStatusPending},
		{name: "Unknown", status: 999, expect: constants.PaymentStatusPending},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ToPaymentStatus(tc.status); got != tc.expect {
				t.Fatalf("unexpected payment status: got %s, want %s", got, tc.expect)
			}
		})
	}
}
