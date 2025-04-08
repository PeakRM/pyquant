import React from 'react';
import { Play, Pause, Settings, X } from 'lucide-react';
import { TrendingUp, Activity } from 'lucide-react';

const SetupRow = ({
  setupName,
  setup,
  position,
  isSelected,
  onSelect,
  onToggleSetup,
  onEditSetup,
  onClosePosition
}) => {
  const performanceValue = position?.unrealized || 0;

  // Format percentage value
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
    <tr
      className={`hover:bg-gray-50 cursor-pointer ${isSelected ? 'bg-blue-50' : ''}`}
      onClick={() => onSelect(setupName)}
    >
      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{setupName}</td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{setup.market}</td>
      <td className="px-6 py-4 whitespace-nowrap text-center">
        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${setup.active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
          {setup.active ? 'Active' : 'Inactive'}
        </span>
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-center text-gray-500">{setup.timeframe}</td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-center text-gray-500">{setup.schedule || 'N/A'}</td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-center text-gray-500">
        {position?.quantity ? `${position.quantity} @ ${position.cost_basis?.toFixed(2)}` : 'No position'}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-center">
        <div className="flex items-center justify-center">
          {getPerformanceIndicator(performanceValue)}
          <span className={`ml-1 text-sm ${performanceValue > 0 ? 'text-green-600' : performanceValue < 0 ? 'text-red-600' : 'text-gray-500'}`}>
            {formatDollar(performanceValue)}
          </span>
        </div>
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-center text-sm font-medium">
        <div className="flex justify-center space-x-2" onClick={(e) => e.stopPropagation()}>
          <button
            onClick={(e) => {
              e.stopPropagation();
              onToggleSetup(setupName);
            }}
            className={`p-1.5 rounded-full ${setup.active ? 'bg-red-100 text-red-600 hover:bg-red-200' : 'bg-green-100 text-green-600 hover:bg-green-200'}`}
          >
            {setup.active ? <Pause size={16} /> : <Play size={16} />}
          </button>
          <button
            onClick={(e) => {
              e.stopPropagation();
              onEditSetup(setupName);
            }}
            className="p-1.5 bg-gray-100 text-gray-600 rounded-full hover:bg-gray-200"
          >
            <Settings size={16} />
          </button>
          {position?.quantity !== 0 && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                onClosePosition(setupName);
              }}
              className="p-1.5 bg-yellow-100 text-yellow-600 rounded-full hover:bg-yellow-200"
              title="Close Position"
            >
              <X size={16} />
            </button>
          )}
        </div>
      </td>
    </tr>
  );
};

export default SetupRow;

