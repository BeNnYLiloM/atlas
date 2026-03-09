<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useChannelsStore, useAuthStore, useWorkspaceStore } from '@/stores'
import { useCallsStore } from '@/stores/calls'
import ChannelContextMenu from '@/components/chat/ChannelContextMenu.vue'
import ChannelSettingsModal from '@/components/chat/ChannelSettingsModal.vue'
import type { NotificationLevel } from '@/types'

const router = useRouter()
const route = useRoute()
const channelsStore = useChannelsStore()
const callsStore = useCallsStore()
const authStore = useAuthStore()
const workspaceStore = useWorkspaceStore()

// Контекстное меню
const ctxMenu = ref<{ x: number; y: number; channelId: string } | null>(null)
const settingsChannelId = ref<string | null>(null)
const deleteConfirmChannelId = ref<string | null>(null)
const deleting = ref(false)

// Collapsed state для категорий: categoryId -> boolean
const collapsedCategories = ref<Record<string, boolean>>({})

function toggleCategory(categoryId: string) {
  collapsedCategories.value[categoryId] = !collapsedCategories.value[categoryId]
}

function isCategoryCollapsed(categoryId: string): boolean {
  return !!collapsedCategories.value[categoryId]
}

type ChannelGroup = { id: string | null; name: string | null; isPrivate: boolean; channels: typeof channelsStore.textChannels }

// Сгруппированные каналы по категориям
const groupedTextChannels = computed<ChannelGroup[]>(() => {
  const cats = channelsStore.categories
  const uncategorized = channelsStore.textChannels.filter(ch => !ch.category_id)
  const groups: ChannelGroup[] = []
  if (uncategorized.length > 0 || (isAdmin.value && cats.length === 0)) {
    groups.push({ id: null, name: null, isPrivate: false, channels: uncategorized })
  }
  for (const cat of cats) {
    const catChannels = channelsStore.textChannels.filter(ch => ch.category_id === cat.id)
    groups.push({ id: cat.id, name: cat.name, isPrivate: cat.is_private, channels: catChannels })
  }
  return groups
})

const groupedVoiceChannels = computed<ChannelGroup[]>(() => {
  const cats = channelsStore.categories
  const uncategorized = channelsStore.voiceChannels.filter(ch => !ch.category_id)
  const groups: ChannelGroup[] = []
  if (uncategorized.length > 0) {
    groups.push({ id: null, name: null, isPrivate: false, channels: uncategorized })
  }
  for (const cat of cats) {
    const catChannels = channelsStore.voiceChannels.filter(ch => ch.category_id === cat.id)
    if (catChannels.length > 0) {
      groups.push({ id: cat.id, name: cat.name, isPrivate: cat.is_private, channels: catChannels })
    }
  }
  return groups
})

const emit = defineEmits<{ createChannel: [] }>()

const isAdmin = computed(() => {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return false
  const members = workspaceStore.membersMap[wsId] ?? []
  const role = members.find(m => m.user_id === authStore.user?.id)?.role
  return role === 'owner' || role === 'admin'
})

function openContextMenu(e: MouseEvent, channelId: string) {
  e.preventDefault()
  // Корректируем позицию чтобы меню не выходило за правый/нижний край
  const menuWidth = 224
  const menuHeight = 200
  const x = Math.min(e.clientX, window.innerWidth - menuWidth - 8)
  const y = Math.min(e.clientY, window.innerHeight - menuHeight - 8)
  ctxMenu.value = { x, y, channelId }
}

function closeCtxMenu() {
  ctxMenu.value = null
}

async function handleSetNotification(channelId: string, level: NotificationLevel) {
  await channelsStore.updateNotifications(channelId, level)
}

async function handleMute(channelId: string, minutes: number | null) {
  void minutes
  // Заглушить = выставить 'nothing', разблокировка — отдельная логика
  await channelsStore.updateNotifications(channelId, 'nothing')
}

async function handleMarkRead(channelId: string) {
  await channelsStore.markAsRead(channelId)
}

async function handleDeleteChannel(channelId: string) {
  deleting.value = true
  try {
    await channelsStore.deleteChannel(channelId)
  } finally {
    deleting.value = false
    deleteConfirmChannelId.value = null
  }
}

function selectChannel(channelId: string) {
  channelsStore.setCurrentChannel(channelId)
  router.push(`/channels/${channelId}`)
}

function joinVoiceChannel(channelId: string) {
  callsStore.toggleVoiceChannel(channelId)
}

function isActive(channelId: string): boolean {
  return route.params.channelId === channelId
}

function getUnreadCount(channelId: string): number {
  const channel = channelsStore.channels.find(c => c.id === channelId)
  return channel?.unread_count || 0
}

function getMentionCount(channelId: string): number {
  return channelsStore.getMentionCount(channelId)
}
</script>

<template>
  <div class="p-3 space-y-2">
    <!-- Text channels grouped by category -->
    <template
      v-for="group in groupedTextChannels"
      :key="group.id ?? '__uncategorized_text'"
    >
      <div v-if="group.channels.length > 0 || (isAdmin && group.id !== null)">
        <!-- Category header -->
        <div class="px-2 mb-1 flex items-center justify-between group/header">
          <button
            class="flex items-center gap-1 text-xs font-semibold text-dark-500 uppercase tracking-wider hover:text-dark-300 transition-colors min-w-0"
            @click="group.id ? toggleCategory(group.id) : undefined"
          >
            <svg
              v-if="group.id"
              class="w-3 h-3 shrink-0 transition-transform"
              :class="isCategoryCollapsed(group.id) ? '-rotate-90' : ''"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path
                fill-rule="evenodd"
                d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
                clip-rule="evenodd"
              />
            </svg>
            <svg
              v-if="group.isPrivate"
              class="w-3 h-3 shrink-0 text-dark-500"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path
                fill-rule="evenodd"
                d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z"
                clip-rule="evenodd"
              />
            </svg>
            <span class="truncate">{{ group.name ?? 'Текстовые каналы' }}</span>
          </button>
          <button
            v-if="isAdmin"
            class="opacity-0 group-hover/header:opacity-100 w-4 h-4 flex items-center justify-center text-dark-500 hover:text-white transition-all shrink-0"
            title="Создать канал"
            @click.stop="emit('createChannel')"
          >
            <svg
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              class="w-4 h-4"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2.5"
                d="M12 4v16m8-8H4"
              />
            </svg>
          </button>
        </div>

        <!-- Channel list -->
        <div
          v-if="!group.id || !isCategoryCollapsed(group.id)"
          class="space-y-0.5"
        >
          <button
            v-for="channel in group.channels"
            :key="channel.id"
            class="w-full px-2 py-1.5 flex items-center gap-2 rounded-lg transition-colors text-left"
            :class="[
              isActive(channel.id)
                ? 'bg-dark-700 text-white'
                : 'text-dark-400 hover:text-dark-100 hover:bg-dark-800'
            ]"
            @click="selectChannel(channel.id)"
            @contextmenu="openContextMenu($event, channel.id)"
          >
            <svg
              class="w-5 h-5 flex-shrink-0"
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
            <span
              class="truncate text-sm"
              :class="getUnreadCount(channel.id) > 0 ? 'font-semibold' : ''"
            >
              {{ channel.name }}
            </span>

            <!-- Right side indicators -->
            <div class="ml-auto flex items-center gap-1 shrink-0">
              <template v-if="channelsStore.getNotificationLevel(channel.id) === 'nothing'">
                <svg
                  class="w-3.5 h-3.5 text-dark-500"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  title="Уведомления отключены"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z"
                    clip-rule="evenodd"
                  />
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2"
                  />
                </svg>
              </template>
              <template v-else-if="channelsStore.getNotificationLevel(channel.id) === 'mentions'">
                <span
                  v-if="getMentionCount(channel.id) > 0"
                  class="px-1.5 py-0.5 text-xs font-semibold bg-atlas-500 text-white rounded-full min-w-[20px] text-center"
                >
                  {{ getMentionCount(channel.id) > 99 ? '99+' : getMentionCount(channel.id) }}
                </span>
                <span
                  v-else-if="getUnreadCount(channel.id) > 0"
                  class="w-2 h-2 rounded-full bg-atlas-500 shrink-0"
                />
                <svg
                  class="w-3.5 h-3.5 text-dark-500"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fill-rule="evenodd"
                    d="M14.243 5.757a6 6 0 10-.986 9.284 1 1 0 111.087 1.678A8 8 0 1118 10a3 3 0 01-4.8 2.401A4 4 0 1114 10a1 1 0 102 0c0-1.537-.586-2.987-1.757-4.243zM12 10a2 2 0 10-4 0 2 2 0 004 0z"
                    clip-rule="evenodd"
                  />
                </svg>
              </template>
              <template v-else>
                <span
                  v-if="getUnreadCount(channel.id) > 0"
                  class="px-1.5 py-0.5 text-xs font-semibold bg-atlas-500 text-white rounded-full min-w-[20px] text-center"
                >
                  {{ getUnreadCount(channel.id) > 99 ? '99+' : getUnreadCount(channel.id) }}
                </span>
                <svg
                  v-else-if="channel.is_private"
                  class="w-3.5 h-3.5 text-dark-500"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fill-rule="evenodd"
                    d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z"
                    clip-rule="evenodd"
                  />
                </svg>
              </template>
            </div>
          </button>
        </div>
      </div>
    </template>

    <!-- Voice channels grouped by category -->
    <template
      v-for="group in groupedVoiceChannels"
      :key="(group.id ?? '__uncategorized_voice') + '_voice'"
    >
      <div v-if="group.channels.length > 0 || isAdmin">
        <div class="px-2 mb-1 flex items-center justify-between group/header">
          <button
            class="flex items-center gap-1 text-xs font-semibold text-dark-500 uppercase tracking-wider hover:text-dark-300 transition-colors min-w-0"
            @click="group.id ? toggleCategory(group.id + '_v') : undefined"
          >
            <svg
              v-if="group.id"
              class="w-3 h-3 shrink-0 transition-transform"
              :class="isCategoryCollapsed(group.id + '_v') ? '-rotate-90' : ''"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path
                fill-rule="evenodd"
                d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
                clip-rule="evenodd"
              />
            </svg>
            <span class="truncate">{{ group.name ?? 'Голосовые каналы' }}</span>
          </button>
          <button
            v-if="isAdmin"
            class="opacity-0 group-hover/header:opacity-100 w-4 h-4 flex items-center justify-center text-dark-500 hover:text-white transition-all shrink-0"
            title="Создать канал"
            @click.stop="emit('createChannel')"
          >
            <svg
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              class="w-4 h-4"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2.5"
                d="M12 4v16m8-8H4"
              />
            </svg>
          </button>
        </div>
        <div
          v-if="!group.id || !isCategoryCollapsed(group.id + '_v')"
          class="space-y-0.5"
        >
          <div
            v-for="channel in group.channels"
            :key="channel.id"
          >
            <button
              class="w-full px-2 py-1.5 flex items-center gap-2 rounded-lg transition-colors text-left group"
              :class="[
                callsStore.isInChannel(channel.id)
                  ? 'bg-green-900/30 text-green-400'
                  : 'text-dark-400 hover:text-dark-100 hover:bg-dark-800'
              ]"
              :disabled="callsStore.loading"
              @click="joinVoiceChannel(channel.id)"
            >
              <!-- Иконка канала -->
              <svg
                class="w-4 h-4 flex-shrink-0"
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
              <span class="truncate text-sm flex-1">{{ channel.name }}</span>

              <!-- Спиннер пока подключаемся -->
              <svg
                v-if="callsStore.loading && !callsStore.isInCall"
                class="w-3.5 h-3.5 animate-spin text-atlas-400"
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

              <!-- Иконка "выйти" когда в канале -->
              <svg
                v-else-if="callsStore.isInChannel(channel.id)"
                class="w-3.5 h-3.5 opacity-0 group-hover:opacity-100 transition-opacity text-red-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                title="Выйти из канала"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
                />
              </svg>
            </button>

            <!-- Участники в канале (как в Discord) -->
            <div
              v-if="callsStore.isInChannel(channel.id) && callsStore.participants.length > 0"
              class="ml-6 mt-0.5 space-y-0.5"
            >
              <div
                v-for="participant in callsStore.participants"
                :key="participant"
                class="flex items-center gap-1.5 px-2 py-0.5 text-xs text-dark-400"
              >
                <div class="w-5 h-5 rounded-full bg-dark-700 flex items-center justify-center text-[10px] text-dark-300">
                  {{ participant.slice(0, 1).toUpperCase() }}
                </div>
                <span class="truncate">{{ participant }}</span>
              </div>
            </div>

            <!-- Текущий пользователь в канале -->
            <div
              v-if="callsStore.isInChannel(channel.id)"
              class="ml-6 mt-0.5"
            >
              <div class="flex items-center gap-1.5 px-2 py-0.5 text-xs text-green-400">
                <div class="w-5 h-5 rounded-full bg-green-900/50 flex items-center justify-center">
                  <svg
                    class="w-3 h-3"
                    fill="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path d="M12 14c1.66 0 3-1.34 3-3V5c0-1.66-1.34-3-3-3S9 3.34 9 5v6c0 1.66 1.34 3 3 3z" />
                    <path d="M17 11c0 2.76-2.24 5-5 5s-5-2.24-5-5H5c0 3.53 2.61 6.43 6 6.92V21h2v-3.08c3.39-.49 6-3.39 6-6.92h-2z" />
                  </svg>
                </div>
                <span>Вы</span>
                <span
                  v-if="callsStore.isMuted"
                  class="text-red-400"
                >(без звука)</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- Empty state -->
    <div
      v-if="channelsStore.textChannels.length === 0 && channelsStore.voiceChannels.length === 0 && !channelsStore.loading"
      class="text-center py-8"
    >
      <svg
        class="w-12 h-12 mx-auto text-dark-600 mb-3"
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
      <p class="text-dark-500 text-sm">
        Нет каналов
      </p>
      <p class="text-dark-600 text-xs mt-1">
        Создайте первый канал
      </p>
    </div>

    <!-- Loading state -->
    <div
      v-if="channelsStore.loading"
      class="flex justify-center py-8"
    >
      <svg
        class="animate-spin w-6 h-6 text-atlas-500"
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
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        />
      </svg>
    </div>
  </div>

  <!-- Контекстное меню канала -->
  <ChannelContextMenu
    v-if="ctxMenu"
    :x="ctxMenu.x"
    :y="ctxMenu.y"
    :channel-id="ctxMenu.channelId"
    :notification-level="channelsStore.getNotificationLevel(ctxMenu.channelId)"
    :is-admin="isAdmin"
    @close="closeCtxMenu"
    @set-notification="handleSetNotification(ctxMenu!.channelId, $event)"
    @mute="handleMute(ctxMenu!.channelId, $event)"
    @mark-read="handleMarkRead(ctxMenu!.channelId)"
    @open-settings="settingsChannelId = ctxMenu!.channelId; closeCtxMenu()"
    @delete-channel="deleteConfirmChannelId = ctxMenu!.channelId; closeCtxMenu()"
  />

  <!-- Модал настроек канала (открывается из контекстного меню) -->
  <ChannelSettingsModal
    v-if="settingsChannelId"
    :open="!!settingsChannelId"
    :channel-id="settingsChannelId"
    @close="settingsChannelId = null"
  />

  <!-- Диалог подтверждения удаления канала -->
  <Teleport to="body">
    <div
      v-if="deleteConfirmChannelId"
      class="fixed inset-0 z-[300] flex items-center justify-center bg-black/60"
      @click.self="deleteConfirmChannelId = null"
    >
      <div class="bg-dark-800 border border-dark-600 rounded-xl shadow-2xl p-6 w-[360px]">
        <h3 class="text-base font-semibold text-white mb-2">
          Удалить канал
        </h3>
        <p class="text-sm text-dark-300 mb-5">
          Вы уверены? Канал
          <span class="font-semibold text-white">#{{ channelsStore.channels.find(c => c.id === deleteConfirmChannelId)?.name }}</span>
          и все его сообщения будут удалены навсегда.
        </p>
        <div class="flex gap-3">
          <button
            class="flex-1 px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-lg transition-colors disabled:opacity-50"
            :disabled="deleting"
            @click="handleDeleteChannel(deleteConfirmChannelId!)"
          >
            {{ deleting ? 'Удаление...' : 'Удалить' }}
          </button>
          <button
            class="flex-1 px-4 py-2 text-sm font-medium text-dark-300 border border-dark-600 hover:bg-dark-700 rounded-lg transition-colors"
            @click="deleteConfirmChannelId = null"
          >
            Отмена
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>


