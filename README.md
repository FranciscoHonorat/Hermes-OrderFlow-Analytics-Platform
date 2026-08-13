# OrderFlow Analytics Platform

O **OrderFlow Analytics Platform** é um projeto em desenvolvimento para demonstrar na prática conceitos de arquitetura de software, sistemas distribuídos, engenharia de dados e infraestrutura moderna.

A proposta é construir uma plataforma de gestão de pedidos com fluxo analítico em tempo real, cobrindo pilares como:

- **Clean Architecture**
- **Domain-Driven Design (DDD)**
- **Hexagonal Architecture**
- **CQRS**
- **Event-driven architecture**
- **OLTP vs OLAP**
- **Containerização e orquestração com Docker e Kubernetes**

---

## 🎯 Objetivo do projeto

Criar uma solução de ponta a ponta para:

- processar pedidos de forma consistente;
- publicar eventos de domínio;
- integrar múltiplos serviços;
- gerar dados analíticos para métricas e relatórios;
- demonstrar boas práticas de desenvolvimento com Go e arquitetura modular.

---

## 🏗️ Arquitetura proposta

O projeto foi pensado em torno de múltiplos serviços independentes, cada um com responsabilidade clara:

- **order-service**: contexto principal de pedidos
- **inventory-service**: gestão de estoque
- **analytics-service**: processamento analítico e métricas
- **notification-service**: notificações
- **cdc-connector**: captura de mudanças e publicação de eventos

A estrutura segue uma organização baseada em camadas:

- **domain**: entidades, value objects, eventos e regras de negócio
- **application**: casos de uso, comandos e queries
- **infrastructure**: adaptações para bancos, APIs, cache, mensageria e integrações externas
- **cmd/server**: ponto de entrada e wiring do serviço

---

## 🛠️ Stack tecnológica

### Linguagem e runtime
- Go 1.22+

### Dados e processamento
- PostgreSQL 16 para OLTP
- ClickHouse 24 para OLAP
- Kafka 3 para eventos e integração assíncrona
- Redis 7 para cache

### Infraestrutura
- Docker
- Docker Compose
- Kubernetes (k3s/minikube)
- Prometheus + Grafana para observabilidade

---

## 📦 Estrutura do repositório

```text
services/
  order-service/
  inventory-service/
  analytics-service/
  notification-service/
  cdc-connector/
infra/
  k8s/
  docker/
shared/
  events/
```

---

## 🚧 Status atual

O projeto está em fase inicial de desenvolvimento. A implementação do contexto de pedidos já possui estrutura base e exemplos de domínio, enquanto a integração com persistência, infraestrutura e fluxos distribuídos ainda está sendo evoluída.

O foco atual é consolidar:

- a camada de domínio;
- as regras de negócio e invariantes;
- os contratos de eventos;
- a separação entre aplicação e infraestrutura.

---

## 🧪 Testes

A estratégia de testes inclui:

- testes unitários para regras de negócio;
- testes de aplicação para fluxos de uso;
- testes de integração quando houver dependência real de infraestrutura.

---

## 🔬 Destaque técnico

Antes de integrar infraestrutura mais complexa, o projeto priorizou a otimização do núcleo do domínio. Em especial, o value object de dinheiro foi analisado com foco em desempenho, alocação e impacto no garbage collector, com estudo prático de escape analysis em Go.

Esse trabalho reforça a ideia de que a base do sistema precisa ser eficiente e bem modelada antes de escalar para ambientes distribuídos.

---

## 🗺️ Próximos passos

Os próximos marcos previstos incluem:

- completar o fluxo de pedidos com validações e eventos;
- implementar repositórios e integrações com PostgreSQL;
- introduzir publicação de eventos com Kafka;
- expandir a camada analítica com ClickHouse;
- preparar deploy com Docker e Kubernetes.

---

## 📚 Referência arquitetural

A documentação de arquitetura do projeto está em:

- [ARCHITECTURE.md](ARCHITECTURE.md)
- [orderflow-roadmap.md](orderflow-roadmap.md)

---

## 🤝 Contribuição

Contribuições são bem-vindas. O objetivo é manter o projeto alinhado com boas práticas de arquitetura, testes e evolução incremental.

