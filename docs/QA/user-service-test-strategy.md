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

However, before starting this QA phase, the `user-service` needs to be at least minimally implemented:

[x] Domain implemented (`User` aggregate: `ChangeRole`, `Activate`, `Deactivate`, `ChangePassword`; `UserID`/`Email`/`Role` value objects; domain events)
[x] Application implemented (`RegisterUser`, `ChangeRole`, `ActivateUser`, `DeactivateUser`, `Login` command handlers)
[x] Ports defined (`input`/`output` under `internal/application/port`, including the new `PasswordHasher`/`TokenIssuer` ports)
[x] Adapters implemented (Postgres repository/queries/unit of work, bcrypt hasher, JWT issuer/verifier, HTTP handlers, auth middleware)
[x] PostgreSQL operational (verified 2026-09-19 via `go test -tags=integration` against a real Postgres in a testcontainers container)
[x] Migrations operational (verified 2026-09-19 — the integration suite applies `migrations/*.up.sql` via `golang-migrate`)
[ ] HTTP operational (handlers and middleware are unit-tested with mocks; not yet exercised manually against a running server bound to a real port)
[ ] Docker operational (Dockerfile exists, not yet built/run in this pass)
[ ] Main endpoints working via manual testing (not yet done in this pass)
[x] `go test ./...` executing successfully (verified 2026-09-19, all packages green; run via `make test`)

Status as of 2026-09-19: code-complete, unit-tested (domain aggregate, value objects, application command handlers, bcrypt/JWT adapters, HTTP handlers, auth middleware, config) and integration-tested against a real Postgres (repository, queries, unit of work/outbox atomicity, unique-email violation). Still open: a manual smoke test against the actual `docker-compose` stack (not an ephemeral test container), the Docker image build, and confirming the `cdc-connector` relay actually picks up and publishes `user.registered` to Kafka end to end.

### Note — two test suites, two commands

- `make test` / `go test ./...` (per module): unit tests only — domain, application (mocked ports), HTTP handlers and middleware (mocked use cases/verifier), bcrypt/JWT adapters (real, but fast — bcrypt cost lowered to `bcrypt.MinCost` in tests), config. Fast, no external dependencies, safe to run on every commit.
- `make test-integration` / `go test -tags=integration ./...` (per module): spins up a real Postgres via testcontainers-go, applies migrations, and exercises the Postgres adapters end to end. Requires Docker.

Once all requirements above are met:

This will be the potential risk matrix:

R1 — Weak or duplicate credentials may be accepted at registration (no password strength policy enforced today beyond non-empty)
R2 — Password may be stored insecurely or in a reversible form
R3 — A JWT may be forged, or an expired/malformed token may be accepted
R4 — Authorization may be bypassed — a non-admin request may reach an admin-only route
R5 — A client may escalate its own role by supplying `role` at self-registration
R6 — Login error messages may allow account enumeration (distinguishing "no such email" from "wrong password")
R7 — The last remaining admin account may be deactivated, locking out all administrative access
R8 — A `user.registered`/`user.role_changed`/etc. event written to the outbox may never reach Kafka if the `cdc-connector` relay is down (see ADR-003, LEARNING-005)
R9 — A JWT issued before an account is deactivated remains valid until expiry — deactivation is not enforced per-request, only checked at login (see ADR-007, Security Considerations)

## Coverage so far

Unit tests:

- `internal/domain/valueobject/{userID,email,role}_test.go`: construction/validation (including email normalization to lowercase+trim, and the closed role enum), `Equal`, `IsZero`, JSON marshal/unmarshal.
- `internal/domain/event/event_test.go`: all five domain event constructors (`UserRegistered`, `UserRoleChanged`, `UserDeactivated`, `UserActivated`, `PasswordChanged`).
- `internal/domain/user/user_test.go`: constructor validation, `ChangeRole`/`Deactivate`/`Activate`/`ChangePassword` happy and failure paths (including idempotency guards), event emission, and that a failed mutation leaves state and events untouched; `RestoreUser` never emits events (mitigates R1, R2's write side).
- `internal/application/command/user_commands_test.go`: all five command handlers against mocked repository/outbox/unit of work/hasher/issuer — verifies events are drained to the outbox only on success, role is always forced to `"user"` on self-registration regardless of client input (mitigates R5), and `Login` collapses unknown-email and wrong-password into the same `ErrInvalidCredentials` (mitigates R6) while a deactivated account gets a distinct `ErrAccountDeactivated`.
- `internal/infrastructure/security/{bcrypt_hasher,jwt_issuer}_test.go`: real bcrypt hash/compare round trip (mitigates R2); real JWT issue/verify round trip, tampered-secret rejection, expired-token rejection, malformed-token rejection (mitigates R3).
- `internal/infrastructure/http/middleware/{auth,require_role}_test.go`: missing/malformed Authorization header → 401, invalid token → 401, disallowed role → 403, missing claims → 401 (mitigates R4).
- `internal/infrastructure/http/handler/handler_test.go`: HTTP handlers against mocked use cases/queries — malformed body → 400, use case error → 422, invalid credentials → 401, deactivated account → 403, not found → 404, query error → 500 (mitigates R5, R6).
- `internal/config/env_test.go`: `getEnv`/`getRequiredEnv` edge cases, plus `JWT_SECRET` missing fails closed and `BOOTSTRAP_ADMIN_EMAIL`/`BOOTSTRAP_ADMIN_PASSWORD` must be set together.

Integration tests (build tag `integration`, real Postgres via testcontainers):

- `internal/infrastructure/persistence/postgres/integration_test.go`:
  - `TestUserRepository`: save + find-by-id + find-by-email round trip, update via the `ON CONFLICT DO UPDATE` branch, unique-email violation mapped to `domainErrors.ErrEmailAlreadyRegistered` (mitigates R1's uniqueness half), and the not-found error path.
  - `TestUserQueries`: `GetUserByID` and `ListUsers` against real rows.
  - `TestUnitOfWork`: the full Transactional Outbox path — user and domain event committed atomically on success, and both left unpersisted when the use case returns an error after saving. Directly validates ADR-003 also holds for this bounded context.

## Open work

- R7 (last-admin lockout): no invariant today prevents deactivating the only remaining admin account. Deliberately deferred — same spirit as inventory-service's deferred "create SKU" endpoint — needs a count-of-active-admins check before `Deactivate`/`ChangeRole` succeeds, not yet implemented.
- R9 (deactivation not enforced per-request): `Auth` middleware only validates the JWT's signature and expiry, not whether the account behind `sub` is still active. A deactivated user's still-valid token keeps working until it expires. Would need either a short `JWT_EXPIRY`-only mitigation (current approach) or a per-request DB check (defeats some of the point of a stateless JWT) or a revocation list — none implemented yet.
- R1 (password strength): `RegisterUser` only rejects an empty password; no minimum length/complexity policy. Not covered by any test asserting a weak password is rejected, because none is rejected today.
- Manual smoke test against the real `docker-compose.yml` stack (not testcontainers) — including confirming the Kafka topic `user-service.user.registered` actually receives a message via the `cdc-connector` relay — not yet performed in this pass (see LEARNING-005 for why this step matters and isn't optional).
