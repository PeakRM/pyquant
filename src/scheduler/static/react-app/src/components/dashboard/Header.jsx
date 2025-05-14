import React from 'react';
import { BarChart2, Plus } from 'lucide-react';

const Header = ({ onAddStrategy, onOpenContractTool }) => {
  return (
    <header className="bg-white shadow-md">
      <div className="max-w-7xl mx-auto px-4 py-4 sm:px-6 lg:px-8 flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900 flex items-center">
          <BarChart2 className="mr-2" /> 
          Dashboard
        </h1>
        <div className="flex space-x-2">
          <button 
            onClick={onAddStrategy} 
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition flex items-center"
          >
            <Plus size={18} className="mr-1" /> Add Strategy
          </button>
          <button 
            onClick={onOpenContractTool}
            className="px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300 transition"
          >
            Contract ID Tool
          </button>
        </div>
      </div>
    </header>
  );
};

export default Header;
