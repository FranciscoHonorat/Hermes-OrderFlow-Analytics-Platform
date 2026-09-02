# ADR-002 — Uso de Arquitetura Hexagonal no Order Service

- Status: Accepted
- Date: 2026-08-30

## Context

O Order Service possui regras de negócio relacionadas ao ciclo de vida
dos pedidos.

Essas regras precisam permanecer independentes de detalhes externos
como:

- HTTP;
- PostgreSQL;
- Kafka;
- Docker;
- bibliotecas de infraestrutura.

Durante o desenvolvimento do projeto, também foi identificado que o
serviço precisaria utilizar diferentes mecanismos de infraestrutura,
como persistência em PostgreSQL e comunicação assíncrona através de
eventos.

Isso cria o risco de acoplamento entre as regras de negócio e as
tecnologias utilizadas para executá-las.

## Decision

Foi adotada a Arquitetura Hexagonal (Ports & Adapters) para o
Order Service.

A aplicação e o domínio permanecem no centro da arquitetura, enquanto
as integrações externas são realizadas através de ports e adapters.

Estruturalmente:

    Adapters
        ↓
      Ports
        ↓
    Application
        ↓
      Domain
        ↑
      Ports
        ↑
    Adapters

As dependências devem apontar para dentro da aplicação e do domínio.

O domínio não deve depender diretamente de PostgreSQL, Kafka, HTTP ou
outras tecnologias externas.

## Motivation

A principal motivação para essa decisão foi isolar as regras de negócio
das preocupações de infraestrutura.

Isso permite que regras relacionadas à Order sejam desenvolvidas e
testadas sem que o domínio precise conhecer os detalhes de:

- banco de dados;
- transporte HTTP;
- mensageria;
- mecanismos de persistência.

Esse isolamento também facilita a substituição ou evolução dos
componentes externos sem exigir alterações nas regras centrais do
domínio.

## Security Considerations

A Arquitetura Hexagonal não é, por si só, um mecanismo de segurança.

Entretanto, o isolamento entre domínio e infraestrutura contribui para
uma arquitetura com fronteiras mais claras.

Isso permite concentrar controles de segurança nos pontos apropriados,
como:

- autenticação nos adapters de entrada;
- autorização na camada apropriada da aplicação;
- validação de entrada nas bordas;
- proteção das credenciais nos adapters de infraestrutura;
- controle de acesso ao banco;
- comunicação segura com serviços externos.

Dessa forma, uma dependência externa não precisa ser introduzida
diretamente nas regras de negócio.

## Alternatives Considered

### Arquitetura em camadas tradicional

Uma arquitetura em camadas poderia separar apresentação, aplicação,
domínio e infraestrutura.

Entretanto, a Arquitetura Hexagonal deixa explícita a direção das
dependências e o conceito de ports e adapters, o que atende melhor ao
objetivo do projeto de manter o domínio independente da infraestrutura.

### Acoplamento direto à infraestrutura

Outra possibilidade seria permitir que os casos de uso utilizassem
diretamente PostgreSQL, Kafka ou HTTP.

Essa alternativa reduziria a quantidade inicial de abstrações, porém
aumentaria o acoplamento e dificultaria testes e substituição das
tecnologias externas.

## Trade-offs

### Benefícios

- Isolamento das regras de negócio.
- Menor acoplamento com infraestrutura.
- Maior testabilidade.
- Facilita substituição de adapters.
- Fronteiras arquiteturais explícitas.
- Permite evolução independente das integrações externas.

### Custos

- Maior quantidade de interfaces e abstrações.
- Estrutura inicial mais complexa.
- Exige disciplina para manter as dependências apontando para dentro.
- Pode gerar abstrações desnecessárias se aplicada sem necessidade.

## Consequences

O Order Service passa a possuir uma separação explícita entre:

- domínio;
- aplicação;
- ports;
- adapters.

Por exemplo:

    application/port/output
            ↓
    OrderRepository
            ↓
    PostgreSQL Adapter

A aplicação conhece o contrato do repository, mas não precisa conhecer
a implementação PostgreSQL.

Da mesma forma:

    Application
        ↓
    OutboxRepository
        ↓
    PostgreSQL

Essa separação permite que a infraestrutura seja modificada sem
alterar diretamente as regras centrais do domínio.

## Bounded Contexts

A adoção da Arquitetura Hexagonal neste contexto não significa que todos
os Bounded Contexts do projeto deverão obrigatoriamente utilizar a mesma
arquitetura interna.

Cada Bounded Context deve ser avaliado de acordo com:

- complexidade do domínio;
- criticidade;
- requisitos de consistência;
- integrações externas;
- necessidade de testabilidade;
- requisitos de segurança;
- custo da abstração.

A arquitetura deve servir ao contexto e não ser aplicada como uma
regra universal sem justificativa.

## Related Decisions

- ADR-001 — Uso do PostgreSQL
- ADR-003 — Uso de Transactional Outbox
- ADR-004 — Uso de Kafka