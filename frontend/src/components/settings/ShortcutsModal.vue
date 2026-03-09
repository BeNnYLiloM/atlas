<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useUIStore, SHORTCUTS } from '@/stores/ui'

const uiStore = useUIStore()

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && uiStore.shortcutsVisible) {
    uiStore.shortcutsVisible = false
  }
  if ((e.ctrlKey || e.metaKey) && e.key === '/') {
    e.preventDefault()
    uiStore.toggleShortcuts()
  }
}

onMounted(() => document.addEventListener('keydown', onKeydown))
onUnmounted(() => document.removeEventListener('keydown', onKeydown))
</script>

<template>
  <Teleport to="body">
    <div
      v-if="uiStore.shortcutsVisible"
      class="fixed inset-0 z-50 flex items-center justify-center px-4"
      role="dialog"
      aria-modal="true"
      aria-label="Горячие клавиши"
      @click.self="uiStore.shortcutsVisible = false"
    >
      <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" />
      <div class="relative bg-dark-900 rounded-2xl border border-dark-700 shadow-2xl w-full max-w-md p-6">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-semibold text-dark-100">Горячие клавиши</h2>
          <button
            class="text-dark-500 hover:text-dark-300 transition-colors"
            aria-label="Закрыть"
            @click="uiStore.shortcutsVisible = false"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="space-y-2">
          <div
            v-for="(shortcut, idx) in SHORTCUTS"
            :key="idx"
            class="flex items-center justify-between py-2 border-b border-dark-800 last:border-0"
          >
            <span class="text-sm text-dark-300">{{ shortcut.description }}</span>
            <div class="flex items-center gap-1">
              <kbd
                v-for="key in shortcut.keys"
                :key="key"
                class="px-2 py-0.5 text-xs bg-dark-700 text-dark-300 rounded border border-dark-600 font-mono"
              >{{ key }}</kbd>
            </div>
          </div>
        </div>

        <p class="mt-4 text-xs text-dark-500 text-center">Нажмите Ctrl+/ чтобы закрыть</p>
      </div>
    </div>
  </Teleport>
</template>
