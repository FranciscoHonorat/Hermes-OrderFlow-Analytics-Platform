<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import BaseButton from '@/components/BaseButton.vue'
import StatusPill from '@/components/StatusPill.vue'
import { cancelOrder, confirmOrder, getOrder, shipOrder } from '@/api/orders'
import { useToastStore } from '@/stores/toast'
import { extractErrorMessage } from '@/utils/errors'
import { formatCents, formatDateTime } from '@/utils/format'
import type { OrderDTO } from '@/types/order'

const props = defineProps<{ id: string }>()
const router = useRouter()
const toast = useToastStore()

const order = ref<OrderDTO | null>(null)
const loading = ref(true)
const error = ref('')
const busy = ref(false)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await getOrder(props.id)
    order.value = res.data
  } catch (err) {
    error.value = extractErrorMessage(err)
  } finally {
    loading.value = false
  }
}

async function runAction(action: (id: string) => Promise<unknown>, label: string) {
  busy.value = true
  try {
    await action(props.id)
    toast.push(`Pedido ${label}`, 'success')
    await load()
  } catch (err) {
    toast.push(extractErrorMessage(err), 'error')
  } finally {
    busy.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <BaseButton variant="secondary" @click="router.push('/orders')">← Voltar</BaseButton>

    <p v-if="loading">Carregando…</p>
    <p v-else-if="error" class="error">{{ error }}</p>

    <div v-else-if="order" class="detail">
      <h1>Pedido {{ order.id }}</h1>
      <StatusPill :label="order.status" />

      <dl>
        <dt>Cliente</dt>
        <dd>{{ order.customer_id }}</dd>
        <dt>Total</dt>
        <dd>{{ formatCents(order.total, order.currency) }}</dd>
        <dt>Criado em</dt>
        <dd>{{ formatDateTime(order.created_at) }}</dd>
        <dt>Atualizado em</dt>
        <dd>{{ formatDateTime(order.updated_at) }}</dd>
      </dl>

      <h2>Itens</h2>
      <p class="note">
        O backend não retorna os itens do pedido nesta consulta hoje (a leitura não faz
        join com a tabela de itens) — este campo aparece vazio mesmo para pedidos com itens
        de fato registrados. A moeda também é sempre exibida como "{{ order.currency }}",
        fixo no backend.
      </p>
      <ul v-if="order.items.length">
        <li v-for="item in order.items" :key="item.product_id">
          {{ item.product_id }} — qtd {{ item.quantity }} — {{ formatCents(item.price, order.currency) }}
        </li>
      </ul>

      <div class="actions">
        <BaseButton :disabled="busy" @click="runAction(confirmOrder, 'confirmado')">Confirmar</BaseButton>
        <BaseButton variant="secondary" :disabled="busy" @click="runAction(shipOrder, 'enviado')">
          Enviar
        </BaseButton>
        <BaseButton variant="danger" :disabled="busy" @click="runAction(cancelOrder, 'cancelado')">
          Cancelar
        </BaseButton>
      </div>
    </div>
  </div>
</template>

<style scoped>
.detail {
  margin-top: var(--space-4);
}

dl {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: var(--space-1) var(--space-4);
  margin: var(--space-4) 0;
}

dt {
  color: var(--color-text-muted);
}

dd {
  margin: 0;
}

.note {
  background: #fbf0dd;
  color: var(--color-warning);
  padding: var(--space-3);
  border-radius: var(--radius);
  font-size: 0.85rem;
}

.actions {
  display: flex;
  gap: var(--space-2);
  margin-top: var(--space-4);
}

.error {
  color: var(--color-danger);
}
</style>
