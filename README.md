# OrderFlow Analytics Platform

O **OrderFlow Analytics Platform** é um projeto em desenvolvimento para demonstrar na prática conceitos de arquitetura de software, sistemas distribuídos, engenharia de dados e infraestrutura moderna.

A proposta é construir uma plataforma de gestão de pedidos com fluxo analítico em tempo real, cobrindo pilares como:

- **Clean Architecture**
- **Domain-Driven Design (DDD)**
- **Hexagonal Architecture**
- **CQRS**
- **Event-driven architecture** (Transactional Outbox + relay para Kafka)
- **OLTP vs OLAP**
- **Containerização e orquestração com Docker e Kubernetes**

---

## 🎯 Objetivo do projeto

Criar uma solução de ponta a ponta para:

- processar pedidos de forma consistente;
- publicar eventos de domínio com garantia de entrega (outbox);
- integrar múltiplos serviços de forma assíncrona;
- gerar dados analíticos para métricas e relatórios;
- demonstrar boas práticas de desenvolvimento com Go e arquitetura modular.

---

## 🏗️ Arquitetura

Cada serviço é independente, tem seu próprio banco no PostgreSQL e grava eventos de domínio na tabela `outbox` na mesma transação da escrita de negócio. O **cdc-connector** lê essas tabelas e publica os eventos no Kafka.

```text
                ┌──────────────┐
  frontend ───▶ │ order-service│──┐
  (Vue 3)  ───▶ │ inventory-svc│──┼──▶ PostgreSQL (orderflow / inventory / users)
           ───▶ │ user-service │──┘            │  tabelas de negócio + outbox
                └──────────────┘               ▼
                                        ┌──────────────┐        ┌───────┐
                                        │ cdc-connector│──────▶ │ Kafka │
                                        └──────────────┘        └───────┘
```

| Serviço | Responsabilidade | Status |
|---|---|---|
| **order-service** | Contexto principal de pedidos (criar, confirmar, cancelar, enviar) | ✅ Implementado |
| **inventory-service** | Gestão de estoque (reservar, liberar, repor, estoque baixo) | ✅ Implementado |
| **user-service** | Cadastro, login, emissão de JWT e gestão de papéis | ✅ Implementado |
| **cdc-connector** | Relay do outbox de cada serviço para tópicos Kafka | ✅ Implementado |
| **analytics-service** | Processamento analítico e métricas (ClickHouse) | 🕓 Planejado |
| **notification-service** | Notificações | 🕓 Planejado |

A estrutura interna dos serviços segue a arquitetura hexagonal:

- **domain**: entidades, value objects, eventos e regras de negócio
- **application**: casos de uso, comandos e queries
- **infrastructure**: adaptadores para banco, mensageria e integrações externas
- **cmd/server**: ponto de entrada e wiring do serviço

O módulo **shared** é o shared kernel usado por todos os serviços: eventos base, outbox, conexão PostgreSQL, produtor Kafka e validação de JWT.

---

## 🛠️ Stack tecnológica

### Backend
- Go (workspace multi-módulo via `go.work`)
- PostgreSQL 16 para OLTP, com migrações via `golang-migrate`
- Kafka (Confluent 7.6) para eventos
- Autenticação JWT compartilhada entre os serviços

### Frontend
- Vue 3 + Vite + TypeScript, Pinia e Vue Router

### Infraestrutura
- Docker e Docker Compose
- Swagger UI para a documentação OpenAPI
- Kubernetes (k3s/minikube): manifestos iniciais em `infra/k8s`

### Planejado
- ClickHouse para OLAP
- Redis para cache
- Prometheus + Grafana para observabilidade

---

## 📦 Estrutura do repositório

```text
services/
  order-service/
  inventory-service/
  user-service/
  cdc-connector/
  analytics-service/      # esqueleto
  notification-service/   # esqueleto
shared/                   # shared kernel: auth, events, outbox, messaging, postgres
frontend/                 # dashboard Vue 3
infra/
  docker/postgres-init/   # cria os bancos inventory e users
  k8s/
docs/
  decisions/              # ADRs
  learning/               # notas de estudo
  QA/                     # estratégias de teste por serviço
  openapi/                # specs OpenAPI
  postman/                # collection + environment
  performance/            # gap analysis de performance
```

---

## 🚀 Como rodar

Pré-requisitos: Docker com Compose, Go e Node.js (para o frontend).

```bash
# sobe Postgres, Kafka, migrações, os quatro serviços e o Swagger UI
make docker-up

# derruba tudo
make docker-down
```

| Componente | Endereço |
|---|---|
| order-service | http://localhost:8080 |
| inventory-service | http://localhost:8081 |
| cdc-connector (health) | http://localhost:8082 |
| user-service | http://localhost:8083 |
| Swagger UI | http://localhost:8090 |
| PostgreSQL | `localhost:5433` (user/senha `postgres`) |
| Kafka (acesso pelo host) | `localhost:9093` |

Frontend:

```bash
cd frontend
npm install
npm run dev   # http://localhost:5173
```

Login: crie uma conta em `/register` ou use o admin de bootstrap definido no `docker-compose.yml` (`BOOTSTRAP_ADMIN_EMAIL` / `BOOTSTRAP_ADMIN_PASSWORD`).

> ⚠️ As credenciais e o `JWT_SECRET` do compose são apenas para desenvolvimento.

---

## 🧪 Testes

```bash
make test               # testes unitários de todos os módulos do workspace
make test-integration   # testes de integração (requer Docker, usa testcontainers)
```

A estratégia de testes inclui:

- testes unitários para regras de negócio e casos de uso;
- testes de integração com PostgreSQL e Kafka reais;
- validação ponta a ponta contra a stack do `docker-compose`.

O detalhamento por serviço está em [docs/QA](docs/QA).

---

## 🔬 Destaque técnico

Antes de integrar infraestrutura mais complexa, o projeto priorizou a otimização do núcleo do domínio. Em especial, o value object de dinheiro foi analisado com foco em desempenho, alocação e impacto no garbage collector, com estudo prático de escape analysis em Go.

Na integração entre serviços, a publicação de eventos usa **Transactional Outbox** + **relay (cdc-connector)**, o que evita o problema de dual write: o evento só existe se a transação de negócio for confirmada, e é entregue ao Kafka com semântica at-least-once.

---

## 🗺️ Próximos passos

- implementar o **analytics-service** consumindo os eventos do Kafka e gravando em ClickHouse;
- implementar o **notification-service**;
- introduzir cache com Redis;
- adicionar observabilidade com Prometheus + Grafana;
- completar os manifestos de deploy em Kubernetes.

---

## 📚 Documentação

- [ARCHITECTURE.md](ARCHITECTURE.md)
- [Decisões de arquitetura (ADRs)](docs/decisions)
- [Notas de aprendizado](docs/learning)
- [inventory-service e cdc-connector](docs/architecture/inventory-service-e-cdc-connector.md)
- [Documentação da API (OpenAPI/Swagger)](docs/openapi/index.html)
- [Collection Postman](docs/postman)
- [Gap analysis de performance](docs/performance/gap-analysis.md)
- [Frontend](frontend/README.md)

---

## 🤝 Contribuição

Contribuições são bem-vindas. O objetivo é manter o projeto alinhado com boas práticas de arquitetura, testes e evolução incremental.
