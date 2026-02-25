<template>
  <div class="forget-reset-page">
    <div class="forget-reset-card card">
      <h1 class="forget-reset-title"><KeyRound class="title-icon" size="32" /> 重置密码</h1>
      <template v-if="verified">
        <form @submit.prevent="onSubmit" class="forget-reset-form">
          <p v-if="username" class="username-hint">为 <strong>{{ username }}</strong> 重置密码</p>
          <div class="field">
            <label>新密码</label>
            <input v-model="newPassword" type="password" placeholder="新密码" required minlength="4" />
          </div>
          <div class="field">
            <label>确认密码</label>
            <input v-model="confirmPassword" type="password" placeholder="确认新密码" required minlength="4" />
          </div>
          <p v-if="message" class="message" :class="{ error: !ok }">{{ message }}</p>
          <button type="submit" class="btn btn-primary" :disabled="loading">提交</button>
          <p class="back-link">
            <router-link to="/login" class="btn btn-default">返回登录</router-link>
          </p>
        </form>
      </template>
      <template v-else>
        <p v-if="verifyError" class="message error">{{ verifyError }}</p>
        <p class="back-link">
          <router-link to="/forget" class="btn btn-default">重新获取链接</router-link>
          <router-link to="/login" class="btn btn-outline">去登录</router-link>
        </p>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAlert } from '../composables/useAlert'
import { KeyRound } from 'lucide-vue-next'

const API = '/api'
const route = useRoute()
const router = useRouter()
const { show: showAlert } = useAlert()
const token = computed(() => (route.query.p as string) || '')
const username = ref('')
const verified = ref(false)
const verifyError = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const message = ref('')
const ok = ref(true)
const loading = ref(false)

onMounted(async () => {
  const p = token.value
  if (!p) {
    verifyError.value = '链接无效，缺少参数'
    return
  }
  try {
    const res = await fetch(`${API}/forget/verify?p=${encodeURIComponent(p)}`, { credentials: 'include' })
    const data = await res.json()
    if (data.ok && data.username) {
      username.value = data.username
      verified.value = true
      return
    }
    verifyError.value = data.msg || '链接已过期或无效'
  } catch {
    verifyError.value = '验证失败，请重试'
  }
})

async function onSubmit() {
  if (newPassword.value !== confirmPassword.value) {
    ok.value = false
    message.value = '两次输入的密码不一致'
    return
  }
  loading.value = true
  message.value = ''
  ok.value = true
  try {
    const params = new URLSearchParams({
      p: token.value,
      new_password: newPassword.value,
    })
    const res = await fetch(`${API}/forget/reset`, {
      method: 'POST',
      body: params,
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      credentials: 'include',
    })
    const data = await res.json()
    if (data.ok) {
      showAlert(data.msg || '修改成功！', 'success')
      router.push('/login')
      return
    }
    ok.value = false
    message.value = data.msg || '重置失败'
  } catch {
    ok.value = false
    message.value = '网络错误，请重试'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.forget-reset-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--content-padding);
  background: var(--color-bg);
}
.forget-reset-card {
  width: 100%;
  max-width: 360px;
}
.forget-reset-title {
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
.username-hint {
  margin-bottom: var(--space-lg);
  font-size: 0.95rem;
  color: var(--color-text);
}
.forget-reset-form .field {
  margin-bottom: var(--space-lg);
}
.forget-reset-form button[type="submit"] {
  width: 100%;
  margin-top: var(--space-sm);
}
.back-link {
  margin-top: var(--space-lg);
  text-align: center;
}
.back-link .btn + .btn {
  margin-top: var(--space-sm);
}
.back-link .btn {
  display: block;
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
