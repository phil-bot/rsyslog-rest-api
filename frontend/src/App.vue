<template>
  <div :data-theme="theme" class="app-root">
    <RouterView />
  </div>
</template>

<script setup>
import { ref, provide, onMounted } from 'vue'

const theme = ref('light')

onMounted(() => {
  const saved = localStorage.getItem('rsyslox_theme')
  if (saved) {
    theme.value = saved
  } else if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
    theme.value = 'dark'
  }
})

provide('theme', theme)
provide('toggleTheme', () => {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
  localStorage.setItem('rsyslox_theme', theme.value)
})
</script>

<style>
.app-root {
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>
