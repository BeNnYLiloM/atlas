<script setup lang="ts" generic="T extends string | number">
import { ref, computed } from 'vue'

interface Option {
  value: T
  label: string
}

interface Props {
  modelValue: T
  options: Option[]
  placeholder?: string
  disabled?: boolean
  small?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  small: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: T]
  'open': []
  'close': []
}>()

const open = ref(false)

const selectedLabel = computed(
  () => props.options.find(o => o.value === props.modelValue)?.label ?? props.placeholder ?? ''
)

function onClickOutside(e: MouseEvent) {
  const el = document.getElementById('select-dropdown-' + uid)
  if (el && !el.contains(e.target as Node)) {
    open.value = false
    emit('close')
    document.removeEventListener('mousedown', onClickOutside)
  }
}

const uid = Math.random().toString(36).slice(2)

function toggle() {
  if (props.disabled) return
  if (!open.value) {
    document.addEventListener('mousedown', onClickOutside, { once: false })
    emit('open')
  } else {
    document.removeEventListener('mousedown', onClickOutside)
    emit('close')
  }
  open.value = !open.value
}

function select(value: T) {
  emit('update:modelValue', value)
  document.removeEventListener('mousedown', onClickOutside)
  emit('close')
  open.value = false
}
</script>

<template>
  <div
    :id="'select-dropdown-' + uid"
    class="relative"
  >
    <button
      type="button"
      class="w-full flex items-center justify-between gap-2 border rounded-lg bg-surface text-primary transition-colors focus:outline-none"
      :class="[
        small ? 'px-2 py-1 text-xs' : 'px-3 py-2 text-sm',
        open ? 'border-accent' : 'border-default hover:border-strong',
        disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer',
      ]"
      :disabled="disabled"
      @click="toggle"
    >
      <span :class="!modelValue && placeholder ? 'text-subtle' : ''">{{ selectedLabel }}</span>
      <svg
        class="w-3.5 h-3.5 text-muted shrink-0 transition-transform duration-150"
        :class="open ? 'rotate-180' : ''"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M19 9l-7 7-7-7"
        />
      </svg>
    </button>

    <Transition
      enter-active-class="transition-all duration-150 origin-top"
      enter-from-class="opacity-0 scale-y-95"
      enter-to-class="opacity-100 scale-y-100"
      leave-active-class="transition-all duration-100 origin-top"
      leave-from-class="opacity-100 scale-y-100"
      leave-to-class="opacity-0 scale-y-95"
    >
      <div
        v-if="open"
        class="absolute z-50 left-0 right-0 mt-1 bg-elevated border border-strong rounded-lg shadow-2xl py-1 max-h-60 overflow-y-auto"
      >
        <button
          v-for="opt in options"
          :key="String(opt.value)"
          type="button"
          class="w-full flex items-center justify-between px-3 py-2 transition-colors text-left"
          :class="[
            small ? 'text-xs py-1.5' : 'text-sm',
            modelValue === opt.value
              ? 'text-accent bg-accent-dim'
              : 'text-secondary hover:bg-overlay hover:text-primary',
          ]"
          @click="select(opt.value)"
        >
          <span>{{ opt.label }}</span>
          <svg
            v-if="modelValue === opt.value"
            class="w-3.5 h-3.5 text-accent shrink-0"
            fill="currentColor"
            viewBox="0 0 20 20"
          >
            <path
              fill-rule="evenodd"
              d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
              clip-rule="evenodd"
            />
          </svg>
        </button>
      </div>
    </Transition>
  </div>
</template>
