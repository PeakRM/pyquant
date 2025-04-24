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

// Trade represents a trade record from the database with only essential fields
type Trade struct {
	ID           int64     `json:"id"`
	StrategyName string    `json:"strategy_name"`
	Symbol       string    `json:"symbol"`
	Exchange     string    `json:"exchange"`
	Side         string    `json:"side"`
	Quantity     int       `json:"quantity"`
	Price        float64   `json:"price"`
	Status       string    `json:"status"`
	BrokerOrderID int      `json:"broker_order_id"`
	UpdatedAt    time.Time `json:"updated_at"`
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

// SSETradesHandler handles the SSE endpoint for streaming recent trades
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

	// Keep the connection open until client disconnects
	for {
		select {
		case <-clientClosed:
			log.Println("Client closed connection")
			return
		default:
			// Fetch trades from the last 24 hours
			trades, err := fetchRecentTrades()
			if err != nil {
				log.Printf("Error fetching trades: %v", err)
				// Send error event to client
				fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
				flusher.Flush()
				time.Sleep(5 * time.Second) // Wait before retrying
				continue
			}

			// Convert trades to a map with trade ID as key for the front-end
			tradesMap := make(map[string]Trade)
			for _, trade := range trades {
				// Use ID as string key
				key := fmt.Sprintf("trade-%d", trade.ID)
				tradesMap[key] = trade
			}

			// Convert the map to JSON
			tradesJSON, err := json.Marshal(tradesMap)
			if err != nil {
				log.Printf("Error marshaling trades map: %v", err)
				continue
			}

			// Send the event with all trades at once
			fmt.Fprintf(w, "data: %s\n\n", tradesJSON)
			flusher.Flush()

			// Send a heartbeat to keep connection alive
			fmt.Fprintf(w, "event: heartbeat\ndata: %s\n\n", time.Now().Format(time.RFC3339))
			flusher.Flush()

			// Wait before polling again
			time.Sleep(2 * time.Second)
		}
	}
}

// fetchRecentTrades retrieves trades from the last 24 hours
func fetchRecentTrades() ([]Trade, error) {
	// Calculate timestamp for 24 hours ago
	oneDayAgo := time.Now().Add(-24 * time.Hour)

	// Query to get trades from the last 24 hours
	query := `
		SELECT id, strategy_name, exchange, symbol, side, quantity,
			price, broker_order_id, status, last_updated_at
		FROM trades
		WHERE last_updated_at >= $1
		ORDER BY last_updated_at DESC
	`

	// Execute query
	rows, err := db.Query(query, oneDayAgo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Process results
	var trades []Trade
	for rows.Next() {
		var t Trade
		var updatedAt time.Time

		err := rows.Scan(
			&t.ID, &t.StrategyName, &t.Exchange, &t.Symbol,
			&t.Side, &t.Quantity, &t.Price, &t.BrokerOrderID,
			&t.Status, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

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
