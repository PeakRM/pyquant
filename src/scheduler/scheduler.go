package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"scheduler/handlers"
	pb "scheduler/tradepb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// -----------------------------------------------------------------
// Data Structures
// -----------------------------------------------------------------

// Setup represents one named setup in the config
type Setup struct {
	Market     string   `json:"market"`
	ContractId int      `json:"contract_id"`
	Active     bool     `json:"active"`
	Timeframe  string   `json:"timeframe"`
	Schedule   string   `json:"schedule"`
	MarketData []string `json:"market_data"`
}

// Strategy represents one strategy with multiple setups
type Strategy struct {
	ScriptPath   string           `json:"script_path"`
	StrategyType string           `json:"strategy_type"`
	Setups       map[string]Setup `json:"setups"`
}
type Position struct {
	Symbol     string  `json:"symbol"`
	Exchange   string  `json:"exchange"`
	Quantity   int     `json:"quantity"`
	CostBasis  float64 `json:"cost_basis"`
	Datetime   string  `json:"datetime"`
	ContractId int     `json:"contract_id"`
	Status     string  `json:"status"`
}

// strategies is a map of "StrategyName" -> Strategy
var strategies map[string]Strategy
var positions map[string]Position

// Used to signal frontend to refersh strategy config data to mirror backend.
var refreshStrategyConfigChan = make(chan string, 50) // Increase if needed

// Keep track of running processes by "StrategyName|SetupName"
var (
	runningProcs = make(map[string]*exec.Cmd)
	runningMu    sync.Mutex
)

// -----------------------------------------------------------------
// Main
// -----------------------------------------------------------------

func main() {
	// 1. Load from JSON
	shared_strategy_config := GetSharedFilePath("strategy-config.json")
	if err := loadStrategies(shared_strategy_config); err != nil {
		log.Fatalf("Failed to load strategies: %v", err)
	}
	setupShutdown("strategy-config.json")

	// 1a. Start process that checks for unexcpected Strategy Crashes
	// Start monitoring every 30 seconds
	monitorScripts(30 * time.Second)

	// Initialize the database connection
	err := handlers.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer handlers.CloseDB()

	// 2. Handle endpoints
	corsMiddleware := func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			handler.ServeHTTP(w, r)
		})
	}

	http.Handle("/strategies", corsMiddleware(http.HandlerFunc(handleListStrategies)))
	http.Handle("/strategies/", corsMiddleware(http.HandlerFunc(handleStrategyActions)))           // e.g. POST /strategies/{strategyName}/{setupName}/toggle
	http.Handle("/streamPositions", corsMiddleware(http.HandlerFunc(positionStreamHandler)))       // handle positions
	http.Handle("/streamTrades", corsMiddleware(http.HandlerFunc(handlers.SSETradesHandler)))      // handle positions
	http.Handle("/refreshStrategyConfig", corsMiddleware(http.HandlerFunc(refreshStrategyConfig))) // tells front end refresh strategies due to backend changes
	http.Handle("/uploadNewStrategy", corsMiddleware(http.HandlerFunc(newStrategyHandler)))
	http.Handle("/updateSetup", corsMiddleware(http.HandlerFunc(updateSetup)))
	http.Handle("/addSetup", corsMiddleware(http.HandlerFunc(addSetupHandler)))
	http.Handle("/proxy/quote/", corsMiddleware(http.HandlerFunc(proxyQuote)))
	http.Handle("/proxy/historicalData", corsMiddleware(http.HandlerFunc(proxyHistoricalData)))
	http.Handle("/proxy/contractId", corsMiddleware(http.HandlerFunc(proxyContractId)))

	// 3. Serve frontend from ./static/
	http.Handle("/", http.FileServer(http.Dir("./static/react-app/build")))

	// Start server
	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}

// -----------------------------------------------------------------
// Loading & Saving System State
// -----------------------------------------------------------------

// load to strategy config state
func loadStrategyFile(filePath string) (map[string]Strategy, error) {
	temp := make(map[string]Strategy)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return temp, err
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return temp, err
	}
	return temp, nil

}

// load strategy config from file (JSON)
func loadStrategies(filePath string) error {
	temp, err := loadStrategyFile(filePath)
	if err != nil {
		return err
	}
	strategies = temp
	fmt.Println(strategies)
	return nil
}

// load positions from file (JSON)
func loadPositions(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	temp := make(map[string]Position)
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	positions = temp
	fmt.Println(positions)
	return nil
}

// Save updated 'strategies' map to the JSON file
func saveStrategies(filePath string) error {
	data, err := json.MarshalIndent(strategies, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func addStrategyToConfigFile(scriptPath, strategyName, typeVal, setupName, market, timeframe, schedule, additionalData string, contractId int) {
	_, ok := strategies[strategyName]
	if ok {
		log.Print("Strategy already exists, select differnt name.")
		return
	}
	setup := Setup{
		Market:     market,
		ContractId: contractId,
		Active:     false,
		Timeframe:  timeframe,
		Schedule:   schedule,
		MarketData: strings.Split(additionalData, ","),
	}
	setups := map[string]Setup{setupName: setup}
	strategies[strategyName] = Strategy{
		ScriptPath:   scriptPath,
		StrategyType: typeVal,
		Setups:       setups,
	}
	shared_strategy_config := GetSharedFilePath("strategy-config.json")
	// 4) Persist to JSON
	if err := saveStrategies(shared_strategy_config); err != nil {
		log.Println("Failed to save config: ", err.Error())
		return
	}
}

func setupShutdown(strategyFilename string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		err := saveStrategies(strategyFilename)
		if err != nil {
			fmt.Println("Error saving Strategy config to file before shutdown..")
		}
		os.Exit(0)
	}()
}

// -----------------------------------------------------------------
// Route Handlers
// -----------------------------------------------------------------

// handleListStrategies GET /strategies -> returns entire strategies map as JSON
func handleListStrategies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(strategies)
	// fmt.Println("strategies")
}

// handlePositionData GET /strategies -> returns entire position map as JSON
func positionStreamHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			shared_positions := GetSharedFilePath("positions.json")
			if err := loadPositions(shared_positions); err != nil {
				log.Fatalf("Failed to load positions: %v", err)
				continue
			}
			// Marshal positions into JSON
			data, err := json.Marshal(positions)
			if err != nil {
				log.Printf("Failed to marshal positions: %v", err)
				continue
			}
			// SSE requires the "data:" prefix + double newline
			fmt.Fprintf(w, "data: %s\n\n", data)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}

func monitorScripts(checkInterval time.Duration) {
	go func() {
		for {
			time.Sleep(checkInterval)

			runningMu.Lock()
			for key, cmd := range runningProcs {
				if cmd == nil || cmd.ProcessState == nil {
					// fmt.Println(cmd, cmd.Process, cmd.Process.Pid)
					continue
				}

				// Check if process is still running
				if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
					fmt.Printf("process info:\n CMD: %s\n Process:%+v\n State:%s", cmd, cmd.Process, cmd.ProcessState)
					log.Printf("Script %s has stopped unexpectedly", key)

					// Get strategy and setup names
					parts := strings.Split(key, "|")
					if len(parts) != 2 {
						continue
					}
					strategyName := parts[0]
					setupName := parts[1]
					strat, ok := strategies[strategyName]
					if !ok {
						return
					}
					setup, ok := strat.Setups[setupName]
					if !ok {
						return
					}

					setup.Active = false

					// 3) Update the local strategies map
					strat.Setups[setupName] = setup
					strategies[strategyName] = strat
					shared_strategy_config := GetSharedFilePath("strategy-config.json")
					// 4) Persist to JSON
					if err := saveStrategies(shared_strategy_config); err != nil {
						// http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
						return
					}
					refreshStrategyConfigChan <- key
					// Remove from running processes
					delete(runningProcs, key)
				}
			}
			runningMu.Unlock()
		}
	}()
}

func refreshStrategyConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case msg := <-refreshStrategyConfigChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

// handleStrategyActions handles requests like:
// POST /strategies/{strategyName}/{setupName}/toggle
func handleStrategyActions(w http.ResponseWriter, r *http.Request) {
	parts := splitPath(r.URL.Path) // e.g. ["strategies","StrategyA","StrategyA-ZF","toggle"]
	if len(parts) < 2 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	// parts[0] = "strategies"
	// parts[1] = strategyName
	// parts[2] = setupName
	// parts[3] = action

	if len(parts) < 3 {
		http.Error(w, "Setup name required", http.StatusBadRequest)
		return
	}
	strategyName := parts[1]
	setupName := parts[2]

	if len(parts) < 4 {
		http.Error(w, "Action required (toggle)", http.StatusBadRequest)
		return
	}
	action := parts[3]

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	switch action {
	case "toggle":
		toggleSetup(strategyName, setupName, w, r)
	case "close-position":
		closePosition(strategyName, setupName, w, r)
	default:
		http.Error(w, "Unknown action", http.StatusNotFound)
	}
}

func newStrategyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the incoming multipart/form-data (up to 10MB here)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	// Grab form fields (non-file)
	strategyName := r.FormValue("strategyName")
	typeVal := r.FormValue("type")
	setupName := r.FormValue("setupName")
	market := r.FormValue("market")
	contractIdString := r.FormValue("contract_id")
	timeframe := r.FormValue("timeframe")
	schedule := r.FormValue("schedule")
	additionalData := r.FormValue("otherMarketData")

	contractId, err := strconv.Atoi(contractIdString)
	if err != nil {
		fmt.Println("Error converting string:", err)
		contractId = 999999
	}

	// Grab the file from the form data
	file, handler, err := r.FormFile("uploaded_file")
	if err != nil {
		// The user may or may not have uploaded a file. Handle accordingly.
		// If a file is required, respond with an error:
		http.Error(w, "File not found in form data", http.StatusBadRequest)
		return

		// Or if file is optional, you can just set file=nil and skip storing.
		// log.Println("[INFO] No file was uploaded.")
	} else {
		defer file.Close()

		// Create a local directory (inside container) to store the upload.
		uploadDir := "C:/Users/Jon/Projects/pyquant/src/scheduler/strategies"
		// if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		// Check if local directory exists - it should
		if _, err := exists(uploadDir); err != nil {
			http.Error(w, "Unable to find directory", http.StatusInternalServerError)
			return
		}
		log.Printf("[INFO] Found directory: %s\n", uploadDir)

		// Build a full path
		filePath := filepath.Join(uploadDir, handler.Filename)
		log.Printf("[INFO] Saving file to: %s\n", filePath)
		// Create the file on disk
		dst, err := os.Create(filePath)
		log.Println("[INFO] Destination:", dst)

		if err != nil {
			log.Println("[ERROR] Unable to create file:", err)
			http.Error(w, "Unable to create file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Copy the uploaded file to the created file on disk
		_, err = io.Copy(dst, file)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		log.Printf("[INFO] File uploaded successfully: %s\n", handler.Filename)
		// (Optional) Do something with these fields—e.g., store them in a DB:
		log.Printf("[INFO] strategyName=%s, type=%s, setupName=%s, market=%s, contractId=%d, timeframe=%s, schedule=%s, additionalData=%s",
			strategyName, typeVal, setupName, market, contractId, timeframe, schedule, additionalData,
		)
		addStrategyToConfigFile(filePath, strategyName, typeVal, setupName, market, timeframe, schedule, additionalData, contractId)
	}

	// Return a success response
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Data received and processed!")

}

func addSetupHandler(w http.ResponseWriter, r *http.Request) {
	// 1) Parse the request body
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get setupName from form data
	newSetupName := r.FormValue("setupName")
	if newSetupName == "" {
		http.Error(w, "Setup name is required", http.StatusBadRequest)
		return
	}

	// 2) Find the strategy & setup
	//var foundStrategy string
	//var foundSetup Setup
	//mvar strategyFound bool

	strategyName := r.FormValue("strategyName")
	if _, ok := strategies[strategyName].Setups[newSetupName]; ok {
		http.Error(w, "Setup name already exists, enter a different name.", http.StatusBadRequest)
		return
	}

	//for stratName, strat := range strategies {
	// 	if setup, ok := strat.Setups[setupName]; ok {
	// 		foundStrategy = stratName
	// 		foundSetup = setup
	// 		strategyFound = true
	// 		break
	// 	}
	// }

	//if !strategyFound {
	//http.Error(w, "Setup not found in any strategy", http.StatusNotFound)
	//return
	//}
	contractIdFmtd, _ := strconv.Atoi(r.FormValue("contract_id"))
	// 3) Update setup fields from form data
	// Preserve existing values that we don't want to modify
	newSetup := Setup{
		Market:     r.FormValue("market"),
		ContractId: contractIdFmtd,
		Timeframe:  r.FormValue("timeframe"),
		Schedule:   r.FormValue("schedule"),
		MarketData: strings.Split(r.FormValue("market_data"), ","),
		Active:     false,
	}
	fmt.Println(r.FormValue("market_data"))
	fmt.Println(newSetup)

	// 4) Update the local strategies map
	//strat := strategies[foundStrategy]
	//strat.Setups[setupName] = foundSetup
	strategies[strategyName].Setups[newSetupName] = newSetup

	// 5) Persist to JSON
	shared_strategy_config := GetSharedFilePath("strategy-config.json")
	if err := saveStrategies(shared_strategy_config); err != nil {
		http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Println(strategyName, newSetup)
}

func proxyQuote(w http.ResponseWriter, r *http.Request) {
	// Get query params from original request
	parts := splitPath(r.URL.Path) // e.g. ["strategies","StrategyA","StrategyA-ZF","toggle"]
	if len(parts) < 2 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	// parts[0] = "proxy"
	// parts[1] = "quote"
	// parts[2] = exchange
	// parts[3] = contractId

	if len(parts) < 3 {
		http.Error(w, "Contract ID required", http.StatusBadRequest)
		return
	}
	exchange := parts[2]   //string
	contractId := parts[3] //int
	// targetURL := fmt.Sprintf("http://localhost:8000/api/IB/quote/%s/%s", exchange, contractId)
	targetURL := fmt.Sprintf("http://broker_api:8000/api/IB/quote/%s/%s", exchange, contractId)
	// Create a new request
	resp, err := http.Get(targetURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body directly to the response writer
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response: %v", err)
	}
}

func proxyHistoricalData(w http.ResponseWriter, r *http.Request) {
	// Get query params from original request
	// targetURL := "http://localhost:8000/api/IB/historicalData?" + r.URL.RawQuery
	targetURL := "http://broker_api:8000/api/IB/historicalData?" + r.URL.RawQuery

	// Create a new request
	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println("Error: ", err)
		fmt.Println("Request: ", proxyReq)
		return
	}

	// Copy headers
	for header, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(header, value)
		}
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for header, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	io.Copy(w, resp.Body)
	fmt.Println(resp.Body)
}

func proxyContractId(w http.ResponseWriter, r *http.Request) {
	// Get query params from original request
	// targetURL := "http://localhost:8000/api/IB/contract-id"
	targetURL := "http://broker_api:8000/api/IB/contract-id"
	// Create a new request
	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println("Error: ", err)
		fmt.Println("Request: ", proxyReq)
		return
	}

	// Copy headers
	for header, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(header, value)
		}
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for header, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	io.Copy(w, resp.Body)
	fmt.Println(resp.Body)
}

// -----------------------------------------------------------------
// gRPC Client Functions
// -----------------------------------------------------------------

// createTradeServiceClient creates a gRPC client to communicate with the backend service
func createTradeServiceClient() (pb.TradeServiceClient, *grpc.ClientConn, error) {
	// Determine the URL based on environment
	fmt.Println(os.Getenv("ENVIRONMENT"))
	var serverAddr string
	if os.Getenv("ENVIRONMENT") == "production" || os.Getenv("ENVIRONMENT") == "docker" {
		serverAddr = "backend:50051"
	} else {
		serverAddr = "localhost:50051"
	}

	fmt.Println("Connecting to backend at: ", serverAddr)
	// Create a connection to the gRPC server
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to backend: %v", err)
	}
	fmt.Println("Connected to backend")
	// Create a client using the connection
	client := pb.NewTradeServiceClient(conn)
	return client, conn, nil
}

// closePosition handles the close-position endpoint
func closePosition(strategyName, setupName string, w http.ResponseWriter, r *http.Request) {
	// 1) Find the position in the positions map
	// positionId := fmt.Sprintf("%s-%s", strategyName, setupName)
	position, ok := positions[setupName]
	if !ok {
		http.Error(w, "Position not found", http.StatusNotFound)
		return
	}

	// 2) Check if there's an active position to close
	if position.Quantity == 0 {
		http.Error(w, "No active position to close", http.StatusBadRequest)
		return
	}

	// 3) Determine the side for the close order (opposite of current position)
	side := "SELL"
	if position.Quantity < 0 {
		side = "BUY"
	}

	// 4) Create a gRPC client to communicate with the backend
	client, conn, err := createTradeServiceClient()
	if err != nil {
		http.Error(w, "Failed to connect to backend: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// 5) Create a trade message
	trade := &pb.Trade{
		StrategyName: strategyName,
		ContractId:   int32(position.ContractId),
		Exchange:     position.Exchange,
		Symbol:       position.Symbol,
		Side:         side,
		Quantity:     strconv.Itoa(abs(position.Quantity)),
		OrderType:    "MKT", // Use market order for closing positions
		Broker:       "IB",  // Default broker
	}

	// 6) Send the trade to the backend service
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.SendTrade(ctx, trade)
	if err != nil {
		http.Error(w, "Failed to send trade to backend: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 7) Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": resp.Status})
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// -----------------------------------------------------------------
// Strategy Toggle/Update Logic
// -----------------------------------------------------------------

func toggleSetup(strategyName, setupName string, w http.ResponseWriter, r *http.Request) {
	// 1) Find the strategy & setup
	strat, ok := strategies[strategyName]
	if !ok {
		http.Error(w, "Strategy not found", http.StatusNotFound)
		return
	}
	setup, ok := strat.Setups[setupName]
	if !ok {
		http.Error(w, "Setup not found", http.StatusNotFound)
		return
	}

	// 2) If setup.Active == true, we want to stop it
	//    If setup.Active == false, we want to start it
	if setup.Active {
		// Stop
		stopScript(strategyName, setupName)
		// Mark as inactive
		setup.Active = false
	} else {
		// Start
		if err := startScript(strat.ScriptPath, strategyName, setupName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Mark as active
		setup.Active = true
	}

	// 3) Update the local strategies map
	strat.Setups[setupName] = setup
	strategies[strategyName] = strat
	shared_strategy_config := GetSharedFilePath("strategy-config.json")
	// 4) Persist to JSON
	if err := saveStrategies(shared_strategy_config); err != nil {
		http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func updateSetup(w http.ResponseWriter, r *http.Request) {
	// 1) Parse the request body
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get setupName from form data
	setupName := r.FormValue("setupName")
	if setupName == "" {
		http.Error(w, "Setup name is required", http.StatusBadRequest)
		return
	}

	// 2) Find the strategy & setup
	var foundStrategy string
	var foundSetup Setup
	var strategyFound bool

	for stratName, strat := range strategies {
		if setup, ok := strat.Setups[setupName]; ok {
			foundStrategy = stratName
			foundSetup = setup
			strategyFound = true
			break
		}
	}

	if !strategyFound {
		http.Error(w, "Setup not found in any strategy", http.StatusNotFound)
		return
	}

	// 3) Update setup fields from form data
	// Preserve existing values that we don't want to modify
	foundSetup.Market = r.FormValue("market")
	foundSetup.ContractId, _ = strconv.Atoi(r.FormValue("contract_id"))
	foundSetup.Timeframe = r.FormValue("timeframe")
	foundSetup.Schedule = r.FormValue("schedule")
	foundSetup.MarketData = strings.Split(r.FormValue("market_data"), ",")

	// 4) Update the local strategies map
	strat := strategies[foundStrategy]
	strat.Setups[setupName] = foundSetup
	strategies[foundStrategy] = strat

	// 5) Persist to JSON
	shared_strategy_config := GetSharedFilePath("strategy-config.json")
	if err := saveStrategies(shared_strategy_config); err != nil {
		http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 6) If the setup is currently active, restart it with new configuration
	if foundSetup.Active {
		stopScript(foundStrategy, setupName)

		if err := startScript(strat.ScriptPath, foundStrategy, setupName); err != nil {
			http.Error(w, "Failed to restart script: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// -----------------------------------------------------------------
// Start / Stop Script
// -----------------------------------------------------------------

// startScript spawns a python process for the given setup
func startScript(scriptPath, strategyName, setupName string) error {
	runningMu.Lock()
	key := strategyName + "|" + setupName
	if _, exists := runningProcs[key]; exists {
		runningMu.Unlock()
		return nil // already running, no-op
	}
	runningMu.Unlock()

	venvPythonPath, err := GetSharedVenvPath()
	if err != nil {
		return err
	}
	// First, verify the Python interpreter and the script exist
	if err := checkPythonAndScript(venvPythonPath, scriptPath); err != nil {
		fmt.Println(err)
		return err
	}

	cmd := exec.Command(venvPythonPath, scriptPath, setupName)
	fmt.Println(cmd.Process)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	// Pipe stdout to server log
	go func() {
		io.Copy(log.Writer(), stdout)
	}()

	if err := cmd.Start(); err != nil {
		return err
	}
	fmt.Println("Running", scriptPath, setupName)

	runningMu.Lock()
	runningProcs[key] = cmd
	runningMu.Unlock()

	return nil
}

// stopScript kills the process if it's running
func stopScript(strategyName, setupName string) {
	runningMu.Lock()
	defer runningMu.Unlock()

	key := strategyName + "|" + setupName
	cmd, exists := runningProcs[key]
	if !exists {
		return
	}
	_ = cmd.Process.Kill() // ignoring error for brevity
	delete(runningProcs, key)
}

// checkPythonAndScript verifies the Python executable and script are present.
func checkPythonAndScript(venvPythonPath, scriptPath string) error {
	if !pathExists(venvPythonPath) {
		return fmt.Errorf("python interpreter not found at path: %s", venvPythonPath)
	}
	if !pathExists(scriptPath) {
		return fmt.Errorf("python script not found at path: %s", scriptPath)
	}
	return nil
}

// -----------------------------------------------------------------
// Utility Funtions
// -----------------------------------------------------------------

func splitPath(p string) []string {
	filtered := []string{}
	for _, segment := range splitOnSlash(p) {
		if segment != "" {
			filtered = append(filtered, segment)
		}
	}
	return filtered
}

func splitOnSlash(p string) []string {
	start := 0
	var res []string
	for i, c := range p {
		if c == '/' {
			if i > start {
				res = append(res, p[start:i])
			}
			start = i + 1
		}
	}
	if start < len(p) {
		res = append(res, p[start:])
	}
	return res
}

// Helper function
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// pathExists checks whether a given file path exists.
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func GetSharedVenvPath() (string, error) {
	if pathExists("/usr/local/bin/python") {
		return "/usr/local/bin/python", nil
	}
	if pathExists("C:/Users/Jon/Projects/pyquant/.venv") {
		return "C:/Users/Jon/Projects/pyquant/.venv/Scripts/python.exe", nil
	}
	return "", fmt.Errorf("no python interpreter path found")

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
