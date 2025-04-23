package database

import (
	"fmt"
	"log"
	"time"
)

// SaveTradeInstruction stores a new trade instruction in the database
func SaveTradeInstruction(strategyName string, contractID int32, exchange, symbol, side, orderType, broker string, quantity float64) (int64, error) {
	query := `
	INSERT INTO trades (
		strategy_name, contract_id, exchange, symbol, side, quantity, order_type, broker,
		trading_date, status, created_at, last_updated_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	RETURNING id
	`

	var id int64
	err := db.QueryRow(
		query,
		strategyName,
		contractID,
		exchange,
		symbol,
		side,
		quantity,
		orderType,
		broker,
		time.Now().Format("2006-01-02"), // current trading date
		"Pending",
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to save trade instruction: %v", err)
	}
	return id, nil
}

// UpdateTradeToSubmitted updates a trade record to submitted status with broker order ID
func UpdateTradeToSubmitted(id int64, brokerOrderID int, price float64) error {
	query := `
	UPDATE trades
	SET status = 'Submitted', broker_order_id = $1, price = $2, last_updated_at = $3
	WHERE id = $4
	`
	_, err := db.Exec(query, brokerOrderID, price, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update trade to submitted: %v", err)
	}
	return nil
}

// UpdateTradeStatus updates the status of a trade
func UpdateTradeStatus(brokerOrderID int, status string, filledPrice float64) error {
	// Get current date in YYYY-MM-DD format for the trading day reference
	tradingDate := time.Now().Format("2006-01-02")

	query := `
	UPDATE trades
	SET status = $1, last_updated_at = $2, price = CASE WHEN $1 = 'Filled' THEN $3 ELSE price END
	WHERE broker_order_id = $4 AND trading_date = $5
	`

	result, err := db.Exec(query, status, time.Now(), filledPrice, brokerOrderID, tradingDate)
	if err != nil {
		return fmt.Errorf("failed to update trade status: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %v", err)
	}

	if rows == 0 {
		log.Printf("Warning: No trade found with broker ID %d for date %s", brokerOrderID, tradingDate)
	}

	return nil
}

// GetPendingTrades retrieves all pending trades
func GetPendingTrades() ([]Trade, error) {
	query := `
	SELECT id, strategy_name, contract_id, exchange, symbol, side, quantity,
	       order_type, broker, price, broker_order_id, trading_date, status, created_at, last_updated_at
	FROM trades
	WHERE status IN ('Pending', 'Submitted')
	ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending trades: %v", err)
	}
	defer rows.Close()

	var trades []Trade
	for rows.Next() {
		var trade Trade

		err := rows.Scan(
			&trade.ID, &trade.StrategyName, &trade.ContractID,
			&trade.Exchange, &trade.Symbol, &trade.Side, &trade.Quantity,
			&trade.OrderType, &trade.Broker, &trade.Price, &trade.BrokerOrderID, &trade.TradingDate,
			&trade.Status, &trade.CreatedAt, &trade.LastUpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning trade row: %v", err)
		}

		trades = append(trades, trade)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trade rows: %v", err)
	}

	return trades, nil
}

// GetRecentTradesBySymbol gets recent trades for a specific symbol
func GetRecentTradesBySymbol(symbol string, limit int) ([]Trade, error) {
	query := `
	SELECT id, strategy_name, contract_id, exchange, symbol, side, quantity,
	       order_type, broker, price, broker_order_id, trading_date, status, created_at, last_updated_at
	FROM trades
	WHERE symbol = $1
	ORDER BY created_at DESC
	LIMIT $2
	`

	rows, err := db.Query(query, symbol, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query trades by symbol: %v", err)
	}
	defer rows.Close()

	var trades []Trade
	for rows.Next() {
		var trade Trade

		err := rows.Scan(
			&trade.ID, &trade.StrategyName, &trade.ContractID,
			&trade.Exchange, &trade.Symbol, &trade.Side, &trade.Quantity,
			&trade.OrderType, &trade.Broker, &trade.Price, &trade.BrokerOrderID, &trade.TradingDate,
			&trade.Status, &trade.CreatedAt, &trade.LastUpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning trade row: %v", err)
		}

		trades = append(trades, trade)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trade rows: %v", err)
	}

	return trades, nil
}

// GetTradesByStrategyAndDate gets trades for a specific strategy within a date range
func GetTradesByStrategyAndDate(strategy string, startDate, endDate string) ([]Trade, error) {
	query := `
	SELECT id, strategy_name, contract_id, exchange, symbol, side, quantity,
	       order_type, broker, price, broker_order_id, trading_date, status, created_at, last_updated_at
	FROM trades
	WHERE strategy_name = $1 AND trading_date BETWEEN $2 AND $3
	ORDER BY created_at DESC
	`

	rows, err := db.Query(query, strategy, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query trades by strategy and date: %v", err)
	}
	defer rows.Close()

	var trades []Trade
	for rows.Next() {
		var trade Trade

		err := rows.Scan(
			&trade.ID, &trade.StrategyName, &trade.ContractID,
			&trade.Exchange, &trade.Symbol, &trade.Side, &trade.Quantity,
			&trade.OrderType, &trade.Broker, &trade.Price, &trade.BrokerOrderID, &trade.TradingDate,
			&trade.Status, &trade.CreatedAt, &trade.LastUpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning trade row: %v", err)
		}

		trades = append(trades, trade)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trade rows: %v", err)
	}

	return trades, nil
}
