import { defineStore } from 'pinia'
import type { OrderItemDTO } from '@/types/order'

export const useCartStore = defineStore('cart', {
  state: () => ({
    customerId: '',
    items: [] as OrderItemDTO[],
    isOpen: false,
  }),
  getters: {
    count: (state) => state.items.length,
    totalCents: (state) => state.items.reduce((sum, i) => sum + i.price * i.quantity, 0),
  },
  actions: {
    open() {
      this.isOpen = true
    },
    close() {
      this.isOpen = false
    },
    toggle() {
      this.isOpen = !this.isOpen
    },
    addItem() {
      this.items.push({ product_id: '', quantity: 1, price: 0 })
    },
    removeItem(index: number) {
      this.items.splice(index, 1)
    },
    clear() {
      this.customerId = ''
      this.items = []
    },
  },
})
