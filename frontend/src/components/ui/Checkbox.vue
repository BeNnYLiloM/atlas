<script setup lang="ts">
interface Props {
  modelValue: boolean
  disabled?: boolean
  label?: string
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'change': [value: boolean]
}>()

function toggle() {
  if (props.disabled) return
  const next = !props.modelValue
  emit('update:modelValue', next)
  emit('change', next)
}
</script>

<template>
  <button
    type="button"
    role="checkbox"
    :aria-checked="modelValue"
    :disabled="disabled"
    class="inline-flex items-center gap-2 group"
    :class="disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'"
    @click="toggle"
  >
    <span
      class="w-4 h-4 rounded flex items-center justify-center shrink-0 border transition-colors"
      :class="modelValue
        ? 'bg-atlas-600 border-atlas-500'
        : 'bg-dark-900 border-dark-600 group-hover:border-dark-400'"
    >
      <svg
        v-if="modelValue"
        class="w-2.5 h-2.5 text-white"
        fill="none"
        stroke="currentColor"
        stroke-width="3"
        viewBox="0 0 12 12"
      >
        <path stroke-linecap="round" stroke-linejoin="round" d="M1.5 6l3 3 6-6" />
      </svg>
    </span>
    <span v-if="label" class="text-sm text-dark-200">{{ label }}</span>
  </button>
</template>
