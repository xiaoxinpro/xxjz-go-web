<template>
  <div class="user-page">
    <AppHeader />
    <main class="main">
      <div class="tabs">
        <button type="button" class="tab" :class="{ active: activeTab === 'account' }" @click="activeTab = 'account'">账号设置</button>
        <button type="button" class="tab" :class="{ active: activeTab === 'password' }" @click="activeTab = 'password'">修改密码</button>
      </div>

      <div v-if="activeTab === 'account'" class="panel">
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

      <div v-if="activeTab === 'password'" class="panel">
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

      <p class="back-link"><router-link to="/home">返回主页</router-link></p>
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
.user-page { min-height: 100vh; padding-bottom: 1rem; }
.main { padding: 1rem; max-width: 400px; margin: 0 auto; }
.tabs { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
.tab { padding: 0.5rem 1rem; border: 1px solid #ddd; background: #fff; cursor: pointer; border-radius: 4px; }
.tab.active { background: #19a7f0; color: #fff; border-color: #19a7f0; }
.panel { margin-bottom: 1rem; }
.demo-tip { color: #c00; margin-bottom: 0.5rem; font-size: 0.9rem; }
.form .field { margin-bottom: 0.5rem; }
.form .field label { display: block; margin-bottom: 0.2rem; }
.form .field input { width: 100%; max-width: 280px; }
.back-link { margin-top: 1rem; }
.back-link a { color: #19a7f0; text-decoration: none; }
.btn { margin-top: 0.5rem; }
</style>
