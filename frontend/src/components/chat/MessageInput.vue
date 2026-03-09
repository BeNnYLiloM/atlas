<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useMessagesStore, useThreadStore, useWorkspaceStore } from '@/stores'
import { useWebSocketStore } from '@/stores/websocket'
import { rolesApi } from '@/api'
import { filesApi, FILE_LIMITS, formatFileSize } from '@/api/files'
import type { UploadedFile } from '@/api/files'
import type { WorkspaceRole } from '@/types'
import FilePreview from './FilePreview.vue'
import EmojiPicker from './EmojiPicker.vue'

const props = withDefaults(
  defineProps<{
    channelId: string
    parentId?: string
    placeholder?: string
    slowmodeSeconds?: number
  }>(),
  {
    placeholder: 'Написать сообщение...',
    slowmodeSeconds: 0,
  }
)

const messagesStore = useMessagesStore()
const threadStore = useThreadStore()
const wsStore = useWebSocketStore()
const workspaceStore = useWorkspaceStore()

// --- @mention ---
const mentionQuery = ref('')           // текст после @
const mentionActive = ref(false)       // показывать dropdown
const mentionIndex = ref(0)            // выбранный элемент
const wsRoles = ref<WorkspaceRole[]>([])
let mentionStartOffset = -1            // позиция символа @

type MentionItem =
  | { kind: 'member'; id: string; name: string; display: string; color?: string }
  | { kind: 'role';   id: string; name: string; display: string; color: string }

const mentionRoles = computed<MentionItem[]>(() => {
  const q = mentionQuery.value.toLowerCase()
  return wsRoles.value
    .filter(r => !r.is_system && r.name.toLowerCase().includes(q))
    .slice(0, 5)
    .map(r => ({ kind: 'role' as const, id: r.id, name: r.name, display: r.name, color: r.color }))
})

const mentionMembers = computed<MentionItem[]>(() => {
  const q = mentionQuery.value.toLowerCase()
  const wsId = workspaceStore.currentWorkspaceId
  const members = wsId ? (workspaceStore.membersMap[wsId] ?? []) : []
  return members
    .filter(m => m.display_name.toLowerCase().includes(q))
    .slice(0, 8)
    .map(m => ({ kind: 'member' as const, id: m.user_id, name: m.display_name, display: m.display_name }))
})

// Плоский список для навигации клавишами (роли + участники)
const mentionItems = computed<MentionItem[]>(() => [
  ...mentionRoles.value,
  ...mentionMembers.value,
])

async function loadWsRoles() {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId || wsRoles.value.length) return
  wsRoles.value = await rolesApi.list(wsId)
}

// Получить текст до курсора в contenteditable
function getTextBeforeCaret(): string {
  const sel = window.getSelection()
  if (!sel || sel.rangeCount === 0) return ''
  const range = sel.getRangeAt(0).cloneRange()
  range.setStart(editorRef.value!, 0)
  return range.toString()
}

function detectMention() {
  const textBefore = getTextBeforeCaret()
  const match = textBefore.match(/@(\w*)$/)
  if (match) {
    mentionQuery.value = match[1]
    mentionActive.value = true
    mentionIndex.value = 0
    mentionStartOffset = textBefore.length - match[0].length
    loadWsRoles()
  } else {
    mentionActive.value = false
    mentionQuery.value = ''
  }
}

function insertMention(item: MentionItem) {
  const el = editorRef.value
  if (!el) return

  // Находим текстовый узел и позицию @ чтобы удалить "@query"
  const sel = window.getSelection()
  if (!sel || sel.rangeCount === 0) return

  // Восстанавливаем selection до позиции @
  // Проходим по DOM и находим узел+offset где начинается @query
  const atLength = 1 + mentionQuery.value.length // длина "@query"
  let caretOffset = mentionStartOffset            // offset @ в plain-text

  let remaining = caretOffset
  let targetNode: Text | null = null
  let targetOffset = 0

  function findNode(node: Node) {
    if (targetNode) return
    if (node.nodeType === Node.TEXT_NODE) {
      const tn = node as Text
      const len = tn.length
      if (remaining <= len) {
        targetNode = tn
        targetOffset = remaining
      } else {
        remaining -= len
      }
    } else {
      node.childNodes.forEach(findNode)
    }
  }
  findNode(el)

  if (!targetNode) return
  const node = targetNode as Text

  // Удаляем "@query" и вставляем span-упоминание
  const range = document.createRange()
  range.setStart(node, targetOffset)
  range.setEnd(node, Math.min(targetOffset + atLength, node.length))
  range.deleteContents()

  // Создаём span-упоминание
  const color = item.kind === 'role' ? item.color : '#5865f2'
  const span = document.createElement('span')
  span.className = 'mention-chip'
  span.contentEditable = 'false'
  span.dataset.mentionId = item.id
  span.dataset.mentionKind = item.kind
  span.dataset.mentionName = item.display
  span.style.cssText = `background:${color}22;color:${color};border:1px solid ${color}55;`
  span.textContent = `@${item.display}`

  range.insertNode(span)

  // Ставим курсор после span + пробел
  const space = document.createTextNode('\u00A0')
  span.after(space)
  const after = document.createRange()
  after.setStartAfter(space)
  after.collapse(true)
  sel.removeAllRanges()
  sel.addRange(after)

  isEmpty.value = false
  mentionActive.value = false
  mentionQuery.value = ''
}

const editorRef = ref<HTMLDivElement | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)
const showEmojiPicker = ref(false)
const pendingFiles = ref<File[]>([])
const uploadError = ref<string | null>(null)
const uploading = ref(false)
const isEmpty = ref(true)

// --- Slowmode ---
const slowmodeRemaining = ref(0)
let slowmodeTimer: ReturnType<typeof setInterval> | null = null

function startSlowmodeCountdown(seconds: number) {
  slowmodeRemaining.value = seconds
  if (slowmodeTimer) clearInterval(slowmodeTimer)
  slowmodeTimer = setInterval(() => {
    slowmodeRemaining.value--
    if (slowmodeRemaining.value <= 0) {
      clearInterval(slowmodeTimer!)
      slowmodeTimer = null
      slowmodeRemaining.value = 0
    }
  }, 1000)
}

const isSlowmodeActive = computed(() => slowmodeRemaining.value > 0 && !props.parentId)

const EMOJI_RE = /(\p{Emoji_Presentation}|\p{Extended_Pictographic})(\u200d(\p{Emoji_Presentation}|\p{Extended_Pictographic})|\uFE0F|\u20E3)*/gu

let typingTimeout: ReturnType<typeof setTimeout> | null = null

// Получить plain-text из contenteditable
function getPlainText(): string {
  const el = editorRef.value
  if (!el) return ''

  function extractText(node: Node): string {
    if (node.nodeType === Node.TEXT_NODE) return node.textContent ?? ''
    if (node instanceof HTMLElement) {
      // mention-chip: возвращаем @name
      if (node.classList.contains('mention-chip')) {
        return `@${node.dataset.mentionName ?? node.textContent?.slice(1) ?? ''}`
      }
      if (node.nodeName === 'BR' || node.nodeName === 'DIV') {
        return '\n' + Array.from(node.childNodes).map(extractText).join('')
      }
    }
    return Array.from(node.childNodes).map(extractText).join('')
  }

  return Array.from(el.childNodes).map(extractText).join('')
}

// Сохранить позицию курсора
function saveCaret(): number {
  const sel = window.getSelection()
  if (!sel || sel.rangeCount === 0) return 0
  const range = sel.getRangeAt(0)
  const pre = range.cloneRange()
  pre.selectNodeContents(editorRef.value!)
  pre.setEnd(range.endContainer, range.endOffset)
  return pre.toString().length
}

// Восстановить позицию курсора по символьному офсету
function restoreCaret(offset: number) {
  const el = editorRef.value
  if (!el) return
  const sel = window.getSelection()
  if (!sel) return

  let remaining = offset
  let found = false

  function walk(node: Node) {
    if (found || !sel) return
    if (node.nodeType === Node.TEXT_NODE) {
      const len = node.textContent?.length ?? 0
      if (remaining <= len) {
        const range = document.createRange()
        range.setStart(node, remaining)
        range.collapse(true)
        sel.removeAllRanges()
        sel.addRange(range)
        found = true
      } else {
        remaining -= len
      }
    } else {
      node.childNodes.forEach(walk)
    }
  }
  walk(el)
  if (!found && sel) {
    // Ставим в конец
    const range = document.createRange()
    range.selectNodeContents(el)
    range.collapse(false)
    sel.removeAllRanges()
    sel.addRange(range)
  }
}

// Обернуть эмодзи в span в текстовых узлах, не трогая mention-chip
function wrapEmojis() {
  const el = editorRef.value
  if (!el) return

  const text = getPlainText()
  isEmpty.value = text.trim() === ''

  // Проверяем есть ли эмодзи только в текстовых узлах (не в чипах)
  EMOJI_RE.lastIndex = 0
  const textNodes: Text[] = []
  function collectTextNodes(node: Node) {
    if (node.nodeType === Node.TEXT_NODE) {
      textNodes.push(node as Text)
    } else if (node instanceof HTMLElement && !node.classList.contains('mention-chip')) {
      node.childNodes.forEach(collectTextNodes)
    }
  }
  collectTextNodes(el)

  let hasEmoji = false
  for (const tn of textNodes) {
    EMOJI_RE.lastIndex = 0
    if (EMOJI_RE.test(tn.textContent ?? '')) { hasEmoji = true; break }
  }
  if (!hasEmoji) return

  const caretPos = saveCaret()

  for (const tn of textNodes) {
    const raw = tn.textContent ?? ''
    EMOJI_RE.lastIndex = 0
    if (!EMOJI_RE.test(raw)) continue
    EMOJI_RE.lastIndex = 0

    const parts: Array<{ text: string; isEmoji: boolean }> = []
    let last = 0
    let m: RegExpExecArray | null
    while ((m = EMOJI_RE.exec(raw)) !== null) {
      if (m.index > last) parts.push({ text: raw.slice(last, m.index), isEmoji: false })
      parts.push({ text: m[0], isEmoji: true })
      last = m.index + m[0].length
    }
    if (last < raw.length) parts.push({ text: raw.slice(last), isEmoji: false })

    const frag = document.createDocumentFragment()
    for (const p of parts) {
      if (p.isEmoji) {
        const s = document.createElement('span')
        s.className = 'emoji'
        s.contentEditable = 'true'
        s.textContent = p.text
        frag.appendChild(s)
      } else {
        frag.appendChild(document.createTextNode(p.text))
      }
    }
    tn.parentNode?.replaceChild(frag, tn)
  }

  restoreCaret(caretPos)
}


function onInput() {
  wrapEmojis()
  handleTyping()
  detectMention()
}

// Скрываем dropdown при потере фокуса (с задержкой чтобы mousedown на кнопке успел сработать)
let blurTimeout: ReturnType<typeof setTimeout> | null = null

function onEditorBlur() {
  blurTimeout = setTimeout(() => {
    mentionActive.value = false
    mentionQuery.value = ''
  }, 150)
}

function onEditorFocus() {
  if (blurTimeout) {
    clearTimeout(blurTimeout)
    blurTimeout = null
  }
  // Ждём установки курсора браузером, затем проверяем наличие @
  setTimeout(detectMention, 0)
}

function onPaste(e: ClipboardEvent) {
  e.preventDefault()
  const text = e.clipboardData?.getData('text/plain') ?? ''
  document.execCommand('insertText', false, text)
}

function clearEditor() {
  if (editorRef.value) {
    editorRef.value.innerHTML = ''
    isEmpty.value = true
  }
}

// --- Typing ---
function sendTypingEvent(typing: boolean) {
  if (!props.parentId) {
    wsStore.send('typing', { channel_id: props.channelId, typing })
  }
}

function handleTyping() {
  sendTypingEvent(true)
  if (typingTimeout) clearTimeout(typingTimeout)
  typingTimeout = setTimeout(() => {
    sendTypingEvent(false)
    typingTimeout = null
  }, 3000)
}

// --- Emoji / media insert ---
function insertEmoji(emoji: string) {
  const el = editorRef.value
  if (!el) return
  el.focus()
  const sel = window.getSelection()
  if (!sel || sel.rangeCount === 0) {
    el.textContent = (el.textContent ?? '') + emoji
    wrapEmojis()
    return
  }
  const range = sel.getRangeAt(0)
  range.deleteContents()
  const textNode = document.createTextNode(emoji)
  range.insertNode(textNode)
  range.setStartAfter(textNode)
  range.collapse(true)
  sel.removeAllRanges()
  sel.addRange(range)
  wrapEmojis()
}

async function onPickerSelect(payload: { type: 'emoji' | 'gif' | 'sticker'; value: string }) {
  showEmojiPicker.value = false
  if (payload.type === 'emoji') {
    insertEmoji(payload.value)
  } else {
    await sendMediaMessage(payload.value)
  }
}

async function sendMediaMessage(url: string) {
  try {
    const message = await messagesStore.sendMessage({
      channel_id: props.channelId,
      content: url,
      parent_id: props.parentId,
    })
    if (props.parentId && message) {
      threadStore.addThreadReply(props.parentId, message)
    }
  } catch {
    // Ошибка обрабатывается в store
  }
}

function toggleEmojiPicker() {
  showEmojiPicker.value = !showEmojiPicker.value
}

// --- Files ---
function onFileSelect(event: Event) {
  const input = event.target as HTMLInputElement
  if (!input.files) return
  for (const file of Array.from(input.files)) {
    if (file.size > FILE_LIMITS.free.maxSizeBytes) {
      uploadError.value = `Файл "${file.name}" превышает лимит ${formatFileSize(FILE_LIMITS.free.maxSizeBytes)}`
      continue
    }
    pendingFiles.value.push(file)
  }
  input.value = ''
}

function removePendingFile(index: number) {
  pendingFiles.value.splice(index, 1)
}

function onDropFiles(event: DragEvent) {
  event.preventDefault()
  for (const file of Array.from(event.dataTransfer?.files ?? [])) {
    if (file.size <= FILE_LIMITS.free.maxSizeBytes) {
      pendingFiles.value.push(file)
    }
  }
}

// --- Send ---
const hasContent = computed(() => !isEmpty.value || pendingFiles.value.length > 0)

async function sendMessage() {
  const text = getPlainText().trim()
  if (!text && pendingFiles.value.length === 0) return

  if (typingTimeout) {
    clearTimeout(typingTimeout)
    typingTimeout = null
  }
  sendTypingEvent(false)

  try {
    uploading.value = true
    uploadError.value = null

    const uploadedFiles: UploadedFile[] = []
    for (const file of pendingFiles.value) {
      const uploaded = await filesApi.upload(file)
      uploadedFiles.push(uploaded)
    }

    const message = await messagesStore.sendMessage({
      channel_id: props.channelId,
      content: text || ' ',
      parent_id: props.parentId,
    })

    if (props.parentId && message) {
      threadStore.addThreadReply(props.parentId, message)
    }

    clearEditor()
    pendingFiles.value = []

    // Запускаем slowmode каунтдаун (только для обычных сообщений)
    if (!props.parentId && props.slowmodeSeconds > 0) {
      startSlowmodeCountdown(props.slowmodeSeconds)
    }
  } catch {
    // Ошибка обрабатывается в store
  } finally {
    uploading.value = false
  }
}

function onKeydown(event: KeyboardEvent) {
  if (mentionActive.value && mentionItems.value.length) {
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      mentionIndex.value = (mentionIndex.value + 1) % mentionItems.value.length
      return
    }
    if (event.key === 'ArrowUp') {
      event.preventDefault()
      mentionIndex.value = (mentionIndex.value - 1 + mentionItems.value.length) % mentionItems.value.length
      return
    }
    if (event.key === 'Enter' || event.key === 'Tab') {
      event.preventDefault()
      insertMention(mentionItems.value[mentionIndex.value])
      return
    }
    if (event.key === 'Escape') {
      mentionActive.value = false
      return
    }
  }

  // Удаление mention-chip целиком по Backspace
  if (event.key === 'Backspace') {
    const sel = window.getSelection()
    if (sel && sel.rangeCount > 0 && sel.isCollapsed) {
      const range = sel.getRangeAt(0)
      const prev = range.startContainer.nodeType === Node.TEXT_NODE && range.startOffset === 0
        ? range.startContainer.previousSibling
        : null
      const chip = prev instanceof HTMLElement && prev.classList.contains('mention-chip') ? prev : null
      if (chip) {
        event.preventDefault()
        chip.remove()
        isEmpty.value = !editorRef.value?.textContent?.trim()
        return
      }
    }
  }

  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    sendMessage()
  }
}

// --- Click outside emoji picker ---
function onDocumentClick(event: MouseEvent) {
  if (showEmojiPicker.value) {
    const target = event.target as HTMLElement
    if (!target.closest('.emoji-picker-wrapper')) {
      showEmojiPicker.value = false
    }
  }
}

onMounted(() => {
  document.addEventListener('click', onDocumentClick, true)
})

onUnmounted(() => {
  document.removeEventListener('click', onDocumentClick, true)
  if (typingTimeout) {
    clearTimeout(typingTimeout)
    sendTypingEvent(false)
  }
  if (slowmodeTimer) clearInterval(slowmodeTimer)
})
</script>

<template>
  <div class="px-4 pb-4">
    <!-- Pending files preview -->
    <div v-if="pendingFiles.length > 0" class="flex gap-2 flex-wrap mb-2 px-1">
      <FilePreview
        v-for="(file, idx) in pendingFiles"
        :key="idx"
        :file="file"
        removable
        @remove="removePendingFile(idx)"
      />
    </div>

    <!-- Upload error -->
    <p v-if="uploadError" class="text-xs text-red-400 mb-1 px-1">{{ uploadError }}</p>

    <!-- Input area with drag-drop + mention dropdown anchor -->
    <div
      class="relative bg-dark-800 rounded-xl border border-dark-700 focus-within:border-atlas-500/50 transition-colors"
      @dragover.prevent
      @drop="onDropFiles"
    >
      <!-- @mention dropdown — абсолютно над инпутом -->
      <div
        v-if="mentionActive && mentionItems.length"
        class="absolute bottom-full left-0 right-0 mb-2 bg-dark-800 border border-dark-600 rounded-xl shadow-xl overflow-hidden z-50"
      >
        <div class="max-h-64 overflow-y-auto py-1">
          <!-- Группа: Роли -->
          <template v-if="mentionRoles.length">
            <div class="px-3 pt-1.5 pb-0.5">
              <span class="text-[10px] font-semibold uppercase tracking-wider text-dark-400">Роли</span>
            </div>
            <button
              v-for="item in mentionRoles"
              :key="'role-' + item.id"
              class="w-full flex items-center gap-2.5 px-3 py-1.5 text-left transition-colors"
              :class="mentionItems.indexOf(item) === mentionIndex ? 'bg-atlas-600/20 text-white' : 'text-dark-200 hover:bg-dark-700'"
              @mousedown.prevent="insertMention(item)"
              @mousemove="mentionIndex = mentionItems.indexOf(item)"
            >
              <span
                class="w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold shrink-0"
                :style="{ backgroundColor: (item.color ?? '#888') + '33', color: item.color ?? '#888' }"
              >
                <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4z"/>
                </svg>
              </span>
              <span class="text-sm truncate font-medium" :style="{ color: item.color }">{{ item.display }}</span>
            </button>
          </template>

          <!-- Разделитель между группами -->
          <div
            v-if="mentionRoles.length && mentionMembers.length"
            class="my-1 mx-3 border-t border-dark-700"
          />

          <!-- Группа: Участники -->
          <template v-if="mentionMembers.length">
            <div class="px-3 pt-1.5 pb-0.5">
              <span class="text-[10px] font-semibold uppercase tracking-wider text-dark-400">Участники</span>
            </div>
            <button
              v-for="item in mentionMembers"
              :key="'member-' + item.id"
              class="w-full flex items-center gap-2.5 px-3 py-1.5 text-left transition-colors"
              :class="mentionItems.indexOf(item) === mentionIndex ? 'bg-atlas-600/20 text-white' : 'text-dark-200 hover:bg-dark-700'"
              @mousedown.prevent="insertMention(item)"
              @mousemove="mentionIndex = mentionItems.indexOf(item)"
            >
              <span class="w-6 h-6 rounded-full bg-atlas-600 flex items-center justify-center text-white text-xs font-semibold shrink-0">
                {{ item.display[0]?.toUpperCase() }}
              </span>
              <span class="text-sm truncate">{{ item.display }}</span>
            </button>
          </template>
        </div>
      </div>

      <!-- Contenteditable editor -->
      <div class="relative">
        <div
          ref="editorRef"
          contenteditable="true"
          role="textbox"
          aria-multiline="true"
          :aria-placeholder="placeholder"
          class="editor w-full px-4 pt-3 pb-2 text-dark-100 text-sm leading-relaxed focus:outline-none min-h-[44px] max-h-[200px] overflow-y-auto break-words"
          @input="onInput"
          @keydown="onKeydown"
          @paste="onPaste"
          @blur="onEditorBlur"
          @focus="onEditorFocus"
        />
        <span
          v-if="isEmpty"
          class="pointer-events-none absolute left-4 top-3 text-dark-500 text-sm select-none"
        >{{ placeholder }}</span>
      </div>

      <!-- Toolbar -->
      <div class="flex items-center gap-1 px-3 py-1.5 border-t border-dark-700/50">
        <input
          ref="fileInputRef"
          type="file"
          multiple
          class="hidden"
          @change="onFileSelect"
        />
        <button
          class="p-2 rounded text-dark-500 hover:text-dark-300 hover:bg-dark-700"
          title="Прикрепить файл (до 10 MB)"
          @click="fileInputRef?.click()"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
          </svg>
        </button>
        <div class="relative emoji-picker-wrapper">
          <button
            class="p-2 rounded text-dark-500 hover:text-dark-300 hover:bg-dark-700"
            :class="{ 'text-atlas-400': showEmojiPicker }"
            title="Эмодзи"
            @click="toggleEmojiPicker"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </button>
          <div
            v-if="showEmojiPicker"
            class="absolute bottom-10 left-0 z-50 shadow-xl rounded-xl overflow-hidden"
          >
            <EmojiPicker @select="onPickerSelect($event)" />
          </div>
        </div>
        <div class="flex-1" />

        <!-- Slowmode countdown -->
        <div
          v-if="isSlowmodeActive"
          class="flex items-center gap-1.5 mr-2 px-2 py-1 bg-amber-500/10 rounded-lg"
        >
          <svg class="w-3.5 h-3.5 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span class="text-xs text-amber-400 font-mono font-medium">{{ slowmodeRemaining }}с</span>
        </div>

        <span v-else class="text-xs text-dark-600 mr-2">Enter — отправить, Shift+Enter — перенос</span>

        <!-- Send button -->
        <button
          class="p-2 rounded-lg transition-colors"
          :class="[
            hasContent && !isSlowmodeActive
              ? 'bg-atlas-600 text-white hover:bg-atlas-500'
              : 'text-dark-500 cursor-not-allowed'
          ]"
          :disabled="!hasContent || messagesStore.sending || uploading || isSlowmodeActive"
          @click="sendMessage"
        >
          <svg
            v-if="messagesStore.sending || uploading"
            class="animate-spin w-5 h-5"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
          <svg
            v-else
            class="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.editor {
  white-space: pre-wrap;
  word-break: break-word;
}

.editor:empty {
  min-height: 44px;
}

.editor :deep(.emoji) {
  font-size: 1.5em;
  vertical-align: middle;
  line-height: 0;
  display: inline-block;
}

.editor :deep(.mention-chip) {
  display: inline-flex;
  align-items: center;
  border-radius: 4px;
  padding: 1px 5px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: default;
  user-select: all;
  vertical-align: baseline;
  line-height: 1.4;
  white-space: nowrap;
}
</style>
