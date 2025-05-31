import React from 'react';

/**
 * KPI Metrics Dashboard
 * Displays a set of key performance indicators in a grid layout
 */
export const KPIMetricsDashboard = ({ metrics }) => {
  return (
    <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-6">
      {metrics.map((metric, index) => (
        <KPIMetricCard key={index} metric={metric} />
      ))}
    </div>
  );
};

/**
 * KPI Metric Card
 * Individual card showing a single KPI with title, value and trend
 */
export const KPIMetricCard = ({ metric }) => {
  if (!metric) return null; // Add this guard
  const { title, value, change, isPositive } = metric;
  const getValueColor=() => {
    if (title.includes('P&L') || title.includes('PnL')){
      return isPositive ? 'text-green-600':'text-red-600';
    }
    return 'text-gray-800';
  }
  const formatCurrency = (value) => {
    const num = parseFloat(value);
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(num);
  };
 

 return (
    <div className="bg-white rounded-lg p-4 shadow-sm border border-gray-200">
      <h3 className="text-gray-600 text-sm font-medium">{title}</h3>
      <div className="flex items-end mt-1">
      {/*  <span className="text-2xl font-bold text-gray-800">{value}</span>*/}

	<span className={`text-2xl font-bold ${getValueColor()}`}>
  		{formatCurrency(value)}
	</span>
      {change && (
          <span className={`ml-2 text-sm ${isPositive ? 'text-green-500' : 'text-red-500'}`}>
            {change}
          </span>
        )}
      </div>
    </div>
  );
};

/**
 * Strategy KPI Bar
 * Displays KPIs specific to strategy performance in a horizontal layout
 */
export const StrategyKPIBar = ({ activeStrategies, totalStrategies, totalPositions, unrealizedPnl, realizedPnl }) => {
  return (
    <div className="flex-1 flex justify-around border-l border-r border-gray-200 px-6">
      <KPIItem
        icon="Layers"
        label="Active"
        value={<>{activeStrategies} <span className="text-xs text-gray-500">/ {totalStrategies}</span></>}
      />
      
      <KPIItem 
        icon="Activity" 
        label="Positions" 
        value={totalPositions} 
      />
      
      <KPIItem 
        icon="DollarSign" 
        label="Unrealized P&L" 
        value={`$${unrealizedPnl.toFixed(2)}`}
        valueClass={unrealizedPnl >= 0 ? 'text-green-600' : 'text-red-600'}
      />
      
      <KPIItem 
        icon="DollarSign" 
        label="Realized P&L" 
        value={`$${realizedPnl.toFixed(2)}`}
        valueClass={realizedPnl >= 0 ? 'text-green-600' : 'text-red-600'}
      />
    </div>
  );
};

/**
 * KPI Item
 * Individual KPI item with icon and value
 */
export const KPIItem = ({ icon, label, value, valueClass = '' }) => {
  // Note: This component expects Lucide icons to be passed as strings
  // In your implementation, you'll need to either:
  // 1. Import all needed icons at the top and reference them here
  // 2. Pass the actual icon component instead of a string
  
  return (
    <div className="flex items-center">
      <div className="bg-blue-50 rounded-full p-1.5 mr-2">
        {/* Placeholder for icon */}
        <div className="w-4 h-4 text-blue-600"></div>
      </div>
      <div>
        <span className="text-xs text-gray-500">{label}</span>
        <p className={`font-medium text-gray-800 ${valueClass}`}>{value}</p>
      </div>
    </div>
  );
};

export default {
  KPIMetricsDashboard,
  KPIMetricCard,
  StrategyKPIBar,
  KPIItem
};
