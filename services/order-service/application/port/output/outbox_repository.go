package output

import (
	"github.com/FranciscoHonorat/ordemflow/shared/outbox"
)

// OutboxRepository é a porta usada pela aplicação; a definição real do
// contrato vive em shared/outbox, reaproveitada por todo bounded context
// que grava eventos de domínio na tabela outbox.
type OutboxRepository = outbox.Repository
