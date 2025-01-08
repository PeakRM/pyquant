from broker_interface import BrokerInterface
from ibroker import IBroker
from typing import List

class BrokerManager:
    
    def __init__(self):
        self.broker: BrokerInterface

    def switch_broker(self, broker_name: str):
        if broker_name == 'IB':
            self.broker = IBroker()
        # elif broker_name == 'TDA':
        #     self.broker = TDAmeritradeBroker()
        else:
            raise ValueError("Unsupported broker")

    async def request_quote(self, symbol: str) -> dict:
        return await self.broker.request_quote(symbol)

    async def request_fills(self) -> List[dict]:
        return await self.broker.request_fills()

    async def request_fills_for_order(self, order_id: int) -> List[dict]:
        return await self.broker.request_fills_for_order(order_id)

    async def place_limit_order(self, symbol: str, qty: int, price: float) -> dict:
        return await self.broker.place_limit_order(symbol, qty, price)

    async def place_market_order(self, symbol: str, qty: int) -> dict:
        return await self.broker.place_market_order(symbol, qty)

    async def get_historical_data(self, symbol: str, start_date: str, end_date: str) -> dict:
        return await self.broker.get_historical_data(symbol, start_date, end_date)

    async def validate_contract(self, symbol: str, exchange: str, currency: str) -> dict:
        return await self.broker.validate_contract(symbol, exchange, currency)
