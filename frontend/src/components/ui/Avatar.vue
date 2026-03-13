<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  src?: string | null
  name: string
  size?: 'xs' | 'sm' | 'md' | 'lg'
  status?: 'online' | 'offline' | 'away' | 'dnd' | null
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
  status: null,
})

const sizeClasses = {
  xs: 'w-6 h-6 text-xs',
  sm: 'w-8 h-8 text-sm',
  md: 'w-10 h-10 text-base',
  lg: 'w-14 h-14 text-xl',
}

const statusSizeClasses = {
  xs: 'w-2 h-2',
  sm: 'w-2.5 h-2.5',
  md: 'w-3 h-3',
  lg: 'w-4 h-4',
}

const statusColors = {
  online:  'bg-emerald-500',
  away:    'bg-amber-500',
  dnd:     'bg-red-500',
  offline: 'bg-muted-fill',
}

const initials = computed(() => {
  return props.name
    .split(' ')
    .map(n => n[0])
    .slice(0, 2)
    .join('')
    .toUpperCase()
})

// Генерируем цвет на основе имени
const bgColor = computed(() => {
  const colors = [
    'bg-accent',
    'bg-emerald-600',
    'bg-amber-600',
    'bg-rose-600',
    'bg-violet-600',
    'bg-cyan-600',
  ]
  const hash = props.name.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
  return colors[hash % colors.length]
})
</script>

<template>
  <div class="relative inline-flex">
    <img
      v-if="props.src"
      :src="props.src"
      :alt="props.name"
      :class="[
        'rounded-full object-cover',
        sizeClasses[props.size],
      ]"
    >
    <div
      v-else
      :class="[
        'rounded-full flex items-center justify-center font-semibold text-primary',
        sizeClasses[props.size],
        bgColor,
      ]"
    >
      {{ initials }}
    </div>
    <span
      v-if="props.status"
      :class="[
        'absolute bottom-0 right-0 rounded-full border-2 border-subtle',
        statusSizeClasses[props.size],
        statusColors[props.status],
      ]"
    />
  </div>
</template>

