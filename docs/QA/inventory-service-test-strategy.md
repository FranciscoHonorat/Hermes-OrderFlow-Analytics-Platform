We will develop a comprehensive suite of tests to fully explore every part of the project and every possible aspect, demonstrating that the service functions according to the domain and adheres to its contracts, handles known failures, and relies on an automated foundation that prevents regressions.

Following this logic of requirements:

Requirement
↓
Risk
↓
Test
↓
Evidence
↓
Result
↓
Decision

However, before starting this QA phase, the `inventory-service` needs to be at least minimally implemented:

[x] Domain implemented (`StockItem` aggregate: `Reserve`, `Release`, `Replenish`, `SKU` value object, domain events)
[x] Application implemented (`ReserveStock`, `ReleaseStock`, `ReplenishStock` command handlers)
[x] Ports defined (`input`/`output` under `internal/application/port`)
[x] Adapters implemented (Postgres repository/queries/unit of work, HTTP handlers)
[x] PostgreSQL operational (verified 2026-09-18 via `go test -tags=integration` against a real Postgres in a testcontainers container — not the `docker-compose` stack itself, see note below)
[x] Migrations operational (verified 2026-09-18 — the integration suite applies `migrations/*.up.sql` via `golang-migrate`, per ADR-001, before every run)
[ ] HTTP operational (handlers are unit-tested with mocked use cases; not yet exercised manually against a running server bound to a real port)
[ ] Docker operational (Dockerfile exists, not yet built/run in this pass)
[ ] Main endpoints working via manual testing (not yet done in this pass)
[x] `go test ./...` executing successfully (verified 2026-09-18, all packages green; run via `make test`, which loops over every workspace module — running it from the repo root directly fails, see note in order-service's strategy doc)

Status as of 2026-09-18: code-complete, unit-tested (domain aggregate, application command handlers, HTTP handlers, config) and integration-tested against a real Postgres (repository, queries, unit of work/outbox atomicity). Still open: a manual smoke test against the actual `docker-compose` stack (not an ephemeral test container), the Docker image build, and the order-service ⇄ inventory-service saga — that last one is Fase 3 of the roadmap, once the `cdc-connector` relay exists to actually move events between services.

### Note — two test suites, two commands

- `make test` / `go test ./...` (per module): unit tests only — domain, application (mocked ports), HTTP handlers (mocked use cases), config. Fast, no external dependencies, safe to run on every commit.
- `make test-integration` / `go test -tags=integration ./...` (per module): spins up a real Postgres via testcontainers-go, applies migrations, and exercises the Postgres adapters end to end. Requires Docker. Not wired into `.github/workflows/go.yml` yet — CI currently only runs the unit suite (and even that needs the per-module loop fix in `make test`/the workflow, see the note in order-service's strategy doc).

Once all requirements above are met:

This will be the potential risk matrix:

R1 — A stock reservation may exceed available quantity (oversell)
R2 — Stock state may be persisted incorrectly (available/reserved out of sync)
R3 — API may accept invalid data (negative quantity, empty SKU)
R4 — API may return an incorrect contract
R5 — Concurrent reservations on the same SKU may lead to an inconsistent state (no optimistic/pessimistic locking implemented yet — `Save` is a plain upsert)
R6 — Internal errors may leak information
R7 — Future changes may break existing behavior
R8 — A domain event written to the outbox may never reach Kafka if the `cdc-connector` relay is down or missing (see ADR-003, LEARNING-004)

## Coverage so far

Unit tests:

- `internal/domain/valueobject/sku_test.go`: `NewSKU` validation (including whitespace trimming), `Equal`, `IsZero`, JSON marshal/unmarshal.
- `internal/domain/event/event_test.go`: all four domain event constructors (`StockReserved`, `StockReleased`, `StockReplenished`, `LowStockDetected`).
- `internal/domain/stock/stockItem_test.go`: constructor validation, `Reserve`/`Release`/`Replenish` happy and failure paths, event emission (including `LowStockDetected`), and that a failed `Reserve` leaves state and events untouched (mitigates R1, R2).
- `internal/application/command/stock_commands_test.go`: command handlers against mocked repository/outbox/unit of work — verifies events are drained to the outbox only on success, and that no partial state is saved on failure (mitigates R1, R2, R8's write side).
- `internal/infrastructure/http/handler/stock_handler_test.go`: HTTP handlers against mocked use cases/queries — malformed body → 400, use case error → 422, not found → 404, query error → 500 (mitigates R3, R4).
- `internal/config/env_test.go`: `getEnv`/`getRequiredEnv` edge cases (missing, empty, whitespace-only values).

Integration tests (build tag `integration`, real Postgres via testcontainers):

- `internal/infrastructure/persistence/postgres/integration_test.go`:
  - `TestStockRepository`: save + find round trip, update via the `ON CONFLICT DO UPDATE` branch, and the not-found error path (mitigates R2).
  - `TestStockQueries`: `GetStockBySKU` and `ListLowStock` against real rows.
  - `TestUnitOfWork`: the full Transactional Outbox path — stock item and domain event committed atomically on success (then drained via `shared/outbox.FetchUnprocessed`/`MarkProcessed`, closing the gap flagged in the order-service strategy doc), and both left unpersisted when the use case returns an error after saving (proves the rollback, not just an application-level no-op). Directly validates ADR-003, ADR-005 and LEARNING-003/004.

## Open work

- R5 (concurrency): `StockRepository.Save` does a plain `ON CONFLICT DO UPDATE`, with no row-level locking or optimistic concurrency check (e.g. a version column). Two concurrent `Reserve` calls on the same SKU can both read the same `available` value before either writes — a lost-update race, not covered by the current integration suite (it doesn't run concurrent requests). Worth revisiting once load testing is in scope.
- R6: no test asserts that error responses avoid leaking internal details (today `c.JSON(..., gin.H{"error": err.Error()})` forwards the raw Go error string to the client).
- R8: `FetchUnprocessed`/`MarkProcessed` are now integration-tested directly, but the actual relay loop (polling + publish + mark) only exists once the `cdc-connector` (Fase 2 of the roadmap) is implemented.
