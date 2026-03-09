<script setup lang="ts">
import { useUIStore } from '@/stores/ui'
import type { Theme } from '@/stores/ui'

const uiStore = useUIStore()

const themes: { value: Theme; label: string; icon: string }[] = [
  { value: 'dark', label: 'Тёмная', icon: '🌙' },
  { value: 'light', label: 'Светлая', icon: '☀️' },
  { value: 'high-contrast', label: 'Высокий контраст', icon: '◑' },
]
</script>

<template>
  <div class="flex items-center gap-1 p-1 rounded-lg bg-dark-800 border border-dark-700">
    <button
      v-for="t in themes"
      :key="t.value"
      class="px-2.5 py-1 rounded text-xs transition-all"
      :class="[
        uiStore.theme === t.value
          ? 'bg-atlas-600 text-white'
          : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700'
      ]"
      :aria-label="`Переключить на тему: ${t.label}`"
      @click="uiStore.setTheme(t.value)"
    >
      <span aria-hidden="true">{{ t.icon }}</span>
      <span class="ml-1">{{ t.label }}</span>
    </button>
  </div>
</template>
