from abc import ABC, abstractmethod
from typing import List, Optional, Dict, Any
from datetime import datetime
from fastapi import HTTPException
import ib_insync
from ib_insync import util
from models import *
from dotenv import load_dotenv
import os
from pathlib import Path
import random
from datetime import datetime, timedelta
import numpy as np
import re

# Load environment variables
env_path = Path('.env')
if not env_path.exists():
    raise FileNotFoundError(f"Environment file not found at {env_path}")
load_dotenv(env_path)

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
        bar_size: str
    ) -> List[Dict[str, Any]]:
        pass

    @abstractmethod
    async def validate_contract(self, contract: Contract) -> bool:
        pass

# Interactive Brokers Implementation
class IBBroker(BrokerInterface):
    def __init__(self):
        self.host = os.getenv('IB_HOST', '127.0.0.1')
        self.port = int(os.getenv('IB_PORT', '7497'))
        self.client_id = int(os.getenv('IB_CLIENT_ID', '1'))
        self.ib = ib_insync.IB()
        self._connected = False

    async def connect(self) -> bool:
        if not self._connected:
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

    def _convert_contract(self, contract: Optional[Contract], contract_id:Optional[int],exchange:Optional[str]) -> ib_insync.Contract:
        if contract is None:
            try:
                return ib_insync.Contract(conId=contract_id, exchange=exchange)
            except Exception:
                raise ValueError(f"You did not pass the correct parameters: \n\t{contract}\n\t{contract_id}\n\t{exchange}")


        if contract.contract_type == ContractType.STOCK:
            return ib_insync.Stock(contract.symbol, contract.exchange or "SMART", contract.currency)
        elif contract.contract_type == ContractType.FUTURE:
            return ib_insync.Future(contract.symbol, contract.expiry, contract.exchange)
        elif contract.contract_type == ContractType.ETF:
            return ib_insync.Stock(contract.symbol, contract.exchange or "SMART", contract.currency)
        raise ValueError(f"Unsupported contract type: {contract.contract_type}")

    async def get_quote(self, contract: Contract) -> Quote:
        await self.connect()
        ib_contract = self._convert_contract(contract)
        
        try:
            await self.ib.qualifyContractsAsync(ib_contract)
            tickers = self.ib.reqMktData(ib_contract, snapshot=True)
            
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

    async def get_quote_by_contract_id(self,exchange:str, contract_id:int) -> Quote:
        await self.connect()
        ib_contract = ib_insync.Contract(conId=contract_id, exchange=exchange)
        
        try:
            await self.ib.qualifyContractsAsync(ib_contract)
            tickers = self.ib.reqMktData(ib_contract, snapshot=True)

            return Quote(
                symbol=ib_contract.symbol,
                bid=tickers.bid,
                ask=tickers.ask,
                last=tickers.last,
                timestamp=datetime.now()
            )
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Failed to get quote: {str(e)}")
  
    async def get_fills(self, order_id: Optional[str] = None) -> List[Fill]:
        await self.connect()
        fills = []
        
        try:
            # if order_id is None:
            #     for fill in self.ib.fills():
            #         # print (order.order.orderId)
            #         # if fill[1]['orderId'] == orderId:
            #         fills.append(Fill(
            #             order_id = fill[1]['orderId'],
            #             order_status = "Filled",
            #             qty = fill[1]['cumQty'],
            #             avg_fill_price = fill[1]['avgPrice'],
            #             time_submitted = fill[-1],# datetime.datetime.now().strftime("%y-%m-%d %H:%M:%S")
            #             conId = fill[0]["conId"]))
            # else:
            trades = self.ib.trades()
            for trade in trades:
                if order_id and trade.order.orderId != order_id:
                    continue
                    
                fills.append(Fill(
                    order_id=str(trade.order.orderId),
                    contract=Contract(
                        symbol=trade.contract.symbol,
                        contract_type=ContractType(trade.contract.secType),
                        exchange=trade.contract.exchange,
                        currency=trade.contract.currency,
                        expiry=getattr(trade.contract, 'lastTradeDateOrContractMonth', None)
                    ),
                    execution_time=trade.time,
                    quantity=trade.execution.shares,
                    price=trade.execution.price,
                    side=OrderSide.BUY if trade.order.action == "BUY" else OrderSide.SELL
                ))
            
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Failed to get fills: {str(e)}")
            
        return fills

    async def place_order(self, order: Order) -> str:
        await self.connect()
        ib_contract = self._convert_contract(contract_id=order.trade.contract_id,
                                             exhcange=order.trade.exchange)
        
        try:
            await self.ib.qualifyContractsAsync(ib_contract)
            
            ib_order = ib_insync.Order(
                action="BUY" if order.trade.side == OrderSide.BUY else "SELL",
                totalQuantity=order.trade.quantity,
                # orderType="MKT" if order.order_type == OrderType.MARKET else "LMT",
                orderType="LMT",
                # lmtPrice=order.limit_price if order.order_type == OrderType.LIMIT else None
                lmtPrice=order.price)
            
            trade = await self.ib.placeOrderAsync(ib_contract, ib_order)
            return str(trade.order.orderId)# TODO TEST THIS
            
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Failed to place order: {str(e)}")

    async def get_historical_data(
        self,
        contract: Contract,
        start_time: datetime,
        end_time: datetime,
        bar_size: str
    ) -> List[Dict[str, Any]]:
        await self.connect()
        ib_contract = self._convert_contract(contract)
        
        try:
            await self.ib.qualifyContractsAsync(ib_contract)
            bars = await self.ib.reqHistoricalDataAsync(
                ib_contract,
                endDateTime=end_time,
                durationStr=self._calculate_duration(start_time, end_time),
                barSizeSetting=bar_size,
                whatToShow='TRADES',
                useRTH=True
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

    async def validate_contract(self, contract: Contract) -> bool:
        await self.connect()
        ib_contract = self._convert_contract(contract)
        
        try:
            contracts = await self.ib.qualifyContractsAsync(ib_contract)
            print(contracts)
            return len(contracts) > 0
        except Exception:
            return False

    async def get_contract_id(self, contract: Contract) -> int:
        await self.connect()
        ib_contract = self._convert_contract(contract)
        
        try:
            contracts = await self.ib.qualifyContractsAsync(ib_contract)
            print(contracts)
            if len(contracts)==1:
                return contracts[0].conId
        except Exception:
            return False

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
        
    async def get_current_bar_open_price(self, contract_id:int, exchange:str) -> float:
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
                useRTH=True,
                formatDate=1)
            # print(time.ctime())
            # print(util.df(bars))
            return util.df(bars)['open'].iloc[-1]
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Failed to get current bar open: {str(e)}")


class TestBroker(BrokerInterface):
    def __init__(self):
        self._connected = False
        self._orders = {}  # Store orders by order_id
        self._fills = {}   # Store fills by order_id
        self._positions = {}  # Store positions by symbol
        self._prices = {}  # Store simulated prices by symbol
        
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
        return f"TEST_{int(datetime.now().timestamp())}_{random.randint(1000, 9999)}"

    async def _simulate_price(self, symbol: str) -> dict:
        """Generate simulated prices for a symbol."""
        await self.connect()
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

    async def get_fills(self, order_id: Optional[str] = None) -> List[Fill]:
        await self.connect()
        if not self._connected:
            raise HTTPException(status_code=500, detail="Not connected")
        
        fills = []
        if order_id:
            if order_id in self._fills:
                fills.append(self._fills[order_id])
        else:
            fills.extend(self._fills.values())
        
        return fills

    async def place_order(self, order_request: Order) -> str:
        await self.connect()
        if not self._connected:
            raise HTTPException(status_code=500, detail="Not connected")
        
        order_id = self._generate_order_id()
        self._orders[order_id] = order_request
        
        # Simulate immediate fill for market orders
        if order_request.order_type == OrderType.MARKET:
            prices = self._simulate_price(order_request.contract.symbol)
            fill_price = prices["ask"] if order_request.side == OrderSide.BUY else prices["bid"]
            
            fill = Fill(
                order_id=order_id,
                contract=order_request.contract,
                execution_time=datetime.now(),
                quantity=order_request.quantity,
                price=fill_price,
                side=order_request.side
            )
            self._fills[order_id] = fill
            
            # Update position
            position_delta = order_request.quantity if order_request.side == OrderSide.BUY else -order_request.quantity
            self._positions[order_request.contract.symbol] = self._positions.get(order_request.contract.symbol, 0) + position_delta
        
        # For limit orders, simulate fill only if price is favorable
        elif order_request.order_type == OrderType.LIMIT:
            prices = self._simulate_price(order_request.contract.symbol)
            if ((order_request.side == OrderSide.BUY and prices["ask"] <= order_request.limit_price) or
                (order_request.side == OrderSide.SELL and prices["bid"] >= order_request.limit_price)):
                
                fill = Fill(
                    order_id=order_id,
                    contract=order_request.contract,
                    execution_time=datetime.now(),
                    quantity=order_request.quantity,
                    price=order_request.limit_price,
                    side=order_request.side
                )
                self._fills[order_id] = fill
                
                position_delta = order_request.quantity if order_request.side == OrderSide.BUY else -order_request.quantity
                self._positions[order_request.contract.symbol] = self._positions.get(order_request.contract.symbol, 0) + position_delta
        
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
        

# Update BrokerFactory to include TestBroker
class BrokerFactory:
    _brokers: Dict[str, BrokerInterface] = {}
    
    @classmethod
    def get_broker(cls, broker_type: str) -> BrokerInterface:
        if broker_type not in cls._brokers:
            if broker_type == "IB":
                cls._brokers[broker_type] = IBBroker()
            elif broker_type == "TEST":
                cls._brokers[broker_type] = TestBroker()
            else:
                raise ValueError(f"Unsupported broker type: {broker_type}")
        return cls._brokers[broker_type]
