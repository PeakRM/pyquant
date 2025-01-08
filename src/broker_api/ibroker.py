from ib_async import IB
from typing import List
from broker_interface import BrokerInterface

class IBroker(BrokerInterface):
    
    def __init__(self):
        self.ib = IB()

    async def connect(self):
        if not self.ib.is_connected():
            await self.ib.connect('127.0.0.1', 7497, clientId=123)

    async def request_quote(self, symbol: str) -> dict:
        await self.connect()
        return await self.ib.request_quote(symbol)

    async def request_fills(self) -> List[dict]:
        await self.connect()
        return await self.ib.request_fills()

    async def request_fills_for_order(self, order_id: int) -> List[dict]:
        await self.connect()
        return await self.ib.request_fills_for_order(order_id)

    async def place_limit_order(self, symbol: str, qty: int, price: float) -> dict:
        await self.connect()
        return await self.ib.place_limit_order(symbol, qty, price)

    async def place_market_order(self, symbol: str, qty: int) -> dict:
        await self.connect()
        return await self.ib.place_market_order(symbol, qty)

    async def get_historical_data(self, symbol: str, start_date: str, end_date: str) -> dict:
        await self.connect()
        return await self.ib.get_historical_data(symbol, start_date, end_date)

    async def validate_contract(self, symbol: str, exchange: str, currency: str) -> dict:
        await self.connect()
        return await self.ib.validate_contract(symbol, exchange, currency)
