import { createApp } from 'vue'
import router from './router'
import App from './App.vue'
import { pinia } from './plugins/pinia'
import { useAuthStore } from '@/stores'
import './assets/main.css'
import './assets/themes/dark.css'
import './assets/themes/light.css'
import './assets/themes/high-contrast.css'

async function bootstrap() {
  const app = createApp(App)

  app.use(pinia)

  const authStore = useAuthStore(pinia)
  await authStore.initialize()

  app.use(router)
  await router.isReady()
  app.mount('#app')
}

bootstrap()
