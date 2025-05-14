package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

// Initialize sets up the database connection
func Initialize() error {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "tradeuser")
	password := getEnv("DB_PASSWORD", "tradepass")
	dbname := getEnv("DB_NAME", "tradedb")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	// Connect with a retry mechanism
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

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Create tables if they don't exist
	err = createTables()
	if err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	log.Println("Database connection established successfully")
	return nil
}

// CreateTables creates the necessary tables if they don't exist
func createTables() error {
	// Create simplified trades table
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS trades (
		id SERIAL PRIMARY KEY,
		strategy_name VARCHAR(100) NOT NULL,
		contract_id INTEGER NOT NULL,
		exchange VARCHAR(50) NOT NULL,
		symbol VARCHAR(50) NOT NULL,
		side VARCHAR(10) NOT NULL,
		quantity FLOAT NOT NULL,
		order_type VARCHAR(10) NOT NULL,
		broker VARCHAR(20) NOT NULL,
		price FLOAT NOT NULL DEFAULT 0,
		broker_order_id INTEGER NOT NULL DEFAULT 0,
		trading_date VARCHAR(10) NOT NULL,
		status VARCHAR(20) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		last_updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE UNIQUE INDEX IF NOT EXISTS trades_broker_order_id_trading_date_idx
	ON trades (broker_order_id, trading_date)
	WHERE broker_order_id > 0;
	`)
	if err != nil {
		return err
	}

	// Create indexes for better query performance
	_, err = db.Exec(`
	CREATE INDEX IF NOT EXISTS idx_trades_status ON trades(status);
	CREATE INDEX IF NOT EXISTS idx_trades_symbol ON trades(symbol);
	CREATE INDEX IF NOT EXISTS idx_trades_strategy ON trades(strategy_name);
	CREATE INDEX IF NOT EXISTS idx_trades_created ON trades(created_at);
	`)

	return err
}

// GetDB returns the database instance
func GetDB() *sql.DB {
	return db
}

// Close closes the database connection
func Close() {
	if db != nil {
		db.Close()
	}
}

// Helper function to get environment variable with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
