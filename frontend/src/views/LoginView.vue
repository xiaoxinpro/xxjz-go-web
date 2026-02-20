<template>
  <div class="login-page">
    <div class="login-card">
      <h1>小歆记账</h1>
      <form @submit.prevent="onSubmit" class="login-form">
        <input v-model="username" type="text" placeholder="用户名" required />
        <input v-model="password" type="password" placeholder="密码" required />
        <p v-if="message" class="message" :class="{ error: !ok }">{{ message }}</p>
        <button type="submit" :disabled="loading">登录</button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()
const username = ref('')
const password = ref('')
const message = ref('')
const ok = ref(true)
const loading = ref(false)

async function onSubmit() {
  loading.value = true
  message.value = ''
  const res = await userStore.login(username.value, password.value)
  ok.value = res.ok
  message.value = res.message
  loading.value = false
  if (res.ok) {
    router.push('/home')
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}
.login-card {
  background: #fff;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.08);
  width: 100%;
  max-width: 360px;
}
.login-card h1 {
  margin: 0 0 1.5rem;
  font-size: 1.5rem;
  text-align: center;
  color: #333;
}
.login-form input {
  width: 100%;
  padding: 0.75rem 1rem;
  margin-bottom: 1rem;
  border: 1px solid #ddd;
  border-radius: 8px;
  font-size: 1rem;
}
.login-form button {
  width: 100%;
  padding: 0.75rem;
  background: #198754;
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  cursor: pointer;
}
.login-form button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}
.message { margin: 0.5rem 0; font-size: 0.9rem; color: #198754; }
.message.error { color: #dc3545; }
</style>
