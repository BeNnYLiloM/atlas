<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { useChannelsStore, useMessagesStore, useUIStore } from '@/stores'
import { channelsApi } from '@/api'
import ChatHeader from '@/components/chat/ChatHeader.vue'
import MessageList from '@/components/chat/MessageList.vue'
import MessageInput from '@/components/chat/MessageInput.vue'
import TypingIndicator from '@/components/chat/TypingIndicator.vue'
import ThreadView from '@/components/chat/ThreadView.vue'
import MemberList from '@/components/chat/MemberList.vue'

const route = useRoute()
const channelsStore = useChannelsStore()
const messagesStore = useMessagesStore()
const uiStore = useUIStore()

const canWrite = ref(true)
const messageListRef = ref<{ scrollToBottom: () => void } | null>(null)

watch(
  () => route.params.channelId,
  async (channelId) => {
    if (channelId && typeof channelId === 'string') {
      await messagesStore.fetchMessages(channelId)

      const messages = messagesStore.getMessages(channelId)
      const lastMessageId = messages.length > 0 ? messages[messages.length - 1].id : undefined

      channelsStore.setCurrentChannel(channelId, lastMessageId)

      // Скроллим вниз после загрузки сообщений
      await nextTick()
      messageListRef.value?.scrollToBottom()

      // Проверяем права на запись
      canWrite.value = await channelsApi.checkCanWrite(channelId).catch(() => true)
    }
  },
  { immediate: true }
)
</script>

<template>
  <div class="flex-1 flex min-h-0">
    <div class="flex-1 flex flex-col min-h-0">
      <template v-if="channelsStore.currentChannel">
        <!-- Header -->
        <ChatHeader :channel="channelsStore.currentChannel" />

        <!-- Messages -->
        <MessageList
          ref="messageListRef"
          :channel-id="channelsStore.currentChannel.id"
          class="flex-1 min-h-0"
        />

        <!-- Typing indicator -->
        <TypingIndicator :channel-id="channelsStore.currentChannel.id" />

        <!-- Input -->
        <MessageInput
          v-if="canWrite"
          :channel-id="channelsStore.currentChannel.id"
          :slowmode-seconds="channelsStore.currentChannel.slowmode_seconds"
        />
        <div
          v-else
          class="px-4 py-3 mx-4 mb-4 bg-dark-800 rounded-lg text-sm text-dark-400 text-center border border-dark-700"
        >
          У вас нет прав для отправки сообщений в этот канал
        </div>
      </template>

      <!-- Empty state -->
      <div
        v-else
        class="flex-1 flex items-center justify-center"
      >
        <div class="text-center">
          <div class="w-20 h-20 mx-auto mb-4 rounded-2xl bg-dark-800 flex items-center justify-center">
            <svg
              class="w-10 h-10 text-dark-500"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="1.5"
                d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
              />
            </svg>
          </div>
          <h3 class="text-lg font-semibold text-dark-200 mb-1">
            Выберите канал
          </h3>
          <p class="text-dark-500 text-sm">
            Выберите канал из списка слева для начала общения
          </p>
        </div>
      </div>
    </div>

    <!-- Thread sidebar -->
    <ThreadView />

    <!-- Member list sidebar -->
    <MemberList v-if="uiStore.memberListVisible && channelsStore.currentChannel" />
  </div>
</template>

