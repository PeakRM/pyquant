import pandas_market_calendars as mcal
import datetime
import pandas as pd
from typing import Dict, Union, Literal, Any
from pydantic import BaseModel
import sys
import requests
import httpx
import json
from definitions import *
import logging
import trade_client


import datetime as dt
import time
import pytz


# Create and configure logger
logging.basicConfig(filename="newfile.log",
                    format='%(asctime)s %(message)s')

# Creating an object
logger = logging.getLogger()

# Setting the threshold of logger to DEBUG
logger.setLevel(logging.DEBUG)



STRATEGY_NAME="Test"
def run(account_data: AccountData, market1:str):

    r = requests.get(f"{BROKER_API}/historicalData?symbol={market1}&securityType=STK&broker=IB&bar_size=1%20day&lookback=60%20D")
    assert r.status_code == httpx.codes.OK, r.raise_for_status()
    market1_data = pd.DataFrame.from_records(r.json())
    market1_data['date'] = pd.to_datetime(market1_data.date)
    trades = []
    logging.info(account_data)


    if int(account_data.position.quantity) != 0:
        r = requests.get(f"{BROKER_API}/currentBarOpen?symbol={market1}&securityType=FUT&broker=IB")
        assert r.status_code == httpx.codes.OK, r.raise_for_status()
        current_bar_open = pd.DataFrame.from_records(r.json())
        if current_bar_open > account_data.position.cost_basis:
            try:
                logger.info("Generating closing trade")
                trade = Trade(strategy_name=STRATEGY_NAME,
                            symbol=account_data.position.symbol,
                            contract_id=account_data.position.contract_id,
                            exchange=account_data.position.exchange,
                            side='SELL',
                            quantity= 1)
            except Exception as e:
                logger.error(e)
            logger.info(trade)
            trades.append(trade)
                
    if int(account_data.position.quantity) == 0:
        logger.info("Generating opening trade")
        try:
            trade = Trade(strategy_name=STRATEGY_NAME,
                          symbol=account_data.position.symbol,
                          contract_id=account_data.position.contract_id,
                          exchange=account_data.position.exchange,
                          side='BUY',
                          quantity=1)
            trades.append(trade)
            logger.info(trades)
        except Exception as e:
            logger.error(e)


    if len(trades) == 2:
        logger.info("Rolling position")
        trades = [Trade(strategy_name=STRATEGY_NAME,
                        symbol=account_data.position.symbol,
                         contract_id=account_data.position.contract_id,
                         exchange=account_data.position.exchange,
                        side='HOLD',
                        quantity=0)]   
    logger.info(f"pys3: {json.dumps(trades[0].model_dump())}")
    return trades

def generate_test_trade():
    return  [Trade(strategy_name=STRATEGY_NAME,
                        symbol='MES',
                         contract_id=99999999,
                         exchange='cme',
                        side='BUY',
                        quantity=1)]  




def is_within_est_business_hours() -> bool:
    # Define EST timezone
    est = pytz.timezone('US/Eastern')
    
    # Get the current time in EST
    now_est = dt.datetime.now(est)
    print("Current Time: ", now_est )
    
    # Define start and end times in EST
    start_time = dt.time(9, 30)  # 9:30 AM
    end_time = dt.time(23, 59)   # 4:30 PM
    
    # Check if current time is within the range
    return start_time <= now_est.time() <= end_time


if __name__ == "__main__":

    while True:

        if is_within_est_business_hours():

            # trade=run(account_data=AccountData(**{
            #                 "position":{
            #                 "symbol":"MES",
            #                 "exchange":"CME",
            #                 "quantity":0.0,
            #                 "cost_basis":0.0,
            #                 "datetime": "2024-08-25T09:00:00Z",
            #                 "contract_id":654503314},
            #                 "buying_power":16307.37
            #                 }) ,market1="SPY")
            trade = generate_test_trade()

            if trade:
                try:
                    trade_client.send_trade(trade[0])
                except Exception as e:
                    print("Failed to submit trade:", e)

            time.sleep(60)
        time.sleep(1)


