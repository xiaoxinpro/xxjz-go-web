<template>
  <div class="init-page">
    <div class="init-card card">
      <h1 class="init-title">{{ appStore.title }} · 初始化</h1>
      <p class="intro">首次使用请创建管理员账号，或导入旧数据库。</p>

      <section class="section card">
        <h2 class="card-title"><UserPlus class="section-icon" size="20" /> 创建管理员账号</h2>
        <form @submit.prevent="onSetup" class="form">
          <div class="field">
            <label>用户名</label>
            <input v-model="setup.username" type="text" placeholder="用户名" required />
          </div>
          <div class="field">
            <label>密码</label>
            <input v-model="setup.password" type="password" placeholder="密码" required />
          </div>
          <div class="field">
            <label>邮箱</label>
            <input v-model="setup.email" type="email" placeholder="邮箱" required />
          </div>
          <p v-if="setup.message" class="message" :class="{ error: !setup.ok }">{{ setup.message }}</p>
          <button type="submit" class="btn btn-primary" :disabled="setup.loading">创建并进入</button>
        </form>
      </section>

      <div class="divider">或</div>

      <section class="section card">
        <h2 class="card-title"><Upload class="section-icon" size="20" /> 导入旧数据库</h2>
        <p class="hint">上传从旧版（ThinkPHP/MySQL）导出的 xxjz.sql 文件，将自动转换为当前数据库。</p>
        <form @submit.prevent="onImport" class="form">
          <div class="field">
            <label>选择 .sql 文件</label>
            <input ref="fileInput" type="file" accept=".sql" @change="onFileChange" />
          </div>
          <p v-if="importForm.message" class="message" :class="{ error: !importForm.ok }">{{ importForm.message }}</p>
          <button type="submit" class="btn btn-primary" :disabled="importForm.loading || !importForm.file">导入</button>
        </form>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '../stores/app'
import { UserPlus, Upload } from 'lucide-vue-next'

const API = '/api'
const router = useRouter()
const appStore = useAppStore()
const fileInput = ref<HTMLInputElement | null>(null)

const setup = reactive({
  username: '',
  password: '',
  email: '',
  message: '',
  ok: true,
  loading: false,
})

const importForm = reactive({
  file: null as File | null,
  message: '',
  ok: true,
  loading: false,
})

async function onSetup() {
  setup.loading = true
  setup.message = ''
  try {
    const res = await fetch(`${API}/init/setup`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username: setup.username,
        password: setup.password,
        email: setup.email,
      }),
      credentials: 'include',
    })
    const data = await res.json()
    setup.ok = data.ok
    setup.message = data.msg || (data.ok ? '创建成功' : '创建失败')
    if (data.ok) {
      router.push('/login')
    }
  } catch (e) {
    setup.ok = false
    setup.message = '网络错误'
  }
  setup.loading = false
}

function onFileChange(e: Event) {
  const target = e.target as HTMLInputElement
  importForm.file = target.files?.[0] ?? null
  importForm.message = ''
}

async function onImport() {
  if (!importForm.file) return
  importForm.loading = true
  importForm.message = ''
  try {
    const formData = new FormData()
    formData.append('file', importForm.file)
    const res = await fetch(`${API}/init/import`, {
      method: 'POST',
      body: formData,
      credentials: 'include',
    })
    const data = await res.json()
    importForm.ok = data.ok
    importForm.message = data.msg || (data.ok ? '导入成功' : '导入失败')
    if (data.ok) {
      if (data.initialized) {
        router.push('/login')
      } else {
        importForm.message = '导入成功，请刷新页面后登录'
        const statusRes = await fetch(`${API}/init/status`, { credentials: 'include' })
        const statusData = await statusRes.json().catch(() => ({}))
        if (statusData.initialized) router.push('/login')
      }
    }
  } catch (e) {
    importForm.ok = false
    importForm.message = '网络错误'
  }
  importForm.loading = false
}
</script>

<style scoped>
.init-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--content-padding);
  background: var(--color-bg);
}
.init-card {
  width: 100%;
  max-width: 420px;
}
.init-title {
  margin: 0 0 0.5rem;
  font-size: 1.35rem;
  text-align: center;
  color: var(--color-text);
}
.intro {
  margin: 0 0 var(--space-xl);
  font-size: 0.9rem;
  color: var(--color-text-muted);
  text-align: center;
}
.section {
  margin-bottom: var(--space-xl);
}
.section .card-title {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}
.section-icon {
  flex-shrink: 0;
  color: var(--color-primary);
}
.hint {
  margin: 0 0 var(--space-md);
  font-size: 0.85rem;
  color: var(--color-text-muted);
}
.form .field input[type="file"] {
  font-size: 0.9rem;
  min-height: auto;
  padding: var(--space-sm);
}
.form button {
  width: 100%;
  margin-top: var(--space-sm);
}
.divider {
  text-align: center;
  color: var(--color-text-light);
  font-size: 0.9rem;
  margin: var(--space-lg) 0;
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
