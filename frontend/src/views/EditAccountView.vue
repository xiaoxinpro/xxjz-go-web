<template>
  <div class="edit-page">
    <AppHeader />
    <main class="main page-main">
      <p v-if="loadError" class="error">{{ loadError }}</p>
      <template v-else-if="loaded">
        <div class="form-card card">
          <h2 class="card-title">编辑记账</h2>
        <form class="form" enctype="multipart/form-data" @submit.prevent="onSubmit">
            <div class="field">
              <label>金额</label>
              <input v-model.number="form.money" type="number" step="0.01" min="0" required />
            </div>
            <div v-if="funds.length > 1" class="field">
              <label>账户</label>
              <select v-model.number="form.fid" required>
                <option v-for="f in funds" :key="f.fundsid" :value="f.fundsid">{{ f.fundsname }}</option>
              </select>
            </div>
            <div v-else class="field" style="display:none">
              <input v-model.number="form.fid" type="hidden" />
            </div>
            <div class="field">
              <label>类别</label>
              <select v-model.number="form.typeid" @change="onTypeChange">
                <option :value="1">收入</option>
                <option :value="2">支出</option>
              </select>
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
                <input v-model.trim="form.acremark" type="text" class="remark-input" />
                <span class="upload-wrap">
                  <button type="button" class="btn btn-default btn-upload" :disabled="uploading" @click="triggerFileInput">
                    <ImagePlus v-if="!uploading" size="18" /> {{ uploading ? '上传中...' : '上传图片' }}
                  </button>
                  <input ref="fileInput" type="file" accept=".jpg,.jpeg,.png,.gif" multiple class="hidden-input" @change="onFileChange" />
                </span>
              </div>
            </div>
            <div v-if="imageList.length > 0" class="image-section">
              <p class="divider" />
              <ul class="image-gallery">
                <li v-for="img in imageList" :key="img.id" class="image-item">
                  <button type="button" class="image-delete" aria-label="删除图片" @click="deleteImage(img.id)"><Trash2 size="14" /></button>
                  <a :href="img.url" target="_blank" rel="noopener" class="image-link">
                    <div class="zoom-image" :style="{ backgroundImage: 'url(' + img.url + ')' }" />
                    <span class="image-name">{{ img.name }}</span>
                    <span class="image-time">{{ formatImageTime(img.time) }}</span>
                  </a>
                </li>
              </ul>
            </div>
            <div class="field">
              <label>时间</label>
              <input v-model="form.actime" type="date" required />
            </div>
            <button type="submit" class="btn btn-primary">修改</button>
            <button type="button" class="btn btn-danger" @click="onDelete"><Trash2 size="18" /> 删除</button>
            <router-link to="/home" class="btn btn-default">返回</router-link>
        </form>
        </div>
      </template>
      <p v-else class="muted">加载中...</p>
    </main>
    <NavBars current="home" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import AppHeader from '../components/AppHeader.vue'
import NavBars from '../components/NavBars.vue'
import { ImagePlus, Trash2 } from 'lucide-vue-next'

const API = '/api'
const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

function base64Json (obj: Record<string, unknown>): string {
  const s = JSON.stringify(obj)
  const bytes = new TextEncoder().encode(s)
  let binary = ''
  for (let i = 0; i < bytes.length; i++) binary += String.fromCharCode(bytes[i])
  return btoa(binary)
}

interface Fund { fundsid: number; fundsname: string }
interface ImageItem { id: number; name: string; url: string; time: number }

const acid = computed(() => Number(route.params.id) || 0)
const funds = ref<Fund[]>([])
const classMap = ref<{ in: Record<string, string>; out: Record<string, string> }>({ in: {}, out: {} })
const form = ref({
  money: 0 as number,
  fid: -1 as number,
  typeid: 2 as number,
  acclassid: 0 as number,
  acremark: '',
  actime: '',
})
const imageList = ref<ImageItem[]>([])
const loaded = ref(false)
const loadError = ref('')
const uploading = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

const classOptions = computed(() => form.value.typeid === 1 ? classMap.value.in : classMap.value.out)

function formatImageTime (ts: number): string {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0')
}

function loadFunds () {
  return fetch(`${API}/funds?type=get`, { credentials: 'include' })
    .then(res => res.json())
    .then((data: { uid: number; data: unknown }) => {
      if (data.uid <= 0) return
      funds.value = Array.isArray(data.data) ? data.data as Fund[] : []
    })
}

function loadAclass () {
  return fetch(`${API}/aclass?type=get`, { credentials: 'include' })
    .then(res => res.json())
    .then((data: { uid: number; data: { in?: Record<string, string>; out?: Record<string, string> } }) => {
      if (data.uid <= 0) return
      const d = data.data || {}
      classMap.value = { in: d.in || {}, out: d.out || {} }
    })
}

function loadAccount () {
  const uid = userStore.uid
  if (uid <= 0 || acid.value <= 0) return Promise.resolve()
  const data = base64Json({ acid: acid.value, jiid: uid })
  return fetch(`${API}/account?type=get_id&data=${encodeURIComponent(data)}`, { credentials: 'include' })
    .then(res => res.json())
    .then((data: { uid: number; data: Record<string, unknown> | string }) => {
      if (data.uid <= 0) {
        loadError.value = typeof data.data === 'string' ? data.data : '请先登录'
        return
      }
      const d = data.data
      if (!d || typeof d !== 'object' || Array.isArray(d)) {
        loadError.value = '记录不存在或无权访问'
        return
      }
      const r = d as { id?: number; money?: number; fid?: number; typeid?: number; classid?: number; time?: string; mark?: string }
      form.value.money = Number(r.money) || 0
      form.value.fid = Number(r.fid) ?? -1
      form.value.typeid = Number(r.typeid) || 2
      form.value.acclassid = Number(r.classid) || 0
      form.value.actime = (r.time as string) || ''
      form.value.acremark = (r.mark as string) || ''
    })
}

function loadImages () {
  if (acid.value <= 0) return Promise.resolve()
  const data = base64Json({ acid: acid.value })
  return fetch(`${API}/account?type=get_image&data=${encodeURIComponent(data)}`, { credentials: 'include' })
    .then(res => res.json())
    .then((data: { uid: number; data: { ret?: boolean; msg?: ImageItem[] } }) => {
      if (data.uid <= 0) return
      const d = data.data
      if (d && d.ret && Array.isArray(d.msg)) {
        imageList.value = d.msg
      } else {
        imageList.value = []
      }
    })
}

function onTypeChange () {
  const opts = classOptions.value
  const ids = Object.keys(opts)
  if (ids.length > 0 && !ids.includes(String(form.value.acclassid))) {
    form.value.acclassid = Number(ids[0])
  }
}

watch(acid, () => {
  if (acid.value > 0) {
    loaded.value = false
    loadError.value = ''
    loadAccount().then(() => loadImages()).then(() => { loaded.value = true })
  }
})

function triggerFileInput () {
  fileInput.value?.click()
}

function onFileChange (e: Event) {
  const input = e.target as HTMLInputElement
  const files = input.files
  if (!files || files.length === 0) return
  uploading.value = true
  const fd = new FormData()
  fd.append('acid', String(acid.value))
  for (let i = 0; i < files.length; i++) {
    fd.append('file[]', files[i])
  }
  fetch(`${API}/account/upload`, {
    method: 'POST',
    credentials: 'include',
    body: fd,
  })
    .then(res => res.json())
    .then((data: { uid: number; upload?: ImageItem[]; data?: string }) => {
      uploading.value = false
      input.value = ''
      if (data.uid <= 0) {
        alert(data.data || '请先登录')
        return
      }
      if (Array.isArray(data.upload) && data.upload.length > 0) {
        imageList.value = [...imageList.value, ...data.upload]
      }
      if (data.data && !data.upload) {
        alert(data.data)
      }
    })
    .catch(() => {
      uploading.value = false
      input.value = ''
      alert('上传失败')
    })
}

function deleteImage (id: number) {
  if (!confirm('确定要删除这张图片吗？')) return
  const data = base64Json({ acid: acid.value, id })
  fetch(`${API}/account`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({ type: 'del_image', data }).toString(),
  })
    .then(res => res.json())
    .then((data: { uid: number; data: { ret?: boolean } }) => {
      if (data.uid <= 0) return
      if (data.data?.ret) {
        imageList.value = imageList.value.filter((img) => img.id !== id)
      }
    })
}

function onSubmit () {
  const uid = userStore.uid
  if (uid <= 0) return
  const payload = {
    acid: acid.value,
    acmoney: form.value.money,
    acclassid: form.value.acclassid,
    actime: form.value.actime || new Date().toISOString().slice(0, 10),
    acremark: form.value.acremark || '',
    zhifu: form.value.typeid,
    fid: form.value.fid > 0 ? form.value.fid : -1,
  }
  const body = new URLSearchParams()
  body.set('type', 'edit')
  body.set('data', base64Json(payload))
  fetch(`${API}/account`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: body.toString(),
  })
    .then(res => res.json())
    .then((data: { uid: number; data: { ret?: boolean; msg?: string } }) => {
      if (data.uid <= 0) {
        alert(typeof data.data === 'string' ? data.data : '请先登录')
        return
      }
      const d = data.data
      if (d?.ret) {
        alert(d.msg || '修改成功')
        router.push('/home')
      } else {
        alert(d?.msg || '修改失败')
      }
    })
    .catch(() => alert('请求失败'))
}

function onDelete () {
  if (!confirm('确定要删除这条记账吗？')) return
  const body = new URLSearchParams()
  body.set('type', 'del')
  body.set('data', base64Json({ acid: acid.value }))
  fetch(`${API}/account`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: body.toString(),
  })
    .then(res => res.json())
    .then((data: { uid: number; data: { ret?: boolean; msg?: string } }) => {
      if (data.uid <= 0) return
      if (data.data?.ret) {
        alert(data.data.msg || '已删除')
        router.push('/home')
      } else {
        alert(data.data?.msg || '删除失败')
      }
    })
    .catch(() => alert('请求失败'))
}

onMounted(() => {
  if (acid.value <= 0) {
    loadError.value = '无效的记账ID'
    return
  }
  loaded.value = false
  loadError.value = ''
  Promise.all([loadFunds(), loadAclass()]).then(() => loadAccount()).then(() => loadImages()).then(() => { loaded.value = true })
})
</script>

<style scoped>
.edit-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
.main {
  flex: 1;
  padding-bottom: 4.5rem;
}
.form-card .form button,
.form-card .form .btn {
  width: 100%;
  margin-bottom: var(--space-sm);
}
.form-card .form .btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-xs);
}
.remark-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-sm);
  align-items: center;
}
.remark-input { flex: 1; min-width: 120px; }
.upload-wrap { flex-shrink: 0; }
.btn-upload { white-space: nowrap; }
.hidden-input { position: absolute; width: 0; height: 0; opacity: 0; pointer-events: none; }

.divider { border: none; border-top: 1px dashed var(--color-border); margin: var(--space-lg) 0; }
.image-section { margin: var(--space-md) 0; }
.image-gallery {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: var(--space-md);
}
.image-item { position: relative; }
.image-delete {
  position: absolute;
  top: 4px;
  right: 4px;
  z-index: 1;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.6);
  color: #fff;
  cursor: pointer;
  padding: 0;
}
.image-link { display: block; text-decoration: none; color: var(--color-text); }
.zoom-image {
  width: 100%;
  padding-bottom: 100%;
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  border-radius: var(--radius-md);
}
.image-name, .image-time { display: block; font-size: 0.75rem; color: var(--color-text-muted); margin-top: var(--space-xs); }
.error { color: var(--color-danger); padding: var(--space-md); }
.muted { color: var(--color-text-muted); padding: var(--space-md); }
</style>
