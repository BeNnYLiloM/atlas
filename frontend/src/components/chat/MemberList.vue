<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { channelsApi } from '@/api'
import { useWorkspaceStore, useChannelsStore, useAuthStore } from '@/stores'
import { useProjectsStore } from '@/stores/projects'
import { useNavigationStore } from '@/stores/navigation'
import { useDMStore } from '@/stores/dm'
import type { ChannelMemberInfo, WorkspaceMember, WorkspaceRole } from '@/types'
import { Avatar } from '@/components/ui'

const workspaceStore = useWorkspaceStore()
const channelsStore = useChannelsStore()
const projectsStore = useProjectsStore()
const navigationStore = useNavigationStore()
const authStore = useAuthStore()
const dmStore = useDMStore()

async function openDM(userId: string) {
  if (userId === authStore.user?.id) return
  await dmStore.openDM(userId)
}

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

  // Если активна секция проекта — показываем только участников проекта
  if (navigationStore.activeSection === 'project' && navigationStore.activeProjectId) {
    return projectsStore.currentMembers.map((pm) => {
      const wm = workspaceMembers.value.find(m => m.user_id === pm.user_id)
      if (wm) return toEnriched(wm)
      return {
        user_id: pm.user_id,
        display_name: pm.display_name,
        avatar_url: pm.avatar_url,
        system_role: 'member' as const,
        custom_roles: [],
        presence: workspaceStore.getPresence(pm.user_id),
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

const PRESENCE_ORDER: Record<string, number> = { online: 0, dnd: 1, away: 2, offline: 3 }

function isActive(presence: string) {
  return presence === 'online' || presence === 'away' || presence === 'dnd'
}


// Разбиваем каждую группу на активных/офлайн, сортируем по presence внутри
type SplitGroup = Group & { active: EnrichedMember[]; offline: EnrichedMember[] }

const splitGroups = computed<SplitGroup[]>(() =>
  groups.value.map(g => {
    const sorted = [...g.members].sort((a, b) =>
      (PRESENCE_ORDER[a.presence] ?? 3) - (PRESENCE_ORDER[b.presence] ?? 3)
      || a.display_name.localeCompare(b.display_name)
    )
    return {
      ...g,
      active:  sorted.filter(m => isActive(m.presence)),
      offline: sorted.filter(m => !isActive(m.presence)),
    }
  })
)

const onlineCount = computed(() => enrichedMembers.value.filter(m => isActive(m.presence)).length)
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

function presenceToAvatarStatus(presence: string): 'online' | 'away' | 'dnd' | 'offline' {
  if (presence === 'online' || presence === 'away' || presence === 'dnd') return presence
  return 'offline'
}
</script>

<template>
  <aside class="w-60 shrink-0 flex flex-col bg-surface border-l border-subtle overflow-hidden">
    <!-- Header -->
    <div class="h-14 px-4 flex items-center gap-2 border-b border-subtle shrink-0">
      <span class="text-sm font-semibold text-secondary">Участники</span>
      <span class="text-xs text-subtle ml-auto">{{ onlineCount }}/{{ totalCount }}</span>
    </div>

    <!-- Spinner -->
    <div
      v-if="loading"
      class="flex-1 flex items-center justify-center"
    >
      <div class="w-5 h-5 border-2 border-accent border-t-transparent rounded-full animate-spin" />
    </div>

    <div
      v-else
      class="flex-1 overflow-y-auto py-2 scrollbar-thin"
    >
      <template
        v-for="group in splitGroups"
        :key="group.id"
      >
        <template v-if="group.active.length || group.offline.length">
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
              :class="!group.color ? 'text-subtle' : ''"
            >
              {{ group.label }} — {{ group.active.length + group.offline.length }}
            </span>
          </div>

          <!-- Active members (online / away / dnd) -->
          <div
            v-for="member in group.active"
            :key="member.user_id"
            class="flex items-center gap-2.5 px-3 py-1.5 mx-1 rounded-md hover:bg-elevated cursor-pointer group transition-colors"
            @click="openDM(member.user_id)"
          >
            <Avatar
              :name="member.display_name"
              :src="member.avatar_url"
              size="sm"
              :status="presenceToAvatarStatus(member.presence)"
            />
            <div class="min-w-0 flex-1">
              <div class="text-sm text-secondary truncate group-hover:text-primary transition-colors">
                {{ member.display_name }}
              </div>
              <div
                v-if="member.system_role !== 'member'"
                class="text-xs text-subtle truncate"
              >
                {{ member.system_role === 'owner' ? 'Владелец' : 'Администратор' }}
              </div>
            </div>
            <button
              v-if="member.user_id !== authStore.user?.id"
              class="opacity-0 group-hover:opacity-100 p-1 rounded text-muted hover:text-primary transition-all shrink-0"
              title="Написать сообщение"
              @click.stop="openDM(member.user_id)"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
            </button>
          </div>

          <!-- Offline members -->
          <div
            v-for="member in group.offline"
            :key="member.user_id"
            class="flex items-center gap-2.5 px-3 py-1.5 mx-1 rounded-md hover:bg-elevated cursor-pointer group transition-colors opacity-50 hover:opacity-100"
            @click="openDM(member.user_id)"
          >
            <Avatar
              :name="member.display_name"
              :src="member.avatar_url"
              size="sm"
              status="offline"
            />
            <div class="min-w-0 flex-1">
              <div class="text-sm text-subtle truncate group-hover:text-tertiary transition-colors">
                {{ member.display_name }}
              </div>
              <div
                v-if="member.system_role !== 'member'"
                class="text-xs text-faint truncate"
              >
                {{ member.system_role === 'owner' ? 'Владелец' : 'Администратор' }}
              </div>
            </div>
            <button
              v-if="member.user_id !== authStore.user?.id"
              class="opacity-0 group-hover:opacity-100 p-1 rounded text-muted hover:text-primary transition-all shrink-0"
              title="Написать сообщение"
              @click.stop="openDM(member.user_id)"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
            </button>
          </div>
        </template>
      </template>

      <div
        v-if="!splitGroups.length"
        class="px-4 py-8 text-center text-sm text-subtle"
      >
        Нет участников
      </div>
    </div>
  </aside>
</template>
