<script setup lang="ts">
import { ref } from 'vue'
import { useWorkspaceStore } from '@/stores'
import { Modal, Input, Button } from '@/components/ui'

const workspaceStore = useWorkspaceStore()

const showDropdown = ref(false)
const showCreateModal = ref(false)
const newWorkspaceName = ref('')
const creating = ref(false)

async function createWorkspace() {
  if (!newWorkspaceName.value.trim()) return

  creating.value = true
  try {
    await workspaceStore.createWorkspace({
      name: newWorkspaceName.value.trim(),
    })
    showCreateModal.value = false
    newWorkspaceName.value = ''
  } finally {
    creating.value = false
  }
}

function selectWorkspace(id: string) {
  workspaceStore.setCurrentWorkspace(id)
  showDropdown.value = false
}
</script>

<template>
  <div class="relative">
    <button
      class="w-full p-4 flex items-center gap-3 hover:bg-elevated transition-colors border-b border-subtle"
      @click="showDropdown = !showDropdown"
    >
      <!-- Workspace avatar -->
      <div class="w-10 h-10 rounded-xl overflow-hidden bg-gradient-to-br from-[var(--accent)] to-[var(--accent-dim)] flex items-center justify-center text-white font-bold shadow-lg shrink-0">
        <img
          v-if="workspaceStore.currentWorkspace?.icon_url"
          :src="workspaceStore.currentWorkspace.icon_url"
          class="w-full h-full object-cover"
          alt=""
        >
        <span v-else>{{ workspaceStore.currentWorkspace?.name?.[0]?.toUpperCase() || 'A' }}</span>
      </div>
      <div class="flex-1 text-left min-w-0">
        <p class="font-semibold text-primary truncate">
          {{ workspaceStore.currentWorkspace?.name || 'Выберите воркспейс' }}
        </p>
        <p class="text-xs text-muted">
          Воркспейс
        </p>
      </div>
      <svg
        class="w-5 h-5 text-muted transition-transform"
        :class="{ 'rotate-180': showDropdown }"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M19 9l-7 7-7-7"
        />
      </svg>
    </button>

    <!-- Dropdown -->
    <Transition
      enter-active-class="transition-all duration-200"
      enter-from-class="opacity-0 -translate-y-2"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition-all duration-150"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 -translate-y-2"
    >
      <div
        v-if="showDropdown"
        class="absolute top-full left-0 right-0 mt-1 mx-2 bg-elevated border border-default rounded-xl shadow-xl z-50 overflow-hidden"
      >
        <div class="max-h-64 overflow-y-auto">
          <button
            v-for="workspace in workspaceStore.workspaces"
            :key="workspace.id"
            class="w-full p-3 flex items-center gap-3 hover:bg-overlay transition-colors"
            :class="{ 'bg-overlay': workspace.id === workspaceStore.currentWorkspaceId }"
            @click="selectWorkspace(workspace.id)"
          >
            <div class="w-8 h-8 rounded-lg overflow-hidden bg-gradient-to-br from-[var(--accent-dim)] to-[var(--accent-dim)] flex items-center justify-center text-white text-sm font-bold shrink-0">
              <img
                v-if="workspace.icon_url"
                :src="workspace.icon_url"
                class="w-full h-full object-cover"
                alt=""
              >
              <span v-else>{{ workspace.name[0].toUpperCase() }}</span>
            </div>
            <span class="text-sm text-primary truncate">{{ workspace.name }}</span>
            <svg
              v-if="workspace.id === workspaceStore.currentWorkspaceId"
              class="w-4 h-4 text-accent ml-auto"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path
                fill-rule="evenodd"
                d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                clip-rule="evenodd"
              />
            </svg>
          </button>
        </div>

        <div class="border-t border-default p-2">
          <button
            class="w-full p-2 flex items-center gap-2 text-tertiary hover:text-primary hover:bg-overlay rounded-lg transition-colors text-sm"
            @click="showCreateModal = true; showDropdown = false"
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
                d="M12 4v16m8-8H4"
              />
            </svg>
            Создать воркспейс
          </button>
        </div>
      </div>
    </Transition>

    <!-- Create workspace modal -->
    <Modal
      :open="showCreateModal"
      title="Создать воркспейс"
      @close="showCreateModal = false"
    >
      <form
        class="space-y-4"
        @submit.prevent="createWorkspace"
      >
        <Input
          v-model="newWorkspaceName"
          label="Название"
          placeholder="Моя компания"
        />
        <div class="flex gap-3 pt-2">
          <Button
            variant="secondary"
            class="flex-1"
            @click="showCreateModal = false"
          >
            Отмена
          </Button>
          <Button
            type="submit"
            :loading="creating"
            class="flex-1"
          >
            Создать
          </Button>
        </div>
      </form>
    </Modal>
  </div>
</template>

