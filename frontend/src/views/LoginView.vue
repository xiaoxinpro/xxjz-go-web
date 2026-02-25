<template>
  <div class="login-page">
    <div class="login-card card">
      <h1 class="login-title"><Wallet class="title-icon" size="32" /> {{ appStore.title }}</h1>
      <form @submit.prevent="onSubmit" class="login-form">
        <div class="field">
          <label>用户名</label>
          <input v-model="username" type="text" placeholder="用户名" required />
        </div>
        <div class="field">
          <label>密码</label>
          <input v-model="password" type="password" placeholder="密码" required />
        </div>
        <p v-if="message" class="message" :class="{ error: !ok }">{{ message }}</p>
        <button type="submit" class="btn btn-primary" :disabled="loading">登录</button>
        <p class="login-links">
          <router-link to="/regist" class="btn-link">注册</router-link>
          <span class="sep">|</span>
          <router-link to="/forget" class="btn-link">忘记密码</router-link>
        </p>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '../stores/user'
import { useAppStore } from '../stores/app'
import { Wallet } from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const appStore = useAppStore()
const redirectTo = computed(() => {
  const r = route.query.redirect as string
  return r && r.startsWith('/') ? r : '/home'
})
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
    router.push(redirectTo.value)
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--content-padding);
  background: var(--color-bg);
}
.login-card {
  width: 100%;
  max-width: 360px;
}
.login-title {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-md);
  margin: 0 0 var(--space-xl);
  font-size: 1.5rem;
  text-align: center;
  color: var(--color-text);
}
.title-icon {
  color: var(--color-primary);
  flex-shrink: 0;
}
.login-form .field {
  margin-bottom: var(--space-lg);
}
.login-form button[type="submit"] {
  width: 100%;
  margin-top: var(--space-sm);
}
.login-links {
  margin-top: var(--space-lg);
  text-align: center;
  font-size: 0.9rem;
}
.login-links .sep {
  margin: 0 var(--space-sm);
  color: var(--color-text-muted, #666);
}
.login-links .btn-link {
  color: var(--color-primary);
}
.message {
  margin: var(--space-sm) 0;
  font-size: 0.9rem;
  color: var(--color-success);
}
.message.error {
  color: var(--color-danger);
}
</style>
