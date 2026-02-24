<template>
  <header class="app-header">
    <div class="header-left">
      <router-link to="/user" class="header-link" title="设置">
        <span class="icon-user">用户</span>
      </router-link>
    </div>
    <h1 class="header-title">
      <router-link to="/home" class="header-link">{{ title }}</router-link>
    </h1>
    <div class="header-right" ref="dropdownRef">
      <button type="button" class="dropdown-trigger" aria-haspopup="true" :aria-expanded="menuOpen" @click="toggleMenu">
        <span class="icon-menu">≡</span>
      </button>
      <div class="dropdown-panel" v-show="menuOpen">
        <ul class="dropdown-list">
          <li><router-link to="/chart" class="dropdown-item" @click="closeMenu">年度统计</router-link></li>
          <li class="divider"></li>
          <li><router-link to="/funds" class="dropdown-item" @click="closeMenu">资金账户</router-link></li>
          <li class="divider"></li>
          <li><router-link to="/class" class="dropdown-item" @click="closeMenu">分类管理</router-link></li>
          <li class="divider"></li>
          <li><router-link to="/user" class="dropdown-item" @click="closeMenu">设置选项</router-link></li>
          <li class="divider"></li>
          <li><a href="javascript:void(0)" class="dropdown-item" @click="onLogout">退出登录</a></li>
        </ul>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()

const title = ref('小歆记账')
const menuOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

async function loadVersion() {
  try {
    const res = await fetch('/api/version', { credentials: 'include' })
    const data = await res.json()
    if (data.title) title.value = data.title
  } catch {
    /* keep default title */
  }
}

function toggleMenu() {
  menuOpen.value = !menuOpen.value
}

function closeMenu() {
  menuOpen.value = false
}

function onLogout() {
  closeMenu()
  userStore.logout()
  router.push('/login')
}

function onClickOutside(e: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target as Node)) {
    menuOpen.value = false
  }
}

onMounted(() => {
  loadVersion()
  document.addEventListener('click', onClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', onClickOutside)
})
</script>

<style scoped>
.app-header {
  position: sticky;
  top: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 48px;
  padding: 0 12px;
  background-color: #19a7f0;
  color: #fff;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.12);
}

.header-left,
.header-right {
  flex-shrink: 0;
  min-width: 48px;
}

.header-right {
  position: relative;
  display: flex;
  justify-content: flex-end;
}

.header-title {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
  flex: 1;
  text-align: center;
}

.header-link {
  color: inherit;
  text-decoration: none;
}

.header-link:hover {
  opacity: 0.9;
}

.icon-user {
  font-size: 0.95rem;
}

.dropdown-trigger {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  padding: 0;
  border: none;
  background: transparent;
  color: inherit;
  font-size: 1.4rem;
  cursor: pointer;
  border-radius: 4px;
}

.dropdown-trigger:hover {
  background: rgba(255, 255, 255, 0.2);
}

.dropdown-panel {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 4px;
  min-width: 160px;
  background: #fff;
  color: #333;
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  overflow: hidden;
}

.dropdown-list {
  list-style: none;
  margin: 0;
  padding: 6px 0;
}

.dropdown-list li {
  margin: 0;
}

.dropdown-item {
  display: block;
  padding: 10px 16px;
  color: #333;
  text-decoration: none;
  font-size: 0.95rem;
}

.dropdown-item:hover {
  background: #f0f0f0;
}

.dropdown-list .divider {
  height: 1px;
  margin: 4px 0;
  padding: 0;
  background: #e0e0e0;
}
</style>
