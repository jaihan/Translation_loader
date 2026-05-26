# WYZauto Translation Loader

A Go implementation skeleton for the WYZauto take-home exercise.

## How to run

### Prerequisites

- Go 1.20+
- Docker
- Docker Compose
- Make

### Start PostgreSQL

```bash
make up
```

Equivalent:

```bash
docker compose up -d
```

### Install dependencies

```bash
go mod tidy
```

### Run the app

```bash
make run
```

Equivalent:

```bash
go run ./cmd/app
```

## How to test

### Run all tests

```bash
make test
```

Equivalent:

```bash
go test ./... -v
```

### Run unit tests only

```bash
make unit
```

### Run integration tests

```bash
make integration
```

### Stop PostgreSQL

```bash
make down
```

## Project structure

```text
cmd/app/main.go
internal/domain
internal/repository
internal/cache
internal/loader
internal/service
tests/unit
tests/integration
migrations/schema.sql
```

## Design goals

- Avoid N+1 queries by bulk-loading translations.
- Hide PostgreSQL behind repository and loader interfaces.
- Support locale filtering.
- Use a thread-safe in-process TTL cache.
- Invalidate cache by entity ID, not by global flush.
- Treat missing translations as data gaps, not hard failures.

## Architecture analogy

The TranslationLoader is like a librarian with a cart.

Bad approach: ask the librarian to walk to the shelf once per product.

Good approach: give the librarian a full list and collect everything in one trip.

That is the difference between N+1 queries and bulk loading.

## Query strategy

```sql
SELECT entity_type, entity_id, locale, field_name, field_value, updated_at
FROM translation
WHERE entity_type = $1
  AND entity_id = ANY($2)
  AND locale = ANY($3)
ORDER BY entity_id, locale, field_name;
```

This loads all required translations for one entity type and a batch of IDs in one round-trip.

## Cache design

The cache is like a small kitchen fridge.

Frequently requested data stays close to the worker, but every item has an expiry label.

- Cache key includes entity type, entity ID, and requested locales.
- TTL prevents stale data from living forever.
- `Invalidate(entityID)` removes only that entity's cached translations.
- Cache access is protected by `sync.RWMutex`.

## Missing translation fallback

Missing translations should not panic or break indexing.

Fallback order:

```text
requested locale
↓
English
↓
empty string
```

## Sync consistency note

If translations are updated while a sync is running, a document could accidentally mix old and new values.

The safer production approach is to capture a sync snapshot timestamp at the start of the run and load translations with:

```sql
updated_at <= $sync_started_at
```

Analogy: printing a newspaper edition. Even if news changes during printing, one edition should be internally consistent.

## Delta sync design question

To support delta sync, I would use a cursor based on `(updated_at, id)` rather than only `updated_at`.

```sql
SELECT entity_type, entity_id, locale, field_name, field_value, updated_at, id
FROM translation
WHERE (updated_at, id) > ($1, $2)
ORDER BY updated_at, id
LIMIT $3;
```

Flow:

```text
Read last cursor
↓
Fetch changed translation rows
↓
Group by entity_type and entity_id
↓
Reload affected product documents
↓
Update cursor after successful indexing
```

Main trade-off:

- Benefit: much less database and Elasticsearch work.
- Cost: more cursor bookkeeping and retry complexity.

Using `(updated_at, id)` prevents missing rows when multiple rows share the same timestamp.

## What I would improve with more time

- Add request coalescing to avoid duplicate DB loads under concurrent sync.
- Add metrics for cache hit ratio, cache miss ratio, and load duration.
- Add structured logging for missing translation fallbacks.
- Consider Redis if multiple sync workers need shared cache state.
# wyzauto_submission
