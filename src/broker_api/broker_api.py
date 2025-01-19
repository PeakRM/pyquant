from fastapi import FastAPI, HTTPException
from broker_interface import BrokerFactory
from models import Contract, Order
from datetime import datetime 
from typing import Optional
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI(title="Multi-Broker Trading API")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:8080"],
    allow_credentials=True,
    allow_methods=["POST", "OPTIONS"],
    allow_headers=["Content-Type"],
)

@app.get("/")
async def get_index() -> str:
    return "Multi-Broker Trading API"

@app.post("/api/{broker}/quote")
async def get_quote(broker: str, contract: Contract):
    broker_instance = BrokerFactory.get_broker(broker)
    return await broker_instance.get_quote(contract)

@app.post("/api/{broker}/quote/{exchange}/{contract_id}")
async def get_quote_by_contract_id(broker: str, exchange:str, contract_id:int):
    if broker=="TEST": 
        raise NotImplementedError("Implement this fuction in TEST broker first.")
    broker_instance = BrokerFactory.get_broker(broker)
    return await broker_instance.get_quote_by_contract_id(exchange, contract_id)

@app.get("/api/{broker}/fills")
async def get_fills(broker: str, order_id: Optional[str] = None):
    broker_instance = BrokerFactory.get_broker(broker)
    return await broker_instance.get_fills(order_id)

@app.post("/api/{broker}/order")
async def place_order(broker: str, order: Order):
    broker_instance = BrokerFactory.get_broker(broker)
    return await broker_instance.place_order(order)

@app.post("/api/{broker}/historicalData")
async def get_historical_data(
    broker: str,
    contract: Contract,
    start_time: datetime,
    end_time: datetime,
    bar_size: str
):
    broker_instance = BrokerFactory.get_broker(broker)
    return await broker_instance.get_historical_data(contract, start_time, end_time, bar_size)

@app.post("/api/{broker}/validate-contract")
async def validate_contract(broker: str, contract: Contract):
    broker_instance = BrokerFactory.get_broker(broker)
    return await broker_instance.validate_contract(contract)

@app.post("/api/{broker}/contract-id")
async def get_contract_id(broker: str, contract: Contract):
    broker_instance = BrokerFactory.get_broker(broker)
    return await broker_instance.get_contract_id(contract)

@app.post("/api/{broker}/currentBarOpen")
async def get_quote_by_contract_id(broker: str, exchange:str, contract_id:int):
    if broker=="TEST": 
        raise NotImplementedError("Implement this fuction in TEST broker first.")
    broker_instance = BrokerFactory.get_broker(broker)
    return await broker_instance.get_current_bar_open(exchange, contract_id)
