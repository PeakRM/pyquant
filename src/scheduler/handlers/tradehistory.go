package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// Trade represents a trade record from the database
type Trade struct {
	ID            int64     `json:"id"`
	StrategyName  string    `json:"strategy_name"`
	ContractID    int       `json:"contract_id"`
	Exchange      string    `json:"exchange"`
	Symbol        string    `json:"symbol"`
	Side          string    `json:"side"`
	Quantity      int       `json:"quantity"`
	Price         float64   `json:"price"`
	BrokerOrderID int       `json:"broker_order_id"`
	TradingDate   string    `json:"trading_date"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

var db *sql.DB

// InitDB initializes the database connection
func InitDB() error {
	host := getEnv("DB_HOST", "postgres")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "tradeuser")
	password := getEnv("DB_PASSWORD", "tradepass")
	dbname := getEnv("DB_NAME", "tradedb")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	// Connect with retries to allow database to initialize
	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Printf("Failed to connect to DB, retrying in 3 seconds... (attempt %d/10)", i+1)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 5)

	log.Println("Database connection established")
	return nil
}

// CloseDB closes the database connection
func CloseDB() {
	if db != nil {
		db.Close()
	}
}

// SSETradesHandler handles the SSE endpoint for streaming today's trades
func SSETradesHandler(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*") // For CORS

	// Get the client's flusher to flush data in chunks
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Create a channel to signal client disconnection
	clientClosed := r.Context().Done()

	// Query parameter for filtering by strategy or symbol
	strategy := r.URL.Query().Get("strategy")
	symbol := r.URL.Query().Get("symbol")

	// Start with an initial fetch of today's trades
	today := time.Now().Format("2006-01-02")
	// lastID := int64(0)

	// Keep the connection open until client disconnects
	for {
		select {
		case <-clientClosed:
			log.Println("Client closed connection")
			return
		default:
			// Fetch all trades
			// trades, err := fetchNewTrades(today, lastID, strategy, symbol)
			trades, err := fetchTrades(today, strategy, symbol)
			if err != nil {
				log.Printf("Error fetching trades: %v", err)
				// Send error event to client
				fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
				flusher.Flush()
				time.Sleep(5 * time.Second) // Wait before retrying
				continue
			}

			// Send each trade as an SSE event
			for _, trade := range trades {
				// if trade.ID > lastID {
				// 	lastID = trade.ID
				// }

				// Convert trade to JSON
				tradeJSON, err := json.Marshal(trade)
				if err != nil {
					log.Printf("Error marshaling trade: %v", err)
					continue
				}

				// Send the event
				fmt.Fprintf(w, "event: trade\ndata: %s\n\n", tradeJSON)
				flusher.Flush()
			}

			// If requested, send a heartbeat to keep connection alive
			fmt.Fprintf(w, "event: heartbeat\ndata: %s\n\n", time.Now().Format(time.RFC3339))
			flusher.Flush()

			// Wait before polling again
			time.Sleep(2 * time.Second)
		}
	}
}

// fetchNewTrades retrieves trades from today that have an ID greater than lastID
func fetchNewTrades(date string, lastID int64, strategy, symbol string) ([]Trade, error) {
	var query string
	var args []interface{}

	// Base query
	baseQuery := `
		SELECT id, strategy_name, contract_id, exchange, symbol, side, quantity,
			price, broker_order_id, trading_date, status, created_at, last_updated_at
		FROM trades
		WHERE trading_date = $1 AND id > $2
	`

	// Add filters based on parameters
	if strategy != "" && symbol != "" {
		query = baseQuery + " AND strategy_name = $3 AND symbol = $4 ORDER BY id ASC"
		args = []interface{}{date, lastID, strategy, symbol}
	} else if strategy != "" {
		query = baseQuery + " AND strategy_name = $3 ORDER BY id ASC"
		args = []interface{}{date, lastID, strategy}
	} else if symbol != "" {
		query = baseQuery + " AND symbol = $3 ORDER BY id ASC"
		args = []interface{}{date, lastID, symbol}
	} else {
		query = baseQuery + " ORDER BY id ASC"
		args = []interface{}{date, lastID}
	}

	// Execute query
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Process results
	var trades []Trade
	for rows.Next() {
		var t Trade
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&t.ID, &t.StrategyName, &t.ContractID, &t.Exchange, &t.Symbol,
			&t.Side, &t.Quantity, &t.Price, &t.BrokerOrderID, &t.TradingDate,
			&t.Status, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

		t.CreatedAt = createdAt
		t.UpdatedAt = updatedAt
		trades = append(trades, t)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return trades, nil
}

// s retrieves trades from today
func fetchTrades(date string, strategy, symbol string) ([]Trade, error) {
	var query string
	var args []interface{}

	// Base query
	baseQuery := `
		SELECT id, strategy_name, contract_id, exchange, symbol, side, quantity,
			price, broker_order_id, trading_date, status, created_at, last_updated_at
		FROM trades
		WHERE trading_date = $1
	`

	// Add filters based on parameters
	if strategy != "" && symbol != "" {
		query = baseQuery + " AND strategy_name = $3 AND symbol = $4 ORDER BY id ASC"
		args = []interface{}{date, strategy, symbol}
	} else if strategy != "" {
		query = baseQuery + " AND strategy_name = $3 ORDER BY id ASC"
		args = []interface{}{date, strategy}
	} else if symbol != "" {
		query = baseQuery + " AND symbol = $3 ORDER BY id ASC"
		args = []interface{}{date, symbol}
	} else {
		query = baseQuery + " ORDER BY id ASC"
		args = []interface{}{date}
	}

	// Execute query
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Process results
	var trades []Trade
	for rows.Next() {
		var t Trade
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&t.ID, &t.StrategyName, &t.ContractID, &t.Exchange, &t.Symbol,
			&t.Side, &t.Quantity, &t.Price, &t.BrokerOrderID, &t.TradingDate,
			&t.Status, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

		t.CreatedAt = createdAt
		t.UpdatedAt = updatedAt
		trades = append(trades, t)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return trades, nil
}

// Helper function to get environment variable with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
