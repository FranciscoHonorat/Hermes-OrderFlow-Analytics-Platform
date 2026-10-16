<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import BaseButton from '@/components/BaseButton.vue'
import DataTable from '@/components/DataTable.vue'
import FormField from '@/components/FormField.vue'
import StatusPill from '@/components/StatusPill.vue'
import { getStock, listLowStock, releaseStock, replenishStock, reserveStock } from '@/api/inventory'
import { useToastStore } from '@/stores/toast'
import { extractErrorMessage } from '@/utils/errors'
import { formatDateTime } from '@/utils/format'
import type { StockDTO } from '@/types/inventory'

const route = useRoute()
const toast = useToastStore()
const lowStock = ref<StockDTO[]>([])
const loading = ref(true)

const lookupSku = ref('')
const lookupResult = ref<StockDTO | null>(null)
const lookupError = ref('')

const form = reactive({ sku: '', quantity: 1 })
const submitting = ref(false)

const columns = [
  { key: 'sku', label: 'SKU' },
  { key: 'available', label: 'Disponível' },
  { key: 'reserved', label: 'Reservado' },
  { key: 'minimum_quantity', label: 'Mínimo' },
  { key: 'updated_at', label: 'Atualizado em' },
]

async function refreshLowStock() {
  loading.value = true
  try {
    const res = await listLowStock()
    lowStock.value = res.data
  } catch (err) {
    toast.push(extractErrorMessage(err), 'error')
  } finally {
    loading.value = false
  }
}

async function onLookup() {
  lookupError.value = ''
  lookupResult.value = null
  if (!lookupSku.value) return
  try {
    const res = await getStock(lookupSku.value)
    lookupResult.value = res.data
  } catch (err) {
    lookupError.value = extractErrorMessage(err)
  }
}

async function runMutation(action: typeof reserveStock, label: string) {
  submitting.value = true
  try {
    await action({ sku: form.sku, quantity: Number(form.quantity) })
    toast.push(`${label} aplicado para ${form.sku}`, 'success')
    await refreshLowStock()
  } catch (err) {
    toast.push(extractErrorMessage(err), 'error')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  refreshLowStock()

  const skuFromQuery = route.query.sku
  if (typeof skuFromQuery === 'string' && skuFromQuery) {
    lookupSku.value = skuFromQuery
    onLookup()
  }
})
</script>

<template>
  <div>
    <h1>Estoque</h1>

    <section class="panel">
      <h2>Buscar SKU</h2>
      <div class="lookup">
        <input v-model="lookupSku" placeholder="sku-demo" @keyup.enter="onLookup" />
        <BaseButton variant="secondary" @click="onLookup">Buscar</BaseButton>
      </div>
      <p v-if="lookupError" class="error">{{ lookupError }}</p>
      <p v-else-if="lookupResult">
        <strong>{{ lookupResult.sku }}</strong> — disponível: {{ lookupResult.available }},
        reservado: {{ lookupResult.reserved }}, mínimo: {{ lookupResult.minimum_quantity }}
      </p>
    </section>

    <section class="panel">
      <h2>Movimentar estoque</h2>
      <div class="form-row">
        <FormField label="SKU" v-model="form.sku">
          <input v-model="form.sku" />
        </FormField>
        <FormField label="Quantidade" v-model="form.quantity">
          <input v-model.number="form.quantity" type="number" min="1" />
        </FormField>
      </div>
      <div class="form-actions">
        <BaseButton :loading="submitting" @click="runMutation(reserveStock, 'Reserva')">
          Reservar
        </BaseButton>
        <BaseButton
          variant="secondary"
          :loading="submitting"
          @click="runMutation(releaseStock, 'Liberação')"
        >
          Liberar
        </BaseButton>
        <BaseButton
          variant="secondary"
          :loading="submitting"
          @click="runMutation(replenishStock, 'Reposição')"
        >
          Repor
        </BaseButton>
      </div>
    </section>

    <section class="panel">
      <h2>Itens com estoque baixo</h2>
      <DataTable :items="lowStock" :columns="columns" :loading="loading" empty-message="Nenhum item abaixo do mínimo.">
        <template #cell-updated_at="{ item }">{{ formatDateTime(item.updated_at) }}</template>
        <template #cell-available="{ item }">
          <StatusPill :label="String(item.available)" tone="warning" />
        </template>
      </DataTable>
    </section>
  </div>
</template>

<style scoped>
.panel {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: var(--space-4);
  margin-bottom: var(--space-5);
}

.panel h2 {
  margin-top: 0;
  font-size: 1rem;
}

.lookup {
  display: flex;
  gap: var(--space-2);
  margin-bottom: var(--space-2);
}

.lookup input {
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  flex: 1;
}

.form-row {
  display: flex;
  gap: var(--space-4);
}

.form-actions {
  display: flex;
  gap: var(--space-2);
}

.error {
  color: var(--color-danger);
}
</style>
