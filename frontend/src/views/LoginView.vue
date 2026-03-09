<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores'
import { Button, Input } from '@/components/ui'

const router = useRouter()
const authStore = useAuthStore()

const form = reactive({
  email: '',
  password: '',
})

const errors = reactive({
  email: '',
  password: '',
  general: '',
})

const loading = ref(false)

function validate(): boolean {
  errors.email = ''
  errors.password = ''
  errors.general = ''

  if (!form.email) {
    errors.email = 'Введите email'
    return false
  }
  if (!form.email.includes('@')) {
    errors.email = 'Некорректный email'
    return false
  }
  if (!form.password) {
    errors.password = 'Введите пароль'
    return false
  }
  return true
}

async function onSubmit() {
  if (!validate()) return

  loading.value = true
  try {
    await authStore.login({
      email: form.email,
      password: form.password,
    })
    router.push('/')
  } catch {
    errors.general = 'Неверный email или пароль'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-pattern p-4">
    <!-- Декоративные элементы фона -->
    <div class="absolute inset-0 overflow-hidden pointer-events-none">
      <div class="absolute -top-40 -right-40 w-80 h-80 bg-atlas-600/10 rounded-full blur-3xl" />
      <div class="absolute -bottom-40 -left-40 w-80 h-80 bg-amber-500/10 rounded-full blur-3xl" />
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
          Корпоративный мессенджер
        </p>
      </div>

      <!-- Форма -->
      <div class="card">
        <h2 class="text-xl font-semibold text-white mb-6">
          Вход в систему
        </h2>

        <form
          class="space-y-4"
          @submit.prevent="onSubmit"
        >
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
            Войти
          </Button>
        </form>

        <div class="mt-6 text-center">
          <p class="text-dark-400 text-sm">
            Нет аккаунта?
            <router-link
              to="/register"
              class="text-atlas-400 hover:text-atlas-300 font-medium"
            >
              Зарегистрироваться
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

