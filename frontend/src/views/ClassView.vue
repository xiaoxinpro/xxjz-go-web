<template>
  <div class="class-page">
    <AppHeader />
    <main class="main page-main container">
      <div class="section card">
        <h2 class="card-title">新建分类</h2>
        <form class="form" @submit.prevent="onAdd">
          <div class="field">
            <label>名称</label>
            <input v-model.trim="addForm.classname" type="text" placeholder="分类名称" required />
          </div>
          <div class="field">
            <label>类别</label>
            <select v-model.number="addForm.classtype">
              <option :value="2">支出</option>
              <option :value="1">收入</option>
            </select>
          </div>
          <button type="submit" class="btn btn-primary">添加</button>
        </form>
      </div>

      <div class="section card">
        <h2 class="card-title">收入分类</h2>
        <div class="table-wrap">
        <table class="list-table">
          <thead>
            <tr><th>分类名称</th><th>操作</th></tr>
          </thead>
          <tbody>
            <tr v-for="c in classInList" :key="c.classid">
              <td>{{ c.classname }}</td>
              <td>
                <button type="button" class="btn-link" @click="turnType(c, 2)">转为支出</button>
                <button type="button" class="btn-link" @click="startEdit(c)">编辑</button>
                <button type="button" class="btn-link" @click="confirmDel(c)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
        </div>
        <p v-if="classInList.length === 0" class="muted">暂无收入分类</p>
      </div>

      <div class="section card">
        <h2 class="card-title">支出分类</h2>
        <div class="table-wrap">
        <table class="list-table">
          <thead>
            <tr><th>分类名称</th><th>操作</th></tr>
          </thead>
          <tbody>
            <tr v-for="c in classOutList" :key="c.classid">
              <td>{{ c.classname }}</td>
              <td>
                <button type="button" class="btn-link" @click="turnType(c, 1)">转为收入</button>
                <button type="button" class="btn-link" @click="startEdit(c)">编辑</button>
                <button type="button" class="btn-link" @click="confirmDel(c)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
        </div>
        <p v-if="classOutList.length === 0" class="muted">暂无支出分类</p>
      </div>

      <!-- 编辑弹窗 -->
      <div v-if="editing" class="modal-mask" @click.self="editing = null">
        <div class="modal">
          <h3>编辑分类</h3>
          <form @submit.prevent="onEdit">
            <div class="field">
              <label>名称</label>
              <input v-model.trim="editForm.classname" type="text" required />
            </div>
            <div class="field">
              <label>类别</label>
              <select v-model.number="editForm.classtype">
                <option :value="1">收入</option>
                <option :value="2">支出</option>
              </select>
            </div>
            <button type="submit" class="btn btn-primary">保存</button>
            <button type="button" class="btn" @click="editing = null">取消</button>
          </form>
        </div>
      </div>

      <p class="back-link"><router-link to="/home" class="btn btn-default">返回</router-link></p>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAlert } from '../composables/useAlert'
import { useConfirm } from '../composables/useConfirm'
import AppHeader from '../components/AppHeader.vue'

const { show: showAlert } = useAlert()
const { showConfirm } = useConfirm()

const API = '/api'

function base64Json(obj: Record<string, unknown>): string {
  const s = JSON.stringify(obj)
  const bytes = new TextEncoder().encode(s)
  let binary = ''
  for (let i = 0; i < bytes.length; i++) binary += String.fromCharCode(bytes[i])
  return btoa(binary)
}

interface ClassRow { classid: number; classname: string; classtype: number; ufid?: number; sort?: number }
const classAllList = ref<ClassRow[]>([])
const addForm = ref({ classname: '', classtype: 2 })
const editing = ref<ClassRow | null>(null)
const editForm = ref({ classname: '', classtype: 1 })

const classInList = computed(() => classAllList.value.filter((c) => c.classtype === 1))
const classOutList = computed(() => classAllList.value.filter((c) => c.classtype === 2))

async function loadClass() {
  const [resIn, resOut] = await Promise.all([
    fetch(`${API}/aclass?type=getindata`, { credentials: 'include' }),
    fetch(`${API}/aclass?type=getoutdata`, { credentials: 'include' }),
  ])
  const dataIn = await resIn.json()
  const dataOut = await resOut.json()
  const inArr = Array.isArray(dataIn.data) ? dataIn.data : []
  const outArr = Array.isArray(dataOut.data) ? dataOut.data : []
  classAllList.value = [...inArr, ...outArr]
}

async function onAdd() {
  const body = new URLSearchParams()
  body.set('type', 'add')
  body.set('data', base64Json({ classname: addForm.value.classname, classtype: addForm.value.classtype }))
  const res = await fetch(API + '/aclass', { method: 'POST', body, credentials: 'include' })
  const data = await res.json()
  const out = data.data
  if (Array.isArray(out) && out[0] === true) {
    addForm.value = { classname: '', classtype: 2 }
    loadClass()
  } else {
    showAlert(Array.isArray(out) ? out[1] : (data.data || '添加失败'), 'error')
  }
}

function startEdit(c: ClassRow) {
  editing.value = c
  editForm.value = { classname: c.classname, classtype: c.classtype }
}

async function onEdit() {
  if (!editing.value) return
  const body = new URLSearchParams()
  body.set('type', 'edit')
  body.set('data', base64Json({ classid: editing.value.classid, classname: editForm.value.classname, classtype: editForm.value.classtype }))
  const res = await fetch(API + '/aclass', { method: 'POST', body, credentials: 'include' })
  const data = await res.json()
  const out = data.data
  if (Array.isArray(out) && out[0] === true) {
    editing.value = null
    loadClass()
  } else {
    showAlert(Array.isArray(out) ? out[1] : (data.data || '保存失败'), 'error')
  }
}

async function turnType(c: ClassRow, classtype: number) {
  const body = new URLSearchParams()
  body.set('type', 'edit')
  body.set('data', base64Json({ classid: c.classid, classname: c.classname, classtype }))
  const res = await fetch(API + '/aclass', { method: 'POST', body, credentials: 'include' })
  const data = await res.json()
  const out = data.data
  if (Array.isArray(out) && out[0] === true) loadClass()
  else showAlert(Array.isArray(out) ? out[1] : (data.data || '操作失败'), 'error')
}

async function confirmDel(c: ClassRow) {
  const ok = await showConfirm(`确定删除分类「${c.classname}」？`, 'warning')
  if (!ok) return
  doDel(c.classid)
}

async function doDel(classid: number) {
  const body = new URLSearchParams()
  body.set('type', 'del')
  body.set('data', base64Json({ classid }))
  const res = await fetch(API + '/aclass', { method: 'POST', body, credentials: 'include' })
  const data = await res.json()
  const out = data.data
  if (Array.isArray(out) && out[0] === true) loadClass()
  else showAlert(Array.isArray(out) ? out[1] : (data.data || '删除失败'), 'error')
}

onMounted(() => loadClass())
</script>

<style scoped>
.class-page { min-height: 100vh; padding-bottom: var(--space-xl); }
.section { margin-bottom: var(--space-xl); }
.section .card-title { margin-bottom: var(--space-md); }
.form .field label { display: block; }
.table-wrap { margin-bottom: var(--space-md); }
.btn-link { margin-right: var(--space-sm); }
.modal .btn { margin-right: var(--space-sm); margin-top: var(--space-sm); }
.back-link { margin-top: var(--space-lg); }
</style>
