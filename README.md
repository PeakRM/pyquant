# pyquant
Event Driven Microservice Based Algorithmic Trading Engine.

Order Execution, Reconciliation and Dashboard backend are all built in Go. 
Strategies are (currently) designed to be written in Python, but "backend" service
can be extended to use other languages, including Go.

Broker_API service is how the system communicates to different
brokers (IBKR, TDA, etc.) to centralize execution and reconciliation. 

## Roadmap
    - Add ib gateway to docker-compose - DONE
    - Update dockerfile for new frontend - done
    - Fix position and open pnl tracking on frontend - DONE
    - Fix chart on frontend - DONE
    - Fix logging in backend
    - Add order type to trade instruction - ADDED TO NEW FEATURES BRANCH
    - Add price to trade instruction
    - Add Close position button to front end - ADDED TO NEW FEATURES BRANCH
    - Add more meaningful stats to dashboard
        - Positions, Order lists
        - Buying Power, NLV, Cash @ broker
    - Add database to store strategy-config, positions, trade instructions generated, orders and fills
    - Add Risk Manager to manage overall portfolio risk
    - Integrate Databento data-feed
    - Improve limit order price/execution logic


