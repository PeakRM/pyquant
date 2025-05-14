#!/usr/bin/env python3
"""
Moving Average Crossover Strategy

This strategy uses two moving averages (fast and slow) to generate buy and sell signals.
When the fast MA crosses above the slow MA, a buy signal is generated.
When the fast MA crosses below the slow MA, a sell signal is generated.

Configuration parameters:
    - symbol: The trading symbol to analyze (e.g., "AAPL")
    - timeframe: Candlestick timeframe (e.g., "1h", "1d")
    - fast_ma: Fast moving average period (e.g., 20)
    - slow_ma: Slow moving average period (e.g., 50)
    - quantity: Number of shares to trade
    - stop_loss_pct: Stop loss percentage (e.g., 2.5)
"""

import json
import sys
import time
import requests
import pandas as pd
import numpy as np
import logging
from datetime import datetime

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class MovingAverageCrossoverStrategy:
    def __init__(self, config_path):
        """
        Initialize the strategy with configuration parameters.
        
        Args:
            config_path (str): Path to the JSON configuration file
        """
        self.config = self._load_config(config_path)
        self.symbol = self.config.get("symbol", "AAPL")
        self.timeframe = self.config.get("timeframe", "1h")
        self.fast_ma = self.config.get("fast_ma", 20)
        self.slow_ma = self.config.get("slow_ma", 50)
        self.quantity = self.config.get("quantity", 10)
        self.stop_loss_pct = self.config.get("stop_loss_pct", 2.5)
        
        self.positions = {}  # To track open positions
        
        logger.info(f"Initialized Moving Average Crossover strategy for {self.symbol}")
        logger.info(f"Parameters: Fast MA={self.fast_ma}, Slow MA={self.slow_ma}, Quantity={self.quantity}")
    
    def _load_config(self, config_path):
        """
        Load configuration from a JSON file.
        
        Args:
            config_path (str): Path to the JSON configuration file
            
        Returns:
            dict: Configuration parameters
        """
        try:
            with open(config_path, "r") as f:
                config = json.load(f)
            return config
        except Exception as e:
            logger.error(f"Error loading configuration: {e}")
            sys.exit(1)
    
    def fetch_market_data(self):
        """
        Fetch historical market data for the specified symbol and timeframe.
        
        In a real implementation, this would call a market data API.
        This example uses mock data for demonstration purposes.
        
        Returns:
            pandas.DataFrame: Historical market data with OHLCV columns
        """
        logger.info(f"Fetching market data for {self.symbol} ({self.timeframe})")
        
        # In a real implementation, fetch data from an API
        # For example:
        # api_url = f"https://api.example.com/v1/historical?symbol={self.symbol}&timeframe={self.timeframe}"
        # response = requests.get(api_url)
        # data = response.json()
        
        # Mock data generation for demonstration
        np.random.seed(42)  # For reproducibility
        num_candles = 200
        
        # Start with a base price and generate random movements
        base_price = 100.0
        daily_volatility = 0.01
        prices = [base_price]
        
        for _ in range(num_candles - 1):
            # Random price movement
            change_pct = np.random.normal(0.0002, daily_volatility)
            price = prices[-1] * (1 + change_pct)
            prices.append(price)
        
        # Generate OHLCV data
        dates = pd.date_range(end=datetime.now(), periods=num_candles, freq=self.timeframe)
        data = []
        
        for i, date in enumerate(dates):
            price = prices[i]
            high_low_range = price * daily_volatility * 2
            open_price = price - (high_low_range / 4) + (np.random.random() * high_low_range / 2)
            close_price = price - (high_low_range / 4) + (np.random.random() * high_low_range / 2)
            high_price = max(open_price, close_price) + (np.random.random() * high_low_range / 2)
            low_price = min(open_price, close_price) - (np.random.random() * high_low_range / 2)
            volume = np.random.randint(50000, 200000)
            
            data.append({
                'timestamp': date,
                'open': open_price,
                'high': high_price,
                'low': low_price,
                'close': close_price,
                'volume': volume
            })
        
        df = pd.DataFrame(data)
        
        logger.info(f"Retrieved {len(df)} candles of market data")
        return df
    
    def calculate_indicators(self, df):
        """
        Calculate technical indicators based on the market data.
        
        Args:
            df (pandas.DataFrame): Market data with OHLCV columns
            
        Returns:
            pandas.DataFrame: Market data with added indicator columns
        """
        logger.info("Calculating technical indicators")
        
        # Calculate moving averages
        df['fast_ma'] = df['close'].rolling(window=self.fast_ma).mean()
        df['slow_ma'] = df['close'].rolling(window=self.slow_ma).mean()
        
        # Calculate signals
        df['signal'] = 0
        df.loc[df['fast_ma'] > df['slow_ma'], 'signal'] = 1
        df.loc[df['fast_ma'] < df['slow_ma'], 'signal'] = -1
        
        # Detect crossovers
        df['signal_change'] = df['signal'].diff()
        
        # Drop NaN values resulting from the indicators calculation
        df = df.dropna()
        
        return df
    
    def generate_trading_signals(self, df):
        """
        Generate buy/sell trading signals based on the indicators.
        
        Args:
            df (pandas.DataFrame): Market data with indicator columns
            
        Returns:
            list: Trading signals
        """
        logger.info("Generating trading signals")
        
        signals = []
        latest_price = df['close'].iloc[-1]
        
        # Look for signal changes in the last candle
        latest_signal_change = df['signal_change'].iloc[-1]
        
        if latest_signal_change > 0:
            # Fast MA crossed above Slow MA: Buy Signal
            logger.info(f"BUY SIGNAL: {self.symbol} at ${latest_price:.2f}")
            signals.append({
                'type': 'buy',
                'symbol': self.symbol,
                'price': latest_price,
                'quantity': self.quantity,
                'timestamp': df.index[-1]
            })
            
            # Output signal in the format expected by the execution service
            print(f"SIGNAL:BUY:{self.symbol}:{self.quantity}@{latest_price:.2f}")
            
        elif latest_signal_change < 0:
            # Fast MA crossed below Slow MA: Sell Signal
            logger.info(f"SELL SIGNAL: {self.symbol} at ${latest_price:.2f}")
            signals.append({
                'type': 'sell',
                'symbol': self.symbol,
                'price': latest_price,
                'quantity': self.quantity,
                'timestamp': df.index[-1]
            })
            
            # Output signal in the format expected by the execution service
            print(f"SIGNAL:SELL:{self.symbol}:{self.quantity}@{latest_price:.2f}")
        
        # Check for stop loss on existing positions
        for pos_id, position in self.positions.items():
            if position['direction'] == 'long' and latest_price < position['entry_price'] * (1 - self.stop_loss_pct / 100):
                logger.info(f"STOP LOSS: Closing long position {pos_id} at ${latest_price:.2f}")
                signals.append({
                    'type': 'close',
                    'position_id': pos_id,
                    'symbol': self.symbol,
                    'price': latest_price,
                    'timestamp': df.index[-1]
                })
                
                # Output signal in the format expected by the execution service
                print(f"SIGNAL:CLOSE:{self.symbol}:{latest_price:.2f}")
                
            elif position['direction'] == 'short' and latest_price > position['entry_price'] * (1 + self.stop_loss_pct / 100):
                logger.info(f"STOP LOSS: Closing short position {pos_id} at ${latest_price:.2f}")
                signals.append({
                    'type': 'close',
                    'position_id': pos_id,
                    'symbol': self.symbol,
                    'price': latest_price,
                    'timestamp': df.index[-1]
                })
                
                # Output signal in the format expected by the execution service
                print(f"SIGNAL:CLOSE:{self.symbol}:{latest_price:.2f}")
        
        return signals
    
    def run(self):
        """
        Execute the strategy once.
        """
        try:
            # Fetch market data
            df = self.fetch_market_data()
            
            # Calculate indicators
            df = self.calculate_indicators(df)
            
            # Generate trading signals
            signals = self.generate_trading_signals(df)
            
            # Return the signals
            return signals
            
        except Exception as e:
            logger.error(f"Error running strategy: {e}")
            return []

def main():
    """
    Main entry point for the strategy script.
    """
    if len(sys.argv) != 2:
        logger.error("Usage: python ma_crossover_strategy.py <config_path>")
        sys.exit(1)
    
    config_path = sys.argv[1]
    
    # Initialize and run the strategy
    strategy = MovingAverageCrossoverStrategy(config_path)
    signals = strategy.run()
    
    # Log the signals
    for signal in signals:
        logger.info(f"Signal: {signal}")
    
    # Success
    logger.info("Strategy execution completed successfully")
    sys.exit(0)

if __name__ == "__main__":
    main()
