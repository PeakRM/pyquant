import pandas_market_calendars as mcal
from datetime import datetime
import pandas as pd
from typing import Dict, Union, Literal, Any, List
from pydantic import BaseModel
import sys
import requests
import httpx
import json
from definitions import *
import logging
import trade_client
from config import load_and_parse_config

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



# def run(account_data: AccountData, market1:str):
def run(account_data: AccountData, market1:str):
    today = dt.datetime.today()
    r = requests.post(f"{BROKER_API}/historicalData",data=dict(symbol=market1,
                                                               securityType="STK",
                                                               bar_size="1 day",
                                                               end_date=today,
                                                               start_date=today - dt.timedelta(days=60))
                                                            #    lookback="60D")
                                                               )
    assert r.status_code == httpx.codes.OK, r.raise_for_status()
    market1_data = pd.DataFrame.from_records(r.json())
    market1_data['date'] = pd.to_datetime(market1_data.date)
    trades = []
    logging.info(account_data)


    if int(account_data.position.quantity) != 0:
        r = requests.post(f"{BROKER_API}/currentBarOpen",data=dict(contract_id=account_data.position.contract_id,
                                                                   exchange=account_data.position.exchange))
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

def check_position(symbol:str, strategy_name:str)->str:
    with open("/shared/positions.json",'r') as position_file:
        positions = json.load(position_file)
    try:
        strategy_position = positions[f"{strategy_name}-{symbol}"]
        position_status = strategy_position['status'].lower() 
    except KeyError:
        print("no position found in file")
        position_status=""
    return  position_status
def get_position_data(symbol:str, strategy_name:str):
    with open("/shared/positions.json",'r') as position_file:
        positions = json.load(position_file)
    try:
        strategy_position = positions[f"{strategy_name}-{symbol}"]
        return strategy_position
    except KeyError:
        print("no position found in file")
        position_status=""
    # return  position_status

def generate_test_trade(symbol:str, exchange:str, strategy_name:str, contract_id:int) -> List[Trade]:
    position_data = get_position_data(symbol=symbol, strategy_name=strategy_name)
    position_status=position_data["status"]
    r = requests.post(f"{BROKER_API}/currentBarOpen",data=dict(contract_id=contract_id,
                                                                exchange=exchange))
    assert r.status_code == httpx.codes.OK, r.raise_for_status()
    current_bar_open = pd.DataFrame.from_records(r.json())
    if position_status == "filled":
        if current_bar_open > float(position_data["cost_basis"]):
            return  [Trade(strategy_name=strategy_name,
                        symbol=symbol,
                         contract_id=contract_id,
                         exchange=exchange,
                         side='SELL',
                        quantity=1)] 
    elif position_status == "pending":
        return []
    else:
        # no open position
        print(position_status)

        return  [Trade(strategy_name=strategy_name,
                        symbol=symbol,
                         contract_id=99999999,
                         exchange=exchange,
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

def initialize(config_data):
    mkt = config_data["market"].split(":")
    strategy_settings = {}
    strategy_settings["exchange"] = mkt[0]
    strategy_settings["symbol"] = mkt[1]
    strategy_settings["contract_id"] = float(config_data["contract_id"])
    strategy_settings["interval"] = get_interval(config_data["timeframe"])
    return strategy_settings

def get_interval(freq:str)->int:
    if freq=="1 min":
        return 60
    if freq=="5 min":
        return 60*5
    elif freq=="15 min":
        return 60*15
    elif freq=="1 hour":
        return 60*60
    elif freq=="4 hour":
        return 60*60*4
    else:
        return 60*60*24

if __name__ == "__main__":

    setup_name = sys.argv[1]
    config_data = load_and_parse_config("/shared/strategy-config.json", setup_name=setup_name)
    strategy_settings = initialize(config_data)
    strategy_settings["name"] = setup_name.split("-")[0]
    while True:

        # if is_within_est_business_hours():
        if True:
            run_time = datetime.now()
            # trade=run(symbol=strategy_settings['symbol'],
            #           exchange=strategy_settings['exchange'],
            #           strategy_name=strategy_settings["name"])
                    #   =AccountData(**{
                            # "position":{
            #                 "symbol":"MES",
            #                 "exchange":"CME",
            #                 "quantity":0.0,
            #                 "cost_basis":0.0,
            #                 "datetime": "2024-08-25T09:00:00Z",
            #                 "contract_id":654503314},
            #                 "buying_power":16307.37
            #                 }) ,market1="SPY")
            trade = generate_test_trade(symbol=strategy_settings['symbol'],
                                        exchange=strategy_settings['exchange'],
                                        strategy_name=strategy_settings["name"],
                                        strategy_name=strategy_settings["contract_id"])
            print(trade)

            if trade:
                try:
                    trade_client.send_trade(trade[0])
                except Exception as e:
                    print("Failed to submit trade:", e)

            time.sleep(strategy_settings["interval"] + (datetime.now()-run_time).seconds)
        time.sleep(1)


