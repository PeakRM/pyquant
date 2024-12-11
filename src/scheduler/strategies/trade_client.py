import grpc
import trade_pb2
import trade_pb2_grpc
import time
from definitions import Trade as TradeInstruction


def send_trade(trade: TradeInstruction) -> None:
    # Connect to the server
    channel = grpc.insecure_channel('localhost:50051')
    stub = trade_pb2_grpc.TradeServiceStub(channel)

    # Create a Trade message
    trade = trade_pb2.Trade(
        strategy_name=trade.strategy_name,
        contract_id=trade.contract_id,
        exchange=trade.exchange,
        symbol=trade.symbol,
        side=trade.side,
        quantity=str(trade.quantity) # Serialize as a string
    )

    # Send the Trade message
    response = stub.SendTrade(trade)
    print("Server response:", response.status)

if __name__ == "__main__":
    while True:
        time.sleep(5)
        send_trade()
