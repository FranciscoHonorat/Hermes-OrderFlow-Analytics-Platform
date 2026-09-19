<script setup lang="ts">
import { useToastStore } from '@/stores/toast'

const toastStore = useToastStore()
</script>

<template>
  <div class="toasts">
    <div
      v-for="toast in toastStore.toasts"
      :key="toast.id"
      class="toast"
      :class="toast.variant"
      @click="toastStore.dismiss(toast.id)"
    >
      {{ toast.message }}
    </div>
  </div>
</template>

<style scoped>
.toasts {
  position: fixed;
  top: var(--space-4);
  right: var(--space-4);
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  z-index: 200;
}

.toast {
  padding: var(--space-3) var(--space-4);
  border-radius: var(--radius);
  color: white;
  box-shadow: var(--shadow);
  cursor: pointer;
  max-width: 320px;
}

.success {
  background: var(--color-success);
}
.error {
  background: var(--color-danger);
}
.info {
  background: var(--color-primary);
}
</style>
