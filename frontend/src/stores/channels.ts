import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { channelsApi, categoriesApi } from '@/api'
import type { Channel, ChannelCategory, ChannelCategoryCreate, ChannelCreate, ChannelUpdate, ChannelWithUnread, NotificationLevel } from '@/types'

function toChannelWithUnread(channel: Channel, unreadCount = 0, mentionCount = 0, notificationLevel: NotificationLevel = 'all'): ChannelWithUnread {
  return {
    ...channel,
    unread_count: unreadCount,
    mention_count: mentionCount,
    notification_level: notificationLevel,
  }
}

export const useChannelsStore = defineStore('channels', () => {
  const channels = ref<ChannelWithUnread[]>([])
  const categories = ref<ChannelCategory[]>([])
  const currentChannelId = ref<string | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Typing indicators: channelId -> array of userIds who are typing
  const typingUsers = ref<Record<string, string[]>>({})

  // Notification levels per channel: channelId -> level
  const notificationLevels = ref<Record<string, NotificationLevel>>({})

  // Mention counts per channel: channelId -> count (упоминания текущего юзера)
  const mentionCounts = ref<Record<string, number>>({})

  const currentChannel = computed(() =>
    channels.value.find(c => c.id === currentChannelId.value) ?? null
  )

  const textChannels = computed(() =>
    channels.value.filter(c => c.type === 'text')
  )

  const voiceChannels = computed(() =>
    channels.value.filter(c => c.type === 'voice')
  )

  async function fetchCategories(workspaceId: string) {
    try {
      const all = await categoriesApi.list(workspaceId)
      // В воркспейс-контексте показываем только категории без project_id
      categories.value = all.filter(c => c.project_id === null)
    } catch {
      categories.value = []
    }
  }

  async function fetchChannels(workspaceId: string) {
    loading.value = true
    error.value = null
    try {
      await fetchCategories(workspaceId)
      const all = await channelsApi.list(workspaceId)
      // В воркспейс-контексте показываем только каналы без project_id
      channels.value = all.filter(ch => ch.project_id === null)
      for (const ch of channels.value) {
        if (ch.notification_level) {
          notificationLevels.value[ch.id] = ch.notification_level
        }
        mentionCounts.value[ch.id] = ch.mention_count ?? 0
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Ошибка загрузки каналов'
    } finally {
      loading.value = false
    }
  }

  async function createCategory(workspaceId: string, data: ChannelCategoryCreate): Promise<ChannelCategory> {
    const cat = await categoriesApi.create(workspaceId, data)
    addCategory(cat)
    return cat
  }

  async function renameCategory(workspaceId: string, categoryId: string, name: string): Promise<void> {
    const updated = await categoriesApi.update(workspaceId, categoryId, { name })
    const idx = categories.value.findIndex(c => c.id === categoryId)
    if (idx !== -1) categories.value[idx] = updated
  }

  async function toggleCategoryPrivacy(workspaceId: string, categoryId: string, isPrivate: boolean): Promise<void> {
    const updated = await categoriesApi.update(workspaceId, categoryId, { is_private: isPrivate })
    const idx = categories.value.findIndex(c => c.id === categoryId)
    if (idx !== -1) categories.value[idx] = updated
  }

  async function deleteCategory(workspaceId: string, categoryId: string): Promise<void> {
    await categoriesApi.delete(workspaceId, categoryId)
    removeCategory(categoryId)
  }

  async function fetchProjectChannels(_workspaceId: string, projectId: string) {
    loading.value = true
    error.value = null
    try {
      const [projectCats, projectChannels] = await Promise.all([
        categoriesApi.listByProject(projectId),
        channelsApi.listByProject(projectId),
      ])
      categories.value = projectCats
      channels.value = projectChannels
      for (const ch of channels.value) {
        if (ch.notification_level) {
          notificationLevels.value[ch.id] = ch.notification_level
        }
        mentionCounts.value[ch.id] = ch.mention_count ?? 0
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Ошибка загрузки каналов проекта'
    } finally {
      loading.value = false
    }
  }

  async function createChannel(data: ChannelCreate) {
    loading.value = true
    error.value = null
    try {
      const channel = await channelsApi.create(data)
      channels.value.push(toChannelWithUnread(channel))
      if (channel.type === 'text') {
        currentChannelId.value = channel.id
      }
      return channel
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Ошибка создания канала'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function deleteChannel(channelId: string) {
    await channelsApi.delete(channelId)
    removeChannel(channelId)
  }

  async function updateChannelSettings(channelId: string, data: ChannelUpdate) {
    const updated = await channelsApi.update(channelId, data)
    const idx = channels.value.findIndex(c => c.id === channelId)
    if (idx !== -1) {
      const current = channels.value[idx]
      channels.value[idx] = toChannelWithUnread(
        updated,
        current.unread_count,
        current.mention_count,
        current.notification_level,
      )
    }
    return updated
  }

  async function updateNotifications(channelId: string, level: NotificationLevel) {
    await channelsApi.updateNotifications(channelId, level)
    notificationLevels.value[channelId] = level

    const channel = channels.value.find(c => c.id === channelId)
    if (channel) {
      channel.notification_level = level
    }
  }

  function getNotificationLevel(channelId: string): NotificationLevel {
    return notificationLevels.value[channelId] ?? 'all'
  }

  async function markAsRead(channelId: string, messageId?: string) {
    try {
      await channelsApi.markAsRead(channelId, messageId)
      // Обновляем unread count локально
      const channel = channels.value.find(c => c.id === channelId)
      if (channel) {
        channel.unread_count = 0
      }
    } catch (e) {
      console.error('Failed to mark channel as read:', e)
    }
  }

  function setCurrentChannel(id: string | null, lastMessageId?: string) {
    currentChannelId.value = id || null
    if (id && lastMessageId) {
      markAsRead(id, lastMessageId)
    }
  }

  // Real-time методы
  function addChannel(channel: Channel) {
    const exists = channels.value.find(c => c.id === channel.id)
    if (!exists) {
      channels.value.push(toChannelWithUnread(channel))
      console.log('[Channels] Added channel:', channel.name)
    }
  }

  function updateChannel(channel: Channel) {
    const index = channels.value.findIndex(c => c.id === channel.id)
    if (index !== -1) {
      const current = channels.value[index]
      channels.value[index] = toChannelWithUnread(
        channel,
        current.unread_count,
        current.mention_count,
        current.notification_level,
      )
      console.log('[Channels] Updated channel:', channel.name)
    }
  }

  function removeChannel(channelId: string) {
    const index = channels.value.findIndex(c => c.id === channelId)
    if (index !== -1) {
      channels.value.splice(index, 1)
      console.log('[Channels] Removed channel:', channelId)

      if (currentChannelId.value === channelId) {
        currentChannelId.value = null
      }
    }
  }

  function incrementUnread(channelId: string) {
    const channel = channels.value.find(c => c.id === channelId)
    if (channel) {
      channel.unread_count++
    }
  }

  function incrementMention(channelId: string) {
    mentionCounts.value[channelId] = (mentionCounts.value[channelId] ?? 0) + 1

    const channel = channels.value.find(c => c.id === channelId)
    if (channel) {
      channel.mention_count = mentionCounts.value[channelId]
    }
  }

  function getMentionCount(channelId: string): number {
    return mentionCounts.value[channelId] ?? 0
  }

  function clearUnread(channelId: string) {
    const channel = channels.value.find(c => c.id === channelId)
    if (channel) {
      channel.unread_count = 0
      channel.mention_count = 0
    }
    mentionCounts.value[channelId] = 0
  }

  // Typing indicators
  function setUserTyping(channelId: string, userId: string, isTyping: boolean) {
    if (!typingUsers.value[channelId]) {
      typingUsers.value[channelId] = []
    }
    const arr = typingUsers.value[channelId]
    const idx = arr.indexOf(userId)
    if (isTyping && idx === -1) {
      arr.push(userId)
    } else if (!isTyping && idx !== -1) {
      arr.splice(idx, 1)
    }
  }

  function getTypingUsers(channelId: string): string[] {
    return typingUsers.value[channelId] ?? []
  }

  // --- Category real-time handlers ---
  function addCategory(cat: ChannelCategory) {
    if (!categories.value.find(c => c.id === cat.id)) {
      categories.value.push(cat)
      categories.value.sort((a, b) => a.position - b.position)
    }
  }

  function updateCategory(cat: ChannelCategory) {
    const idx = categories.value.findIndex(c => c.id === cat.id)
    if (idx !== -1) {
      categories.value[idx] = cat
      categories.value.sort((a, b) => a.position - b.position)
    }
  }

  function removeCategory(categoryId: string) {
    const idx = categories.value.findIndex(c => c.id === categoryId)
    if (idx !== -1) categories.value.splice(idx, 1)
    // Каналы этой категории становятся uncategorized
    for (const ch of channels.value) {
      if (ch.category_id === categoryId) ch.category_id = null
    }
  }

  function $reset() {
    channels.value = []
    categories.value = []
    currentChannelId.value = null
    loading.value = false
    error.value = null
    typingUsers.value = {}
    notificationLevels.value = {}
    mentionCounts.value = {}
  }

  return {
    channels,
    categories,
    currentChannelId,
    currentChannel,
    textChannels,
    voiceChannels,
    loading,
    error,
    typingUsers,
    notificationLevels,
    mentionCounts,
    getMentionCount,
    incrementMention,
    fetchChannels,
    fetchCategories,
    fetchProjectChannels,
    createChannel,
    createCategory,
    renameCategory,
    toggleCategoryPrivacy,
    deleteCategory,
    deleteChannel,
    updateChannelSettings,
    updateNotifications,
    getNotificationLevel,
    markAsRead,
    setCurrentChannel,
    addChannel,
    updateChannel,
    removeChannel,
    incrementUnread,
    clearUnread,
    setUserTyping,
    getTypingUsers,
    addCategory,
    updateCategory,
    removeCategory,
    $reset,
  }
})
