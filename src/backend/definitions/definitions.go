package definitions

const BROKER_API = "http://127.0.0.1:8000"

// Define the Trade struct to match the Python model
type Trade struct {
	StrategyName string      `json:"strategy_name"`
	ContractID   int         `json:"contract_id"`
	Exchange     string      `json:"exchange"`
	Symbol       string      `json:"symbol"`
	Side         string      `json:"side"`            // Literal['BUY', 'SELL', 'HOLD']
	Quantity     interface{} `json:"quantity"`        // Can be int or float
	OrderType    string      `json:"order_type"`      // Literal['MKT', 'LMT']
	Broker       string      `json:"broker"`          // Literal['IB', 'TDA', etc.]
	Price        float64     `json:"price,omitempty"` // Optional price for limit orders
}

type Position struct {
	Symbol     string  `json:"symbol"`
	Exchange   string  `json:"exchange"`
	Quantity   int     `json:"quantity"`
	CostBasis  float64 `json:"cost_basis"`
	Datetime   string  `json:"datetime"`
	ContractID int     `json:"contract_id"`
	Status     string  `json:"status"`
}

type Contract struct {
	Symbol       string `json:"symbol"`
	ContractType string `json:"contract_type"`
	Exchange     string `json:"exchange"`
	Currency     string `json:"currency"`
	Expiry       string `json:"expiry"`
}

// // Struct to hold shared state
// type Tracker struct {
// 	mu      sync.Mutex
// 	seenSet map[string]bool // Tracks objects already added to the channel
// }

// func NewTracker() *Tracker {
// 	return &Tracker{
// 		seenSet: make(map[string]bool),
// 	}
// }

// func (t *Tracker) AddIfNotSeen(trade Trade) bool {
// 	t.mu.Lock()
// 	defer t.mu.Unlock()

// 	// Create a unique key based on object properties
// 	key := fmt.Sprintf("%s:%s", trade.StrategyName, trade.Symbol)

// 	if t.seenSet[key] {
// 		return false // Object already seen, don't add
// 	}

// 	// Mark object as seen
// 	t.seenSet[key] = true
// 	return true
// }

// func (t *Tracker) Remove(trade Trade) {
// 	t.mu.Lock()
// 	defer t.mu.Unlock()

// 	key := fmt.Sprintf("%s:%s", trade.StrategyName, trade.Symbol)
// 	delete(t.seenSet, key) // Remove from the tracker
// }
