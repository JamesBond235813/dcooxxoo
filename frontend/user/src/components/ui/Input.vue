<template>
  <div class="w-full">
    <label v-if="label" :for="inputId" class="block text-sm font-medium text-[var(--ui-text-primary)] mb-1.5">
      {{ label }}
      <span v-if="required" class="text-[var(--ui-danger)]">*</span>
    </label>
    <div class="relative">
      <input
        :id="inputId"
        :type="type"
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        :required="required"
        :class="inputClasses"
        @input="handleInput"
        @blur="handleBlur"
        @focus="handleFocus"
      />
      <div v-if="error" class="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none">
        <svg class="h-5 w-5 text-[var(--ui-danger)]" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
        </svg>
      </div>
    </div>
    <p v-if="error" class="mt-1.5 text-sm text-[var(--ui-danger)]">{{ error }}</p>
    <p v-else-if="helper" class="mt-1.5 text-sm text-[var(--ui-text-muted)]">{{ helper }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

interface Props {
  modelValue: string | number
  type?: string
  label?: string
  placeholder?: string
  error?: string
  helper?: string
  disabled?: boolean
  required?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  type: 'text',
  disabled: false,
  required: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  blur: []
  focus: []
}>()

const inputId = ref(`input-${Math.random().toString(36).substring(2, 9)}`)

const inputClasses = computed(() => {
  const base = 'block w-full rounded-lg border px-4 py-2.5 text-[var(--ui-text-primary)] placeholder-[var(--ui-text-muted)] transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-offset-0 disabled:opacity-50 disabled:cursor-not-allowed'

  if (props.error) {
    return `${base} border-[var(--ui-danger)] focus:border-[var(--ui-danger)] focus:ring-[var(--ui-danger)]`
  }

  return `${base} border-[var(--ui-border)] bg-[var(--ui-bg-elevated)] focus:border-[var(--ui-accent)] focus:ring-[var(--ui-accent)]`
})

const handleInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', target.value)
}

const handleBlur = () => {
  emit('blur')
}

const handleFocus = () => {
  emit('focus')
}
</script>
