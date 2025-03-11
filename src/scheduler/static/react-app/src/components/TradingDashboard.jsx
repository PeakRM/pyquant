import React, { useState, useEffect } from 'react';
import { AlertCircle, Play, Pause, Settings, Plus, X, RefreshCw, TrendingUp, BarChart2, Activity, Calendar } from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

const TICK_VALUE_MAP = new Map([
  ["MES", 5.0],
  ["MGC", 10.0],
  
]);

// Main Dashboard Component
export default function TradingDashboard() {
  // State management
  const [strategies, setStrategies] = useState({});
  const [positions, setPositions] = useState({});
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
  const SCHEDULER_API_BASE = window.location.hostname === 'localhost' ? 'http://localhost:8080' : 'http://scheduler';

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
        
        if (position.quantity !== 0) {
          const conId = position.contract_id;
          if (!priceMap.has(conId)) {
            try {
              const exchange = position.exchange;
              const response = await fetch(`${SCHEDULER_API_BASE}/proxy/quote/${exchange}/${conId}`);
              const data = await response.json();
              priceMap.set(conId, data.last);
              
              // Calculate unrealized value
              position.unrealized = ((priceMap.get(conId) - position.cost_basis )*TICK_VALUE_MAP.get(position.symbol)) * 
              Math.sign(position.quantity);
            } catch (error) {
              console.error('Error fetching price:', error);
            }
          } else {
            position.unrealized = ((priceMap.get(conId) - position.cost_basis )*TICK_VALUE_MAP.get(position.symbol)) * 
              Math.sign(position.quantity);
            
          }
        }
      }
      
      setCurrentPrices(priceMap);
      setPositions(newPositions);
    };
    
    // Set up strategy config refresh streaming
    const refreshSource = new EventSource(`${SCHEDULER_API_BASE}/refreshStrategyConfig`);
    refreshSource.onmessage = (event) => {
      console.log("Strategy update notification:", event.data);
      fetchStrategies();
    };
    
    return () => {
      positionSource.close();
      refreshSource.close();
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
    
    // Optimistically toggle the active state
    const currentActive = updatedStrategies[strategyName].setups[setupName].active;
    updatedStrategies[strategyName].setups[setupName].active = !currentActive;
    
    // Update state immediately (UI responds instantly)
    setStrategies(updatedStrategies);
    
    try {
      // Make API call in background
      await fetch(`${SCHEDULER_API_BASE}/strategies/${strategyName}/${setupName}/toggle`, { method: 'POST' });
    } catch (error) {
      console.error("Failed to toggle setup:", error);
      // Revert the change if the API call fails
      updatedStrategies[strategyName].setups[setupName].active = currentActive;
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

  // Format percentage value
  const formatPercent = (value) => {
    if (value === null || value === undefined) return '';
    return `${(value * 100).toFixed(2)}%`;
  };
  const formatDollar = (value) => {
    if (value === null || value === undefined) return '';
    return `$${(value).toFixed(2)}`;
  };

  // Display performance indicator based on value
  const getPerformanceIndicator = (value) => {
    if (value === null || value === undefined) return null;
    
    if (value > 0.02) return <TrendingUp className="text-green-500" />;
    if (value < -0.02) return <TrendingUp className="text-red-500 transform rotate-180" />;
    return <Activity className="text-gray-500" />;
  };
  
  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-200 text-gray-800">
      {/* Header */}
      <header className="bg-white shadow-md">
        <div className="max-w-7xl mx-auto px-4 py-4 sm:px-6 lg:px-8 flex justify-between items-center">
          <h1 className="text-2xl font-bold text-gray-900 flex items-center">
            <BarChart2 className="mr-2" /> 
            Dashboard
          </h1>
          <div className="flex space-x-2">
            <button 
              onClick={() => setIsNewStrategyModalOpen(true)} 
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition flex items-center"
            >
              <Plus size={18} className="mr-1" /> Add Strategy
            </button>
            <button 
              onClick={() => {
                setIsSidebarOpen(true);
                setContractResult("");
              }}
              className="px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300 transition"
            >
              Contract ID Tool
            </button>
          </div>
        </div>
      </header>
      
      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 py-6 sm:px-6 lg:px-8">
        {/* Chart Section */}
        <div className="bg-white rounded-lg shadow-md p-4 mb-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-lg font-medium text-gray-800">
              {selectedSetup ? `Chart: ${selectedSetup.setupName} (${selectedSetup.market})` : 'Select a trading setup to view chart'}
            </h2>
          </div>
          
          <div className="h-64">
            {chartLoading ? (
              <div className="flex justify-center items-center h-full">
                <RefreshCw className="animate-spin text-blue-500" size={40} />
              </div>
            ) : chartData.length > 0 ? (
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis 
                    dataKey="date" 
                    angle={-45}
                    tickFormatter={(value) => {
                      const date = new Date(value);
                      // return date.toLocaleDateString();
                      const timeFormatter = new Intl.DateTimeFormat('en-US', {
                        hour: '2-digit',
                        minute: '2-digit',
                        // second: '2-digit',
                        hour12: false // Use 24-hour format
                      });
                      
                      const formattedTime = timeFormatter.format(date);
                      return formattedTime;
                    }}
                  />
                  <YAxis type="number" domain={['auto', 'auto']} />
                  <Tooltip />
                  <Legend />
                  <Line type="monotone" dataKey="close" stroke="#3B82F6" name="Close Price" dot={false} />
                </LineChart>
              </ResponsiveContainer>
            ) : (
              <div className="flex justify-center items-center h-full text-gray-500">
                <Calendar className="mr-2" /> Select a trading setup to view historical data
              </div>
            )}
          </div>
        </div>
      
        {loading ? (
          <div className="flex justify-center py-20">
            <RefreshCw className="animate-spin text-blue-500" size={40} />
          </div>
        ) : (
          <div className="space-y-6">
            {Object.keys(strategies).length === 0 ? (
              <div className="bg-white p-8 rounded-lg shadow-md text-center">
                <AlertCircle className="mx-auto text-gray-400 mb-4" size={48} />
                <h3 className="text-lg font-medium">No strategies found</h3>
                <p className="mt-2 text-gray-500">Add your first strategy to get started</p>
              </div>
            ) : (
              Object.entries(strategies).map(([strategyName, strategy]) => (
                <div key={strategyName} className="bg-white rounded-lg shadow-md overflow-hidden">
                  <div className="px-6 py-4 bg-gray-50 border-b border-gray-200 flex justify-between items-center">
                    <div>
                      <h2 className="text-xl font-semibold text-gray-800">{strategyName}</h2>
                      <p className="text-sm text-gray-500">{strategy.strategy_type} · {Object.keys(strategy.setups).length} setups</p>
                    </div>
                    <button 
                      onClick={() => openAddSetupModal(strategyName)}
                      className="px-3 py-1 bg-blue-100 text-blue-700 rounded-md hover:bg-blue-200 transition flex items-center text-sm"
                    >
                      <Plus size={16} className="mr-1" /> Add Setup
                    </button>
                  </div>
                  
                  <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-gray-200">
                      <thead className="bg-gray-50">
                        <tr>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Setup</th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Market</th>
                          <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                          <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Timeframe</th>
                          <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Schedule</th>
                          <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Position</th>
                          <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Performance</th>
                          <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                        </tr>
                      </thead>
                      <tbody className="bg-white divide-y divide-gray-200">
                        {Object.entries(strategy.setups).map(([setupName, setup]) => {
                          const position = positions[setupName] || { quantity: 0 };
                          const performanceValue = position.unrealized || 0.;
                          const isSelected = selectedSetup && selectedSetup.setupName === setupName;
                          
                          return (
                            <tr 
                              key={setupName} 
                              className={`hover:bg-gray-50 cursor-pointer ${isSelected ? 'bg-blue-50' : ''}`}
                              onClick={() => selectSetupForChart(strategyName, setupName)}
                            >
                              <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{setupName}</td>
                              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{setup.market}</td>
                              <td className="px-6 py-4 whitespace-nowrap text-center">
                                <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${setup.active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
                                  {setup.active ? 'Active' : 'Inactive'}
                                </span>
                              </td>
                              <td className="px-6 py-4 whitespace-nowrap text-sm text-center text-gray-500">{setup.timeframe}</td>
                              <td className="px-6 py-4 whitespace-nowrap text-sm text-center text-gray-500">{setup.schedule}</td>
                              <td className="px-6 py-4 whitespace-nowrap text-sm text-center font-medium">
                                {position.quantity !== 0 ? (
                                  <span className={position.quantity > 0 ? 'text-green-600' : 'text-red-600'}>
                                    {position.quantity}
                                  </span>
                                ) : (
                                  <span className="text-gray-400">-</span>
                                )}
                              </td>
                              <td className="px-6 py-4 whitespace-nowrap text-center">
                                <div className="flex items-center justify-center space-x-1">
                                  {getPerformanceIndicator(performanceValue)}
                                  <span className={`text-sm ${performanceValue > 0 ? 'text-green-600' : performanceValue < 0 ? 'text-red-600' : 'text-gray-500'}`}>
                                    {formatDollar(performanceValue)}
                                  </span>
                                </div>
                              </td>
                              <td className="px-6 py-4 whitespace-nowrap text-center text-sm font-medium" onClick={(e) => e.stopPropagation()}>
                                <div className="flex items-center justify-center space-x-2">
                                  <button
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      toggleSetup(strategyName, setupName);
                                    }}
                                    className={`p-1 rounded-full ${setup.active ? 'bg-red-100 hover:bg-red-200 text-red-600' : 'bg-green-100 hover:bg-green-200 text-green-600'}`}
                                    title={setup.active ? 'Stop' : 'Start'}
                                  >
                                    {setup.active ? <Pause size={18} /> : <Play size={18} />}
                                  </button>
                                  <button
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      openEditSetupModal(strategyName, setupName);
                                    }}
                                    className="p-1 rounded-full bg-gray-100 hover:bg-gray-200 text-gray-600"
                                    title="Configure"
                                  >
                                    <Settings size={18} />
                                  </button>
                                </div>
                              </td>
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  </div>
                </div>
              ))
            )}
          </div>
        )}
      </main>
      
      {/* New Strategy Modal */}
      {isNewStrategyModalOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg shadow-xl max-w-md w-full relative max-h-screen overflow-y-auto">
            <button 
              onClick={() => setIsNewStrategyModalOpen(false)}
              className="absolute top-4 right-4 text-gray-400 hover:text-gray-600"
            >
              <X size={24} />
            </button>
            
            <h2 className="text-xl font-semibold mb-4">Add New Strategy</h2>
            
            <form onSubmit={handleNewStrategySubmit} className="space-y-4">
              <div>
                <label className="block font-medium mb-1" htmlFor="fileInput">Upload Script</label>
                <input 
                  className="block w-full border rounded p-2" 
                  type="file" 
                  id="fileInput" 
                  name="uploaded_file"
                  required
                />
              </div>
              
              <div>
                <label className="block font-medium mb-1" htmlFor="strategyName">Strategy Name</label>
                <input 
                  className="block w-full border rounded p-2" 
                  type="text" 
                  id="strategyName" 
                  name="strategyName"
                  required
                />
              </div>
              
              <div>
                <label className="block font-medium mb-1" htmlFor="typeSelect">Type</label>
                <select 
                  className="block w-full border rounded p-2" 
                  id="typeSelect" 
                  name="type"
                >
                  <option value="Rebalance">Rebalance</option>
                  <option value="Alpha">Alpha</option>
                  <option value="Other">Other</option>
                </select>
              </div>
              
              <h3 className="font-semibold text-lg mt-6">Initial Setup</h3>
              
              <div>
                <label className="block font-medium mb-1" htmlFor="setupName">Setup Name</label>
                <input 
                  className="block w-full border rounded p-2" 
                  type="text" 
                  id="setupName" 
                  name="setupName"
                  required
                />
              </div>
              
              <div>
                <label className="block font-medium mb-1" htmlFor="market">Market</label>
                <input 
                  className="block w-full border rounded p-2" 
                  type="text" 
                  id="market" 
                  name="market"
                  required
                />
              </div>
              
              <div>
                <label className="block font-medium mb-1" htmlFor="contract_id">Contract ID</label>
                <input 
                  className="block w-full border rounded p-2" 
                  type="text" 
                  id="contract_id" 
                  name="contract_id"
                  required
                />
              </div>
              
              <div>
                <label className="block font-medium mb-1" htmlFor="timeframe">Timeframe</label>
                <input 
                  className="block w-full border rounded p-2" 
                  type="text" 
                  id="timeframe" 
                  name="timeframe"
                  required
                />
              </div>
              
              <div>
                <label className="block font-medium mb-1" htmlFor="schedule">Schedule</label>
                <input 
                  className="block w-full border rounded p-2" 
                  type="text" 
                  id="schedule" 
                  name="schedule"
                  required
                />
              </div>
              
              <div>
                <label className="block font-medium mb-1" htmlFor="additionalData">Additional Data (comma separated)</label>
                <input 
                  className="block w-full border rounded p-2" 
                  type="text" 
                  id="additionalData" 
                  name="additionalData"
                />
              </div>
              
              <button 
                type="submit" 
                className="w-full py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition mt-2"
              >
                Add Strategy
              </button>
            </form>
          </div>
        </div>
      )}
      
      {/* Edit Setup Modal */}
      {isEditSetupModalOpen && selectedSetup && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg shadow-xl max-w-md w-full relative">
            <button 
              onClick={() => setIsEditSetupModalOpen(false)}
              className="absolute top-4 right-4 text-gray-400 hover:text-gray-600"
            >
              <X size={24} />
            </button>
            
            <h2 className="text-xl font-semibold mb-4">Edit Setup Configuration</h2>
            
            <form onSubmit={handleEditSetupSubmit} className="space-y-4">
              <div>
                <label className="block font-medium mb-1" htmlFor="editSetupName">Setup Name</label>
                <input 
                  className="block w-full border rounded p-2 bg-gray-100" 
                  type="text" 
                  id="editSetupName" 
                  name="setupName"
                  defaultValue={selectedSetup.setupName}
                  readOnly
                />
              </div>
              
              <div>
                <label className="block font-medium mb-1" htmlFor="editMarket">Market</label>
                <input 
                  className="block w-full border rounded p-2 bg-gray-100" 
                  type="text" 
                  id="editMarket" 
                  name="market"
                  defaultValue={selectedSetup.market}
                  readOnly
                />
              </div>
              
              <div>
                <label className="block font-medium mb-1" htmlFor="editContractId">Contract ID</label>
                <input 
                  className="block w-full border rounded p-2" 
                  type="text" 
                  id="editContractId" 
                  name="contract_id"
                  defaultValue={selectedSetup.contractId}
                />
              </div>
              
              <div>
                <label className="block font-medium mb-1" htmlFor="editTimeframe">Timeframe</label>
                <input 
                  className="block w-full border rounded p-2" 
                  type="text" 
                  id="editTimeframe" 
                  name="timeframe"
                  defaultValue={selectedSetup.timeframe}
                />
              </div>
              
              <div>
                <label className="block font-medium mb-1" htmlFor="editSchedule">Schedule</label>
                <input 
                  className="block w-full border rounded p-2" 
                  type="text" 
                  id="editSchedule" 
                  name="schedule"
                  defaultValue={selectedSetup.schedule}
                />
              </div>
              
              <div>
                <label className="block font-medium mb-1" htmlFor="editOtherMarketData">Other Market Data</label>
                <input 
                  className="block w-full border rounded p-2" 
                  type="text" 
                  id="editOtherMarketData" 
                  name="otherMarketData"
                  defaultValue={selectedSetup.marketData ? selectedSetup.marketData.join(',') : ''}
                />
              </div>
              
              <button 
                type="submit" 
                className="w-full py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition mt-2"
              >
                Save Changes
              </button>
            </form>
          </div>
        </div>
      )}
      
      {/* Add Setup Modal */}
      {isAddSetupModalOpen && selectedStrategy && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg shadow-xl max-w-md w-full relative">
            <button 
              onClick={() => setIsAddSetupModalOpen(false)}
              className="absolute top-4 right-4 text-gray-400 hover:text-gray-600"
            >
              <X size={24} />
            </button>
            
            <h2 className="text-xl font-semibold mb-4">Add New Setup</h2>
            
            <form onSubmit={handleAddSetupSubmit} className="space-y-4">
              <div>
                <label className="block font-medium mb-1" htmlFor="addStrategyName">Strategy</label>
                <input 
                  className="block w-full border rounded p-2 bg-gray-100" 
                  type="text" 
                  id="addStrategyName" 
                  name="strategyName"
                  defaultValue={selectedStrategy}
                  readOnly
                />
              </div>
              
              <div>
                <label className="block font-medium mb-1" htmlFor="addSetupName">Setup Name</label>
                <input 
                  className="block w-full border rounded p-2" 
                  type="text" 
                  id="addSetupName" 
                  name="setupName"
                  defaultValue={`${selectedStrategy}-`}
                  required
                />
              </div>
              
              <div>
                <label className="block font-medium mb-1" htmlFor="addMarket">Market</label>
                <input 
                  className="block w-full border rounded p-2" 
                  type="text" 
                  id="addMarket" 
                  name="market"
                  required
                />
              </div>
              
              <div>
                <label className="block font-medium mb-1" htmlFor="addContractId">Contract ID</label>
                <input 
                  className="block w-full border rounded p-2" 
                  type="text" 
                  id="addContractId" 
                  name="contract_id"
                  required
                />
              </div>
              
              <div>
                <label className="block font-medium mb-1" htmlFor="addTimeframe">Timeframe</label>
                <input 
                  className="block w-full border rounded p-2" 
                  type="text" 
                  id="addTimeframe" 
                  name="timeframe"
                  required
                />
              </div>
              
              <div>
                <label className="block font-medium mb-1" htmlFor="addSchedule">Schedule</label>
                <input 
                  className="block w-full border rounded p-2" 
                  type="text" 
                  id="addSchedule" 
                  name="schedule"
                  required
                />
              </div>
              
              <div>
                <label className="block font-medium mb-1" htmlFor="addOtherMarketData">Other Market Data</label>
                <input 
                  className="block w-full border rounded p-2" 
                  type="text" 
                  id="addOtherMarketData" 
                  name="otherMarketData"
                />
              </div>
              
              <button 
                type="submit" 
                className="w-full py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition mt-2"
              >
                Add Setup
              </button>
            </form>
          </div>
        </div>
      )}
      
      {/* Contract ID Tool Sidebar */}
      {isSidebarOpen && (
        <div className="fixed inset-0 z-50 flex">
          <div 
            className="fixed inset-0 bg-black bg-opacity-50" 
            onClick={() => setIsSidebarOpen(false)}
          ></div>
          
          <div className="fixed top-0 right-0 h-full w-80 bg-white shadow-lg overflow-y-auto z-10 p-6">
            <button 
              onClick={() => setIsSidebarOpen(false)}
              className="absolute top-4 right-4 text-gray-400 hover:text-gray-600"
            >
              <X size={24} />
            </button>
            
            <h2 className="text-xl font-semibold mb-6">Get Contract ID (IBKR)</h2>
            
            <form id="contractForm" className="space-y-4" onSubmit={handleGetContractId}>
              <div>
                <label className="block text-sm font-medium mb-1">Symbol</label>
                <input 
                  type="text" 
                  name="symbol" 
                  className="w-full border rounded p-2"
                  required
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium mb-1">Contract Type</label>
                <select 
                  name="contract_type" 
                  className="w-full border rounded p-2"
                >
                  <option value="FUT">FUT</option>
                  <option value="STK">STK</option>
                </select>
              </div>
              
              <div>
                <label className="block text-sm font-medium mb-1">Exchange</label>
                <input 
                  type="text" 
                  name="exchange" 
                  className="w-full border rounded p-2"
                  required
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium mb-1">Currency</label>
                <input 
                  type="text" 
                  name="currency" 
                  defaultValue="USD"
                  className="w-full border rounded p-2"
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium mb-1">Expiry</label>
                <input 
                  type="text" 
                  name="expiry" 
                  className="w-full border rounded p-2"
                />
              </div>
              
              <button 
                type="submit" 
                className="w-full py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition mt-4"
              >
                Get Contract ID
              </button>
            </form>
            
            <div id="result" className={`mt-6 p-4 bg-gray-100 rounded ${!contractResult ? 'hidden' : ''}`}>
              {contractResult && (
                <div>
                  <h3 className="font-medium mb-2">Result:</h3>
                  <pre className="whitespace-pre-wrap text-sm">
                    {JSON.stringify(contractResult, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}