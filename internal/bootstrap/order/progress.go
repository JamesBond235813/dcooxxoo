package orderwiring

import (
	"context"
	"strings"
	"sync"
	"time"

	procurementcontract "github.com/dujiao-next/internal/modules/procurement/contract"
	procurementdomain "github.com/dujiao-next/internal/modules/procurement/domain"
)

type progressOrderRepository interface {
	GetByLocalOrderNo(localOrderNo string) (*procurementdomain.Order, error)
}

type cachedOrderProgress struct {
	value   procurementcontract.Progress
	expires time.Time
}

type orderProgressAdapter struct {
	orders      progressOrderRepository
	connections procurementcontract.ConnectionProvider
	mu          sync.Mutex
	cache       map[uint]cachedOrderProgress
}

func newOrderProgressAdapter(orders progressOrderRepository, connections procurementcontract.ConnectionProvider) *orderProgressAdapter {
	return &orderProgressAdapter{orders: orders, connections: connections}
}

// GetLiveProgressByLocalOrderNo is called only after the order handler has
// verified the requesting user or guest owns this order.
func (a *orderProgressAdapter) GetLiveProgressByLocalOrderNo(orderNo string) (*procurementcontract.Progress, error) {
	order, err := a.orders.GetByLocalOrderNo(orderNo)
	if err != nil || order == nil {
		return nil, err
	}
	progress := procurementcontract.Progress{Status: order.Status}
	if order.Status != "accepted" || order.UpstreamOrderID == 0 {
		return &progress, nil
	}
	now := time.Now()
	a.mu.Lock()
	if cached, ok := a.cache[order.ID]; ok && now.Before(cached.expires) {
		a.mu.Unlock()
		return &cached.value, nil
	}
	a.mu.Unlock()
	connection, err := a.connections.Open(order.ConnectionID)
	if err != nil || connection == nil {
		return &progress, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	detail, err := connection.GetOrder(ctx, order.UpstreamOrderID)
	if err == nil && detail != nil {
		progress.Status = strings.ToLower(strings.TrimSpace(detail.Status))
		stage := strings.TrimSpace(detail.Stage)
		if len(stage) <= 64 && publicStage(stage) {
			progress.Stage = stage
		}
		for _, event := range detail.Events {
			eventStage := strings.TrimSpace(event.Stage)
			eventStatus := strings.TrimSpace(event.Status)
			if len(eventStage) > 64 || !publicStage(eventStage) || len(eventStatus) > 64 || !publicStage(eventStatus) || (eventStage == "" && eventStatus == "") {
				continue
			}
			at := strings.TrimSpace(event.At)
			if _, parseErr := time.Parse(time.RFC3339, at); parseErr != nil {
				at = ""
			}
			progress.Events = append(progress.Events, procurementcontract.ProgressEvent{Stage: eventStage, Status: eventStatus, At: at})
		}
		if len(progress.Events) > 20 {
			progress.Events = progress.Events[len(progress.Events)-20:]
		}
	}
	a.mu.Lock()
	if a.cache == nil || len(a.cache) > 512 {
		a.cache = make(map[uint]cachedOrderProgress)
	}
	a.cache[order.ID] = cachedOrderProgress{value: progress, expires: now.Add(8 * time.Second)}
	a.mu.Unlock()
	return &progress, nil
}

func publicStage(stage string) bool {
	for _, char := range stage {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '_' && char != '-' {
			return false
		}
	}
	return true
}
