<template>
  <header class="app-header">
    <div class="header-left">
      <router-link to="/user" class="header-link" title="设置" aria-label="用户设置">
        <User class="header-icon" size="22" />
      </router-link>
    </div>
    <h1 class="header-title">
      <router-link to="/home" class="header-link">{{ title }}</router-link>
    </h1>
    <div class="header-right" ref="dropdownRef">
      <button type="button" class="dropdown-trigger" aria-haspopup="true" :aria-expanded="menuOpen" @click="toggleMenu" aria-label="菜单">
        <Menu class="header-icon" size="22" />
      </button>
      <div class="dropdown-panel" v-show="menuOpen">
        <ul class="dropdown-list">
          <li><router-link to="/chart" class="dropdown-item" @click="closeMenu"><BarChart3 size="18" class="dropdown-icon" /> 年度统计</router-link></li>
          <li class="divider"></li>
          <li><router-link to="/funds" class="dropdown-item" @click="closeMenu"><Wallet size="18" class="dropdown-icon" /> 资金账户</router-link></li>
          <li class="divider"></li>
          <li><router-link to="/class" class="dropdown-item" @click="closeMenu"><Tags size="18" class="dropdown-icon" /> 分类管理</router-link></li>
          <li class="divider"></li>
          <li><router-link to="/user" class="dropdown-item" @click="closeMenu"><Settings size="18" class="dropdown-icon" /> 设置选项</router-link></li>
          <li class="divider"></li>
          <li><a href="javascript:void(0)" class="dropdown-item" @click="onLogout"><LogOut size="18" class="dropdown-icon" /> 退出登录</a></li>
        </ul>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { User, Menu, BarChart3, Wallet, Tags, Settings, LogOut } from 'lucide-vue-next'

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
  min-height: var(--touch-min, 48px);
  padding: 0 12px;
  background-color: var(--color-primary);
  color: #fff;
  box-shadow: var(--shadow-md);
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
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 44px;
  min-height: 44px;
  color: inherit;
  text-decoration: none;
}

.header-link:hover {
  opacity: 0.9;
}

.header-icon {
  flex-shrink: 0;
}

.dropdown-trigger {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 44px;
  min-height: 44px;
  padding: 0;
  border: none;
  background: transparent;
  color: inherit;
  cursor: pointer;
  border-radius: var(--radius-md);
}

.dropdown-trigger:hover {
  background: rgba(255, 255, 255, 0.2);
}

.dropdown-panel {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 4px;
  min-width: 180px;
  background: var(--color-bg-card);
  color: var(--color-text);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  overflow: hidden;
}

.dropdown-list {
  list-style: none;
  margin: 0;
  padding: var(--space-sm) 0;
}

.dropdown-list li {
  margin: 0;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-md) var(--space-lg);
  color: var(--color-text);
  text-decoration: none;
  font-size: 0.95rem;
}

.dropdown-item:hover {
  background: var(--color-bg);
}

.dropdown-icon {
  flex-shrink: 0;
  opacity: 0.8;
}

.dropdown-list .divider {
  height: 1px;
  margin: var(--space-xs) 0;
  padding: 0;
  background: var(--color-border);
}
</style>
