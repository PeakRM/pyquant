from enum import Enum

from schwab.orders.common import Duration, Session, __BaseInstrument

class FuturesInstruction(Enum):
    '''
    Instructions for opening and closing equity positions.
    '''

    #: Open a long equity position
    BUY = 'BUY'

    #: Close a long equity position
    SELL = 'SELL'

    #: Open a short equity position
    SELL_SHORT = 'SELL_SHORT'

    #: Close a short equity position
    BUY_TO_COVER = 'BUY_TO_COVER'


class EquityInstrument(__BaseInstrument):
    '''Represents an equity when creating order legs.'''

    def __init__(self, symbol):
        super().__init__('FUTURES', symbol)

##########################################################################
# Buy orders


def future_buy_market(symbol, quantity):
    '''
    Returns a pre-filled :class:`~schwab.orders.generic.OrderBuilder` for an equity
    buy market order.
    '''
    from schwab.orders.common import OrderStrategyType, OrderType, Session
    from schwab.orders.generic import OrderBuilder

    return (OrderBuilder()
            .set_order_type(OrderType.MARKET)
            .set_session(Session.NORMAL)
            .set_duration(Duration.DAY)
            .set_order_strategy_type(OrderStrategyType.SINGLE)
            .add_equity_leg(FuturesInstruction.BUY, symbol, quantity))


def future_buy_limit(symbol, quantity, price):
    '''
    Returns a pre-filled :class:`~schwab.orders.generic.OrderBuilder` for an equity
    buy limit order.
    '''
    from schwab.orders.common import OrderStrategyType, OrderType, Session
    from schwab.orders.generic import OrderBuilder

    return (OrderBuilder()
            .set_order_type(OrderType.LIMIT)
            .set_price(price)
            .set_session(Session.NORMAL)
            .set_duration(Duration.DAY)
            .set_order_strategy_type(OrderStrategyType.SINGLE)
            .add_equity_leg(FuturesInstruction.BUY, symbol, quantity))

def future_sell_market(symbol, quantity):
    '''
    Returns a pre-filled :class:`~schwab.orders.generic.OrderBuilder` for an equity
    buy market order.
    '''
    from schwab.orders.common import OrderStrategyType, OrderType, Session
    from schwab.orders.generic import OrderBuilder

    return (OrderBuilder()
            .set_order_type(OrderType.MARKET)
            .set_session(Session.NORMAL)
            .set_duration(Duration.DAY)
            .set_order_strategy_type(OrderStrategyType.SINGLE)
            .add_equity_leg(FuturesInstruction.SELL, symbol, quantity))


def future_sell_limit(symbol, quantity, price):
    '''
    Returns a pre-filled :class:`~schwab.orders.generic.OrderBuilder` for an equity
    buy limit order.
    '''
    from schwab.orders.common import OrderStrategyType, OrderType, Session
    from schwab.orders.generic import OrderBuilder

    return (OrderBuilder()
            .set_order_type(OrderType.LIMIT)
            .set_price(price)
            .set_session(Session.NORMAL)
            .set_duration(Duration.DAY)
            .set_order_strategy_type(OrderStrategyType.SINGLE)
            .add_equity_leg(FuturesInstruction.SELL, symbol, quantity))
