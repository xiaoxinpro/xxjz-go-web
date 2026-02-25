import { reactive } from 'vue'

export type AlertType = 'info' | 'success' | 'warning' | 'error'

const state = reactive({
  visible: false,
  message: '',
  type: 'info' as AlertType,
})

export function useAlert () {
  function show (message: string, type: AlertType = 'info') {
    state.message = message
    state.type = type
    state.visible = true
  }

  function close () {
    state.visible = false
  }

  return { show, close, state }
}
