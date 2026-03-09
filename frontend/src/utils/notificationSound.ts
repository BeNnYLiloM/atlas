// Генерирует звук уведомления через Web Audio API без внешних файлов
// Два варианта: 'message' — тихий пинг, 'mention' — более заметный двойной пинг

type SoundType = 'message' | 'mention'

let audioCtx: AudioContext | null = null

function getAudioContext(): AudioContext | null {
  if (!audioCtx) {
    try {
      audioCtx = new AudioContext()
    } catch {
      return null
    }
  }
  return audioCtx
}

// Разблокируем AudioContext при первом взаимодействии пользователя.
// Браузеры блокируют воспроизведение звука до первого gesture.
export function unlockAudioContext() {
  const ctx = getAudioContext()
  if (ctx && ctx.state === 'suspended') {
    ctx.resume()
  }
}

function playTone(
  ctx: AudioContext,
  frequency: number,
  startTime: number,
  duration: number,
  volume: number,
) {
  const oscillator = ctx.createOscillator()
  const gainNode = ctx.createGain()

  oscillator.connect(gainNode)
  gainNode.connect(ctx.destination)

  oscillator.type = 'sine'
  oscillator.frequency.setValueAtTime(frequency, startTime)
  // Плавное нарастание и затухание
  gainNode.gain.setValueAtTime(0, startTime)
  gainNode.gain.linearRampToValueAtTime(volume, startTime + 0.01)
  gainNode.gain.exponentialRampToValueAtTime(0.001, startTime + duration)

  oscillator.start(startTime)
  oscillator.stop(startTime + duration)
}

export function playNotificationSound(type: SoundType = 'message') {
  const ctx = getAudioContext()
  if (!ctx) return

  const play = () => {
    const now = ctx.currentTime
    if (type === 'mention') {
      playTone(ctx, 880, now, 0.15, 0.3)
      playTone(ctx, 1100, now + 0.18, 0.2, 0.25)
    } else {
      playTone(ctx, 660, now, 0.18, 0.2)
    }
  }

  if (ctx.state === 'suspended') {
    ctx.resume().then(play)
  } else {
    play()
  }
}

// Настройки звука — хранятся в localStorage
const SOUND_KEY = 'atlas_sound_enabled'
const MENTION_SOUND_KEY = 'atlas_mention_sound_enabled'

export function isSoundEnabled(): boolean {
  return localStorage.getItem(SOUND_KEY) !== 'false'
}

export function isMentionSoundEnabled(): boolean {
  return localStorage.getItem(MENTION_SOUND_KEY) !== 'false'
}

export function setSoundEnabled(enabled: boolean) {
  localStorage.setItem(SOUND_KEY, String(enabled))
}

export function setMentionSoundEnabled(enabled: boolean) {
  localStorage.setItem(MENTION_SOUND_KEY, String(enabled))
}
