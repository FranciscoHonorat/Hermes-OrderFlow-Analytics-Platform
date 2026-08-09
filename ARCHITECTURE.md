Padrão de Eventos de Domínio

Regra: Todos os eventos de domínio devem utilizar composição com BaseEvent para reaproveitamento de propriedades fundamentais (AggregateID, OccurredAT, EventName).

Evitar criação de structs avulsas com tipagens soltas no construtor.

Certo:

type OrderCustomEvent struct {
    BaseEvent
}

Padrão de Handlers e Mutações de Agregados

Regra: Modificação de estados de entidades só acontecem dentro do Agregado através de métodos explícitos (.Ship(), .Deliver()) validando invariantes. Repositório ou Handlers não modificam propriedades internas diretamente.