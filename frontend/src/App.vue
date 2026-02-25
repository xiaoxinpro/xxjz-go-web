<template>
  <div class="app">
    <router-view />
    <AlertModal />
    <ConfirmModal />
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import AlertModal from './components/AlertModal.vue'
import ConfirmModal from './components/ConfirmModal.vue'
import { useAppStore } from './stores/app'

const appStore = useAppStore()
appStore.loadVersion()

watch(() => appStore.title, (t) => {
  document.title = t || '小歆记账'
}, { immediate: true })
</script>

<style>
@import './assets/styles/variables.css';
@import './assets/styles/base.css';
</style>
