import React, { useState } from 'react';

/**
 * Trading Activity Component
 * A card with tabs for positions and trades
 */
const TradingActivityComponent = ({
  positions,
  trades,
  initialTab = 'positions'
}) => {
  const [activeTab, setActiveTab] = useState(initialTab);

  // Count positions with non-zero quantity
  const activePositionsCount = Object.values(positions).filter(position => position.quantity !== 0).length;

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden h-full">
      {/* Header with tabs */}
      <div
        className="p-3 flex items-center justify-between bg-gray-50 border-b border-gray-200"
      >
        <div className="flex items-center">
        <h2 className="text-base font-small text-gray-700">Trading Activity</h2>
        <span className="ml-2 text-gray-500 text-base">
            {activeTab === 'positions'
              ? `${activePositionsCount} active positions`
              : 'Last 24 hours'}
          </span>
        </div>

        <div className="flex items-center">
          {/* Tab Toggle */}
          <div className="bg-gray-100 rounded-full p-1 flex">
            <button
              className={`px-3 py-1 text-xs rounded-full transition-colors ${
                activeTab === 'positions'
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-700 hover:text-gray-900'
              }`}
              onClick={() => setActiveTab('positions')}
            >
              Positions
            </button>
            <button
              className={`px-3 py-1 text-xs rounded-full transition-colors ${
                activeTab === 'trades'
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-700 hover:text-gray-900'
              }`}
              onClick={() => setActiveTab('trades')}
            >
              Trades
            </button>
          </div>
        </div>
      </div>

      {/* Content container */}
      <div
        className="overflow-x-auto"
        style={{
          height: '275px',
          overflowY: 'auto'
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
  // Filter out positions with quantity equal to 0
  const activePositions = Object.entries(positions).filter(([_, position]) => position.quantity !== 0);

  return (
    <table className="w-full text-sm">
      <thead className="bg-gray-50 text-left text-xs uppercase sticky top-0 z-10">
      <tr>
          <th className="px-3 py-2 text-gray-500 font-small">Market</th>
          <th className="px-3 py-2 text-gray-500 font-small">Pos</th>
          <th className="px-3 py-2 text-gray-500 font-small">Avg Entry</th>
          {/* <th className="px-3 py-2">Current</th> */}
          <th className="px-3 py-2 text-gray-500 font-small">Unrealized P&L</th>
          {/* <th className="px-3 py-2">Realized P&L</th> */}
          {/* <th className="px-3 py-2 text-gray-500 font-small">Value</th> */}
          <th className="px-3 py-2 text-gray-500 font-small">Setup</th>
        </tr>
      </thead>
      <tbody className="divide-y divide-gray-200">
        {activePositions.map(([setupName, position], index) => (
          <tr
            key={setupName}
            className= {index % 2 === 0 ? 'bg-white' : 'bg-gray-50'}
          >
            <td className="px-3 py-2 font-small text-gray-700">{position.symbol}</td>
            <td className="px-3 py-2 text-center" >
              <span className={position.quantity > 0 ? 'text-green-600' : 'text-red-600'}>
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
            {/* <td className="px-3 py-2 text-gray-700 text-center" >${position.value?.toFixed(2)}</td> */}
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
  // Sort trades by updated_at in descending order (newest first)
  const sortedTrades = Object.entries(trades)
    .sort(([, tradeA], [, tradeB]) => {
      return new Date(tradeB.updated_at) - new Date(tradeA.updated_at);
    });

  return (
    <table className="w-full text-sm">
      <thead className="bg-gray-50 text-left text-xs uppercase sticky top-0 z-10">
        <tr>
          <th className="px-3 py-2 text-gray-500 font-small">ID</th>
          <th className="px-3 py-2 text-gray-500 font-small">Strategy</th>
          <th className="px-3 py-2 text-gray-500 font-small">Market</th>
          <th className="px-3 py-2 text-gray-500 font-small">Side</th>
          {/* <th className="px-3 py-2">Type</th> */}
          <th className="px-3 py-2 text-gray-500 font-small">Qty</th>
          <th className="px-3 py-2 text-gray-500 font-small">Price</th>
          <th className="px-3 py-2 text-gray-500 font-small">Status</th>
          <th className="px-3 py-2 text-gray-500 font-small">Time</th>
        </tr>
      </thead>
      <tbody className="divide-y divide-gray-200">
        {sortedTrades.map(([setupName, trade], index) => (
          <tr
            key={setupName}
            className={index % 2 === 0 ?  'bg-white' : 'bg-gray-50'}
          >
            <td className="px-3 py-2 font-small text-gray-700">{trade.broker_order_id}</td>
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