<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

interface TurnstileRenderOptions {
  sitekey: string
  theme?: 'auto' | 'light' | 'dark'
  size?: 'normal' | 'flexible' | 'compact'
  callback?: (token: string) => void
  'expired-callback'?: () => void
  'error-callback'?: (code?: string) => void
  'timeout-callback'?: () => void
}

interface TurnstileAPI {
  render: (container: HTMLElement, options: TurnstileRenderOptions) => string
  reset: (widgetID?: string) => void
  remove: (widgetID: string) => void
}

declare global {
  interface Window {
    turnstile?: TurnstileAPI
  }
}

const props = defineProps<{
  modelValue?: string
  siteKey: string
}>()

const emit = defineEmits<{
  (event: 'update:modelValue', value: string): void
}>()
const { t } = useI18n()

const containerRef = ref<HTMLElement | null>(null)
const widgetID = ref<string>('')
const errorCode = ref('')

let turnstileScriptPromise: Promise<void> | null = null

const loadScript = async () => {
  if (window.turnstile) return
  if (!turnstileScriptPromise) {
    turnstileScriptPromise = new Promise((resolve, reject) => {
      const existing = document.querySelector<HTMLScriptElement>('script[data-turnstile="1"]')
      if (existing) {
        existing.addEventListener('load', () => resolve(), { once: true })
        existing.addEventListener('error', () => reject(new Error('failed to load turnstile script')), { once: true })
        return
      }
      const script = document.createElement('script')
      script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit'
      script.async = true
      script.defer = true
      script.dataset.turnstile = '1'
      script.onload = () => resolve()
      script.onerror = () => { script.remove(); turnstileScriptPromise = null; reject(new Error('failed to load turnstile script')) }
      document.head.appendChild(script)
    })
  }
  await turnstileScriptPromise
}

const renderWidget = async () => {
  if (!props.siteKey || !containerRef.value) {
    emit('update:modelValue', '')
    return
  }

  await loadScript()
  if (!window.turnstile || !containerRef.value) return

  if (widgetID.value) {
    window.turnstile.remove(widgetID.value)
    widgetID.value = ''
  }

  containerRef.value.innerHTML = ''
  errorCode.value = ''
  widgetID.value = window.turnstile.render(containerRef.value, {
    sitekey: props.siteKey,
    theme: 'auto',
    size: 'flexible',
    callback: (token: string) => { errorCode.value = ''; emit('update:modelValue', token || '') },
    'expired-callback': () => { emit('update:modelValue', ''); window.turnstile?.reset(widgetID.value) },
    'timeout-callback': () => { emit('update:modelValue', ''); window.turnstile?.reset(widgetID.value) },
    'error-callback': (code?: string) => { errorCode.value = String(code || 'unknown'); emit('update:modelValue', '') },
  })
}

const reset = () => {
  if (window.turnstile && widgetID.value) {
    window.turnstile.reset(widgetID.value)
  } else {
    void renderWidget().catch(() => { errorCode.value = 'script-load' })
  }
  errorCode.value = ''
  emit('update:modelValue', '')
}

watch(
  () => props.siteKey,
  () => {
    renderWidget().catch(() => {
      errorCode.value = 'script-load'
      emit('update:modelValue', '')
    })
  },
)

onMounted(() => {
  renderWidget().catch(() => {
    errorCode.value = 'script-load'
    emit('update:modelValue', '')
  })
})

onBeforeUnmount(() => {
  if (window.turnstile && widgetID.value) {
    window.turnstile.remove(widgetID.value)
  }
})

defineExpose({
  reset,
})
</script>

<template>
  <div class="min-w-0">
    <div ref="containerRef" class="min-h-[65px] w-full"></div>
    <div v-if="errorCode" class="mt-2 flex items-center justify-between gap-2 text-xs text-destructive" role="alert">
      <span>{{ t('mobileCheckout.verifyFailed', { code: errorCode }) }}</span>
      <button type="button" class="shrink-0 underline" @click="reset">{{ t('mobileCheckout.retry') }}</button>
    </div>
  </div>
</template>
