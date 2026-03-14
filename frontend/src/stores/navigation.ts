import { defineStore } from 'pinia'
import { ref } from 'vue'

export type NavSection = 'channels' | 'dm' | 'project'

export const useNavigationStore = defineStore('navigation', () => {
  const activeSection = ref<NavSection>('channels')
  const activeProjectId = ref<string | null>(null)

  function setSection(section: NavSection, projectId?: string) {
    activeSection.value = section
    activeProjectId.value = projectId ?? null
  }

  function setProject(projectId: string) {
    activeSection.value = 'project'
    activeProjectId.value = projectId
  }

  return { activeSection, activeProjectId, setSection, setProject }
})
