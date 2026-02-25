<template>
  <div class="home-page">
    <AppHeader />
    <main class="main page-main">
      <div class="welcome card" v-if="welcomeText">
        <strong>{{ userStore.username }}，{{ welcomeText }}</strong>
      </div>

      <!-- 本日/本月/本年 -->
      <div class="stat-grid">
        <div class="stat-section card">
          <table class="stat-table">
            <thead>
              <tr><th>统计</th><th>本日</th><th>本月</th><th>本年</th></tr>
            </thead>
            <tbody>
              <tr class="row-out"><td>支出</td><td>{{ formatMoney(stat.TodayOutMoney) }}</td><td>{{ formatMoney(stat.MonthOutMoney) }}</td><td>{{ formatMoney(stat.YearOutMoney) }}</td></tr>
              <tr class="row-in"><td>收入</td><td>{{ formatMoney(stat.TodayInMoney) }}</td><td>{{ formatMoney(stat.MonthInMoney) }}</td><td>{{ formatMoney(stat.YearInMoney) }}</td></tr>
              <tr class="row-balance"><td>剩余</td><td>{{ formatMoney(todayBalance) }}</td><td>{{ formatMoney(monthBalance) }}</td><td>{{ formatMoney(yearBalance) }}</td></tr>
            </tbody>
          </table>
        </div>
        <div class="stat-section card">
          <table class="stat-table">
            <thead>
              <tr><th>统计</th><th>昨日</th><th>上月</th><th>去年</th></tr>
            </thead>
            <tbody>
              <tr class="row-out"><td>支出</td><td>{{ formatMoney(stat.LastTodayOutMoney) }}</td><td>{{ formatMoney(stat.LastMonthOutMoney) }}</td><td>{{ formatMoney(stat.LastYearOutMoney) }}</td></tr>
              <tr class="row-in"><td>收入</td><td>{{ formatMoney(stat.LastTodayInMoney) }}</td><td>{{ formatMoney(stat.LastMonthInMoney) }}</td><td>{{ formatMoney(stat.LastYearInMoney) }}</td></tr>
              <tr class="row-balance"><td>剩余</td><td>{{ formatMoney(lastTodayBalance) }}</td><td>{{ formatMoney(lastMonthBalance) }}</td><td>{{ formatMoney(lastYearBalance) }}</td></tr>
            </tbody>
          </table>
        </div>
      </div>

      <div id="money-table" class="summary card">
        <p class="summary-line">
          总收入: <span class="money-in">{{ formatMoney(stat.SumInMoney) }}</span>
          总支出: <span class="money-out">{{ formatMoney(stat.SumOutMoney) }}</span>
          剩余: <span class="money-balance">{{ formatMoney(sumBalance) }}</span>
          <span class="badge">不含转账</span>
        </p>
      </div>

      <div class="table-wrap">
        <table class="list-table">
          <thead>
            <tr>
              <th>分类</th><th>金额</th><th>收支</th><th>时间</th><th>备注</th><th>操作</th>
            </tr>
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
                  <button type="button" class="btn-link" title="删除" @click.prevent="confirmDel(row.id, true)"><Trash2 size="16" /></button>
                </template>
                <template v-else>
                  <router-link :to="'/edit/transfer/' + row.id" class="btn-link" title="编辑"><Pencil size="16" /></router-link>
                  <button type="button" class="btn-link" title="删除" @click.prevent="confirmDel(row.id, false)"><Trash2 size="16" /></button>
                </template>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="pagination-wrap" v-if="pagemax > 1">
        <div class="pagination">
          <button type="button" class="pagination-btn" :disabled="page <= 1" @click="goPage(page - 1)">上一页</button>
          <span class="pagination-info">第 {{ page }} / {{ pagemax }} 页</span>
          <span class="page-select">
            <select :value="page" @change="goPage(Number(($event.target as HTMLSelectElement).value))">
              <option v-for="p in pageRange" :key="p" :value="p">{{ p }}</option>
            </select>
          </span>
          <button type="button" class="pagination-btn" :disabled="page >= pagemax" @click="goPage(page + 1)">下一页</button>
        </div>
      </div>
      <p v-else-if="pagemax === 1 && listData.length > 0" class="more-link">
        <button type="button" class="btn btn-outline" @click="scrollToMoneyTable">更多记账明细</button>
      </p>
    </main>
    <NavBars current="home" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useUserStore } from '../stores/user'
import { useRouter } from 'vue-router'
import { useAlert } from '../composables/useAlert'
import { useConfirm } from '../composables/useConfirm'
import AppHeader from '../components/AppHeader.vue'
import NavBars from '../components/NavBars.vue'
import { Pencil, Trash2 } from 'lucide-vue-next'

const { show: showAlert } = useAlert()
const { showConfirm } = useConfirm()

const API = '/api'

function base64Json (obj: Record<string, unknown>): string {
  const s = JSON.stringify(obj)
  const bytes = new TextEncoder().encode(s)
  let binary = ''
  for (let i = 0; i < bytes.length; i++) binary += String.fromCharCode(bytes[i])
  return btoa(binary)
}

const userStore = useUserStore()
const router = useRouter()

const welcomeText = ref('欢迎使用！')
const moneyDecimals = ref(2)
const moneyPoint = ref('.')
const moneyThousands = ref(',')

const stat = ref<Record<string, number>>({
  TodayInMoney: 0, TodayOutMoney: 0, MonthInMoney: 0, MonthOutMoney: 0, YearInMoney: 0, YearOutMoney: 0,
  LastTodayInMoney: 0, LastTodayOutMoney: 0, LastMonthInMoney: 0, LastMonthOutMoney: 0, LastYearInMoney: 0, LastYearOutMoney: 0,
  SumInMoney: 0, SumOutMoney: 0,
})
const listData = ref<Array<{ id: number; money: number; classid: number; class: string; typeid: number; type: string; funds: string; time: number; mark: string }>>([])
const page = ref(1)
const pagemax = ref(1)

const todayBalance = computed(() => (stat.value.TodayInMoney || 0) - (stat.value.TodayOutMoney || 0))
const monthBalance = computed(() => (stat.value.MonthInMoney || 0) - (stat.value.MonthOutMoney || 0))
const yearBalance = computed(() => (stat.value.YearInMoney || 0) - (stat.value.YearOutMoney || 0))
const lastTodayBalance = computed(() => (stat.value.LastTodayInMoney || 0) - (stat.value.LastTodayOutMoney || 0))
const lastMonthBalance = computed(() => (stat.value.LastMonthInMoney || 0) - (stat.value.LastMonthOutMoney || 0))
const lastYearBalance = computed(() => (stat.value.LastYearInMoney || 0) - (stat.value.LastYearOutMoney || 0))
const sumBalance = computed(() => (stat.value.SumInMoney || 0) - (stat.value.SumOutMoney || 0))
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

function loadVersion () {
  fetch(`${API}/version`, { credentials: 'include' })
    .then(res => res.json())
    .then((data: { message?: string; account?: Record<string, unknown> }) => {
      if (data.message) welcomeText.value = data.message
      const acc = data.account as Record<string, unknown> | undefined
      if (acc) {
        if (typeof acc.MONEY_FORMAT_DECIMALS === 'number') moneyDecimals.value = acc.MONEY_FORMAT_DECIMALS
        if (typeof acc.MONEY_FORMAT_POINT === 'string') moneyPoint.value = acc.MONEY_FORMAT_POINT
        if (typeof acc.MONEY_FORMAT_THOUSANDS === 'string') moneyThousands.value = acc.MONEY_FORMAT_THOUSANDS
      }
    })
    .catch(() => {})
}

function loadStatistic () {
  fetch(`${API}/statistic`, { credentials: 'include' })
    .then(res => res.json())
    .then((data: { uid: number; data: Record<string, number> | string }) => {
      if (data.uid > 0 && typeof data.data === 'object' && data.data !== null) {
        stat.value = { ...stat.value, ...data.data }
      }
    })
    .catch(() => {})
}

function loadFind (p: number) {
  const uid = userStore.uid
  if (uid <= 0) return
  const data = base64Json({ jiid: uid, page: p })
  fetch(`${API}/find?type=all&data=${encodeURIComponent(data)}`, { credentials: 'include' })
    .then(res => res.json())
    .then((data: { uid: number; data: { ret?: boolean; msg?: { data?: unknown[]; page?: number; pagemax?: number; count?: number } } }) => {
      if (data.uid <= 0) return
      const d = data.data
      if (d && d.ret && d.msg) {
        listData.value = (d.msg.data as Array<{ id: number; money: number; classid: number; class: string; typeid: number; type: string; funds: string; time: number; mark: string }>) || []
        page.value = d.msg.page ?? 1
        pagemax.value = d.msg.pagemax ?? 1
      }
    })
    .catch(() => {})
}

function goPage (p: number) {
  if (p < 1 || p > pagemax.value) return
  page.value = p
  loadFind(p)
}

async function confirmDel (id: number, isAccount: boolean) {
  const ok = await showConfirm('确定删除这条记录？', 'warning')
  if (!ok) return
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
    loadFind(page.value)
  } else {
    showAlert(msg, 'error')
  }
}

function scrollToMoneyTable () {
  const el = document.getElementById('money-table')
  if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

onMounted(() => {
  loadVersion()
  loadStatistic()
  loadFind(1)
})
</script>

<style scoped>
.home-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
.main {
  flex: 1;
  padding-bottom: 4.5rem;
}

.welcome.card {
  background: var(--color-balance-bg);
}

.stat-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--space-lg);
  margin-bottom: var(--space-lg);
}
@media (min-width: 640px) {
  .stat-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (min-width: 1024px) {
  .stat-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.stat-section.card {
  padding: var(--space-md);
  margin-bottom: 0;
}
.stat-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.85rem;
}
.stat-table th,
.stat-table td {
  padding: var(--space-sm) var(--space-xs);
  text-align: center;
  border: 1px solid var(--color-border);
}
.stat-table th {
  background: var(--color-bg-card);
  font-weight: 600;
}
.stat-table .row-out {
  background: var(--color-expense-bg);
}
.stat-table .row-in {
  background: var(--color-income-bg);
}
.stat-table .row-balance {
  background: var(--color-balance-bg);
}

.summary.card {
  margin-bottom: var(--space-lg);
}
.summary-line {
  margin: 0;
  font-size: 0.9rem;
}

.table-wrap {
  margin-bottom: var(--space-lg);
}
.list-table {
  font-size: 0.85rem;
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

.pagination-wrap {
  display: flex;
  justify-content: center;
  margin: var(--space-xl) 0 var(--space-lg);
}
.pagination-wrap .pagination {
  display: inline-flex;
  align-items: center;
  gap: var(--space-md);
  flex-wrap: wrap;
  justify-content: center;
  margin: 0;
  padding: var(--space-sm) var(--space-lg);
  background: var(--color-bg-card);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--color-border);
}
.pagination-wrap .pagination-btn {
  min-height: var(--touch-min);
  padding: var(--space-sm) var(--space-md);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border);
  background: var(--color-bg-card);
  color: var(--color-text);
  cursor: pointer;
  font-size: 0.9rem;
}
.pagination-wrap .pagination-btn:hover:not(:disabled) {
  background: var(--color-bg);
  border-color: var(--color-primary);
  color: var(--color-primary);
}
.pagination-wrap .pagination-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.pagination-wrap .pagination-info {
  font-size: 0.9rem;
  color: var(--color-text-muted);
}
.pagination-wrap .page-select select {
  min-height: 36px;
  padding: var(--space-xs) var(--space-sm);
  border-radius: var(--radius-sm);
  border: 1px solid var(--color-border);
  background: var(--color-bg-card);
  font-size: 0.9rem;
}

.more-link {
  margin-bottom: var(--space-lg);
}
</style>
