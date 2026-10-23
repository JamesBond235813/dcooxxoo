package application

import (
	"testing"
	"time"

	"github.com/dujiao-next/internal/constants"
	settingsintegration "github.com/dujiao-next/internal/modules/settings/schema/integration"
	"github.com/dujiao-next/internal/shared/jsonmap"
)

func TestSyncedSpecValuesPlusHalfStaysInPhilippines(t *testing.T) {
	existing := jsonmap.JSON{"payment_regions": []interface{}{map[string]interface{}{"country": "US", "currency": "USD"}}}
	incoming := jsonmap.JSON{"plan": "plus_half", "payment_regions": []map[string]interface{}{{"country": "CL", "currency": "CLP"}}}
	got := syncedSpecValues(existing, incoming, constants.ConnectionProtocolZovoCard, "plus_half")
	regions, ok := got["payment_regions"].([]map[string]string)
	if !ok || len(regions) != 1 || regions[0]["country"] != "PH" || regions[0]["currency"] != "PHP" {
		t.Fatalf("expected only PH/PHP, got %#v", got["payment_regions"])
	}
	if got["plan"] != "plus_half" {
		t.Fatalf("plan field was lost: %#v", got)
	}
}

// fakeSettingsProvider 以固定同步间隔实现 SettingsProvider。
type fakeSettingsProvider struct {
	interval time.Duration
}

func (f fakeSettingsProvider) GetUpstreamSyncConfig(fallbackInterval string) (settingsintegration.UpstreamSyncConfig, error) {
	cfg := settingsintegration.DefaultUpstreamSyncConfig()
	cfg.IntervalMinutes = int(f.interval / time.Minute)
	return cfg, nil
}

func (f fakeSettingsProvider) GetUpstreamSyncInterval(fallbackInterval string) (time.Duration, error) {
	return f.interval, nil
}

func TestComputeFullSyncIntervalFloorsAt24h(t *testing.T) {
	svc := &Service{settings: fakeSettingsProvider{interval: 5 * time.Minute}}

	// 5m 同步间隔 × 3 = 15m < 24h，期望落到 24h floor
	got := svc.computeFullSyncInterval()
	if got != fullSyncIntervalFloor {
		t.Fatalf("expected floor=24h, got %v", got)
	}
}

func TestComputeFullSyncIntervalScalesWithLongInterval(t *testing.T) {
	svc := &Service{settings: fakeSettingsProvider{interval: 12 * time.Hour}}

	// 12h * 3 = 36h，应使用 scaled 值
	got := svc.computeFullSyncInterval()
	want := 36 * time.Hour
	if got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestComputeFullSyncIntervalWithoutSettings(t *testing.T) {
	svc := &Service{}
	got := svc.computeFullSyncInterval()
	if got != fullSyncIntervalFloor {
		t.Fatalf("expected floor when settings=nil, got %v", got)
	}
}
