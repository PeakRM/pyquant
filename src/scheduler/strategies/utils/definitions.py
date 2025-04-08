from pydantic import BaseModel
from typing import Union, Literal
import datetime

# BROKER_API = "http://127.0.0.1:8000"
BROKER_API = "http://broker_api:8000"

class Position(BaseModel):
    symbol: str
    exchange: str
    quantity: float
    cost_basis: float
    datetime: Union[str, datetime.datetime]
    contract_id: int
    status: str

class AccountData(BaseModel):
    position: Position
    buying_power: float

class Trade(BaseModel):
    strategy_name: str
    contract_id: int
    exchange: str
    symbol: str
    side: Literal['BUY', 'SELL', 'HOLD']
    quantity: int
    order_type: Literal['MKT', 'LMT'] = 'LMT'  # Default to limit order
    broker: str = 'IB'  # Default to Interactive Brokers
