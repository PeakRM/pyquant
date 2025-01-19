from enum import Enum
from pydantic import BaseModel
from typing import Optional, Union
from datetime import datetime

# Data Models
class OrderType(str, Enum):
    MARKET = "MARKET"
    LIMIT = "LIMIT"

class OrderSide(str, Enum):
    BUY = "BUY"
    SELL = "SELL"
    HOLD = "HOLD"

class ContractType(str, Enum):
    STOCK = "STK"
    FUTURE = "FUT"
    ETF = "ETF"

class Contract(BaseModel):
    symbol: str
    contract_type: ContractType
    exchange: Optional[str] = None
    currency: str = "USD"
    expiry: Optional[str] = None

class Trade(BaseModel):
    strategy_name: str
    contract_id: int
    exchange: str
    symbol: str
    side: OrderSide
    quantity: Union[int, float]

class Order(BaseModel):
    trade: Trade
    price: float
    timestamp: str
    # contract: Contract
    # order_type: OrderType
    # side: OrderSide
    # quantity: float
    # limit_price: Optional[float] = None

class Fill(BaseModel):
    order_id: str
    contract: Contract
    execution_time: datetime
    quantity: float
    price: float
    side: OrderSide

class Quote(BaseModel):
    symbol: str
    bid: Optional[float]
    ask: Optional[float]
    last: Optional[float]
    timestamp: datetime
