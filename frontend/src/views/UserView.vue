<template>
  <div class="user-page">
    <AppHeader />
    <main class="main page-main container">
      <div class="tabs card">
        <button type="button" class="tab" :class="{ active: activeTab === 'account' }" @click="activeTab = 'account'">账号设置</button>
        <button type="button" class="tab" :class="{ active: activeTab === 'password' }" @click="activeTab = 'password'">修改密码</button>
      </div>

      <div v-if="activeTab === 'account'" class="panel card">
        <p v-if="isDemo" class="demo-tip">Demo 账号无法修改账号信息。</p>
        <form class="form" @submit.prevent="onUpdateUsername">
          <div class="field">
            <label>登录账号</label>
            <input v-model.trim="accountForm.username" type="text" :disabled="isDemo" required />
          </div>
          <div class="field">
            <label>邮箱</label>
            <input v-model.trim="accountForm.email" type="text" :disabled="isDemo" />
          </div>
          <div class="field">
            <label>验证密码</label>
            <input v-model="accountForm.password" type="password" placeholder="当前密码" :disabled="isDemo" />
          </div>
          <button type="submit" class="btn btn-primary" :disabled="isDemo">保存</button>
        </form>
      </div>

      <div v-if="activeTab === 'password'" class="panel card">
        <p v-if="isDemo" class="demo-tip">Demo 账号无法修改密码。</p>
        <form class="form" @submit.prevent="onUpdatePassword">
          <div class="field">
            <label>旧密码</label>
            <input v-model="pwdForm.old" type="password" :disabled="isDemo" required />
          </div>
          <div class="field">
            <label>新密码</label>
            <input v-model="pwdForm.new" type="password" :disabled="isDemo" required />
          </div>
          <div class="field">
            <label>确认新密码</label>
            <input v-model="pwdForm.confirm" type="password" :disabled="isDemo" required />
          </div>
          <button type="submit" class="btn btn-primary" :disabled="isDemo">修改密码</button>
        </form>
      </div>

      <p class="back-link"><router-link to="/home" class="btn btn-default">返回</router-link></p>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import AppHeader from '../components/AppHeader.vue'
import { useUserStore } from '../stores/user'

const API = '/api'
const router = useRouter()
const userStore = useUserStore()

const activeTab = ref<'account' | 'password'>('account')
const demoUsername = ref('')
const accountForm = ref({ username: '', email: '', password: '' })
const pwdForm = ref({ old: '', new: '', confirm: '' })

const isDemo = computed(() => {
  return !!demoUsername.value && userStore.username === demoUsername.value
})

function base64Json(obj: Record<string, unknown>): string {
  const s = JSON.stringify(obj)
  const bytes = new TextEncoder().encode(s)
  let binary = ''
  for (let i = 0; i < bytes.length; i++) binary += String.fromCharCode(bytes[i])
  return btoa(binary)
}

async function loadVersion() {
  try {
    const res = await fetch(`${API}/version`, { credentials: 'include' })
    const data = await res.json()
    if (data.demo && data.demo.username) demoUsername.value = data.demo.username
  } catch {
    /* ignore */
  }
}

async function loadUser() {
  const res = await fetch(`${API}/user?type=get&uid=${userStore.uid}`, { credentials: 'include' })
  const data = await res.json()
  if (data.username) accountForm.value.username = data.username
  if (data.email != null) accountForm.value.email = data.email
}

async function onUpdateUsername() {
  if (isDemo.value) return
  const body = new URLSearchParams()
  body.set('type', 'updataUsername')
  body.set('uid', String(userStore.uid))
  body.set('data', base64Json({ username: accountForm.value.username, email: accountForm.value.email, password: accountForm.value.password }))
  const res = await fetch(API + '/user', { method: 'POST', body, credentials: 'include' })
  const data = await res.json()
  if (data.uid > 0) {
    userStore.username = data.username || accountForm.value.username
    alert('保存成功，请重新登录后生效。')
    userStore.logout()
    router.push('/login')
  } else {
    alert(data.username || '保存失败')
  }
}

async function onUpdatePassword() {
  if (isDemo.value) return
  if (pwdForm.value.new !== pwdForm.value.confirm) {
    alert('两次输入的新密码不一致')
    return
  }
  const body = new URLSearchParams()
  body.set('type', 'updataPassword')
  body.set('uid', String(userStore.uid))
  body.set('data', base64Json({ old: pwdForm.value.old, new: pwdForm.value.new }))
  const res = await fetch(API + '/user', { method: 'POST', body, credentials: 'include' })
  const data = await res.json()
  if (data.uid > 0) {
    alert('密码已修改，请重新登录。')
    userStore.logout()
    router.push('/login')
  } else {
    alert(data.username || '修改失败')
  }
}

onMounted(() => {
  loadVersion()
  loadUser()
})
</script>

<style scoped>
.user-page { min-height: 100vh; padding-bottom: var(--space-xl); }
.tabs.card { display: flex; flex-wrap: wrap; gap: var(--space-sm); margin-bottom: var(--space-lg); }
.tab {
  padding: var(--space-md) var(--space-lg);
  border: 1px solid var(--color-border);
  background: var(--color-bg-card);
  cursor: pointer;
  border-radius: var(--radius-md);
  min-height: var(--touch-min);
  font-size: 0.95rem;
}
.tab.active { background: var(--color-primary); color: #fff; border-color: var(--color-primary); }
.panel { margin-bottom: var(--space-lg); }
.demo-tip { color: var(--color-danger); margin-bottom: var(--space-md); font-size: 0.9rem; }
.form .field label { display: block; }
.form .field input { width: 100%; max-width: 320px; }
.back-link { margin-top: var(--space-lg); }
.btn { margin-top: var(--space-sm); }
</style>
