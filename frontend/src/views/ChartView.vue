<template>
  <div class="chart-page">
    <AppHeader />
    <main class="main page-main">
      <div class="year-nav card">
        <router-link :to="`/chart`" class="btn btn-outline" @click.prevent="setYear(year - 1)">« {{ year - 1 }}年</router-link>
        <span class="year-current">{{ year }}年</span>
        <router-link :to="`/chart`" class="btn btn-outline" @click.prevent="setYear(year + 1)">{{ year + 1 }}年 »</router-link>
      </div>
      <p v-if="error" class="error">{{ error }}</p>
      <p v-else-if="!chartData && !loading" class="empty">暂无数据</p>
      <template v-else-if="chartData && !isErrorData">
        <div ref="chartYearRef" class="chart-box"></div>
        <div v-if="hasPieData" class="charts-row">
          <div ref="chartInRef" class="chart-pie"></div>
          <div ref="chartOutRef" class="chart-pie"></div>
        </div>
      </template>
      <p class="back-link"><router-link to="/home" class="btn btn-default">返回</router-link></p>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import AppHeader from '../components/AppHeader.vue'
import * as echarts from 'echarts'

const API = '/api'

const year = ref(new Date().getFullYear())
const chartData = ref<Record<string, unknown> | null>(null)
const loading = ref(false)
const error = ref('')
const chartYearRef = ref<HTMLElement | null>(null)
const chartInRef = ref<HTMLElement | null>(null)
const chartOutRef = ref<HTMLElement | null>(null)
let chartYear: echarts.ECharts | null = null
let chartIn: echarts.ECharts | null = null
let chartOut: echarts.ECharts | null = null

const isErrorData = computed(() => {
  const d = chartData.value
  return d && typeof (d as Record<string, unknown>).uid === 'number' && typeof (d as Record<string, unknown>).data === 'string'
})

const hasPieData = computed(() => {
  const d = chartData.value as Record<string, unknown> | undefined
  if (!d || isErrorData.value) return false
  const inSum = d.InSumClassMoney as Record<string, number> | undefined
  const outSum = d.OutSumClassMoney as Record<string, number> | undefined
  const inTotal = inSum && Object.values(inSum).some((v) => v > 0)
  const outTotal = outSum && Object.values(outSum).some((v) => v > 0)
  return !!inTotal || !!outTotal
})

function dateTs(y: number) {
  return Math.floor(new Date(y, 0, 1).getTime() / 1000)
}

async function load() {
  loading.value = true
  error.value = ''
  chartData.value = null
  try {
    const res = await fetch(`${API}/chart?type=year&date=${dateTs(year.value)}`, { credentials: 'include' })
    const data = await res.json()
    if (data.uid === 0 && data.data) {
      error.value = typeof data.data === 'string' ? data.data : '未登录'
      return
    }
    if (data.uid && typeof data.data === 'string') {
      error.value = data.data
      return
    }
    chartData.value = data
  } catch (e) {
    error.value = (e as Error).message || '加载失败'
  } finally {
    loading.value = false
  }
}

function setYear(y: number) {
  if (y >= 2000 && y <= 2100) year.value = y
}

function buildBarOption() {
  const d = chartData.value as Record<string, unknown>
  const inMoney = (d.InMoney as Record<string, number>) || {}
  const outMoney = (d.OutMoney as Record<string, number>) || {}
  const surplus = (d.SurplusMoney as Record<string, number>) || {}
  const surplusSum = (d.SurplusSumMoney as Record<string, number>) || {}
  const months = ['1月', '2月', '3月', '4月', '5月', '6月', '7月', '8月', '9月', '10月', '11月', '12月']
  const arr = (m: Record<string, number>) => months.map((_, i) => m[String(i + 1)] ?? 0)
  return {
    title: { text: `${year.value}年收入与支出金额汇总`, top: 'top', left: 'center' },
    tooltip: { trigger: 'axis' },
    legend: { data: ['收入', '支出', '月剩余', '年剩余'], bottom: 10, left: 'center' },
    grid: { left: '3%', right: '4%', bottom: '15%', containLabel: true },
    xAxis: { type: 'category', data: months },
    yAxis: { type: 'value' },
    series: [
      { name: '收入', type: 'bar', data: arr(inMoney) },
      { name: '支出', type: 'bar', data: arr(outMoney) },
      { name: '月剩余', type: 'line', smooth: true, data: arr(surplus) },
      { name: '年剩余', type: 'line', smooth: true, data: arr(surplusSum) },
    ],
  }
}

function buildPieOption(title: string, subtext: string, nameValue: Record<string, number>) {
  const data = Object.entries(nameValue)
    .filter(([, v]) => v > 0)
    .map(([name, value]) => ({ name, value }))
  return {
    title: { text: title, subtext, left: 'center' },
    tooltip: { trigger: 'item', formatter: '{b}: {c} 元 ({d}%)' },
    series: [{ type: 'pie', radius: '55%', center: ['50%', '55%'], data }],
  }
}

function renderCharts() {
  if (!chartData.value || isErrorData.value) return
  const d = chartData.value as Record<string, unknown>
  if (chartYearRef.value) {
    chartYear = echarts.init(chartYearRef.value)
    chartYear.setOption(buildBarOption())
  }
  const inSum = (d.InSumClassMoney as Record<string, number>) || {}
  const outSum = (d.OutSumClassMoney as Record<string, number>) || {}
  const inTotal = Object.values(inSum).reduce((a, b) => a + b, 0)
  const outTotal = Object.values(outSum).reduce((a, b) => a + b, 0)
  if (chartInRef.value && Object.keys(inSum).length) {
    chartIn = echarts.init(chartInRef.value)
    chartIn.setOption(buildPieOption(`${year.value}年收入分类汇总`, `总收入 ${inTotal.toFixed(2)} 元`, inSum))
  }
  if (chartOutRef.value && Object.keys(outSum).length) {
    chartOut = echarts.init(chartOutRef.value)
    chartOut.setOption(buildPieOption(`${year.value}年支出分类汇总`, `总支出 ${outTotal.toFixed(2)} 元`, outSum))
  }
}

watch(year, () => {
  load()
})

watch([chartData, year], () => {
  if (chartYear) chartYear.dispose()
  if (chartIn) chartIn.dispose()
  if (chartOut) chartOut.dispose()
  chartYear = chartIn = chartOut = null
  nextTick(() => renderCharts())
})

onMounted(() => {
  load()
})

onUnmounted(() => {
  if (chartYear) chartYear.dispose()
  if (chartIn) chartIn.dispose()
  if (chartOut) chartOut.dispose()
})
</script>

<style scoped>
.chart-page {
  min-height: 100vh;
  padding-bottom: var(--space-xl);
}
.main {
  max-width: 900px;
}
.year-nav.card {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-lg);
  margin-bottom: var(--space-lg);
}
.year-current { font-weight: 600; color: var(--color-text); }
.error, .empty {
  color: var(--color-danger);
  text-align: center;
  padding: var(--space-lg);
}
.chart-box { height: 320px; width: 100%; min-height: 240px; }
.charts-row {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--space-lg);
  margin-top: var(--space-lg);
}
@media (min-width: 640px) {
  .charts-row { grid-template-columns: 1fr 1fr; }
}
.chart-pie { height: 280px; min-height: 220px; }
.back-link { margin-top: var(--space-xl); }
</style>
