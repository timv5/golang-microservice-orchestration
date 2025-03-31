# golang-microservice-orchestration
## Purpose
Project is a simple example of how to implement orchestration in microservice environment.

## Description
This project contains 2 standalone microservices (order-service & payment-service). Order service contains rest api
and makes a call to payment service. Transaction is commited on both sides, meaning someone needs to orchestrate a rollback,
and for this is responsible an orchestration-service.

## Order Service
Is a main service responsible for processing orders. It orchestrates order creation, payment processing,
and rollbacks in case of errors using a combination of Redis, and RabbitMQ.
Built in Go, the service is designed with a clean architecture that separates configuration, business logic, data persistence,
and message handling.

## Technologies Used
- Go (Golang)
- Redis - For caching and orchestration state
- RabbitMQ - For event handling and orchestration
- Postgres
- Docker

## Getting Started
Prerequisites
- Go 1.17+ installed
- A running PostgreSQL database
- A running Redis instance
- A running RabbitMQ instance

Project contains docker-compose file which can be run with: docker-compose -f docker-compose.yml up -d
This will run: redis, postgres, RabbitMQ.
