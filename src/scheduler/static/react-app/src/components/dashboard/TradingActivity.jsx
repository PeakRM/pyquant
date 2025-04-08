import React, { useState, useRef, useEffect } from 'react';
import { ChevronDown, ChevronUp } from 'lucide-react';

/**
 * Trading Activity Component
 * A collapsible card with tabs for positions and trades
 */
const TradingActivityComponent = ({ 
  positions, 
  trades, 
  initialTab = 'positions', 
  initialCollapsed = false
}) => {
  const [isCollapsed, setIsCollapsed] = useState(initialCollapsed);
  const [activeTab, setActiveTab] = useState(initialTab);
  const [containerHeight, setContainerHeight] = useState('auto');
  const containerRef = useRef(null);
  
  // Measure and update the container height when needed
  useEffect(() => {
    if (containerRef.current && !isCollapsed) {
      const height = containerRef.current.scrollHeight;
      setContainerHeight(`${height}px`);
    }
  }, [positions, trades, isCollapsed, activeTab]);
  
  return (
    <div className="bg-white rounded-lg shadow-md overflow-hidden mb-6">
      {/* Header with tabs */}
      <div 
        className="px-6 py-4 bg-gray-50 border-b border-gray-200 flex justify-between items-center"
        onClick={() => setIsCollapsed(!isCollapsed)}
      >
        <div>
          <h2 className="text-xl font-semibold text-gray-800">Trading Activity</h2>
          <p className="text-sm text-gray-500">
            {activeTab === 'positions' 
              ? `${Object.keys(positions).length} active positions` 
              : 'Last 24 hours'}
          </p>
        </div>
        
        <div className="flex items-center">
          {/* Tab Toggle */}
          <div className="bg-gray-50 rounded-full p-1 flex mr-4">
            <button 
              className={`px-3 py-1 text-xs rounded-full border transition-colors ${
                activeTab === 'positions' 
                  ? 'bg-blue-600 text-white' 
                  : 'text-gray-300 hover:text-black'
              }`}
              onClick={(e) => {
                e.stopPropagation();
                setActiveTab('positions');
              }}
            >
              Positions
            </button>
            <button 
              className={`px-3 py-1 text-xs rounded-full border transition-colors ${
                activeTab === 'trades' 
                  ? 'bg-blue-600 text-white' 
                  : 'text-gray-300 hover:text-black'
              }`}
              onClick={(e) => {
                e.stopPropagation();
                setActiveTab('trades');
              }}
            >
              Trades
            </button>
          </div>
          
          {isCollapsed ? 
            <ChevronDown size={18} className="text-gray-400" /> : 
            <ChevronUp size={18} className="text-gray-400" />}
        </div>
      </div>
      
      {/* Content container with smooth transition */}
      <div 
        ref={containerRef}
        className="overflow-x-auto overflow-hidden transition-all duration-300 ease-in-out"
        style={{ 
          maxHeight: isCollapsed ? '0px' : containerHeight,
          opacity: isCollapsed ? 0 : 1
        }}
      >
        {activeTab === 'positions' ? (
          <PositionsTable positions={positions} />
        ) : (
          <TradesTable trades={trades} />
        )}
      </div>
    </div>
  );
};

/**
 * Positions Table Component
 */
const PositionsTable = ({ positions }) => {
  return (
    <table className="min-w-full divide-y divide-gray-200 text-sm">
      <thead className="bg-gray-50 text-left text-xs uppercase">
        <tr>
          <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Symbol</th>
          <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Pos</th>
          <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Avg Entry</th>
          {/* <th className="px-3 py-2">Current</th> */}
          <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Unrealized P&L</th>
          {/* <th className="px-3 py-2">Realized P&L</th> */}
          <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Value</th>
          <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Setup</th>
        </tr>
      </thead>
      <tbody className="bg-white divide-y divide-gray-200">
        {Object.entries(positions).map(([setupName, position], index) => (
          <tr 
            key={setupName} 
            className= {index % 2 === 1 ? 'bg-gray-50' : 'bg-white'}
          >
            <td className="px-3 py-2 font-medium text-center">{position.symbol}</td>
            <td className="px-3 py-2 text-center" >
              <span className={position.quantity > 0 ? 'text-green-500' : position.quantity < 0 ? 'text-red-500' : 'text-gray-500'}>
                {position.quantity > 0 ? '+' : ''}{position.quantity}
              </span>
            </td>
            <td className="px-3 py-2 text-gray-500 text-center">{position.cost_basis?.toFixed(2)}</td>
            {/* <td className="px-3 py-2 text-gray-300">{position.current_price?.toFixed(2)}</td> */}
            <td className="px-3 py-2 text-center">
              <span className={position.unrealized >= 0 ? 'text-green-500' : 'text-red-500'}>
                ${position.unrealized?.toFixed(2)}
              </span>
            </td>
            {/* <td className="px-3 py-2">
              <span className={position.realized >= 0 ? 'text-green-500' : position.realized < 0 ? 'text-red-500' : 'text-gray-400'}>
                ${position.realized?.toFixed(2)}
              </span>
            </td> */}
            <td className="px-3 py-2 text-gray-500 text-center" >${position.value?.toFixed(2)}</td>
            <td className="px-3 py-2 text-center">
              <div className="flex flex-wrap gap-1 text-center">
                <span className="bg-blue-900 bg-opacity-30 text-blue-400 text-xs px-2 py-0.5 rounded-full text-center">
                  {setupName}
                </span>
              </div>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
};

/**
 * Trades Table Component
 */
const TradesTable = ({ trades }) => {
  return (
    <table className="w-full text-sm">
      <thead className="bg-gray-50 text-left text-xs uppercase">
        <tr>
          <th className="px-3 py-2">ID</th>
          <th className="px-3 py-2">Strategy</th>
          <th className="px-3 py-2">Market</th>
          <th className="px-3 py-2">Side</th>
          <th className="px-3 py-2">Type</th>
          <th className="px-3 py-2">Qty</th>
          <th className="px-3 py-2">Price</th>
          <th className="px-3 py-2">Status</th>
          <th className="px-3 py-2">Time</th>
        </tr>
      </thead>
      <tbody className="divide-y divide-gray-700">
        {Object.entries(trades).map(([setupName, trade], index) => (
          <tr 
            key={setupName} 
            className={index % 2 === 0 ? 'bg-gray-800' : 'bg-gray-750'}
          >
            <td className="px-3 py-2 font-medium text-gray-300">{trade.id}</td>
            <td className="px-3 py-2">{trade.strategy}</td>
            <td className="px-3 py-2 text-gray-300">{trade.market}</td>
            <td className="px-3 py-2">
              <span className={trade.side === 'BUY' ? 'text-green-500' : 'text-red-500'}>
                {trade.side}
              </span>
            </td>
            <td className="px-3 py-2 text-gray-300">{trade.type}</td>
            <td className="px-3 py-2 text-gray-300">{trade.quantity}</td>
            <td className="px-3 py-2 text-gray-300">${trade.price.toFixed(2)}</td>
            <td className="px-3 py-2">
              <span className={`text-xs px-2 py-0.5 rounded-full
                ${trade.status === 'Filled' ? 'bg-green-900 bg-opacity-30 text-green-400' : 
                  trade.status === 'Cancelled' ? 'bg-red-900 bg-opacity-30 text-red-400' : 
                  'bg-yellow-900 bg-opacity-30 text-yellow-400'}`}>
                {trade.status}
              </span>
            </td>
            <td className="px-3 py-2 text-gray-300">{trade.timestamp}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
};

export { PositionsTable, TradesTable };
export default TradingActivityComponent;