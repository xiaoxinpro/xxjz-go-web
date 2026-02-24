<template>
  <div class="init-page">
    <div class="init-card">
      <h1>小歆记账 · 初始化</h1>
      <p class="intro">首次使用请创建管理员账号，或导入旧数据库。</p>

      <section class="section">
        <h2>创建管理员账号</h2>
        <form @submit.prevent="onSetup" class="form">
          <input v-model="setup.username" type="text" placeholder="用户名" required />
          <input v-model="setup.password" type="password" placeholder="密码" required />
          <input v-model="setup.email" type="email" placeholder="邮箱" required />
          <p v-if="setup.message" class="message" :class="{ error: !setup.ok }">{{ setup.message }}</p>
          <button type="submit" :disabled="setup.loading">创建并进入</button>
        </form>
      </section>

      <div class="divider">或</div>

      <section class="section">
        <h2>导入旧数据库</h2>
        <p class="hint">上传从旧版（ThinkPHP/MySQL）导出的 xxjz.sql 文件，将自动转换为当前数据库。</p>
        <form @submit.prevent="onImport" class="form">
          <input ref="fileInput" type="file" accept=".sql" @change="onFileChange" />
          <p v-if="importForm.message" class="message" :class="{ error: !importForm.ok }">{{ importForm.message }}</p>
          <button type="submit" :disabled="importForm.loading || !importForm.file">导入</button>
        </form>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'

const API = '/api'
const router = useRouter()
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
  padding: 1rem;
  background: #f5f5f5;
}
.init-card {
  background: #fff;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.08);
  width: 100%;
  max-width: 420px;
}
.init-card h1 {
  margin: 0 0 0.5rem;
  font-size: 1.35rem;
  text-align: center;
  color: #333;
}
.intro {
  margin: 0 0 1.5rem;
  font-size: 0.9rem;
  color: #666;
  text-align: center;
}
.section {
  margin-bottom: 1.5rem;
}
.section h2 {
  margin: 0 0 0.75rem;
  font-size: 1rem;
  color: #333;
}
.hint {
  margin: 0 0 0.75rem;
  font-size: 0.85rem;
  color: #666;
}
.form input[type="text"],
.form input[type="password"],
.form input[type="email"] {
  width: 100%;
  padding: 0.75rem 1rem;
  margin-bottom: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 8px;
  font-size: 1rem;
}
.form input[type="file"] {
  margin-bottom: 0.75rem;
  font-size: 0.9rem;
}
.form button {
  width: 100%;
  padding: 0.75rem;
  background: #198754;
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  cursor: pointer;
}
.form button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}
.divider {
  text-align: center;
  color: #999;
  font-size: 0.9rem;
  margin: 1rem 0;
}
.message { margin: 0.5rem 0; font-size: 0.9rem; color: #198754; }
.message.error { color: #dc3545; }
</style>
