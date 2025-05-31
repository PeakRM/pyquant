import psycopg2
import os
def round_to_increment(number, increment):
    """
    Rounds a number to the nearest multiple of the specified increment.
    
    Args:
        number (float): The number to round
        increment (float): The minimum change value (e.g., 0.25, 0.10, 1.00)
    
    Returns:
        float: The rounded number
    
    Examples:
        >>> round_to_increment(100.38, 0.25)
        100.5
        >>> round_to_increment(100.38, 0.1)
        100.4
        >>> round_to_increment(100.38, 1.0)
        100.0
        >>> round_to_increment(7.23, 0.5)
        7.0
        >>> round_to_increment(7.26, 0.5)
        7.5
    """
    return round(round(number / increment) * increment, 2)

def get_futures_multiplier(symbol: str)-> float:
    user=os.getenv('DB_USER')
    host=os.getenv('DB_HOST')
    password=os.getenv('DB_PASSWORD')
    name=os.getenv('DB_NAME')
    port=os.getenv('DB_PORT')
    try:
        with psycopg2.connect(f"postgresql://{user}:{password}@{host}:{port}/{name}") as conn:
            with conn.cursor() as cursor:
                cursor.execute(f"SELECT multiplier FROM futures_contracts WHERE symbol='{symbol}'")
                results = cursor.fetchall()
                print(results)
                return results
    except psycopg2.Error as e:
        print(f"Database error: {e}")
        return 1.0

        
def get_futures_minimum_tick(symbol: str)-> float:
    user=os.getenv('DB_USER')
    host=os.getenv('DB_HOST')
    password=os.getenv('DB_PASSWORD')
    name=os.getenv('DB_NAME')
    port=os.getenv('DB_PORT')
    try:
        with psycopg2.connect(f"postgresql://{user}:{password}@{host}:{port}/{name}") as conn:
            with conn.cursor() as cursor:
                cursor.execute("SELECT minimum_tick FROM futures_contracts WHERE symbol=%s", (symbol,))
                result = cursor.fetchone()
                return float(result[0]) if result else 1.0
    except psycopg2.Error as e:
        print(f"Database error: {e}")
        return 1.