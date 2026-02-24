<template>
  <div class="home-page">
    <AppHeader />
    <main class="main">
      <div class="welcome alert" v-if="welcomeText">
        <strong>{{ userStore.username }}，{{ welcomeText }}</strong>
      </div>

      <!-- 本日/本月/本年 -->
      <div class="stat-section">
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
      <!-- 昨日/上月/去年 -->
      <div class="stat-section">
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

      <div id="money-table" class="summary alert">
        <p class="summary-line">
          总收入: <span class="money-in">{{ formatMoney(stat.SumInMoney) }}</span>
          总支出: <span class="money-out">{{ formatMoney(stat.SumOutMoney) }}</span>
          剩余: <span class="money-balance">{{ formatMoney(sumBalance) }}</span>
          <span class="badge">不含转账</span>
        </p>
      </div>

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
            <td>
              <template v-if="row.classid > 0">
                <router-link :to="'/edit/' + row.id">编辑</router-link>
                <a href="javascript:void(0)" role="button" @click.prevent="confirmDel(row.id, true)"> 删除</a>
              </template>
              <template v-else>
                <router-link :to="'/edit/transfer/' + row.id">编辑</router-link>
                <a href="javascript:void(0)" role="button" @click.prevent="confirmDel(row.id, false)"> 删除</a>
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
      <p v-else-if="pagemax === 1 && listData.length > 0" class="more-link">
        <button type="button" class="link-btn" @click="scrollToMoneyTable">更多记账明细</button>
      </p>
    </main>
    <NavBars current="home" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useUserStore } from '../stores/user'
import { useRouter } from 'vue-router'
import AppHeader from '../components/AppHeader.vue'
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
    loadFind(page)
  } else {
    alert(msg)
  }
}

function scrollToMoneyTable () {
  const el = document.getElementById('money-table')
  if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function logout () {
  userStore.logout()
  router.push('/login')
}

onMounted(() => {
  loadVersion()
  loadStatistic()
  loadFind(1)
})
</script>

<style scoped>
.home-page { min-height: 100vh; display: flex; flex-direction: column; }
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  background: #fff;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
}
.user { font-size: 0.9rem; color: #666; }
.main { flex: 1; padding: 1rem 1.5rem; padding-bottom: 4rem; }

.alert { padding: 0.75rem 1rem; margin-bottom: 1rem; border-radius: 6px; background: #f5f5f5; }
.welcome.alert { background: #e8f4fc; }
.summary.alert { background: #f0f0f0; }
.summary-line { margin: 0; font-size: 0.9rem; }
.summary-line .money-in { color: #0a0; }
.summary-line .money-out { color: #c00; }
.summary-line .money-balance { color: #07c; }
.badge { margin-left: 6px; padding: 2px 6px; border-radius: 10px; background: #ddd; font-size: 0.75rem; }

.stat-section { margin-bottom: 1rem; }
.stat-table { width: 100%; border-collapse: collapse; border-radius: 6px; overflow: hidden; box-shadow: 0 1px 2px rgba(0,0,0,0.06); }
.stat-table th, .stat-table td { padding: 0.5rem 0.4rem; text-align: center; border: 1px solid #eee; }
.stat-table th { background: #fff; font-weight: 600; }
.stat-table .row-out { background: #ffebee; }
.stat-table .row-in { background: #e8f5e9; }
.stat-table .row-balance { background: #e3f2fd; }

.list-table { width: 100%; border-collapse: collapse; font-size: 0.85rem; margin-bottom: 1rem; }
.list-table th, .list-table td { padding: 0.5rem 0.4rem; border: 1px solid #eee; }
.list-table th { background: #fafafa; }
.list-table .money-in { color: #0a0; }
.list-table .money-out { color: #c00; }
.list-table .time { white-space: nowrap; }
.list-table a { margin-right: 0.25rem; }

.pagination { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 1rem; }
.pagination button:disabled { opacity: 0.5; cursor: not-allowed; }
.more-link { margin-bottom: 1rem; }
.more-link a { color: #07c; }
.more-link .link-btn {
  background: none; border: none; padding: 0; font-size: inherit; color: #07c; cursor: pointer; text-decoration: none;
}
.more-link .link-btn:hover { text-decoration: underline; }

.btn-outline {
  padding: 0.5rem 1rem;
  border: 1px solid #ddd;
  background: #fff;
  border-radius: 6px;
  cursor: pointer;
}
</style>
