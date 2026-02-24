<template>
  <nav class="nav-bars">
    <router-link to="/home" class="nav-item" :class="{ active: current === 'home' }">
      <span class="nav-icon">主页</span>
      <span class="nav-label">主页</span>
    </router-link>
    <router-link to="/add" class="nav-item" :class="{ active: current === 'add' }">
      <span class="nav-icon">记账</span>
      <span class="nav-label">记账</span>
    </router-link>
    <router-link to="/find" class="nav-item" :class="{ active: current === 'find' }">
      <span class="nav-icon">查询</span>
      <span class="nav-label">查询</span>
    </router-link>
  </nav>
</template>

<script setup lang="ts">
defineProps<{ current: 'home' | 'add' | 'find' }>()
</script>

<style scoped>
.nav-bars {
  display: flex;
  justify-content: space-around;
  align-items: center;
  height: 56px;
  background: #19a7f0;
  color: #fff;
}
.nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0.25rem 1rem;
  color: rgba(255,255,255,0.85);
  text-decoration: none;
  font-size: 0.9rem;
}
.nav-item.active { color: #fff; font-weight: 600; }
.nav-icon { margin-bottom: 2px; }
.nav-label { font-size: 0.75rem; }
</style>
