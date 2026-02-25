<template>
  <div class="add-page">
    <header class="header">
      <span>小歆记账</span>
    </header>
    <main class="main">
      <div class="tabs">
        <button type="button" class="tab" :class="{ active: activeTab === 'out' }" @click="activeTab = 'out'">支出</button>
        <button type="button" class="tab" :class="{ active: activeTab === 'in' }" @click="activeTab = 'in'">收入</button>
        <router-link v-if="funds.length > 1" to="/add/transfer" class="tab">转账</router-link>
      </div>

      <form class="form" @submit.prevent="onSubmit">
        <fieldset>
          <legend>{{ activeTab === 'out' ? '添加支出' : '添加收入' }}</legend>
          <div class="field">
            <label>金额</label>
            <input v-model.number="form.money" type="number" step="0.01" min="0" placeholder="输入金额" required />
          </div>
          <div v-if="funds.length > 1" class="field">
            <label>账户</label>
            <select v-model.number="form.fid" required>
              <option v-for="f in funds" :key="f.fundsid" :value="f.fundsid">{{ f.fundsname }}</option>
            </select>
          </div>
          <div v-else class="field" style="display:none">
            <input v-model="form.fid" type="hidden" :value="funds[0]?.fundsid ?? -1" />
          </div>
          <div class="field">
            <label>分类</label>
            <select v-model.number="form.acclassid" required>
              <option v-for="(name, id) in classOptions" :key="String(id)" :value="Number(id)">{{ name }}</option>
            </select>
          </div>
          <div class="field">
            <label>备注</label>
            <div class="remark-row">
              <input v-model.trim="form.acremark" type="text" placeholder="备注" class="remark-input" />
              <span class="upload-wrap">
                <button type="button" class="btn btn-upload" @click="triggerFileInput">上传图片</button>
                <input ref="fileInput" type="file" accept=".jpg,.jpeg,.png,.gif" multiple class="hidden-input" @change="onFileSelect" />
              </span>
            </div>
            <p v-if="pendingFiles.length > 0" class="file-hint">已选 {{ pendingFiles.length }} 张图片，提交后将随记账一起上传</p>
          </div>
          <div class="field">
            <label>时间</label>
            <input v-model="form.actime" type="date" required />
          </div>
          <p><button type="submit" class="btn btn-primary">添加</button></p>
          <p><router-link to="/home" class="btn btn-default">返回</router-link></p>
        </fieldset>
      </form>

      <div class="list-section">
        <table class="list-table">
          <thead>
            <tr><th>分类</th><th>金额</th><th>收支</th><th>时间</th><th>备注</th></tr>
          </thead>
          <tbody>
            <tr v-for="row in listData" :key="row.classid === 0 ? 't' + row.id : 'a' + row.id">
              <td>{{ row.class }}</td>
              <td :class="row.typeid === 1 ? 'money-in' : 'money-out'">{{ formatMoney(row.money) }}</td>
              <td>{{ row.funds }} {{ row.type }}</td>
              <td class="time">{{ formatTime(row.time) }}</td>
              <td>{{ row.mark }}</td>
            </tr>
          </tbody>
        </table>
        <p v-if="listData.length === 0" class="muted">暂无记录</p>
      </div>
    </main>
    <NavBars current="add" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useUserStore } from '../stores/user'
import NavBars from '../components/NavBars.vue'

const API = '/api'

function base64Json (obj: Record<string, unknown>): string {
  const s = JSON.stringify(obj)
  const bytes = new TextEncoder().encode(s)
  let binary = ''
  for (let i = 0; i < bytes.length; i++) binary += String.fromCharCode(bytes[i])
  return btoa(binary)
}

const userStore = useUserStore()
const activeTab = ref<'in' | 'out'>('out')

interface Fund { fundsid: number; fundsname: string }
const funds = ref<Fund[]>([])
const classMap = ref<{ in: Record<string, string>; out: Record<string, string> }>({ in: {}, out: {} })

const form = ref({
  money: 0 as number,
  fid: -1 as number,
  acclassid: 0 as number,
  acremark: '',
  actime: '',
})
const listData = ref<Array<{ id: number; money: number; classid: number; class: string; typeid: number; type: string; funds: string; time: number; mark: string }>>([])
const pendingFiles = ref<File[]>([])
const fileInput = ref<HTMLInputElement | null>(null)
const moneyDecimals = ref(2)
const moneyPoint = ref('.')
const moneyThousands = ref(',')

const classOptions = computed(() => activeTab.value === 'out' ? classMap.value.out : classMap.value.in)

function formatMoney (v: number | undefined): string {
  if (v == null || Number.isNaN(v)) return '0'
  const dec = moneyDecimals.value
  const pt = moneyPoint.value
  const th = moneyThousands.value
  const parts = Number(v).toFixed(dec).split('.')
  const intPart = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, th)
  return dec > 0 ? intPart + pt + parts[1] : intPart
}

function formatTime (ts: number): string {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0')
}

function loadFunds () {
  fetch(`${API}/funds?type=get`, { credentials: 'include' })
    .then(res => res.json())
    .then((data: { uid: number; data: unknown }) => {
      if (data.uid <= 0) return
      const arr = Array.isArray(data.data) ? data.data as Fund[] : []
      funds.value = arr
      if (arr.length > 0 && (form.value.fid === -1 || form.value.fid === 0)) {
        form.value.fid = arr[0].fundsid
      }
    })
    .catch(() => {})
}

function loadAclass () {
  fetch(`${API}/aclass?type=get`, { credentials: 'include' })
    .then(res => res.json())
    .then((data: { uid: number; data: { in?: Record<string, string>; out?: Record<string, string> } }) => {
      if (data.uid <= 0) return
      const d = data.data || {}
      classMap.value = { in: d.in || {}, out: d.out || {} }
      const opts = activeTab.value === 'out' ? classMap.value.out : classMap.value.in
      const ids = Object.keys(opts)
      if (ids.length > 0 && !ids.includes(String(form.value.acclassid))) {
        form.value.acclassid = Number(ids[0])
      }
    })
    .catch(() => {})
}

function loadFind () {
  const uid = userStore.uid
  if (uid <= 0) return
  const data = base64Json({ jiid: uid, page: 1 })
  fetch(`${API}/find?type=all&data=${encodeURIComponent(data)}`, { credentials: 'include' })
    .then(res => res.json())
    .then((data: { uid: number; data: { ret?: boolean; msg?: { data?: unknown[] } } }) => {
      if (data.uid <= 0) return
      const d = data.data
      if (d && d.ret && d.msg && Array.isArray(d.msg.data)) {
        listData.value = d.msg.data as typeof listData.value
      }
    })
    .catch(() => {})
}

watch(activeTab, () => {
  const opts = classOptions.value
  const ids = Object.keys(opts)
  if (ids.length > 0) form.value.acclassid = Number(ids[0])
})

function triggerFileInput () {
  fileInput.value?.click()
}

function onFileSelect (e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files) {
    pendingFiles.value = Array.from(input.files)
  }
  input.value = ''
}

function uploadPendingFiles (acid: number): Promise<void> {
  if (pendingFiles.value.length === 0) return Promise.resolve()
  const fd = new FormData()
  fd.append('acid', String(acid))
  pendingFiles.value.forEach((file) => fd.append('file[]', file))
  return fetch(`${API}/account/upload`, {
    method: 'POST',
    credentials: 'include',
    body: fd,
  })
    .then(res => res.json())
    .then((data: { uid: number; upload?: unknown[]; data?: string }) => {
      pendingFiles.value = []
      if (data.uid <= 0) return
      if (!Array.isArray(data.upload) && data.data) {
        console.warn('上传图片:', data.data)
      }
    })
    .catch(() => {
      pendingFiles.value = []
    })
}

function onSubmit () {
  const uid = userStore.uid
  if (uid <= 0) return
  const zhifu = activeTab.value === 'out' ? 2 : 1
  const fid = form.value.fid
  const payload = {
    acmoney: Number(form.value.money),
    acclassid: form.value.acclassid,
    actime: form.value.actime || new Date().toISOString().slice(0, 10),
    acremark: form.value.acremark || '',
    zhifu,
    fid: fid > 0 ? fid : -1,
  }
  const body = new URLSearchParams()
  body.set('type', 'add')
  body.set('data', base64Json(payload))
  fetch(`${API}/account`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: body.toString(),
  })
    .then(res => res.json())
    .then((data: { uid: number; data: { ret?: boolean; msg?: string; acid?: number } }) => {
      if (data.uid <= 0) {
        alert(data.data && typeof data.data === 'string' ? data.data : '请先登录')
        return
      }
      const d = data.data as { ret?: boolean; msg?: string; acid?: number }
      if (d && d.ret) {
        const acid = d.acid
        if (acid != null && pendingFiles.value.length > 0) {
          uploadPendingFiles(acid).then(() => {
            alert(d.msg || '添加成功')
            form.value.money = 0
            form.value.acremark = ''
            form.value.actime = new Date().toISOString().slice(0, 10)
            loadFind()
          })
        } else {
          alert(d.msg || '添加成功')
          form.value.money = 0
          form.value.acremark = ''
          form.value.actime = new Date().toISOString().slice(0, 10)
          pendingFiles.value = []
          loadFind()
        }
      } else {
        alert(d && d.msg ? d.msg : '添加失败')
      }
    })
    .catch(() => alert('请求失败'))
}

onMounted(() => {
  form.value.actime = new Date().toISOString().slice(0, 10)
  loadFunds()
  loadAclass()
  loadFind()
  fetch(`${API}/version`, { credentials: 'include' })
    .then(res => res.json())
    .then((data: { account?: Record<string, unknown> }) => {
      const acc = data.account as Record<string, unknown> | undefined
      if (acc) {
        if (typeof acc.MONEY_FORMAT_DECIMALS === 'number') moneyDecimals.value = acc.MONEY_FORMAT_DECIMALS
        if (typeof acc.MONEY_FORMAT_POINT === 'string') moneyPoint.value = acc.MONEY_FORMAT_POINT
        if (typeof acc.MONEY_FORMAT_THOUSANDS === 'string') moneyThousands.value = acc.MONEY_FORMAT_THOUSANDS
      }
    })
    .catch(() => {})
})
</script>

<style scoped>
.add-page { min-height: 100vh; display: flex; flex-direction: column; }
.header { padding: 1rem 1.5rem; background: #fff; box-shadow: 0 1px 3px rgba(0,0,0,0.08); }
.main { flex: 1; padding: 1rem 1.5rem; padding-bottom: 4rem; }

.tabs { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
.tab { padding: 0.5rem 1rem; border: 1px solid #ddd; background: #fff; border-radius: 6px; text-decoration: none; color: #333; cursor: pointer; }
.tab.active { background: #19a7f0; color: #fff; border-color: #19a7f0; }

.form fieldset { border: none; padding: 0; }
.form legend { font-weight: 600; margin-bottom: 0.75rem; }
.field { margin-bottom: 0.75rem; }
.field label { display: block; margin-bottom: 0.25rem; font-size: 0.9rem; }
.field input, .field select { width: 100%; padding: 0.5rem; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
.remark-row { display: flex; gap: 0.5rem; align-items: center; }
.remark-input { flex: 1; }
.upload-wrap { flex-shrink: 0; }
.hidden-input { position: absolute; width: 0; height: 0; opacity: 0; }
.file-hint { font-size: 0.85rem; color: #666; margin-top: 0.25rem; }
.btn { display: inline-block; padding: 0.5rem 1rem; border-radius: 6px; text-align: center; text-decoration: none; cursor: pointer; border: 1px solid #ddd; background: #fff; color: #333; }
.btn-upload { white-space: nowrap; }
.btn-primary { background: #19a7f0; color: #fff; border-color: #19a7f0; width: 100%; margin-bottom: 0.5rem; }
.btn-default { width: 100%; }

.list-section { margin-top: 1.5rem; }
.list-table { width: 100%; border-collapse: collapse; font-size: 0.85rem; }
.list-table th, .list-table td { padding: 0.4rem; border: 1px solid #eee; }
.list-table th { background: #fafafa; }
.list-table .money-in { color: #0a0; }
.list-table .money-out { color: #c00; }
.list-table .time { white-space: nowrap; }
.muted { color: #999; font-size: 0.9rem; }
</style>
