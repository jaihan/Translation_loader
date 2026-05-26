# WYZauto Translation Loader

A production-oriented Translation Loader implementation designed to efficiently load and cache localized product translations while avoiding common performance pitfalls such as N+1 queries.

The solution emphasizes:

- Bulk translation loading
- Interface-driven architecture
- Thread-safe caching
- Entity-scoped cache invalidation
- Comprehensive testing
- Clean separation of concerns
- Production-ready design principles

---

# Objective

Given a list of products and a requested locale, efficiently load translations while:

- Avoiding N+1 database queries
- Supporting cache-based lookups
- Providing graceful fallback behavior
- Remaining testable and maintainable

---

# High-Level Architecture

```text
                       ┌─────────────────────┐
                       │     Application     │
                       └──────────┬──────────┘
                                  │
                                  ▼
                       ┌─────────────────────┐
                       │ Translation Service │
                       └──────────┬──────────┘
                                  │
                                  ▼
                       ┌─────────────────────┐
                       │ Translation Loader  │
                       └──────────┬──────────┘
                                  │
               ┌──────────────────┴──────────────────┐
               │                                     │
               ▼                                     ▼
     ┌───────────────────┐                ┌───────────────────┐
     │     TTL Cache     │                │ Translation Repo  │
     └─────────┬─────────┘                └─────────┬─────────┘
               │                                    │
               │ Cache Miss                         │
               ▼                                    ▼
        ┌──────────────────────────────────────────────────┐
        │                    PostgreSQL                    │
        └──────────────────────────────────────────────────┘
```

---

# Project Structure

```text
cmd            -> application startup
domain         -> business models
repository     -> database access
cache          -> caching infrastructure
loader         -> translation abstraction
service        -> business use case
tests/unit     -> logic verification
tests/integration -> DB verification
migrations     -> schema definition
```

---

# Design Decisions

## 1. Bulk Translation Loading

### Problem

A naïve implementation loads translations individually:

```text
Load Product 1
→ Query Translations

Load Product 2
→ Query Translations

Load Product 3
→ Query Translations
```

For N products:

```text
1 product query
N translation queries
```

Result:

```text
N+1 Query Problem
```

---

### Solution

Translations are loaded in bulk:

```sql
SELECT *
FROM translations
WHERE entity_type = $1
AND entity_id = ANY($2)
```

This reduces:

```text
1 product query
1 translation query
```

Benefits:

- Predictable performance
- Lower database load
- Better scalability

---

## 2. Interface-Based Design

Business logic never depends directly on PostgreSQL.

```go
type TranslationRepository interface {
    LoadTranslations(
        ctx context.Context,
        entityType string,
        entityIDs []string,
        locale string,
    ) ([]domain.Translation, error)
}
```

Benefits:

- Easy unit testing
- Mock-friendly
- Future database replacement

Example:

```text
Loader
  ↓
Repository Interface
  ↓
Postgres Repository
```

---

## 3. Thread-Safe TTL Cache

Translations are cached using:

```text
(entity_type, entity_id, locale)
```

as the cache key.

Implementation uses:

```go
sync.RWMutex
```

to guarantee safe concurrent access.

Benefits:

- Reduced database load
- Improved response time
- Safe for concurrent goroutines

---

## 4. Entity-Level Invalidation

Instead of flushing the entire cache:

```text
BAD

Cache.Clear()
```

only affected entities are invalidated:

```text
GOOD

Invalidate(Product-123)
```

Benefits:

- Higher cache hit ratio
- Less database traffic
- Better scalability

---

# Query Strategy

The implementation intentionally avoids N+1 queries.

Example:

```text
100 Products
```

Naïve:

```text
1 Product Query
100 Translation Queries

Total = 101 Queries
```

Current Solution:

```text
1 Product Query
1 Translation Query

Total = 2 Queries
```

Benefits:

- Reduced latency
- Lower PostgreSQL load
- Better throughput

---

# Cache Strategy

## Cache Key

```text
product:123:en
product:123:de
product:123:fr
```

---

## TTL

Default:

```text
5 minutes
```

Configurable based on workload requirements.

Trade-offs:

| TTL   | Advantage          | Disadvantage    |
| ----- | ------------------ | --------------- |
| Short | Fresh data         | More DB traffic |
| Long  | Better performance | Stale data risk |

---

# Missing Translation Handling

The system handles missing translations gracefully.

Example:

```text
Requested Locale: de

Translation Exists:
✓ Return Translation

Translation Missing:
✓ Return Empty Translation
✓ Continue Processing
```

No panic.

No partial failure.

No broken document generation.

---

# Error Handling

Errors are wrapped with contextual information.

Example:

```go
return fmt.Errorf(
    "load translations for entity %s: %w",
    entityID,
    err,
)
```

Produces:

```text
load translations for entity 123:
query translations:
connection timeout
```

Benefits:

- Easier debugging
- Better observability
- Faster incident resolution

---

# Testing Strategy

## Unit Tests

Purpose:

Validate business logic in isolation.

Characteristics:

- Repository mocked
- Cache mocked
- No database dependency

Run:

```bash
make unit
```

---

## Integration Tests

Purpose:

Validate repository and loader behaviour against PostgreSQL.

Characteristics:

- Real PostgreSQL instance
- Real SQL execution
- Schema applied automatically

Run:

```bash
make integration
```

---

# Running the Project

## Prerequisites

- Go 1.24+
- Docker
- Docker Compose

Verify:

```bash
go version
docker version
docker compose version
```

---

# Setup

Clone repository:

```bash
git clone <repository-url>
cd wyzauto-translation-loader
```

Install dependencies:

```bash
go mod tidy
```

---

# Start PostgreSQL

```bash
make up
```

Verify:

```bash
docker compose ps
```

Expected:

```text
postgres   running
```

---

# Apply Database Schema

```bash
docker compose exec postgres psql \
  -U wyzauto \
  -d wyzauto \
  -f /docker-entrypoint-initdb.d/schema.sql
```

Or automatically during container startup.

---

# Run Application

```bash
make run
```

---

# Run Unit Tests

```bash
make unit
```

Expected:

```text
PASS
```

---

# Run Integration Tests

```bash
make integration
```

Expected:

```text
PASS
```

---

# Run All Tests

```bash
make test
```

---

# Coverage Report

```bash
make coverage
```

Example:

```text
coverage: 90.3% of statements
```

---

# Makefile Commands

| Command          | Description           |
| ---------------- | --------------------- |
| make up          | Start PostgreSQL      |
| make down        | Stop PostgreSQL       |
| make restart     | Restart PostgreSQL    |
| make logs        | View PostgreSQL logs  |
| make run         | Run application       |
| make test        | Run all tests         |
| make unit        | Run unit tests        |
| make integration | Run integration tests |
| make coverage    | Generate coverage     |
| make fmt         | Format code           |
| make vet         | Run go vet            |
| make clean       | Clean test cache      |

---

# Production Considerations

If this system were deployed to production, I would additionally consider:

### Distributed Cache

Replace in-memory cache with:

- Redis
- KeyDB

Benefits:

- Shared cache across instances
- Better horizontal scaling

---

### Observability

Add:

- Structured logging
- Prometheus metrics
- OpenTelemetry tracing

Metrics:

```text
cache_hit_total
cache_miss_total
translation_load_duration
translation_query_duration
```

---

### Connection Pooling

Tune:

```go
SetMaxOpenConns()
SetMaxIdleConns()
SetConnMaxLifetime()
```

to match workload characteristics.

---

### Cache Stampede Protection

Prevent multiple requests from loading the same translations simultaneously.

Possible solutions:

- Singleflight
- Request coalescing

---

### Delta Synchronization

Instead of rebuilding everything:

```text
Load only changed entities
```

Benefits:

- Reduced processing time
- Lower database load

---

# Design Trade-Offs

| Decision              | Benefit                 | Trade-Off           |
| --------------------- | ----------------------- | ------------------- |
| Bulk loading          | Avoids N+1              | More memory usage   |
| TTL cache             | Faster reads            | Potential staleness |
| Interface abstraction | Testability             | Slight complexity   |
| Entity invalidation   | Better cache efficiency | More bookkeeping    |
| In-memory cache       | Simplicity              | Not distributed     |

---

# Evaluation Rubric Mapping

| Requirement      | Solution                    |
| ---------------- | --------------------------- |
| Query efficiency | Bulk translation loading    |
| Interface design | Repository abstraction      |
| Cache design     | TTL + entity invalidation   |
| Error handling   | Contextual wrapped errors   |
| Test quality     | Unit + integration tests    |
| Readability      | Clear package separation    |
| Design question  | Trade-off analysis included |

---

# Conclusion

This implementation prioritizes:

- Correctness
- Performance
- Testability
- Maintainability

while remaining intentionally simple and easy to evolve for future production requirements.
