import grpc
import utils.trade_pb2 as trade_pb2
import utils.trade_pb2_grpc as trade_pb2_grpc
import time
from utils.definitions import Trade as TradeInstruction


def send_trade(trade: TradeInstruction) -> None:
    # Connect to the server
    # channel = grpc.insecure_channel('localhost:50051') # for local development
    # try:
    channel = grpc.insecure_channel('backend:50051') # for docker container  with service "backend"
    # except Exception:
        # channel = grpc.insecure_channel('localhost:50051') # for local development
    stub = trade_pb2_grpc.TradeServiceStub(channel)
    try:
        # Create a Trade message
        trade = trade_pb2.Trade(
            strategy_name=trade.strategy_name,
            contract_id=trade.contract_id,
            exchange=trade.exchange,
            symbol=trade.symbol,
            side=trade.side,
            quantity=str(trade.quantity), # Serialize as a string
            order_type=trade.order_type,  # Add order type (MKT, LMT)
            broker=trade.broker,          # Add broker (IB, TDA, etc.)
            price=str(trade.price) # Serialize as a string
        )

        # Send the Trade message
        response = stub.SendTrade(trade)
        print("Server response:", response.status)
    except Exception as e:
        print("Unable to send trade to backend: ", e)
        

if __name__ == "__main__":
    # while True:
        # time.sleep(30)
    send_trade(TradeInstruction(strategy_name='Test',
                                contract_id=691171673,
                                    exchange='CME',
                                    symbol='MES',
                                    side='BUY',
                                    quantity=1,
                                    order_type='LMT',
                                    price=6040.0,
                                    broker='IB_TEST'
                                    )
    )
