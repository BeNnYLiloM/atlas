import { onMounted, onUnmounted, ref } from 'vue'

const IDLE_TIMEOUT_MS = 15 * 60 * 1000 // 15 минут

// Список событий, которые считаются активностью
const ACTIVITY_EVENTS: (keyof WindowEventMap)[] = [
  'mousemove',
  'mousedown',
  'keydown',
  'touchstart',
  'scroll',
  'wheel',
]

interface UseIdleDetectorOptions {
  onIdle: () => void
  onActive: () => void
  timeoutMs?: number
}

export function useIdleDetector({ onIdle, onActive, timeoutMs = IDLE_TIMEOUT_MS }: UseIdleDetectorOptions) {
  const isIdle = ref(false)
  let timer: ReturnType<typeof setTimeout> | null = null

  function resetTimer() {
    if (timer !== null) clearTimeout(timer)

    // Если был idle — сообщаем о возврате активности
    if (isIdle.value) {
      isIdle.value = false
      onActive()
    }

    timer = setTimeout(() => {
      isIdle.value = true
      onIdle()
    }, timeoutMs)
  }

  function onVisibilityChange() {
    if (!document.hidden) {
      // Вкладка снова в фокусе — считаем активностью
      resetTimer()
    }
  }

  onMounted(() => {
    for (const event of ACTIVITY_EVENTS) {
      window.addEventListener(event, resetTimer, { passive: true })
    }
    document.addEventListener('visibilitychange', onVisibilityChange)
    // Запускаем таймер сразу
    resetTimer()
  })

  onUnmounted(() => {
    if (timer !== null) clearTimeout(timer)
    for (const event of ACTIVITY_EVENTS) {
      window.removeEventListener(event, resetTimer)
    }
    document.removeEventListener('visibilitychange', onVisibilityChange)
  })

  return { isIdle }
}
