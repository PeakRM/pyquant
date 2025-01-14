from fastapi import FastAPI, HTTPException
from typing import List
from broker_interface import BrokerFactory
from models import Contract, OrderRequest
from datetime import datetime 
from typing import Optional
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:8080"],  # Your frontend URL
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)
# FastAPI Application
app = FastAPI(title="Multi-Broker Trading API")


@app.post("/api/{broker}/quote")
async def get_quote(broker: str, contract: Contract):
    broker_instance = BrokerFactory.get_broker(broker)
    return await broker_instance.get_quote(contract)

@app.get("/api/{broker}/fills")
async def get_fills(broker: str, order_id: Optional[str] = None):
    broker_instance = BrokerFactory.get_broker(broker)
    return await broker_instance.get_fills(order_id)

@app.post("/api/{broker}/order")
async def place_order(broker: str, order_request: OrderRequest):
    broker_instance = BrokerFactory.get_broker(broker)
    return await broker_instance.place_order(order_request)

@app.post("/api/{broker}/historical")
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

