<script setup lang="ts">
import { onMounted, ref } from 'vue'
import BaseButton from '@/components/BaseButton.vue'
import DataTable from '@/components/DataTable.vue'
import StatusPill from '@/components/StatusPill.vue'
import { cancelOrder, confirmOrder, listOrders, shipOrder } from '@/api/orders'
import { useCartStore } from '@/stores/cart'
import { useToastStore } from '@/stores/toast'
import { extractErrorMessage } from '@/utils/errors'
import { formatCents, formatDateTime } from '@/utils/format'
import type { OrderDTO, OrderStatus } from '@/types/order'

const toast = useToastStore()
const cart = useCartStore()

const orders = ref<OrderDTO[]>([])
const loading = ref(true)
const busyId = ref<string | null>(null)
const customerIdFilter = ref('')

const columns = [
  { key: 'id', label: 'ID' },
  { key: 'customer_id', label: 'Cliente' },
  { key: 'status', label: 'Status' },
  { key: 'total', label: 'Total' },
  { key: 'created_at', label: 'Criado em' },
]

const terminalStatuses: OrderStatus[] = ['CANCELLED', 'DELIVERED']

function statusTone(status: OrderStatus) {
  if (status === 'CANCELLED') return 'danger'
  if (status === 'DELIVERED' || status === 'SHIPPED') return 'success'
  return 'neutral'
}

async function refresh() {
  loading.value = true
  try {
    const res = await listOrders({
      limit: 20,
      customer_id: customerIdFilter.value || undefined,
    })
    orders.value = res.data
  } catch (err) {
    toast.push(extractErrorMessage(err), 'error')
  } finally {
    loading.value = false
  }
}

async function runAction(action: (id: string) => Promise<unknown>, id: string, label: string) {
  busyId.value = id
  try {
    await action(id)
    toast.push(`Pedido ${label}`, 'success')
    await refresh()
  } catch (err) {
    toast.push(extractErrorMessage(err), 'error')
  } finally {
    busyId.value = null
  }
}

onMounted(refresh)
</script>

<template>
  <div>
    <div class="header">
      <h1>Pedidos</h1>
      <BaseButton @click="cart.open()">🛒 Novo pedido</BaseButton>
    </div>


    <div class="filters">
      <input v-model="customerIdFilter" placeholder="Filtrar por customer_id (uuid)" @keyup.enter="refresh" />
      <BaseButton variant="secondary" @click="refresh">Filtrar</BaseButton>
    </div>

    <DataTable :items="orders" :columns="columns" :loading="loading" empty-message="Nenhum pedido encontrado.">
      <template #cell-id="{ item }">
        <RouterLink :to="`/orders/${item.id}`">{{ item.id.slice(0, 8) }}…</RouterLink>
      </template>
      <template #cell-status="{ item }">
        <StatusPill :label="item.status" :tone="statusTone(item.status)" />
      </template>
      <template #cell-total="{ item }">{{ formatCents(item.total, item.currency) }}</template>
      <template #cell-created_at="{ item }">{{ formatDateTime(item.created_at) }}</template>

      <template #actions="{ item }">
        <div class="actions" v-if="!terminalStatuses.includes(item.status)">
          <BaseButton
            variant="secondary"
            :disabled="busyId === item.id"
            @click="runAction(confirmOrder, item.id, 'confirmado')"
          >
            Confirmar
          </BaseButton>
          <BaseButton
            variant="secondary"
            :disabled="busyId === item.id"
            @click="runAction(shipOrder, item.id, 'enviado')"
          >
            Enviar
          </BaseButton>
          <BaseButton
            variant="danger"
            :disabled="busyId === item.id"
            @click="runAction(cancelOrder, item.id, 'cancelado')"
          >
            Cancelar
          </BaseButton>
        </div>
      </template>
    </DataTable>
  </div>
</template>

<style scoped>
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filters {
  display: flex;
  gap: var(--space-2);
  margin-bottom: var(--space-4);
}

.filters input {
  flex: 1;
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
}

.actions {
  display: flex;
  gap: var(--space-2);
}
</style>
