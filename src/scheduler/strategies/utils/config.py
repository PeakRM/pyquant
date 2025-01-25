import json
from typing import Dict, Any



def load_strategy_config(filepath: str) -> Dict[str, Any]:
    try:
        with open(filepath, 'r') as f:
            return json.load(f)
    except Exception as e:
        print("Strategy Config unable to load: ",e)
        return {}

def parse_strategy_config(config_data: Dict[str, Any],setup_name:str) -> Dict[str, Any]:
    strategy_name= setup_name.split('-')[0]
    try:
        return config_data[strategy_name]['setups'][setup_name]
    except Exception as e:
        print("Unable to Parse Configuration: ",e)
        print(config_data)
        print(setup_name)
        return {}


def load_and_parse_config(filepath: str, setup_name:str) -> Dict[str, Any]:
    try:
        data = load_strategy_config(filepath=filepath)
        print(type(data),data)
        return parse_strategy_config(config_data=data, setup_name=setup_name)
    except Exception as e:
        print("Unable to Load / Parse Configuration: ",e)
        return {}
    
if __name__ == "__main__":

    print(load_and_parse_config("C:\\Users\\Jon\\Projects\\pyquant\\shared_files\\strategy-config.json", "Test-MYM"))