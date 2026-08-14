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

[ ] Domain implemented
[ ] Application implemented
[ ] Ports defined
[ ] Adapters implemented
[ ] PostgreSQL operational
[ ] Migrations operational
[ ] HTTP operational
[ ] Docker operational
[ ] Main endpoints working via manual testing
[ ] `go test ./...` executing successfully

Once all requirements are met:

This will be the potential risk matrix:

R1 — Order may be created incorrectly
R2 — Order may be persisted incorrectly
R3 — API may accept invalid data
R4 — API may return an incorrect contract
R5 — Concurrency may lead to an inconsistent state
R6 — Internal errors may leak information
R7 — Future changes may break existing behavior