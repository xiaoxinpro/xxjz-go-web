import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../stores/user'

const API = '/api'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/home' },
    { path: '/init', name: 'Init', component: () => import('../views/InitView.vue'), meta: { init: true } },
    { path: '/login', name: 'Login', component: () => import('../views/LoginView.vue'), meta: { guest: true } },
    { path: '/home', name: 'Home', component: () => import('../views/HomeView.vue'), meta: { requiresAuth: true } },
  ],
})

router.beforeEach(async (to) => {
  const res = await fetch(`${API}/init/status`, { credentials: 'include' })
  let initialized = true
  try {
    const data = await res.json()
    initialized = !!data.initialized
  } catch {
    /* assume initialized on error */
  }
  if (!initialized && to.path !== '/init') return { path: '/init' }
  if (initialized && to.path === '/init') return { path: '/login' }

  if (to.meta.requiresAuth) {
    const userStore = useUserStore()
    if (userStore.uid <= 0) return { path: '/login', query: { redirect: to.fullPath } }
  }
})

export default router
