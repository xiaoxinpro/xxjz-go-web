<template>
  <div class="funds-page">
    <AppHeader />
    <main class="main page-main container">
      <div class="section card">
        <h2 class="card-title">新建资金账户</h2>
        <form class="form" @submit.prevent="onAdd">
          <div class="field">
            <label>账户名称</label>
            <input v-model.trim="addForm.fundsname" type="text" placeholder="账户名称" required />
          </div>
          <!-- 初始金额：后端已支持，暂时隐藏避免误解，后续可恢复展示 -->
          <div class="field" style="display: none;">
            <label>初始金额</label>
            <input v-model.number="addForm.fundsmoney" type="number" step="0.01" placeholder="0" />
          </div>
          <button type="submit" class="btn btn-primary">添加</button>
        </form>
      </div>

      <div class="section card">
        <h2 class="card-title">管理资金账户</h2>
        <div class="table-wrap">
        <table class="list-table">
          <thead>
            <tr>
              <th>账户名称</th>
              <th>收入金额</th>
              <th>支出金额</th>
              <th>剩余金额</th>
              <th>记录数</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="f in fundsList" :key="f.fundsid">
              <td>{{ f.fundsname }}</td>
              <td class="money-in">{{ formatMoney(moneyOf(f).in) }}</td>
              <td class="money-out">{{ formatMoney(moneyOf(f).out) }}</td>
              <td class="money-balance">{{ formatMoney(moneyOf(f).over) }}</td>
              <td>{{ moneyOf(f).count }}</td>
              <td>
                <button type="button" class="btn-link" @click="startEdit(f)">编辑</button>
                <button type="button" class="btn-link" @click="startDelete(f)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
        </div>
        <p v-if="fundsList.length === 0" class="muted">暂无资金账户</p>
      </div>

      <!-- 编辑弹窗 -->
      <div v-if="editing" class="modal-mask" @click.self="editing = null">
        <div class="modal">
          <h3>编辑账户</h3>
          <form @submit.prevent="onEdit">
            <div class="field">
              <label>账户名称</label>
              <input v-model.trim="editForm.fundsname" type="text" required />
            </div>
            <button type="submit" class="btn btn-primary">保存</button>
            <button type="button" class="btn" @click="editing = null">取消</button>
          </form>
        </div>
      </div>

      <!-- 删除：选择合并目标 -->
      <div v-if="deleting" class="modal-mask" @click.self="deleting = null">
        <div class="modal">
          <h3>删除「{{ deleting.fundsname }}」</h3>
          <p>请选择将该账户下的记录合并到哪个账户：</p>
          <form @submit.prevent="onDelete">
            <div class="field">
              <label>合并到</label>
              <select v-model.number="deleteTargetId" required>
                <option v-for="f in fundsList.filter(x => x.fundsid !== deleting.fundsid)" :key="f.fundsid" :value="f.fundsid">{{ f.fundsname }}</option>
              </select>
            </div>
            <button type="submit" class="btn btn-primary">确认删除</button>
            <button type="button" class="btn" @click="deleting = null">取消</button>
          </form>
        </div>
      </div>

      <p class="back-link"><router-link to="/home" class="btn btn-default">返回</router-link></p>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AppHeader from '../components/AppHeader.vue'

const API = '/api'

function base64Json(obj: Record<string, unknown>): string {
  const s = JSON.stringify(obj)
  const bytes = new TextEncoder().encode(s)
  let binary = ''
  for (let i = 0; i < bytes.length; i++) binary += String.fromCharCode(bytes[i])
  return btoa(binary)
}

interface MoneyStats { in: number; out: number; over: number; count: number }
interface FundRow {
  fundsid: number
  fundsname: string
  uid?: number
  sort?: number
  money?: MoneyStats
}
const fundsList = ref<FundRow[]>([])
const moneyDecimals = ref(2)

function moneyOf(f: FundRow): MoneyStats {
  const m = f.money
  if (m && typeof m.in === 'number' && typeof m.out === 'number' && typeof m.over === 'number' && typeof m.count === 'number') {
    return m
  }
  return { in: 0, out: 0, over: 0, count: 0 }
}

function formatMoney(v: number | undefined): string {
  if (v == null || Number.isNaN(v)) return '0'
  return Number(v).toFixed(moneyDecimals.value)
}
const addForm = ref({ fundsname: '', fundsmoney: 0 })
const editing = ref<FundRow | null>(null)
const editForm = ref({ fundsname: '' })
const deleting = ref<FundRow | null>(null)
const deleteTargetId = ref<number>(0)

async function loadFunds() {
  const res = await fetch(`${API}/funds?type=get`, { credentials: 'include' })
  const data = await res.json()
  if (data.uid && Array.isArray(data.data)) fundsList.value = data.data
  else fundsList.value = []
}

async function onAdd() {
  const body = new URLSearchParams()
  body.set('type', 'add')
  body.set('data', base64Json({ fundsname: addForm.value.fundsname, fundsmoney: addForm.value.fundsmoney || 0 }))
  const res = await fetch(API + '/funds', { method: 'POST', body, credentials: 'include' })
  const data = await res.json()
  const out = data.data
  if (Array.isArray(out) && out[0] === true) {
    addForm.value = { fundsname: '', fundsmoney: 0 }
    loadFunds()
  } else {
    alert(Array.isArray(out) ? out[1] : (data.data || '添加失败'))
  }
}

function startEdit(f: FundRow) {
  editing.value = f
  editForm.value = { fundsname: f.fundsname }
}

async function onEdit() {
  if (!editing.value) return
  const body = new URLSearchParams()
  body.set('type', 'edit')
  body.set('data', base64Json({ fundsid: editing.value.fundsid, fundsname: editForm.value.fundsname }))
  const res = await fetch(API + '/funds', { method: 'POST', body, credentials: 'include' })
  const data = await res.json()
  const out = data.data
  if (Array.isArray(out) && out[0] === true) {
    editing.value = null
    loadFunds()
  } else {
    alert(Array.isArray(out) ? out[1] : (data.data || '保存失败'))
  }
}

function startDelete(f: FundRow) {
  deleting.value = f
  const others = fundsList.value.filter((x) => x.fundsid !== f.fundsid)
  deleteTargetId.value = others[0]?.fundsid ?? 0
}

async function onDelete() {
  if (!deleting.value) return
  const body = new URLSearchParams()
  body.set('type', 'del')
  body.set('data', base64Json({ fundsid_old: deleting.value.fundsid, fundsid_new: deleteTargetId.value }))
  const res = await fetch(API + '/funds', { method: 'POST', body, credentials: 'include' })
  const data = await res.json()
  const out = data.data
  if (Array.isArray(out) && out[0] === true) {
    deleting.value = null
    loadFunds()
  } else {
    alert(Array.isArray(out) ? out[1] : (data.data || '删除失败'))
  }
}

onMounted(() => {
  loadFunds()
  fetch(`${API}/version`, { credentials: 'include' })
    .then(res => res.json())
    .then((data: { account?: Record<string, unknown> }) => {
      const acc = data.account as Record<string, unknown> | undefined
      if (acc && typeof acc.MONEY_FORMAT_DECIMALS === 'number') {
        moneyDecimals.value = acc.MONEY_FORMAT_DECIMALS
      }
    })
    .catch(() => {})
})
</script>

<style scoped>
.funds-page { min-height: 100vh; padding-bottom: var(--space-xl); }
.section { margin-bottom: var(--space-xl); }
.section .card-title { margin-bottom: var(--space-md); }
.form .field label { display: block; }
.table-wrap { margin-bottom: var(--space-md); }
.list-table td:nth-child(2), .list-table td:nth-child(3), .list-table td:nth-child(4),
.list-table td:nth-child(5) { text-align: right; }
.btn-link { margin-right: var(--space-sm); }
.modal .btn { margin-right: var(--space-sm); margin-top: var(--space-sm); }
.back-link { margin-top: var(--space-lg); }
</style>
