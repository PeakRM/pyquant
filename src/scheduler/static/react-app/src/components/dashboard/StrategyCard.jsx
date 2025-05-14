import React, { useState } from 'react';
import { Plus, ChevronUp, ChevronDown } from 'lucide-react';
import SetupRow from './SetupRow';

const StrategyCard = ({
  strategyName,
  strategy,
  positions,
  selectedSetup,
  onSelectSetup,
  onToggleSetup,
  onEditSetup,
  onAddSetup,
  onClosePosition
}) => {
  const [isStrategyListCollapsed, setIsStrategyListCollapsed] = useState(false);

  return (
    <div className="bg-white rounded-lg shadow-md overflow-hidden">
      <div className="px-6 py-4 bg-gray-50 border-b border-gray-200 flex justify-between items-center"
                     onClick={() => setIsStrategyListCollapsed(!isStrategyListCollapsed)}>
        <div>
          <h2 className="text-xl font-semibold text-gray-800">{strategyName}</h2>
          <p className="text-sm text-gray-500">{strategy.strategy_type} · {Object.keys(strategy.setups).length} setups</p>
        </div>
        <div className="flex items-center ml-2">
          <button
            onClick={() => onAddSetup(strategyName)}
            className="px-3 py-1.5 bg-blue-100 text-blue-700 rounded-lg hover:bg-blue-200 transition flex items-center text-sm mr-3"
          >
            <Plus size={16} className="mr-1" /> Add Setup
          </button>
            {isStrategyListCollapsed ? 
              <ChevronDown size={20} className="text-gray-400" /> : 
              <ChevronUp size={20} className="text-gray-400" />}
        </div>
      </div>

      <div className={`overflow-x-auto ${isStrategyListCollapsed ? 'hidden' : 'block'}`}>
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
              const isSelected = selectedSetup && selectedSetup.setupName === setupName;

              return (
                <SetupRow
                  key={setupName}
                  setupName={setupName}
                  setup={setup}
                  position={position}
                  isSelected={isSelected}
                  onSelect={() => onSelectSetup(strategyName, setupName)}
                  onToggleSetup={() => onToggleSetup(strategyName, setupName)}
                  onEditSetup={() => onEditSetup(strategyName, setupName)}
                  onClosePosition={() => onClosePosition(strategyName, setupName)}
                />
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default StrategyCard;

