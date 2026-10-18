package fivesim

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFivesimClient_BuyActivation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.URL.Path == "/user/buy/activation/vietnam/any/openai" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"id": 123456,
				"phone": "+84912345678",
				"operator": "viettel",
				"product": "openai",
				"price": 0.13,
				"status": "PENDING",
				"expires": "2026-10-13T18:00:00Z",
				"country": "vietnam"
			}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-token")
	order, err := client.BuyActivation(context.Background(), "vietnam", "any", "openai")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.Phone != "+84912345678" {
		t.Errorf("expected phone +84912345678, got %s", order.Phone)
	}
	if order.ID != 123456 {
		t.Errorf("expected id 123456, got %d", order.ID)
	}
}

func TestFivesimClient_RejectsMissingToken(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		http.Error(w, "unexpected request", http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "  ")
	_, err := client.BuyActivation(context.Background(), "vietnam", "any", "openai")
	if err == nil {
		t.Fatal("expected missing token error")
	}
	if called {
		t.Fatal("request was sent without a configured token")
	}
}

func TestFivesimClient_CheckOrder_Received(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user/check/123456" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"id": 123456,
				"phone": "+84912345678",
				"status": "RECEIVED",
				"sms": [
					{
						"sender": "OpenAI",
						"text": "Your OpenAI code is 654321",
						"code": "654321"
					}
				]
			}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-token")
	order, err := client.CheckOrder(context.Background(), 123456)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.Status != "RECEIVED" {
		t.Errorf("expected status RECEIVED, got %s", order.Status)
	}
	if len(order.SMS) != 1 || order.SMS[0].Code != "654321" {
		t.Errorf("expected code 654321, got %+v", order.SMS)
	}
}
