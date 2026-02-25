<template>
  <div class="transfer-page">
    <AppHeader />
    <main class="main page-main">
      <div class="tabs card">
        <router-link to="/add" class="tab">
          <ArrowDownCircle size="20" class="tab-icon" /> 支出
        </router-link>
        <router-link to="/add" class="tab">
          <ArrowUpCircle size="20" class="tab-icon" /> 收入
        </router-link>
        <router-link to="/add/transfer" class="tab" :class="{ active: $route.path === '/add/transfer' }">
          <ArrowLeftRight size="20" class="tab-icon" /> 转账
        </router-link>
      </div>

      <div class="form-card card">
        <h2 class="card-title"><ArrowRightLeft size="22" class="title-icon" /> 账户转账</h2>
        <form class="form" @submit.prevent="onSubmit">
          <div class="field">
            <label>转账金额</label>
            <input v-model="form.money" type="number" step="0.01" min="0" placeholder="输入转账金额" required />
          </div>
          <div class="field">
            <label>转出账户</label>
            <select v-model="form.source_fid" required>
              <option value="" disabled>选择转出账户</option>
              <option v-for="f in funds" :key="f.fundsid" :value="f.fundsid">{{ f.fundsname }}</option>
            </select>
          </div>
          <div class="field">
            <label>转入账户</label>
            <select v-model="form.target_fid" required>
              <option value="" disabled>选择转入账户</option>
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
import { ArrowRightLeft, ArrowDownCircle, ArrowUpCircle, ArrowLeftRight } from 'lucide-vue-next'

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
  money: '' as number | string,
  source_fid: '' as number | string,
  target_fid: '' as number | string,
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
  const moneyNum = Number(form.value.money)
  if (!Number.isFinite(moneyNum) || moneyNum <= 0) {
    alert('请输入有效转账金额')
    return
  }
  const src = form.value.source_fid === '' ? NaN : Number(form.value.source_fid)
  const tgt = form.value.target_fid === '' ? NaN : Number(form.value.target_fid)
  if (!Number.isFinite(src) || !Number.isFinite(tgt)) {
    alert('请选择转出账户和转入账户')
    return
  }
  if (src === tgt) {
    alert('转出账户与转入账户不能相同')
    return
  }
  const payload = {
    money: moneyNum,
    source_fid: src,
    target_fid: tgt,
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
        form.value.money = ''
        form.value.source_fid = ''
        form.value.target_fid = ''
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
.tabs.card {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-sm);
  margin-bottom: var(--space-lg);
}
.tab {
  display: inline-flex;
  align-items: center;
  gap: var(--space-xs);
  padding: var(--space-md) var(--space-lg);
  border: 1px solid var(--color-border);
  background: var(--color-bg-card);
  border-radius: var(--radius-md);
  text-decoration: none;
  color: var(--color-text);
  cursor: pointer;
  font-size: 0.95rem;
  min-height: var(--touch-min);
}
.tab:hover {
  background: var(--color-bg);
}
.tab.active {
  background: var(--color-primary);
  color: #fff;
  border-color: var(--color-primary);
}
.tab-icon {
  flex-shrink: 0;
}

.form button,
.form .btn {
  width: 100%;
  margin-bottom: var(--space-sm);
}
</style>
