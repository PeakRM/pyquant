# pyquant
Event Driven Microservice Based Algorithmic Trading Engine.

Order Execution, Reconciliation and Dashboard backend are all built in Go. 
Strategies are (currently) designed to be written in Python, but "backend" service
can be extended to use other languages, including Go.

Broker_API service is currently test API written in Python (FastAPI) but the production version,
which has already been developed locally, behaves like an API to communicate with  various
brokers (IBKR, TDA, etc.) to allow systems to use one service to talk to any broker.

