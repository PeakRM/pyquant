// package main

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"
// 	"os"
// 	"os/exec"
// 	"os/signal"
// 	"pytrader/definitions"
// 	"pytrader/scheduler"
// 	"strings"
// 	"sync"
// 	"syscall"
// 	"time"
// )

// // import (
// // 	"context"
// // 	"fmt"
// // 	"math/rand"
// // 	"net/http"
// // 	"os"
// // 	"os/signal"
// // 	"pytrader/definitions"
// // 	"sync"
// // 	"syscall"
// // 	"time"
// // )

// // // import (
// // // 	"bytes"
// // // 	"encoding/json"
// // // 	"fmt"
// // // 	"io/ioutil"
// // // 	"log"
// // // 	"net/http"
// // // 	"os"
// // // 	"pytrader/definitions"
// // // 	"pytrader/scheduler"

// // // 	"github.com/robfig/cron"
// // // )

// // import (
// // 	"bytes"
// // 	"encoding/json"
// // 	"fmt"
// // 	"io/ioutil"
// // 	"log"
// // 	"net/http"
// // 	"os"
// // 	"os/exec"
// // 	"pytrader/definitions"
// // 	"pytrader/scheduler"
// // 	"strconv"
// // 	"strings"
// // 	"sync"
// // 	"time"
// // )

// type TradeSetup struct {
// 	Market     string   `json:"market"`
// 	Active     bool     `json:"active"`
// 	Timeframe  string   `json:"timeframe"`
// 	Schedule   string   `json:"schedule"`
// 	MarketData []string `json:"market_data"`
// }

// type StrategyConfig struct {
// 	ScriptPath   string                `json:"script_path"`
// 	StrategyType string                `json:"strategy_type"`
// 	Setups       map[string]TradeSetup `json:"setups"`
// }

// // OrderStatus represents the status of an order. Maintianed within program
// type OrderStatus struct { //FIXME - need ot add exchange and contractID to save to positions
// 	StrategyName string
// 	Symbol       string
// 	OrderID      int
// 	Status       string // "pending", "filled", "cancelled"
// 	Timestamp    string
// 	ContractID   int
// 	Exchange     string
// }

// // Response from API
// type OrderStatusResponse struct {
// 	Status    string  `json:"status"`     //: order_status,
// 	Quantity  float64 `json:"quantity"`   //: qty,
// 	CostBasis float64 `json:"cost_basis"` //: avg_fill_price,
// 	Datetime  string  `json:"datetime"`   //: time_submitted
// 	// ContractID int       `json:"conId"`
// }

// type OrderId struct {
// 	StrategyName string
// 	Symbol       string
// 	Id           int
// }

// type Order struct {
// 	definitions.Trade         // Embed the Trade struct
// 	Price             float64 `json:"price"`
// }

// type StrategyConfigFileData map[string]StrategyConfig
// type AllCurrentPositions map[string]map[string]definitions.Position // keys: StrategyName, Symbol
// type PositionJSONFileData map[string]map[string]definitions.Position

// // var (
// // 	tradeQueue       = make(chan definitions.Trade, 100) // Channel for trade queue
// // 	orderUpdateQueue = make(chan OrderStatus, 100)       // Channel for order status updates
// // 	orderStatus      = make(map[int]OrderStatus)         // In-memory order status tracker
// // 	positions        = make(AllCurrentPositions)         // In-memory positions
// // 	mutex            = &sync.Mutex{}                     // Mutex for shared resources
// // )

// // Function to execute a Python script and return its outut
// func executePythonScript(scriptPath string, jsonData string, additionalMarkets string,
// ) {
// 	defer wg.Done()
// 	cmd := exec.Command("python", scriptPath, jsonData, additionalMarkets)

// 	output, err := cmd.CombinedOutput() // Use CombinedOutput to capture stderr as well
// 	if err != nil {
// 		fmt.Printf("Error executing script %s: %s\nOutput: %s", scriptPath, err, string(output))
// 		return
// 	}

// 	// Process the output
// 	outputStr := string(output)
// 	fmt.Printf("%v\n", outputStr)

// 	// var result map[string]interface{}
// 	var trade definitions.Trade
// 	err = json.Unmarshal([]byte(outputStr), &trade)
// 	if err != nil {
// 		fmt.Println("Error parsing JSON output:", err)
// 		return
// 	}

// 	// results <- trade
// 	tradeQueue <- trade
// }

// func runScriptLoop(schedule string) ([]definitions.Trade, error) {

// 	fmt.Println("Running Script loop - ", schedule)
// 	// Load strategy configuration
// 	strategyConfigFile := "C:\\Users\\Jon\\PythonScripts\\pytrader\\backend\\strategies\\strategy-config.json"
// 	strategyConfigData, err := ioutil.ReadFile(strategyConfigFile)
// 	if err != nil {
// 		fmt.Println("Error reading strategy configuration file:", err)
// 		return nil, err
// 	}

// 	// Parse strategy configuration JSON
// 	var strategies map[string]StrategyConfig
// 	err = json.Unmarshal(strategyConfigData, &strategies)
// 	if err != nil {
// 		fmt.Println("Error parsing strategy configuration JSON:", err)
// 		return nil, err
// 	}

// 	// var wg sync.WaitGroup
// 	// results := make(chan definitions.Trade)

// 	// Collect all trade results
// 	var tradeResults []definitions.Trade

// 	// // Use a WaitGroup to wait for the collector goroutine
// 	// var collectorWG sync.WaitGroup
// 	// collectorWG.Add(1)

// 	// // Collector goroutine to gather results - #UNDO HERE FOR 12/3
// 	// go func() {
// 	// 	defer collectorWG.Done()
// 	// 	for trade := range results {
// 	// 		tradeResults = append(tradeResults, trade)
// 	// 	}
// 	// }()

// 	// Loop through each strategy in the configuration
// 	for strategyName, strategy := range strategies {
// 		fmt.Println("Processing strategy:", strategyName)

// 	signalGeneration:
// 		// Loop through each setup for the strategy
// 		for _, setup := range strategy.Setups {

// 			if setup.Schedule != schedule {
// 				// Skip strategies that don't fit schedule setups
// 				fmt.Printf("%s - %s did not pass the Time check: %s\n", strategyName, setup.Market, setup.Schedule)
// 				continue signalGeneration
// 			}
// 			fmt.Printf("%s - %s passed the Time check: %s\n", strategyName, setup.Market, setup.Schedule)

// 			if !setup.Active {
// 				fmt.Printf("%s - %s did not pass the Active check: %t\n", strategyName, setup.Market, setup.Active)
// 				// Skip inactive setups
// 				continue signalGeneration
// 			}
// 			fmt.Printf("%s - %s passed the Active check: %t\n", strategyName, setup.Market, setup.Active)

// 			// Check if strategy setup [StrategyName-Symbol combo] is in the positions data
// 			market := strings.Split(setup.Market, ":")
// 			exchange, symbol := market[0], market[1]

// 			// Protect shared access to positions with mutex
// 			mutex.Lock()
// 			if _, ok := positions[strategyName]; !ok {
// 				positions[strategyName] = make(map[string]definitions.Position)
// 			}
// 			if _, ok := positions[strategyName][symbol]; !ok {
// 				fmt.Println("Adding New Position to position map")
// 				// Add new position
// 				newPosition := definitions.Position{
// 					Symbol:     symbol,
// 					Exchange:   exchange,
// 					Quantity:   0,
// 					CostBasis:  0.0,
// 					Datetime:   "NONE",
// 					ContractID: 0,
// 				}
// 				fmt.Println(newPosition)
// 				positions[strategyName][symbol] = newPosition
// 			}
// 			for _, order := range orderStatus {
// 				if order.Status != "Filled" && order.Symbol == symbol && order.Exchange == exchange && order.StrategyName == strategyName {
// 					fmt.Printf("Pending order exists for %s - %s. Skipping signal generation.", strategyName, market)
// 					continue signalGeneration
// 				}
// 			}

// 			strategyPositions := positions[strategyName][symbol]
// 			mutex.Unlock()

// 			// Fetch buying power from the local server
// 			buyingPower, err := scheduler.FetchBuyingPower()
// 			if err != nil {
// 				fmt.Println("Error fetching buying power:", err)
// 				continue
// 			}
// 			fmt.Println("Current Buying Power: ", buyingPower)

// 			// Prepare the data for the Python script
// 			data := map[string]interface{}{
// 				"position":     strategyPositions,
// 				"buying_power": buyingPower,
// 			}

// 			// Convert the dictionary to a JSON string
// 			jsonData, err := json.Marshal(data)
// 			if err != nil {
// 				fmt.Println("Error converting dictionary to JSON:", err)
// 				continue
// 			}
// 			fmt.Println("Sending Position Data & Buying Power: ", string(jsonData))

// 			// Convert market_data list to a comma-separated string
// 			marketDataStr := strings.Join(setup.MarketData, ",")
// 			fmt.Println("List of Required Additional Market Data: ", marketDataStr)

// 			fmt.Println("Running: ", strategyName)
// 			wg.Add(1)
// 			// Execute the corresponding Python script
// 			go executePythonScript(strategy.ScriptPath, string(jsonData), marketDataStr)
// 		}
// 	}

// 	// Start a goroutine to close the results channel when all scripts are done
// 	// go func() {
// 	// 	wg.Wait()
// 	// 	close(results)
// 	// }()

// 	// // Wait for the collector goroutine to finish
// 	// collectorWG.Wait()

// 	// fmt.Println(tradeResults)
// 	// return tradeResults, nil

// 	// Generate the JSON file name with current date
// 	// fileName := fmt.Sprintf("trades_%s.json", time.Now().Format("01-02-2006-17.06.06"))

// }

// func generateTrades(schedule string) {

// 	newTrades, err := runScriptLoop(schedule)
// 	if err != nil {
// 		fmt.Printf("Error generating %s trades: %s", schedule, err)
// 	}

// 	for _, newTrade := range newTrades {
// 		// Push trade to the queue
// 		tradeQueue <- newTrade
// 		fmt.Printf("Generated trade from %s strategy: %+v\n", schedule, newTrade)
// 	}
// }

// func runStrategy(schedule string, frequency time.Duration) {
// 	for range time.Tick(frequency) {
// 		select {
// 		case <-ctx.Done():
// 			fmt.Println("Stopping trade generation.")
// 			return
// 		default:
// 			// fmt.Println("Running Strategies: ", schedule)
// 			generateTrades(schedule)
// 		}
// 	}
// }

// // // Function to send a GET request to localhost:8081/api/[contract_id] and retrieve the price
// // func getQuote(contractID int, exchange string) float64 {
// // 	url := fmt.Sprintf("http://127.0.0.1:8000/quoteByConId?conId=%d&exchange=%s", contractID, exchange)
// // 	resp, err := http.Get(url)
// // 	if err != nil {
// // 		fmt.Println("Error sending GET request:", err)
// // 		return -1.0 // Default price if there is an error
// // 	}
// // 	defer resp.Body.Close()

// // 	// Parse the response body to extract the price
// // 	var response struct {
// // 		Price float64 `json:"price"`
// // 	}
// // 	err = json.NewDecoder(resp.Body).Decode(&response)
// // 	if err != nil {
// // 		fmt.Println("Error decoding response:", err)
// // 		return -1.0 // Default price if there is an error
// // 	}

// // 	fmt.Printf("GET request to %s returned price: %f\n", url, response.Price)
// // 	if response.Price == 0.0 {
// // 		return -1.0
// // 	}
// // 	return response.Price
// // }

// func transmitOrder(order Order) (int, error) {
// 	url := "http://127.0.0.1:8000/placeLimitOrder?broker=IB"
// 	orderJSON, err := json.Marshal(order)
// 	if err != nil {
// 		fmt.Println("Error marshaling order to JSON:", err)
// 		return 0, err
// 	}
// 	fmt.Println("Order Spec: ", string(orderJSON))

// 	// req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(orderJSON)))
// 	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(orderJSON))
// 	if err != nil {
// 		fmt.Println("Error creating POST request:", err)
// 		return 0, err
// 	}

// 	// Set headers for the POST request
// 	req.Header.Add("Content-Type", "application/json")
// 	fmt.Println("Transmitting Order...")
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		fmt.Println("Error sending POST request:", err)
// 		return 0, err
// 	}
// 	defer resp.Body.Close()

// 	fmt.Printf("POST request to %s completed with status: %s\n", url, resp.Status)
// 	// read response body
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		fmt.Println(err)
// 		return 0, err

// 	}
// 	// print response body
// 	//convert string to int
// 	n, err := strconv.Atoi(string(body))
// 	//check if error occured
// 	if err != nil {
// 		//executes if there is any error
// 		fmt.Println(err)
// 		return 0, err

// 	}
// 	fmt.Println("Order Sent")

// 	return n, nil
// }

// // ExecutionHandler processes trades from the tradeQueue.
// func ExecutionHandler() {

// 	for trade := range tradeQueue {

// 		price := getQuote(trade.ContractID, trade.Exchange)
// 		fmt.Println("Quote: ", price)
// 		if price == -1.0 {
// 			fmt.Println("Error pulling price. Market is not open or contract is not valid")
// 			fmt.Println("Order not sent, trade removed from queue: ", trade)
// 			continue
// 		}

// 		// Create an Order by combining the Trade with the retrieved Price
// 		order := Order{
// 			Trade: trade,
// 			Price: price,
// 		}

// 		// Send POST request with the complete Order
// 		orderID, err := transmitOrder(order)
// 		if err != nil {
// 			fmt.Println(err)
// 		} else {

// 			fmt.Println("Submitted Order:", order)

// 			orderDetails := OrderStatus{
// 				StrategyName: trade.StrategyName,
// 				Symbol:       trade.Symbol,
// 				OrderID:      orderID,
// 				Timestamp:    time.Now().String(),
// 				Status:       "pending",
// 				ContractID:   trade.ContractID,
// 				Exchange:     trade.Exchange,
// 			}

// 			// Now update shared resources with mutex
// 			mutex.Lock()
// 			orderStatus[orderID] = orderDetails
// 			mutex.Unlock()

// 			// Send to orderUpdateQueue without holding the mutex
// 			orderUpdateQueue <- orderDetails
// 			fmt.Println("Order Added to Queue")
// 			go monitorOrderFill(orderDetails)

// 		}
// 	}

// }

// func monitorOrderFill(order OrderStatus) {
// 	timeout := time.After(5 * time.Minute)
// 	ticker := time.NewTicker(5 * time.Second) // Adjust the interval as needed
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-timeout:
// 			fmt.Printf("Order ID %d for %s is still not filled after 1 minute.\n", order.OrderID, order.Symbol)
// 			return
// 		case <-ticker.C:
// 			// Check order status
// 			statusUpdated, err := checkAndUpdateOrderStatus(order)
// 			if err != nil {
// 				fmt.Printf("Error checking order status for Order ID %d: %v\n", order.OrderID, err)
// 				continue
// 			}
// 			if statusUpdated {
// 				// Order has been filled or cancelled
// 				return
// 			}
// 		}
// 	}
// }

// func checkAndUpdateOrderStatus(order OrderStatus) (bool, error) {
// 	// Update order status
// 	url := fmt.Sprintf("http://127.0.0.1:8000/orderStatus?orderId=%d&broker=IB", order.OrderID)
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		return false, fmt.Errorf("error sending GET request: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	var orderStatusResponse OrderStatusResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&orderStatusResponse); err != nil {
// 		return false, fmt.Errorf("error decoding response: %v", err)
// 	}

// 	// Update order status in memory
// 	mutex.Lock()
// 	defer mutex.Unlock()

// 	existingOrder, exists := orderStatus[order.OrderID]
// 	if !exists {
// 		return false, fmt.Errorf("order ID %d not found in orderStatus map", order.OrderID)
// 	}

// 	existingOrder.Status = orderStatusResponse.Status
// 	existingOrder.Timestamp = time.Now().String()
// 	orderStatus[order.OrderID] = existingOrder

// 	// Check if the order is filled
// 	if orderStatusResponse.Status == "Filled" {
// 		// Update positions
// 		updatePosition(existingOrder, orderStatusResponse)
// 		// Remove the order from orderStatus map
// 		delete(orderStatus, order.OrderID)
// 		fmt.Printf("Order ID %d for %s has been filled.\n", order.OrderID, order.Symbol)
// 		return true, nil
// 	}

// 	// You can also handle other statuses like "Cancelled" if needed
// 	return false, nil
// }

// func updatePosition(order OrderStatus, orderStatusResponse OrderStatusResponse) {
// 	// Access the strategy's positions
// 	mutex.Lock()
// 	defer mutex.Unlock()

// 	positionsBySymbol, strategyExists := positions[order.StrategyName]
// 	if !strategyExists {
// 		positionsBySymbol = make(map[string]definitions.Position)
// 		positions[order.StrategyName] = positionsBySymbol
// 	}

// 	// Access the position for the symbol
// 	position, symbolExists := positionsBySymbol[order.Symbol]
// 	if !symbolExists {
// 		// Initialize a new position if it doesn't exist
// 		position = definitions.Position{
// 			Symbol:     order.Symbol,
// 			Exchange:   order.Exchange,
// 			ContractID: order.ContractID,
// 		}
// 	}

// 	// Update the position fields
// 	position.Quantity = orderStatusResponse.Quantity
// 	position.CostBasis = orderStatusResponse.CostBasis
// 	position.Datetime = orderStatusResponse.Datetime //.Format(time.RFC3339)

// 	// Assign the updated position back to the map
// 	positionsBySymbol[order.Symbol] = position

// 	fmt.Printf("Position updated for %s: %+v\n", order.Symbol, position)

// 	// Optionally, persist the positions to a file
// 	positionsFilename := "C:/Users/Jon/PythonScripts/pytrader/backend/positions.json"
// 	if err := backupToFile(positionsFilename, positions); err != nil {
// 		fmt.Println("Error backing up positions:", err)
// 	} else {
// 		fmt.Println("Positions backed up successfully")
// 	}
// }

// // // func ReconcileTrades() {
// // // 	ticker := time.NewTicker(30 * time.Second)
// // // 	defer ticker.Stop()

// // // 	positionsFilename := "C:/Users/Jon/PythonScripts/pytrader/backend/positions.json"

// // // 	for range ticker.C {
// // // 		// Process pending orders without holding the mutex
// // // 		var hasOrders bool = true
// // // 		for hasOrders {
// // // 			select {
// // // 			case pendingOrder := <-orderUpdateQueue:
// // // 				fmt.Println("Checking for fills: ", pendingOrder)

// // // 				// Lock mutex only when accessing shared resources
// // // 				mutex.Lock()
// // // 				// Update order status and positions
// // // 				updateOrderAndPosition(pendingOrder, positionsFilename)
// // // 				mutex.Unlock()
// // // 			default:
// // // 				// No more orders to process
// // // 				fmt.Println("No orders to reconcile at this time")
// // // 				hasOrders = false
// // // 			}
// // // 		}
// // // 	}
// // // }

// // // Helper function to update order status and positions
// // // func updateOrderAndPosition(pendingOrder OrderStatus, positionsFilename string) {
// // // 	// Update order status
// // // 	url := fmt.Sprintf("http://127.0.0.1:8000/fillsByOrderId?orderId=%d&broker=IB", pendingOrder.OrderID)
// // // 	resp, err := http.Get(url)
// // // 	if err != nil {
// // // 		fmt.Println("Error sending GET request:", err)
// // // 		// return 0.0 // Default price if there is an error
// // // 	}
// // // 	defer resp.Body.Close()
// // // 	body, err := ioutil.ReadAll(resp.Body)
// // // 	if err != nil {
// // // 		fmt.Println(err)
// // // 	}

// // // 	var orderStatusResponse OrderStatusResponse
// // // 	json.Unmarshal(body, &orderStatusResponse)

// // // 	// updatXe order quantity in file
// // // 	// update in memeory positions
// // // 	if order, exists := orderStatus[pendingOrder.OrderID]; exists {
// // // 		//FIXME - Can probably move these three lines ot else statement of if order == filled.
// // // 		order.Status = pendingOrder.Status
// // // 		order.Timestamp = pendingOrder.Timestamp
// // // 		orderStatus[pendingOrder.OrderID] = pendingOrder

// // // 		// Update positions if the order is filled
// // // 		if pendingOrder.Status == "Filled" {
// // // 			delete(orderStatus, pendingOrder.OrderID)
// // // 			// Update positions when the order is filled
// // // 			pos, exists := positions[order.StrategyName][order.Symbol]
// // // 			if exists {
// // // 				pos.Quantity = orderStatusResponse.Quantity
// // // 				pos.CostBasis = orderStatusResponse.CostBasis
// // // 				pos.Datetime = orderStatusResponse.Datetime.String()
// // // 			} else {
// // // 				// Create a new position if not exists
// // // 				positions[order.StrategyName][order.Symbol] = definitions.Position{
// // // 					Symbol:     order.Symbol,
// // // 					Exchange:   order.Exchange, //FIXME - Feed this in from OrderStatus Or OrderStatusResponse (needs to eb added there first)
// // // 					Quantity:   orderStatusResponse.Quantity,
// // // 					CostBasis:  orderStatusResponse.CostBasis,
// // // 					Datetime:   orderStatusResponse.Datetime.String(),
// // // 					ContractID: order.ContractID,
// // // 				}
// // // 			}
// // // 		}
// // // 	}

// // // 	// Backup positions
// // // 	if err := backupToFile(positionsFilename, positions); err != nil {
// // // 		fmt.Println("Error backing up positions:", err)
// // // 	} else {
// // // 		fmt.Println("Positions backed up successfully")
// // // 	}
// // // }

// // // ReconciliationProcess updates order statuses and positions.
// // // func ReconcileTrades() {
// // // 	ticker := time.NewTicker(30 * time.Second)
// // // 	defer ticker.Stop()

// // // 	positionsFilename := "C:/Users/Jon/PythonScripts/pytrader/backend/positions.json"

// // // 	for range ticker.C {

// // // 		mutex.Lock()
// // // 		for pendingOrder := range orderUpdateQueue {
// // // 			fmt.Println("Checking for fills: ", pendingOrder)

// // // 			// Update order status
// // // 			url := fmt.Sprintf("http://127.0.0.1:8000/orderStatus?orderId=%d&broker=IB", pendingOrder.OrderID)
// // // 			resp, err := http.Get(url)
// // // 			if err != nil {
// // // 				fmt.Println("Error sending GET request:", err)
// // // 				// return 0.0 // Default price if there is an error
// // // 			}
// // // 			defer resp.Body.Close()
// // // 			body, err := ioutil.ReadAll(resp.Body)
// // // 			if err != nil {
// // // 				fmt.Println(err)
// // // 			}

// // // 			var orderStatusResponse OrderStatusResponse
// // // 			json.Unmarshal(body, &orderStatusResponse)

// // // 			// updatXe order quantity in file
// // // 			// update in memeory positions
// // // 			if order, exists := orderStatus[pendingOrder.OrderID]; exists {
// // // 				//FIXME - Can probably move these three lines ot else statement of if order == filled.
// // // 				order.Status = pendingOrder.Status
// // // 				order.Timestamp = pendingOrder.Timestamp
// // // 				orderStatus[pendingOrder.OrderID] = pendingOrder

// // // 				// Update positions if the order is filled
// // // 				if pendingOrder.Status == "Filled" {
// // // 					delete(orderStatus, pendingOrder.OrderID)
// // // 					// Update positions when the order is filled
// // // 					pos, exists := positions[order.StrategyName][order.Symbol]
// // // 					if exists {
// // // 						pos.Quantity = orderStatusResponse.Quantity
// // // 						pos.CostBasis = orderStatusResponse.CostBasis
// // // 						pos.Datetime = orderStatusResponse.Datetime.String()
// // // 					} else {
// // // 						// Create a new position if not exists
// // // 						positions[order.StrategyName][order.Symbol] = definitions.Position{
// // // 							Symbol:     order.Symbol,
// // // 							Exchange:   order.Exchange, //FIXME - Feed this in from OrderStatus Or OrderStatusResponse (needs to eb added there first)
// // // 							Quantity:   orderStatusResponse.Quantity,
// // // 							CostBasis:  orderStatusResponse.CostBasis,
// // // 							Datetime:   orderStatusResponse.Datetime.String(),
// // // 							ContractID: order.ContractID,
// // // 						}
// // // 					}

// // // 					// Backup positions
// // // 					if err := backupToFile(positionsFilename, positions); err != nil {
// // // 						fmt.Println("Error backing up positions:", err)
// // // 					} else {
// // // 						fmt.Println("Positions backed up successfully")
// // // 					}

// // // 				}
// // // 			}

// // // 		}
// // // 		mutex.Unlock()
// // // 	}
// // // }

// // func UpdatePosition(positionFile PositionJSONFileData, order OrderStatus, orderStatusResponse OrderStatusResponse) error {
// // 	// Access the strategy's positions
// // 	positionsBySymbol, strategyExists := positionFile[order.StrategyName]
// // 	if !strategyExists {
// // 		return fmt.Errorf("StrategyName '%s' not found in PositionFile", order.StrategyName)
// // 	}

// // 	// Access the position for the symbol
// // 	position, err := positionsBySymbol[order.Symbol]
// // 	if !err {
// // 		return fmt.Errorf("Symbol '%s' not found for StrategyName '%s' in PositionFile",
// // 			order.Symbol, order.StrategyName)
// // 	}

// // 	// Update the position fields
// // 	// position.Exchange = float64(orderStatus.Exchange)  Don't need this because that
// // 	position.Quantity = float64(orderStatusResponse.Quantity)
// // 	position.CostBasis = float64(orderStatusResponse.CostBasis)
// // 	position.Datetime = orderStatusResponse.Datetime.Format(time.RFC3339)

// // 	// Assign the updated position back to the map
// // 	positionsBySymbol[order.Symbol] = position

// // 	return nil
// // }

// // // BackupProcessor periodically writes in-memory data to JSON files.
// // func BackupProcessor(interval time.Duration, positionsFile, orderStatusFile, tradeQueueFile string) {
// // 	ticker := time.NewTicker(interval)
// // 	defer ticker.Stop()

// // 	for range ticker.C {
// // 		// Lock the shared resources
// // 		mutex.Lock()
// // 		fmt.Println("Starting backup...")

// // 		// Backup positions
// // 		if err := backupToFile(positionsFile, positions); err != nil {
// // 			fmt.Println("Error backing up positions:", err)
// // 		} else {
// // 			fmt.Println("Positions backed up successfully")
// // 		}

// // 		// Backup order status
// // 		if err := backupToFile(orderStatusFile, orderStatus); err != nil {
// // 			fmt.Println("Error backing up order status:", err)
// // 		} else {
// // 			fmt.Println("Order status backed up successfully")
// // 		}
// // 		// Backup trade queue (convert channel to slice for serialization)
// // 		trades := drainTradeQueue()
// // 		if err := backupToFile(tradeQueueFile, trades); err != nil {
// // 			fmt.Println("Error backing up trade queue:", err)
// // 		}
// // 		fmt.Println("Backup Completed")
// // 		mutex.Unlock()
// // 	}
// // }

// // // drainTradeQueue empties the tradeQueue and returns its contents as a slice.
// // func drainTradeQueue() []definitions.Trade {
// // 	var trades []definitions.Trade
// // 	for {
// // 		select {
// // 		case trade := <-tradeQueue:
// // 			trades = append(trades, trade)
// // 		default:
// // 			return trades // Exit when the channel is empty
// // 		}
// // 	}
// // }

// // // backupToFile writes a given data structure to a JSON file.
// // func backupToFile(filename string, data interface{}) error {
// // 	// Create or overwrite the file
// // 	file, err := os.Create(filename)
// // 	if err != nil {
// // 		return fmt.Errorf("failed to create file %s: %w", filename, err)
// // 	}
// // 	defer file.Close()

// // 	// Serialize data to JSON
// // 	encoder := json.NewEncoder(file)
// // 	encoder.SetIndent("", "  ") // Pretty-print JSON for readability
// // 	if err := encoder.Encode(data); err != nil {
// // 		return fmt.Errorf("failed to encode data to JSON: %w", err)
// // 	}

// // 	return nil
// // }

// // // RecoverSystemState loads system state from backup files if they exist and are valid.
// // func RecoverSystemState(positionsFile string, orderStatusFile string) error {
// // 	// Recover positions
// // 	if err := loadFromFile(positionsFile, &positions); err != nil {
// // 		return fmt.Errorf("failed to recover positions: %w", err)
// // 	}
// // 	fmt.Println("Positions recovered successfully:", positions)

// // 	// Recover order status
// // 	if err := loadFromFile(orderStatusFile, &orderStatus); err != nil {
// // 		return fmt.Errorf("failed to recover order status: %w", err)
// // 	}
// // 	fmt.Println("Order status recovered successfully:", orderStatus)

// // 	return nil
// // }

// // // loadFromFile loads JSON data from a file into the provided target.
// // func loadFromFile(filename string, target interface{}) error {
// // 	// Check if file exists
// // 	if _, err := os.Stat(filename); os.IsNotExist(err) {
// // 		return fmt.Errorf("file %s does not exist", filename)
// // 	}

// // 	// Open the file
// // 	file, err := os.Open(filename)
// // 	if err != nil {
// // 		return fmt.Errorf("failed to open file %s: %w", filename, err)
// // 	}
// // 	defer file.Close()

// // 	// Decode JSON data into the target
// // 	decoder := json.NewDecoder(file)
// // 	if err := decoder.Decode(target); err != nil {
// // 		return fmt.Errorf("failed to decode JSON from file %s: %w", filename, err)
// // 	}

// // 	return nil
// // }

// // func RecoverTradeQueue(tradeQueueFile string) error {
// // 	var trades []definitions.Trade
// // 	if err := loadFromFile(tradeQueueFile, &trades); err != nil {
// // 		return fmt.Errorf("failed to recover trade queue: %w", err)
// // 	}

// // 	// Push recovered trades back into the tradeQueue channel
// // 	for _, trade := range trades {
// // 		tradeQueue <- trade
// // 	}

// // 	fmt.Println("Trade queue recovered successfully:", trades)
// // 	return nil
// // }

// // func main() {

// // 	// Recover system state from backup files
// // 	if err := RecoverSystemState("positions_backup.json", "order_status_backup.json"); err != nil {
// // 		fmt.Println("System state recovery failed:", err)
// // 	} else {
// // 		fmt.Println("System state recovered successfully")
// // 	}

// // 	// Recover trade queue
// // 	if err := RecoverTradeQueue("trade_queue_backup.json"); err != nil {
// // 		fmt.Println("Trade queue recovery failed:", err)
// // 	} else {
// // 		fmt.Println("Trade queue recovered successfully")
// // 	}
// // 	// Start strategy schedulers
// // 	go runStrategy("Intraday", 1*time.Minute)

// // 	go ExecutionHandler() // Start execution handler
// // 	// go ReconcileTrades()  // Reconciles order statuses
// // 	go BackupProcessor(5*time.Minute, "positions_backup.json",
// // 		"order_status_backup.json",
// // 		"trade_queue_backup.json") // Periodically backs up data

// // 	// Prevent the main function from exiting
// // 	select {}
// // }

// // // func main() {
// // // go test.TestExecution()

// // // c := cron.New()
// // // fmt.Println("Scheduling....")
// // //	c.AddFunc("0 * * * * *", scheduler.RunScriptsIntraday) // Runs every minute
// // // c.AddFunc("0 55 17 * * *", runScriptsFuturesDailyOpen)       // Runs every day
// // // c.AddFunc("0 20 20 * * *", runScriptsFuturesDailyOpen)       // Runs every day?
// // // c.AddFunc("0 55 17 * * *", runScriptsEquitiesDailyOpenClose) // Runs every day
// // // c.AddFunc("0 25 9 * * *", runScriptsEquitiesDailyOpenClose)  // Runs every day
// // // c.AddFunc("0 55 15 * * *", runScriptsEquitiesDailyOpenClose) // Runs every day

// // //c.Start()

// // // select {}
// // // http.HandleFunc("/api/dashboard-config", handleConfig)
// // // http.HandleFunc("/api/open-positions", handleOpenPositions)
// // // http.HandleFunc("/api/markets/pause", pauseMarket)
// // // http.HandleFunc("/api/markets/resume", resumeMarket)

// // // log.Println("Starting server on :8081")
// // // if err := http.ListenAndServe(":8081", nil); err != nil {
// // // log.Fatalf("Could not start server: %s\n", err.Error())
// // //}

// // // }

// // // Function to unpack positions data from a file
// // func unpackPositionsData(filePath string) (PositionJSONFileData, error) {
// // 	var positions PositionJSONFileData
// // 	// Read the file content
// // 	byteValue, err := ioutil.ReadFile(filePath)
// // 	if err != nil || len(byteValue) == 0 {
// // 		return nil, err
// // 	}

// // 	// Unmarshal JSON data into a slice of Trade structs
// // 	err = json.Unmarshal(byteValue, &positions)
// // 	if err != nil {
// // 		return nil, err
// // 	}

// // 	return positions, nil
// // }

// // func handleConfig(w http.ResponseWriter, r *http.Request) {
// // 	// Add CORS headers
// // 	w.Header().Set("Access-Control-Allow-Origin", "*")
// // 	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
// // 	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

// // 	// Handle preflight (OPTIONS) requests
// // 	if r.Method == http.MethodOptions {
// // 		w.WriteHeader(http.StatusOK)
// // 		return
// // 	}

// // 	switch r.Method {
// // 	case "GET":
// // 		getConfig(w, r)
// // 	case "POST":
// // 		updateConfig(w, r)
// // 	default:
// // 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// // 	}
// // }

// // func getConfig(w http.ResponseWriter, r *http.Request) {
// // 	configPath := "C:/Users/Jon/PythonScripts/pytrader/backend/strategies/strategy-config.json"

// // 	file, err := os.Open(configPath)
// // 	if err != nil {
// // 		http.Error(w, "Could not read config file", http.StatusInternalServerError)
// // 		return
// // 	}
// // 	defer file.Close()

// // 	data, err := ioutil.ReadAll(file)
// // 	if err != nil {
// // 		http.Error(w, "Could not read file data", http.StatusInternalServerError)
// // 		return
// // 	}

// // 	w.Header().Set("Content-Type", "application/json")
// // 	w.Write(data)
// // }

// // func updateConfig(w http.ResponseWriter, r *http.Request) {
// // 	var strategies map[string]StrategyConfig

// // 	// Debug: Print the incoming request body
// // 	body, _ := ioutil.ReadAll(r.Body)
// // 	log.Println("Received Payload:", string(body))

// // 	// Reset the body reader since we've consumed it
// // 	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

// // 	if err := json.NewDecoder(r.Body).Decode(&strategies); err != nil {
// // 		http.Error(w, "Invalid request payload", http.StatusBadRequest)
// // 		log.Println("Error decoding JSON:", err)
// // 		return
// // 	}

// // 	// Debug: Print the decoded strategies
// // 	log.Println("Decoded Strategies:", strategies)

// // 	configPath := "C:/Users/Jon/PythonScripts/pytrader/backend/strategies/strategy-config.json"
// // 	data, err := json.MarshalIndent(strategies, "", "  ")
// // 	if err != nil {
// // 		http.Error(w, "Could not marshal JSON", http.StatusInternalServerError)
// // 		log.Println("Error marshaling JSON:", err)
// // 		return
// // 	}
// // 	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
// // 		http.Error(w, "Could not write to config file", http.StatusInternalServerError)
// // 		log.Println("Error writing to file:", err)
// // 		return
// // 	}

// // 	w.WriteHeader(http.StatusOK)
// // 	w.Write([]byte("Configuration updated successfully"))
// // }

// // func handleOpenPositions(w http.ResponseWriter, r *http.Request) {
// // 	// Add CORS headers
// // 	w.Header().Set("Access-Control-Allow-Origin", "*")
// // 	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
// // 	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

// // 	// Handle preflight (OPTIONS) requests
// // 	if r.Method == http.MethodOptions {
// // 		w.WriteHeader(http.StatusOK)
// // 		return
// // 	}

// // 	switch r.Method {
// // 	case "GET":
// // 		getOpenPositions(w, r)
// // 	// case "POST":
// // 	// updateOpenPositionsHandler(w, r)
// // 	default:
// // 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// // 	}
// // }

// // func getOpenPositions(w http.ResponseWriter, r *http.Request) {
// // 	configPath := "C:\\Users\\Jon\\PythonScripts\\pytrader\\backend\\positions.json"

// // 	file, err := os.Open(configPath)
// // 	if err != nil {
// // 		http.Error(w, "Could not read open positions file", http.StatusInternalServerError)
// // 		return
// // 	}
// // 	defer file.Close()

// // 	data, err := ioutil.ReadAll(file)
// // 	if err != nil {
// // 		http.Error(w, "Could not read file data", http.StatusInternalServerError)
// // 		return
// // 	}

// // 	w.Header().Set("Content-Type", "application/json")
// // 	w.Write(data)
// // }

// // func loadStrategies(filename string) (StrategyConfigFileData, error) {
// // 	var strategyFile StrategyConfigFileData

// // 	file, err := ioutil.ReadFile(filename)
// // 	if err != nil {
// // 		return nil, err
// // 	}

// // 	err = json.Unmarshal(file, &strategyFile)
// // 	if err != nil {
// // 		return nil, err
// // 	}

// // 	return strategyFile, nil
// // }

// // // func updateOpenPositionsHandler(w http.ResponseWriter, r *http.Request) {
// // // 	var newPositions OpenPositions
// // // 	// var newPositions Position

// // // 	if err := json.NewDecoder(r.Body).Decode(&newPositions); err != nil {
// // // 		http.Error(w, "Invalid request payload", http.StatusBadRequest)
// // // 		return
// // // 	}

// // // 	updateOpenPositions(newPositions)

// // // 	w.WriteHeader(http.StatusOK)
// // // 	w.Write([]byte("Open positions updated successfully"))
// // // }

// // // func updateOpenPositions(newPositions OpenPositions) {
// // // 	configPath := "positions.json"

// // // 	// Read existing data
// // // 	// existingPositions := make(Positions)
// // // 	// if file, err := os.Open(configPath); err == nil {
// // // 	// 	defer file.Close()
// // // 	// 	data, err := ioutil.ReadAll(file)
// // // 	// 	if err == nil {
// // // 	// 		_ = json.Unmarshal(data, &existingPositions)
// // // 	// 	}
// // // 	// }

// // // 	// Load positions data
// // // 	positionsFile := "positions.json"
// // // 	existingPositions, err := ioutil.ReadFile(positionsFile)
// // // 	if err != nil {
// // // 		fmt.Println("Error reading positions file:", err)
// // // 		return
// // // 	}

// // // 	// Parse positions JSON
// // // 	var positions map[string]map[string]Position
// // // 	err = json.Unmarshal(existingPositions, &positions)
// // // 	if err != nil {
// // // 		fmt.Println("Error parsing positions JSON:", err)
// // // 		return
// // // 	}

// // // 	// Update or add new positions
// // // 	for strategy_name, strategy_position := range newPositions {
// // // 		if _, exists := existingPositions[strategy_name]; !exists {
// // // 			existingPositions[strategy_name] = strategy_position
// // // 		} else {
// // // 			for symbol, position_detail := range strategy_position {
// // // 				existingPosition, exists := existingPositions[strategy_name][symbol]
// // // 				if !exists || existingPosition != position_detail {
// // // 					existingPositions[strategy_name][symbol] = position_detail
// // // 				}
// // // 			}
// // // 		}
// // // 	}

// // // 	// Write updated data back to file
// // // 	data, err := json.MarshalIndent(existingPositions, "", "  ")
// // // 	if err == nil {
// // // 		_ = ioutil.WriteFile(configPath, data, 0644)
// // // 	}
// // // }

// // // // Pause market handler
// // // func pauseMarket(w http.ResponseWriter, r *http.Request) {

// // // 	// Add CORS headers
// // // 	w.Header().Set("Access-Control-Allow-Origin", "*")
// // // 	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
// // // 	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

// // // 	// Handle preflight (OPTIONS) requests
// // // 	if r.Method == http.MethodOptions {
// // // 		w.WriteHeader(http.StatusOK)
// // // 		return
// // // 	}
// // // 	var data struct {
// // // 		StrategyName string `json:"strategyName"`
// // // 		Market       string `json:"market"`
// // // 	}

// // // 	err := json.NewDecoder(r.Body).Decode(&data)
// // // 	if err != nil {
// // // 		http.Error(w, err.Error(), http.StatusBadRequest)
// // // 		return
// // // 	}

// // // 	// Load the strategies
// // // 	strategies, err := loadStrategies("strategy-config.json")
// // // 	if err != nil {
// // // 		http.Error(w, err.Error(), http.StatusInternalServerError)
// // // 		return
// // // 	}

// // // 	strategy, exists := strategies.Setup[data.Market]
// // // 	if !exists {
// // // 		http.Error(w, "Strategy not found", http.StatusNotFound)
// // // 		return
// // // 	}

// // // 	// Move market from active to inactive
// // // 	var newActiveMarkets []string
// // // 	for _, market := range strategy {
// // // 		if market != data.Market {
// // // 			newActiveMarkets = append(newActiveMarkets, market)
// // // 		}
// // // 	}

// // // 	strategy.ActiveMarkets = newActiveMarkets
// // // 	strategy.InactiveMarkets = append(strategy.InactiveMarkets, data.Market)
// // // 	strategies[data.StrategyName] = strategy

// // // 	// Save the updated strategies back to the file
// // // 	err = saveStrategies("dashboard_config.json", strategies)
// // // 	if err != nil {
// // // 		http.Error(w, err.Error(), http.StatusInternalServerError)
// // // 		return
// // // 	}

// // // 	w.WriteHeader(http.StatusOK)
// // // }

// // // func resumeMarket(w http.ResponseWriter, r *http.Request) {

// // // 	// Add CORS headers
// // // 	w.Header().Set("Access-Control-Allow-Origin", "*")
// // // 	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
// // // 	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

// // // 	// Handle preflight (OPTIONS) requests
// // // 	if r.Method == http.MethodOptions {
// // // 		w.WriteHeader(http.StatusOK)
// // // 		return
// // // 	}
// // // 	var data struct {
// // // 		StrategyName string `json:"strategyName"`
// // // 		Market       string `json:"market"`
// // // 	}

// // // 	err := json.NewDecoder(r.Body).Decode(&data)
// // // 	if err != nil {
// // // 		http.Error(w, err.Error(), http.StatusBadRequest)
// // // 		return
// // // 	}

// // // 	// Load the strategies
// // // 	strategies, err := loadStrategies("dashboard_config.json")
// // // 	if err != nil {
// // // 		http.Error(w, err.Error(), http.StatusInternalServerError)
// // // 		return
// // // 	}

// // // 	strategy, exists := strategies[data.StrategyName]
// // // 	if !exists {
// // // 		http.Error(w, "Strategy not found", http.StatusNotFound)
// // // 		return
// // // 	}

// // // 	// Move market from inactive to active
// // // 	var newInactiveMarkets []string
// // // 	for _, market := range strategy.InactiveMarkets {
// // // 		if market != data.Market {
// // // 			newInactiveMarkets = append(newInactiveMarkets, market)
// // // 		}
// // // 	}

// // // 	strategy.InactiveMarkets = newInactiveMarkets
// // // 	strategy.ActiveMarkets = append(strategy.ActiveMarkets, data.Market)
// // // 	strategies[data.StrategyName] = strategy

// // // 	// Save the updated strategies back to the file
// // // 	err = saveStrategies("dashboard_config.json", strategies)
// // // 	if err != nil {
// // // 		http.Error(w, err.Error(), http.StatusInternalServerError)
// // // 		return
// // // 	}

// // // 	w.WriteHeader(http.StatusOK)
// // // }

// // // func saveStrategies(filename string, strategies Strategies) error {
// // // 	data, err := json.MarshalIndent(strategies, "", "  ")
// // // 	if err != nil {
// // // 		return err
// // // 	}

// // // 	err = ioutil.WriteFile(filename, data, 0644)
// // // 	if err != nil {
// // // 		return err
// // // 	}

// // // 	return nil
// // // }

// var (
// 	tradeQueue       = make(chan definitions.Trade, 100) // Channel for trade queue
// 	orderUpdateQueue = make(chan OrderStatus, 100)       // Channel for order status updates
// 	orderStatus      = make(map[int]OrderStatus)         // In-memory order status tracker
// 	positions        = make(AllCurrentPositions)         // In-memory positions
// 	mutex            = &sync.Mutex{}                     // Mutex for shared resources
// 	wg               sync.WaitGroup
// 	ctx              context.Context
// )

// func main() {
// 	// Channels for pipeline
// 	// tradeQueue := make(chan Data, 10)
// 	// enrichedChan := make(chan Data, 10)
// 	// statusChan := make(chan Data, 10)

// 	ctx, cancel := context.WithCancel(context.Background())

// 	// Recover system state from backup files
// 	if err := RecoverSystemState("positions_backup.json", "order_status_backup.json"); err != nil {
// 		fmt.Println("System state recovery failed:", err)
// 	} else {
// 		fmt.Println("System state recovered successfully")
// 	}

// 	// Recover trade queue
// 	if err := RecoverTradeQueue("trade_queue_backup.json"); err != nil {
// 		fmt.Println("Trade queue recovery failed:", err)
// 	} else {
// 		fmt.Println("Trade queue recovered successfully")
// 	}

// 	// Start goroutines
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		// Start strategy schedulers
// 		runStrategy("Intraday", 1*time.Minute)
// 		// Run Iterate through stragegies here
// 	}()

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		ExecutionHandler() // Start execution handler
// 	}()

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		ReconcileTrades(ctx)
// 	}()

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		go BackupProcessor(5*time.Minute, "positions_backup.json",
// 			"order_status_backup.json",
// 			"trade_queue_backup.json")
// 	}()

// 	// Start HTTP server for shutdown
// 	go startShutdownServer()

// 	// Handle system interrupts for shutdown
// 	go handleSystemSignals()

// 	// Wait for all goroutines to complete
// 	wg.Wait()
// 	fmt.Println("System shutdown complete.")
// }

// // Data generation goroutine
// // func generateTrades(ctx context.Context, tradeQueue chan<- definitions.Trade) {
// // 	id := 0
// // 	for {
// // 		select {
// // 		case <-ctx.Done():
// // 			fmt.Println("Stopping trade generation.")
// // 			return
// // 		default:
// // 			schedu
// // 		}
// // 	}
// // }

// // Data enrichment goroutine
// // func enrichData(ctx context.Context, tradeQueue <-chan definitions.Trade, enrichedChan chan<- Data) {
// // 	for {
// // 		select {
// // 		case <-ctx.Done():
// // 			fmt.Println("Stopping data enrichment.")
// // 			return
// // 		case data := <-tradeQueue:
// // 			data.Info = fmt.Sprintf("%s (enriched)", data.Info)
// // 			fmt.Println("Enriched:", data)
// // 			enrichedChan <- data
// // 		}
// // 	}
// // }

// // API request goroutine
// func sendToAPI(ctx context.Context, enrichedChan <-chan Data, statusChan chan<- Data) {
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			fmt.Println("Stopping API requests.")
// 			return
// 		case data := <-enrichedChan:
// 			fmt.Println("Sending to API:", data)
// 			statusChan <- data
// 		}
// 	}
// }

// func monitorAPIStatus(ctx context.Context, statusChan <-chan Data) {
// 	// Map to track pending items
// 	pendingItems := make(map[int]Data)
// 	var mu sync.Mutex

// 	// Goroutine to continuously check statuses
// 	go func() {
// 		ticker := time.NewTicker(2 * time.Second) // Check every 2 seconds
// 		defer ticker.Stop()

// 		for {
// 			select {
// 			case <-ctx.Done():
// 				fmt.Println("Stopping API status monitoring.")
// 				return
// 			case <-ticker.C:
// 				// Check statuses periodically
// 				mu.Lock()
// 				for id, data := range pendingItems {
// 					status := checkAPIStatus(data) // Simulate API status check
// 					fmt.Printf("Checked status for ID=%d: %s\n", id, status)

// 					if status == "Complete" {
// 						fmt.Printf("ID=%d is complete, removing from pending list.\n", id)
// 						delete(pendingItems, id) // Remove completed items
// 					}
// 				}
// 				mu.Unlock()
// 			}
// 		}
// 	}()

// 	// Goroutine to receive new items for monitoring
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			fmt.Println("Stopping new item reception for monitoring.")
// 			return
// 		case data := <-statusChan:
// 			mu.Lock()
// 			fmt.Printf("Adding ID=%d to monitoring list.\n", data.ID)
// 			pendingItems[data.ID] = data
// 			mu.Unlock()
// 		}
// 	}
// }

// // // Simulate an API status check
// // func checkAPIStatus(data Data) string {
// // 	// Mock: Return "Complete" randomly for demonstration
// // 	if rand.Intn(5) == 0 { // 20% chance of completion
// // 		return "Complete"
// // 	}
// // 	return "InProgress"
// // }

// // HTTP server to handle shutdown request
// func startShutdownServer(cancel context.CancelFunc) {
// 	http.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Println("Shutdown request received.")
// 		cancel() // Trigger the shutdown
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte("Shutting down the system..."))
// 	})

// 	if err := http.ListenAndServe(":8080", nil); err != nil {
// 		fmt.Println("HTTP server error:", err)
// 	}
// }

// // Handle system interrupts (Ctrl+C, SIGTERM)
// func handleSystemSignals(cancel context.CancelFunc) {
// 	sigChan := make(chan os.Signal, 1)
// 	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
// 	<-sigChan // Wait for a signal
// 	fmt.Println("System signal received, shutting down...")
// 	cancel() // Trigger the shutdown
// }

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
	"pytrader/definitions"
	pb "pytrader/tradepb"
	"strconv"
	"strings"
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
	Id       int     `json:"id"`
	Price    float64 `json:"price"`
	Status   string  `json:"status"`
	Quantity float64 `json:"quantity"`
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

// Function to send a GET request to localhost:8081/api/[contract_id] and retrieve the price
func fetchPriceQuote(contractID int32, exchange string) (float64, error) {
	// url := fmt.Sprintf("http://127.0.0.1:8000/quoteByConId?conId=%d&exchange=%s", contractID, exchange)
	url := fmt.Sprintf("http://broker_api:8000/quoteByConId?conId=%d&exchange=%s", contractID, exchange)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error sending GET request:", err)
		return -1.0, err // Default price if there is an error
	}
	defer resp.Body.Close()

	// Parse the response body to extract the price
	var response struct {
		Price float64 `json:"price"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Println("Error decoding quote:", err)
		return -1.0, err // Default price if there is an error
	}

	fmt.Printf("GET request to %s returned price: %f\n", url, response.Price)
	if response.Price == 0.0 {
		return -1.0, &MyError{}
	}
	return response.Price, nil
}

// Send order to BrokerAPI
func transmitOrder(order Order, testTrade bool) (int, error) {
	if testTrade {
		fmt.Println("Test Trade --> ")
		return rand.Intn(1000), nil
	}
	// url := "http://127.0.0.1:8000/placeLimitOrder?broker=IB"
	url := "http://broker_api:8000/placeLimitOrder?broker=IB"
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

			fmt.Println("curr_pos:", current_pos)
			open := 1.
			// an open position is postive
			// check if we have a selling order
			if trade.Side == "SELL" {
				// if so switch sign of order
				open = -1.
			}
			//convert trade quantity from string to float64
			q, err := strconv.ParseFloat(strings.TrimSpace(trade.Quantity), 64)
			if err != nil {
				// if it errors, you're fucked
				fmt.Println("shit ", err)
				continue
			}
			// check if the direction of the new trade (open*quantity) is equal to the dirction of the current position.
			if math.Signbit(open*q) == math.Signbit(float64(current_pos.Quantity)) {
				// if so, skip the trade because we have an open position
				fmt.Println(open, q, current_pos.Quantity)
				fmt.Printf("%sOpen/Pending order exists, trade skipped: %s - %s - %t \n", workerInfo, trade, i, ok)
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

		// url := fmt.Sprintf("http://127.0.0.1:8000/fills?Id=%d", orderResp.OrderId)
		url := fmt.Sprintf("http://broker_api:8000/fills?Id=%d", orderResp.OrderId)
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
			if fill.Status != "filled" {
				continue
			}
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
	positions.Store(positionId, definitions.Position{
		Symbol:     orderResp.Order.Trade.Symbol,
		Exchange:   orderResp.Order.Trade.Exchange,
		Quantity:   0,
		CostBasis:  0.0,
		Datetime:   time.Now().String(),
		ContractID: int(orderResp.Order.Trade.ContractId),
		Status:     "pending",
	})
	// Marshal to JSON file
	if err := SyncMapToJSONFile(&positions, "/shared/positions.json"); err != nil {
		fmt.Println("Error marshalling sync.Map to JSON:", err)
		return
	}

}

func updatePositionsToFilled(orderResp OrderResponse, costBasis float64, quantity int) {
	fmt.Println("Updating Positions for Filled Order")
	positionId := fmt.Sprintf("%s-%s", orderResp.Order.Trade.StrategyName, orderResp.Order.Trade.Symbol)
	posAdj := 1
	if orderResp.Order.Trade.Side == "SELL" {
		posAdj = -1
	}
	status := "filled"
	positionMap, ok := positions.Load(positionId)
	if ok {
		pos, ok := positionMap.(definitions.Position)
		if ok {

			fmt.Print("Position Map", pos)
			quantity += (posAdj * pos.Quantity)

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

	// Marshal to JSON file
	if err := SyncMapToJSONFile(&positions, "/shared/positions.json"); err != nil {
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
			fmt.Println("Error loading position from JSON -->", v)
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

	// Unmarshal from JSON file
	if err := SyncMapFromJSONFile(&positions, "/shared/positions.json"); err != nil {
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
