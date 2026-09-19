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

However, before starting this QA phase, the `order-service` needs to be at least minimally implemented:

[x] Domain implemented
[x] Application implemented
[x] Ports defined
[x] Adapters implemented (Postgres, HTTP; Kafka publisher and the outbox relay now live in `shared/`, see below)
[ ] PostgreSQL operational (migrations exist, not yet run against a live database in this pass)
[ ] Migrations operational (not yet run against a live database in this pass)
[ ] HTTP operational (not yet exercised manually against a running server in this pass)
[ ] Docker operational (not yet built/run in this pass)
[ ] Main endpoints working via manual testing (not yet done in this pass)
[x] `go test ./...` executing successfully (verified 2026-09-18, all packages green)

Status as of 2026-09-18: code-complete and unit-tested (domain, application, config). Integration-level verification (real Postgres, real HTTP server, Docker image) has not been exercised in this pass — code correctness is confirmed, feature correctness against a running stack is not. Don't treat the checkboxes above as a substitute for that manual pass.

### Note — Shared Kernel (ADR-005)

`BaseEvent`/`DomainEvent`, the Kafka producer/publisher, and the outbox
repository moved out of `order-service` into `shared/` so the
`inventory-service` (and future bounded contexts) can reuse them
instead of duplicating the pattern. Their tests moved with them:

- `shared/events/base_event_test.go` (was `domain/event/base_event_test.go`)
- `shared/outbox` and `shared/messaging/kafka` currently have no dedicated tests — they're exercised indirectly through `application/command` tests via mocks. `FetchUnprocessed`/`MarkProcessed` are now integration-tested from the inventory-service side (`services/inventory-service/internal/infrastructure/persistence/postgres/integration_test.go`, `TestUnitOfWork`), since both bounded contexts share the same `shared/outbox.PostgresRepository`. Order-service itself has no integration test yet that exercises its own outbox table the same way — worth adding for parity when the `cdc-connector` starts depending on it.

### Note — `go test ./...` at the repo root does not work

`go.work` makes this a multi-module workspace; `go <cmd> ./...` only resolves from inside one of the member module directories (`shared/`, `services/order-service/`, `services/inventory-service/`), never from the repo root — `.github/workflows/go.yml` currently runs `go build -v ./...` / `go test -v ./...` from the root and fails for this reason, independent of any change to the code. `make test` (and the new `make test-integration`) now loop over each module explicitly and work correctly; the CI workflow still needs the same fix (not yet applied).

Once all requirements are met:

This will be the potential risk matrix:

R1 — Order may be created incorrectly
R2 — Order may be persisted incorrectly
R3 — API may accept invalid data
R4 — API may return an incorrect contract
R5 — Concurrency may lead to an inconsistent state
R6 — Internal errors may leak information
R7 — Future changes may break existing behavior