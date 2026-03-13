<script setup lang="ts">
import { useUIStore, ACCENT_COLORS } from '@/stores/ui'
import type { Theme, AccentColor } from '@/stores/ui'

const uiStore = useUIStore()

interface ThemeOption {
  value: Theme
  label: string
  // Цвета для мини-превью — независимы от текущей темы
  preview: {
    sidebar: string
    sidebarItem: string
    header: string
    message: string
    messageBg: string
    dot: string
  }
}

const themes: ThemeOption[] = [
  {
    value: 'dark',
    label: 'Тёмная',
    preview: {
      sidebar: '#0f1117',
      sidebarItem: '#1c2128',
      header: '#161b22',
      message: '#e6edf3',
      messageBg: '#1c2128',
      dot: '#3fb950',
    },
  },
  {
    value: 'light',
    label: 'Светлая',
    preview: {
      sidebar: '#f6f8fa',
      sidebarItem: '#eaeef2',
      header: '#ffffff',
      message: '#1f2328',
      messageBg: '#eaeef2',
      dot: '#1a7f37',
    },
  },
  {
    value: 'high-contrast',
    label: 'Контраст',
    preview: {
      sidebar: '#000000',
      sidebarItem: '#0d1117',
      header: '#0d1117',
      message: '#ffffff',
      messageBg: '#161b22',
      dot: '#56d364',
    },
  },
]
</script>

<template>
  <div class="space-y-5">
    <!-- Выбор базовой темы -->
    <div>
      <p class="text-xs font-medium text-subtle uppercase tracking-wider mb-3">
        Тема интерфейса
      </p>
      <div class="grid grid-cols-3 gap-3">
        <button
          v-for="theme in themes"
          :key="theme.value"
          type="button"
          class="group relative rounded-xl border-2 overflow-hidden transition-all duration-150 focus-visible:outline-none"
          :class="uiStore.theme === theme.value
            ? 'border-accent shadow-[0_0_0_1px_var(--accent-500-hex)]'
            : 'border-default hover:border-strong'"
          :aria-pressed="uiStore.theme === theme.value"
          :aria-label="`Тема: ${theme.label}`"
          @click="uiStore.setTheme(theme.value)"
        >
          <!-- Мини-превью интерфейса -->
          <div
            class="w-full aspect-[4/3] relative overflow-hidden"
            :style="{ background: theme.preview.header }"
          >
            <!-- Сайдбар -->
            <div
              class="absolute left-0 top-0 bottom-0 w-[38%] flex flex-col gap-[3px] p-[5px]"
              :style="{ background: theme.preview.sidebar }"
            >
              <!-- Иконка воркспейса -->
              <div
                class="w-full h-[12px] rounded-[3px] mb-[3px] flex items-center gap-[3px] px-[3px]"
                :style="{ background: theme.preview.sidebarItem }"
              >
                <div
                  class="w-[5px] h-[5px] rounded-full shrink-0"
                  :style="{ background: theme.preview.dot }"
                />
                <div
                  class="flex-1 h-[2px] rounded-full opacity-50"
                  :style="{ background: theme.preview.message }"
                />
              </div>
              <!-- Каналы -->
              <div
                v-for="i in 3"
                :key="i"
                class="w-full h-[7px] rounded-[2px] opacity-40"
                :style="{ background: i === 2 ? theme.preview.sidebarItem : 'transparent' }"
              >
                <div
                  class="h-full rounded-[2px]"
                  :style="{ background: i === 2 ? theme.preview.message : theme.preview.sidebarItem, opacity: i === 2 ? '0.6' : '1' }"
                />
              </div>
            </div>

            <!-- Основная область -->
            <div class="absolute left-[38%] right-0 top-0 bottom-0 flex flex-col">
              <!-- Хедер канала -->
              <div
                class="h-[14px] border-b flex items-center px-[5px] shrink-0"
                :style="{ background: theme.preview.header, borderColor: theme.preview.sidebarItem }"
              >
                <div
                  class="h-[3px] w-[40%] rounded-full opacity-60"
                  :style="{ background: theme.preview.message }"
                />
              </div>
              <!-- Сообщения -->
              <div class="flex-1 flex flex-col justify-end gap-[3px] p-[4px]">
                <div class="flex items-end gap-[3px]">
                  <div
                    class="w-[8px] h-[8px] rounded-full shrink-0"
                    :style="{ background: theme.preview.sidebarItem }"
                  />
                  <div
                    class="h-[5px] rounded-[2px] flex-1"
                    :style="{ background: theme.preview.messageBg }"
                  />
                </div>
                <div class="flex items-end gap-[3px]">
                  <div
                    class="w-[8px] h-[8px] rounded-full shrink-0"
                    :style="{ background: theme.preview.sidebarItem }"
                  />
                  <div
                    class="h-[5px] rounded-[2px] w-[55%]"
                    :style="{ background: theme.preview.messageBg }"
                  />
                </div>
              </div>
            </div>

            <!-- Галочка выбора -->
            <div
              v-if="uiStore.theme === theme.value"
              class="absolute top-1.5 right-1.5 w-4 h-4 rounded-full flex items-center justify-center"
              :style="{ background: 'var(--accent-500-hex)' }"
            >
              <svg
                class="w-2.5 h-2.5 text-white"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="3"
                  d="M5 13l4 4L19 7"
                />
              </svg>
            </div>
          </div>

          <!-- Подпись -->
          <div
            class="py-1.5 text-center text-xs font-medium border-t"
            :class="uiStore.theme === theme.value ? 'text-primary border-default' : 'text-muted border-subtle'"
          >
            {{ theme.label }}
          </div>
        </button>
      </div>
    </div>

    <!-- Акцентный цвет -->
    <div>
      <p class="text-xs font-medium text-subtle uppercase tracking-wider mb-3">
        Акцентный цвет
      </p>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="accent in ACCENT_COLORS"
          :key="accent.id"
          type="button"
          class="relative w-8 h-8 rounded-full transition-transform duration-150 hover:scale-110 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/50 focus-visible:ring-offset-2 "
          :style="{ background: accent.hex500 }"
          :aria-pressed="uiStore.accentColor === accent.id"
          :aria-label="accent.label"
          :title="accent.label"
          @click="uiStore.setAccentColor(accent.id as AccentColor)"
        >
          <!-- Кольцо выбора -->
          <span
            v-if="uiStore.accentColor === accent.id"
            class="absolute inset-0 rounded-full"
            :style="{ boxShadow: `0 0 0 2px var(--color-bg-primary), 0 0 0 4px ${accent.hex500}` }"
          />
          <!-- Галочка -->
          <span
            v-if="uiStore.accentColor === accent.id"
            class="absolute inset-0 flex items-center justify-center"
          >
            <svg
              class="w-3.5 h-3.5 text-white"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="3"
                d="M5 13l4 4L19 7"
              />
            </svg>
          </span>
        </button>
      </div>
      <p class="mt-2 text-xs text-faint">
        Применяется к кнопкам, ссылкам и активным элементам
      </p>
    </div>
  </div>
</template>
