<template>
  <div class="find-page">
    <header class="header">
      <span>小歆记账</span>
    </header>
    <main class="main">
      <form class="form" @submit.prevent="onQuery">
        <fieldset>
          <legend>查询账目</legend>
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
          <p><button type="submit" class="btn btn-primary">查询</button></p>
          <p><button type="button" class="btn btn-danger" @click="onReset">重置</button></p>
          <p><router-link to="/home" class="btn btn-default">返回</router-link></p>
        </fieldset>
      </form>

      <div v-if="searched" class="result-section">
        <div id="money-table" class="summary alert">
          <p class="summary-line">
            收入: <span class="money-in">{{ formatMoney(sumIn) }}</span>
            支出: <span class="money-out">{{ formatMoney(sumOut) }}</span>
            剩余: <span class="money-balance">{{ formatMoney(sumIn - sumOut) }}</span>
            <span class="badge">{{ isTransfer ? '含转账' : '不含转账' }}</span>
          </p>
        </div>
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
              <td>
                <template v-if="row.classid > 0">
                  <router-link :to="'/edit/' + row.id">编辑</router-link>
                  <a href="javascript:void(0)" @click="() => {}"> 删除</a>
                </template>
                <template v-else>
                  <router-link :to="'/edit/transfer/' + row.id">编辑</router-link>
                  <a href="javascript:void(0)" @click="() => {}"> 删除</a>
                </template>
              </td>
            </tr>
          </tbody>
        </table>
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
.find-page { min-height: 100vh; display: flex; flex-direction: column; }
.header { padding: 1rem 1.5rem; background: #fff; box-shadow: 0 1px 3px rgba(0,0,0,0.08); }
.main { flex: 1; padding: 1rem 1.5rem; padding-bottom: 4rem; }

.form fieldset { border: none; padding: 0; }
.form legend { font-weight: 600; margin-bottom: 0.75rem; }
.field { margin-bottom: 0.75rem; }
.field label { display: block; margin-bottom: 0.25rem; font-size: 0.9rem; }
.field input, .field select { width: 100%; padding: 0.5rem; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
.btn { display: inline-block; padding: 0.5rem 1rem; border-radius: 6px; text-align: center; text-decoration: none; cursor: pointer; border: 1px solid #ddd; background: #fff; color: #333; width: 100%; margin-bottom: 0.5rem; }
.btn-primary { background: #19a7f0; color: #fff; border-color: #19a7f0; }
.btn-danger { background: #e74c3c; color: #fff; border-color: #e74c3c; }
.btn-default { }

.result-section { margin-top: 1.5rem; }
.summary.alert { padding: 0.75rem 1rem; margin-bottom: 1rem; border-radius: 6px; background: #f0f0f0; }
.summary-line { margin: 0; font-size: 0.9rem; }
.summary-line .money-in { color: #0a0; }
.summary-line .money-out { color: #c00; }
.summary-line .money-balance { color: #07c; }
.badge { margin-left: 6px; padding: 2px 6px; border-radius: 10px; background: #ddd; font-size: 0.75rem; }

.list-table { width: 100%; border-collapse: collapse; font-size: 0.85rem; margin-bottom: 1rem; }
.list-table th, .list-table td { padding: 0.4rem; border: 1px solid #eee; }
.list-table th { background: #fafafa; }
.list-table .money-in { color: #0a0; }
.list-table .money-out { color: #c00; }
.list-table .time { white-space: nowrap; }
.list-table a { margin-right: 0.25rem; }

.pagination { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 1rem; }
.pagination button:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
