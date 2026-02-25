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
    { path: '/add/transfer', name: 'AddTransfer', component: () => import('../views/AddTransferView.vue'), meta: { requiresAuth: true } },
    { path: '/add', name: 'Add', component: () => import('../views/AddView.vue'), meta: { requiresAuth: true } },
    { path: '/find', name: 'Find', component: () => import('../views/FindView.vue'), meta: { requiresAuth: true } },
    { path: '/edit/transfer/:id', name: 'EditTransfer', component: () => import('../views/EditPlaceholderView.vue'), meta: { requiresAuth: true } },
    { path: '/edit/:id', name: 'EditAccount', component: () => import('../views/EditAccountView.vue'), meta: { requiresAuth: true } },
    { path: '/chart', name: 'Chart', component: () => import('../views/ChartView.vue'), meta: { requiresAuth: true } },
    { path: '/funds', name: 'Funds', component: () => import('../views/FundsView.vue'), meta: { requiresAuth: true } },
    { path: '/class', name: 'Class', component: () => import('../views/ClassView.vue'), meta: { requiresAuth: true } },
    { path: '/user', name: 'User', component: () => import('../views/UserView.vue'), meta: { requiresAuth: true } },
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
    if (userStore.uid <= 0) {
      const ok = await userStore.restoreFromSession()
      if (!ok) return { path: '/login', query: { redirect: to.fullPath } }
    }
  }
})

export default router
