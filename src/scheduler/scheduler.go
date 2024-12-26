package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
)

// -----------------------------------------------------------------
// Data Structures
// -----------------------------------------------------------------

// Setup represents one named setup in the config
type Setup struct {
	Market     string   `json:"market"`
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

// strategies is a map of "StrategyName" -> Strategy
var strategies map[string]Strategy

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

	// 3. Serve frontend from ./static/
	http.Handle("/", http.FileServer(http.Dir("./static")))

	// Start server
	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// -----------------------------------------------------------------
// JSON Loading & Saving
// -----------------------------------------------------------------

func loadStrategies(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	temp := make(map[string]Strategy)
	if err := json.Unmarshal(data, &temp); err != nil {
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

	cmd := exec.Command(venvPythonPath, scriptPath)
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
	fmt.Println("Running", scriptPath)

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
