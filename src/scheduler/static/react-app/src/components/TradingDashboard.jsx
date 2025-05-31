import React, { useState, useEffect } from 'react';

// Import components
import Header from './dashboard/Header';
import ChartSection from './dashboard/ChartSection';
import StrategyList from './dashboard/StrategyList';
import TradingActivityComponent from './dashboard/TradingActivity';
import { NewStrategyModal, AddSetupModal, EditSetupModal, ContractIdSidebar } from './dashboard/Modals';
import { KPIMetricsDashboard } from './dashboard/KPI';

const TICK_VALUE_MAP = new Map([
  ["MES", 5.0],
  ["MGC", 10.0],
  ["MCL", 100.0],
  ["MYM", 1.0],
  ["MNQ", 2.0],
]);
const sampleTrades = [
  {
    id: "ORD-42381",
    strategy: "VBO-MES",
    market: "CME:MES",
    side: "BUY",
    type: "MARKET",
    quantity: 1,
    price: 5020.25,
    status: "FILLED",
    timestamp: "13:24:36"
  },
  // ... more trades

];

const sampleKPIMetrics = [
{
  title: "Strategies",
  value: "3",
  change: "+10%",
  isPositive: true
},
{
  title: "Total Positions",
  value: "10",
  change: "-5%",
  isPositive: false
},
{
  title: "Trade Count",
  value: "3",
  change: "+10%",
  isPositive: true
},
{
  title: "Realized PnL",
  value: "10",
  change: "-5%",
  isPositive: false
},
];
// Main Dashboard Component
export default function TradingDashboard() {
  // State management
  const [strategies, setStrategies] = useState({});
  const [positions, setPositions] = useState({});
  const [trades, setTrades] = useState({});
  const [currentPrices, setCurrentPrices] = useState(new Map());
  const [isNewStrategyModalOpen, setIsNewStrategyModalOpen] = useState(false);
  const [isEditSetupModalOpen, setIsEditSetupModalOpen] = useState(false);
  const [isAddSetupModalOpen, setIsAddSetupModalOpen] = useState(false);
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);
  const [selectedSetup, setSelectedSetup] = useState(null);
  const [selectedStrategy, setSelectedStrategy] = useState(null);
  const [loading, setLoading] = useState(true);
  const [chartData, setChartData] = useState([]);
  const [chartLoading, setChartLoading] = useState(false);
  const [contractResult, setContractResult] = useState(null);
  const SCHEDULER_API_BASE = window.location.hostname === 'localhost' ? 'http://localhost:8080' : '';
  const [kpiMetrics, setKPIMetrics] = useState({
    maintMarginReq: { title: '', value: '', change: '', isPositive: false },
    netLiquidation: { title: '', value: '', change: '', isPositive: false },
    unrealizedPnl: { title: '', value: '', change: '', isPositive: false },
    realizedPnL: { title: '', value: '', change: '', isPositive: false }
  });
  // Fetch strategies on component mount
  useEffect(() => {
    fetchStrategies();

    // Set up position streaming
    const positionSource = new EventSource(`${SCHEDULER_API_BASE}/streamPositions`);

    positionSource.onmessage = async (event) => {
      const newPositions = JSON.parse(event.data);
      const priceMap = new Map(currentPrices);

      for (const setupName in newPositions) {
        const position = newPositions[setupName];
        const tick_value = TICK_VALUE_MAP.has(position.symbol) ? TICK_VALUE_MAP.get(position.symbol) : 1.0
        if (position.quantity !== 0) {
          const conId = position.contract_id;
          if (!priceMap.has(conId)) {
            try {
              const exchange = position.exchange;
              const response = await fetch(`${SCHEDULER_API_BASE}/proxy/quote/${exchange}/${conId}`);
              const data = await response.json();
              priceMap.set(conId, data.last);
              // Calculate unrealized value
              position.unrealized = ((priceMap.get(conId) - position.cost_basis )*tick_value) * Math.sign(position.quantity);
            } catch (error) {
              console.error('Error fetching price:', error);
            }
          } else {
            position.unrealized = ((priceMap.get(conId) - position.cost_basis )*tick_value) * Math.sign(position.quantity);

          }
        }
      }

      setCurrentPrices(priceMap);
      setPositions(newPositions);
    };

    // Set up trade streaming
    const tradeSource = new EventSource(`${SCHEDULER_API_BASE}/streamTrades`);
    tradeSource.onmessage = (event) => {
      const newTrades = JSON.parse(event.data);
      setTrades(newTrades);
    };


    // Set up strategy config refresh streaming
    const refreshSource = new EventSource(`${SCHEDULER_API_BASE}/refreshStrategyConfig`);
    refreshSource.onmessage = (event) => {
      console.log("Strategy update notification:", event.data);
      fetchStrategies();
    };

    // const { title, value, change, isPositive } = metric;
    // Set up KPI metrics streaming
    const kpiSource = new EventSource(`${SCHEDULER_API_BASE}/streamKPIMetrics`);
    kpiSource.onmessage = (event) => {
      console.log('Raw event data:', event.data); // Add this line
      try {
        const newMetrics = JSON.parse(event.data);
        console.log('Parsed metrics:', newMetrics); // Check structure
        setKPIMetrics(newMetrics);
      } catch (error) {
        console.error('Error parsing KPI metrics:', error);
      }
    };

    return () => {
      positionSource.close();
      refreshSource.close();
      kpiSource.close();
      tradeSource.close();
    };
  }, []);

  // Fetch strategies from backend
  const fetchStrategies = async () => {
    setLoading(true);
    try {
      const response = await fetch(`${SCHEDULER_API_BASE}/strategies`);
      const data = await response.json();
      console.log(data);
      setStrategies(data);
    } catch (error) {
      console.error("Failed to fetch strategies:", error);
    } finally {
      setLoading(false);
    }
  };

  // Fetch historical data for selected setup
  const fetchHistoricalData = async (strategyName, setupName) => {
    setChartLoading(true);

    const setup = strategies[strategyName].setups[setupName];

    try {
      // Calculate dates based on timeframe
      const endDate = new Date();
      let startDate = new Date();
      const barSize = setup.timeframe; // Assuming timeframe matches bar_size format

      // Set start date based on timeframe
      if (barSize.includes('min')) {
        startDate.setDate(startDate.getDate() - 1); // Last 1 day for minute bars
      } else if (barSize.includes('hour')) {
        startDate.setDate(startDate.getDate() - 7); // Last 7 days for hourly bars
      } else if (barSize.includes('day')) {
        startDate.setMonth(startDate.getMonth() - 3); // Last 3 months for daily bars
      } else {
        startDate.setMonth(startDate.getMonth() - 6); // Default to last 6 months
      }

      const startTime = startDate.toISOString();
      const endTime = endDate.toISOString();

      const payload = {
        contract_id: setup.contract_id,
        exchange: setup.market.split(":")[0],
        currency: "USD"
      };

      const url = `${SCHEDULER_API_BASE}/proxy/historicalData?start_time=${startTime}&end_time=${endTime}&bar_size=${barSize}`;
      console.log(url);
      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(payload)
      });

      if (response.ok) {
        const data = await response.json();

        // Format data for chart
        const formattedData = data.map(bar => ({
          date: new Date(bar.timestamp).toLocaleString(),
          open: bar.open,
          high: bar.high,
          low: bar.low,
          close: bar.close,
          volume: bar.volume,
        }));

        setChartData(formattedData);
      } else {
        console.error("Failed to fetch historical data:", await response.text());
      }
    } catch (error) {
      console.error("Error fetching historical data:", error);
    } finally {
      setChartLoading(false);
    }
  };

  // Toggle a strategy setup on/off with optimistic update
  const toggleSetup = async (strategyName, setupName) => {
    // Create a copy of strategies to modify
    const updatedStrategies = {...strategies};

    // Optimistically toggle the enabled state
    const currentEnabled = updatedStrategies[strategyName].setups[setupName].enabled;
    updatedStrategies[strategyName].setups[setupName].enabled = !currentEnabled;

    // Update state immediately (UI responds instantly)
    setStrategies(updatedStrategies);

    try {
      // Make API call in background
      await fetch(`${SCHEDULER_API_BASE}/strategies/${strategyName}/${setupName}/toggle`, { method: 'POST' });
    } catch (error) {
      console.error("Failed to toggle setup:", error);
      // Revert the change if the API call fails
      updatedStrategies[strategyName].setups[setupName].enabled = currentEnabled;
      setStrategies(updatedStrategies);
    }
  };

  // Select a setup for charting
  const selectSetupForChart = (strategyName, setupName) => {
    console.log(strategyName, setupName);
    setSelectedSetup({
      strategyName,
      setupName,
      ...strategies[strategyName].setups[setupName]
    });

    fetchHistoricalData(strategyName, setupName);
  };

  // Open edit setup modal with selected setup data
  const openEditSetupModal = (strategyName, setupName) => {
    setSelectedSetup({
      strategyName,
      setupName,
      ...strategies[strategyName].setups[setupName]
    });
    setIsEditSetupModalOpen(true);
  };

  // Open add setup modal for a strategy
  const openAddSetupModal = (strategyName) => {
    setSelectedStrategy(strategyName);
    setIsAddSetupModalOpen(true);
  };

  // Handle new strategy form submission
  const handleNewStrategySubmit = async (e) => {
    e.preventDefault();

    const formData = new FormData(e.target);

    try {
      const response = await fetch(`${SCHEDULER_API_BASE}/uploadNewStrategy`, {
        method: 'POST',
        body: formData
      });

      if (response.ok) {
        setIsNewStrategyModalOpen(false);
        fetchStrategies();
      } else {
        const errorText = await response.text();
        alert("Failed to add strategy: " + errorText);
      }
    } catch (error) {
      console.error("Error adding strategy:", error);
    }
  };

  // Handle edit setup form submission
  const handleEditSetupSubmit = async (e) => {
    e.preventDefault();

    const formData = new FormData(e.target);
    const urlEncodedData = new URLSearchParams(formData);

    try {
      const response = await fetch(`${SCHEDULER_API_BASE}/updateSetup`, {
        method: 'POST',
        headers: {'Content-Type': 'application/x-www-form-urlencoded'},
        body: urlEncodedData
      });

      if (response.ok) {
        setIsEditSetupModalOpen(false);
        fetchStrategies();
      } else {
        const errorText = await response.text();
        alert("Failed to update setup: " + errorText);
      }
    } catch (error) {
      console.error("Error updating setup:", error);
    }
  };

  // Handle add setup form submission
  const handleAddSetupSubmit = async (e) => {
    e.preventDefault();

    const formData = new FormData(e.target);
    const urlEncodedData = new URLSearchParams(formData);

    try {
      const response = await fetch(`${SCHEDULER_API_BASE}/addSetup`, {
        method: 'POST',
        headers: {'Content-Type': 'application/x-www-form-urlencoded'},
        body: urlEncodedData
      });

      if (response.ok) {
        setIsAddSetupModalOpen(false);
        fetchStrategies();
      } else {
        const errorText = await response.text();
        alert("Failed to add setup: " + errorText);
      }
    } catch (error) {
      console.error("Error adding setup:", error);
    }
  };

  // Handle add setup form submission
  const handleGetContractId = async (e) => {
    e.preventDefault();
    const formData = new FormData(e.target);
    const payload = Object.fromEntries(formData);
    console.log('Sending request to:', `${SCHEDULER_API_BASE}/proxy/contractId`);
    console.log('With payload:', payload);

    try {
      const response = await fetch(`${SCHEDULER_API_BASE}/proxy/contractId`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        // credentials: 'include',
        body: JSON.stringify(payload)
      });

      const result = await response.json();
      setContractResult(result); // Store the result in state
    } catch (error) {
      console.error('Error:', error);
      setContractResult({ error: error.message }); // Store error in state
    }
  };

  // Close position for a setup
  const closePosition = async (strategyName, setupName) => {
    if (!window.confirm(`Are you sure you want to close the position for ${setupName}?`)) {
      return;
    }

    try {
      const response = await fetch(`${SCHEDULER_API_BASE}/strategies/${strategyName}/${setupName}/close-position`, {
        method: 'POST'
      });

      if (response.ok) {
        alert(`Position for ${setupName} closed successfully`);
        // Refresh positions - in a real implementation, the EventSource would update this
        // but we'll force a refresh for now
        fetchStrategies();
      } else {
        const errorText = await response.text();
        alert(`Failed to close position: ${errorText}`);
      }
    } catch (error) {
      console.error('Error closing position:', error);
      alert(`Error closing position: ${error.message}`);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-200 text-gray-800">
      {/* Header */}
      <Header
        onAddStrategy={() => setIsNewStrategyModalOpen(true)}
        onOpenContractTool={() => {
          setIsSidebarOpen(true);
          setContractResult("");
        }}
      />

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 py-6 sm:px-6 lg:px-8">
        {/* KPI Metrics Dashboard */}
        <KPIMetricsDashboard 
          metrics={[
            kpiMetrics.netLiquidation,
            kpiMetrics.maintMarginReq,
            kpiMetrics.realizedPnl,
            kpiMetrics.unrealizedPnl
          ]} 
        />

        {/* Trading Activity and Chart Section side by side */}
        <div className="flex flex-col md:flex-row md:space-x-6 mb-6">
          {/* Trading Activity Component */}
          <div className="w-full md:w-1/2">
            <TradingActivityComponent
              positions={positions}
              trades={trades}
              initialTab="positions"
            />
          </div>
          {/* Chart Section */}
          <div className="w-full md:w-1/2">
            <ChartSection
              selectedSetup={selectedSetup}
              chartData={chartData}
              chartLoading={chartLoading}
            />
          </div>
        </div>

        {/* Strategy List */}
        <StrategyList
          loading={loading}
          strategies={strategies}
          positions={positions}
          selectedSetup={selectedSetup}
          onSelectSetup={selectSetupForChart}
          onToggleSetup={toggleSetup}
          onEditSetup={openEditSetupModal}
          onAddSetup={openAddSetupModal}
          onClosePosition={closePosition}
        />
      </main>

      {/* Modals */}
      <NewStrategyModal
        isOpen={isNewStrategyModalOpen}
        onClose={() => setIsNewStrategyModalOpen(false)}
        onSubmit={handleNewStrategySubmit}
      />

      <EditSetupModal
        isOpen={isEditSetupModalOpen}
        onClose={() => setIsEditSetupModalOpen(false)}
        onSubmit={handleEditSetupSubmit}
        setup={selectedSetup}
        strategyName={selectedStrategy}
      />

      <AddSetupModal
        isOpen={isAddSetupModalOpen}
        onClose={() => setIsAddSetupModalOpen(false)}
        onSubmit={handleAddSetupSubmit}
        strategyName={selectedStrategy}
      />

      <ContractIdSidebar
        isOpen={isSidebarOpen}
        onClose={() => setIsSidebarOpen(false)}
        onSubmit={handleGetContractId}
        contractResult={contractResult}
      />
    </div>
  );
}
