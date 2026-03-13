<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useSearchStore } from '@/stores/search'
import { useWorkspaceStore } from '@/stores/workspace'
import SearchResults from './SearchResults.vue'

const searchStore = useSearchStore()
const workspaceStore = useWorkspaceStore()
const inputRef = ref<HTMLInputElement | null>(null)

watch(
  () => searchStore.isOpen,
  (open) => {
    if (open) {
      setTimeout(() => inputRef.value?.focus(), 50)
    }
  }
)

function onInput(e: Event) {
  const value = (e.target as HTMLInputElement).value
  searchStore.search(value, workspaceStore.currentWorkspaceId ?? undefined)
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    searchStore.close()
  }
}

// Горячая клавиша Ctrl+K / Cmd+K + глобальный Escape
function onGlobalKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault()
    if (searchStore.isOpen) {
      searchStore.close()
    } else {
      searchStore.open()
    }
  } else if (e.key === 'Escape' && searchStore.isOpen) {
    e.preventDefault()
    searchStore.close()
  }
}

onMounted(() => document.addEventListener('keydown', onGlobalKeydown))
onUnmounted(() => document.removeEventListener('keydown', onGlobalKeydown))
</script>

<template>
  <!-- Trigger button in header -->
  <button
    class="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-elevated border border-default text-muted hover:text-secondary hover:border-strong transition-colors text-sm"
    @click="searchStore.open()"
  >
    <svg
      class="w-4 h-4"
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
      />
    </svg>
    <span>Поиск</span>
    <kbd class="ml-2 text-xs bg-overlay px-1.5 py-0.5 rounded">Ctrl+K</kbd>
  </button>

  <!-- Search modal overlay -->
  <Teleport to="body">
    <div
      v-if="searchStore.isOpen"
      class="fixed inset-0 z-50 flex items-start justify-center pt-20 px-4 bg-black/60 backdrop-blur-sm"
      @click.self="searchStore.close()"
    >
      <!-- Search panel -->
      <div class="w-full max-w-2xl bg-surface rounded-2xl border border-default shadow-2xl overflow-hidden">
        <!-- Input -->
        <div class="flex items-center gap-3 px-4 py-3 border-b border-default">
          <svg
            class="w-5 h-5 text-muted flex-shrink-0"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            />
          </svg>
          <input
            ref="inputRef"
            :value="searchStore.query"
            type="text"
            placeholder="Поиск по сообщениям..."
            class="flex-1 bg-transparent text-primary placeholder:text-subtle focus:outline-none text-base"
            @input="onInput"
            @keydown="onKeydown"
          >
          <div
            v-if="searchStore.loading"
            class="w-4 h-4 border-2 border-accent border-t-transparent rounded-full animate-spin"
          />
          <kbd class="text-xs text-subtle bg-elevated px-1.5 py-0.5 rounded">Esc</kbd>
        </div>

        <!-- Results -->
        <SearchResults />

        <!-- Empty / initial state -->
        <div
          v-if="!searchStore.query && !searchStore.loading"
          class="px-4 py-8 text-center"
        >
          <p class="text-subtle text-sm">
            Введите запрос для поиска по сообщениям
          </p>
        </div>
      </div>
    </div>
  </Teleport>
</template>
