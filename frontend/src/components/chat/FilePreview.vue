<script setup lang="ts">
import { computed } from 'vue'
import { formatFileSize, isImageFile } from '@/api/files'
import type { UploadedFile } from '@/api/files'

const props = defineProps<{
  file: UploadedFile | File
  removable?: boolean
}>()

const emit = defineEmits<{
  remove: []
}>()

const isUploaded = computed(() => 'url' in props.file)

const fileName = computed(() =>
  isUploaded.value ? (props.file as UploadedFile).original_name : (props.file as File).name
)

const fileSize = computed(() =>
  formatFileSize(isUploaded.value ? (props.file as UploadedFile).size_bytes : (props.file as File).size)
)

const mimeType = computed(() =>
  isUploaded.value ? (props.file as UploadedFile).mime_type : (props.file as File).type
)

const isImage = computed(() => isImageFile(mimeType.value))

const imageUrl = computed(() => {
  if (!isImage.value) return null
  if (isUploaded.value) return (props.file as UploadedFile).url
  return URL.createObjectURL(props.file as File)
})
</script>

<template>
  <div class="relative inline-flex flex-col items-center group">
    <!-- Image preview -->
    <div
      v-if="isImage && imageUrl"
      class="w-32 h-32 rounded-lg overflow-hidden bg-dark-800 border border-dark-700"
    >
      <img :src="imageUrl" :alt="fileName" class="w-full h-full object-cover" />
    </div>

    <!-- File icon for non-images -->
    <div
      v-else
      class="flex items-center gap-2 px-3 py-2 rounded-lg bg-dark-800 border border-dark-700 max-w-48"
    >
      <svg class="w-5 h-5 text-atlas-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
      </svg>
      <div class="min-w-0">
        <p class="text-xs text-dark-200 truncate">{{ fileName }}</p>
        <p class="text-xs text-dark-500">{{ fileSize }}</p>
      </div>
    </div>

    <!-- Remove button -->
    <button
      v-if="removable"
      class="absolute -top-1.5 -right-1.5 w-5 h-5 rounded-full bg-dark-600 text-dark-300 hover:bg-red-600 hover:text-white flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
      @click="emit('remove')"
    >
      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
      </svg>
    </button>

    <!-- Download link for uploaded files -->
    <a
      v-if="isUploaded"
      :href="(file as UploadedFile).url"
      target="_blank"
      class="text-xs text-atlas-400 hover:text-atlas-300 mt-1"
    >
      Скачать
    </a>
  </div>
</template>
