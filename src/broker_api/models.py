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
    # HOLD = "HOLD"

class ContractType(str, Enum):
    STOCK = "STK"
    FUTURE = "FUT"
    ETF = "ETF"

class Contract(BaseModel):
    symbol: Optional[str] = None
    contract_type: Optional[ContractType]=None
    contract_id: Optional[int] = None
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
    order_id: int
    contract_id: int
    time: datetime
    quantity: int
    price: float
    side: OrderSide

class Quote(BaseModel):
    symbol: str
    bid: Optional[float]
    ask: Optional[float]
    last: Optional[float]
    timestamp: datetime
