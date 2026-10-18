<template>
  <button
    :type="type"
    :disabled="disabled || loading"
    :class="buttonClasses"
    @click="handleClick"
  >
    <span v-if="loading" class="inline-block w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin mr-2"></span>
    <slot />
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger'
  size?: 'sm' | 'md' | 'lg'
  type?: 'button' | 'submit' | 'reset'
  disabled?: boolean
  loading?: boolean
  fullWidth?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'md',
  type: 'button',
  disabled: false,
  loading: false,
  fullWidth: false,
})

const emit = defineEmits<{
  click: [event: MouseEvent]
}>()

const buttonClasses = computed(() => {
  const base = 'inline-flex items-center justify-center font-medium rounded-lg transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2'

  const variants = {
    primary: 'bg-[var(--ui-accent)] hover:bg-[var(--ui-accent-hover)] text-white focus:ring-[var(--ui-accent)] disabled:opacity-50 disabled:cursor-not-allowed shadow-sm hover:shadow-md',
    secondary: 'bg-[var(--ui-bg-soft)] hover:bg-[var(--ui-bg-muted)] text-[var(--ui-text-primary)] focus:ring-[var(--ui-accent)] disabled:opacity-50 disabled:cursor-not-allowed',
    outline: 'border-2 border-[var(--ui-accent)] text-[var(--ui-accent)] hover:bg-[var(--ui-accent-soft)] focus:ring-[var(--ui-accent)] disabled:opacity-50 disabled:cursor-not-allowed',
    ghost: 'text-[var(--ui-accent)] hover:bg-[var(--ui-accent-soft)] focus:ring-[var(--ui-accent)] disabled:opacity-50 disabled:cursor-not-allowed',
    danger: 'bg-[var(--ui-danger)] hover:bg-[#C75B5B] text-white focus:ring-[var(--ui-danger)] disabled:opacity-50 disabled:cursor-not-allowed shadow-sm hover:shadow-md',
  }

  const sizes = {
    sm: 'px-3 py-1.5 text-sm',
    md: 'px-4 py-2.5 text-base',
    lg: 'px-6 py-3 text-lg',
  }

  const width = props.fullWidth ? 'w-full' : ''

  return [base, variants[props.variant], sizes[props.size], width].filter(Boolean).join(' ')
})

const handleClick = (event: MouseEvent) => {
  if (!props.disabled && !props.loading) {
    emit('click', event)
  }
}
</script>
