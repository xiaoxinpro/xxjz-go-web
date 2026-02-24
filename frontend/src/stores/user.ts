import { defineStore } from 'pinia'
import { ref } from 'vue'

const API = '/api'

export const useUserStore = defineStore('user', () => {
  const uid = ref(0)
  const username = ref('')

  async function login(user: string, pass: string): Promise<{ ok: boolean; message: string }> {
    const params = new URLSearchParams({ username: user, password: pass })
    const res = await fetch(`${API}/login`, {
      method: 'POST',
      body: params,
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      credentials: 'include',
    })
    const data = await res.json()
    const id = Number(data.uid)
    if (id > 0) {
      uid.value = id
      username.value = data.uname || user
      return { ok: true, message: '' }
    }
    return { ok: false, message: data.uname || '登录失败' }
  }

  function logout() {
    uid.value = 0
    username.value = ''
  }

  /** Restore login state from server session (e.g. after refresh). Returns true if session is valid. */
  async function restoreFromSession(): Promise<boolean> {
    const res = await fetch(`${API}/user?type=get`, { credentials: 'include' })
    const data = await res.json().catch(() => ({}))
    const id = Number(data.uid)
    if (id > 0) {
      uid.value = id
      username.value = data.username || ''
      return true
    }
    return false
  }

  return { uid, username, login, logout, restoreFromSession }
})
