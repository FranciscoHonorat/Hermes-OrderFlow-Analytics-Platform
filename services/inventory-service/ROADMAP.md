# Roadmap — Inventory Service

Este documento descreve o plano de evolução do **Inventory Service**, o
serviço responsável pela gestão de estoque dentro do OrderFlow Analytics
Platform.

## Decisão arquitetural

Diferente do `order-service` (Arquitetura Hexagonal), o `inventory-service`
adota uma **Arquitetura em Camadas (Layered Architecture)**.

Justificativa:

- o domínio de estoque é mais simples: poucas regras de negócio,
  poucas integrações externas e baixa necessidade de trocar
  implementações de infraestrutura;
- a ADR-002 já prevê que cada Bounded Context deve escolher a
  arquitetura mais adequada à sua complexidade, em vez de aplicar
  hexagonal de forma universal;
- uma API de CRUD/consulta de estoque não justifica o custo de
  ports & adapters completos.

Camadas propostas:

```text
cmd/server/          → wiring e ponto de entrada
internal/
  handler/            → HTTP (entrada), tradução request/response
  service/            → regras de negócio e casos de uso
  repository/         → acesso a dados (Postgres)
  model/              → entidades e DTOs do domínio de estoque
  config/             → env, database, http, messaging
migrations/           → schema do Postgres
```

Regra de dependência: `handler → service → repository → model`.
Uma camada só conhece a camada imediatamente abaixo; `model` não
depende de nenhuma outra.

Princípios do [ARCHITECTURE.md](../../ARCHITECTURE.md) que continuam
valendo mesmo em camadas:

- validação e invariantes dentro do próprio model (ex.: quantidade
  não pode ser negativa), nunca soltas em handlers;
- erros tipados e explícitos, sem panic em fluxo de negócio;
- testes unitários para `service` sem depender de banco real;
- separação clara entre o que é regra de negócio (`service`) e o que
  é infraestrutura (`repository`).

---

## Fase 0 — Fundação do serviço

- [ x] Adicionar `inventory-service` ao `go.work`
- [ x] `config/env.go` + `config/env_test.go` (seguindo o padrão TDD já
      usado no `order-service`)
- [ ] `config/database.go` (conexão Postgres)
- [ ] `config/http.go` (servidor HTTP e rotas)
- [ ] `cmd/server/main.go` com wiring inicial
- [ ] Handler de health check (`/health`)
- [ ] Dockerfile + entrada no `docker-compose.yml`

## Fase 1 — Modelo de domínio de estoque

- [ ] `model.SKU` / `model.ProductID` — identificador do produto
- [ ] `model.StockItem` — quantidade disponível, quantidade reservada,
      quantidade mínima (invariante: disponível >= 0)
- [ ] `model.Reservation` — reserva de estoque associada a um pedido
- [ ] Erros de domínio tipados (`ErrInsufficientStock`,
      `ErrItemNotFound`, `ErrInvalidQuantity`)
- [ ] Testes unitários dos invariantes do model

## Fase 2 — Camada de persistência

- [ ] Migration inicial: tabela `stock_items` (+ índice por SKU)
- [ ] Migration: tabela `stock_reservations`
- [ ] `repository.StockRepository` (interface) + implementação Postgres
- [ ] Testes de integração do repositório (Postgres real, via
      docker-compose ou testcontainers)

## Fase 3 — Camada de serviço (casos de uso)

- [ ] `CheckAvailability(sku, quantity)`
- [ ] `ReserveStock(orderID, items)` — reserva atômica, falha se
      indisponível
- [ ] `ReleaseStock(orderID)` — libera reserva (cancelamento de pedido)
- [ ] `ConfirmStock(orderID)` — efetiva a baixa após confirmação do
      pedido
- [ ] `Replenish(sku, quantity)` — reposição de estoque
- [ ] Testes de aplicação cobrindo os fluxos acima (sucesso e falha)

## Fase 4 — Camada HTTP

- [ ] `GET /stock/{sku}` — consulta de disponibilidade
- [ ] `POST /stock/{sku}/replenish` — reposição manual/administrativa
- [ ] `POST /reservations` — reserva manual (uso administrativo/testes)
- [ ] `DELETE /reservations/{orderID}` — liberação de reserva
- [ ] Middleware de request ID (reaproveitar padrão do `order-service`)
- [ ] Validação de payload na borda (handler), nunca no service

## Fase 5 — Integração assíncrona com Order Service

- [ ] Consumer Kafka: `OrderPlaced` → `ReserveStock`
- [ ] Consumer Kafka: `OrderCancelled` → `ReleaseStock`
- [ ] Consumer Kafka: `OrderConfirmed` → `ConfirmStock`
- [ ] Publicar eventos de domínio: `StockReserved`, `StockRejected`,
      `StockReplenished` (via `shared/events`)
- [ ] Definir estratégia de idempotência no consumo (chave única por
      evento processado, já que não há outbox transacional nesta
      camada)

## Fase 6 — Cache e performance

- [ ] Cache Redis para consulta de disponibilidade (`GET /stock/{sku}`)
- [ ] Estratégia de invalidação no `ReserveStock`/`Replenish`/`ConfirmStock`

## Fase 7 — Observabilidade e deploy

- [ ] Métricas Prometheus (reservas feitas, rejeitadas, latência)
- [ ] Manifests em `infra/k8s/inventory-service`
- [ ] Logs estruturados com request ID propagado

## Fase 8 — Qualidade

- [ ] Cobertura de testes unitários no `service` e `model`
- [ ] Testes de integração no `repository`
- [ ] Testes de contrato para os eventos consumidos/publicados

---

## Fora de escopo (por enquanto)

- Outbox transacional (faz sentido no `order-service` pela criticidade
  de consistência; aqui a idempotência no consumer é suficiente
  dado o escopo do serviço).
- CQRS explícito (o volume de leitura/escrita não justifica separar
  modelos de leitura e escrita neste momento).

## Próximo passo sugerido

Registrar a decisão de arquitetura em camadas como ADR
(`docs/decisions/ADR-005-use-layered-architecture-inventory-service.md`),
seguindo o mesmo formato do ADR-002, para manter o histórico de
decisões consistente entre os serviços.
