import React from 'react';
import { X } from 'lucide-react';

// New Strategy Modal
export const NewStrategyModal = ({ isOpen, onClose, onSubmit }) => {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="fixed inset-0 bg-black bg-opacity-50" onClick={onClose}></div>
      <div className="bg-white rounded-lg shadow-xl w-full max-w-md max-h-[90vh] overflow-y-auto relative z-10">
        <div className="sticky top-0 bg-white border-b p-4 flex justify-between items-center">
          <h2 className="text-xl font-semibold">Add New Strategy</h2>
          <button onClick={onClose} className="text-gray-500 hover:text-gray-700 p-1">
            <X size={20} />
          </button>
        </div>
        
        <form onSubmit={onSubmit}>
          <div className="p-4 space-y-4">

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
            
            <div className="mt-4">
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
            
            <div className="mt-4">
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
                id="otherMarketData" 
                name="otherMarketData"
              />
            </div>
            <button 
              type="submit" 
              className="w-full py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition mt-6"
            >
              Create Strategy
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

// Add Setup Modal
export const AddSetupModal = ({ isOpen, onClose, onSubmit, strategyName }) => {
  if (!isOpen) return null;
  console.log(strategyName);
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="fixed inset-0 bg-black bg-opacity-50" onClick={onClose}></div>
      <div className="bg-white rounded-lg shadow-xl p-6 w-full max-w-md relative z-10">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-semibold">Add Setup to {strategyName}</h2>
          <button onClick={onClose} className="text-gray-500 hover:text-gray-700">
            <X size={20} />
          </button>
        </div>
        
        <form onSubmit={onSubmit} className='space-y-4'>
          <div>
            <label className="block font-medium mb-1" htmlFor="addStrategyName">Strategy</label>
            <input 
              className="block w-full border rounded p-2 bg-gray-100" 
              type="text" 
              id="addStrategyName" 
              name="strategyName"
              defaultValue={strategyName}
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
              defaultValue={`${strategyName}-`}
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
              placeholder="e.g. CME:ES"
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
              placeholder='e.g. 123456789'
              required
            />
          </div>
          
          <div>
            <label className="block font-medium mb-1" htmlFor="addTimeframe">Timeframe</label>
            <select 
              className="block w-full border rounded p-2" 
              id="addTimeframe" 
              name="timeframe"
            >
              <option value="30 sec">30 Seconds</option>
              <option value="1 min">1 Minute</option>
              <option value="5 min">5 Minutes</option>
              <option value="15 min">15 Minutes</option>
              <option value="30 min">30 Minutes</option>
              <option value="1 hour">1 Hour</option>
              <option value="4 hour">4 Hours</option>
              <option value="1 day">Daily</option>
            </select>
          </div>
          
          <div>
            <label className="block font-medium mb-1" htmlFor="addSchedule">Schedule</label>
            <input 
              className="block w-full border rounded p-2" 
              type="text" 
              id="addSchedule" 
              name="schedule"
              placeholder='e.g. Intraday'
              required
            />
          </div>
          
          <div>
            <label className="block font-medium mb-1" htmlFor="addOtherMarketData">Other Market Data</label>
            <input 
              className="block w-full border rounded p-2" 
              type="text" 
              id="addOtherMarketData" 
              name="market_data"
              placeholder='e.g. 1111111:CME:1 day, 2222222:CBOT:1 min'
            />
          </div>
          
          <button 
            type="submit" 
            className="w-full py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition mt-6"
          >
            Add Setup
          </button>
        </form>
      </div>
    </div>
  );
};

// Edit Setup Modal
export const EditSetupModal = ({ isOpen, onClose, onSubmit, setup, strategyName }) => {
  if (!isOpen || !setup) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="fixed inset-0 bg-black bg-opacity-50" onClick={onClose}></div>
      <div className="bg-white rounded-lg shadow-xl p-6 w-full max-w-md relative z-10">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-semibold">Edit Setup: {setup.setupName}</h2>
          <button onClick={onClose} className="text-gray-500 hover:text-gray-700">
            <X size={20} />
          </button>
        </div>
        
        <form onSubmit={onSubmit} className='space-y-4'>
          <div>
            <label className="block font-medium mb-1" htmlFor="editSetupName">Setup Name</label>
            <input 
              className="block w-full border rounded p-2 bg-gray-100" 
              type="text" 
              id="editSetupName" 
              name="setupName"
              defaultValue={setup.setupName}
              readOnly
            />
          </div>
          
          <div>
            <label className="block font-medium mb-1" htmlFor="editMarket">Market</label>
            <input 
              className="block w-full border rounded p-2" 
              type="text" 
              id="editMarket" 
              name="market"
              defaultValue={setup.market}
              required
            />
          </div>
          
          <div>
            <label className="block font-medium mb-1" htmlFor="editContractId">Contract ID</label>
            <input 
              className="block w-full border rounded p-2" 
              type="text" 
              id="editContractId" 
              name="contract_id"
              defaultValue={setup.contract_id}
            />
          </div>
          
          <div className="mt-4">
            <label className="block font-medium mb-1" htmlFor="editTimeframe">Timeframe</label>
            <select 
              className="block w-full border rounded p-2" 
              id="editTimeframe" 
              name="timeframe"
              defaultValue={setup.timeframe}
            >
              <option value="30 sec">30 Seconds</option>
              <option value="1 min">1 Minute</option>
              <option value="5 min">5 Minutes</option>
              <option value="15 min">15 Minutes</option>
              <option value="30 min">30 Minutes</option>
              <option value="1 hour">1 Hour</option>
              <option value="4 hour">4 Hours</option>
              <option value="1 day">Daily</option>
            </select>
          </div>
          
          <div className="mt-4">
            <label className="block font-medium mb-1" htmlFor="editSchedule">Schedule</label>
            <input 
              className="block w-full border rounded p-2" 
              type="text" 
              id="editSchedule" 
              name="schedule"
              defaultValue={setup.schedule || ''}
              placeholder="e.g. 0 9 * * 1-5 (9am weekdays)"
            />
          </div>
          
          <div className="mt-4">
            <label className="block font-medium mb-1" htmlFor="editOtherMarketData">Other Market Data</label>
            <input 
              className="block w-full border rounded p-2" 
              type="text" 
              id="editMarketData"                     // editOtherMarketData
              name="market_data"
              defaultValue={setup.market_data || ''} //defaultValue={selectedSetup.market_data ? selectedSetup.market_data.join(',') : ''}
            />
          </div>
          
          <button 
            type="submit" 
            className="w-full py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition mt-6"
          >
            Save Changes
          </button>
        </form>
      </div>
    </div>
  );
};

// Contract ID Tool Sidebar
export const ContractIdSidebar = ({ isOpen, onClose, onSubmit, contractResult }) => {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex">
      <div 
        className="fixed inset-0 bg-black bg-opacity-50" 
        onClick={onClose}
      ></div>
      <div className="fixed right-0 top-0 h-full w-96 bg-white shadow-lg overflow-y-auto">
        <div className="p-4 border-b">
          <div className="flex justify-between items-center">
            <h2 className="text-lg font-medium">Contract ID Lookup</h2>
            <button onClick={onClose} className="text-gray-500 hover:text-gray-700">
              <X size={20} />
            </button>
          </div>
        </div>
        
        <div className="p-4">
          <form onSubmit={onSubmit}>
            <div>
              <label className="block font-medium mb-1" htmlFor="symbol">Symbol</label>
              <input 
                className="block w-full border rounded p-2" 
                type="text" 
                id="symbol" 
                name="symbol"
                placeholder="e.g. ES"
                required
              />
            </div>
            
            <div className="mt-4">
              <label className="block font-medium mb-1" htmlFor="exchange">Exchange</label>
              <input 
                className="block w-full border rounded p-2" 
                type="text" 
                id="exchange" 
                name="exchange"
                placeholder="e.g. CME"
                required
              />
            </div>
            
            <div className="mt-4">
              <label className="block font-medium mb-1" htmlFor="contract_type">Security Type</label>
              <select 
                className="block w-full border rounded p-2" 
                id="contract_type" 
                name="contract_type"
              >
                <option value="FUT">Futures</option>
                <option value="STK">Stock</option>
                <option value="OPT">Option</option>
                <option value="CASH">Forex</option>
              </select>
            </div>
            
            <div className="mt-4">
              <label className="block font-medium mb-1" htmlFor="expiry">Expiry (for futures)</label>
              <input 
                className="block w-full border rounded p-2" 
                type="text" 
                id="expiry" 
                name="expiry"
                placeholder="e.g. 202312"
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
            
            <button 
              type="submit" 
              className="w-full py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition mt-4"
            >
              Lookup Contract
            </button>
          </form>
          
          {contractResult && (
            <div className="mt-6 p-4 bg-gray-50 rounded-md">
              <h3 className="font-medium mb-2">Result:</h3>
              {contractResult.error ? (
                <div className="text-red-600">{contractResult.error}</div>
              ) : (
                <pre className="whitespace-pre-wrap text-sm bg-gray-100 p-2 rounded">
                  {JSON.stringify(contractResult, null, 2)}
                </pre>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
