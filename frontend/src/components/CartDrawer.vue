<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import BaseButton from '@/components/BaseButton.vue'
import { placeOrder } from '@/api/orders'
import { useCartStore } from '@/stores/cart'
import { useToastStore } from '@/stores/toast'
import { extractErrorMessage } from '@/utils/errors'
import { formatCents } from '@/utils/format'

const cart = useCartStore()
const toast = useToastStore()
const router = useRouter()
const placing = ref(false)

if (cart.items.length === 0) cart.addItem()

async function onCheckout() {
  placing.value = true
  try {
    await placeOrder({
      customer_id: cart.customerId,
      items: cart.items.map((i) => ({
        product_id: i.product_id,
        quantity: Number(i.quantity),
        price: Number(i.price),
      })),
    })
    toast.push('Pedido criado com sucesso', 'success')
    cart.clear()
    cart.close()
    router.push('/orders')
  } catch (err) {
    toast.push(extractErrorMessage(err), 'error')
  } finally {
    placing.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <div v-if="cart.isOpen" class="overlay" @click.self="cart.close()">
      <aside class="drawer">
        <header>
          <h3>Carrinho (montar pedido)</h3>
          <button class="close" @click="cart.close()">✕</button>
        </header>

        <div class="body">
          <label class="field">
            <span>Customer ID (uuid)</span>
            <input v-model="cart.customerId" placeholder="11111111-1111-1111-1111-111111111111" />
          </label>

          <div v-for="(item, index) in cart.items" :key="index" class="item-row">
            <input v-model="item.product_id" placeholder="product_id (uuid)" />
            <input v-model.number="item.quantity" type="number" min="1" placeholder="qtd" />
            <input v-model.number="item.price" type="number" min="0" placeholder="preço (centavos)" />
            <button class="remove" @click="cart.removeItem(index)" v-if="cart.items.length > 1">✕</button>
          </div>

          <BaseButton variant="secondary" @click="cart.addItem()">+ item</BaseButton>

          <p class="total">Total estimado: {{ formatCents(cart.totalCents) }}</p>
        </div>

        <footer>
          <BaseButton
            :loading="placing"
            :disabled="!cart.customerId || cart.items.some((i) => !i.product_id)"
            @click="onCheckout"
          >
            Finalizar pedido
          </BaseButton>
        </footer>
      </aside>
    </div>
  </Teleport>
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  z-index: 150;
  display: flex;
  justify-content: flex-end;
}

.drawer {
  background: var(--color-surface);
  width: min(380px, 92vw);
  height: 100%;
  display: flex;
  flex-direction: column;
  box-shadow: var(--shadow);
}

header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-4);
  border-bottom: 1px solid var(--color-border);
}

h3 {
  margin: 0;
  font-size: 1rem;
}

.close {
  border: none;
  background: none;
  cursor: pointer;
}

.body {
  padding: var(--space-4);
  overflow-y: auto;
  flex: 1;
}

.field {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  margin-bottom: var(--space-4);
  font-size: 0.85rem;
  color: var(--color-text-muted);
}

.field input {
  padding: var(--space-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
}

.item-row {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr auto;
  gap: var(--space-2);
  margin-bottom: var(--space-2);
}

.item-row input {
  padding: var(--space-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  min-width: 0;
}

.remove {
  border: none;
  background: none;
  color: var(--color-danger);
  cursor: pointer;
}

.total {
  font-weight: 600;
  margin-top: var(--space-4);
}

footer {
  padding: var(--space-4);
  border-top: 1px solid var(--color-border);
}
</style>
