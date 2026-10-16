<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import BaseButton from '@/components/BaseButton.vue'
import FormField from '@/components/FormField.vue'
import { useAuthStore } from '@/stores/auth'
import { extractErrorMessage } from '@/utils/errors'

const auth = useAuthStore()
const router = useRouter()

const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

async function onSubmit() {
  loading.value = true
  error.value = ''
  try {
    await auth.login(email.value, password.value)
    router.push('/')
  } catch (err) {
    error.value = extractErrorMessage(err)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <form class="login-card" @submit.prevent="onSubmit">
      <h1>Hermes OrderFlow</h1>
      <p class="subtitle">Entre com sua conta do user-service</p>

      <FormField label="Email" v-model="email">
        <input v-model="email" type="email" required autocomplete="username" />
      </FormField>
      <FormField label="Senha" v-model="password">
        <input v-model="password" type="password" required autocomplete="current-password" />
      </FormField>

      <p v-if="error" class="error">{{ error }}</p>

      <BaseButton type="submit" :loading="loading">Entrar</BaseButton>
    </form>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
}

.login-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  padding: var(--space-6);
  width: min(360px, 90vw);
}

h1 {
  margin: 0 0 var(--space-1);
  font-size: 1.3rem;
}

.subtitle {
  color: var(--color-text-muted);
  margin: 0 0 var(--space-5);
  font-size: 0.9rem;
}

.error {
  color: var(--color-danger);
  font-size: 0.85rem;
}
</style>
