import React from 'react';
import { RefreshCw, Calendar } from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

const ChartSection = ({ selectedSetup, chartData, chartLoading }) => {
  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden h-full">
      <div className="p-3 flex items-center justify-between bg-gray-50 border-b border-gray-200">
        <h2 className="text-base font-small text-gray-700">
          {selectedSetup ? `Chart: ${selectedSetup.setupName} (${selectedSetup.market})` : 'Market Chart'}
        </h2>
      </div>

      <div className="h-[275px]">
        {chartLoading ? (
          <div className="flex justify-center items-center h-full">
            <RefreshCw className="animate-spin text-blue-500" size={40} />
          </div>
        ) : chartData.length > 0 ? (
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={chartData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }} >
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis
                dataKey="date"
                angle={-45}
                tickFormatter={(value) => {
                  const date = new Date(value);
                  const timeFormatter = new Intl.DateTimeFormat('en-US', {
                    hour: '2-digit',
                    minute: '2-digit',
                    hour12: false
                  });

                  const formattedTime = timeFormatter.format(date);
                  return formattedTime;
                }}
              />
              <YAxis type="number" domain={['auto', 'auto']} />
              <Tooltip />
              <Legend />
              <Line type="monotone"
                    dataKey="close"
                    stroke="#3B82F6"
                    name="Close Price"
                    strokeWidth={2}
                    dot={false}
                    activeDot={{ r: 6 }}  />
            </LineChart>
          </ResponsiveContainer>
        ) : (
          <div className="flex justify-center items-center h-full text-gray-500">
            <Calendar className="mr-2" /> Select a trading setup to view historical data
          </div>
        )}
      </div>
    </div>
  );
};

export default ChartSection;
