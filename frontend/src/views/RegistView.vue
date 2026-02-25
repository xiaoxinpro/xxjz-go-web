<template>
  <div class="regist-page">
    <div class="regist-card card">
      <h1 class="regist-title"><UserPlus class="title-icon" size="32" /> 注册</h1>
      <form @submit.prevent="onSubmit" class="regist-form">
        <div class="field">
          <label>用户名</label>
          <input v-model="username" type="text" placeholder="用户名" required minlength="2" />
        </div>
        <div class="field">
          <label>密码</label>
          <input v-model="password" type="password" placeholder="密码" required minlength="4" />
        </div>
        <div class="field">
          <label>邮箱</label>
          <input v-model="email" type="email" placeholder="邮箱" required />
        </div>
        <p v-if="message" class="message" :class="{ error: !ok }">{{ message }}</p>
        <button type="submit" class="btn btn-primary" :disabled="loading">注册</button>
        <p class="back-link">
          <router-link to="/login" class="btn btn-default">返回登录</router-link>
        </p>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAlert } from '../composables/useAlert'
import { UserPlus } from 'lucide-vue-next'

const API = '/api'
const router = useRouter()
const { show: showAlert } = useAlert()
const username = ref('')
const password = ref('')
const email = ref('')
const message = ref('')
const ok = ref(true)
const loading = ref(false)

async function onSubmit() {
  loading.value = true
  message.value = ''
  ok.value = true
  try {
    const params = new URLSearchParams({
      username: username.value,
      password: password.value,
      email: email.value,
    })
    const res = await fetch(`${API}/regist`, {
      method: 'POST',
      body: params,
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      credentials: 'include',
    })
    const data = await res.json()
    const uid = Number(data.uid)
    if (uid > 0) {
      showAlert(data.msg || '注册成功！', 'success')
      router.push('/login')
      return
    }
    ok.value = false
    message.value = data.msg || '注册失败'
  } catch {
    ok.value = false
    message.value = '网络错误，请重试'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.regist-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--content-padding);
  background: var(--color-bg);
}
.regist-card {
  width: 100%;
  max-width: 360px;
}
.regist-title {
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
.regist-form .field {
  margin-bottom: var(--space-lg);
}
.regist-form button[type="submit"] {
  width: 100%;
  margin-top: var(--space-sm);
}
.regist-form .back-link {
  margin-top: var(--space-lg);
  text-align: center;
}
.regist-form .back-link .btn {
  width: 100%;
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
