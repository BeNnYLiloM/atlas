<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import data from '@emoji-mart/data'
import { Picker } from 'emoji-mart'

type TabType = 'emoji' | 'gif' | 'sticker'

interface GiphyResult {
  id: string
  title: string
  images: {
    fixed_width_small: { url: string; webp: string }
    original: { url: string }
  }
}

const emit = defineEmits<{
  select: [payload: { type: 'emoji' | 'gif' | 'sticker'; value: string }]
}>()

const GIPHY_KEY = import.meta.env.VITE_GIPHY_API_KEY as string

const activeTab = ref<TabType>('emoji')
const containerRef = ref<HTMLDivElement | null>(null)
const searchQuery = ref('')
const gifResults = ref<GiphyResult[]>([])
const stickerResults = ref<GiphyResult[]>([])
const loading = ref(false)

let picker: InstanceType<typeof Picker> | null = null
let searchTimeout: ReturnType<typeof setTimeout> | null = null

function mountPicker() {
  if (!containerRef.value || picker) return
  picker = new Picker({
    data,
    locale: 'ru',
    theme: 'dark',
    onEmojiSelect: (emoji: { native: string }) => {
      emit('select', { type: 'emoji', value: emoji.native })
    },
  })
  containerRef.value.appendChild(picker as unknown as Node)
}

function unmountPicker() {
  if (containerRef.value && picker) {
    try {
      containerRef.value.removeChild(picker as unknown as Node)
    } catch {
      // already removed
    }
    picker = null
  }
}

async function fetchGiphy(type: 'gif' | 'sticker', q: string) {
  loading.value = true
  try {
    const base = type === 'sticker'
      ? 'https://api.giphy.com/v1/stickers'
      : 'https://api.giphy.com/v1/gifs'
    const endpoint = q
      ? `${base}/search?api_key=${GIPHY_KEY}&q=${encodeURIComponent(q)}&limit=30&rating=pg`
      : `${base}/trending?api_key=${GIPHY_KEY}&limit=30&rating=pg`

    const res = await fetch(endpoint)
    const json = await res.json()
    if (type === 'gif') gifResults.value = json.data ?? []
    else stickerResults.value = json.data ?? []
  } catch {
    if (type === 'gif') gifResults.value = []
    else stickerResults.value = []
  } finally {
    loading.value = false
  }
}

function onSearch() {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    if (activeTab.value === 'gif') fetchGiphy('gif', searchQuery.value)
    else if (activeTab.value === 'sticker') fetchGiphy('sticker', searchQuery.value)
  }, 400)
}

function getThumb(item: GiphyResult): string {
  return item.images.fixed_width_small?.webp ?? item.images.fixed_width_small?.url ?? item.images.original.url
}

function selectMedia(item: GiphyResult, type: 'gif' | 'sticker') {
  emit('select', { type, value: item.images.original.url })
}

watch(activeTab, (tab) => {
  searchQuery.value = ''
  if (tab === 'emoji') {
    unmountPicker()
    setTimeout(mountPicker, 50)
  } else {
    unmountPicker()
    fetchGiphy(tab, '')
  }
})

onMounted(() => {
  mountPicker()
})

onUnmounted(() => {
  unmountPicker()
  if (searchTimeout) clearTimeout(searchTimeout)
})
</script>

<template>
  <div
    class="w-[360px] bg-surface border border-default rounded-xl flex flex-col"
    style="height: 420px;"
  >
    <!-- Tab bar -->
    <div class="flex border-b border-default px-2 pt-2 gap-1 shrink-0">
      <button
        v-for="tab in (['emoji', 'gif', 'sticker'] as TabType[])"
        :key="tab"
        class="px-3 py-1.5 text-xs font-medium rounded-t transition-colors"
        :class="activeTab === tab
          ? 'text-accent border-b-2 border-accent -mb-px bg-elevated'
          : 'text-muted hover:text-secondary'"
        @click="activeTab = tab"
      >
        <span v-if="tab === 'emoji'">😀 Эмодзи</span>
        <span v-else-if="tab === 'gif'">GIF</span>
        <span v-else>Стикеры</span>
      </button>
    </div>

    <!-- Search (GIF / Sticker) -->
    <div
      v-if="activeTab !== 'emoji'"
      class="px-3 pt-2 pb-1 shrink-0"
    >
      <input
        v-model="searchQuery"
        type="text"
        :placeholder="activeTab === 'gif' ? 'Поиск GIF...' : 'Поиск стикеров...'"
        class="w-full bg-elevated border border-default rounded-lg px-3 py-1.5 text-sm text-primary placeholder:text-subtle focus:outline-none focus:border-accent/60"
        @input="onSearch"
      >
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-hidden">
      <!-- Emoji tab -->
      <div
        v-show="activeTab === 'emoji'"
        ref="containerRef"
        class="emoji-picker-host h-full"
      />

      <!-- GIF / Sticker grid -->
      <div
        v-if="activeTab !== 'emoji'"
        class="h-full overflow-y-auto p-2"
      >
        <div
          v-if="loading"
          class="flex items-center justify-center h-full"
        >
          <svg
            class="animate-spin w-6 h-6 text-accent"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            />
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
            />
          </svg>
        </div>
        <div
          v-else-if="(activeTab === 'gif' ? gifResults : stickerResults).length === 0"
          class="flex items-center justify-center h-full text-subtle text-sm"
        >
          Ничего не найдено
        </div>
        <div
          v-else
          class="columns-3 gap-1.5 space-y-1.5"
        >
          <div
            v-for="item in (activeTab === 'gif' ? gifResults : stickerResults)"
            :key="item.id"
            class="break-inside-avoid cursor-pointer rounded-lg overflow-hidden hover:ring-2 hover:ring-[var(--accent)] transition-all"
            @click="selectMedia(item, activeTab as 'gif' | 'sticker')"
          >
            <img
              :src="getThumb(item)"
              :alt="item.title"
              class="w-full object-cover"
              loading="lazy"
            >
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.emoji-picker-host :deep(em-emoji-picker) {
  --border-radius: 0px;
  --background-rgb: 17, 17, 28;
  --category-icon-size: 18px;
  width: 100%;
  height: 100%;
  max-height: 380px;
  border: none;
}
</style>
