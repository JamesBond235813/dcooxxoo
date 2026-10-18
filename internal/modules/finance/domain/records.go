package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// ManualReceipt records owner-entered cash receipts and upstream funding transfers.
type ManualReceipt struct {
	ID           uint            `gorm:"primarykey" json:"id"`
	EntryDate    string          `gorm:"type:date;not null;index" json:"entry_date"`
	EntryType    string          `gorm:"type:varchar(32);not null;index" json:"entry_type"`
	Amount       decimal.Decimal `gorm:"type:decimal(20,2);not null" json:"amount"`
	Currency     string          `gorm:"type:varchar(16);not null;default:'CNY'" json:"currency"`
	Counterparty string          `gorm:"type:varchar(160)" json:"counterparty,omitempty"`
	Reference    string          `gorm:"type:varchar(160)" json:"reference,omitempty"`
	OrderNo      string          `gorm:"type:varchar(80);index" json:"order_no,omitempty"`
	Note         string          `gorm:"type:text" json:"note,omitempty"`
	CreatedBy    uint            `gorm:"not null;default:0;index" json:"created_by"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    *time.Time      `gorm:"index" json:"-"`
}

func (ManualReceipt) TableName() string { return "finance_manual_receipts" }

// CashFlowEntry records a manually confirmed cash movement or account transfer.
// Source-system events are read from their owning modules and are not duplicated here.
type CashFlowEntry struct {
	ID                   uint            `gorm:"primarykey" json:"id"`
	EventDate            string          `gorm:"type:date;not null;index" json:"event_date"`
	Direction            string          `gorm:"type:varchar(16);not null;index" json:"direction"` // in, out, transfer
	Category             string          `gorm:"type:varchar(48);not null;index" json:"category"`
	Account              string          `gorm:"type:varchar(48);not null;index" json:"account"`
	CounterAccount       string          `gorm:"type:varchar(48)" json:"counter_account,omitempty"`
	Amount               decimal.Decimal `gorm:"type:decimal(20,6);not null" json:"amount"`
	Currency             string          `gorm:"type:varchar(16);not null;default:'CNY'" json:"currency"`
	ExchangeRateSnapshot decimal.Decimal `gorm:"type:decimal(16,6);not null;default:1" json:"exchange_rate_snapshot"`
	CNYAmount            decimal.Decimal `gorm:"type:decimal(20,6);not null;default:0" json:"cny_amount"`
	Counterparty         string          `gorm:"type:varchar(160)" json:"counterparty,omitempty"`
	Reference            string          `gorm:"type:varchar(160);index" json:"reference,omitempty"`
	OrderNo              string          `gorm:"type:varchar(80);index" json:"order_no,omitempty"`
	RelatedSource        string          `gorm:"type:varchar(48)" json:"related_source,omitempty"`
	RelatedSourceID      string          `gorm:"type:varchar(120)" json:"related_source_id,omitempty"`
	Note                 string          `gorm:"type:text" json:"note,omitempty"`
	CreatedBy            uint            `gorm:"not null;default:0;index" json:"created_by"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
	DeletedAt            *time.Time      `gorm:"index" json:"-"`
}

func (CashFlowEntry) TableName() string { return "finance_cash_flow_entries" }

// UpstreamFinanceRecord is a sanitized, locally persisted ZovoCard ledger row.
// Reports read this table; only the explicit sync action contacts ZovoCard.
type UpstreamFinanceRecord struct {
	ID                   uint            `gorm:"primarykey" json:"id"`
	ConnectionID         uint            `gorm:"not null;index;uniqueIndex:idx_finance_upstream_source" json:"connection_id"`
	Source               string          `gorm:"type:varchar(32);not null;uniqueIndex:idx_finance_upstream_source" json:"source"`
	SourceKey            string          `gorm:"type:varchar(128);not null;uniqueIndex:idx_finance_upstream_source" json:"-"`
	EventDate            string          `gorm:"type:date;not null;index" json:"event_date"`
	EventTime            time.Time       `gorm:"index" json:"event_time"`
	EventType            string          `gorm:"type:varchar(64);index" json:"event_type"`
	Status               string          `gorm:"type:varchar(64);index" json:"status"`
	Direction            string          `gorm:"type:varchar(16);not null;index" json:"direction"`
	Category             string          `gorm:"type:varchar(48);not null;index" json:"category"`
	Account              string          `gorm:"type:varchar(48);not null" json:"account"`
	Amount               decimal.Decimal `gorm:"type:decimal(20,6);not null" json:"amount"`
	Currency             string          `gorm:"type:varchar(16);not null" json:"currency"`
	ExchangeRateSnapshot decimal.Decimal `gorm:"type:decimal(16,6);not null;default:1" json:"exchange_rate_snapshot"`
	CNYAmount            decimal.Decimal `gorm:"type:decimal(20,6);not null;default:0" json:"cny_amount"`
	Reference            string          `gorm:"type:varchar(160);index" json:"reference,omitempty"`
	Description          string          `gorm:"type:text" json:"description,omitempty"`
	RawJSON              string          `gorm:"type:json" json:"-"`
	SyncedAt             time.Time       `gorm:"not null;index" json:"synced_at"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

func (UpstreamFinanceRecord) TableName() string { return "finance_upstream_records" }

// CDKDelivery records user delivery allocations against upstream stock.
type CDKDelivery struct {
	ID             uint       `gorm:"primarykey" json:"id"`
	DeliveryDate   string     `gorm:"type:date;not null;index" json:"delivery_date"`
	Plan           string     `gorm:"type:varchar(40);not null;index" json:"plan"`
	ProductID      uint       `gorm:"not null;default:0;index" json:"product_id"`
	Recipient      string     `gorm:"type:varchar(160);not null;index" json:"recipient"`
	Quantity       int        `gorm:"not null" json:"quantity"`
	UpstreamCDKIDs string     `gorm:"type:text" json:"-"`
	Note           string     `gorm:"type:text" json:"note,omitempty"`
	CreatedBy      uint       `gorm:"not null;default:0;index" json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `gorm:"index" json:"-"`
}

func (CDKDelivery) TableName() string { return "finance_cdk_deliveries" }
