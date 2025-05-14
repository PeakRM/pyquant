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
	"os/signal"
	"path/filepath"
	"pytrader/database"
	"pytrader/definitions"
	"syscall"

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
	OrderType    string  `json:"order_type"`      // MKT, LMT
	Broker       string  `json:"broker"`          // IB, TDA, etc.
	Price        float64 `json:"price,omitempty"` // Optional price for limit orders
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

	// Convert price string to float64 if provided
	var price float64 = 0.0
	if trade.Price != "" {
		price, err = strconv.ParseFloat(trade.Price, 64)
		if err != nil {
			log.Printf("Warning: Failed to convert price '%s' to float64: %v", trade.Price, err)
			// Continue with price = 0.0
		}
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
		price,
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

// func processNewTrades(workerId int) {
func processNewTrades() {
	for tradeWithID := range tradeChannel {
		trade := tradeWithID.Trade
		tradeID := tradeWithID.TradeID

		// workerInfo := fmt.Sprintf("Worker %d ==>", workerId)

		// Create key for Order
		positionId := fmt.Sprintf("%s-%s", trade.StrategyName, trade.Symbol)

		// deduplication
		i, ok := positions.Load(positionId)
		// If position is found
		if ok {
			//load position to struct
			current_pos, ok1 := i.(definitions.Position)
			if !ok1 {
				fmt.Printf("Issue reading position: i: %s \n cp: %t\n", i, ok1)
				continue
			}

			if current_pos.Status == "Pending" {
				fmt.Printf("Pending order exists, trade skipped: %s - %s - %t \n", trade, i, ok)
				continue
			}
		}
		// log.Printf("%sProcessing trade: %s\n", workerInfo, trade)
		var lmtPrice float64 = 0.0 // Limit price for limit orders

		// Check if price is provided in the trade instruction
		if trade.Price != "" {
			// Convert price string to float64
			var err error
			lmtPrice, err = strconv.ParseFloat(trade.Price, 64)
			if err != nil {
				log.Printf("Failed to convert price '%s' to float64: %v", trade.Price, err)
				// Fall back to fetching price if conversion fails
				lmtPrice = 0.0
			} else {
				log.Printf("Using provided price: %f\n", lmtPrice)
			}
		}

		// If price is not provided or conversion failed, and it's not a market order, fetch price
		if trade.OrderType == "MKT" {
			log.Printf("Market order, skipping price quote: %s\n", trade)
			lmtPrice = 0.0
		} else if lmtPrice != 0.0 {
			log.Printf("Using provided price: %f\n", lmtPrice)
		} else {
			// Fetch price quote
			quote, err := fetchPriceQuote(trade.ContractId, trade.Exchange, trade.Broker)
			if err != nil {
				log.Printf("Failed to fetch price for symbol %s: %v", trade.Symbol, err)
				continue
			}
			lmtPrice = quote.Bid
			if trade.Side == "SELL" {
				lmtPrice = quote.Ask
			}
		}

		quantity, err := strconv.ParseFloat(trade.Quantity, 64)
		if err != nil {
			log.Printf("Failed to convert Quantity string to float64 for symbol %s: %v", trade.Quantity, err)
			continue
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
				Price:        lmtPrice, // Include the price in the trade instruction
			},
			PriceQuote: lmtPrice,
			Timestamp:  time.Now(),
		}

		// Send order
		orderId, err := transmitOrder(order, false)
		if err != nil {
			log.Printf("Failed to submit order for strategy-symbol %s-%s: %v", trade.StrategyName, trade.Symbol, err)
			continue
		}

		// Update the trade record with the broker order ID
		if tradeID > 0 {
			err = database.UpdateTradeToSubmitted(tradeID, orderId, lmtPrice)
			if err != nil {
				log.Printf("Warning: Failed to update trade status to Submitted in database: %v", err)
			}
		}

		// Save Order Id received from API call to broker
		orderResponse := OrderResponse{
			Order:   order,
			OrderId: orderId,
		}
		updatePositionsToPending(orderResponse)
		log.Println("Sending order response to channel")
		orderResponseChannel <- &orderResponse

		// go monitorFill(orderResponse)
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
			fmt.Printf("Order %s: %d\n", trade.Status, orderResp.OrderId)

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
				log.Printf("Warning: Failed to update trade status to Filled/Cancelled in database: %v\n", err)
			}

			updatePositionsFromResponse(orderResp, trade.Status, trade.Price,
				int(direction*math.Abs(float64(trade.Quantity))))
			fmt.Printf("Order %s: %d -- %f\n", trade.Status, orderResp.OrderId, trade.Price)
			isFilled = true
		}
		time.Sleep(time.Second)
	}
}

func sendOrdersToFillMonitor() {
	for orderResponse := range orderResponseChannel {
		log.Println("Transfering order response to check for Fills")
		key := fmt.Sprintf("%v-%d", orderResponse.Order.Timestamp, orderResponse.OrderId)
		orderResponseQueue.Store(key, orderResponse)
	}
}

func queryTradesAtBroker() []Trade {
	url := "http://broker_api:8000/api/IB/trades"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error sending GET request:", err)
		return nil
	}
	defer resp.Body.Close()

	// Parse the response body
	var response []Trade
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Println("Error decoding fills:", err)
		return nil
	}
	return response
}

type MatchedTrades struct {
	OrderResponse OrderResponse
	Trade
}

// returns the intersection of trade list an dorder responses
func findOrderInTrades(slice1 []Trade, slice2 *sync.Map) []MatchedTrades {
	intersection := []MatchedTrades{}
	slice2.Range(func(key, value interface{}) bool {

		val2, ok := value.(*OrderResponse) // Type assertion for the value from sync.Map
		if !ok {
			log.Printf("Unable to assert Order Response: %v (type: %T)\n", value, value)
			return true
		}

		for _, val1 := range slice1 {
			log.Printf("Matching Trade -> OrderID %d, Trade ID: %d,  Trade Status: %s", val2.OrderId, val1.Id, val1.Status)

			if val1.Id == val2.OrderId && val1.Status == "Filled" {
				log.Println("Trade Trades")

				intersection = append(intersection, MatchedTrades{OrderResponse: *val2, Trade: val1})
				break // Avoid duplicates from slice2
			}
		}
		return true
	})

	return intersection
}

func monitorFills(done chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			//if order resposne queue is not empty
			// if len(orderResponseQueue) != 0 {
			var trades []Trade
			// orderResponseQueue.Range(func(key, value any) {

			// query borker api for list of trades
			trades = queryTradesAtBroker()
			// check for orderIds in Trades
			ordersFoundInTrades := findOrderInTrades(trades, &orderResponseQueue)
			// for each order filled, Update system state
			for _, order := range ordersFoundInTrades {
				log.Println("Reconciling Trades")

				fmt.Printf("Order Filled: %d\n", order.OrderResponse.OrderId)
				// check for orderrespos.id
				direction := 1.0
				if order.OrderResponse.Order.TradeInstruction.Side == "SELL" {
					direction = -1.0
				}

				// Update trade status in database
				err := database.UpdateTradeStatus(
					order.OrderResponse.OrderId,
					order.Trade.Status,
					order.Trade.Price,
				)
				if err != nil {
					log.Printf("Warning: Failed to update trade status to Filled/Cancelled in database: %v\n", err)
				}
				// update positions json
				updatePositionsFromResponse(order.OrderResponse,
					order.Trade.Status,
					order.Trade.Price,
					int(direction*math.Abs(float64(order.Trade.Quantity))))
				fmt.Printf("Order %s: %d -- %f\n", order.Trade.Status,
					order.OrderResponse.OrderId,
					order.Trade.Price)

				// remove from orderResponse queue
				orq_key := fmt.Sprintf("%v-%d", order.OrderResponse.Order.Timestamp, order.OrderResponse.OrderId)
				orderResponseQueue.Delete(orq_key) // change this to map[int]OrderResponse

			}
		}
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

// Create a single shutdown function
func shutdown() {
	// Signal all goroutines to terminate
	close(done)

	// Save positions to file
	if err := SyncMapToJSONFile(&positions, GetSharedFilePath("positions.json")); err != nil {
		log.Printf("Error saving positions: %v", err)
	}

	// Close channels
	close(tradeChannel)
	close(orderResponseChannel)

	// Give goroutines time to finish
	time.Sleep(1 * time.Second)

	// Database is already being closed with defer in main()
}

var tradeChannel = make(chan *TradeWithID, 100)           // Buffered channel for trades
var orderResponseChannel = make(chan *OrderResponse, 100) // Channel for order response pointers
var orderResponseQueue sync.Map                           // map[int]OrderResponse
type poolFunction func(int)

var done = make(chan struct{})

var positions sync.Map // hols positions

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
	// startWorkerPool(5, processNewTrades) //could be used elsewhere, when iterating in new fill monitor
	go processNewTrades()
	go sendOrdersToFillMonitor()
	go monitorFills(done)
	// Start the gRPC server
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterTradeServiceServer(grpcServer, &server{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// Start a goroutine to handle shutdown
	go func() {
		<-sigChan
		log.Println("Shutting down...")
		shutdown()
		grpcServer.GracefulStop()
		os.Exit(0)
	}()
	log.Println("Server is running on port 50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
