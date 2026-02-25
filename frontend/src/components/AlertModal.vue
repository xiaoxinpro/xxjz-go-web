<template>
  <div v-if="state.visible" class="modal-mask" role="dialog" aria-modal="true" :aria-labelledby="'alert-title-' + state.type">
    <div class="modal alert-modal" :class="'alert-modal--' + state.type">
      <div class="alert-icon-wrap" :class="'alert-icon-wrap--' + state.type" aria-hidden="true">
        <Info v-if="state.type === 'info'" size="28" class="alert-icon" />
        <CheckCircle v-else-if="state.type === 'success'" size="28" class="alert-icon" />
        <AlertTriangle v-else-if="state.type === 'warning'" size="28" class="alert-icon" />
        <AlertCircle v-else size="28" class="alert-icon" />
      </div>
      <p :id="'alert-title-' + state.type" class="alert-message">{{ state.message }}</p>
      <button type="button" class="btn alert-btn" :class="alertBtnClass" @click="close">确定</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAlert } from '../composables/useAlert'
import { Info, CheckCircle, AlertTriangle, AlertCircle } from 'lucide-vue-next'

const { close, state } = useAlert()

const alertBtnClass = computed(() => {
  const t = state.type
  if (t === 'success') return 'btn-primary'
  if (t === 'error') return 'btn-danger'
  return 'btn-default'
})
</script>

<style scoped>
.alert-modal {
  text-align: center;
}
.alert-icon-wrap {
  margin-bottom: var(--space-md);
  display: flex;
  justify-content: center;
  align-items: center;
}
.alert-icon-wrap--info .alert-icon {
  color: var(--color-primary);
}
.alert-icon-wrap--success .alert-icon {
  color: var(--color-success);
}
.alert-icon-wrap--warning .alert-icon {
  color: var(--color-warning);
}
.alert-icon-wrap--error .alert-icon {
  color: var(--color-danger);
}
.alert-message {
  margin: 0 0 var(--space-xl);
  font-size: 1rem;
  line-height: 1.5;
  color: var(--color-text);
  white-space: pre-wrap;
  word-break: break-word;
}
.alert-btn {
  width: 100%;
  min-width: 120px;
}
</style>
