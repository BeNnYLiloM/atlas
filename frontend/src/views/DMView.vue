<script setup lang="ts">
import { ref, watch, nextTick, computed, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useMessagesStore } from '@/stores'
import { useDMStore } from '@/stores/dm'
import { useNavigationStore } from '@/stores/navigation'
import { useCallsStore } from '@/stores/calls'
import { Avatar } from '@/components/ui'
import MessageList from '@/components/chat/MessageList.vue'
import MessageInput from '@/components/chat/MessageInput.vue'
import TypingIndicator from '@/components/chat/TypingIndicator.vue'
import ThreadView from '@/components/chat/ThreadView.vue'
import CallPanel from '@/components/calls/CallPanel.vue'

const route = useRoute()
const messagesStore = useMessagesStore()
const dmStore = useDMStore()
const navigationStore = useNavigationStore()
const callsStore = useCallsStore()

const isInCall = computed(() =>
  callsStore.isInCall && callsStore.currentChannelId === activeChannelId.value,
)

const messageListRef = ref<{ scrollToBottom: () => void } | null>(null)

const activeChannelId = computed(() =>
  typeof route.params.channelId === 'string' ? route.params.channelId : null,
)

// activeDM используется только для заголовка — чат рендерится по activeChannelId
const activeDM = computed(() =>
  activeChannelId.value
    ? dmStore.dmList.find((d) => d.channelId === activeChannelId.value) ?? null
    : null,
)

// Показываем чат как только есть channelId — не ждём activeDM
const showChat = computed(() => !!activeChannelId.value)

const STATUS_COLORS: Record<string, string> = {
  online: '#3fb950',
  away: '#d29922',
  dnd: '#f85149',
  offline: '#6e7681',
}

function statusColor(status: string): string {
  return STATUS_COLORS[status] ?? STATUS_COLORS.offline
}

async function markActiveChannelRead() {
  const channelId = activeChannelId.value
  if (!channelId) return
  const msgs = messagesStore.messagesByChannel[channelId]
  const lastMsgId = msgs && msgs.length > 0 ? msgs[msgs.length - 1].id : undefined
  await dmStore.clearUnread(channelId, lastMsgId)
}

// При смене канала — грузим сообщения, сбрасываем unread, скроллим вниз
watch(
  activeChannelId,
  async (channelId, prevChannelId) => {
    if (prevChannelId) dmStore.activeChannelId = null
    if (!channelId) return
    dmStore.activeChannelId = channelId
    await messagesStore.fetchMessages(channelId)
    await markActiveChannelRead()
    await nextTick()
    await nextTick()
    messageListRef.value?.scrollToBottom()
  },
  { immediate: true },
)

// При возврате на DM-секцию (переключение воркспейс→директ) — сбрасываем unread снова
watch(
  () => navigationStore.activeSection,
  (section) => {
    if (section === 'dm' && activeChannelId.value) {
      void markActiveChannelRead()
    }
  },
)

onUnmounted(() => {
  dmStore.activeChannelId = null
})
</script>

<template>
  <div class="flex-1 flex min-h-0">
    <div class="flex-1 flex flex-col min-h-0">
      <template v-if="showChat && activeChannelId">
        <!-- DM Header -->
        <div class="h-12 flex items-center gap-3 px-4 border-b border-subtle shrink-0">
          <template v-if="activeDM">
            <div class="relative">
              <Avatar
                :name="activeDM.peer.displayName"
                :src="activeDM.peer.avatarUrl ?? undefined"
                size="sm"
              />
              <span
                class="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border-2 border-[var(--bg-base)]"
                :style="{ background: statusColor(activeDM.peer.status) }"
              />
            </div>
            <span class="font-semibold text-primary flex-1">{{ activeDM.peer.displayName }}</span>

            <!-- Кнопка звонка -->
            <button
              :title="isInCall ? 'Завершить звонок' : 'Позвонить'"
              class="p-1.5 rounded-lg transition-colors"
              :class="isInCall
                ? 'text-red-400 bg-red-600/20 hover:bg-red-600 hover:text-white'
                : 'text-muted hover:text-primary hover:bg-elevated'"
              @click="isInCall ? callsStore.leaveCall() : callsStore.joinCall(activeChannelId!, true, true)"
            >
              <svg
                v-if="!isInCall"
                class="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
              </svg>
              <svg
                v-else
                class="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 8l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2M5 3a2 2 0 00-2 2v1c0 8.284 6.716 15 15 15h1a2 2 0 002-2v-3.28a1 1 0 00-.684-.948l-4.493-1.498a1 1 0 00-1.21.502l-1.13 2.257a11.042 11.042 0 01-5.516-5.517l2.257-1.128a1 1 0 00.502-1.21L9.228 3.683A1 1 0 008.279 3H5z" />
              </svg>
            </button>
          </template>
          <div v-else class="w-8 h-8 rounded-full bg-elevated animate-pulse" />
        </div>

        <!-- Messages -->
        <MessageList
          ref="messageListRef"
          :channel-id="activeChannelId"
          class="flex-1 min-h-0"
        />

        <TypingIndicator :channel-id="activeChannelId" />

        <MessageInput :channel-id="activeChannelId" :slowmode-seconds="0" />

        <CallPanel />
      </template>

      <!-- Empty state -->
      <div v-else class="flex-1 flex items-center justify-center">
        <div class="text-center">
          <div class="w-20 h-20 mx-auto mb-4 rounded-2xl bg-elevated flex items-center justify-center">
            <svg class="w-10 h-10 text-subtle" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
          </div>
          <h3 class="text-lg font-semibold text-secondary mb-1">Личные сообщения</h3>
          <p class="text-subtle text-sm">Выберите диалог из списка слева</p>
        </div>
      </div>
    </div>

    <!-- Thread sidebar -->
    <ThreadView />
  </div>
</template>
