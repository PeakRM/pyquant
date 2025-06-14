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
        bar_size: str
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
        await self.connect()
        ib_contract = self._convert_contract(contract)

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
                print(trade)
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
        await self.connect()
        ib_contract = self._convert_contract(contract_id=order.trade.contract_id,
                                             exchange=order.trade.exchange)

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

    async def get_historical_data(self, contract: Contract, start_time: datetime, end_time: datetime, bar_size: str) -> List[Dict[str, Any]]:
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

    async def get_historical_data_by_contract_id(self, contract_id: int, exchange:str, start_time: datetime,
                                                 end_time: datetime, bar_size: str) -> List[Dict[str, Any]]:
        await self.connect()
        ib_contract = self._convert_contract(contract_id=contract_id, exchange=exchange)

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
            return util.df(bars)
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Failed to get historical data: {str(e)}")

    async def validate_contract(self, contract: Contract) -> bool:
        await self.connect()
        ib_contract = self._convert_contract(contract)

        try:
            contracts = await self.ib.qualifyContractsAsync(ib_contract)
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

# Update BrokerFactory to include TestBroker
class BrokerFactory:
    _brokers: Dict[str, BrokerInterface] = {}

    @classmethod
    def get_broker(cls, broker_type: str) -> BrokerInterface:
        if broker_type not in cls._brokers:
            if broker_type == "IB":
                cls._brokers[broker_type] = IBKRBroker()
            elif broker_type == "TEST":
                cls._brokers[broker_type] = TestBroker()
            else:
                raise ValueError(f"Unsupported broker type: {broker_type}")
        return cls._brokers[broker_type]


