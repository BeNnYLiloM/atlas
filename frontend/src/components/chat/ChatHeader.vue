<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Channel } from '@/types'
import InviteMemberModal from '@/components/workspace/InviteMemberModal.vue'
import ChannelSettingsModal from '@/components/chat/ChannelSettingsModal.vue'
import { useAuthStore, useWorkspaceStore, useUIStore } from '@/stores'

defineProps<{
  channel: Channel
}>()

const authStore = useAuthStore()
const workspaceStore = useWorkspaceStore()
const uiStore = useUIStore()

const showInviteModal = ref(false)
const showSettings = ref(false)

const isAdmin = computed(() => {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return false
  const members = workspaceStore.membersMap[wsId] ?? []
  const role = members.find(m => m.user_id === authStore.user?.id)?.role
  return role === 'owner' || role === 'admin'
})
</script>

<template>
  <header class="h-14 px-4 flex items-center gap-3 border-b border-subtle bg-surface/50 backdrop-blur-sm">
    <!-- Channel icon + name + topic -->
    <div class="flex items-center gap-2 min-w-0">
      <svg
        v-if="channel.type === 'text'"
        class="w-5 h-5 text-muted shrink-0"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14"
        />
      </svg>
      <svg
        v-else
        class="w-5 h-5 text-muted shrink-0"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z"
        />
      </svg>
      <h1 class="font-semibold text-primary shrink-0">
        {{ channel.name }}
      </h1>
      <svg
        v-if="channel.is_private"
        class="w-4 h-4 text-subtle shrink-0"
        fill="currentColor"
        viewBox="0 0 20 20"
      >
        <path
          fill-rule="evenodd"
          d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z"
          clip-rule="evenodd"
        />
      </svg>
      <!-- Topic separator + text -->
      <template v-if="channel.topic">
        <span class="text-faint shrink-0">|</span>
        <span class="text-sm text-muted truncate max-w-xs">{{ channel.topic }}</span>
      </template>
    </div>

    <div class="flex-1" />

    <!-- Actions -->
    <div class="flex items-center gap-1">
      <template v-if="isAdmin">
        <button
          class="btn-ghost p-2 rounded-lg text-muted hover:text-primary"
          title="Пригласить участника"
          @click="showInviteModal = true"
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
              d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z"
            />
          </svg>
        </button>
        <button
          class="btn-ghost p-2 rounded-lg transition-colors"
          :class="showSettings ? 'text-accent' : 'text-muted hover:text-primary'"
          title="Настройки канала"
          @click="showSettings = true"
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
              d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
            />
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
            />
          </svg>
        </button>
      </template>
      <button
        class="btn-ghost p-2 rounded-lg transition-colors"
        :class="uiStore.memberListVisible ? 'text-accent' : 'text-muted hover:text-primary'"
        title="Участники канала"
        @click="uiStore.toggleMemberList()"
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
            d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
          />
        </svg>
      </button>
    </div>

    <InviteMemberModal
      :open="showInviteModal"
      @close="showInviteModal = false"
    />
    <ChannelSettingsModal
      :open="showSettings"
      :channel-id="channel.id"
      @close="showSettings = false"
    />
  </header>
</template>


