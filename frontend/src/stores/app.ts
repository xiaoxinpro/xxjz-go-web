import { defineStore } from 'pinia'
import { ref } from 'vue'

const API = '/api'

export const useAppStore = defineStore('app', () => {
  const title = ref('小歆记账')

  async function loadVersion () {
    try {
      const res = await fetch(`${API}/version`, { credentials: 'include' })
      const data = await res.json()
      if (data.title) title.value = data.title
    } catch {
      /* keep default */
    }
  }

  return { title, loadVersion }
})
