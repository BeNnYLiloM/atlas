<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import data from '@emoji-mart/data'
import { Picker } from 'emoji-mart'

const emit = defineEmits<{
  select: [emoji: string]
}>()

const containerRef = ref<HTMLDivElement | null>(null)
let picker: InstanceType<typeof Picker> | null = null

onMounted(() => {
  picker = new Picker({
    data,
    locale: 'ru',
    theme: 'dark',
    previewPosition: 'none',
    skinTonePosition: 'none',
    onEmojiSelect: (emoji: { native: string }) => {
      emit('select', emoji.native)
    },
  })
  if (containerRef.value) {
    containerRef.value.appendChild(picker as unknown as Node)
  }
})

onUnmounted(() => {
  if (containerRef.value && picker) {
    try { containerRef.value.removeChild(picker as unknown as Node) } catch { /* */ }
    picker = null
  }
})
</script>

<template>
  <div
    ref="containerRef"
    class="reaction-picker"
  />
</template>

<style scoped>
.reaction-picker :deep(em-emoji-picker) {
  --border-radius: 12px;
  --background-rgb: 17, 17, 28;
  --category-icon-size: 16px;
  width: 320px;
  height: 340px;
  border: none;
}
</style>
