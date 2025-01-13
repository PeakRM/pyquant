package main

import (
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
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
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
	if err := loadStrategies("/shared/strategy-config.json"); err != nil {
		log.Fatalf("Failed to load strategies: %v", err)
	}

	// 2. Handle endpoints
	http.HandleFunc("/strategies", handleListStrategies)
	// e.g. POST /strategies/{strategyName}/{setupName}/toggle
	http.HandleFunc("/strategies/", handleStrategyActions)
	// handle positions
	http.HandleFunc("/streamPositions", positionStreamHandler)

	http.HandleFunc("/uploadNewStrategy", newStrategyHandler)

	// 3. Serve frontend from ./static/
	http.Handle("/", http.FileServer(http.Dir("./static")))

	// Start server
	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// -----------------------------------------------------------------
// JSON Loading & Saving
// -----------------------------------------------------------------

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

func loadStrategies(filePath string) error {
	temp, err := loadStrategyFile(filePath)
	if err != nil {
		return err
	}
	strategies = temp
	fmt.Println(strategies)
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

// -----------------------------------------------------------------
// Handlers
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

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			if err := loadPositions("/shared/positions.json"); err != nil {
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
	default:
		http.Error(w, "Unknown action", http.StatusNotFound)
	}
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
	additionalData := r.FormValue("additionalData")

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
		uploadDir := "/app/strategies"
		// if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		// Check if local directory exists - it should
		if _, err := exists(uploadDir); err != nil {
			http.Error(w, "Unable to find directory", http.StatusInternalServerError)
			return
		}
		// 	http.Error(w, "Unable to create upload directory", http.StatusInternalServerError)
		// 	return
		// }  if not make it.

		// Build a full path
		filePath := filepath.Join(uploadDir, handler.Filename)
		dst, err := os.Create(filePath)
		if err != nil {
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
	// 4) Persist to JSON
	if err := saveStrategies("/shared/strategy-config.json"); err != nil {
		log.Printf("Failed to save config: " + err.Error())
		return
	}
}

// -----------------------------------------------------------------
// Toggle Logic
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

	// 4) Persist to JSON
	if err := saveStrategies("/shared/strategy-config.json"); err != nil {
		http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
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

	venvPythonPath := "/usr/local/bin/python"

	// First, verify the Python interpreter and the script exist
	if err := checkPythonAndScript(venvPythonPath, scriptPath); err != nil {
		fmt.Println(err)
		// ex, err := os.Getwd()
		// if err != nil {
		// 	panic(err)
		// }
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

// pathExists checks whether a given file path exists.
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
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
// Utility for splitting paths
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
