<template>
  <div class="find-page">
    <AppHeader />
    <main class="main page-main">
      <div class="filter-card card">
        <h2 class="card-title"><Search size="20" class="title-icon" /> 查询账目</h2>
        <form class="form" @submit.prevent="onQuery">
          <div class="field">
            <label>资金账户</label>
            <select v-model="query.fid">
              <option value="">全部账户</option>
              <option v-for="f in funds" :key="f.fundsid" :value="String(f.fundsid)">{{ f.fundsname }}</option>
            </select>
          </div>
          <div class="field">
            <label>收支类别</label>
            <select v-model="query.zhifu" @change="onTypeChange">
              <option value="">全部类别</option>
              <option value="1">收入</option>
              <option value="2">支出</option>
              <option value="3">转账</option>
            </select>
          </div>
          <div class="field">
            <label>选择分类</label>
            <select v-model="query.acclassid">
              <option value="">全部</option>
              <option v-for="(name, id) in classOptionsForType" :key="String(id)" :value="String(id)">{{ name }}</option>
            </select>
          </div>
          <div class="field">
            <label>起始时间</label>
            <input v-model="query.starttime" type="date" />
          </div>
          <div class="field">
            <label>终止时间</label>
            <input v-model="query.endtime" type="date" />
          </div>
          <div class="field">
            <label>备注信息</label>
            <input v-model.trim="query.acremark" type="text" placeholder="备注" />
          </div>
          <div class="form-actions">
            <button type="submit" class="btn btn-primary">查询</button>
            <button type="button" class="btn btn-danger" @click="onReset">重置</button>
            <router-link to="/home" class="btn btn-default">返回</router-link>
          </div>
        </form>
      </div>

      <div v-if="searched" class="result-section">
        <div id="money-table" class="summary card">
          <p class="summary-line">
            收入: <span class="money-in">{{ formatMoney(sumIn) }}</span>
            支出: <span class="money-out">{{ formatMoney(sumOut) }}</span>
            剩余: <span class="money-balance">{{ formatMoney(sumIn - sumOut) }}</span>
            <span class="badge">{{ isTransfer ? '含转账' : '不含转账' }}</span>
          </p>
        </div>
        <div class="table-wrap">
          <table class="list-table">
            <thead>
              <tr><th>分类</th><th>金额</th><th>收支</th><th>时间</th><th>备注</th><th>操作</th></tr>
            </thead>
            <tbody>
              <tr v-for="row in listData" :key="row.classid === 0 ? 't' + row.id : 'a' + row.id">
                <td>{{ row.class }}</td>
                <td :class="row.typeid === 1 ? 'money-in' : 'money-out'">{{ formatMoney(row.money) }}</td>
                <td>{{ row.funds }} {{ row.type }}</td>
                <td class="time">{{ formatTime(row.time) }}</td>
                <td>{{ row.mark }}</td>
                <td class="actions">
                  <template v-if="row.classid > 0">
                    <router-link :to="'/edit/' + row.id" class="btn-link" title="编辑"><Pencil size="16" /></router-link>
                    <button type="button" class="btn-link" title="删除" @click="confirmDel(row.id, true)"><Trash2 size="16" /></button>
                  </template>
                  <template v-else>
                    <router-link :to="'/edit/transfer/' + row.id" class="btn-link" title="编辑"><Pencil size="16" /></router-link>
                    <button type="button" class="btn-link" title="删除" @click="confirmDel(row.id, false)"><Trash2 size="16" /></button>
                  </template>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="pagination" v-if="pagemax > 1">
          <button type="button" :disabled="page <= 1" @click="goPage(page - 1)">上一页</button>
          <span class="page-select">
            <select :value="page" @change="goPage(Number(($event.target as HTMLSelectElement).value))">
              <option v-for="p in pageRange" :key="p" :value="p">{{ p }}</option>
            </select>
          </span>
          <button type="button" :disabled="page >= pagemax" @click="goPage(page + 1)">下一页</button>
        </div>
      </div>
    </main>
    <NavBars current="find" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useUserStore } from '../stores/user'
import AppHeader from '../components/AppHeader.vue'
import NavBars from '../components/NavBars.vue'
import { Search, Pencil, Trash2 } from 'lucide-vue-next'

const API = '/api'

function base64Json (obj: Record<string, unknown>): string {
  const s = JSON.stringify(obj)
  const bytes = new TextEncoder().encode(s)
  let binary = ''
  for (let i = 0; i < bytes.length; i++) binary += String.fromCharCode(bytes[i])
  return btoa(binary)
}

const userStore = useUserStore()

interface Fund { fundsid: number; fundsname: string }
const funds = ref<Fund[]>([])
const classMap = ref<{ in: Record<string, string>; out: Record<string, string> }>({ in: {}, out: {} })

const query = ref({
  fid: '',
  zhifu: '',
  acclassid: '',
  starttime: '',
  endtime: '',
  acremark: '',
})

const searched = ref(false)
const listData = ref<Array<{ id: number; money: number; classid: number; class: string; typeid: number; type: string; funds: string; time: number; mark: string }>>([])
const page = ref(1)
const pagemax = ref(1)
const sumIn = ref(0)
const sumOut = ref(0)
const isTransfer = ref(false)
const moneyDecimals = ref(2)
const moneyPoint = ref('.')
const moneyThousands = ref(',')

const classOptionsForType = computed(() => {
  const z = query.value.zhifu
  if (z === '1') return classMap.value.in
  if (z === '2') return classMap.value.out
  if (z === '3') return { inTransfer: '转入', outTransfer: '转出' } as Record<string, string>
  return {}
})

const pageRange = computed(() => Array.from({ length: pagemax.value }, (_, i) => i + 1))

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

function onTypeChange () {
  query.value.acclassid = ''
}

function loadFunds () {
  fetch(`${API}/funds?type=get`, { credentials: 'include' })
    .then(res => res.json())
    .then((data: { uid: number; data: unknown }) => {
      if (data.uid <= 0) return
      funds.value = Array.isArray(data.data) ? data.data as Fund[] : []
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
    })
    .catch(() => {})
}

function buildFindData (p: number): Record<string, unknown> {
  const q = query.value
  const uid = userStore.uid
  const data: Record<string, unknown> = { jiid: uid, page: p }
  if (q.fid) data.fid = Number(q.fid)
  if (q.zhifu) data.zhifu = Number(q.zhifu)
  if (q.acclassid) {
    if (q.acclassid === 'inTransfer' || q.acclassid === 'outTransfer') {
      data.acclassid = q.acclassid
    } else {
      data.acclassid = Number(q.acclassid)
    }
  }
  if (q.starttime) data.starttime = q.starttime
  if (q.endtime) data.endtime = q.endtime
  if (q.acremark) data.acremark = q.acremark
  return data
}

function doFind (p: number) {
  const uid = userStore.uid
  if (uid <= 0) return
  const data = base64Json(buildFindData(p))
  fetch(`${API}/find?type=all&data=${encodeURIComponent(data)}`, { credentials: 'include' })
    .then(res => res.json())
    .then((data: { uid: number; data: { ret?: boolean; msg?: { data?: unknown[]; page?: number; pagemax?: number; count?: number; SumInMoney?: number; SumOutMoney?: number; isTransfer?: boolean } } }) => {
      if (data.uid <= 0) return
      const d = data.data
      if (d && d.ret && d.msg) {
        const msg = d.msg
        listData.value = (msg.data as typeof listData.value) || []
        page.value = msg.page ?? 1
        pagemax.value = msg.pagemax ?? 1
        sumIn.value = msg.SumInMoney ?? 0
        sumOut.value = msg.SumOutMoney ?? 0
        isTransfer.value = !!msg.isTransfer
      }
    })
    .catch(() => {})
}

function onQuery () {
  searched.value = true
  page.value = 1
  doFind(1)
}

function onReset () {
  query.value = { fid: '', zhifu: '', acclassid: '', starttime: '', endtime: '', acremark: '' }
  searched.value = false
  listData.value = []
  page.value = 1
  pagemax.value = 1
  sumIn.value = 0
  sumOut.value = 0
}

function goPage (p: number) {
  if (p < 1 || p > pagemax.value) return
  page.value = p
  doFind(p)
}

async function confirmDel (id: number, isAccount: boolean) {
  if (!confirm('确定删除这条记录？')) return
  const body = new URLSearchParams()
  body.set('type', 'del')
  body.set('data', isAccount ? base64Json({ acid: id }) : base64Json({ tid: id }))
  const url = isAccount ? `${API}/account` : `${API}/transfer`
  const res = await fetch(url, { method: 'POST', body, credentials: 'include' })
  const data = await res.json()
  const out = data.data
  const ret = out && typeof out.ret === 'boolean' ? out.ret : false
  const msg = out && typeof out.msg === 'string' ? out.msg : (data.data || '删除失败')
  if (ret) {
    doFind(page)
  } else {
    alert(msg)
  }
}

onMounted(() => {
  loadFunds()
  loadAclass()
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
.find-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
.main {
  flex: 1;
  padding-bottom: 4.5rem;
}
.filter-card .card-title {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}
.title-icon {
  color: var(--color-primary);
}
.form-actions {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}
.form-actions .btn {
  width: 100%;
}

.result-section {
  margin-top: var(--space-lg);
}
.summary.card {
  margin-bottom: var(--space-lg);
}
.table-wrap {
  margin-bottom: var(--space-lg);
}
.actions {
  white-space: nowrap;
}
.actions .btn-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 32px;
  min-height: 32px;
  padding: var(--space-xs);
  margin-right: var(--space-xs);
  color: var(--color-primary);
}
.actions .btn-link:hover {
  color: var(--color-primary-hover);
}
.actions button.btn-link {
  color: var(--color-danger);
}
.actions button.btn-link:hover {
  color: var(--color-expense);
}
</style>
