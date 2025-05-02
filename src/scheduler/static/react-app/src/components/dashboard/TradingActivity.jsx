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
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden mb-6">
      {/* Header with tabs */}
      <div
        className="p-3 flex items-center justify-between cursor-pointer bg-gray-50 border-b border-gray-200"
        onClick={() => setIsCollapsed(!isCollapsed)}
      >
        <div className="flex items-center">
        <h2 className="text-lg font-medium text-gray-700">Trading Activity</h2>
        <span className="ml-2 text-gray-500 text-sm">
            {activeTab === 'positions'
              ? `${Object.keys(positions).length} active positions`
              : 'Last 24 hours'}
          </span>
        </div>

        <div className="flex items-center">
          {/* Tab Toggle */}
          <div className="bg-gray-100 rounded-full p-1 flex mr-4">
            <button
              className={`px-3 py-1 text-xs rounded-full transition-colors ${
                activeTab === 'positions'
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-700 hover:text-gray-900'
              }`}
              onClick={(e) => {
                e.stopPropagation();
                setActiveTab('positions');
              }}
            >
              Positions
            </button>
            <button
              className={`px-3 py-1 text-xs rounded-full transition-colors ${
                activeTab === 'trades'
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-700 hover:text-gray-900'
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
            <ChevronDown size={18} className="text-gray-500" /> :
            <ChevronUp size={18} className="text-gray-500" />}
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
    <table className="w-full text-sm">
      <thead className="bg-gray-50 text-left text-xs uppercase">
      <tr>
          <th className="px-3 py-2 text-gray-500 font-medium">Market</th>
          <th className="px-3 py-2 text-gray-500 font-medium">Pos</th>
          <th className="px-3 py-2 text-gray-500 font-medium">Avg Entry</th>
          {/* <th className="px-3 py-2">Current</th> */}
          <th className="px-3 py-2 text-gray-500 font-medium">Unrealized P&L</th>
          {/* <th className="px-3 py-2">Realized P&L</th> */}
          <th className="px-3 py-2 text-gray-500 font-medium">Value</th>
          <th className="ppx-3 py-2 text-gray-500 font-medium">Setup</th>
        </tr>
      </thead>
      <tbody className="divide-y divide-gray-200">
        {Object.entries(positions).map(([setupName, position], index) => (
          <tr
            key={setupName}
            className= {index % 2 === 0 ? 'bg-white' : 'bg-gray-50'}
          >
            <td className="px-3 py-2 font-medium text-gray-700">{position.symbol}</td>
            <td className="px-3 py-2 text-center" >
              <span className={position.quantity > 0 ? 'text-green-600' : position.quantity < 0 ? 'text-red-600' : 'text-gray-500'}>
                {position.quantity > 0 ? '+' : ''}{position.quantity}
              </span>
            </td>
            <td className="px-3 py-2 text-gray-700 text-center">{position.cost_basis?.toFixed(2)}</td>
            {/* <td className="px-3 py-2 text-gray-700 text-center">{position.current_price?.toFixed(2)}</td> */}
            <td className="px-3 py-2 text-center">
              <span className={position.unrealized >= 0 ? 'text-green-600' : 'text-red-600'}>
                ${position.unrealized?.toFixed(2)}
              </span>
            </td>
            {/* <td className="px-3 py-2">
              <span className={position.realized >= 0 ? 'text-green-600' : position.realized < 0 ? 'text-red-600' : 'text-gray-500'}>
                ${position.realized?.toFixed(2)}
              </span>
            </td> */}
            <td className="px-3 py-2 text-gray-700 text-center" >${position.value?.toFixed(2)}</td>
            <td className="px-3 py-2 text-center">
              <div className="flex flex-wrap gap-1 text-center">
                <span className="bg-blue-100 text-blue-700 text-xs px-2 py-0.5 rounded-full text-center">
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
          <th className="px-3 py-2 text-gray-500 font-medium">ID</th>
          <th className="px-3 py-2 text-gray-500 font-medium">Strategy</th>
          <th className="px-3 py-2 text-gray-500 font-medium">Market</th>
          <th className="px-3 py-2 text-gray-500 font-medium">Side</th>
          {/* <th className="px-3 py-2">Type</th> */}
          <th className="px-3 py-2 text-gray-500 font-medium">Qty</th>
          <th className="px-3 py-2 text-gray-500 font-medium">Price</th>
          <th className="px-3 py-2 text-gray-500 font-medium">Status</th>
          <th className="px-3 py-2 text-gray-500 font-medium">Time</th>
        </tr>
      </thead>
      <tbody className="divide-y divide-gray-200">
        {Object.entries(trades).map(([setupName, trade], index) => (
          <tr
            key={setupName}
            className={index % 2 === 0 ?  'bg-white' : 'bg-gray-50'}
          >
            <td className="px-3 py-2 font-medium text-gray-700">{trade.broker_order_id}</td>
            <td className="px-3 py-2 text-gray-700">{trade.strategy_name}</td>
            <td className="px-3 py-2 text-gray-700">{trade.exchange}:{trade.symbol}</td>
            <td className="px-3 py-2">
              <span className={trade.side === 'BUY' ? 'text-green-600' : 'text-red-600'}>
                {trade.side}
              </span>
            </td>
            {/* <td className="px-3 py-2 text-gray-300">{trade.type}</td> */}
            <td className="px-3 py-2 text-gray-700">{trade.quantity}</td>
            <td className="px-3 py-2 text-gray-700">${trade.price.toFixed(2)}</td>
            <td className="px-3 py-2">
              <span className={`text-xs px-2 py-0.5 rounded-full
                ${trade.status.toUpperCase() === 'FILLED' ?  'bg-green-100 text-green-700' :
                  trade.status.toUpperCase() === 'CANCELLED' ? 'bg-red-100 text-red-700':
                  'bg-yellow-100 text-yellow-700'}`}>
                {trade.status.toUpperCase()}
              </span>
            </td>
            <td className="px-3 py-2 text-gray-700">{new Date(trade.updated_at).toLocaleTimeString()}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
};

export { PositionsTable, TradesTable };
export default TradingActivityComponent;