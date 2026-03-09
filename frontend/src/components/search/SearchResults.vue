<script setup lang="ts">
import { useSearchStore } from '@/stores/search'
import { useRouter } from 'vue-router'
import { useChannelsStore } from '@/stores/channels'

const searchStore = useSearchStore()
const channelsStore = useChannelsStore()
const router = useRouter()

function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' })
}

function goToMessage(channelId: string) {
  channelsStore.setCurrentChannel(channelId)
  router.push(`/channels/${channelId}`)
  searchStore.close()
}

function getChannelName(channelId: string): string {
  const ch = channelsStore.channels.find(c => c.id === channelId)
  return ch ? `#${ch.name}` : '#канал'
}
</script>

<template>
  <div class="max-h-96 overflow-y-auto">
    <!-- Results list -->
    <div v-if="searchStore.results.length > 0">
      <div class="px-3 py-2 text-xs text-dark-500 font-medium sticky top-0 bg-dark-900 border-b border-dark-800">
        Найдено {{ searchStore.total }} сообщений
      </div>
      <div
        v-for="result in searchStore.results"
        :key="result.message.id"
        class="px-4 py-3 hover:bg-dark-800 cursor-pointer transition-colors border-b border-dark-800/50 last:border-0"
        @click="goToMessage(result.message.channel_id)"
      >
        <div class="flex items-center gap-2 mb-1">
          <span class="text-xs font-medium text-atlas-400">{{ getChannelName(result.message.channel_id) }}</span>
          <span class="text-xs text-dark-600">·</span>
          <span class="text-xs text-dark-500">{{ result.message.user?.display_name }}</span>
          <span class="text-xs text-dark-600 ml-auto">{{ formatDate(result.message.created_at) }}</span>
        </div>
        <!-- Highlighted text -->
        <p
          class="text-sm text-dark-300 line-clamp-2"
          v-html="result.highlight || result.message.content"
        />
      </div>
    </div>

    <!-- No results -->
    <div
      v-else-if="searchStore.query && !searchStore.loading"
      class="px-4 py-8 text-center"
    >
      <svg class="w-10 h-10 text-dark-600 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <p class="text-dark-500 text-sm">Ничего не найдено по запросу «{{ searchStore.query }}»</p>
    </div>
  </div>
</template>
