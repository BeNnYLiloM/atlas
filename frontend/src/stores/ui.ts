import { defineStore } from 'pinia'
import { ref } from 'vue'

export type Theme = 'dark' | 'light' | 'high-contrast'
export type AccentColor = 'indigo' | 'violet' | 'blue' | 'teal' | 'green' | 'rose' | 'orange'

export const ACCENT_COLORS: Array<{ id: AccentColor; label: string; hex500: string; hex600: string }> = [
  { id: 'indigo', label: 'Индиго',    hex500: '#6366f1', hex600: '#4f46e5' },
  { id: 'violet', label: 'Фиолетовый', hex500: '#8b5cf6', hex600: '#7c3aed' },
  { id: 'blue',   label: 'Синий',     hex500: '#3b82f6', hex600: '#2563eb' },
  { id: 'teal',   label: 'Бирюзовый', hex500: '#14b8a6', hex600: '#0d9488' },
  { id: 'green',  label: 'Зелёный',   hex500: '#22c55e', hex600: '#16a34a' },
  { id: 'rose',   label: 'Розовый',   hex500: '#f43f5e', hex600: '#e11d48' },
  { id: 'orange', label: 'Оранжевый', hex500: '#f97316', hex600: '#ea580c' },
]

export const useUIStore = defineStore('ui', () => {
  const theme = ref<Theme>((localStorage.getItem('atlas-theme') as Theme) ?? 'dark')
  const accentColor = ref<AccentColor>((localStorage.getItem('atlas-accent') as AccentColor) ?? 'indigo')
  const shortcutsVisible = ref(false)
  const memberListVisible = ref(true)

  function setTheme(newTheme: Theme) {
    theme.value = newTheme
    document.documentElement.setAttribute('data-theme', newTheme)
    localStorage.setItem('atlas-theme', newTheme)
  }

  function setAccentColor(accent: AccentColor) {
    accentColor.value = accent
    document.documentElement.setAttribute('data-accent', accent)
    localStorage.setItem('atlas-accent', accent)
  }

  function initTheme() {
    document.documentElement.setAttribute('data-theme', theme.value)
    document.documentElement.setAttribute('data-accent', accentColor.value)
  }

  function toggleShortcuts() {
    shortcutsVisible.value = !shortcutsVisible.value
  }

  function toggleMemberList() {
    memberListVisible.value = !memberListVisible.value
  }

  return {
    theme,
    accentColor,
    shortcutsVisible,
    memberListVisible,
    setTheme,
    setAccentColor,
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
