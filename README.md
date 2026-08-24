# Spndex

**Spndex is a work-in-progress expense tracking and analytics application built with Go.**

The backend provides a versioned REST API for authentication, users, expenses, and analytics. A frontend application is currently in development.

The project focuses on building a maintainable backend with clear separation of concerns, authentication, testing against real infrastructure, and observability.

> **🚧 Work in progress — the application is not yet complete still learning.**

---

![alt text](login.png)

## Architecture

Spndex uses a layered architecture designed to keep HTTP handling, business logic, and infrastructure concerns separate.

```text
                         HTTP Request
                              │
                              ▼
                         ┌────────┐
                         │ Router │
                         └────┬───┘
                              │
                              ▼
                        ┌───────────┐
                        │Middleware │
                        └─────┬─────┘
                              │
                              ▼
                         ┌─────────┐
                         │Handler  │
                         └────┬────┘
                              │
                              ▼
                         ┌─────────┐
                         │ Service │
                         └────┬────┘
                              │
                              ▼
                      ┌──────────────┐
                      │  Repository  │
                      └──────┬───────┘
                             │
                    ┌────────┴────────┐
                    ▼                 ▼
               PostgreSQL           Redis
```

### Handlers

Handlers are responsible for HTTP concerns such as:

* Request binding
* Validation
* Extracting request context
* Calling services
* Building HTTP responses

Business logic is kept out of the handler layer.

### Services

Services contain application and business logic.

They depend on repository interfaces rather than concrete database implementations.

This allows the same business logic to be tested using fake repositories without requiring a database.

### Repositories

Repositories handle infrastructure-specific operations.

PostgreSQL is used for persistent application data and Redis is used for authentication sessions.

This keeps infrastructure concerns isolated from the service layer.

---

## Authentication

Authentication uses JWTs with Redis-backed sessions.

```text
Login
  │
  ├── Validate credentials
  │
  ├── Create Redis session
  │
  └── Generate JWT containing session ID
```

Authenticated requests use the JWT to identify the session.

Logging out removes the corresponding Redis session:

```text
Logout
   │
   ▼
Delete Redis Session
   │
   ▼
Session Invalidated
```

This provides server-side session invalidation instead of relying exclusively on JWT expiration.

---

## Testing

Testing is split between unit tests and integration tests.

### Service Tests

Service tests use fake repository implementations.

```text
Service
   │
   ▼
Fake Repository
   │
   ▼
Unit Test
```

This allows business logic to be tested without PostgreSQL or Redis.

Tests cover areas such as:

* Business rules
* Validation
* Error handling
* Authentication
* Session behavior
* Partial updates

### Repository Tests

Repository tests use real PostgreSQL and Redis instances through Testcontainers.

```text
Repository Test
      │
      ▼
 Testcontainers
    ┌─┴─┐
    ▼   ▼
Postgres Redis
```

This allows SQL and Redis behavior to be tested against the actual infrastructure rather than mocked implementations.

Repository tests cover:

* CRUD operations
* Filtering
* Pagination
* Ownership checks
* Soft deletion
* Analytics queries
* Redis session operations
* Serialization
* Session expiration

---

## Observability

Observability is built into the backend rather than added as an afterthought.

The application uses:

* **Structured logging** with Zerolog
* **Request IDs**
* **Distributed tracing**
* **OpenTelemetry**
* **New Relic**

A request can be followed through the application:

```text
HTTP Request
     ↓
Middleware
     ↓
Handler
     ↓
Service
     ↓
Repository
     ↓
PostgreSQL / Redis
```

This makes it easier to trace requests and diagnose failures across application layers.

---

## Database Design

PostgreSQL is the primary application database.

The application uses:

* UUID identifiers
* Foreign keys
* Database constraints
* Soft deletion
* Created/updated timestamps
* Parameterized queries
* Database migrations

Expense amounts are stored as integers representing the smallest unit.

For example:

```text
34.99 → 3499
```

This avoids floating-point precision issues when calculating expense totals and averages.

---

## API

The API is versioned under:

```text
/api/v1
```

### Authentication

```text
POST /auth/register
POST /auth/login
POST /auth/logout
```

### Users

```text
GET    /users
GET    /users/:id
PATCH  /users/:id
DELETE /users/:id
```

### Expenses

```text
POST   /expenses
GET    /expenses
GET    /expenses/:id
PATCH  /expenses/:id
DELETE /expenses/:id
```

### Analytics

```text
GET /expenses/category-totals
GET /expenses/monthly-expenses
GET /expenses/dashboard
GET /expenses/spending-trends
```

The API is documented with OpenAPI and includes interactive documentation for exploring endpoints and request/response schemas.

---

## Project Structure

```text
apps/backend/
├── internal/
│   ├── auth/
│   ├── database/
│   ├── handlers/
│   ├── middleware/
│   ├── model/
│   ├── repositories/
│   ├── router/
│   ├── service/
│   ├── tests/
│   └── validation/
│
├── static/
│   └── openapi.json
│
└── main.go
```

The structure separates application responsibilities instead of combining routing, business logic, and infrastructure code.

---

## Engineering Decisions

Some of the main design decisions in the project are:

**Interface-driven services**

Services depend on repository interfaces, making business logic easier to test and keeping infrastructure details isolated.

**Redis-backed sessions**

JWT authentication is combined with server-side sessions so sessions can be invalidated before token expiration.

**Integration testing**

Repository tests use real PostgreSQL and Redis containers instead of mocking the underlying infrastructure.

**Layered architecture**

Handlers, services, and repositories have distinct responsibilities, making the application easier to reason about and change.

**Observability**

Logging, request IDs, tracing, and monitoring are integrated into the application architecture.

**Integer expense amounts**

Expenses are represented using integer amounts rather than floating-point values to avoid precision issues during calculations.

---

## Current Status

Spndex is actively being developed.

### Implemented

* [x] Versioned REST API
* [x] User management
* [x] Expense management
* [x] Expense analytics
* [x] JWT authentication
* [x] Redis-backed sessions
* [x] Session invalidation on logout
* [x] Request validation
* [x] Soft deletion
* [x] OpenAPI documentation
* [x] Service unit tests
* [x] PostgreSQL integration tests
* [x] Redis integration tests
* [x] Structured logging
* [x] Request IDs
* [x] Distributed tracing
* [x] New Relic monitoring

### In Progress

* [ ] Frontend application
* [ ] Additional backend features
* [ ] Additional test coverage
* [ ] Rate limiting
* [ ] Production deployment

---

## Why I Built It

Spndex started as an expense tracking application but has evolved into a project focused on backend engineering and system design.

The project is being used to explore and apply concepts including:

* REST API design
* Layered architecture
* Dependency inversion
* Interface-driven design
* Authentication and session management
* PostgreSQL
* Redis
* Integration testing
* Testcontainers
* Observability
* Distributed tracing

The system is still evolving as new features and architectural requirements are introduced.
