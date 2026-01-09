# Airline Booking System

A high-performance airline booking system built in Go, implementing the system design for Indigo Airlines with direct domestic flights, caching, and distributed locking.

## Features

- **Flight Search**: Fast search with Redis caching (1hr TTL)
- **Booking System**: Distributed locking with Redis to prevent race conditions
- **Payment Processing**: Asynchronous payment flow with Kafka events
- **Data Consistency**: Optimistic locking for seat management
- **REST API**: Clean HTTP API with proper error handling

## Architecture

### Components
- **Search**: Redis cache + PostgreSQL fallback
- **Booking**: Distributed locking + async payment processing
- **Flight Management**: CRUD operations with Kafka events
- **Event Streaming**: Kafka for payment and seat update events

### Tech Stack
- **Backend**: Go (Gin/Gorilla Mux)
- **Database**: PostgreSQL
- **Cache**: Redis
- **Message Queue**: Kafka
- **Container**: Docker & Docker Compose

## Prerequisites

- Docker and Docker Compose
- Go 1.21+

## Quick Start

1. **Clone the repo:**
   ```bash
   git clone <repository>
   cd airline-booking-system
   ```
2. **Install dependencies:**
   ```bash
   go mod tidy
   ```
3. **Run the application without tracing (local dev):**
   - Start core infra (Postgres, Redis, Kafka, Zookeeper):
     ```bash
     docker-compose up -d postgres redis zookeeper kafka
     ```
   - Run the Go server:
     ```bash
     go run cmd/server/main.go
     ```

The API will be available at `http://localhost:8080`.

## API Endpoints

### Flight Search
```http
GET /api/v1/flights/search?source=Delhi&destination=Mumbai&date=2025-01-15
```

### Flight Management
```http
GET    /api/v1/flights/{id}
POST   /api/v1/flights
PUT    /api/v1/flights/{id}
```

### Booking System
```http
POST   /api/v1/bookings
GET    /api/v1/bookings/{id}
GET    /api/v1/users/{userId}/bookings
```

### Health Check
```http
GET /api/v1/health
```

## Sample API Usage

### Search Flights
```bash
curl "http://localhost:8080/api/v1/flights/search?source=Delhi&destination=Mumbai&date=2025-01-15"
```

### Create Booking
```bash
curl -X POST http://localhost:8080/api/v1/bookings \
  -H "Content-Type: application/json" \
  -d '{
    "flight_id": 1,
    "user_id": 123,
    "seats_booked": 2,
    "passenger_details": [
      {"name": "John Doe", "email": "john@example.com", "phone": "1234567890", "age": 30, "gender": "male"},
      {"name": "Jane Doe", "email": "jane@example.com", "phone": "0987654321", "age": 28, "gender": "female"}
    ]
  }'
```

### Create Flight
```bash
curl -X POST http://localhost:8080/api/v1/flights \
  -H "Content-Type: application/json" \
  -d '{
    "source": "Delhi",
    "destination": "Mumbai",
    "timestamp": "2025-01-20T10:00:00Z",
    "available_seats": 150,
    "total_seats": 180,
    "price": 2500.00
  }'
```

## Database Schema

### Flights Table
```sql
CREATE TABLE flights (
    id BIGSERIAL PRIMARY KEY,
    source VARCHAR(100) NOT NULL,
    destination VARCHAR(100) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    available_seats INTEGER NOT NULL,
    total_seats INTEGER NOT NULL,
    flight_status VARCHAR(50) DEFAULT 'scheduled',
    price DECIMAL(10,2) NOT NULL,
    version INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Bookings Table
```sql
CREATE TABLE bookings (
    id BIGSERIAL PRIMARY KEY,
    flight_id BIGINT REFERENCES flights(id),
    user_id BIGINT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    payment_reference_id VARCHAR(255),
    booking_price DECIMAL(10,2) NOT NULL,
    seats_booked INTEGER NOT NULL,
    booking_metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| SERVER_PORT | 8080 | HTTP server port |
| DB_HOST | localhost | PostgreSQL host |
| DB_PORT | 5432 | PostgreSQL port |
| DB_USER | postgres | PostgreSQL user |
| DB_PASSWORD | password | PostgreSQL password |
| DB_NAME | airline_booking | Database name |
| REDIS_HOST | localhost | Redis host |
| REDIS_PORT | 6379 | Redis port |
| KAFKA_BROKERS | localhost:9092 | Kafka brokers |
| CACHE_TTL | 1h | Cache TTL duration |
| LOCK_TTL | 5m | Lock TTL duration |

## Key Design Decisions

1. **Redis Caching**: Caching serach results in redis 
2. **Distributed Locking**: Redis-based locks with 5-minute TTL for booking
3. **Optimistic Locking**: Version-based concurrency control for seat updates
4. **Async Payment**: Event-driven payment processing with Kafka
5. **Data Consistency**: Update inventory only after successful payment

## Performance Characteristics

- **Search**: Sub-millisecond cache hits, ~50ms database queries
- **Booking**: ~100-200ms with locking and validation
- **Concurrent Safety**: Handles 1000+ concurrent bookings safely
- **Scalability**: Horizontal scaling with Redis/Kafka clustering

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o bin/server cmd/server/main.go
```

### Docker Development

- **Build server binary:**
  ```bash
  go build -o bin/server cmd/server/main.go
  ```

- **Run infra only (no tracing):**
  ```bash
  docker-compose up -d postgres redis zookeeper kafka
  ```

- **Run infra + tracing stack:**
  ```bash
  docker-compose up -d postgres redis zookeeper kafka otel-collector jaeger
  ```

## Monitoring & Observability

- Structured logging with request IDs
- Health checks for all dependencies
- Kafka event monitoring for business metrics
- Redis cache hit/miss ratios
 - Distributed tracing with OpenTelemetry (Jaeger via OTel Collector)
 - Metrics collection with Prometheus (API throughput, latency, error rate)

### Distributed Tracing

Tracing is optional and controlled via environment variables:

| Variable               | Default                         | Description |
|------------------------|---------------------------------|-------------|
| TRACING_ENABLED        | false                           | Enable distributed tracing when set to `true` |
| TRACING_SERVICE_NAME   | airline-booking-service         | Service name reported to the tracer backend |
| TRACING_ENDPOINT       | http://localhost:4318           | OTLP HTTP endpoint (e.g. OTel Collector) |
| TRACING_ENVIRONMENT    | local                           | Deployment environment (local, staging, prod, ...) |
| TRACING_SAMPLER_RATIO  | 1.0                             | Sampling ratio between 0.0 and 1.0 |

When `TRACING_ENABLED=true`, the server:

- Initializes an OpenTelemetry tracer provider with an OTLP HTTP exporter.
- Wraps the HTTP router with OpenTelemetry middleware so each API call becomes a span.
- Adds spans in core services (flight search and booking flows).

#### Example: Run with Jaeger via OTel Collector

This repository includes an OpenTelemetry Collector and Jaeger setup in `docker-compose.yml`.

1. Start infra + tracing stack:

```bash
docker-compose up -d postgres redis zookeeper kafka otel-collector jaeger
```

2. Run the application with tracing enabled:

```bash
export TRACING_ENABLED=true
export TRACING_SERVICE_NAME=airline-booking-service
export TRACING_ENDPOINT=http://localhost:4318
go run cmd/server/main.go
```

3. Generate some traffic (flight search, bookings, etc.), then open Jaeger UI:

- URL: `http://localhost:16686`
- Select the `airline-booking-service` service to explore traces.

### Metrics (Prometheus)

This project exposes **Prometheus metrics** at the `/metrics` endpoint on the same port as the API (`8080`).

Key HTTP metrics:

- `http_requests_total{method, path, code}`: total requests, labeled by method, path, and status text.
- `http_request_duration_seconds{method, path}`: histogram of request latency.
- `http_in_flight_requests`: current number of in-flight HTTP requests.

#### Run with Prometheus

1. Start infra + Prometheus (optionally with tracing stack as well):

```bash
docker-compose up -d postgres redis zookeeper kafka otel-collector jaeger prometheus
```

2. Run the application (with or without tracing as described above).

3. Open Prometheus UI:

- URL: `http://localhost:9090`
- Example queries:
  - `sum(rate(http_requests_total[1m])) by (method, path, code)` – request throughput & error rate.
  - `histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, path))` – 95th percentile latency per path.

## Future Enhancements

- [x] Rate limiting and API throttling
- [ ] Circuit breakers for external services
- [x] Distributed tracing (Jaeger)
- [x] Metrics collection (Prometheus)
- [ ] Waitlist system for sold-out flights
- [ ] Email/SMS notifications
- [ ] Multi-city booking support
