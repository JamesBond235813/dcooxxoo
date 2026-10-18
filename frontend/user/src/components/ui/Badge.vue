<template>
  <span :class="badgeClasses">
    <slot />
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  variant?: 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'
  size?: 'sm' | 'md' | 'lg'
  rounded?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'default',
  size: 'md',
  rounded: false,
})

const badgeClasses = computed(() => {
  const base = 'inline-flex items-center justify-center font-medium border'

  const variants = {
    default: 'bg-[var(--ui-bg-soft)] text-[var(--ui-text-secondary)] border-[var(--ui-border)]',
    primary: 'bg-[var(--ui-accent)] text-white border-[var(--ui-accent)]',
    success: 'bg-[var(--ui-success-soft)] text-[var(--ui-success)] border-[var(--ui-success)]',
    warning: 'bg-[var(--ui-warning-soft)] text-[var(--ui-warning)] border-[var(--ui-warning)]',
    danger: 'bg-[var(--ui-danger-soft)] text-[var(--ui-danger)] border-[var(--ui-danger)]',
    info: 'bg-[var(--ui-info-soft)] text-[var(--ui-info)] border-[var(--ui-info)]',
  }

  const sizes = {
    sm: 'px-2 py-0.5 text-xs',
    md: 'px-3 py-1 text-sm',
    lg: 'px-4 py-1.5 text-base',
  }

  const shape = props.rounded ? 'rounded-full' : 'rounded'

  return [base, variants[props.variant], sizes[props.size], shape].filter(Boolean).join(' ')
})
</script>
