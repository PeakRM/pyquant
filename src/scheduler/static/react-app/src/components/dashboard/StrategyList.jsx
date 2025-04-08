import React from 'react';
import { RefreshCw, AlertCircle } from 'lucide-react';
import StrategyCard from './StrategyCard';

const StrategyList = ({
  loading,
  strategies,
  positions,
  selectedSetup,
  onSelectSetup,
  onToggleSetup,
  onEditSetup,
  onAddSetup,
  onClosePosition
}) => {
  if (loading) {
    return (
      <div className="flex justify-center py-20">
        <RefreshCw className="animate-spin text-blue-500" size={40} />
      </div>
    );
  }

  if (Object.keys(strategies).length === 0) {
    return (
      <div className="bg-white p-8 rounded-lg shadow-md text-center">
        <AlertCircle className="mx-auto text-gray-400 mb-4" size={48} />
        <h3 className="text-lg font-medium">No strategies found</h3>
        <p className="mt-2 text-gray-500">Add your first strategy to get started</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {Object.entries(strategies).map(([strategyName, strategy]) => (
        <StrategyCard
          key={strategyName}
          strategyName={strategyName}
          strategy={strategy}
          positions={positions}
          selectedSetup={selectedSetup}
          onSelectSetup={onSelectSetup}
          onToggleSetup={onToggleSetup}
          onEditSetup={onEditSetup}
          onAddSetup={onAddSetup}
          onClosePosition={onClosePosition}
        />
      ))}
    </div>
  );
};

export default StrategyList;

