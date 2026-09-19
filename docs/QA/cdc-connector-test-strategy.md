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

However, before starting this QA phase, the `cdc-connector` needs to be at least minimally implemented:

[x] Core logic implemented (`internal/relay`: polling, publish, mark-processed, per-row and per-source failure isolation)
[x] Ports defined (`OutboxReader`, `Publisher` — small interfaces, no framework leakage into the relay logic)
[x] Adapters implemented (`shared/outbox.PostgresRepository`, `shared/messaging/kafka.Producer`, reused from the shared kernel — ADR-005)
[x] PostgreSQL operational (verified via `go test -tags=integration`, testcontainers)
[x] Kafka operational (verified via `go test -tags=integration`, testcontainers — **and** against the real `docker-compose.yml` broker, see below)
[x] Docker operational (built and run via `docker compose up`, multi-stage Dockerfile copying `shared/` + the service directory)
[x] Main flow working via manual testing (placed a real order through order-service's HTTP API, watched cdc-connector pick up the outbox row within one `POLL_INTERVAL` and publish it, consumed the message back from Kafka)
[x] `go test ./...` executing successfully (verified 2026-09-19, `internal/relay` unit tests green)

Status as of 2026-09-19: this is the first bounded context in the project verified end-to-end against the **real** `docker-compose.yml` stack, not just testcontainers — and that distinction is exactly what surfaced the two real bugs described below (see LEARNING-005). Both fixed and re-verified.

This is the potential risk matrix:

R1 — An event written to the outbox may never reach Kafka (the whole reason this service exists — see ADR-003, LEARNING-004)
R2 — A publish failure may cause an event to be marked processed anyway (silent data loss)
R3 — A crash mid-cycle may cause an event to be published twice (duplicate delivery)
R4 — A failure in one source (e.g. inventory database down) may block delivery for other sources
R5 — Kafka/Postgres misconfiguration in the runtime environment may go undetected because it's tested only against ephemeral containers, not the actual deployment stack (see LEARNING-005)
R6 — Internal errors may leak information (not applicable today — cdc-connector exposes no request-driven API beyond `/health`)
R7 — Future changes may break existing behavior

## Coverage so far

Unit tests (`internal/relay/relay_test.go`, fakes for `OutboxReader`/`Publisher`, no real I/O):

- `TestRelay/Test RunOnce`: publishes and marks on success; does nothing when there are no unprocessed rows; a publish failure leaves the row unmarked (mitigates R2); one row failing does not block the others in the same batch (mitigates R3's blast radius); a fetch failure surfaces as an error without panicking.
- `TestRelay/Test RunOnce with multiple sources`: a failure in one source (e.g. `inventory-service` unreachable) does not prevent the other source (`order-service`) from being processed in the same cycle (mitigates R4).

Integration tests (`internal/relay/integration_test.go`, build tag `integration`, real Postgres + real Kafka via testcontainers):

- `TestRelay_PublishesOutboxRowsToKafka`: an outbox row inserted directly (as the application would, inside a transaction — ADR-003) is fetched, published to `order-service.order.placed`, and marked processed; a real Kafka consumer then reads the message back and asserts the key (`AggregateID`) and value (raw payload) match exactly (mitigates R1 end-to-end, at the code level).

Manual verification against the real `docker-compose.yml` stack (2026-09-19):

- Placed an order via `POST /orders` on `order-service`; within one `POLL_INTERVAL` (2s), `cdc-connector` logged `relay: cycle published=1 err=<nil>` and the event appeared on `order-service.order.placed`, consumed directly via `kafka-console-consumer` inside the `orderflow-kafka` container.
- All 6 expected topics (`order-service.order.{placed,confirmed,shipped}`, `inventory-service.stock.{reserved,released,replenished}`) were auto-created and populated from outbox rows accumulated across earlier manual testing sessions.
- This manual pass is what surfaced R5 in practice: the `docker-compose.yml` Kafka broker was crash-looping (`RestartCount: 31`) due to a listener misconfiguration that predates this session, undetected because nothing had ever tried to actually publish to it before. Fixed (see ADR-004, ADR-006, LEARNING-005) and re-verified with the same manual pass, twice, after the fix.

## Open work

- R3 (exactly-once): delivery is at-least-once by design (ADR-003/004); no test simulates a crash between a successful Kafka publish and `MarkProcessed` to confirm the row would be retried (and thus duplicated) on the next cycle. Worth adding once a real consumer exists to observe the duplicate.
- R5: there is no automated check that catches `docker-compose.yml`/k8s configuration drift the way this manual pass did. Consider a CI job that brings up `docker-compose.yml` for real (not testcontainers) and asserts a message round-trips through it, specifically to guard against regressions like the listener bug.
- No test covers the relay running continuously via `Run` (only `RunOnce` is tested directly) — `Run`'s ticker/shutdown behavior on `ctx` cancellation is exercised only manually (`docker compose stop`).
