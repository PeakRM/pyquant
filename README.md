# pyquant
Event Driven Microservice Based Algorithmic Trading Engine.

Order Execution, Reconciliation and Dashboard backend are all built in Go. 
Strategies are (currently) designed to be written in Python, but "backend" service
can be extended to use other languages, including Go.

Broker_API service is how the system communicates to different
brokers (IBKR, TDA, etc.) to centralize execution and reconciliation. 

## Roadmap
    - New Services
        - Add Risk Manager to manage overall portfolio risk (optional)
        - Integrate Databento data-feed (optional)
    - Backend
        - Improve execution logic (optional)
    - Broker API
        - Add logic to handle stops orders and improve fill monitoring (optional)
        - Add TD Ameritrade broker
        - Add functionality to enable option trading
    - Scehduler / Frontend
        - Fix scheduler to use cron for scheduling strategies
            - Add endpoint to schduler to allow strategies to send heartbeats to frontend
        - Add setup parameters to strategy configuration at setup-level 
        - Change strategy config to list each setup independently
            - Add strategy-group for grouping on frontend
        - Add KPIs for Buying Power, NLV, Cash @ broker
    
