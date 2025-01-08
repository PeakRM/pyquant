from fastapi import FastAPI, HTTPException
from typing import List
from broker_manager import BrokerManager

app = FastAPI()

# Instantiate the broker manager
broker_manager = BrokerManager()


@app.get("/quotes")
async def get_quotes(symbol: str, broker: str = 'IB'):
    """
    Endpoint to request a quote for a given symbol from the specified broker.
    The broker can be switched dynamically using the 'broker' query parameter.
    """
    broker_manager.switch_broker(broker)
    try:
        quote = await broker_manager.request_quote(symbol)
        return {"symbol": symbol, "quote": quote}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/fills")
async def get_fills(broker: str = 'IB'):
    """
    Endpoint to request all fills from the specified broker.
    The broker can be switched dynamically using the 'broker' query parameter.
    """
    broker_manager.switch_broker(broker)
    try:
        fills = await broker_manager.request_fills()
        return {"fills": fills}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/fills/{order_id}")
async def get_fills_for_order(order_id: int, broker: str = 'IB'):
    """
    Endpoint to request fills for a specific order ID from the specified broker.
    The broker can be switched dynamically using the 'broker' query parameter.
    """
    broker_manager.switch_broker(broker)
    try:
        fills = await broker_manager.request_fills_for_order(order_id)
        return {"order_id": order_id, "fills": fills}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/place_limit_order")
async def place_limit_order(symbol: str, qty: int, price: float, broker: str = 'IB'):
    """
    Endpoint to place a limit order with the specified symbol, quantity, and price
    on the specified broker. The broker can be switched dynamically using the 'broker' query parameter.
    """
    broker_manager.switch_broker(broker)
    try:
        order = await broker_manager.place_limit_order(symbol, qty, price)
        return {"order": order}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/place_market_order")
async def place_market_order(symbol: str, qty: int, broker: str = 'IB'):
    """
    Endpoint to place a market order with the specified symbol and quantity
    on the specified broker. The broker can be switched dynamically using the 'broker' query parameter.
    """
    broker_manager.switch_broker(broker)
    try:
        order = await broker_manager.place_market_order(symbol, qty)
        return {"order": order}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/historical_data")
async def get_historical_data(symbol: str, start_date: str, end_date: str, broker: str = 'IB'):
    """
    Endpoint to request historical market data for a given symbol from the specified broker.
    The broker can be switched dynamically using the 'broker' query parameter.
    """
    broker_manager.switch_broker(broker)
    try:
        data = await broker_manager.get_historical_data(symbol, start_date, end_date)
        return {"symbol": symbol, "data": data}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/validate_contract")
async def validate_contract(symbol: str, exchange: str, currency: str, broker: str = 'IB'):
    """
    Endpoint to validate if a contract is valid for trading with the specified broker.
    The broker can be switched dynamically using the 'broker' query parameter.
    """
    broker_manager.switch_broker(broker)
    try:
        contract = await broker_manager.validate_contract(symbol, exchange, currency)
        return {"contract": contract}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
