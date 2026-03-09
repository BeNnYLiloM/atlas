<script setup lang="ts">
interface Props {
  modelValue: string
  type?: 'text' | 'email' | 'password'
  placeholder?: string
  label?: string
  error?: string
  disabled?: boolean
  id?: string
}

const props = withDefaults(defineProps<Props>(), {
  type: 'text',
  placeholder: '',
  disabled: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'input': [event: Event]
}>()

function onInput(event: Event) {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', target.value)
  emit('input', event)
}
</script>

<template>
  <div class="space-y-1.5">
    <label
      v-if="props.label"
      :for="props.id"
      class="block text-sm font-medium text-dark-300"
    >
      {{ props.label }}
    </label>
    <input
      :id="props.id"
      :type="props.type"
      :value="props.modelValue"
      :placeholder="props.placeholder"
      :disabled="props.disabled"
      :class="[
        'input',
        props.error && 'border-red-500 focus:border-red-500 focus:ring-red-500',
      ]"
      @input="onInput"
    >
    <p
      v-if="props.error"
      class="text-xs text-red-400"
    >
      {{ props.error }}
    </p>
  </div>
</template>

