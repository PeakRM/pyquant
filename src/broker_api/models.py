from enum import Enum
from pydantic import BaseModel
from typing import Optional, Union
from datetime import datetime

# Data Models
class OrderType(str, Enum):
    MARKET = "MKT"
    LIMIT = "LMT"

class OrderSide(str, Enum):
    BUY = "BUY"
    SELL = "SELL"
    # HOLD = "HOLD"

class ContractType(str, Enum):
    STOCK = "STK"
    FUTURE = "FUT"
    ETF = "ETF"

class OrderStatus(str, Enum):
    Filled = "Filled"
    Pending = "Pending"
    Cancelled = "Cancelled"
    Submitted = "Submitted"


class Contract(BaseModel):
    symbol: Optional[str] = None
    contract_type: Optional[ContractType]=None
    contract_id: Optional[int] = None
    exchange: Optional[str] = None
    currency: str = "USD"
    expiry: Optional[str] = None

class TradeInstruction(BaseModel):
    strategy_name: str
    contract_id: int
    exchange: str
    symbol: str
    side: OrderSide
    quantity: Union[int, float]
    order_type: OrderType = OrderType.LIMIT  # Default to limit order
    broker: str = 'IB'  # Default to Interactive Brokers

class Order(BaseModel):
    trade: TradeInstruction
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

class Trade(BaseModel):
    order_id: int
    contract_id: int
    time: datetime
    quantity: int
    price: float
    side: OrderSide
    order_status: OrderStatus