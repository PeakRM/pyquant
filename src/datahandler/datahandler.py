import pandas as pd
import databento as db
from dotenv import load_dotenv
import os
from datetime import datetime, timedelta
import sqlite3
from typing import List

# def load_env():
#     try:
#         load_dotenv('./shared/.env')
#         db_path = r"./shared/securities_master.db"
#     except Exception as e:
#         print(e)
#         load_dotenv('../shared_files/.env')
#         db_path = r"../shared_files/securities_master.db"
 



def get_front_month_contracts(root_symbols: List[str]=["ES"], dataset:str="GLBX.MDP3") -> pd.DataFrame:
    root_symbols = [f"{rs.upper()}.FUT" for rs in root_symbols if rs.upper()[-3:] != "FUT"]
    end = datetime.today() - timedelta(days=1)
    start = end - timedelta(days=20)
    cost = client.metadata.get_cost(
            dataset=dataset,
            symbols=root_symbols,
            stype_in="parent",
            start=start.strftime("%Y-%m-%d"),
            end=end.strftime("%Y-%m-%d"),
            schema="statistics")
    print("Cost: ",cost)
    x = input("Would you like to proceed? ")
    if x.lower() != "y":
        return pd.DataFrame()
        
    stats = client.timeseries.get_range(
            dataset=dataset,
            symbols=root_symbols,
            stype_in="parent",
            start=start.strftime("%Y-%m-%d"),
            end=end.strftime("%Y-%m-%d"),
            schema="statistics").to_df()

    stats = stats[stats.stat_type.isin([db.StatType.OPEN_INTEREST])].copy()
    stats["stat"] = stats["stat_type"].map({db.StatType.OPEN_INTEREST: "open interest"})
    stats["ts_ref_date"] = stats["ts_ref"].dt.floor("D")
    return (stats.reset_index()
                 .loc[stats.reset_index().groupby(["ts_ref_date"]).quantity.idxmax(),
                     ['ts_ref_date','symbol']])
   
   
ContractMonths: dict[str,str] = {}
   
   
if __name__=="__main__":
    load_dotenv()
    # db = sqlite3.connect(db_path)
    # print(os.environ.get('DB_API_KEY'))
    client = db.Historical(key=os.environ.get('DB_API_KEY'))
    data = get_front_month_contracts(["MCL","MYM","NQ"])
    print(data)

   
   
   

