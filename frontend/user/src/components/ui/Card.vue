<template>
  <div :class="cardClasses">
    <div v-if="$slots.header" class="px-6 py-4 border-b border-[var(--ui-border)]">
      <slot name="header" />
    </div>
    <div :class="bodyClasses">
      <slot />
    </div>
    <div v-if="$slots.footer" class="px-6 py-4 border-t border-[var(--ui-border)] bg-[var(--ui-bg-soft)]">
      <slot name="footer" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  variant?: 'default' | 'elevated' | 'outlined' | 'accent'
  padding?: boolean
  hoverable?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'default',
  padding: true,
  hoverable: false,
})

const cardClasses = computed(() => {
  const base = 'rounded-lg overflow-hidden transition-all duration-200'

  const variants = {
    default: 'bg-[var(--ui-bg-elevated)] border border-[var(--ui-border)]',
    elevated: 'bg-[var(--ui-bg-elevated)] shadow-[var(--ui-shadow-card)]',
    outlined: 'bg-[var(--ui-bg-elevated)] border-2 border-[var(--ui-border-strong)]',
    accent: 'bg-[var(--ui-accent-soft)] border-2 border-[var(--ui-accent)]',
  }

  const hover = props.hoverable
    ? 'cursor-pointer hover:shadow-[var(--ui-shadow-2)] hover:-translate-y-0.5'
    : ''

  return [base, variants[props.variant], hover].filter(Boolean).join(' ')
})

const bodyClasses = computed(() => {
  return props.padding ? 'p-6' : ''
})
</script>
