package database

import (
	"time"
)

// Trade represents both trade instructions and submitted orders in one structure
type Trade struct {
	ID            int64     `db:"id"`
	StrategyName  string    `db:"strategy_name"`
	ContractID    int       `db:"contract_id"`
	Exchange      string    `db:"exchange"`
	Symbol        string    `db:"symbol"`
	Side          string    `db:"side"`
	Quantity      int       `db:"quantity"`
	Price         float64   `db:"price"`           // Either quote or fill price
	BrokerOrderID int       `db:"broker_order_id"` // 0 for unsubmitted trades
	TradingDate   string    `db:"trading_date"`    // YYYY-MM-DD format
	Status        string    `db:"status"`          // pending, submitted, filled, cancelled, rejected
	CreatedAt     time.Time `db:"created_at"`
	LastUpdatedAt time.Time `db:"last_updated_at"`
}
