import pandas as pd
import databento as db
from dotenv import load_dotenv
import os
from datetime import datetime, timedelta

try:
    load_dotenv('./shared/.env')
except Exception as e:
    load_dotenv('./shared_files/.env')
 
client = db.Historical(key=os.environ.get('DB_API_KEY'))


def get_front_month(root_symbol: str = "ES") -> pd.DataFrame():
   end = datetime.today()
   start = end - timedelta(days=90)
   stats = client.timeseries.get_range(
           dataset="GLBX.MDP3",
          symbols=f"{root_symbol.upper()}.FUT",
          stype_in="parent",
          start=start.strftime("%Y-%m-%d"),
          end=end.strftime("%Y-%m-%d"),
          schema="statistics",
        # and convert it to a DataFrame
    ).to_df()
    
    stats1 = stats[
        stats.stat_type.isin([db.StatType.OPEN_INTEREST])
    ].copy()
    
    stats1["stat"] = stats1["stat_type"].map(
        {
            db.StatType.OPEN_INTEREST: "open interest",
        },
    )
    
    stats1["ts_ref_date"] = stats1["ts_ref"].dt.floor("D")
    return stats1.reset_index().loc[stats1.reset_index().groupby(["ts_ref_date"]).quantity.idxmax(),['ts_ref_date','symbol']]
    

