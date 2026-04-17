# Ride Sharing Microservices Prototype

This repository is a full-stack ride sharing prototype built around small Go services, asynchronous messaging, and a thin Next.js client. It models the core rider journey end to end: preview a trip, choose a ride package, match with an available driver, create a Stripe Checkout session, and mark the trip as paid after the webhook is received.

The project is best understood as a working systems prototype rather than a finished production platform. The service boundaries are clear, the message flow is already meaningful, and the repository is organized in a way that makes it practical to extend.

## What the Project Does

- Lets a rider preview a route and estimate fares for multiple ride packages.
- Persists fare options and trips in MongoDB.
- Publishes trip lifecycle events through RabbitMQ.
- Registers drivers over WebSocket and matches them by package type.
- Creates Stripe Checkout sessions after driver assignment.
- Pushes rider and driver updates over WebSocket.
- Emits OpenTelemetry traces and ships them to Jaeger.

## High-Level Architecture

The system is split into four runtime pieces:

1. `api-gateway`
	- Public entry point for HTTP and WebSocket traffic.
	- Bridges the web client to gRPC services and RabbitMQ events.
	- Handles Stripe webhooks.

2. `trip-service`
	- Owns trip creation, fare generation, route handling, and trip persistence.
	- Stores trips and ride fares in MongoDB.
	- Publishes `trip.event.created` and reacts to driver and payment events.

3. `driver-service`
	- Maintains an in-memory pool of registered drivers.
	- Finds compatible drivers by selected ride package.
	- Emits driver acceptance or decline results back into the event flow.

4. `payment-service`
	- Consumes payment creation commands from RabbitMQ.
	- Creates Stripe Checkout sessions.
	- Publishes payment session events for the rider UI.

The frontend in `web/` is a Next.js application that acts as a simple rider and driver simulator on top of the gateway.

## Request and Event Flow

### Rider flow

1. The rider requests a preview through `POST /trip/preview`.
2. The API Gateway calls `TripService.PreviewTrip` over gRPC.
3. The Trip Service generates route data and fare options, then stores the fares.
4. The rider selects one fare and starts the trip through `POST /trip/start`.
5. The Trip Service creates the trip and publishes `trip.event.created`.
6. The Driver Service consumes that event, looks for a matching driver, and notifies one driver over WebSocket.
7. The driver accepts or declines the trip over WebSocket.
8. The Trip Service updates trip state and publishes either:
	- `trip.event.driver_assigned`, or
	- `trip.event.driver_not_interested` to retry matching.
9. After assignment, the Trip Service publishes `payment.cmd.create_session`.
10. The Payment Service creates a Stripe Checkout session and publishes `payment.event.session_created`.
11. The rider is redirected to Stripe Checkout.
12. Stripe sends a webhook to the API Gateway after successful payment.
13. The API Gateway publishes `payment.event.success`.
14. The Trip Service consumes that event and marks the trip as paid.

### Supporting docs

Architecture notes already live in:

- `docs/architecture/trip-creation-flow-v1.md`
- `docs/architecture/rabbitmq-flow-v1.md`

## Tech Stack

### Backend

- Go 1.25
- gRPC and Protocol Buffers
- RabbitMQ
- MongoDB
- OpenTelemetry + Jaeger
- Stripe Checkout

### Frontend

- Next.js 15
- React 19
- Tailwind CSS
- Leaflet and React Leaflet

### Local infrastructure

- Tilt for local orchestration
- Kubernetes manifests for development and production layouts
- Docker for container builds

## Service and Port Map

| Component | Role | Default Port |
| --- | --- | --- |
| API Gateway | HTTP API, WebSockets, Stripe webhook | `8081` |
| Trip Service | gRPC trip operations | `9083` |
| Driver Service | gRPC driver registration | `9092` |
| Payment Service | Event consumer for Stripe session creation | `9004` |
| Web | Next.js UI | `3000` |
| RabbitMQ | Broker | `5672` |
| RabbitMQ UI | Broker dashboard | `15672` |
| Jaeger UI | Tracing UI | `16686` |

## Repository Layout

```text
.
├── services/
│   ├── api-gateway/        # HTTP, WebSocket, Stripe webhook entry point
│   ├── driver-service/     # Driver registration and matching
│   ├── payment-service/    # Stripe session creation from queue messages
│   └── trip-service/       # Trip domain, fare calculation, MongoDB persistence
├── shared/
│   ├── contracts/          # Shared event and API contracts
│   ├── db/                 # MongoDB helpers
│   ├── messaging/          # RabbitMQ setup, queues, consumers
│   ├── proto/              # Generated protobuf code
│   └── tracing/            # HTTP, gRPC, and RabbitMQ tracing helpers
├── proto/                  # Source .proto definitions
├── infra/                  # Dockerfiles and Kubernetes manifests
├── docs/architecture/      # Event-flow and sequence diagrams
├── tools/                  # Helper utilities such as service scaffolding
└── web/                    # Next.js frontend
```

## RabbitMQ Topology

The project currently uses one main topic exchange:

- Exchange: `x.trip`

Key queues include:

- `q.find_available_drivers`
- `q.driver_cmd_trip`
- `q.driver_trip_response`
- `q.notify_driver_not_found`
- `q.notify_driver_assign`
- `q.payment_trip_response`
- `q.notify_payment_session_created`
- `q.notify_payment_success`
- `q.dead_letter`

Routing keys follow a predictable naming scheme:

- Trip events: `trip.event.*`
- Driver commands: `driver.cmd.*`
- Payment events: `payment.event.*`
- Payment commands: `payment.cmd.*`

## Storage Model

MongoDB is used only by the Trip Service today.

- Collection `ride_fares`
  - stores generated fare options for a user and route
- Collection `trips`
  - stores trip status, selected fare, and assigned driver

The driver registry is currently in memory inside the Driver Service.

## Development Workflow

### Recommended path: Tilt + Kubernetes

This repository is structured first around Tilt and Kubernetes manifests. That is the cleanest way to run the full stack locally.

#### Prerequisites

- Go 1.25+
- Node.js 20+
- Docker
- Kubernetes cluster available locally
- Tilt
- `protoc` with Go and gRPC code generation plugins if you plan to modify `.proto` files
- Stripe test keys if you want the payment flow to work end to end
- A MongoDB instance reachable from the Trip Service

#### Start the stack

1. Install backend and frontend dependencies:

	```bash
	go mod tidy
	cd web && npm install
	```

2. Ensure Kubernetes secrets exist.

	The repository includes development secrets for RabbitMQ and Stripe in `infra/development/k8s/secrets.yaml`. MongoDB is not provisioned in this repository, so you must create that secret yourself:

	```bash
	kubectl create secret generic mongodb \
	  --from-literal=uri='mongodb://<user>:<password>@<host>:<port>'
	```

3. Start Tilt:

	```bash
	tilt up
	```

4. Open the exposed services:

	- Web UI: `http://localhost:3000`
	- API Gateway: `http://localhost:8081`
	- Jaeger: `http://localhost:16686`
	- RabbitMQ management: `http://localhost:15672`

### Running pieces manually

If you are not using Tilt, the minimum runtime dependencies are:

- RabbitMQ
- MongoDB
- Stripe credentials
- the four services
- the Next.js web app

Representative commands:

```bash
go run ./services/api-gateway
go run ./services/driver-service
go run ./services/trip-service/cmd/main.go
go run ./services/payment-service/cmd/main.go
cd web && npm run dev
```

Manual startup is possible, but you will need to provide all required environment variables yourself.

## Environment Variables

The most important runtime variables are:

| Variable | Used By | Purpose |
| --- | --- | --- |
| `RABBITMQ_URI` | backend services | RabbitMQ connection string |
| `MONGODB_URI` | trip-service | MongoDB connection string |
| `JAEGER_ENDPOINT` | backend services | OpenTelemetry export endpoint |
| `TRIP_SERVICE_URL` | api-gateway | gRPC address for Trip Service |
| `DRIVER_SERVICE_URL` | api-gateway | gRPC address for Driver Service |
| `HTTP_ADDR` | api-gateway | HTTP bind address |
| `APP_URL` | payment-service | base URL for Stripe success and cancel redirects |
| `STRIPE_SECRET_KEY` | payment-service | Stripe secret key |
| `STRIPE_WEBHOOK_KEY` | api-gateway | Stripe webhook signing secret |
| `STRIPE_SUCCESS_URL` | payment-service | post-payment success redirect |
| `STRIPE_CANCEL_URL` | payment-service | post-payment cancellation redirect |
| `NEXT_PUBLIC_API_URL` | web | public HTTP API base URL |
| `NEXT_PUBLIC_WEBSOCKET_URL` | web | public WebSocket base URL |
| `NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY` | web | Stripe publishable key |

The frontend also includes a local `web/.env.local` with a Stripe test publishable key for development.

## Useful Commands

Generate protobuf code:

```bash
make generate-proto
```

Scaffold a new service:

```bash
make microservice name=<service-name>
```

## Summary

This codebase already demonstrates a realistic microservices workflow: synchronous HTTP and gRPC calls where they make sense, asynchronous messaging for cross-service coordination, a browser-based client, tracing, and an external payment provider. It is a solid foundation for learning, demos, and iterative system design work.
