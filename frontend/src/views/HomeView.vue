<template>
  <div class="home-page">
    <header class="header">
      <span>小歆记账</span>
      <span class="user">{{ userStore.username || '未登录' }}</span>
    </header>
    <main class="main">
      <p>欢迎使用小歆记账。</p>
      <button @click="logout" class="btn-outline">退出登录</button>
    </main>
  </div>
</template>

<script setup lang="ts">
import { useUserStore } from '../stores/user'
import { useRouter } from 'vue-router'

const userStore = useUserStore()
const router = useRouter()

function logout() {
  userStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.home-page { min-height: 100vh; display: flex; flex-direction: column; }
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  background: #fff;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
}
.user { font-size: 0.9rem; color: #666; }
.main { flex: 1; padding: 1.5rem; }
.btn-outline {
  padding: 0.5rem 1rem;
  border: 1px solid #ddd;
  background: #fff;
  border-radius: 6px;
  cursor: pointer;
}
</style>
