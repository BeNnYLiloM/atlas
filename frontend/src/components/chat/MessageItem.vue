<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import type { Message } from '@/types'
import { Avatar } from '@/components/ui'
import { useThreadStore } from '@/stores'
import ReactionBar from './ReactionBar.vue'
import ReactionEmojiPicker from './ReactionEmojiPicker.vue'
import TaskCreationModal from '@/components/tasks/TaskCreationModal.vue'
import { reactionsApi } from '@/api/reactions'
import { useWorkspaceStore } from '@/stores/workspace'
import { useAuthStore } from '@/stores/auth'
import { useDMStore } from '@/stores/dm'
import type { ReactionGroup } from '@/api/reactions'

const props = defineProps<{
  message: Message
  highlighted?: boolean
}>()

const threadStore = useThreadStore()
const workspaceStore = useWorkspaceStore()
const authStore = useAuthStore()
const dmStore = useDMStore()

function openAuthorDM() {
  const authorId = props.message.user_id
  if (!authorId || authorId === authStore.user?.id) return
  void dmStore.openDM(authorId)
}
const reactions = ref<ReactionGroup[]>([])
const showTaskModal = ref(false)
const showReactionPicker = ref(false)
const reactionBtnRef = ref<HTMLButtonElement | null>(null)
const pickerStyle = ref({ top: '0px', left: '0px' })

function openReactionPicker() {
  if (showReactionPicker.value) {
    showReactionPicker.value = false
    return
  }
  const btn = reactionBtnRef.value
  if (!btn) return
  const rect = btn.getBoundingClientRect()
  const pickerH = 340
  const pickerW = 320
  // Открываем вверх если не хватает места снизу
  const top = rect.bottom + pickerH + 8 > window.innerHeight
    ? rect.top - pickerH - 4
    : rect.bottom + 4
  // Прижимаем к левому краю если выходит за правый
  const left = rect.right - pickerW < 0
    ? rect.left
    : rect.right - pickerW
  pickerStyle.value = { top: `${top}px`, left: `${left}px` }
  showReactionPicker.value = true
}

async function loadReactions() {
  try {
    reactions.value = await reactionsApi.getGrouped(props.message.id)
  } catch {
    // Не критично
  }
}

async function onReactionChange() {
  await loadReactions()
}

async function addReaction(emoji: string) {
  showReactionPicker.value = false
  const workspaceId = workspaceStore.currentWorkspaceId ?? ''
  try {
    await reactionsApi.add(props.message.id, emoji, workspaceId)
    await loadReactions()
  } catch {
    // Не критично
  }
}

onMounted(() => {
  loadReactions()
})

const time = computed(() => {
  return new Date(props.message.created_at).toLocaleTimeString('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
  })
})

const displayName = computed(() => {
  return props.message.user?.display_name || 'Пользователь'
})

// Thread stats для превью
const threadStats = computed(() => {
  return threadStore.getThreadStats(props.message.id)
})

const hasThread = computed(() => threadStats.value.count > 0)

const hasUnreadInThread = computed(() => threadStats.value.unreadCount > 0)

function openThread() {
  threadStore.openThread(props.message.id)
}

const isSingleEmoji = computed(() => {
  const c = props.message.content.trim()
  const emojiRegex = /^(\p{Emoji_Presentation}|\p{Extended_Pictographic})(\u200d(\p{Emoji_Presentation}|\p{Extended_Pictographic})|\uFE0F|\u20E3)*$/u
  return emojiRegex.test(c)
})

const isGiphyMedia = computed(() => {
  const c = props.message.content.trim()
  return /^https:\/\/media\d*\.giphy\.com\/.+\.gif(\?.*)?$/.test(c)
})

// Простой парсинг markdown (жирный, курсив, код)
const formattedContent = computed(() => {
  let content = props.message.content
  // Экранируем HTML
  content = content.replace(/</g, '&lt;').replace(/>/g, '&gt;')
  // Код в бэктиках
  content = content.replace(/`([^`]+)`/g, '<code class="px-1.5 py-0.5 rounded bg-elevated text-accent-strong font-mono text-sm">$1</code>')
  // Жирный
  content = content.replace(/\*\*([^*]+)\*\*/g, '<strong class="font-semibold">$1</strong>')
  // Курсив
  content = content.replace(/\*([^*]+)\*/g, '<em>$1</em>')
  // @упоминания — до эмодзи и ссылок, чтобы не конфликтовать
  const myName = authStore.user?.display_name ?? ''
  content = content.replace(/@([\w\u0400-\u04FF][^\s@]*)/gu, (match, name) => {
    const isSelf = myName && name === myName
    return isSelf
      ? `<span class="mention mention-self">${match}</span>`
      : `<span class="mention">${match}</span>`
  })
  // Эмодзи — оборачиваем в span для увеличения размера
  content = content.replace(
    /(\p{Emoji_Presentation}|\p{Extended_Pictographic})(\u200d(\p{Emoji_Presentation}|\p{Extended_Pictographic})|\uFE0F|\u20E3)*/gu,
    '<span class="emoji">$&</span>'
  )
  // Ссылки
  content = content.replace(
    /(https?:\/\/[^\s]+)/g,
    '<a href="$1" target="_blank" rel="noopener" class="text-accent hover:text-accent-strong underline">$1</a>'
  )
  return content
})

// Сообщение упоминает текущего пользователя
const isMentioned = computed(() => {
  const myName = authStore.user?.display_name
  if (!myName) return false
  return props.message.content.includes(`@${myName}`) || props.message.content.includes('@everyone')
})

const isCallMessage = computed(() => props.message.type === 'call')

const callLabel = computed(() => {
  if (!isCallMessage.value) return ''
  const status = props.message.call_status
  if (status === 'ringing') {
    return 'Звоним...'
  }
  if (status === 'cancelled') {
    return 'Звонок отменён'
  }
  if (status === 'missed') {
    const isMe = props.message.user_id === authStore.user?.id
    return isMe ? 'Звонок не принят' : 'Пропущенный звонок'
  }
  if (status === 'ongoing') {
    return 'Звонок идёт...'
  }
  // ended
  const dur = props.message.call_duration_sec
  if (dur != null) {
    if (dur < 60) return `Звонок · ${dur} сек`
    const m = Math.floor(dur / 60)
    const s = dur % 60
    return s > 0 ? `Звонок · ${m} мин ${s} сек` : `Звонок · ${m} мин`
  }
  return 'Звонок завершён'
})
</script>

<template>
  <div
    class="group flex gap-3 py-1 px-2 -mx-2 rounded-lg transition-colors relative"
    :class="[
      showReactionPicker ? 'bg-elevated/50' : 'hover:bg-elevated/50',
      isMentioned ? 'mentioned-message' : '',
      highlighted ? 'search-highlighted' : '',
    ]"
  >
    <Avatar
      :name="displayName"
      :src="message.user?.avatar_url"
      size="sm"
    />

    <div class="flex-1 min-w-0">
      <div class="flex items-baseline gap-2">
        <span
          class="font-semibold text-primary text-sm hover:underline cursor-pointer"
          :class="{ 'cursor-default hover:no-underline': message.user_id === authStore.user?.id }"
          @click="openAuthorDM"
        >{{ displayName }}</span>
        <span class="text-xs text-subtle">{{ time }}</span>
      </div>
      <!-- Call-сообщение -->
      <div
        v-if="isCallMessage"
        class="mt-1 inline-flex items-center gap-2 px-3 py-2 rounded-lg text-sm"
        :class="message.call_status === 'missed' || message.call_status === 'cancelled'
          ? 'bg-red-500/10 text-red-400'
          : message.call_status === 'ringing' || message.call_status === 'ongoing'
            ? 'bg-green-500/10 text-green-400'
            : 'bg-elevated text-secondary'"
      >
        <svg
          class="w-4 h-4 shrink-0"
          fill="currentColor"
          viewBox="0 0 24 24"
        >
          <path d="M6.62 10.79a15.09 15.09 0 006.59 6.59l2.2-2.2a1 1 0 011.01-.24 11.47 11.47 0 003.58.57 1 1 0 011 1V21a1 1 0 01-1 1A17 17 0 013 5a1 1 0 011-1h3.5a1 1 0 011 1 11.47 11.47 0 00.57 3.58 1 1 0 01-.25 1.01l-2.2 2.2z" />
        </svg>
        <span>{{ callLabel }}</span>
      </div>

      <img
        v-else-if="isGiphyMedia"
        :src="message.content.trim()"
        class="mt-1 rounded-lg max-w-xs max-h-48 object-contain"
        loading="lazy"
      >
      <div
        v-else-if="isSingleEmoji"
        class="text-5xl leading-none mt-0.5 select-none"
      >
        {{ message.content.trim() }}
      </div>
      <div
        v-else
        class="text-secondary text-sm leading-relaxed break-words"
        v-html="formattedContent"
      />
      
      <!-- Reactions -->
      <div
        v-if="reactions.length > 0"
        class="mt-1"
      >
        <ReactionBar
          :message-id="message.id"
          :reactions="reactions"
          @reaction-change="onReactionChange"
        />
      </div>

      <!-- Thread preview -->
      <button
        v-if="hasThread"
        class="mt-2 flex items-center gap-2 text-xs text-accent hover:text-accent-strong transition-colors"
        @click="openThread"
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
            d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z"
          />
        </svg>
        <span class="font-medium">{{ threadStats.count }} {{ threadStats.count === 1 ? 'ответ' : 'ответа' }}</span>
        <span
          v-if="threadStats.lastReplyUser"
          class="text-subtle"
        >{{ threadStats.lastReplyUser }}</span>
        <!-- Unread indicator -->
        <span 
          v-if="hasUnreadInThread" 
          class="ml-1 px-1.5 py-0.5 text-xs font-bold bg-accent text-white rounded-full"
        >
          {{ threadStats.unreadCount }}
        </span>
      </button>
    </div>

    <!-- Actions (visible on hover) -->
    <div
      class="transition-opacity flex items-start gap-0.5"
      :class="showReactionPicker ? 'opacity-100' : 'opacity-0 group-hover:opacity-100'"
    >
      <div class="reaction-picker-wrapper">
        <button
          ref="reactionBtnRef"
          class="p-2 rounded text-subtle hover:text-primary hover:bg-overlay"
          :class="{ 'text-accent bg-overlay': showReactionPicker }"
          title="Реакция"
          @click.stop="openReactionPicker"
        >
          <svg
            class="w-5 h-5"
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
      </div>

      <!-- Teleport пикера в body — фиксированное позиционирование -->
      <Teleport to="body">
        <template v-if="showReactionPicker">
          <!-- Оверлей блокирует hover и клики вне пикера -->
          <div
            class="fixed inset-0 z-[998]"
            @click="showReactionPicker = false"
          />
          <div
            class="fixed z-[999] shadow-2xl rounded-xl overflow-hidden border border-default"
            :style="pickerStyle"
          >
            <ReactionEmojiPicker @select="addReaction($event)" />
          </div>
        </template>
      </Teleport>
      <button
        class="p-2 rounded text-subtle hover:text-primary hover:bg-overlay"
        title="Создать тред"
        @click="openThread"
      >
        <svg
          class="w-5 h-5"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z"
          />
        </svg>
      </button>
      <button
        class="p-2 rounded text-subtle hover:text-primary hover:bg-overlay"
        title="Создать задачу из сообщения"
        @click="showTaskModal = true"
      >
        <svg
          class="w-5 h-5"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
          />
        </svg>
      </button>
    </div>

    <!-- Task creation modal -->
    <TaskCreationModal
      v-if="showTaskModal"
      :message-id="message.id"
      :initial-title="message.content.substring(0, 100)"
      @close="showTaskModal = false"
      @created="showTaskModal = false"
    />
  </div>
</template>

<style scoped>
:deep(.emoji) {
  font-size: 1.75em;
  vertical-align: middle;
  line-height: 0;
  display: inline-block;
}

:deep(.mention) {
  display: inline-flex;
  align-items: center;
  background: rgba(88, 101, 242, 0.15);
  color: #7983f5;
  border-radius: 4px;
  padding: 0 4px;
  font-weight: 500;
  cursor: default;
}

:deep(.mention-self) {
  background: rgba(250, 168, 26, 0.2);
  color: #faa81a;
}

.mentioned-message {
  background: rgba(250, 168, 26, 0.05) !important;
  border-left: 2px solid rgba(250, 168, 26, 0.5);
  padding-left: calc(0.5rem - 2px);
}

.search-highlighted {
  animation: highlight-fade 2.5s ease-out forwards;
}

@keyframes highlight-fade {
  0%   { background-color: rgba(88, 101, 242, 0.25); }
  30%  { background-color: rgba(88, 101, 242, 0.20); }
  100% { background-color: transparent; }
}
</style>
