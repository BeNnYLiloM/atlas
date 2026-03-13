import { createApp } from 'vue'
import router from './router'
import App from './App.vue'
import { pinia } from './plugins/pinia'
import { useAuthStore } from '@/stores'
import { useUIStore } from '@/stores/ui'
import './assets/main.css'
import './assets/themes/dark.css'
import './assets/themes/light.css'
import './assets/themes/high-contrast.css'

// Применяем тему и акцент немедленно — до монтирования Vue,
// чтобы не было вспышки нестилизованного контента (FOUC).
;(function applyStoredTheme() {
  const theme = localStorage.getItem('atlas-theme') ?? 'dark'
  const accent = localStorage.getItem('atlas-accent') ?? 'indigo'
  document.documentElement.setAttribute('data-theme', theme)
  document.documentElement.setAttribute('data-accent', accent)
})()

async function bootstrap() {
  const app = createApp(App)

  app.use(pinia)

  // Синхронизируем store с уже применёнными атрибутами
  useUIStore(pinia).initTheme()

  const authStore = useAuthStore(pinia)
  await authStore.initialize()

  app.use(router)
  await router.isReady()
  app.mount('#app')
}

bootstrap()
