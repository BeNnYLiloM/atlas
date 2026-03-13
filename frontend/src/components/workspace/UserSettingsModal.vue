<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Avatar, Button, Input } from '@/components/ui'
import AvatarCropper from '@/components/ui/AvatarCropper.vue'
import ThemeSwitcher from '@/components/settings/ThemeSwitcher.vue'
import { useAuthStore } from '@/stores'
import {
  isMentionSoundEnabled,
  isSoundEnabled,
  setMentionSoundEnabled,
  setSoundEnabled,
  playNotificationSound,
} from '@/utils/notificationSound'
import { authApi } from '@/api/auth'
import type { AuthSession } from '@/api/auth'

const maxAvatarSizeBytes = 10 * 1024 * 1024

type SettingsSection = 'profile' | 'appearance' | 'notifications' | 'security'

const sections: Array<{ id: SettingsSection; label: string }> = [
  { id: 'profile', label: 'Учётная запись' },
  { id: 'appearance', label: 'Внешний вид' },
  { id: 'notifications', label: 'Уведомления' },
  { id: 'security', label: 'Безопасность' },
]

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ close: [] }>()

const router = useRouter()
const authStore = useAuthStore()

const activeSection = ref<SettingsSection>('profile')
const displayName = ref('')
const profileError = ref('')
const profileSaved = ref(false)
const savingProfile = ref(false)
const avatarUploading = ref(false)
const avatarError = ref('')
const loggingOutEverywhere = ref(false)
const sessions = ref<AuthSession[]>([])
const sessionsLoading = ref(false)
const revokingSessionId = ref<string | null>(null)
const avatarInputRef = ref<HTMLInputElement | null>(null)
const avatarPreviewUrl = ref<string | null>(null)
const cropFile = ref<File | null>(null)

// Password change
const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const passwordError = ref('')
const passwordSuccess = ref(false)
const changingPassword = ref(false)


// Notifications
const soundEnabled = ref(isSoundEnabled())
const mentionSoundEnabled = ref(isMentionSoundEnabled())
const notifyMode = ref<'all' | 'mentions' | 'nothing'>('all')
const dndEnabled = ref(false)
const dndFrom = ref('22:00')
const dndTo = ref('09:00')

// Delete account
const deleteConfirmPassword = ref('')
const deletingAccount = ref(false)
const deleteError = ref('')
const showDeleteConfirm = ref(false)

const currentUser = computed(() => authStore.user)
const currentAvatarSrc = computed(() => avatarPreviewUrl.value ?? currentUser.value?.avatar_url ?? null)
const canRemoveAvatar = computed(() => !!(currentUser.value?.avatar_url || avatarPreviewUrl.value))
const hasProfileChanges = computed(() =>
  displayName.value.trim() !== (currentUser.value?.display_name ?? '').trim(),
)

watch(
  () => props.open,
  (open) => {
    if (!open) return
    activeSection.value = 'profile'
    displayName.value = authStore.user?.display_name ?? ''
    profileError.value = ''
    profileSaved.value = false
    avatarError.value = ''
    resetPreview()
    soundEnabled.value = isSoundEnabled()
    mentionSoundEnabled.value = isMentionSoundEnabled()
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
    passwordError.value = ''
    passwordSuccess.value = false
    deleteConfirmPassword.value = ''
    deleteError.value = ''
    showDeleteConfirm.value = false
    loadSessions()
  },
  { immediate: true },
)

async function loadSessions() {
  sessionsLoading.value = true
  try {
    sessions.value = await authApi.listSessions()
  } catch {
    sessions.value = []
  } finally {
    sessionsLoading.value = false
  }
}

watch(() => props.open, (isOpen) => {
  document.body.style.overflow = isOpen ? 'hidden' : ''
})

onBeforeUnmount(() => {
  resetPreview()
  document.body.style.overflow = ''
})

async function saveProfile() {
  if (!currentUser.value) return

  const nextDisplayName = displayName.value.trim()
  if (nextDisplayName.length < 2) {
    profileError.value = 'Минимум 2 символа'
    return
  }

  savingProfile.value = true
  profileError.value = ''
  profileSaved.value = false

  try {
    await authStore.updateProfile({ display_name: nextDisplayName })
    profileSaved.value = true
    setTimeout(() => { profileSaved.value = false }, 2500)
  } catch (error) {
    profileError.value = error instanceof Error ? error.message : 'Не удалось сохранить'
  } finally {
    savingProfile.value = false
  }
}

function triggerAvatarPicker() {
  avatarInputRef.value?.click()
}

function onAvatarSelected(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return

  avatarError.value = ''
  clearAvatarInput()

  if (!file.type.startsWith('image/')) {
    avatarError.value = 'Только изображения'
    return
  }
  if (file.size > maxAvatarSizeBytes) {
    avatarError.value = 'Максимум 10 МБ'
    return
  }

  // Открываем кроппер вместо немедленной загрузки
  cropFile.value = file
}

async function onCropConfirmed(blob: Blob) {
  cropFile.value = null

  resetPreview()
  avatarPreviewUrl.value = URL.createObjectURL(blob)
  avatarUploading.value = true
  avatarError.value = ''

  try {
    const croppedFile = new File([blob], 'avatar.webp', { type: 'image/webp' })
    await authStore.uploadAvatar(croppedFile)
  } catch (error) {
    avatarError.value = error instanceof Error ? error.message : 'Не удалось загрузить аватар'
    resetPreview()
  } finally {
    avatarUploading.value = false
  }
}

function onCropCancelled() {
  cropFile.value = null
}

async function removeAvatar() {
  avatarError.value = ''
  resetPreview()
  try {
    await authStore.updateProfile({ avatar_url: '' })
  } catch (error) {
    avatarError.value = error instanceof Error ? error.message : 'Не удалось удалить аватар'
  }
}

function clearAvatarInput() {
  if (avatarInputRef.value) avatarInputRef.value.value = ''
}

function resetPreview() {
  if (avatarPreviewUrl.value) {
    URL.revokeObjectURL(avatarPreviewUrl.value)
    avatarPreviewUrl.value = null
  }
}

function toggleSound() {
  soundEnabled.value = !soundEnabled.value
  setSoundEnabled(soundEnabled.value)
  if (soundEnabled.value) playNotificationSound('message')
}

function toggleMentionSound() {
  mentionSoundEnabled.value = !mentionSoundEnabled.value
  setMentionSoundEnabled(mentionSoundEnabled.value)
  if (mentionSoundEnabled.value) playNotificationSound('mention')
}

function formatUserAgent(ua: string): string {
  if (!ua) return 'Неизвестное устройство'
  // Браузер
  const browsers: [RegExp, string][] = [
    [/Edg\/[\d.]+/, 'Edge'],
    [/OPR\/[\d.]+/, 'Opera'],
    [/Chrome\/([\d.]+)/, 'Chrome'],
    [/Firefox\/([\d.]+)/, 'Firefox'],
    [/Safari\/([\d.]+)/, 'Safari'],
  ]
  // ОС
  const oses: [RegExp, string][] = [
    [/Windows NT 10/, 'Windows 10'],
    [/Windows NT 11/, 'Windows 11'],
    [/Windows NT/, 'Windows'],
    [/Mac OS X/, 'macOS'],
    [/Android/, 'Android'],
    [/iPhone|iPad/, 'iOS'],
    [/Linux/, 'Linux'],
  ]
  const browser = browsers.find(([re]) => re.test(ua))
  const os = oses.find(([re]) => re.test(ua))
  return [browser?.[1], os?.[1]].filter(Boolean).join(' · ') || ua.slice(0, 60)
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString('ru-RU', {
    day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit',
  })
}

async function changePassword() {
  passwordError.value = ''
  if (!currentPassword.value) {
    passwordError.value = 'Введите текущий пароль'
    return
  }
  if (newPassword.value.length < 8) {
    passwordError.value = 'Новый пароль — минимум 8 символов'
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    passwordError.value = 'Пароли не совпадают'
    return
  }

  changingPassword.value = true
  try {
    await authApi.changePassword(currentPassword.value, newPassword.value)
    passwordSuccess.value = true
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
    setTimeout(() => { passwordSuccess.value = false }, 3000)
  } catch (e) {
    passwordError.value = e instanceof Error ? e.message : 'Не удалось изменить пароль'
  } finally {
    changingPassword.value = false
  }
}

async function deleteAccount() {
  deleteError.value = ''
  if (!deleteConfirmPassword.value) {
    deleteError.value = 'Введите пароль для подтверждения'
    return
  }
  deletingAccount.value = true
  try {
    await authApi.deleteAccount(deleteConfirmPassword.value)
    emit('close')
    await router.push('/login')
  } catch (e) {
    deleteError.value = e instanceof Error ? e.message : 'Не удалось удалить аккаунт'
  } finally {
    deletingAccount.value = false
  }
}

async function revokeSession(sessionId: string) {
  revokingSessionId.value = sessionId
  try {
    await authApi.revokeSession(sessionId)
    sessions.value = sessions.value.filter((s) => s.id !== sessionId)
  } catch {
    // Тихо игнорируем — можно добавить toast позже
  } finally {
    revokingSessionId.value = null
  }
}

async function logoutAllSessions() {
  loggingOutEverywhere.value = true
  try {
    await authStore.logoutAll()
    emit('close')
    await router.push('/login')
  } finally {
    loggingOutEverywhere.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-200"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-200"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="props.open"
        class="fixed inset-0 z-50 flex items-center justify-center bg-base/80 backdrop-blur-sm p-4"
        @click.self="emit('close')"
      >
        <div class="flex w-full max-w-3xl h-[600px] card overflow-hidden p-0">
          <!-- Sidebar -->
          <aside class="w-52 shrink-0 flex flex-col bg-surface border-r border-default p-3">
            <!-- User info -->
            <div class="flex items-center gap-2.5 px-2 py-2 mb-3">
              <Avatar
                :name="currentUser?.display_name ?? '?'"
                :src="currentAvatarSrc"
                size="sm"
                status="online"
              />
              <div class="min-w-0">
                <p class="text-sm font-semibold text-primary truncate">
                  {{ currentUser?.display_name }}
                </p>
                <p class="text-xs text-subtle truncate">
                  {{ currentUser?.email }}
                </p>
              </div>
            </div>

            <p class="px-2 mb-1.5 text-[11px] font-semibold uppercase tracking-widest text-subtle">
              Настройки
            </p>

            <nav class="space-y-0.5">
              <button
                v-for="section in sections"
                :key="section.id"
                type="button"
                class="w-full text-left px-2 py-1.5 rounded-md text-sm transition-colors"
                :class="activeSection === section.id
                  ? 'bg-overlay text-primary font-medium'
                  : 'text-tertiary hover:bg-elevated hover:text-primary'"
                @click="activeSection = section.id"
              >
                {{ section.label }}
              </button>
            </nav>

            <div class="mt-auto pt-3 border-t border-default">
              <button
                type="button"
                class="w-full text-left px-2 py-1.5 rounded-md text-sm text-muted hover:bg-elevated hover:text-secondary transition-colors"
                @click="emit('close')"
              >
                Закрыть
              </button>
            </div>
          </aside>

          <!-- Content -->
          <div class="flex-1 flex flex-col min-w-0 bg-base/90">
            <!-- Profile -->
            <div
              v-if="activeSection === 'profile'"
              class="flex-1 overflow-y-auto p-6 space-y-5"
            >
              <h2 class="text-base font-semibold text-primary">
                Учётная запись
              </h2>

              <!-- Avatar -->
              <div class="flex items-center gap-4 p-4 rounded-xl bg-surface border border-default">
                <div
                  class="relative cursor-pointer group shrink-0"
                  @click="triggerAvatarPicker"
                >
                  <Avatar
                    :name="currentUser?.display_name ?? '?'"
                    :src="currentAvatarSrc"
                    size="lg"
                    status="online"
                  />
                  <div class="absolute inset-0 flex items-center justify-center rounded-full bg-black/50 opacity-0 transition-opacity group-hover:opacity-100">
                    <svg
                      class="h-5 w-5 text-white"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"
                      />
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M15 13a3 3 0 11-6 0 3 3 0 016 0z"
                      />
                    </svg>
                  </div>
                  <div
                    v-if="avatarUploading"
                    class="absolute inset-0 flex items-center justify-center rounded-full bg-black/60"
                  >
                    <svg
                      class="h-4 w-4 animate-spin text-white"
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
                        d="M4 12a8 8 0 018-8v8H4z"
                      />
                    </svg>
                  </div>
                </div>

                <div class="flex-1 min-w-0">
                  <p class="text-sm font-medium text-primary mb-2">
                    Фото профиля
                  </p>
                  <p class="text-xs text-subtle mb-3">
                    PNG, JPG, GIF или WebP · до 10 МБ
                  </p>
                  <div class="flex items-center gap-2">
                    <button
                      type="button"
                      :disabled="avatarUploading"
                      class="px-3 py-1.5 text-xs font-medium rounded-lg bg-overlay text-primary hover:bg-muted-fill transition-colors disabled:opacity-50"
                      @click="triggerAvatarPicker"
                    >
                      Загрузить
                    </button>
                    <button
                      v-if="canRemoveAvatar"
                      type="button"
                      :disabled="avatarUploading"
                      class="px-3 py-1.5 text-xs font-medium rounded-lg text-muted hover:text-red-400 hover:bg-red-500/10 transition-colors disabled:opacity-50"
                      @click="removeAvatar"
                    >
                      Удалить
                    </button>
                  </div>
                  <p
                    v-if="avatarError"
                    class="mt-2 text-xs text-red-400"
                  >
                    {{ avatarError }}
                  </p>
                </div>

                <input
                  ref="avatarInputRef"
                  type="file"
                  accept="image/*"
                  class="hidden"
                  @change="onAvatarSelected"
                >
              </div>

              <!-- Display name -->
              <div class="p-4 rounded-xl bg-surface border border-default space-y-4">
                <Input
                  v-model="displayName"
                  label="Отображаемое имя"
                  placeholder="Иван Петров"
                  :error="profileError"
                />

                <div class="flex items-center gap-3 pt-1">
                  <Button
                    :loading="savingProfile"
                    :disabled="!hasProfileChanges"
                    @click="saveProfile"
                  >
                    Сохранить
                  </Button>
                  <Transition
                    enter-active-class="transition-opacity duration-200"
                    enter-from-class="opacity-0"
                    leave-active-class="transition-opacity duration-200"
                    leave-to-class="opacity-0"
                  >
                    <span
                      v-if="profileSaved"
                      class="text-xs text-emerald-400"
                    >Сохранено ✓</span>
                  </Transition>
                </div>
              </div>

              <!-- Email (read-only) -->
              <div class="p-4 rounded-xl bg-surface border border-default">
                <p class="text-xs font-medium text-subtle uppercase tracking-wider mb-1.5">
                  Email
                </p>
                <p class="text-sm text-tertiary">
                  {{ currentUser?.email }}
                </p>
                <p class="mt-1.5 text-xs text-faint">
                  Email нельзя изменить через настройки профиля
                </p>
              </div>

              <!-- Change password -->
              <div class="p-4 rounded-xl bg-surface border border-default space-y-3">
                <p class="text-sm font-medium text-primary">
                  Смена пароля
                </p>
                <Input
                  v-model="currentPassword"
                  type="password"
                  label="Текущий пароль"
                  placeholder="••••••••"
                />
                <Input
                  v-model="newPassword"
                  type="password"
                  label="Новый пароль"
                  placeholder="Минимум 8 символов"
                />
                <Input
                  v-model="confirmPassword"
                  type="password"
                  label="Повторите новый пароль"
                  placeholder="••••••••"
                  :error="passwordError"
                />
                <div class="flex items-center gap-3 pt-1">
                  <button
                    type="button"
                    :disabled="changingPassword"
                    class="px-3 py-1.5 rounded-lg text-sm font-medium bg-overlay text-primary hover:bg-muted-fill transition-colors disabled:opacity-50"
                    @click="changePassword"
                  >
                    {{ changingPassword ? 'Сохранение…' : 'Изменить пароль' }}
                  </button>
                  <Transition
                    enter-active-class="transition-opacity duration-200"
                    enter-from-class="opacity-0"
                    leave-active-class="transition-opacity duration-200"
                    leave-to-class="opacity-0"
                  >
                    <span v-if="passwordSuccess" class="text-xs text-emerald-400">Пароль изменён ✓</span>
                  </Transition>
                </div>
              </div>
            </div>

            <!-- Appearance -->
            <div
              v-else-if="activeSection === 'appearance'"
              class="flex-1 overflow-y-auto p-6 space-y-5"
            >
              <h2 class="text-base font-semibold text-primary">
                Внешний вид
              </h2>
              <div class="p-4 rounded-xl bg-surface border border-default space-y-3">
                <p class="text-sm font-medium text-secondary">
                  Цветовая тема
                </p>
                <ThemeSwitcher />
              </div>
            </div>

            <!-- Notifications -->
            <div
              v-else-if="activeSection === 'notifications'"
              class="flex-1 overflow-y-auto p-6 space-y-5"
            >
              <h2 class="text-base font-semibold text-primary">
                Уведомления
              </h2>

              <!-- Notify mode -->
              <div class="p-4 rounded-xl bg-surface border border-default space-y-3">
                <p class="text-sm font-medium text-primary">
                  Уведомлять о
                </p>
                <div class="space-y-2">
                  <label
                    v-for="opt in ([
                      { value: 'all', label: 'Все сообщения' },
                      { value: 'mentions', label: 'Только упоминания и личные сообщения' },
                      { value: 'nothing', label: 'Ничего' },
                    ] as const)"
                    :key="opt.value"
                    class="flex items-center gap-3 cursor-pointer"
                  >
                    <input
                      v-model="notifyMode"
                      type="radio"
                      :value="opt.value"
                      class="accent-[var(--accent)]"
                    >
                    <span class="text-sm text-secondary">{{ opt.label }}</span>
                  </label>
                </div>
              </div>

              <!-- Sounds -->
              <div class="rounded-xl bg-surface border border-default divide-y divide-[var(--border-subtle)]">
                <div class="flex items-center justify-between gap-4 px-4 py-3">
                  <div>
                    <p class="text-sm font-medium text-primary">
                      Звук новых сообщений
                    </p>
                    <p class="text-xs text-subtle mt-0.5">
                      Тихий сигнал при получении сообщения
                    </p>
                  </div>
                  <button
                    type="button"
                    :aria-checked="soundEnabled"
                    role="switch"
                    class="relative inline-flex h-5 w-9 shrink-0 rounded-full transition-colors duration-200"
                    :class="soundEnabled ? 'bg-accent' : 'bg-muted-fill'"
                    @click="toggleSound"
                  >
                    <span
                      class="inline-block h-3.5 w-3.5 mt-[3px] rounded-full bg-white shadow transition-transform duration-200"
                      :class="soundEnabled ? 'translate-x-[18px]' : 'translate-x-[3px]'"
                    />
                  </button>
                </div>

                <div class="flex items-center justify-between gap-4 px-4 py-3">
                  <div>
                    <p class="text-sm font-medium text-primary">
                      Звук упоминаний
                    </p>
                    <p class="text-xs text-subtle mt-0.5">
                      Отдельный сигнал для @имя и @everyone
                    </p>
                  </div>
                  <button
                    type="button"
                    :aria-checked="mentionSoundEnabled"
                    role="switch"
                    class="relative inline-flex h-5 w-9 shrink-0 rounded-full transition-colors duration-200"
                    :class="mentionSoundEnabled ? 'bg-accent' : 'bg-muted-fill'"
                    @click="toggleMentionSound"
                  >
                    <span
                      class="inline-block h-3.5 w-3.5 mt-[3px] rounded-full bg-white shadow transition-transform duration-200"
                      :class="mentionSoundEnabled ? 'translate-x-[18px]' : 'translate-x-[3px]'"
                    />
                  </button>
                </div>
              </div>

              <!-- Do Not Disturb -->
              <div class="p-4 rounded-xl bg-surface border border-default space-y-3">
                <div class="flex items-center justify-between">
                  <div>
                    <p class="text-sm font-medium text-primary">
                      Режим «Не беспокоить»
                    </p>
                    <p class="text-xs text-subtle mt-0.5">
                      Автоматически включать по расписанию
                    </p>
                  </div>
                  <button
                    type="button"
                    :aria-checked="dndEnabled"
                    role="switch"
                    class="relative inline-flex h-5 w-9 shrink-0 rounded-full transition-colors duration-200"
                    :class="dndEnabled ? 'bg-accent' : 'bg-muted-fill'"
                    @click="dndEnabled = !dndEnabled"
                  >
                    <span
                      class="inline-block h-3.5 w-3.5 mt-[3px] rounded-full bg-white shadow transition-transform duration-200"
                      :class="dndEnabled ? 'translate-x-[18px]' : 'translate-x-[3px]'"
                    />
                  </button>
                </div>
                <div v-if="dndEnabled" class="flex items-center gap-3 pt-1">
                  <div class="flex-1">
                    <p class="text-xs text-subtle mb-1">
                      С
                    </p>
                    <input
                      v-model="dndFrom"
                      type="time"
                      class="input w-full text-sm"
                    >
                  </div>
                  <div class="flex-1">
                    <p class="text-xs text-subtle mb-1">
                      До
                    </p>
                    <input
                      v-model="dndTo"
                      type="time"
                      class="input w-full text-sm"
                    >
                  </div>
                </div>
              </div>
            </div>

            <!-- Security -->
            <div
              v-else-if="activeSection === 'security'"
              class="flex-1 overflow-y-auto p-6 space-y-5"
            >
              <h2 class="text-base font-semibold text-primary">
                Безопасность
              </h2>

              <!-- Sessions list -->
              <div class="rounded-xl bg-surface border border-default overflow-hidden">
                <div class="flex items-center justify-between px-4 py-3 border-b border-default">
                  <p class="text-sm font-medium text-primary">
                    Активные сессии
                  </p>
                  <span class="text-xs text-subtle">
                    {{ sessions.length }} {{ sessions.length === 1 ? 'устройство' : sessions.length < 5 ? 'устройства' : 'устройств' }}
                  </span>
                </div>

                <!-- Loading -->
                <div
                  v-if="sessionsLoading"
                  class="flex items-center justify-center py-8"
                >
                  <div class="w-5 h-5 border-2 border-accent border-t-transparent rounded-full animate-spin" />
                </div>

                <!-- Session rows -->
                <div
                  v-else-if="sessions.length"
                  class="divide-y divide-[var(--border-subtle)]"
                >
                  <div
                    v-for="session in sessions"
                    :key="session.id"
                    class="flex items-center gap-3 px-4 py-3"
                  >
                    <!-- Device icon -->
                    <div class="shrink-0 w-8 h-8 rounded-lg bg-elevated flex items-center justify-center">
                      <svg
                        class="w-4 h-4 text-muted"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
                        />
                      </svg>
                    </div>

                    <div class="flex-1 min-w-0">
                      <div class="flex items-center gap-2">
                        <p class="text-sm font-medium text-primary truncate">
                          {{ formatUserAgent(session.user_agent) }}
                        </p>
                        <span
                          v-if="session.is_current"
                          class="shrink-0 text-[10px] font-semibold px-1.5 py-0.5 rounded bg-accent-dim text-accent uppercase tracking-wide"
                        >
                          текущая
                        </span>
                      </div>
                      <p class="text-xs text-subtle mt-0.5">
                        {{ session.ip_address }} · {{ formatDate(session.last_used_at) }}
                      </p>
                    </div>

                    <!-- Revoke button (only for non-current sessions) -->
                    <button
                      v-if="!session.is_current"
                      type="button"
                      :disabled="revokingSessionId === session.id"
                      class="shrink-0 p-1.5 rounded-md text-muted hover:text-red-400 hover:bg-red-500/10 transition-colors disabled:opacity-40"
                      title="Завершить сессию"
                      @click="revokeSession(session.id)"
                    >
                      <svg
                        v-if="revokingSessionId === session.id"
                        class="w-4 h-4 animate-spin"
                        fill="none"
                        viewBox="0 0 24 24"
                      >
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z" />
                      </svg>
                      <svg
                        v-else
                        class="w-4 h-4"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                      </svg>
                    </button>
                  </div>
                </div>

                <div
                  v-else
                  class="py-8 text-center text-sm text-subtle"
                >
                  Нет активных сессий
                </div>
              </div>

              <!-- Logout all -->
              <div class="p-4 rounded-xl bg-surface border border-default">
                <p class="text-sm font-medium text-primary mb-1">
                  Завершить все сессии
                </p>
                <p class="text-xs text-subtle leading-relaxed mb-3">
                  Выходит со всех устройств кроме текущего. Используй если потерял доступ к устройству или подозреваешь несанкционированный вход.
                </p>
                <button
                  type="button"
                  :disabled="loggingOutEverywhere"
                  class="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm font-medium text-red-400 border border-red-500/30 hover:bg-red-500/10 transition-colors disabled:opacity-50"
                  @click="logoutAllSessions"
                >
                  <svg
                    v-if="loggingOutEverywhere"
                    class="h-4 w-4 animate-spin"
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
                      d="M4 12a8 8 0 018-8v8H4z"
                    />
                  </svg>
                  Выйти на всех устройствах
                </button>
              </div>

              <!-- Delete account -->
              <div class="p-4 rounded-xl bg-surface border border-red-500/20">
                <p class="text-sm font-medium text-red-400 mb-1">
                  Удалить аккаунт
                </p>
                <p class="text-xs text-subtle leading-relaxed mb-3">
                  Это действие необратимо. Все данные, включая сообщения и файлы, будут удалены навсегда.
                </p>
                <div v-if="!showDeleteConfirm">
                  <button
                    type="button"
                    class="px-3 py-1.5 rounded-lg text-sm font-medium text-red-400 border border-red-500/30 hover:bg-red-500/10 transition-colors"
                    @click="showDeleteConfirm = true"
                  >
                    Удалить мой аккаунт
                  </button>
                </div>
                <div v-else class="space-y-3">
                  <Input
                    v-model="deleteConfirmPassword"
                    type="password"
                    label="Введите пароль для подтверждения"
                    placeholder="••••••••"
                    :error="deleteError"
                  />
                  <div class="flex items-center gap-2">
                    <button
                      type="button"
                      :disabled="deletingAccount"
                      class="px-3 py-1.5 rounded-lg text-sm font-medium text-white bg-red-500 hover:bg-red-600 transition-colors disabled:opacity-50"
                      @click="deleteAccount"
                    >
                      {{ deletingAccount ? 'Удаление…' : 'Подтвердить удаление' }}
                    </button>
                    <button
                      type="button"
                      class="px-3 py-1.5 rounded-lg text-sm text-muted hover:text-secondary transition-colors"
                      @click="showDeleteConfirm = false; deleteConfirmPassword = ''; deleteError = ''"
                    >
                      Отмена
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>

    <AvatarCropper
      v-if="cropFile"
      :file="cropFile"
      @crop="onCropConfirmed"
      @cancel="onCropCancelled"
    />
  </Teleport>
</template>
