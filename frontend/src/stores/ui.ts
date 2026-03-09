import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export type Theme = 'dark' | 'light' | 'high-contrast'

export const useUIStore = defineStore('ui', () => {
  const theme = ref<Theme>((localStorage.getItem('atlas-theme') as Theme) ?? 'dark')
  const shortcutsVisible = ref(false)
  const memberListVisible = ref(true)

  function setTheme(newTheme: Theme) {
    theme.value = newTheme
    document.documentElement.setAttribute('data-theme', newTheme)
    localStorage.setItem('atlas-theme', newTheme)
  }

  // Инициализация темы при загрузке
  function initTheme() {
    document.documentElement.setAttribute('data-theme', theme.value)
  }

  watch(theme, (t) => {
    document.documentElement.setAttribute('data-theme', t)
  })

  function toggleShortcuts() {
    shortcutsVisible.value = !shortcutsVisible.value
  }

  function toggleMemberList() {
    memberListVisible.value = !memberListVisible.value
  }

  return {
    theme,
    shortcutsVisible,
    memberListVisible,
    setTheme,
    initTheme,
    toggleShortcuts,
    toggleMemberList,
  }
})

export const SHORTCUTS = [
  { keys: ['Ctrl', 'K'], description: 'Открыть поиск' },
  { keys: ['Ctrl', '/'], description: 'Список горячих клавиш' },
  { keys: ['Esc'], description: 'Закрыть модальное окно / поиск' },
  { keys: ['Enter'], description: 'Отправить сообщение' },
  { keys: ['Shift', 'Enter'], description: 'Перенос строки' },
  { keys: ['Alt', '↑'], description: 'Предыдущий канал' },
  { keys: ['Alt', '↓'], description: 'Следующий канал' },
]
