<script setup lang="ts">
import { ref } from 'vue'
import { reactionsApi, QUICK_EMOJIS } from '@/api/reactions'
import type { ReactionGroup } from '@/api/reactions'
import { useWorkspaceStore } from '@/stores/workspace'

const props = defineProps<{
  messageId: string
  reactions: ReactionGroup[]
}>()

const emit = defineEmits<{
  'reaction-change': [messageId: string]
}>()

const workspaceStore = useWorkspaceStore()
const showPicker = ref(false)

async function toggleReaction(emoji: string) {
  const workspaceId = workspaceStore.currentWorkspaceId ?? ''
  const existing = props.reactions.find(r => r.emoji === emoji)

  try {
    if (existing?.mine) {
      await reactionsApi.remove(props.messageId, emoji, workspaceId)
    } else {
      await reactionsApi.add(props.messageId, emoji, workspaceId)
    }
    emit('reaction-change', props.messageId)
  } catch (e) {
    console.error('Failed to toggle reaction:', e)
  }

  showPicker.value = false
}

function getReactionTitle(r: ReactionGroup): string {
  if (r.user_ids.length === 0) return r.emoji
  const names = r.user_ids.slice(0, 3).join(', ')
  return `${names}${r.user_ids.length > 3 ? ' и ещё...' : ''} поставили ${r.emoji}`
}
</script>

<template>
  <div class="flex items-center gap-1 flex-wrap">
    <!-- Существующие реакции -->
    <button
      v-for="reaction in reactions"
      :key="reaction.emoji"
      class="flex items-center gap-1 px-2 py-0.5 rounded-full text-sm border transition-all"
      :class="[
        reaction.mine
          ? 'bg-atlas-600/20 border-atlas-600/50 text-atlas-300'
          : 'bg-dark-800 border-dark-700 text-dark-300 hover:border-dark-500'
      ]"
      :title="getReactionTitle(reaction)"
      :aria-label="`Реакция ${reaction.emoji}: ${reaction.count}`"
      @click="toggleReaction(reaction.emoji)"
    >
      <span>{{ reaction.emoji }}</span>
      <span class="text-xs font-medium">{{ reaction.count }}</span>
    </button>

    <!-- Кнопка добавить реакцию -->
    <div class="relative">
      <button
        class="p-1 rounded-full text-dark-600 hover:text-dark-400 hover:bg-dark-800 transition-colors"
        aria-label="Добавить реакцию"
        @click="showPicker = !showPicker"
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
            d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
      </button>

      <!-- Quick emoji picker -->
      <div
        v-if="showPicker"
        v-click-outside="() => showPicker = false"
        class="absolute bottom-7 left-0 z-10 bg-dark-800 border border-dark-700 rounded-xl p-2 shadow-xl flex gap-1 flex-wrap w-48"
      >
        <button
          v-for="emoji in QUICK_EMOJIS"
          :key="emoji"
          class="w-8 h-8 flex items-center justify-center rounded-lg text-lg hover:bg-dark-700 transition-colors"
          :aria-label="`Поставить реакцию ${emoji}`"
          @click="toggleReaction(emoji)"
        >
          {{ emoji }}
        </button>
      </div>
    </div>
  </div>
</template>

