# Loan Application API

A RESTful API for loan application flow built with **Go**, **Chi router**, and **GORM**. Implements a two-step loan process: **Request** and **Approve**.

## Features

- **Request Loan** — Submit a new loan application with vehicle details
- **Approve Loan** — Approve a submitted loan application

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **Language** | Go 1.25+ |
| **HTTP Router** | [go-chi/chi/v5](https://github.com/go-chi/chi) |
| **ORM** | [GORM](https://gorm.io) + PostgreSQL driver |
| **Database** | PostgreSQL 14 |
| **Migration** | GORM AutoMigrate (development) / goose SQL (production) |
| **Config** | [envconfig](https://github.com/kelseyhightower/envconfig) (12-factor) |
| **Logging** | `log/slog` (structured JSON) |
| **Testing** | `testing` + testify + testcontainers-go |

## Prerequisites

- **Go 1.25+** (recommended 1.25+)
- **PostgreSQL 14+** (or Docker for containerized DB)

## Quick Start

### 1. Install Dependencies

#### Using Docker (recommended)

```bash
docker run -d \
  --name loan-postgres \
  -e POSTGRES_USER=user \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=loan_test \
  -p 5435:5432 \
  postgres:14-bullseye
```

> **Note:** When running via `go run ./cmd/api/main.go`, the app uses **GORM AutoMigrate** which creates the table automatically. The SQL migration file is provided for production deployments using `goose`.

### 2. Configure Environment

```bash
DATABASE_URL=postgresql://user:password@localhost:5435/loan_test
PORT=8086
LOG_LEVEL=info
```

### 3. Run the Application

```bash
go run ./cmd/api/main.go
```

Server starts at `http://localhost:8086` with JSON structured logging.

### 4. Test the Endpoints

```bash
# Health check
curl http://localhost:8086/health

# Request a loan
curl -X POST http://localhost:8086/api/loans/request \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "Bruce",
    "mrp": 100000000,
    "dp": 20000000,
    "vehicle_year": 2018,
    "police_number": "B 1234 BYE",
    "machine_number": "SDR72V25000W201"
  }'

# Approve a loan
curl -X POST http://localhost:8086/api/loans/approve \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "Bruce",
    "police_number": "B 1234 BYE"
  }'
```

## API Documentation

### Health Check

```
GET /health
```

**Response** `200 OK`:
```json
{
  "status": "ok"
}
```

---

### Request Loan

Creates a new loan application with status `submitted`.

```
POST /api/loans/request
```

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user_id` | string | ✅ | User identifier |
| `mrp` | integer | ✅ | Maximum Retail Price (must be > 0) |
| `dp` | integer | ✅ | Down Payment (must be > 0, cannot exceed MRP) |
| `vehicle_year` | integer | ✅ | Year of vehicle (1900–2050) |
| `police_number` | string | ✅ | Vehicle license plate (unique per user) |
| `machine_number` | string | ✅ | Vehicle engine number |

**Example Request:**
```json
{
  "user_id": "Bruce",
  "mrp": 100000000,
  "dp": 20000000,
  "vehicle_year": 2018,
  "police_number": "B 1234 BYE",
  "machine_number": "SDR72V25000W201"
}
```

**Success Response** `201 Created`:
```json
{
  "user_id": "Bruce",
  "loans": [
    {
      "mrp": 100000000,
      "dp": 20000000,
      "vehicle_year": 2018,
      "police_number": "B 1234 BYE",
      "machine_number": "SDR72V25000W201",
      "status": "submitted"
    }
  ]
}
```

**Error Response** `400 Bad Request` (validation):
```json
{
  "error": "validating request: user_id is required"
}
```

---

### Approve Loan

Updates a loan status from `submitted` to `approved`.

```
POST /api/loans/approve
```

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user_id` | string | ✅ | User identifier |
| `police_number` | string | ✅ | Vehicle license plate |

**Example Request:**
```json
{
  "user_id": "Bruce",
  "police_number": "B 1234 BYE"
}
```

**Success Response** `200 OK`:
```json
{
  "user_id": "Bruce",
  "police_number": "B 1234 BYE",
  "message": "Loan updated successfully."
}
```

**Error Response** `404 Not Found`:
```json
{
  "error": "loan_not_found",
  "error_description": "Loan not Found"
}
```

**Error Response** `400 Bad Request` (validation):
```json
{
  "error": "validating request: user_id is required"
}
```

## Configuration

All configuration is via environment variables (12-factor app):

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `DATABASE_URL` | — | ✅ | PostgreSQL connection string |
| `PORT` | `8086` | ❌ | HTTP server port |
| `LOG_LEVEL` | `info` | ❌ | Log level: `info` or `debug` |

## Running Tests

### Unit Tests

```bash
# All unit tests
go test -v ./internal/...

# Specific layer
go test -v ./internal/modules/loan/usecase/...
go test -v ./internal/modules/loan/handler/...
```
