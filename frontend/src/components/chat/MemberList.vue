<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { channelsApi } from '@/api'
import { useWorkspaceStore, useChannelsStore } from '@/stores'
import type { ChannelMemberInfo, WorkspaceMember, WorkspaceRole } from '@/types'

const workspaceStore = useWorkspaceStore()
const channelsStore = useChannelsStore()

const channelMembers = ref<ChannelMemberInfo[]>([])
const loading = ref(false)

const channel = computed(() => channelsStore.currentChannel)
const wsId = computed(() => workspaceStore.currentWorkspaceId)

const workspaceMembers = computed<WorkspaceMember[]>(() =>
  wsId.value ? (workspaceStore.membersMap[wsId.value] ?? []) : []
)

type EnrichedMember = {
  user_id: string
  display_name: string
  avatar_url: string | null
  system_role: 'owner' | 'admin' | 'member'
  custom_roles: WorkspaceRole[]
  presence: string
}

const enrichedMembers = computed<EnrichedMember[]>(() => {
  if (!channel.value) return []

  const toEnriched = (wm: WorkspaceMember): EnrichedMember => ({
    user_id: wm.user_id,
    display_name: wm.nickname ?? wm.display_name,
    avatar_url: wm.avatar_url,
    system_role: wm.role,
    custom_roles: wm.custom_roles ?? [],
    presence: workspaceStore.getPresence(wm.user_id),
  })

  if (channel.value.is_private) {
    return channelMembers.value.map((cm) => {
      const wm = workspaceMembers.value.find(m => m.user_id === cm.user_id)
      if (wm) return toEnriched(wm)
      return {
        user_id: cm.user_id,
        display_name: cm.display_name,
        avatar_url: cm.avatar_url,
        system_role: 'member' as const,
        custom_roles: [],
        presence: workspaceStore.getPresence(cm.user_id),
      }
    })
  }

  return workspaceMembers.value.map(toEnriched)
})

// Группы: кастомные роли (по position DESC) + системные (owner/admin) + "Участники"
type Group = {
  id: string
  label: string
  color: string | null
  members: EnrichedMember[]
}

const SYSTEM_ROLE_ORDER = { owner: 0, admin: 1, member: 2 }

const groups = computed<Group[]>(() => {
  // Собираем уникальные кастомные роли из всех участников, сортируем по position DESC
  const roleMap = new Map<string, WorkspaceRole>()
  for (const m of enrichedMembers.value) {
    for (const r of m.custom_roles) {
      if (!roleMap.has(r.id)) roleMap.set(r.id, r)
    }
  }
  const customRoles = [...roleMap.values()].sort((a, b) => b.position - a.position)

  const result: Group[] = []
  const ungrouped = new Set(enrichedMembers.value.map(m => m.user_id))

  // Группы по кастомным ролям
  for (const role of customRoles) {
    const members = enrichedMembers.value
      .filter(m => m.custom_roles.some(r => r.id === role.id))
      .sort((a, b) => SYSTEM_ROLE_ORDER[a.system_role] - SYSTEM_ROLE_ORDER[b.system_role] || a.display_name.localeCompare(b.display_name))

    if (members.length === 0) continue

    result.push({ id: role.id, label: role.name, color: role.color, members })
    members.forEach(m => ungrouped.delete(m.user_id))
  }

  // Оставшиеся — по системным ролям
  const remaining = enrichedMembers.value
    .filter(m => ungrouped.has(m.user_id))
    .sort((a, b) => SYSTEM_ROLE_ORDER[a.system_role] - SYSTEM_ROLE_ORDER[b.system_role] || a.display_name.localeCompare(b.display_name))

  const ownerAdmins = remaining.filter(m => m.system_role === 'owner' || m.system_role === 'admin')
  const regularMembers = remaining.filter(m => m.system_role === 'member')

  if (ownerAdmins.length > 0) {
    result.push({ id: '__admins__', label: 'Администрация', color: null, members: ownerAdmins })
  }
  if (regularMembers.length > 0) {
    result.push({ id: '__members__', label: 'Участники', color: null, members: regularMembers })
  }

  return result
})

// Разбиваем каждую группу на online/offline
type SplitGroup = Group & { online: EnrichedMember[]; offline: EnrichedMember[] }

const splitGroups = computed<SplitGroup[]>(() =>
  groups.value.map(g => ({
    ...g,
    online: g.members.filter(m => m.presence === 'online'),
    offline: g.members.filter(m => m.presence !== 'online'),
  }))
)

const onlineCount = computed(() => enrichedMembers.value.filter(m => m.presence === 'online').length)
const totalCount = computed(() => enrichedMembers.value.length)

async function loadPrivateMembers(channelId: string) {
  loading.value = true
  try {
    channelMembers.value = await channelsApi.getMembers(channelId)
  } finally {
    loading.value = false
  }
}

watch(
  channel,
  (ch) => {
    if (ch?.is_private) {
      loadPrivateMembers(ch.id)
    } else {
      channelMembers.value = []
    }
  },
  { immediate: true }
)

function getInitials(name: string) {
  return name.split(' ').map(p => p[0]).join('').slice(0, 2).toUpperCase()
}
</script>

<template>
  <aside class="w-60 shrink-0 flex flex-col bg-dark-900 border-l border-dark-800 overflow-hidden">
    <!-- Header -->
    <div class="h-14 px-4 flex items-center gap-2 border-b border-dark-800 shrink-0">
      <span class="text-sm font-semibold text-dark-200">Участники</span>
      <span class="text-xs text-dark-500 ml-auto">{{ onlineCount }}/{{ totalCount }}</span>
    </div>

    <!-- Spinner -->
    <div v-if="loading" class="flex-1 flex items-center justify-center">
      <div class="w-5 h-5 border-2 border-atlas-500 border-t-transparent rounded-full animate-spin" />
    </div>

    <div v-else class="flex-1 overflow-y-auto py-2 scrollbar-thin">
      <template v-for="group in splitGroups" :key="group.id">
        <template v-if="group.online.length || group.offline.length">
          <!-- Group header -->
          <div class="px-3 pt-4 pb-1 flex items-center gap-1.5">
            <span
              v-if="group.color"
              class="w-2 h-2 rounded-full shrink-0"
              :style="{ backgroundColor: group.color }"
            />
            <span
              class="text-xs font-semibold uppercase tracking-wide truncate"
              :style="group.color ? { color: group.color } : {}"
              :class="!group.color ? 'text-dark-500' : ''"
            >
              {{ group.label }} — {{ group.online.length + group.offline.length }}
            </span>
          </div>

          <!-- Online members -->
          <div
            v-for="member in group.online"
            :key="member.user_id"
            class="flex items-center gap-2.5 px-3 py-1.5 mx-1 rounded-md hover:bg-dark-800 cursor-pointer group transition-colors"
          >
            <div class="relative shrink-0">
              <img
                v-if="member.avatar_url"
                :src="member.avatar_url"
                :alt="member.display_name"
                class="w-8 h-8 rounded-full object-cover"
              />
              <div
                v-else
                class="w-8 h-8 rounded-full bg-atlas-600 flex items-center justify-center text-xs font-semibold text-white"
              >
                {{ getInitials(member.display_name) }}
              </div>
              <span class="absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full bg-green-500 border-2 border-dark-900" />
            </div>
            <div class="min-w-0">
              <div class="text-sm text-dark-200 truncate group-hover:text-white transition-colors">
                {{ member.display_name }}
              </div>
              <div v-if="member.system_role !== 'member'" class="text-xs text-dark-500 truncate">
                {{ member.system_role === 'owner' ? 'Владелец' : 'Администратор' }}
              </div>
            </div>
          </div>

          <!-- Offline members -->
          <div
            v-for="member in group.offline"
            :key="member.user_id"
            class="flex items-center gap-2.5 px-3 py-1.5 mx-1 rounded-md hover:bg-dark-800 cursor-pointer group transition-colors"
          >
            <div class="relative shrink-0">
              <img
                v-if="member.avatar_url"
                :src="member.avatar_url"
                :alt="member.display_name"
                class="w-8 h-8 rounded-full object-cover opacity-50"
              />
              <div
                v-else
                class="w-8 h-8 rounded-full bg-dark-700 flex items-center justify-center text-xs font-semibold text-dark-400"
              >
                {{ getInitials(member.display_name) }}
              </div>
              <span class="absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full bg-dark-600 border-2 border-dark-900" />
            </div>
            <div class="min-w-0">
              <div class="text-sm text-dark-500 truncate group-hover:text-dark-300 transition-colors">
                {{ member.display_name }}
              </div>
              <div v-if="member.system_role !== 'member'" class="text-xs text-dark-600 truncate">
                {{ member.system_role === 'owner' ? 'Владелец' : 'Администратор' }}
              </div>
            </div>
          </div>
        </template>
      </template>

      <div v-if="!splitGroups.length" class="px-4 py-8 text-center text-sm text-dark-500">
        Нет участников
      </div>
    </div>
  </aside>
</template>
