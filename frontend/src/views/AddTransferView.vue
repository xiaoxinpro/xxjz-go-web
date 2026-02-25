<template>
  <div class="transfer-page">
    <AppHeader />
    <main class="main page-main">
      <div class="form-card card">
        <h2 class="card-title"><ArrowRightLeft size="22" class="title-icon" /> 账户转账</h2>
        <form class="form" @submit.prevent="onSubmit">
          <div class="field">
            <label>转账金额</label>
            <input v-model.number="form.money" type="number" step="0.01" min="0" placeholder="输入转账金额" required />
          </div>
          <div class="field">
            <label>转出账户</label>
            <select v-model.number="form.source_fid" required>
              <option value="">选择转出账户</option>
              <option v-for="f in funds" :key="f.fundsid" :value="f.fundsid">{{ f.fundsname }}</option>
            </select>
          </div>
          <div class="field">
            <label>转入账户</label>
            <select v-model.number="form.target_fid" required>
              <option value="">选择转入账户</option>
              <option v-for="f in funds" :key="f.fundsid" :value="f.fundsid">{{ f.fundsname }}</option>
            </select>
          </div>
          <div class="field">
            <label>备注</label>
            <input v-model.trim="form.mark" type="text" placeholder="备注" />
          </div>
          <div class="field">
            <label>时间</label>
            <input v-model="form.time" type="date" required />
          </div>
          <button type="submit" class="btn btn-primary">转账</button>
          <router-link to="/add" class="btn btn-default">返回</router-link>
        </form>
      </div>
    </main>
    <NavBars current="add" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useUserStore } from '../stores/user'
import AppHeader from '../components/AppHeader.vue'
import NavBars from '../components/NavBars.vue'
import { ArrowRightLeft } from 'lucide-vue-next'

const API = '/api'

function base64Json (obj: Record<string, unknown>): string {
  const s = JSON.stringify(obj)
  const bytes = new TextEncoder().encode(s)
  let binary = ''
  for (let i = 0; i < bytes.length; i++) binary += String.fromCharCode(bytes[i])
  return btoa(binary)
}

interface Fund { fundsid: number; fundsname: string }
const funds = ref<Fund[]>([])
const form = ref({
  money: 0 as number,
  source_fid: 0 as number,
  target_fid: 0 as number,
  mark: '',
  time: '',
})

function loadFunds () {
  fetch(`${API}/funds?type=get`, { credentials: 'include' })
    .then(res => res.json())
    .then((data: { uid: number; data: unknown }) => {
      if (data.uid <= 0) return
      funds.value = Array.isArray(data.data) ? data.data as Fund[] : []
    })
    .catch(() => {})
}

function onSubmit () {
  if (form.value.source_fid === form.value.target_fid) {
    alert('转出账户与转入账户不能相同')
    return
  }
  const payload = {
    money: Number(form.value.money),
    source_fid: form.value.source_fid,
    target_fid: form.value.target_fid,
    mark: form.value.mark || '',
    time: form.value.time || new Date().toISOString().slice(0, 10),
  }
  const body = new URLSearchParams()
  body.set('type', 'add')
  body.set('data', base64Json(payload))
  fetch(`${API}/transfer`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: body.toString(),
  })
    .then(res => res.json())
    .then((data: { uid: number; data: { ret?: boolean; msg?: string } }) => {
      if (data.uid <= 0) {
        alert('请先登录')
        return
      }
      const d = data.data as { ret?: boolean; msg?: string }
      if (d && d.ret) {
        alert(d.msg || '转账成功')
        form.value.money = 0
        form.value.mark = ''
        form.value.time = new Date().toISOString().slice(0, 10)
      } else {
        alert(d && d.msg ? d.msg : '转账失败')
      }
    })
    .catch(() => alert('请求失败'))
}

onMounted(() => {
  form.value.time = new Date().toISOString().slice(0, 10)
  loadFunds()
})
</script>

<style scoped>
.transfer-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
.main {
  flex: 1;
  padding-bottom: 4.5rem;
}
.form-card .card-title {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}
.title-icon {
  color: var(--color-primary);
}
.form button,
.form .btn {
  width: 100%;
  margin-bottom: var(--space-sm);
}
</style>
