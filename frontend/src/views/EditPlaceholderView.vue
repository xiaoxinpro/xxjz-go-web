<template>
  <div class="placeholder-page">
    <AppHeader />
    <main class="main page-main">
      <div class="card">
        <p class="muted">编辑转账功能开发中。</p>
        <p class="back-link"><router-link to="/home" class="btn btn-default">返回</router-link></p>
      </div>
    </main>
    <NavBars current="home" />
  </div>
</template>

<script setup lang="ts">
import AppHeader from '../components/AppHeader.vue'
import NavBars from '../components/NavBars.vue'
</script>

<style scoped>
.placeholder-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
.main {
  flex: 1;
  padding-bottom: 4.5rem;
}
.back-link {
  margin-top: var(--space-lg);
}
</style>
