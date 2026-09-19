# LEARNING-005 — cdc-connector e o relay do Outbox

- Date: 2026-09-19
- Topic: cdc-connector, Kafka
- Related ADR: ADR-003, ADR-004, ADR-005, ADR-006

## Initial Hypothesis

A hipótese inicial era que implementar o `cdc-connector` seria, em boa
parte, "conectar peças que já existiam": `shared/outbox.Repository` já
tinha `FetchUnprocessed`/`MarkProcessed` prontos desde a Fase 0, e
`shared/messaging/kafka` já tinha um `Producer`. Bastaria escrever o
loop de polling.

## What I Expected

Esperava que o trabalho fosse majoritariamente escrever `relay.Run`,
testar com fakes, subir no `docker-compose` e ver eventos chegando no
Kafka de primeira — afinal, cada peça já tinha sido testada
isoladamente antes (outbox com Postgres real na Fase 1, Kafka com
testcontainers nesta própria Fase 2).

## What I Learned

O relay em si (`internal/relay`) funcionou de primeira, testado com
fakes e depois com Postgres+Kafka reais via testcontainers
(`confluentinc/confluent-local`, modo KRaft). Mas ao subir no
`docker-compose.yml` real do projeto — usando `confluentinc/cp-kafka`
com Zookeeper, não KRaft — o relay simplesmente não publicava nada, e
sem nenhum erro visível.

A investigação revelou três problemas empilhados, cada um mascarando o
próximo:

1. **`kgo.Producer.Publish` sem `AllowAutoTopicCreation()`.** O
   primeiro evento publicado em um tópico novo falhava com
   `UNKNOWN_TOPIC_OR_PARTITION`, porque o cliente `kgo` não pede
   criação automática de tópico por padrão, mesmo com
   `auto.create.topics.enable=true` no broker. Isso só apareceu no
   teste de integração com Kafka real — os testes com fake `Publisher`
   nunca exercitam esse detalhe de protocolo.

2. **`EventPublisher`/`Serializer` eram uma abstração morta.** Ao
   escrever o relay, ficou claro que ele nunca precisaria serializar
   nada — o `Payload` de uma linha do outbox já é o JSON gravado pela
   aplicação. `EventPublisher.Publish` esperava uma interface
   (`EventType()`/`ResourceID()`) que nenhum evento de domínio
   implementa de fato (eles implementam `events.DomainEvent`, com
   nomes de método diferentes). Essa incompatibilidade nunca tinha
   sido pega pelo compilador porque nada jamais chamava
   `EventPublisher.Publish`.

3. **O Kafka do `docker-compose.yml` estava em crash-loop desde
   sempre.** `KAFKA_ADVERTISED_LISTENERS` tinha `PLAINTEXT` e
   `PLAINTEXT_HOST` apontando para a mesma porta (9092), o que o
   broker recusa ("two listeners on the same port"). O container
   ficava reiniciando indefinidamente (`RestartCount` chegou a 31
   durante a investigação), mas isso nunca tinha sido percebido porque
   nenhum serviço, antes do `cdc-connector`, jamais havia tentado se
   conectar de fato ao Kafka — order-service publica no outbox, não no
   broker diretamente (ADR-003).

O terceiro problema foi o mais difícil de diagnosticar porque os
sintomas eram enganosos: o processo do `cdc-connector` continuava
"rodando" (health check respondendo), sem nenhum log de erro, porque a
chamada `ProduceSync` ficava bloqueada indefinidamente esperando o
broker ficar disponível, e só retornava (com `context canceled`) quando
o processo era encerrado. Um relay "silenciosamente travado" parece,
de fora, com um relay "sem nada para processar".

## Evidence

O diagnóstico foi feito eliminando variáveis uma de cada vez:

    Relay com fakes           → passou (prova a lógica)
    Relay com Postgres+Kafka   → passou (prova o contrato real)
    real via testcontainers
    Relay no docker-compose    → travou, sem nenhum log de erro

A diferença entre os dois últimos foi o sinal: o container do
`docker-compose` estava em `RestartCount: 31`
(`docker inspect orderflow-kafka`), e os logs do próprio Kafka
mostravam a exceção fatal de configuração assim que se olhava para
eles diretamente — a informação estava disponível o tempo todo, só não
no lugar óbvio (nos logs do `cdc-connector`, que só reportava
`context canceled` genérico).

## What Changed In My Understanding

Minha compreensão mudou de:

"Se cada peça foi testada isoladamente, a integração delas vai
funcionar."

para:

"Testar cada peça isoladamente prova que ELA está correta contra um
ambiente bem configurado — não prova que o ambiente real está bem
configurado. `docker-compose.yml` nunca tinha sido exercitado no
caminho que o `cdc-connector` foi o primeiro a percorrer (publicar de
fato no Kafka), então um bug de configuração pôde sobreviver ali desde
o início do projeto sem que nenhum teste automatizado o pegasse — os
testes de integração usavam Kafka efêmero via testcontainers, nunca o
Kafka do `docker-compose.yml`."

## Trade-offs I Learned

- Testes de integração com testcontainers validam o *código*, não a
  *configuração de infraestrutura declarada em outro lugar*
  (`docker-compose.yml`, manifests k8s). As duas coisas podem divergir
  silenciosamente até alguém exercitar o caminho real.
- Um processo que "não loga erro" não é o mesmo que um processo "sem
  erros" — pode estar bloqueado dentro de uma chamada que só respeita
  cancelamento de contexto, sem timeout próprio. O `Producer` não tem
  hoje um timeout de produção independente do contexto do chamador.

## Open Questions

- O `kgo.Producer` deveria ter um timeout de produção próprio (ex:
  `kgo.ProduceRequestTimeout`), para falhar rápido e logar um erro
  claro em vez de bloquear indefinidamente quando o broker está
  inacessível?
- Vale a pena um teste de "smoke" que sobe o `docker-compose.yml` real
  (não testcontainers efêmero) em CI, especificamente para pegar
  divergências de configuração como essa? Hoje nada cobre isso.

## Next Experiment

Ao integrar o próximo consumidor de Kafka (a saga order-service ⇄
inventory-service, Fase 3, ou o futuro analytics-service), validar
deliberadamente contra o `docker-compose.yml` real cedo — não confiar
apenas nos testes de integração com testcontainers — já que foi
exatamente essa lacuna que escondeu o bug do Kafka por toda a Fase 0 e
Fase 1 do projeto.
