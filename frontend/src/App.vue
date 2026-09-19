<script setup lang="ts">
import { useRoute } from 'vue-router'
import AppShell from '@/components/layout/AppShell.vue'
import ToastContainer from '@/components/ToastContainer.vue'
import TopHeader from '@/components/layout/TopHeader.vue'

const route = useRoute()
</script>

<template>
  <template v-if="route.meta.public">
    <TopHeader />
    <RouterView />
    <ToastContainer />
  </template>
  <AppShell v-else />
</template>
