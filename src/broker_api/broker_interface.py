from abc import ABC, abstractmethod
from typing import List, Optional, Dict, Any
from datetime import datetime
from fastapi import HTTPException
import ib_async
from ib_async import util
from models import *
from dotenv import load_dotenv
import os
from pathlib import Path
import random
from datetime import datetime, timedelta
import numpy as np
import re
import nest_asyncio
import schwab
from schwab.auth import easy_client
from schwab.orders.equities import equity_buy_market, equity_sell_market, equity_buy_limit, equity_sell_limit
from schwab_orders_futures import future_buy_market, future_sell_market, future_buy_limit, future_sell_limit
import json
from pathlib import Path
nest_asyncio.apply()

# # Load environment variables
# env_path = Path('.env')
# if not env_path.exists():
#     raise FileNotFoundError(f"Environment file not found at {env_path}")
# load_dotenv(env_path)

# Abstract Broker Interface
class BrokerInterface(ABC):
    @abstractmethod
    async def connect(self) -> bool:
        pass

    @abstractmethod
    async def disconnect(self) -> bool:
        pass

    @abstractmethod
    async def get_quote(self, contract: Contract) -> Quote:
        pass

    @abstractmethod
    async def get_fills(self, order_id: Optional[str] = None) -> List[Fill]:
        pass

    @abstractmethod
    async def place_order(self, order_request: Order) -> str:
        pass

    @abstractmethod
    async def get_historical_data(
        self,
        contract: Contract,
        start_time: datetime,
        end_time: datetime,
        bar_size: str,
        rth: bool=True,
    ) -> List[Dict[str, Any]]:
        pass

    @abstractmethod
    async def validate_contract(self, contract: Contract) -> bool:
        pass

    @abstractmethod
    async def get_account_summary(self) -> Dict[str, float]:
        pass

# Interactive Brokers Implementation
class IBKRBroker(BrokerInterface):
    def __init__(self):
        self.host = os.getenv('IB_HOST', '127.0.0.1')
        self.port = int(os.getenv('IB_PORT', '7496'))
        self.client_id = 226
        self.ib = ib_async.IB()
        self._connected = False
        self.pending_trades = {}  # Initialize the pending trades dictionary

    async def connect(self) -> bool:
        # if not self._connected:
        if not self.ib.isConnected():
            try:
                await self.ib.connectAsync(self.host, self.port, self.client_id)
                self._connected = True
            except Exception as e:
                raise HTTPException(status_code=500, detail=f"Connection failed: {str(e)}")
        return self._connected

    async def disconnect(self) -> bool:
        if self._connected:
            self.ib.disconnect()
            self._connected = False
        return True

    def _convert_contract(self, contract: Optional[Contract]=None, contract_id:Optional[int]=None, exchange:Optional[str]=None) -> ib_async.Contract:
        if contract is None or contract.contract_type=="":
            try:
                return ib_async.Contract(conId=contract_id, exchange=exchange)
            except Exception:
                raise ValueError(f"You did not pass the correct parameters: \n\t{contract}\n\t{contract_id}\n\t{exchange}")

        if contract.contract_type == ContractType.STOCK:
            return ib_async.Stock(contract.symbol, contract.exchange or "SMART", contract.currency)
        elif contract.contract_type == ContractType.FUTURE:
            return ib_async.Future(contract.symbol, contract.expiry, contract.exchange)
        elif contract.contract_type == ContractType.ETF:
            return ib_async.Stock(contract.symbol, contract.exchange or "SMART", contract.currency)
        raise ValueError(f"Unsupported contract type: {contract.contract_type}")

    async def get_quote(self, contract: Contract) -> Quote:
        ib_contract = self._convert_contract(contract)
        await self.connect()
        try:
            await self.ib.qualifyContractsAsync(ib_contract)
            tickers = self.ib.reqMktData(ib_contract, snapshot=True)
            self.ib.sleep(1)
            return Quote(
                # contract=contract,
                symbol=contract.symbol,
                bid=tickers.bid,
                ask=tickers.ask,
                last=tickers.last,
                timestamp=datetime.now()
            )
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Failed to get quote: {str(e)}")

    async def get_quote_by_contract_id(self, exchange:str, contract_id:int) -> Quote:
        await self.connect()
        ib_contract = ib_async.Contract(conId=contract_id, exchange=exchange)

        try:
            await self.ib.qualifyContractsAsync(ib_contract)
            tickers = self.ib.reqMktData(ib_contract, snapshot=True)
            self.ib.sleep(1)
            quote = Quote(
                symbol=ib_contract.symbol,
                bid=tickers.bid,
                ask=tickers.ask,
                last=0. if tickers.last== 'nan' else tickers.last,
                timestamp=datetime.now()
            )
            print("Quote: ", quote )

            return quote
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Failed to get quote: {str(e)}")

    async def get_fills(self) -> List[Fill]:
        await self.connect()
        fills = []
        try:
            for trade_id, trade in self.pending_trades.items():  # Iterate through the pending trades dictionary
                if trade.orderStatus.status == 'Filled':  # Check if the trade is filled
                    print("Fills:", trade.fills[0])  # Print the fill details
                    fill = trade.fills[0]  # Get the fill details
                    fills.append(Fill(
                        order_id = fill.execution.orderId, # trade_id
                        contract_id=fill.contract.conId,
                        quantity = fill.execution.cumQty,
                        price = fill.execution.avgPrice,
                        time = fill.time,
                        side="BUY" if fill.execution.side=="BOT" else "SELL"
                    ))  # Append the fill details to the fills list
                    self.pending_trades.pop(trade_id)  # Remove the trade from the pending trades dictionary
            return fills
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Failed to get fills: {str(e)}")

    async def get_trades(self) -> List[Trade]:
        await self.connect()
        trades=[]
        try:
            ib_trades = self.ib.trades()
            for trade in ib_trades:
                quantity=0
                price=0.
                if trade.orderStatus.status == "Filled":
                    quantity=trade.fills[0].execution.shares
                    price=trade.fills[0].execution.price
                    order_id = trade.fills[0].execution.orderId

                t = Trade(
                    order_id=order_id,
                    contract_id=trade.contract.conId,
                    time=datetime.now().strftime("%Y-%m-%dT%H:%M:%SZ"),
                    quantity=quantity,
                    price=price,
                    side=OrderSide.BUY if trade.order.action == "BUY" else OrderSide.SELL,
                    order_status=trade.orderStatus.status)
                trades.append(t)
            return trades
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Failed to get : {str(e)}")

    async def place_order(self, order: Order) -> str:
        ib_contract = self._convert_contract(contract_id=order.trade.contract_id,
                                             exchange=order.trade.exchange)
        await self.connect()
        try:
            await self.ib.qualifyContractsAsync(ib_contract)

            # Determine order type
            order_type = order.trade.order_type #, 'order_type', 'LMT')  # Default to LMT if not specified

            if order_type == 'MKT':
                ib_order = ib_async.MarketOrder(
                    action="BUY" if order.trade.side == OrderSide.BUY else "SELL",
                    totalQuantity=order.trade.quantity
                )
            else:  # Default to LimitOrder
                ib_order = ib_async.LimitOrder(
                    action="BUY" if order.trade.side == OrderSide.BUY else "SELL",
                    totalQuantity=order.trade.quantity,
                    lmtPrice=order.price
                )

            trade = self.ib.placeOrder(ib_contract, ib_order)
            trade_id = str(trade.order.orderId)
            self.pending_trades[trade_id] = trade  # Add the trade to the pending trade dictionary
            return trade_id  # Return the trade ID to the caller

        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Failed to place order: {str(e)}")

    async def get_historical_data(self, contract: Contract, start_time: datetime, end_time: datetime, bar_size: str, rth:bool=True) -> List[Dict[str, Any]]:
        await self.connect()
        if contract.contract_type is None:
            ib_contract = self._convert_contract(contract_id=contract.contract_id, exchange=contract.exchange)
        else:
            ib_contract = self._convert_contract(contract=contract)

        try:
            await self.ib.qualifyContractsAsync(ib_contract)
            bars = await self.ib.reqHistoricalDataAsync(
                ib_contract,
                endDateTime=end_time,
                durationStr=self._calculate_duration(start_time, end_time),
                barSizeSetting=bar_size,
                whatToShow='TRADES',
                useRTH=rth
            )

            return [
                {
                    "timestamp": bar.date,
                    "open": bar.open,
                    "high": bar.high,
                    "low": bar.low,
                    "close": bar.close,
                    "volume": bar.volume
                }
                for bar in bars
            ]
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Failed to get historical data: {str(e)}")

    async def get_historical_data_by_contract_id(self, contract_id: int, exchange:str, start_time: datetime, end_time: datetime, bar_size: str, rth:bool=True) -> List[Dict[str, Any]]:
        ib_contract = self._convert_contract(contract_id=contract_id, exchange=exchange)
        await self.connect()
        try:
            await self.ib.qualifyContractsAsync(ib_contract)
            bars = await self.ib.reqHistoricalDataAsync(
                ib_contract,
                endDateTime=end_time,
                durationStr=self._calculate_duration(start_time, end_time),
                barSizeSetting=bar_size,
                whatToShow='TRADES',
                useRTH=rth 
            )
            return util.df(bars)
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Failed to get historical data: {str(e)}")

    async def validate_contract(self, contract: Contract) -> bool:
        ib_contract = self._convert_contract(contract)
        await self.connect()

        try:
            contracts = await self.ib.qualifyContractsAsync(ib_contract)
            return len(contracts) > 0
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"The service is unable to validate the contract: {str(e)}")

    async def get_contract_id(self, contract: Contract) -> int:
        ib_contract = self._convert_contract(contract)
        await self.connect()
        try:
            contracts = await self.ib.qualifyContractsAsync(ib_contract)
            print(contracts)
            if len(contracts)==1:
                return contracts[0].conId
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"There was an error retrieving the contract ID: {str(e)}")

    def _calculate_duration(self, start_time: datetime, end_time: datetime) -> str:
        # Calculate the duration string based on the time difference
        diff = end_time - start_time
        days = diff.days

        if days <= 1:
            return "1 D"
        elif days <= 7:
            return "1 W"
        elif days <= 31:
            return "1 M"
        elif days <= 365:
            return "1 Y"
        else:
            return "5 Y"

    async def get_current_minute_bar_open(self, contract_id:int, exchange:str) -> float:
        await self.connect()
        ib_contract = self._convert_contract(contract_id=contract_id, exchange=exchange)

        try:
            await self.ib.qualifyContractsAsync(ib_contract)
            bars = self.ib.reqHistoricalData(
                ib_contract,
                endDateTime='',
                durationStr='120 S',
                barSizeSetting='1 min',
                whatToShow='TRADES',
                useRTH=False,
                formatDate=1)
            return util.df(bars)['open'].iloc[-1]
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Failed to get current bar open: {str(e)}")

    async def close_all_positions(self) -> int:
        await self.connect()
        print(self.ib.positions())
        try:
            for open_trade in self.ib.positions():
                if open_trade is None:
                    return 0
                direction = 'BUY' if open_trade.position < 0 else 'SELL'
                order = ib_async.MarketOrder(direction, abs(open_trade.position))
                await self.ib.qualifyContractsAsync(open_trade.contract)
                self.ib.placeOrder(open_trade.contract, order)
        except Exception as e:
            print(e)
            return 0
        return 1

    async def get_positions(self) -> int:
        await self.connect()
        try:
            return self.ib.positions()
        except Exception as e:
            print(e)
            return 0
        return 1

    async def get_account_summary(self)->Dict[str,float]:
        await self.connect()
        try:
            acct_summary=self.ib.accountValues()
            scope=['NetLiquidation','CashBalance','TotalCashBalance','BuyingPower','AvailableFunds',
                   'FullMaintMarginReq','FullInitMarginReq','InitMarginReq',
                   'GrossPositionValue','FuturesPNL','UnrealizedPnL','RealizedPnL']
            # print(acct_summary)
            return {r.tag:float(r.value) for r in acct_summary if (r.tag in scope)&(r.currency=='USD')} 
        except Exception as e:
            print("Error: ", e)
            return {}

# class SchwabBroker(BrokerInterface):
#     def __init__(self):
#         # Schwab API credentials from environment variables
#         self.app_key = os.getenv('SCHWAB_APP_KEY')
#         self.app_secret = os.getenv('SCHWAB_APP_SECRET')
#         self.token_path = os.getenv('SCHWAB_TOKEN_PATH', './schwab_token.json')
#         self.redirect_uri = os.getenv('SCHWAB_REDIRECT_URI', 'https://localhost:8080')
        
#         self.client = None
#         self._connected = False
#         self.pending_orders = {}  # Track pending orders

#     async def connect(self) -> bool:
#         """Connect to Schwab API using token-based authentication"""
#         if not self._connected:
#             try:
#                 # Check if we have required credentials
#                 if not self.app_key or not self.app_secret:
#                     raise ValueError("SCHWAB_APP_KEY and SCHWAB_APP_SECRET must be set in environment variables")
                
#                 # Try to create client with existing token or create new one
#                 token_file = Path(self.token_path)
                
#                 if token_file.exists():
#                     # Load existing token
#                     self.client = schwab.auth.client_from_token_file(
#                         token_path=self.token_path,
#                         api_key=self.app_key,
#                         app_secret=self.app_secret
#                     )
#                 else:
#                     # Create new client - this will require manual authentication
#                     self.client = easy_client(
#                         api_key=self.app_key,
#                         app_secret=self.app_secret,
#                         redirect_uri=self.redirect_uri,
#                         token_path=self.token_path
#                     )
                
#                 # Test connection by getting account info
#                 response = self.client.get_account_numbers()
#                 if response.status_code == 200:
#                     self._connected = True
#                     self.account_numbers = response.json()
#                     # Use the first account number for operations
#                     if self.account_numbers:
#                         self.primary_account = list(self.account_numbers.keys())[0]
#                 else:
#                     raise Exception(f"Failed to authenticate: {response.status_code}")
                    
#             except Exception as e:
#                 raise HTTPException(status_code=500, detail=f"Schwab connection failed: {str(e)}")
        
#         return self._connected

#     async def disconnect(self) -> bool:
#         """Disconnect from Schwab API"""
#         if self._connected:
#             self.client = None
#             self._connected = False
#         return True

#     def _convert_contract_to_schwab_instrument(self, contract: Contract) -> Dict[str, Any]:
#         """Convert our Contract model to Schwab instrument format"""
#         if contract.contract_type == ContractType.STOCK:
#             return {
#                 "symbol": contract.symbol,
#                 "assetType": "EQUITY"
#             }
#         elif contract.contract_type == ContractType.ETF:
#             return {
#                 "symbol": contract.symbol,
#                 "assetType": "EQUITY"
#             }
#         elif contract.contract_type == ContractType.FUTURE:
#             return {
#                 "symbol": contract.symbol,
#                 "assetType": "FUTURE"
#             }
#         else:
#             raise ValueError(f"Unsupported contract type: {contract.contract_type}")

#     async def get_quote(self, contract: Contract) -> Quote:
#         """Get real-time quote for a contract"""
#         await self.connect()
        
#         try:
#             # Get quote from Schwab API
#             response = self.client.get_quote(contract.symbol)
            
#             if response.status_code != 200:
#                 raise Exception(f"Quote request failed: {response.status_code}")
            
#             quote_data = response.json()
            
#             # Extract quote information (structure may vary by asset type)
#             if contract.symbol in quote_data:
#                 quote_info = quote_data[contract.symbol]
                
#                 # Handle different quote structures for different asset types
#                 if quote_info.get('assetType') == 'EQUITY':
#                     return Quote(
#                         symbol=contract.symbol,
#                         bid=quote_info.get('bidPrice', 0.0),
#                         ask=quote_info.get('askPrice', 0.0),
#                         last=quote_info.get('lastPrice', 0.0),
#                         timestamp=datetime.now()
#                     )
#                 else:
#                     # Handle other asset types
#                     return Quote(
#                         symbol=contract.symbol,
#                         bid=quote_info.get('bid', 0.0),
#                         ask=quote_info.get('ask', 0.0),
#                         last=quote_info.get('last', 0.0),
#                         timestamp=datetime.now()
#                     )
#             else:
#                 raise Exception(f"No quote data found for symbol {contract.symbol}")
                
#         except Exception as e:
#             raise HTTPException(status_code=500, detail=f"Failed to get quote: {str(e)}")

#     async def get_quote_by_contract_id(self, exchange: str, contract_id: int) -> Quote:
#         """Get quote by contract ID - Schwab uses symbols, so we'll need to map this"""
#         await self.connect()
        
#         try:
#             # In Schwab's case, we would need to maintain a mapping of contract_id to symbol
#             # For now, we'll use the contract_id as a symbol placeholder
#             symbol = f"SYMBOL_{contract_id}"  # This would need proper mapping in production
            
#             response = self.client.get_quote(symbol)
            
#             if response.status_code != 200:
#                 raise Exception(f"Quote request failed: {response.status_code}")
            
#             quote_data = response.json()
            
#             if symbol in quote_data:
#                 quote_info = quote_data[symbol]
#                 return Quote(
#                     symbol=symbol,
#                     bid=quote_info.get('bidPrice', 0.0),
#                     ask=quote_info.get('askPrice', 0.0),
#                     last=quote_info.get('lastPrice', 0.0),
#                     timestamp=datetime.now()
#                 )
#             else:
#                 raise Exception(f"No quote data found for contract_id {contract_id}")
                
#         except Exception as e:
#             raise HTTPException(status_code=500, detail=f"Failed to get quote: {str(e)}")

#     async def get_fills(self) -> List[Fill]:
#         """Get recent fills/executions"""
#         await self.connect()
        
#         try:
#             # Get orders from the last 7 days
#             from_date = datetime.now() - timedelta(days=7)
            
#             response = self.client.get_orders_for_account(
#                 account_hash=self.primary_account,
#                 from_entered_time=from_date,
#                 to_entered_time=datetime.now(),
#                 status='FILLED'
#             )
            
#             if response.status_code != 200:
#                 raise Exception(f"Orders request failed: {response.status_code}")
            
#             orders_data = response.json()
#             fills = []
            
#             for order in orders_data:
#                 if order.get('status') == 'FILLED':
#                     # Extract fill information
#                     order_id = order.get('orderId')
                    
#                     # Get contract information
#                     instrument = order.get('orderLegCollection', [{}])[0].get('instrument', {})
#                     symbol = instrument.get('symbol', '')
                    
#                     # Get execution details
#                     executions = order.get('orderActivityCollection', [])
                    
#                     for execution in executions:
#                         if execution.get('activityType') == 'EXECUTION':
#                             exec_legs = execution.get('executionLegs', [])
#                             for leg in exec_legs:
#                                 fill = Fill(
#                                     order_id=order_id,
#                                     contract_id=hash(symbol),  # Generate contract_id from symbol
#                                     quantity=leg.get('quantity', 0),
#                                     price=leg.get('price', 0.0),
#                                     time=datetime.fromisoformat(leg.get('time', datetime.now().isoformat())),
#                                     side=OrderSide.BUY if order.get('orderLegCollection', [{}])[0].get('instruction') == 'BUY' else OrderSide.SELL
#                                 )
#                                 fills.append(fill)
            
#             return fills
            
#         except Exception as e:
#             raise HTTPException(status_code=500, detail=f"Failed to get fills: {str(e)}")

#     async def place_order(self, order: Order) -> str:
#         """Place an order with Schwab"""
#         await self.connect()
        
#         try:
#             # Create the appropriate order based on type and side
#             if order.trade.order_type == OrderType.MARKET:
#                 if order.trade.side == OrderSide.BUY:
#                     schwab_order = equity_buy_market(order.trade.symbol, order.trade.quantity)
#                 else:
#                     schwab_order = equity_sell_market(order.trade.symbol, order.trade.quantity)
#             elif order.trade.order_type == OrderType.LIMIT:
#                 if order.trade.side == OrderSide.BUY:
#                     schwab_order = equity_buy_limit(order.trade.symbol, order.trade.quantity, order.price)
#                 else:
#                     schwab_order = equity_sell_limit(order.trade.symbol, order.trade.quantity, order.price)
#             else:
#                 raise ValueError(f"Unsupported order type: {order.trade.order_type}")
            
#             # Place the order
#             response = self.client.place_order(
#                 account_hash=self.primary_account,
#                 order_spec=schwab_order
#             )
            
#             if response.status_code not in [200, 201]:
#                 raise Exception(f"Order placement failed: {response.status_code} - {response.text}")
            
#             # Extract order ID from response headers or body
#             order_id = response.headers.get('Location', '').split('/')[-1]
#             if not order_id:
#                 # If not in headers, try to get from response body
#                 response_data = response.json() if response.text else {}
#                 order_id = response_data.get('orderId', f"SCHWAB_{int(datetime.now().timestamp())}")
            
#             # Store in pending orders for tracking
#             self.pending_orders[order_id] = {
#                 'order': order,
#                 'timestamp': datetime.now(),
#                 'status': 'SUBMITTED'
#             }
            
#             return str(order_id)
            
#         except Exception as e:
#             raise HTTPException(status_code=500, detail=f"Failed to place order: {str(e)}")

#     async def get_historical_data(self, contract: Contract, start_time: datetime, end_time: datetime, bar_size: str) -> List[Dict[str, Any]]:
#         """Get historical price data"""
#         await self.connect()
        
#         try:
#             # Map bar_size to Schwab's frequency format
#             frequency_map = {
#                 "1 min": schwab.client.Client.PriceHistory.Frequency.MINUTE,
#                 "5 mins": schwab.client.Client.PriceHistory.Frequency.MINUTE,
#                 "1 hour": schwab.client.Client.PriceHistory.Frequency.MINUTE,
#                 "1 day": schwab.client.Client.PriceHistory.Frequency.DAILY
#             }
            
#             frequency_type_map = {
#                 "1 min": schwab.client.Client.PriceHistory.FrequencyType.MINUTE,
#                 "5 mins": schwab.client.Client.PriceHistory.FrequencyType.MINUTE,
#                 "1 hour": schwab.client.Client.PriceHistory.FrequencyType.MINUTE,
#                 "1 day": schwab.client.Client.PriceHistory.FrequencyType.DAILY
#             }
            
#             frequency = frequency_map.get(bar_size, schwab.client.Client.PriceHistory.Frequency.MINUTE)
#             frequency_type = frequency_type_map.get(bar_size, schwab.client.Client.PriceHistory.FrequencyType.MINUTE)
            
#             # For minute data, we need to specify the frequency value
#             if bar_size == "5 mins":
#                 frequency = 5
#             elif bar_size == "1 hour":
#                 frequency = 60
#             else:
#                 frequency = 1
            
#             response = self.client.get_price_history(
#                 symbol=contract.symbol,
#                 period_type=schwab.client.Client.PriceHistory.PeriodType.DAY,
#                 frequency_type=frequency_type,
#                 frequency=frequency,
#                 start_datetime=start_time,
#                 end_datetime=end_time
#             )
            
#             if response.status_code != 200:
#                 raise Exception(f"Historical data request failed: {response.status_code}")
            
#             data = response.json()
#             candles = data.get('candles', [])
            
#             historical_data = []
#             for candle in candles:
#                 historical_data.append({
#                     "timestamp": datetime.fromtimestamp(candle['datetime'] / 1000),
#                     "open": candle['open'],
#                     "high": candle['high'],
#                     "low": candle['low'],
#                     "close": candle['close'],
#                     "volume": candle['volume']
#                 })
            
#             return historical_data
            
#         except Exception as e:
#             raise HTTPException(status_code=500, detail=f"Failed to get historical data: {str(e)}")

#     async def validate_contract(self, contract: Contract) -> bool:
#         """Validate if a contract exists and is tradeable"""
#         await self.connect()
        
#         try:
#             # Try to get a quote for the symbol
#             response = self.client.get_quote(contract.symbol)
#             return response.status_code == 200
            
#         except Exception:
#             return False

#     async def get_contract_id(self, contract: Contract) -> int:
#         """Get contract ID for a contract - Schwab uses symbols, so we'll generate a hash"""
#         await self.connect()
        
#         try:
#             # Since Schwab uses symbols rather than contract IDs, we'll generate a consistent hash
#             contract_string = f"{contract.symbol}_{contract.contract_type}_{contract.exchange}_{contract.currency}"
#             return abs(hash(contract_string)) % (10**10)  # Return a 10-digit integer
            
#         except Exception as e:
#             raise HTTPException(status_code=500, detail=f"Failed to get contract ID: {str(e)}")

#     async def get_current_minute_bar_open(self, contract_id: int, exchange: str) -> float:
#         """Get the current minute bar's opening price"""
#         await self.connect()
        
#         try:
#             # We would need to map contract_id back to symbol
#             # For now, using a placeholder approach
#             symbol = f"SYMBOL_{contract_id}"  # This would need proper mapping
            
#             # Get recent minute bars
#             end_time = datetime.now()
#             start_time = end_time - timedelta(minutes=5)
            
#             response = self.client.get_price_history(
#                 symbol=symbol,
#                 period_type=schwab.client.Client.PriceHistory.PeriodType.DAY,
#                 frequency_type=schwab.client.Client.PriceHistory.FrequencyType.MINUTE,
#                 frequency=1,
#                 start_datetime=start_time,
#                 end_datetime=end_time
#             )
            
#             if response.status_code != 200:
#                 raise Exception(f"Price history request failed: {response.status_code}")
            
#             data = response.json()
#             candles = data.get('candles', [])
            
#             if candles:
#                 # Return the open price of the most recent candle
#                 return candles[-1]['open']
#             else:
#                 raise Exception("No recent price data available")
                
#         except Exception as e:
#             raise HTTPException(status_code=500, detail=f"Failed to get current minute bar open: {str(e)}")

#     async def close_all_positions(self) -> int:
#         """Close all open positions"""
#         await self.connect()
        
#         try:
#             # Get all positions
#             response = self.client.get_account(
#                 account_hash=self.primary_account,
#                 fields='positions'
#             )
            
#             if response.status_code != 200:
#                 raise Exception(f"Failed to get positions: {response.status_code}")
            
#             account_data = response.json()
#             positions = account_data.get('securitiesAccount', {}).get('positions', [])
            
#             orders_placed = 0
            
#             for position in positions:
#                 quantity = position.get('longQuantity', 0) - position.get('shortQuantity', 0)
                
#                 if quantity != 0:
#                     instrument = position.get('instrument', {})
#                     symbol = instrument.get('symbol', '')
                    
#                     # Create market order to close position
#                     if quantity > 0:
#                         # Long position - sell to close
#                         order = equity_sell_market(symbol, abs(quantity))
#                     else:
#                         # Short position - buy to close
#                         order = equity_buy_market(symbol, abs(quantity))
                    
#                     # Place the closing order
#                     close_response = self.client.place_order(
#                         account_hash=self.primary_account,
#                         order_spec=order
#                     )
                    
#                     if close_response.status_code in [200, 201]:
#                         orders_placed += 1
            
#             return orders_placed
            
#         except Exception as e:
#             print(f"Error closing positions: {e}")
#             return 0

#     async def get_positions(self) -> List[dict]:
#         """Get all current positions"""
#         await self.connect()
        
#         try:
#             response = self.client.get_account(
#                 account_hash=self.primary_account,
#                 fields='positions'
#             )
            
#             if response.status_code != 200:
#                 raise Exception(f"Failed to get positions: {response.status_code}")
            
#             account_data = response.json()
#             positions_data = account_data.get('securitiesAccount', {}).get('positions', [])
            
#             positions = []
#             for pos in positions_data:
#                 quantity = pos.get('longQuantity', 0) - pos.get('shortQuantity', 0)
                
#                 if quantity != 0:  # Only include non-zero positions
#                     instrument = pos.get('instrument', {})
#                     positions.append({
#                         'symbol': instrument.get('symbol', ''),
#                         'position': quantity,
#                         'avgCost': pos.get('averagePrice', 0.0),
#                         'marketValue': pos.get('marketValue', 0.0),
#                         'contract': {
#                             'symbol': instrument.get('symbol', ''),
#                             'conId': abs(hash(instrument.get('symbol', ''))) % (10**10)
#                         }
#                     })
            
#             return positions
            
#         except Exception as e:
#             print(f"Error getting positions: {e}")
#             return []

#     async def get_trades(self) -> List[Trade]:
#         """Get all recent trades"""
#         await self.connect()
        
#         try:
#             # Get orders from the last 30 days
#             from_date = datetime.now() - timedelta(days=30)
            
#             response = self.client.get_orders_for_account(
#                 account_hash=self.primary_account,
#                 from_entered_time=from_date,
#                 to_entered_time=datetime.now()
#             )
            
#             if response.status_code != 200:
#                 raise Exception(f"Orders request failed: {response.status_code}")
            
#             orders_data = response.json()
#             trades = []
            
#             for order in orders_data:
#                 order_id = order.get('orderId')
#                 status = order.get('status', 'UNKNOWN')
                
#                 # Get instrument details
#                 legs = order.get('orderLegCollection', [])
#                 if legs:
#                     leg = legs[0]
#                     instrument = leg.get('instrument', {})
#                     symbol = instrument.get('symbol', '')
                    
#                     # Determine quantities and prices
#                     quantity = 0
#                     price = 0.0
                    
#                     if status == 'FILLED':
#                         # Get execution details if available
#                         executions = order.get('orderActivityCollection', [])
#                         for execution in executions:
#                             if execution.get('activityType') == 'EXECUTION':
#                                 exec_legs = execution.get('executionLegs', [])
#                                 if exec_legs:
#                                     quantity = exec_legs[0].get('quantity', 0)
#                                     price = exec_legs[0].get('price', 0.0)
#                                     break
                    
#                     # Create trade object
#                     trade = Trade(
#                         order_id=order_id,
#                         contract_id=abs(hash(symbol)) % (10**10),
#                         time=datetime.now(),  # You might want to parse the actual time from order data
#                         quantity=quantity,
#                         price=price,
#                         side=OrderSide.BUY if leg.get('instruction') == 'BUY' else OrderSide.SELL,
#                         order_status=OrderStatus.Filled if status == 'FILLED' else OrderStatus.Submitted
#                     )
                    
#                     trades.append(trade)
            
#             return trades
            
#         except Exception as e:
#             raise HTTPException(status_code=500, detail=f"Failed to get trades: {str(e)}")

#     async def get_account_summary(self) -> Dict[str, float]:
#         """Get account summary information"""
#         await self.connect()
        
#         try:
#             response = self.client.get_account(
#                 account_hash=self.primary_account,
#                 fields='balances'
#             )
            
#             if response.status_code != 200:
#                 raise Exception(f"Account request failed: {response.status_code}")
            
#             account_data = response.json()
#             balances = account_data.get('securitiesAccount', {}).get('currentBalances', {})
            
#             # Map Schwab balance fields to our standard format
#             summary = {
#                 'NetLiquidation': balances.get('liquidationValue', 0.0),
#                 'CashBalance': balances.get('cashBalance', 0.0),
#                 'TotalCashBalance': balances.get('totalCash', 0.0),
#                 'BuyingPower': balances.get('buyingPower', 0.0),
#                 'AvailableFunds': balances.get('availableFunds', 0.0),
#                 'GrossPositionValue': balances.get('longMarketValue', 0.0),
#                 'UnrealizedPnL': balances.get('unrealizedPL', 0.0),
#                 'RealizedPnL': 0.0  # This might need to be calculated from trades
#             }
            
#             return summary
            
#         except Exception as e:
#             print(f"Error getting account summary: {e}")
#             return {}

class TestBroker(BrokerInterface):
    def __init__(self):
        self._connected = False
        self._orders = {}  # Store orders by order_id
        self._fills = {}   # Store fills by order_id
        self._positions = {}  # Store positions by symbol
        self._prices = {}  # Store simulated prices by symbol
        self.pending_trades = {}  # Initialize the pending trades dictionary similar to IBKRBroker

    async def connect(self) -> bool:
        if not self._connected:
            self._connected = True
        return self._connected

    async def disconnect(self) -> bool:
        if self._connected:
            self._connected = False
        return self._connected

    async def _generate_order_id(self) -> str:
        await self.connect()
        if not self._connected:
            raise HTTPException(status_code=500, detail="Not connected")
        return f"TEST_{int(datetime.now().timestamp())}_{random.randint(1000, 9999)}"

    async def _simulate_price(self, symbol: str) -> dict:
        """Generate simulated prices for a symbol."""
        await self.connect()
        if not self._connected:
            raise HTTPException(status_code=500, detail="Not connected")
        if symbol not in self._prices:
            base_price = random.uniform(10, 1000)
            spread = base_price * 0.001  # 0.1% spread
            self._prices[symbol] = {
                "base": base_price,
                "last_update": datetime.now()
            }

        # Update price with random walk
        elapsed = (datetime.now() - self._prices[symbol]["last_update"]).total_seconds()
        if elapsed > 1:  # Update price if more than 1 second has passed
            price_change = random.gauss(0, self._prices[symbol]["base"] * 0.001)
            self._prices[symbol]["base"] += price_change
            self._prices[symbol]["last_update"] = datetime.now()

        base = self._prices[symbol]["base"]
        spread = base * 0.001

        return {
            "bid": base - spread/2,
            "ask": base + spread/2,
            "last": base + random.uniform(-spread/2, spread/2)
        }

    async def get_quote(self, contract: Contract) -> Quote:
        await self.connect()
        if not self._connected:
            raise HTTPException(status_code=500, detail="Not connected")

        prices = self._simulate_price(contract.symbol)

        return Quote(
            contract=contract.symbol,
            bid=prices["bid"],
            ask=prices["ask"],
            last=prices["last"],
            timestamp=datetime.now()
        )

    async def get_fills(self) -> List[Fill]:
        await self.connect()
        if not self._connected:
            raise HTTPException(status_code=500, detail="Not connected")

        fills = []
        try:
            for trade_id, trade in self.pending_trades.items():
                if trade["orderStatus"]["status"] == 'Filled':
                    # Get the fill from the _fills dictionary
                    if trade_id in self._fills:
                        fills.append(self._fills[trade_id])
                        # Remove the trade from pending_trades after retrieving the fill
                        self.pending_trades.pop(trade_id)
            return fills
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Failed to get fills: {str(e)}")

    async def place_order(self, order: Order) -> str:
        await self.connect()
        if not self._connected:
            raise HTTPException(status_code=500, detail="Not connected")

        order_id = await self._generate_order_id()
        self._orders[order_id] = order

        # Create a simulated trade object to mimic IBKRBroker's behavior
        trade = {
            "order": {
                "orderId": order_id,
                "action": "BUY" if order.trade.side == OrderSide.BUY else "SELL",
                "totalQuantity": order.trade.quantity,
                "lmtPrice": order.price if order.trade.order_type == OrderType.LIMIT else 0.0
            },
            "orderStatus": {
                "status": "Submitted"
            },
            "contract": {
                "symbol": order.trade.symbol,
                "conId": order.trade.contract_id,
                "exchange": order.trade.exchange
            }
        }

        # Store the trade in pending_trades
        self.pending_trades[order_id] = trade

        # Simulate immediate fill for market orders
        if order.trade.order_type == OrderType.MARKET:
            # Create a simple contract for this symbol
            contract = Contract(symbol=order.trade.symbol, contract_type=ContractType.STOCK)

            prices = await self._simulate_price(order.trade.symbol)
            fill_price = prices["ask"] if order.trade.side == OrderSide.BUY else prices["bid"]

            fill = Fill(
                order_id=int(order_id.split('_')[1]),  # Extract numeric part of order_id
                contract_id=order.trade.contract_id,
                time=datetime.now(),
                quantity=order.trade.quantity,
                price=fill_price,
                side=order.trade.side
            )

            self._fills[order_id] = fill

            # Update position
            position_delta = order.trade.quantity if order.trade.side == OrderSide.BUY else -order.trade.quantity
            self._positions[order.trade.symbol] = self._positions.get(order.trade.symbol, 0) + position_delta

            # Update the trade status to Filled
            self.pending_trades[order_id]["orderStatus"]["status"] = "Filled"

        # For limit orders, simulate fill only if price is favorable
        elif order.trade.order_type == OrderType.LIMIT:
            prices = await self._simulate_price(order.trade.symbol)
            if ((order.trade.side == OrderSide.BUY and prices["ask"] <= order.price) or
                (order.trade.side == OrderSide.SELL and prices["bid"] >= order.price)):

                fill = Fill(
                    order_id=int(order_id.split('_')[1]),  # Extract numeric part of order_id
                    contract_id=order.trade.contract_id,
                    time=datetime.now(),
                    quantity=order.trade.quantity,
                    price=order.price,
                    side=order.trade.side
                )
                self._fills[order_id] = fill

                position_delta = order.trade.quantity if order.trade.side == OrderSide.BUY else -order.trade.quantity
                self._positions[order.trade.symbol] = self._positions.get(order.trade.symbol, 0) + position_delta

                # Update the trade status to Filled
                self.pending_trades[order_id]["orderStatus"]["status"] = "Filled"

        return order_id

    async def get_historical_data(self, contract: Contract, start_time: datetime, end_time: datetime, bar_size: str) -> List[Dict[str, Any]]:
        await self.connect()
        if not self._connected:
            raise HTTPException(status_code=500, detail="Not connected")

        # Parse bar size to determine number of bars
        bar_seconds = {
            "1 min": 60,
            "5 mins": 300,
            "1 hour": 3600,
            "1 day": 86400
        }.get(bar_size, 300)  # Default to 5 mins

        total_seconds = (end_time - start_time).total_seconds()
        num_bars = int(total_seconds / bar_seconds)

        # Generate simulated price data using random walk
        base_price = random.uniform(10, 1000)
        volatility = base_price * 0.02  # 2% daily volatility

        prices = []
        current_price = base_price

        for i in range(num_bars):
            bar_time = start_time + timedelta(seconds=i * bar_seconds)
            price_change = random.gauss(0, volatility * np.sqrt(bar_seconds / 86400))
            current_price = max(0.01, current_price + price_change)

            # Generate OHLC data
            high = current_price * (1 + random.uniform(0, 0.002))
            low = current_price * (1 - random.uniform(0, 0.002))
            open_price = current_price + random.uniform(-0.001, 0.001) * current_price
            close = current_price

            prices.append({
                "timestamp": bar_time,
                "open": open_price,
                "high": high,
                "low": low,
                "close": close,
                "volume": random.randint(100, 10000)
            })

        return prices

    async def validate_contract(self, contract: Contract) -> bool:
        await self.connect()
        if not self._connected:
            raise HTTPException(status_code=500, detail="Not connected")

        # Simulate contract validation with basic rules
        if contract.contract_type == ContractType.STOCK:
            return bool(re.match(r'^[A-Z]{1,5}$', contract.symbol))
        elif contract.contract_type == ContractType.FUTURE:
            return bool(re.match(r'^[A-Z]{2,4}$', contract.symbol)) and bool(contract.expiry)
        elif contract.contract_type == ContractType.ETF:
            return bool(re.match(r'^[A-Z]{1,5}$', contract.symbol))
        return False

    async def get_contract_id(self, contract: Contract) -> int:
        await self.connect()
        if not self._connected:
            raise HTTPException(status_code=500, detail="Not connected")
        full_string = f"{contract.symbol}{contract.contract_type}{contract.exchange}{contract.currency}{contract.expiry}"
        alphabet = list("_abcdefghijklmnopqrstuvwzyz0123456789")
        output = int("".join([str(alphabet.index(c)) for c in full_string.lower()]))
        return output

    async def get_quote_by_contract_id(self, exchange:str, contract_id:int) -> Quote:
        await self.connect()
        if not self._connected:
            raise HTTPException(status_code=500, detail="Not connected")

        # For test broker, we'll simulate a contract with a random symbol
        symbol = f"SYM_{contract_id}_{exchange}"

        # Generate prices for this symbol
        prices = await self._simulate_price(symbol)

        return Quote(
            symbol=symbol,
            bid=prices["bid"],
            ask=prices["ask"],
            last=prices["last"],
            timestamp=datetime.now()
        )

    async def get_current_minute_bar_open(self, contract_id:int, exchange:str) -> float:
        await self.connect()
        if not self._connected:
            raise HTTPException(status_code=500, detail="Not connected")

        # For test broker, we'll simulate a current minute bar open price
        # Generate a symbol based on contract_id and exchange
        symbol = f"SYM_{contract_id}_{exchange}"

        # Get simulated price for this symbol
        prices = await self._simulate_price(symbol)

        # Return the simulated open price (using last price as a proxy)
        return prices["last"]

    async def close_all_positions(self) -> int:
        await self.connect()
        if not self._connected:
            raise HTTPException(status_code=500, detail="Not connected")

        # Close all positions by creating opposite orders
        try:
            for symbol, position in self._positions.items():
                if position == 0:
                    continue

                # Create a market order in the opposite direction
                side = OrderSide.SELL if position > 0 else OrderSide.BUY
                quantity = abs(position)

                # Create a simple contract for this symbol
                contract = Contract(symbol=symbol, contract_type=ContractType.STOCK)

                # Create a market order
                order_request = Order(
                    trade=TradeInstruction(
                        strategy_name="position_close",
                        contract_id=await self.get_contract_id(contract),
                        exchange="SMART",
                        symbol=symbol,
                        side=side,
                        quantity=quantity,
                        order_type=OrderType.MARKET,
                        broker="TEST"
                    ),
                    price=0.0,  # Market order doesn't need a price
                    timestamp=datetime.now().strftime("%Y-%m-%dT%H:%M:%SZ")
                )

                # Place the order
                await self.place_order(order_request)

                # Reset the position to zero
                self._positions[symbol] = 0

            return 1  # Success
        except Exception as e:
            print(e)
            return 0  # Failure

    async def get_positions(self) -> List[dict]:
        await self.connect()
        if not self._connected:
            raise HTTPException(status_code=500, detail="Not connected")

        positions = []
        for symbol, quantity in self._positions.items():
            if quantity != 0:  # Only include non-zero positions
                # Create a simple contract for this symbol
                contract = Contract(symbol=symbol, contract_type=ContractType.STOCK)

                # Get current price for this symbol
                prices = await self._simulate_price(symbol)

                positions.append({
                    "symbol": symbol,
                    "position": quantity,
                    "avgCost": prices["last"],  # Use current price as average cost for simplicity
                    "contract": {
                        "symbol": symbol,
                        "conId": await self.get_contract_id(contract)
                    }
                })

        return positions

    async def get_trades(self) -> List[Trade]:
        await self.connect()
        if not self._connected:
            raise HTTPException(status_code=500, detail="Not connected")

        trades = []

        # Convert orders and fills to Trade objects
        for order_id, order in self._orders.items():
            # Check if this order has a fill
            fill = self._fills.get(order_id)

            # Determine quantity and price based on fill status
            quantity = 0
            price = 0.0
            status = OrderStatus.Submitted

            if fill:
                quantity = fill.quantity
                price = fill.price
                status = OrderStatus.Filled

            trade = Trade(
                order_id=int(order_id.split('_')[1]),  # Extract numeric part of order_id
                contract_id=order.trade.contract_id,
                time=datetime.now(),
                quantity=quantity,
                price=price,
                side=order.trade.side,
                order_status=status
            )

            trades.append(trade)

        return trades

# class TestIBKR(IBKRBroker):
#     def __init__(self):
#         super().__init__()
#         self.client_id = 0

# Update BrokerFactory to include TestBroker
class BrokerFactory:
    _brokers: Dict[str, BrokerInterface] = {}

    @classmethod
    def get_broker(cls, broker_type: str) -> BrokerInterface:
        if broker_type not in cls._brokers:
            if broker_type == "IB":
                cls._brokers[broker_type] = IBKRBroker()
            # elif broker_type == "IB_TEST":
            #     cls._brokers[broker_type] = TestIBKR()                
            # elif broker_type == "SCHWAB":
            #     cls._brokers[broker_type] = SchwabBroker()
            elif broker_type == "TEST":
                cls._brokers[broker_type] = TestBroker()
            else:
                raise ValueError(f"Unsupported broker type: {broker_type}")
        return cls._brokers[broker_type]


