<template>
  <div v-if="state.visible" class="modal-mask" role="dialog" aria-modal="true" :aria-labelledby="'confirm-title-' + state.type">
    <div class="modal confirm-modal" :class="'confirm-modal--' + state.type">
      <div class="confirm-icon-wrap" :class="'confirm-icon-wrap--' + state.type" aria-hidden="true">
        <AlertTriangle v-if="state.type === 'warning'" size="28" class="confirm-icon" />
        <AlertCircle v-else-if="state.type === 'error'" size="28" class="confirm-icon" />
        <Info v-else size="28" class="confirm-icon" />
      </div>
      <p :id="'confirm-title-' + state.type" class="confirm-message">{{ state.message }}</p>
      <div class="confirm-actions">
        <button type="button" class="btn btn-default confirm-btn" @click="close(false)">取消</button>
        <button type="button" class="btn confirm-btn confirm-btn-ok" :class="confirmOkClass" @click="close(true)">确定</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useConfirm } from '../composables/useConfirm'
import { Info, AlertTriangle, AlertCircle } from 'lucide-vue-next'

const { close, state } = useConfirm()

const confirmOkClass = computed(() => {
  const t = state.type
  if (t === 'error') return 'btn-danger'
  return 'btn-primary'
})
</script>

<style scoped>
.confirm-modal {
  text-align: center;
}
.confirm-icon-wrap {
  margin-bottom: var(--space-md);
  display: flex;
  justify-content: center;
  align-items: center;
}
.confirm-icon-wrap--warning .confirm-icon {
  color: var(--color-warning);
}
.confirm-icon-wrap--error .confirm-icon {
  color: var(--color-danger);
}
.confirm-icon-wrap--info .confirm-icon {
  color: var(--color-primary);
}
.confirm-message {
  margin: 0 0 var(--space-xl);
  font-size: 1rem;
  line-height: 1.5;
  color: var(--color-text);
  white-space: pre-wrap;
  word-break: break-word;
}
.confirm-actions {
  display: flex;
  gap: var(--space-md);
  justify-content: center;
  flex-wrap: wrap;
}
.confirm-btn {
  min-width: 100px;
}
.confirm-btn-ok {
  flex: 1;
}
</style>
