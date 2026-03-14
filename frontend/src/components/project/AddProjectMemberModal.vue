<script setup lang="ts">
import { ref, computed } from 'vue'
import { useProjectsStore, useWorkspaceStore } from '@/stores'
const props = defineProps<{
  projectId: string
  workspaceId: string
  existingMemberIds: string[]
}>()
const emit = defineEmits<{ (e: 'close'): void }>()

const projectsStore = useProjectsStore()
const workspaceStore = useWorkspaceStore()

const search = ref('')
const adding = ref<string | null>(null)

const workspaceMembers = computed(() => {
  const members = workspaceStore.membersMap[props.workspaceId] ?? []
  return members.filter(m => !props.existingMemberIds.includes(m.user_id))
})

const filtered = computed(() => {
  const q = search.value.toLowerCase()
  if (!q) return workspaceMembers.value
  return workspaceMembers.value.filter(m =>
    m.display_name?.toLowerCase().includes(q) ||
    m.nickname?.toLowerCase().includes(q)
  )
})

async function addMember(userId: string) {
  adding.value = userId
  try {
    await projectsStore.addMember(props.projectId, userId)
    emit('close')
  } finally {
    adding.value = null
  }
}
</script>

<template>
  <div class="fixed inset-0 z-[60] flex items-center justify-center bg-black/60" @click.self="emit('close')">
    <div class="bg-base rounded-xl shadow-xl w-full max-w-sm mx-4">
      <div class="flex items-center justify-between px-5 py-4 border-b border-subtle">
        <h2 class="font-semibold">Добавить участника</h2>
        <button class="text-muted hover:text-primary" @click="emit('close')">✕</button>
      </div>

      <div class="px-5 py-3">
        <input
          v-model="search"
          type="text"
          placeholder="Поиск по имени..."
          class="w-full bg-surface border border-subtle rounded px-3 py-2 text-sm focus:outline-none focus:border-accent"
        />
      </div>

      <div class="px-5 pb-4 max-h-64 overflow-y-auto space-y-1">
        <div
          v-for="member in filtered"
          :key="member.user_id"
          class="flex items-center gap-3 py-2"
        >
          <div class="w-8 h-8 rounded-full bg-accent/20 flex items-center justify-center text-sm font-semibold flex-shrink-0">
            {{ (member.display_name ?? member.nickname ?? '?')[0]?.toUpperCase() }}
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-sm truncate">{{ member.display_name ?? member.nickname }}</p>
            <p class="text-xs text-muted">{{ member.role }}</p>
          </div>
          <button
            :disabled="adding === member.user_id"
            class="px-3 py-1 bg-accent text-white rounded text-xs hover:bg-accent/90 disabled:opacity-50"
            @click="addMember(member.user_id)"
          >
            {{ adding === member.user_id ? '...' : 'Добавить' }}
          </button>
        </div>
        <p v-if="filtered.length === 0" class="text-sm text-muted text-center py-4">
          Нет доступных участников
        </p>
      </div>
    </div>
  </div>
</template>
