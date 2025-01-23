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

type Order struct {
	Trade      *pb.Trade // The incoming trade
	PriceQuote float64   // Price quote fetched from the API
	Timestamp  time.Time // Time the order was created
}

type Fill struct {
	Id       int       `json:"order_id"`
	Price    float64   `json:"price"`
	Quantity float64   `json:"quantity"`
	Time     time.Time `json:"time"`
	// Status   string  `json:"status"`
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

// SendTrade implements the SendTrade RPC
func (s *server) SendTrade(ctx context.Context, trade *pb.Trade) (*pb.TradeResponse, error) {
	log.Printf("Received trade: %+v", trade)

	// Send trade to the processing channel
	tradeChannel <- trade

	return &pb.TradeResponse{Status: "Trade received and processing"}, nil
}

// Function to send a GET request to localhost:8081/api/[contract_id] and retrieve the last price
func fetchPriceQuote(contractID int32, exchange string) (float64, error) {
	// url := fmt.Sprintf("http://127.0.0.1:8000/quoteByConId?conId=%d&exchange=%s", contractID, exchange) //local dev test_api
	// url := fmt.Sprintf("http://broker_api:8000/quoteByConId?conId=%d&exchange=%s", contractID, exchange) // docker test test_api
	url := fmt.Sprintf("http://127.0.0.1:8000/api/IB/quote/%s/%d", exchange, contractID) //local dev broker_api
	// url := fmt.Sprintf("http://broker_api:8000/api/IB/quote/%s/%d", exchange, contractID) //docker dev broker_api

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error sending GET request:", err)
		return -1.0, err // Default price if there is an error
	}
	defer resp.Body.Close()

	// Parse the response body to extract the price
	// var response struct {
	// 	Price float64 `json:"price"`
	// }
	var response Quote
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Println("Error decoding quote:", err)
		return -1.0, err // Default price if there is an error
	}

	fmt.Printf("GET request to %s returned price: %f\n", url, response.Last)
	if response.Last == 0.0 {
		return -1.0, &MyError{}
	}
	return response.Last, nil
}

// Send order to BrokerAPI
func transmitOrder(order Order, testTrade bool) (int, error) {
	if testTrade {
		fmt.Println("Test Trade --> ")
		return rand.Intn(1000), nil
	}
	// url := "http://127.0.0.1:8000/placeLimitOrder?broker=IB"
	url := "http://127.0.0.1:8000/api/IB/order"
	// url := "http://broker_api:8000/placeLimitOrder?broker=IB"
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
	// print response body
	//convert string to int
	n, err := strconv.Atoi(string(body))
	//check if error occured
	if err != nil {
		//executes if there is any error
		fmt.Println(err)
		return 0, err

	}
	fmt.Println("Order Sent")

	return n, nil
}

func processNewTrades(workerId int) {
	for trade := range tradeChannel {
		workerInfo := fmt.Sprintf("Worker %d ==>", workerId)

		// Create key for Order - #TODO - Add Timestamp
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

			if current_pos.Status == "pending" {
				fmt.Printf("%sPending order exists, trade skipped: %s - %s - %t \n", workerInfo, trade, i, ok)
				continue
			}
		}

		// Fetch price quote
		price, err := fetchPriceQuote(trade.ContractId, trade.Exchange)
		if err != nil {
			log.Printf("%sFailed to fetch price for symbol %s: %v", workerInfo, trade.Symbol, err)
			continue
		}

		// Create order
		order := Order{
			Trade:      trade,
			PriceQuote: price,
			Timestamp:  time.Now(),
		}

		// Send order
		orderId, err := transmitOrder(order, true)
		if err != nil {
			log.Printf("%sFailed to subimt order for strategy-symbol %s-%s: %v", workerInfo, trade.StrategyName, trade.Symbol, err)
			continue
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

		// url := fmt.Sprintf("http://127.0.0.1:8000/fills?order_id=%d", orderResp.OrderId)
		url := "http://127.0.0.1:8000/fills"
		// url := fmt.Sprintf("http://broker_api:8000/fills?Id=%d", orderResp.OrderId)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Error sending GET request:", err)
			// return -1.0, err // Default price if there is an error
		}
		defer resp.Body.Close()

		// Parse the response body to extract the price
		var response []Fill
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			fmt.Println("Error decoding fills:", err)
			// return -1.0, err // Default price if there is an error
		}

		for _, fill := range response {
			if fill.Id != orderResp.OrderId {
				continue
			}
			// if fill.Status != "filled" {
			// 	continue
			// }
			direction := 1.0
			if orderResp.Order.Trade.Side == "SELL" {
				direction = -1.0
			}

			updatePositionsToFilled(orderResp, fill.Price, int(direction*math.Abs(float64(fill.Quantity))))
			fmt.Println("Order Filled: ", orderResp.OrderId)
			isFilled = true
		}
		time.Sleep(time.Second)
	}
}

func updatePositionsToPending(orderResp OrderResponse) {
	fmt.Println("Updating Positions for Pending Order")
	positionId := fmt.Sprintf("%s-%s", orderResp.Order.Trade.StrategyName, orderResp.Order.Trade.Symbol)

	p, ok := positions.Load(positionId)
	if !ok {
		fmt.Println("Positon does not exist")

		positions.Store(positionId, definitions.Position{
			Symbol:     orderResp.Order.Trade.Symbol,
			Exchange:   orderResp.Order.Trade.Exchange,
			Quantity:   0.0,
			CostBasis:  0.0,
			Datetime:   time.Now().String(),
			ContractID: int(orderResp.Order.Trade.ContractId),
			Status:     "pending",
		})
	} else {
		p, _ := p.(definitions.Position)
		if !ok {
			fmt.Println("Could not assert Position type on p:", p)
		}
		p.Status = "pending"
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
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return filepath.Join("/shared", filename)
	}

	// Development environment
	return filepath.Join("..", "..", "shared_files", filename)
}

func updatePositionsToFilled(orderResp OrderResponse, costBasis float64, quantity int) {
	fmt.Println("Updating Positions for Filled Order")
	positionId := fmt.Sprintf("%s-%s", orderResp.Order.Trade.StrategyName, orderResp.Order.Trade.Symbol)
	status := "filled"
	positionMap, ok := positions.Load(positionId)
	if ok {
		pos, ok := positionMap.(definitions.Position)
		if ok {

			fmt.Print("Position Map", pos)
			quantity += pos.Quantity

		}

	}
	if quantity == 0.0 {
		status = "closed"
	}

	positions.Store(positionId, definitions.Position{
		Symbol:     orderResp.Order.Trade.Symbol,
		Exchange:   orderResp.Order.Trade.Exchange,
		Quantity:   quantity, // * float64(posAdj),
		CostBasis:  costBasis,
		Datetime:   time.Now().String(),
		ContractID: int(orderResp.Order.Trade.ContractId),
		Status:     status,
	})
	shared_positions := GetSharedFilePath("positions.json")
	// Marshal to JSON file
	if err := SyncMapToJSONFile(&positions, shared_positions); err != nil {
		fmt.Println("Error marshalling sync.Map to JSON:", err)
		return
	}

}

// SyncMapFromJSONFile unmarshals a JSON file into a sync.Map.
func SyncMapFromJSONFile(m *sync.Map, filename string) error {
	// alternative here:https://stackoverflow.com/questions/46390409/how-to-decode-json-strings-to-sync-map-instead-of-normal-map-in-go1-9

	byteValue, err := ioutil.ReadFile(filename)
	if err != nil || len(byteValue) == 0 {
		return err
	}

	normalMap := make(map[string]interface{})
	// Unmarshal JSON data into a slice of Trade structs
	err = json.Unmarshal(byteValue, &normalMap)
	if err != nil {
		return err
	}

	// Store each key-value pair back into the sync.Map
	for k, v := range normalMap {
		vpos, ok := v.(definitions.Position)
		if !ok {
			fmt.Println("Error asserting Position from JSON -->", ok, v)
			continue
		}
		m.Store(k, vpos)
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

var tradeChannel = make(chan *pb.Trade, 100) // Buffered channel for trades
type poolFunction func(int)

var positions sync.Map //

func main() {
	// Clear the original map to demonstrate loading from file
	// sm = sync.Map{}
	shared_positions := GetSharedFilePath("positions.json")
	// Unmarshal from JSON file
	if err := SyncMapFromJSONFile(&positions, shared_positions); err != nil {
		fmt.Println("Error unmarshalling JSON to sync.Map:", err)
		return
	}
	// Start the trade processing worker
	startWorkerPool(5, processNewTrades)
	// Start the trade processing worker
	// startWorkerPool(5, tradeReconciliation)
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
