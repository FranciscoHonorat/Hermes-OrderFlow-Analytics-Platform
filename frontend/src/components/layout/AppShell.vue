<script setup lang="ts">
import CartDrawer from '@/components/CartDrawer.vue'
import ToastContainer from '@/components/ToastContainer.vue'
import SideNav from '@/components/layout/SideNav.vue'
import TopHeader from '@/components/layout/TopHeader.vue'
</script>

<template>
  <div class="shell">
    <TopHeader />
    <div class="body">
      <SideNav />
      <main class="content">
        <RouterView />
      </main>
    </div>
    <CartDrawer />
    <ToastContainer />
  </div>
</template>

<style scoped>
.shell {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.body {
  flex: 1;
  display: flex;
}

.content {
  flex: 1;
  padding: var(--space-6);
  overflow-y: auto;
}
</style>
