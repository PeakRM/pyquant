package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"pytrader/database"
	"pytrader/definitions"

	pb "pytrader/tradepb"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
)

// server is used to implement TradeService
type server struct {
	pb.UnimplementedTradeServiceServer
}

type OrderResponse struct {
	Order   Order
	OrderId int
}

type TradeInstruction struct {
	StrategyName string  `json:"strategy_name"`
	ContractId   int     `json:"contract_id"`
	Exchange     string  `json:"exchange"`
	Symbol       string  `json:"symbol"`
	Side         string  `json:"side"`
	Quantity     float64 `json:"quantity"`
	OrderType    string  `json:"order_type"` // MKT, LMT
	Broker       string  `json:"broker"`     // IB, TDA, etc.
}

type Order struct {
	TradeInstruction TradeInstruction `json:"trade"`
	PriceQuote       float64          `json:"price"`
	Timestamp        time.Time        `json:"timestamp"`
}

type Trade struct {
	Id         int       `json:"order_id"`
	Price      float64   `json:"price"`
	Quantity   float64   `json:"quantity"`
	Time       time.Time `json:"time"`
	ContractId int       `json:"contract_id"`
	Side       string    `json:"side"`
	Status     string    `json:"order_status"`
}

type Quote struct {
	Symbol    string  `json:"symbol"`
	Bid       float64 `json:"bid"`
	Ask       float64 `json:"ask"`
	Last      float64 `json:"last"`
	Timestamp string  `json:"timestamp"`
}
type MyError struct{}

func (m *MyError) Error() string {
	return "Failed to get price."
}

// Add a struct to carry the trade along with its database ID
type TradeWithID struct {
	Trade   *pb.Trade
	TradeID int64
}

// SendTrade implements the SendTrade RPC
func (s *server) SendTrade(ctx context.Context, trade *pb.Trade) (*pb.TradeResponse, error) {
	log.Printf("Received trade: %+v", trade)

	// Convert quantity string to float64
	quantity, err := strconv.ParseFloat(trade.Quantity, 64)
	if err != nil {
		log.Printf("Failed to convert quantity '%s' to float64: %v", trade.Quantity, err)
		return &pb.TradeResponse{Status: "Error: Invalid quantity"}, err
	}

	// Save trade instruction to database
	tradeID, err := database.SaveTradeInstruction(
		trade.StrategyName,
		trade.ContractId,
		trade.Exchange,
		trade.Symbol,
		trade.Side,
		trade.OrderType,
		trade.Broker,
		quantity,
	)

	if err != nil {
		log.Printf("Error saving trade instruction: %v", err)
		// Continue processing anyway - we don't want to block the trade
	}

	// Store the trade ID for later use in the channel
	tradeWithID := &TradeWithID{
		Trade:   trade,
		TradeID: tradeID,
	}

	// Send trade to the processing channel
	tradeChannel <- tradeWithID

	return &pb.TradeResponse{Status: "Trade received and processing"}, nil
}

// Function to send a GET request to retrieve the last price
func fetchPriceQuote(contractID int32, exchange string, broker string) (Quote, error) {
	// Default to IB if broker is not specified
	if broker == "" {
		broker = "IB"
	}

	// Determine the URL based on environment
	var baseURL string
	if os.Getenv("ENVIRONMENT") == "production" || os.Getenv("ENVIRONMENT") == "docker" {
		baseURL = "http://broker_api:8000"
	} else {
		baseURL = "http://127.0.0.1:8000"
	}

	url := fmt.Sprintf("%s/api/%s/quote/%s/%d", baseURL, broker, exchange, contractID)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error sending GET request:", err)
		return Quote{}, err // Empty quote if there is an error
	}
	defer resp.Body.Close()

	// Parse the response body to extract the price
	var response Quote
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Println("Error decoding quote:", err)
		return Quote{}, err // Empty quote if there is an error
	}

	fmt.Printf("Bid: %f\tAsk: %f\tLast:%f\n", response.Bid, response.Ask, response.Last)
	if response.Last == 0.0 {
		return Quote{}, &MyError{}
	}
	return response, nil
}

// Send order to BrokerAPI
func transmitOrder(order Order, testTrade bool) (int, error) {
	if testTrade {
		fmt.Println("Test Trade --> ")
		return rand.Intn(1000), nil
	}

	// Use the broker from the order, default to IB if not specified
	broker := order.TradeInstruction.Broker
	if broker == "" {
		broker = "IB"
	}

	// Determine the URL based on environment
	var baseURL string
	if os.Getenv("ENVIRONMENT") == "production" || os.Getenv("ENVIRONMENT") == "docker" {
		baseURL = "http://broker_api:8000"
	} else {
		baseURL = "http://127.0.0.1:8000"
	}

	url := fmt.Sprintf("%s/api/%s/order", baseURL, broker)
	orderJSON, err := json.Marshal(order)
	if err != nil {
		fmt.Println("Error marshaling order to JSON:", err)
		return 0, err
	}
	fmt.Println("Order Spec: ", string(orderJSON))

	// req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(orderJSON)))
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(orderJSON))
	if err != nil {
		fmt.Println("Error creating POST request:", err)
		return 0, err
	}

	// Set headers for the POST request
	req.Header.Add("Content-Type", "application/json")
	fmt.Println("Transmitting Order...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending POST request:", err)
		return 0, err
	}
	defer resp.Body.Close()

	fmt.Printf("POST request to %s completed with status: %s\n", url, resp.Status)
	// read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return 0, err

	}

	var orderIDStr string
	if err := json.Unmarshal(body, &orderIDStr); err != nil {
		return 0, fmt.Errorf("error unmarshaling response: %v", err)
	}

	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		return 0, fmt.Errorf("error converting to int: %v", err)
	}
	fmt.Println("Order Sent\t-->\tID: ", orderID)

	return orderID, nil
}
func processNewTrades(workerId int) {
	for tradeWithID := range tradeChannel {
		trade := tradeWithID.Trade
		tradeID := tradeWithID.TradeID

		workerInfo := fmt.Sprintf("Worker %d ==>", workerId)

		// Create key for Order
		positionId := fmt.Sprintf("%s-%s", trade.StrategyName, trade.Symbol)

		// deduplication
		i, ok := positions.Load(positionId)
		// If position is found
		if ok {
			//load position to struct
			current_pos, ok1 := i.(definitions.Position)
			if !ok1 {
				fmt.Printf("%sIssue reading position: i: %s \n cp: %t\n", workerInfo, i, ok1)
				continue
			}

			if current_pos.Status == "Pending" {
				fmt.Printf("%sPending order exists, trade skipped: %s - %s - %t \n", workerInfo, trade, i, ok)
				continue
			}
		}
		log.Printf("%sProcessing trade: %s\n", workerInfo, trade)
		var quote Quote = Quote{} // Initialize to empty struct
		var quantity float64 = 0.0
		var lmtPrice float64 = 0.0 // Limit price for limit orders
		if trade.OrderType == "MKT" {
			log.Printf("%sMarket order, skipping price quote: %s\n", workerInfo, trade)
			quote.Bid = 0.0
			quote.Ask = 0.0
			quote.Last = 0.0
		} else{
			// Fetch price quote
			
			quote, err := fetchPriceQuote(trade.ContractId, trade.Exchange, trade.Broker)
			if err != nil {
				log.Printf("%sFailed to fetch price for symbol %s: %v", workerInfo, trade.Symbol, err)
				continue
			}
			quantity, err := strconv.ParseFloat(trade.Quantity, 64)
			if err != nil {
				log.Printf("%sFailed to convert Quantity string to float64 for symbol %s: %v", workerInfo, trade.Quantity, err)
				continue
			}
			lmtPrice := quote.Bid
			if trade.Side == "SELL" {
				lmtPrice = quote.Ask
			}
		}
		// Create order
		order := Order{
			TradeInstruction: TradeInstruction{
				StrategyName: trade.StrategyName,
				ContractId:   int(trade.ContractId),
				Exchange:     trade.Exchange,
				Symbol:       trade.Symbol,
				Side:         trade.Side,
				Quantity:     quantity,
				OrderType:    trade.OrderType,
				Broker:       trade.Broker,
			},
			PriceQuote: lmtPrice,
			Timestamp:  time.Now(),
		}

		// Send order
		orderId, err := transmitOrder(order, false)
		if err != nil {
			log.Printf("%sFailed to submit order for strategy-symbol %s-%s: %v", workerInfo, trade.StrategyName, trade.Symbol, err)
			continue
		}

		// Update the trade record with the broker order ID
		if tradeID > 0 {
			err = database.UpdateTradeToSubmitted(tradeID, orderId, lmtPrice)
			if err != nil {
				log.Printf("%sWarning: Failed to update trade status to Submitted in database: %v", workerInfo, err)
			}
		}

		// Save Order Id received from API call to broker
		orderResponse := OrderResponse{
			Order:   order,
			OrderId: orderId,
		}
		updatePositionsToPending(orderResponse)

		go monitorFill(orderResponse)
	}
}

func startWorkerPool(numWorkers int, f poolFunction) {
	for i := 0; i < numWorkers; i++ {
		go f(i)
	}
}
func monitorFill(orderResp OrderResponse) {
	fmt.Println("Monitoring fill")
	isFilled := false
	for !isFilled {
		url := "http://broker_api:8000/api/IB/trades"
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Error sending GET request:", err)
			continue
		}
		defer resp.Body.Close()

		// Parse the response body
		var response []Trade
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			fmt.Println("Error decoding fills:", err)
		}

		for _, trade := range response {
			fmt.Printf("Order %s: %d", trade.Status, orderResp.OrderId)

			if trade.Id != orderResp.OrderId {
				continue
			}
			if trade.Status != "Filled" && trade.Status != "Cancelled" {
				continue
			}

			direction := 1.0
			if orderResp.Order.TradeInstruction.Side == "SELL" {
				direction = -1.0
			}

			// Update trade status in database
			err := database.UpdateTradeStatus(
				orderResp.OrderId,
				trade.Status,
				trade.Price,
			)
			if err != nil {
				log.Printf("Warning: Failed to update trade status to Filled/Cancelled in database: %v", err)
			}

			updatePositionsFromResponse(orderResp, trade.Status, trade.Price,
				int(direction*math.Abs(float64(trade.Quantity))))
			fmt.Printf("Order %s: %d", trade.Status, orderResp.OrderId)
			isFilled = true
		}
		time.Sleep(time.Second)
	}
}

func updatePositionsFromResponse(orderResp OrderResponse, status string, costBasis float64, quantity int) {
	fmt.Printf("Updating Positions for %s Order\n", status)
	positionId := fmt.Sprintf("%s-%s",
		orderResp.Order.TradeInstruction.StrategyName,
		orderResp.Order.TradeInstruction.Symbol)
	positionMap, ok := positions.Load(positionId)
	if ok {
		pos, ok := positionMap.(definitions.Position)
		if ok {
			fmt.Print("Position Map", pos)
			quantity += pos.Quantity
		}

	}
	if quantity == 0.0 || status == "Cancelled" {
		status = "Closed"
	}

	positions.Store(positionId, definitions.Position{
		Symbol:     orderResp.Order.TradeInstruction.Symbol,
		Exchange:   orderResp.Order.TradeInstruction.Exchange,
		Quantity:   quantity, // * float64(posAdj),
		CostBasis:  costBasis,
		Datetime:   time.Now().String(),
		ContractID: int(orderResp.Order.TradeInstruction.ContractId),
		Status:     status,
	})
	shared_positions := GetSharedFilePath("positions.json")
	// Marshal to JSON file
	if err := SyncMapToJSONFile(&positions, shared_positions); err != nil {
		fmt.Println("Error marshalling sync.Map to JSON:", err)
		return
	}

}
func updatePositionsToPending(orderResp OrderResponse) {
	fmt.Println("Updating Positions for Pending Order")
	positionId := fmt.Sprintf("%s-%s",
		orderResp.Order.TradeInstruction.StrategyName,
		orderResp.Order.TradeInstruction.Symbol)

	p, ok := positions.Load(positionId)
	if !ok {
		fmt.Println("Positon does not exist")

		positions.Store(positionId, definitions.Position{
			Symbol:     orderResp.Order.TradeInstruction.Symbol,
			Exchange:   orderResp.Order.TradeInstruction.Exchange,
			Quantity:   0.0,
			CostBasis:  0.0,
			Datetime:   time.Now().String(),
			ContractID: int(orderResp.Order.TradeInstruction.ContractId),
			Status:     "Pending",
		})
	} else {
		p, _ := p.(definitions.Position)
		if !ok {
			fmt.Println("Could not assert Position type on p:", p)
		}
		p.Status = "Pending"
		positions.Store(positionId, p)
	}
	shared_positions := GetSharedFilePath("positions.json")
	// Marshal to JSON file
	if err := SyncMapToJSONFile(&positions, shared_positions); err != nil {
		fmt.Println("Error marshalling sync.Map to JSON:", err)
		return
	}

}

// GetSharedFilePath returns the appropriate path based on environment
func GetSharedFilePath(filename string) string {
	// Check if running in container by looking for /.dockerenv
	if os.Getenv("ENVIRONMENT") == "production" || os.Getenv("ENVIRONMENT") == "docker" {
		return filepath.Join("/shared", filename)
	}

	// Development environment
	return filepath.Join("..", "..", "shared_files", filename)
}

// SyncMapFromJSONFile unmarshals a JSON file into a sync.Map.
func SyncMapFromJSONFile(m *sync.Map, filename string) error {
	// alternative here:https://stackoverflow.com/questions/46390409/how-to-decode-json-strings-to-sync-map-instead-of-normal-map-in-go1-9

	byteValue, err := ioutil.ReadFile(filename)
	if err != nil || len(byteValue) == 0 {
		return err
	}

	normalMap := make(map[string]json.RawMessage)
	// Unmarshal JSON data into a slice of Trade structs
	err = json.Unmarshal(byteValue, &normalMap)
	if err != nil {
		return err
	}

	// Store each key-value pair back into the sync.Map
	for k, v := range normalMap {
		var pos definitions.Position
		err := json.Unmarshal(v, &pos)
		if err != nil {
			fmt.Printf("Error unmarshaling position for key %s: %v\n", k, err)
			continue
		}
		m.Store(k, pos)
	}
	return nil
}

// SyncMapToJSONFile marshals a sync.Map to a JSON file.
func SyncMapToJSONFile(m *sync.Map, filename string) error {

	normalMap := make(map[string]interface{})
	// Convert sync.Map to a normal map
	m.Range(func(key, value interface{}) bool {
		strKey, ok := key.(string)
		if !ok {
			// If keys are not strings, you may choose how to handle it.
			// Here we skip non-string keys.
			return true
		}
		normalMap[strKey] = value
		return true
	})

	// Write updated data back to file
	data, err := json.MarshalIndent(normalMap, "", "  ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		// fmt.Println("Error writing to file:", err)
		return err
	}
	return nil
}

var tradeChannel = make(chan *TradeWithID, 100) // Buffered channel for trades
type poolFunction func(int)

var positions sync.Map //

func main() {
	// Initialize database connection
	err := database.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()
	// Clear the original map to demonstrate loading from file
	shared_positions := GetSharedFilePath("positions.json")

	// Unmarshal from JSON file
	if err := SyncMapFromJSONFile(&positions, shared_positions); err != nil {
		fmt.Println("Error unmarshalling JSON to sync.Map:", err)
		return
	}
	// Start the trade processing worker
	startWorkerPool(5, processNewTrades)

	// Start the gRPC server
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterTradeServiceServer(grpcServer, &server{})

	log.Println("Server is running on port 50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
