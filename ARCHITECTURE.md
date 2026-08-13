# Arquitetura do Projeto

Este documento consolida os padrões arquiteturais e de código que devem ser mantidos no projeto.

## 1. Princípios gerais

- Manter separação clara entre domínio, aplicação e infraestrutura.
- O domínio deve ser o núcleo do sistema: limpo, testável e independente de frameworks, banco de dados e bibliotecas externas.
- A infraestrutura implementa adaptações externas, mas não define regras de negócio.
- O código deve priorizar clareza, encapsulamento, consistência e testabilidade.

## 2. Arquitetura Hexagonal e Clean Architecture

A estrutura do projeto deve seguir a ideia de portas e adaptadores:

- Domain: regras de negócio, entidades, value objects, eventos e interfaces de repositórios.
- Application: casos de uso, comandos, queries e orquestração de fluxos.
- Infrastructure: adaptadores para banco de dados, mensageria, cache, HTTP e outros serviços externos.
- cmd/server: ponto de composição e wiring do sistema.

Regra principal:

- O domínio não deve depender de infraestrutura.
- A infraestrutura deve implementar interfaces definidas pela camada interna.
- O fluxo de dependência deve ser sempre do núcleo para fora.

## 3. DDD (Domain-Driven Design)

O projeto deve seguir os conceitos básicos de DDD:

- Bounded Contexts: cada serviço representa um contexto de negócio bem definido.
- Aggregates: entidades agrupadas que devem preservar invariantes de negócio.
- Entities: objetos com identidade própria e comportamento.
- Value Objects: objetos imutáveis que representam conceitos de domínio com validação própria.
- Domain Events: eventos que representam mudanças relevantes no domínio.

## 4. Padrão de eventos de domínio

Todos os eventos de domínio devem seguir o mesmo padrão estrutural.

Regra:
- Os eventos devem usar composição com BaseEvent para reaproveitar propriedades fundamentais como AggregateID, OccurredAt e EventName.
- Evitar criação de structs avulsas com tipagens soltas no construtor.

Exemplo correto:

```go
type OrderCustomEvent struct {
    BaseEvent
}
```

Principais diretrizes:
- Eventos de domínio devem ser usados para expressar mudanças de negócio.
- Eles devem ser pequenos, claros e semanticamente ricos.
- O agregado é responsável por criar os eventos que representam suas mudanças.

## 5. Padrão de mutação de agregados

Modificações de estado de entidades devem acontecer somente dentro do agregado, por meio de métodos explícitos.

Regra:
- Repositórios e handlers não devem alterar propriedades internas diretamente.
- O agregado deve validar invariantes e encapsular regras de negócio.

Exemplo:

```go
func (o *Order) Confirm(now time.Time) error {
    if o.status != StatusPlaced {
        return ErrInvalidTransition
    }

    o.status = StatusConfirmed
    o.updatedAt = now.UTC()
    return nil
}
```

Não fazer:

```go
// evite
handler.repo.Save(order)
order.Status = StatusConfirmed
```

## 6. Regras para entidades e value objects

- Entidades devem representar conceitos com identidade e comportamento.
- Value objects devem ser imutáveis e validar seus próprios invariantes.
- Validações devem acontecer no próprio objeto, nunca fora dele.
- Objetos de valor devem ser usados para conceitos como Money, Quantity, Address, CustomerID, OrderID, etc.

## 7. CQRS

A aplicação deve separar escrita e leitura sempre que fizer sentido:

- Commands: operações de escrita e alteração de estado.
- Queries: operações de leitura e projeção.

Essa separação melhora clareza, escalabilidade e manutenção.

## 8. Tratamento de erros

- Erros de domínio devem ser explícitos e tipados.
- Erros devem ser retornados, não escondidos.
- Evitar uso de panic em fluxos normais de negócio.
- O código deve preferir retornar erro para o chamador e permitir que a camada apropriada trate a falha.

## 9. Testabilidade

O projeto deve ser construído de forma testável:

- Testes unitários devem cobrir regras de domínio sem depender de banco ou infraestrutura.
- Testes de aplicação devem validar fluxos de uso.
- Testes de integração devem ser usados apenas quando houver dependência real de infraestrutura.

## 10. Responsabilidades por camada

### Domain
- Regras de negócio.
- Entidades, value objects, aggregates.
- Eventos de domínio.
- Interfaces de repositórios.

### Application
- Casos de uso.
- Comandos e queries.
- Coordenação entre domínio e infraestrutura.

### Infrastructure
- Implementação de repositórios.
- Adapters para banco de dados, Kafka, Redis, HTTP e outros serviços.
- Mapeamento entre modelos externos e do domínio.

### Entry points
- Configuração e wiring do sistema.
- Inicialização de serviços e dependências.

## 11. Padrões de código a manter

- Nomear funções e tipos de forma clara e consistente.
- Preferir pequenos métodos com responsabilidade única.
- Evitar lógica de negócio espalhada por handlers ou infraestrutura.
- Manter o código simples, explícito e fácil de evoluir.
- Priorizar composição e encapsulamento em vez de acoplamento forte.

## 12. Resumo das regras de arquitetura

Os padrões abaixo devem ser preservados em todas as mudanças do projeto:

- Arquitetura hexagonal.
- Clean Architecture.
- DDD.
- Ports and Adapters.
- CQRS.
- Eventos de domínio com BaseEvent.
- Mutação de estado somente dentro do agregado.
- Regras de negócio encapsuladas no domínio.
- Tratamento consistente de erros.
- Código testável e com baixo acoplamento.
