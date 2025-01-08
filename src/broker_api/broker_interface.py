from abc import ABC, abstractmethod
from typing import List

class BrokerInterface(ABC):
    
    @abstractmethod
    async def request_quote(self, symbol: str) -> dict:
        pass

    @abstractmethod
    async def request_fills(self) -> List[dict]:
        pass

    @abstractmethod
    async def request_fills_for_order(self, order_id: int) -> List[dict]:
        pass

    @abstractmethod
    async def place_limit_order(self, symbol: str, qty: int, price: float) -> dict:
        pass

    @abstractmethod
    async def place_market_order(self, symbol: str, qty: int) -> dict:
        pass

    @abstractmethod
    async def get_historical_data(self, symbol: str, start_date: str, end_date: str) -> dict:
        pass

    @abstractmethod
    async def validate_contract(self, symbol: str, exchange: str, currency: str) -> dict:
        pass
