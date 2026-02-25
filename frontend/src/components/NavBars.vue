<template>
  <nav class="nav-bars">
    <router-link to="/home" class="nav-item" :class="{ active: current === 'home' }">
      <Home class="nav-icon" size="22" />
      <span class="nav-label">主页</span>
    </router-link>
    <router-link to="/add" class="nav-item" :class="{ active: current === 'add' }">
      <PenLine class="nav-icon" size="22" />
      <span class="nav-label">记账</span>
    </router-link>
    <router-link to="/find" class="nav-item" :class="{ active: current === 'find' }">
      <Search class="nav-icon" size="22" />
      <span class="nav-label">查询</span>
    </router-link>
  </nav>
</template>

<script setup lang="ts">
import { Home, PenLine, Search } from 'lucide-vue-next'
defineProps<{ current: 'home' | 'add' | 'find' }>()
</script>

<style scoped>
.nav-bars {
  display: flex;
  justify-content: space-around;
  align-items: center;
  height: 56px;
  min-height: var(--touch-min, 56px);
  background: var(--color-primary);
  color: #fff;
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 90;
}
.nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-xs) var(--space-lg);
  min-height: var(--touch-min, 44px);
  color: rgba(255, 255, 255, 0.85);
  text-decoration: none;
  font-size: 0.9rem;
}
.nav-item.active {
  color: #fff;
  font-weight: 600;
}
.nav-icon {
  margin-bottom: 2px;
  flex-shrink: 0;
}
.nav-label {
  font-size: 0.75rem;
}
</style>
