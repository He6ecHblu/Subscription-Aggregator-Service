# Subscription Aggregator Service

REST service for aggregating user online subscription data.

This project was created as a technical assessment. It provides CRUDL operations for subscription records and allows calculating the total subscription cost for a selected period with optional filtering by user ID and subscription service name.

---

## Features

* Create subscription records
* Get subscription by ID
* Update subscription
* Delete subscription
* List subscriptions
* Filter subscriptions by user ID and service name
* Calculate total subscription cost for a selected period
* PostgreSQL persistence
* Database migrations
* Structured logging
* Configuration through environment variables
* Swagger API documentation
* Docker Compose startup

---

## Tech Stack

| Area              | Technology                     |
| ----------------- | ------------------------------ |
| Language          | Go                             |
| HTTP router       | chi                            |
| Database          | PostgreSQL                     |
| PostgreSQL driver | pgx                            |
| Migrations        | golang-migrate                 |
| Configuration     | environment variables / `.env` |
| Logging           | slog                           |
| Validation        | go-playground/validator        |
| API documentation | Swagger / swaggo               |
| Containerization  | Docker, Docker Compose         |

---

## Architecture

The project follows a layered architecture:

```text
HTTP Handler
    ↓
Service Layer
    ↓
Repository Layer
    ↓
PostgreSQL
```

### Handler Layer

The handler layer is responsible for:

* receiving HTTP requests;
* parsing path, query, and body parameters;
* validating request format;
* returning JSON responses;
* converting service errors into HTTP responses.

### Service Layer

The service layer contains business logic:

* subscription data validation;
* period validation;
* date parsing logic;
* total cost calculation rules;
* coordination between handlers and repository.

### Repository Layer

The repository layer is responsible for database access:

* creating records;
* reading records;
* updating records;
* deleting records;
* listing records;
* selecting subscriptions that match filters and period boundaries.

This separation keeps HTTP logic, business logic, and SQL queries independent from each other.

---

## Project Structure

```text
subscription-aggregator-service/
├── cmd/
│   └── app/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── domain/
│   │   └── subscription.go
│   ├── handler/
│   │   └── subscription_handler.go
│   ├── service/
│   │   └── subscription_service.go
│   ├── repository/
│   │   └── subscription_repository.go
│   ├── transport/
│   │   └── router.go
│   └── logger/
│       └── logger.go
├── migrations/
│   ├── 000001_create_subscriptions.up.sql
│   └── 000001_create_subscriptions.down.sql
├── docs/
├── .env.example
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── LICENSE
├── README.md
└── DEVELOPMENT_PLAN.md
```

---

## Subscription Model

Each subscription contains:

| Field          | Type          | Description                                        |
| -------------- | ------------- | -------------------------------------------------- |
| `id`           | UUID          | Unique subscription record ID                      |
| `service_name` | string        | Name of the subscription service                   |
| `price`        | integer       | Monthly price in RUB                               |
| `user_id`      | UUID          | User identifier                                    |
| `start_date`   | string        | Subscription start date in `MM-YYYY` format        |
| `end_date`     | string / null | Optional subscription end date in `MM-YYYY` format |
| `created_at`   | timestamp     | Record creation time                               |
| `updated_at`   | timestamp     | Record update time                                 |

Example:

```json
{
  "id": "2b0a9d62-43ec-4ef5-8f7a-49ec0bdf7611",
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025",
  "end_date": "12-2025",
  "created_at": "2026-07-01T12:00:00Z",
  "updated_at": "2026-07-01T12:00:00Z"
}
```

---

## Date Format

The public API accepts dates in the following format:

```text
MM-YYYY
```

Example:

```text
07-2025
```

Internally, dates are stored as PostgreSQL `DATE` values.
For example, `07-2025` is stored as `2025-07-01`.

This makes filtering, comparison, and period calculations more reliable.

Date periods are inclusive. For example, the period from `07-2025` to `12-2025` contains six billable months.

---

## Total Cost Calculation Logic

The service calculates the total cost based on monthly subscription activity.

A subscription is counted for every month in which it is active within the selected period.
The calculation is performed in the service layer. The repository only returns subscriptions that can overlap the requested period.

Example:

```text
Subscription:
- service_name: Yandex Plus
- price: 400 RUB per month
- start_date: 07-2025
- end_date: 12-2025
```

Requested period:

```text
from=07-2025
to=12-2025
```

Calculation:

```text
400 * 6 months = 2400 RUB
```

Expected response:

```json
{
  "total": 2400,
  "currency": "RUB",
  "from": "07-2025",
  "to": "12-2025"
}
```

If `end_date` is not provided, the subscription is considered active until the end of the requested calculation period.

---

## API Endpoints

Base path:

```http
/api/v1
```

### Create Subscription

```http
POST /api/v1/subscriptions
```

Request body:

```json
{
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025",
  "end_date": "12-2025"
}
```

---

### Get Subscription by ID

```http
GET /api/v1/subscriptions/{id}
```

---

### Update Subscription

```http
PUT /api/v1/subscriptions/{id}
```

The `PUT` endpoint performs a full replacement of the subscription record.
All editable fields are required in the request body.

Request body:

```json
{
  "service_name": "Yandex Plus",
  "price": 500,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025",
  "end_date": "12-2025"
}
```

---

### Delete Subscription

```http
DELETE /api/v1/subscriptions/{id}
```

---

### List Subscriptions

```http
GET /api/v1/subscriptions
```

Supported query parameters:

| Parameter      | Description                    |
| -------------- | ------------------------------ |
| `user_id`      | Optional user ID filter        |
| `service_name` | Optional exact service name filter |
| `limit`        | Optional limit for pagination  |
| `offset`       | Optional offset for pagination |

Example:

```http
GET /api/v1/subscriptions?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex Plus&limit=20&offset=0
```

---

### Calculate Total Cost

```http
GET /api/v1/subscriptions/total
```

Supported query parameters:

| Parameter      | Required | Description                      |
| -------------- | -------: | -------------------------------- |
| `from`         |      yes | Period start in `MM-YYYY` format |
| `to`           |      yes | Period end in `MM-YYYY` format   |
| `user_id`      |       no | Optional user ID filter          |
| `service_name` |       no | Optional exact service name filter |

Example:

```http
GET /api/v1/subscriptions/total?from=07-2025&to=12-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex Plus
```

Response:

```json
{
  "total": 2400,
  "currency": "RUB",
  "from": "07-2025",
  "to": "12-2025"
}
```

---

### Health Check

```http
GET /health
```

Response:

```json
{
  "status": "ok"
}
```

---

## Configuration

The application is configured through environment variables.

Example `.env`:

```env
APP_PORT=8080
APP_ENV=local
LOG_LEVEL=info
DATABASE_URL=postgres://subscriptions_user:subscriptions_password@postgres:5432/subscriptions_db?sslmode=disable
```

The repository includes `.env.example` as a template.

The real `.env` file should not be committed.

---

## Database

The project uses PostgreSQL.

The main table is `subscriptions`.

Recommended schema:

```sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name TEXT NOT NULL,
    price INTEGER NOT NULL CHECK (price > 0),
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_service_name ON subscriptions(service_name);
CREATE INDEX idx_subscriptions_period ON subscriptions(start_date, end_date);
```

---

## Running the Project

### 1. Clone the Repository

```bash
git clone git@github.com:your-username/subscription-aggregator-service.git
cd subscription-aggregator-service
```

### 2. Create Environment File

```bash
cp .env.example .env
```

### 3. Start the Application

```bash
docker compose up --build
```

The API should be available at:

```http
http://localhost:8080
```

Swagger UI should be available at:

```http
http://localhost:8080/swagger/index.html
```

---

## Migrations

Database migrations are stored in the `migrations` directory.

Migrations are used to initialize and update the database schema in a reproducible way.

Example migration files:

```text
migrations/
├── 000001_create_subscriptions.up.sql
└── 000001_create_subscriptions.down.sql
```

For Docker Compose startup, migrations are applied through a separate migration service.
This keeps schema setup explicit and avoids mixing application startup with database migration logic.

For local development, migrations may also be applied manually with the migration CLI.

---

## Logging

The application uses structured logging.

The following events are logged:

* application startup;
* database connection;
* incoming HTTP requests;
* validation errors;
* database errors;
* subscription creation;
* subscription update;
* subscription deletion;
* total cost calculation.

Example log fields:

```text
level
time
message
request_id
method
path
status
duration
error
```

---

## Error Response Format

The API returns errors in a consistent JSON format.

Example:

```json
{
  "error": "validation_error",
  "message": "start_date must have MM-YYYY format"
}
```

Possible error types:

| Error              | Description             |
| ------------------ | ----------------------- |
| `validation_error` | Invalid request data    |
| `not_found`        | Resource was not found  |
| `bad_request`      | Invalid request format  |
| `internal_error`   | Unexpected server error |

---

## Validation Rules

The service validates:

* `service_name` is not empty;
* `price` is a positive integer;
* `user_id` is a valid UUID;
* `start_date` has `MM-YYYY` format;
* `end_date`, if provided, has `MM-YYYY` format;
* `end_date` is not earlier than `start_date`;
* calculation period `from` is not later than `to`.

---

## Design Decisions

### PostgreSQL `DATE` Instead of String Dates

Although the API accepts dates as `MM-YYYY`, the database stores them as `DATE`.

This allows reliable filtering, comparison, and period calculations.

---

### Layered Architecture

The project separates handlers, services, and repositories.

This makes the code easier to test, maintain, and extend.

---

### Explicit Total Calculation Rule

The task description leaves room for interpretation regarding total calculation.

This implementation treats subscription price as a monthly recurring cost and counts every active month inside the selected period.
The selected period is inclusive, so `from=07-2025` and `to=12-2025` covers July, August, September, October, November, and December.

The repository does not calculate the monetary total. It loads subscriptions that overlap the requested period, and the service layer calculates active months and the final sum.

---

### Full Update Semantics

The `PUT /api/v1/subscriptions/{id}` endpoint uses full replacement semantics.
Partial updates are intentionally not supported unless a separate `PATCH` endpoint is added later.

---

### Exact Service Name Filtering

The `service_name` filter uses exact matching.
Case-insensitive or partial matching is outside the initial scope and can be added later if needed.

---

### Separate Migration Service

Database migrations are run by a dedicated Docker Compose migration service.
The application expects the database schema to be ready when it starts.

---

### No User Existence Check

The service does not check whether a user exists.

User management is outside the scope of this service according to the task requirements.

---

### Integer Price

Subscription price is stored as an integer number of rubles.

Kopecks are not supported because the task explicitly states that subscription cost is always an integer value in rubles.

---

## License

This project is distributed under a custom non-commercial technical assessment license.

Commercial use is prohibited.

See `LICENSE` for details.

---

## Project Status

Technical assessment project.

The service is designed for demonstration, review, and evaluation purposes.
