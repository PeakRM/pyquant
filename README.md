# pyquant
Event Driven Microservice Based Algorithmic Trading Engine.

Order Execution, Reconciliation and Dashboard backend are all built in Go. 
Strategies are (currently) designed to be written in Python, but "backend" service
can be extended to use other languages, including Go.

Broker_API service is how the system communicates to different
brokers (IBKR, TDA, etc.) to centralize execution and reconciliation. 

## Roadmap
    - Add ib gateway to docker-compose
    - Update dockerfile for new frontend
    - Fix position and open pnl tracking on frontend
    - Fix chart on frontend
    - Fix logging in backend
    - Add order type to trade instruction
    - Add Risk Manager to manage overall portfolio risk
    - Integrate Databento data-feed
    - Improve limit order price/execution logic


