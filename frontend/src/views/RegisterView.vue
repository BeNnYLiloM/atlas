<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores'
import { Button, Input } from '@/components/ui'

const router = useRouter()
const authStore = useAuthStore()

const form = reactive({
  displayName: '',
  email: '',
  password: '',
  confirmPassword: '',
})

const errors = reactive({
  displayName: '',
  email: '',
  password: '',
  confirmPassword: '',
  general: '',
})

const loading = ref(false)

function validate(): boolean {
  errors.displayName = ''
  errors.email = ''
  errors.password = ''
  errors.confirmPassword = ''
  errors.general = ''

  let valid = true

  if (!form.displayName || form.displayName.length < 2) {
    errors.displayName = 'Имя должно содержать минимум 2 символа'
    valid = false
  }
  if (!form.email || !form.email.includes('@')) {
    errors.email = 'Введите корректный email'
    valid = false
  }
  if (!form.password || form.password.length < 8) {
    errors.password = 'Пароль должен содержать минимум 8 символов'
    valid = false
  }
  if (form.password !== form.confirmPassword) {
    errors.confirmPassword = 'Пароли не совпадают'
    valid = false
  }

  return valid
}

async function onSubmit() {
  if (!validate()) return

  loading.value = true
  try {
    await authStore.register({
      display_name: form.displayName,
      email: form.email,
      password: form.password,
    })
    router.push('/')
  } catch {
    errors.general = 'Ошибка регистрации. Возможно, email уже используется.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-pattern p-4">
    <!-- Декоративные элементы фона -->
    <div class="absolute inset-0 overflow-hidden pointer-events-none">
      <div class="absolute -top-40 -left-40 w-80 h-80 bg-atlas-600/10 rounded-full blur-3xl" />
      <div class="absolute -bottom-40 -right-40 w-80 h-80 bg-amber-500/10 rounded-full blur-3xl" />
    </div>

    <div class="w-full max-w-md relative">
      <!-- Логотип -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-atlas-500 to-atlas-700 mb-4 shadow-lg shadow-atlas-500/25">
          <svg
            class="w-8 h-8 text-white"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        </div>
        <h1 class="text-3xl font-bold text-gradient">
          Atlas
        </h1>
        <p class="text-dark-400 mt-2">
          Создайте аккаунт
        </p>
      </div>

      <!-- Форма -->
      <div class="card">
        <h2 class="text-xl font-semibold text-white mb-6">
          Регистрация
        </h2>

        <form
          class="space-y-4"
          @submit.prevent="onSubmit"
        >
          <Input
            v-model="form.displayName"
            type="text"
            label="Имя"
            placeholder="Иван Петров"
            :error="errors.displayName"
          />

          <Input
            v-model="form.email"
            type="email"
            label="Email"
            placeholder="you@company.com"
            :error="errors.email"
          />

          <Input
            v-model="form.password"
            type="password"
            label="Пароль"
            placeholder="••••••••"
            :error="errors.password"
          />

          <Input
            v-model="form.confirmPassword"
            type="password"
            label="Подтверждение пароля"
            placeholder="••••••••"
            :error="errors.confirmPassword"
          />

          <p
            v-if="errors.general"
            class="text-sm text-red-400 text-center"
          >
            {{ errors.general }}
          </p>

          <Button
            type="submit"
            :loading="loading"
            class="w-full"
          >
            Создать аккаунт
          </Button>
        </form>

        <div class="mt-6 text-center">
          <p class="text-dark-400 text-sm">
            Уже есть аккаунт?
            <router-link
              to="/login"
              class="text-atlas-400 hover:text-atlas-300 font-medium"
            >
              Войти
            </router-link>
          </p>
        </div>
      </div>

      <!-- Футер -->
      <p class="text-center text-dark-500 text-xs mt-8">
        © 2025 Atlas. Корпоративная платформа для коммуникаций.
      </p>
    </div>
  </div>
</template>

