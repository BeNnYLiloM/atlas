<script setup lang="ts">
import { computed } from 'vue'
import { useChannelsStore } from '@/stores/channels'
import { useAuthStore } from '@/stores/auth'
import { useWorkspaceStore } from '@/stores/workspace'

const props = defineProps<{
  channelId: string
}>()

const channelsStore = useChannelsStore()
const authStore = useAuthStore()
const workspaceStore = useWorkspaceStore()

const typingUserIds = computed(() => {
  return channelsStore.getTypingUsers(props.channelId)
    .filter(id => id !== authStore.user?.id)
})

function resolveName(userId: string): string {
  const workspaceId = workspaceStore.currentWorkspaceId
  if (!workspaceId) return userId
  const members = workspaceStore.membersMap[workspaceId] ?? []
  const member = members.find(m => m.user_id === userId)
  return member?.display_name || userId
}

const typingText = computed(() => {
  const ids = typingUserIds.value
  if (ids.length === 0) return ''
  if (ids.length === 1) return `${resolveName(ids[0])} печатает...`
  if (ids.length === 2) return `${resolveName(ids[0])} и ${resolveName(ids[1])} печатают...`
  return `${ids.length} человека печатают...`
})
</script>

<template>
  <div
    v-if="typingUserIds.length > 0"
    class="px-4 py-1 flex items-center gap-2 text-xs text-dark-400 min-h-[24px]"
  >
    <span class="flex gap-0.5 items-center">
      <span
        class="w-1.5 h-1.5 bg-dark-400 rounded-full animate-bounce"
        style="animation-delay: 0ms"
      />
      <span
        class="w-1.5 h-1.5 bg-dark-400 rounded-full animate-bounce"
        style="animation-delay: 150ms"
      />
      <span
        class="w-1.5 h-1.5 bg-dark-400 rounded-full animate-bounce"
        style="animation-delay: 300ms"
      />
    </span>
    <span>{{ typingText }}</span>
  </div>
  <div
    v-else
    class="min-h-[24px]"
  />
</template>
