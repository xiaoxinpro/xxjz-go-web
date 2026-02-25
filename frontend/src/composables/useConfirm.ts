import { reactive } from 'vue'
import type { AlertType } from './useAlert'

const state = reactive({
  visible: false,
  message: '',
  type: 'warning' as AlertType,
})

let resolveRef: ((value: boolean) => void) | null = null

export function useConfirm () {
  function showConfirm (message: string, type: AlertType = 'warning'): Promise<boolean> {
    state.message = message
    state.type = type
    state.visible = true
    return new Promise<boolean>((resolve) => {
      resolveRef = resolve
    })
  }

  function close (confirmed: boolean) {
    state.visible = false
    if (resolveRef) {
      resolveRef(confirmed)
      resolveRef = null
    }
  }

  return { showConfirm, close, state }
}
