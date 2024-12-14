from typing import Dict, Any, List

from fastapi import FastAPI

import random

app = FastAPI()


@app.get("/")
def read_root():
    return {"Test Server": "Running"}


@app.get("/quoteByConId")
def read_item(conId: int, exchange: str ) -> Dict[str, float]:
    return {"price":300.0}



@app.get("/fills")
def get_fills(Id: int) -> List[Dict[str, Any]]:
    fill_chance = .3
    fill_status = "pending"
    price, quantity = 0.0, 0.0
    if random.random() < fill_chance:
        fill_status = "filled"
        price = 300.
        quantity=1.
        print("filled")

    return [{"id": 1002, "price":100., "status":"filled", "quantity":1},
            {"id": 1003, "price":245.0, "status":fill_status, "quantity": quantity},
            {"id": Id, "price":price, "status":fill_status, "quantity": quantity},
            {"id": 1004, "price":43.23, "status":"pending","quantity": 0}]
