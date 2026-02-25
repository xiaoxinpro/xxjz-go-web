<template>
  <div class="forget-page">
    <div class="forget-card card">
      <h1 class="forget-title"><KeyRound class="title-icon" size="32" /> 找回密码</h1>
      <form v-if="!sent" @submit.prevent="onSubmit" class="forget-form">
        <div class="field">
          <label>邮箱</label>
          <input v-model="email" type="email" placeholder="输入注册时使用的邮箱" required />
        </div>
        <p v-if="message" class="message" :class="{ error: !ok }">{{ message }}</p>
        <button type="submit" class="btn btn-primary" :disabled="loading">发送重置链接</button>
        <p class="back-link">
          <router-link to="/login" class="btn btn-default">返回登录</router-link>
        </p>
      </form>
      <div v-else class="forget-done">
        <p class="message success">{{ message }}</p>
        <p class="back-link">
          <router-link to="/login" class="btn btn-primary">去登录</router-link>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { KeyRound } from 'lucide-vue-next'

const API = '/api'
const email = ref('')
const message = ref('')
const ok = ref(true)
const loading = ref(false)
const sent = ref(false)

async function onSubmit() {
  loading.value = true
  message.value = ''
  ok.value = true
  try {
    const params = new URLSearchParams({ forget_email: email.value })
    const res = await fetch(`${API}/forget/request`, {
      method: 'POST',
      body: params,
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      credentials: 'include',
    })
    const data = await res.json()
    const uid = data.uid
    const msg = data.msg || ''
    if (uid === true || uid === 'true') {
      sent.value = true
      message.value = msg || '找回密码的链接已发送至您的邮箱，请查收！'
      return
    }
    ok.value = false
    message.value = msg || '发送失败'
  } catch {
    ok.value = false
    message.value = '网络错误，请重试'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.forget-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--content-padding);
  background: var(--color-bg);
}
.forget-card {
  width: 100%;
  max-width: 360px;
}
.forget-title {
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
.forget-form .field {
  margin-bottom: var(--space-lg);
}
.forget-form button[type="submit"] {
  width: 100%;
  margin-top: var(--space-sm);
}
.forget-form .back-link,
.forget-done .back-link {
  margin-top: var(--space-lg);
  text-align: center;
}
.forget-form .back-link .btn,
.forget-done .back-link .btn {
  width: 100%;
}
.forget-done .message.success {
  color: var(--color-success);
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
