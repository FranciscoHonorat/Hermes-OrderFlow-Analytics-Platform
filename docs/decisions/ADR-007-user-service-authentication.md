# ADR-007 — user-service: autenticação, hashing e bounded context dedicado

- Status: Accepted
- Date: 2026-09-19

## Context

Até este ponto do roadmap, nenhum serviço do Hermes OrderFlow tem
autenticação: `order-service` e `inventory-service` expõem suas APIs
sem exigir identidade de quem chama. O `user-service` introduz gestão
de usuários (registro, papéis admin/user) e precisa decidir como provar
identidade nas chamadas subsequentes, como armazenar senhas, e onde
essa responsabilidade deve viver no sistema.

## Decision

`user-service` é um bounded context próprio, seguindo o mesmo padrão
hexagonal + DDD + Transactional Outbox já usado por
`inventory-service` (ADR-002, ADR-003). Autenticação usa JWT assinado
com HS256, sem estado no servidor entre requisições. Senhas são
armazenadas com bcrypt.

## Why JWT e não sessão com store compartilhado

Sessão com store compartilhado (ex: Redis) exigiria introduzir uma
dependência de infraestrutura nova só para autenticação — nenhum outro
serviço do projeto hoje depende de um store de sessão, e cada request
autenticado precisaria de uma consulta extra a esse store antes de
prosseguir. JWT resolve isso sem estado: qualquer serviço com o mesmo
segredo consegue validar um token localmente, sem round-trip a um
store central. O custo é não conseguir revogar um token antes de
expirar — aceito no v1, com expiração curta (`JWT_EXPIRY`, default
1h) como mitigação parcial.

## Why bcrypt e não argon2

`golang.org/x/crypto` já era uma dependência transitiva do projeto
(via testcontainers), mas nenhum pacote de hashing havia sido usado
diretamente ainda. bcrypt é a escolha mais simples e testada em campo
para hashing de senha em aplicações web deste porte, com um único
parâmetro de custo para ajustar. argon2 (vencedor da Password Hashing
Competition) é mais resistente a ataques com hardware dedicado
(GPU/ASIC), mas exige tunar três parâmetros (tempo, memória,
paralelismo) em vez de um, e não muda a superfície de risco real deste
projeto neste estágio. Fica registrado como possível evolução.

## Why um bounded context dedicado e não acoplar a um serviço existente

Identidade e credenciais são um conceito de domínio próprio — nem
`order-service` nem `inventory-service` deveriam ser donos de dados de
autenticação de outros bounded contexts (ver ARCHITECTURE.md, seção 3:
Bounded Contexts). Um serviço dedicado mantém a fronteira clara e
reaproveita o shared kernel (ADR-005) exatamente como os demais.

## Alternatives Considered

### Sessão com cookie + store compartilhado

Permitiria revogação imediata de sessão, mas introduz uma dependência
de infraestrutura (Redis ou equivalente) só para este propósito, e um
ponto de consulta síncrono em todo request autenticado de todo
serviço que viesse a precisar de autenticação. Descartado por
desproporcional ao estágio atual do projeto.

### OAuth2/OIDC com um provedor de identidade externo

Delegaria login e emissão de token a um IdP dedicado (Keycloak,
Auth0). Mais robusto a longo prazo, mas adiciona uma peça de
infraestrutura inteira (e sua operação) para um sistema que hoje tem
três serviços internos falando entre si — não há hoje um consumidor
externo (app mobile, terceiros) que justifique o custo.

### Acoplar autenticação ao order-service ou inventory-service

Evitaria criar um serviço novo, mas violaria a fronteira de bounded
context que o restante do projeto já segue — nenhum dos dois deveria
ganhar a responsabilidade de possuir credenciais de usuário só por
conveniência.

## Trade-offs

### Benefícios

- Sem estado de sessão compartilhado: qualquer serviço futuro que
  precise validar um JWT só precisa do segredo, sem nova dependência
  de infraestrutura.
- Reaproveita o shared kernel inteiro (outbox, eventos, conexão
  Postgres) sem duplicar nada — mesmo padrão do inventory-service.
- Hashing e emissão de token ficam atrás de portas
  (`output.PasswordHasher`, `output.TokenIssuer`), então trocar bcrypt
  por argon2 ou HS256 por RS256 no futuro não exige mudar a camada de
  aplicação.

### Custos

- Um JWT não pode ser revogado antes de expirar — deslogar um usuário
  ou desativar uma conta comprometida não invalida tokens já emitidos
  até que expirem. Mitigado parcialmente por `JWT_EXPIRY` curto e por
  `LoginHandler` verificar `active` a cada login (mas não a cada
  request autenticado — ver Security Considerations).
- Sem refresh token no v1: um token expirado exige novo login, sem
  renovação silenciosa. Escopo adiado deliberadamente.
- `user-service` é o primeiro processo do sistema com um segredo de
  longa duração (`JWT_SECRET`) que precisa ser gerenciado com cuidado
  em produção — nenhum outro serviço do projeto tem hoje esse tipo de
  segredo.

## Consequences

    Cliente
      │ POST /auth/login
      ▼
    user-service ──── bcrypt.Compare ──── users (Postgres)
      │ JWT (HS256, exp=1h)
      ▼
    Cliente reusa o JWT em Authorization: Bearer <token>
      │
      ▼
    user-service (middleware Auth + RequireRole) ou qualquer
    serviço futuro que compartilhe JWT_SECRET

`RegisterUser` sempre força `role="user"` no servidor — nunca confia
em papel vindo do cliente. O primeiro admin é criado via
`BOOTSTRAP_ADMIN_EMAIL`/`BOOTSTRAP_ADMIN_PASSWORD` no startup do
serviço (idempotente — só cria se o email ainda não existir), evitando
uma senha hardcoded em migration.

## Security Considerations

- `JWT_SECRET` é o único env var de segurança do projeto que **não**
  tem default — ausência faz o serviço falhar no boot
  (`getRequiredEnv`), ao contrário de todo outro env var do projeto,
  que tem fallback sensato.
- Um JWT válido continua sendo aceito pelo middleware `Auth` mesmo se
  a conta for desativada depois de emitido — a checagem de `active`
  hoje só acontece em `LoginHandler`, não a cada request autenticado.
  Isso é uma lacuna conhecida do v1 (ver Trade-offs e Open Questions
  da QA strategy), não coberta enquanto não houver revogação/store de
  sessão.
- `LoginHandler` responde com o mesmo erro
  (`ErrInvalidCredentials`) tanto para email inexistente quanto para
  senha incorreta — evita enumeração de contas por diferença de
  mensagem.
- Não existe hoje uma invariante que impeça desativar o último usuário
  admin restante (lockout). Adiado deliberadamente, registrado como
  risco aberto na QA strategy.

## Related Decisions

- ADR-002 — Uso de Hexagonal Architecture
- ADR-003 — Uso de Transactional Outbox
- ADR-005 — Shared Kernel entre bounded contexts
