<script setup lang="ts">
/**
 * AvatarCropper — круглый кроппер на нативном canvas.
 * Поддерживает drag (мышь + тач) и zoom (колесо + pinch).
 * Emit: "crop" с Blob обрезанного изображения, "cancel".
 */
import { onMounted, onBeforeUnmount, ref, watch } from 'vue'

interface Props {
  file: File
  outputSize?: number  // размер выходного изображения в пикселях, по умолчанию 512
  shape?: 'circle' | 'square'  // форма маски, по умолчанию circle
}

const props = withDefaults(defineProps<Props>(), {
  outputSize: 512,
  shape: 'circle',
})

const emit = defineEmits<{
  crop: [blob: Blob]
  cancel: []
}>()

// ─── Refs ─────────────────────────────────────────────────────────────────────

const canvasRef = ref<HTMLCanvasElement | null>(null)

// Состояние изображения
const img = new Image()
let imgLoaded = false

// Позиция и масштаб (в canvas-координатах)
let scale = 1
let offsetX = 0
let offsetY = 0

// Размер canvas (квадрат)
const CANVAS_SIZE = 360

// Drag
let dragging = false
let lastX = 0
let lastY = 0

// Pinch
let lastPinchDist = 0

// ─── Загрузка изображения ──────────────────────────────────────────────────────

watch(() => props.file, loadFile, { immediate: true })

function loadFile(file: File) {
  imgLoaded = false
  const url = URL.createObjectURL(file)
  img.onload = () => {
    URL.revokeObjectURL(url)
    imgLoaded = true
    resetTransform()
    draw()
  }
  img.src = url
}

function resetTransform() {
  // Вписываем изображение так чтобы оно полностью покрывало круг
  const minDim = Math.min(img.naturalWidth, img.naturalHeight)
  scale = CANVAS_SIZE / minDim
  offsetX = (CANVAS_SIZE - img.naturalWidth * scale) / 2
  offsetY = (CANVAS_SIZE - img.naturalHeight * scale) / 2
}

// ─── Рендер ───────────────────────────────────────────────────────────────────

function draw() {
  const canvas = canvasRef.value
  if (!canvas || !imgLoaded) return
  const ctx = canvas.getContext('2d')!

  ctx.clearRect(0, 0, CANVAS_SIZE, CANVAS_SIZE)

  // Изображение
  ctx.drawImage(img, offsetX, offsetY, img.naturalWidth * scale, img.naturalHeight * scale)

  // Затемнение за пределами маски через evenodd path
  const pad = 8
  const inset = pad
  const size = CANVAS_SIZE - pad * 2
  const cornerRadius = props.shape === 'square' ? 20 : CANVAS_SIZE / 2 - 1

  ctx.save()
  ctx.fillStyle = 'rgba(0, 0, 0, 0.55)'
  ctx.beginPath()
  ctx.rect(0, 0, CANVAS_SIZE, CANVAS_SIZE)  // внешний прямоугольник
  if (props.shape === 'square') {
    // Скруглённый прямоугольник против часовой — «вычитается»
    const x = inset, y = inset, w = size, h = size, r = cornerRadius
    ctx.moveTo(x + r, y)
    ctx.lineTo(x + w - r, y)
    ctx.arcTo(x + w, y, x + w, y + r, r)
    ctx.lineTo(x + w, y + h - r)
    ctx.arcTo(x + w, y + h, x + w - r, y + h, r)
    ctx.lineTo(x + r, y + h)
    ctx.arcTo(x, y + h, x, y + h - r, r)
    ctx.lineTo(x, y + r)
    ctx.arcTo(x, y, x + r, y, r)
    ctx.closePath()
  } else {
    const cx = CANVAS_SIZE / 2
    const cy = CANVAS_SIZE / 2
    ctx.arc(cx, cy, cornerRadius, 0, Math.PI * 2, true)
  }
  ctx.fill('evenodd')
  ctx.restore()

  // Обводка маски
  ctx.strokeStyle = 'rgba(255,255,255,0.3)'
  ctx.lineWidth = 1.5
  ctx.beginPath()
  if (props.shape === 'square') {
    const x = inset, y = inset, w = size, h = size, r = cornerRadius
    ctx.moveTo(x + r, y)
    ctx.lineTo(x + w - r, y)
    ctx.arcTo(x + w, y, x + w, y + r, r)
    ctx.lineTo(x + w, y + h - r)
    ctx.arcTo(x + w, y + h, x + w - r, y + h, r)
    ctx.lineTo(x + r, y + h)
    ctx.arcTo(x, y + h, x, y + h - r, r)
    ctx.lineTo(x, y + r)
    ctx.arcTo(x, y, x + r, y, r)
    ctx.closePath()
  } else {
    const cx = CANVAS_SIZE / 2
    const cy = CANVAS_SIZE / 2
    ctx.arc(cx, cy, cornerRadius, 0, Math.PI * 2)
  }
  ctx.stroke()
}

// ─── Зажим позиции (чтобы изображение не уходило за круг) ────────────────────

function clampOffset() {
  const radius = CANVAS_SIZE / 2
  const w = img.naturalWidth * scale
  const h = img.naturalHeight * scale

  // Левый/верхний край изображения не должен заходить правее/ниже левого/верхнего края круга
  offsetX = Math.min(offsetX, 0)
  offsetY = Math.min(offsetY, 0)
  // Правый/нижний край изображения не должен заходить левее/выше правого/нижнего края круга
  offsetX = Math.max(offsetX, CANVAS_SIZE - w)
  offsetY = Math.max(offsetY, CANVAS_SIZE - h)

  // Если изображение меньше диаметра — центрируем (не должно происходить при корректном scale)
  if (w < CANVAS_SIZE) offsetX = (CANVAS_SIZE - w) / 2
  if (h < CANVAS_SIZE) offsetY = (CANVAS_SIZE - h) / 2

  void radius
}

// ─── Zoom ─────────────────────────────────────────────────────────────────────

function applyZoom(delta: number, pivotX: number, pivotY: number) {
  const minScale = Math.max(
    CANVAS_SIZE / img.naturalWidth,
    CANVAS_SIZE / img.naturalHeight,
  )
  const maxScale = minScale * 5

  const prevScale = scale
  scale = Math.min(maxScale, Math.max(minScale, scale * delta))

  // Зум относительно точки пивота
  const ratio = scale / prevScale
  offsetX = pivotX - ratio * (pivotX - offsetX)
  offsetY = pivotY - ratio * (pivotY - offsetY)

  clampOffset()
  draw()
}

// ─── Mouse events ─────────────────────────────────────────────────────────────

function onMouseDown(e: MouseEvent) {
  dragging = true
  lastX = e.clientX
  lastY = e.clientY
}

function onMouseMove(e: MouseEvent) {
  if (!dragging) return
  const dx = e.clientX - lastX
  const dy = e.clientY - lastY
  lastX = e.clientX
  lastY = e.clientY
  offsetX += dx
  offsetY += dy
  clampOffset()
  draw()
}

function onMouseUp() {
  dragging = false
}

function onWheel(e: WheelEvent) {
  e.preventDefault()
  const rect = canvasRef.value!.getBoundingClientRect()
  const pivotX = e.clientX - rect.left
  const pivotY = e.clientY - rect.top
  const delta = e.deltaY < 0 ? 1.08 : 0.93
  applyZoom(delta, pivotX, pivotY)
}

// ─── Touch events ─────────────────────────────────────────────────────────────

function getTouchMidpoint(t1: Touch, t2: Touch) {
  return {
    x: (t1.clientX + t2.clientX) / 2,
    y: (t1.clientY + t2.clientY) / 2,
  }
}

function getTouchDist(t1: Touch, t2: Touch) {
  return Math.hypot(t1.clientX - t2.clientX, t1.clientY - t2.clientY)
}

function onTouchStart(e: TouchEvent) {
  if (e.touches.length === 1) {
    dragging = true
    lastX = e.touches[0].clientX
    lastY = e.touches[0].clientY
  } else if (e.touches.length === 2) {
    dragging = false
    lastPinchDist = getTouchDist(e.touches[0], e.touches[1])
  }
}

function onTouchMove(e: TouchEvent) {
  e.preventDefault()
  if (e.touches.length === 1 && dragging) {
    const dx = e.touches[0].clientX - lastX
    const dy = e.touches[0].clientY - lastY
    lastX = e.touches[0].clientX
    lastY = e.touches[0].clientY
    offsetX += dx
    offsetY += dy
    clampOffset()
    draw()
  } else if (e.touches.length === 2) {
    const dist = getTouchDist(e.touches[0], e.touches[1])
    const delta = dist / lastPinchDist
    lastPinchDist = dist
    const rect = canvasRef.value!.getBoundingClientRect()
    const mid = getTouchMidpoint(e.touches[0], e.touches[1])
    applyZoom(delta, mid.x - rect.left, mid.y - rect.top)
  }
}

function onTouchEnd(e: TouchEvent) {
  if (e.touches.length < 1) dragging = false
}

// ─── Keyboard zoom ────────────────────────────────────────────────────────────

function onKeyDown(e: KeyboardEvent) {
  if (e.key === '+' || e.key === '=') applyZoom(1.08, CANVAS_SIZE / 2, CANVAS_SIZE / 2)
  if (e.key === '-') applyZoom(0.93, CANVAS_SIZE / 2, CANVAS_SIZE / 2)
}

// ─── Глобальные слушатели (чтобы drag работал за пределами canvas) ─────────────

onMounted(() => {
  window.addEventListener('mousemove', onMouseMove)
  window.addEventListener('mouseup', onMouseUp)
  window.addEventListener('keydown', onKeyDown)
})

onBeforeUnmount(() => {
  window.removeEventListener('mousemove', onMouseMove)
  window.removeEventListener('mouseup', onMouseUp)
  window.removeEventListener('keydown', onKeyDown)
})

// ─── Экспорт результата ───────────────────────────────────────────────────────

function applyCrop() {
  const output = document.createElement('canvas')
  output.width = props.outputSize
  output.height = props.outputSize
  const ctx = output.getContext('2d')!

  // Масштабируем координаты к outputSize
  // Для square маска занимает (CANVAS_SIZE - pad*2) пикселей из CANVAS_SIZE
  const pad = props.shape === 'square' ? 8 : 0
  const maskSize = CANVAS_SIZE - pad * 2
  const ratio = props.outputSize / maskSize

  // Клип-маска
  ctx.beginPath()
  if (props.shape === 'square') {
    const r = 20 * ratio
    const s = props.outputSize
    ctx.moveTo(r, 0)
    ctx.lineTo(s - r, 0)
    ctx.arcTo(s, 0, s, r, r)
    ctx.lineTo(s, s - r)
    ctx.arcTo(s, s, s - r, s, r)
    ctx.lineTo(r, s)
    ctx.arcTo(0, s, 0, s - r, r)
    ctx.lineTo(0, r)
    ctx.arcTo(0, 0, r, 0, r)
    ctx.closePath()
  } else {
    ctx.arc(props.outputSize / 2, props.outputSize / 2, props.outputSize / 2, 0, Math.PI * 2)
  }
  ctx.clip()

  ctx.drawImage(
    img,
    (offsetX - pad) * ratio,
    (offsetY - pad) * ratio,
    img.naturalWidth * scale * ratio,
    img.naturalHeight * scale * ratio,
  )

  output.toBlob(
    (blob) => {
      if (blob) emit('crop', blob)
    },
    'image/webp',
    0.92,
  )
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 z-[60] flex items-center justify-center bg-base/90 backdrop-blur-sm p-4">
      <div class="flex flex-col items-center gap-5 w-full max-w-sm card p-6">
        <div class="w-full flex items-center justify-between">
          <h3 class="text-base font-semibold text-primary">
            {{ props.shape === 'square' ? 'Выберите область иконки' : 'Выберите область фото' }}
          </h3>
          <button
            type="button"
            class="btn-ghost p-1 rounded-lg"
            aria-label="Отмена"
            @click="emit('cancel')"
          >
            <svg
              class="w-5 h-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>

        <!-- Canvas -->
        <div
          class="relative cursor-grab active:cursor-grabbing select-none touch-none"
          :class="props.shape === 'circle' ? 'rounded-full overflow-hidden' : 'rounded-2xl overflow-hidden'"
          :style="{ width: `${CANVAS_SIZE}px`, height: `${CANVAS_SIZE}px` }"
        >
          <canvas
            ref="canvasRef"
            :width="CANVAS_SIZE"
            :height="CANVAS_SIZE"
            class="block"
            @mousedown="onMouseDown"
            @wheel.passive="false"
            @wheel="onWheel"
            @touchstart.prevent="onTouchStart"
            @touchmove.prevent="onTouchMove"
            @touchend="onTouchEnd"
          />
        </div>

        <p class="text-xs text-subtle text-center">
          Перетащи или прокрути для масштабирования
        </p>

        <div class="flex w-full gap-3">
          <button
            type="button"
            class="flex-1 px-4 py-2.5 text-sm font-medium rounded-lg bg-elevated text-secondary hover:bg-overlay transition-colors"
            @click="emit('cancel')"
          >
            Отмена
          </button>
          <button
            type="button"
            class="flex-1 px-4 py-2.5 text-sm font-medium rounded-lg bg-accent text-white hover:bg-accent-light transition-colors"
            @click="applyCrop"
          >
            Применить
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
