<template>
  <section class="rounded-2xl border border-primary/20 bg-card p-4 shadow-sm sm:p-5" aria-live="polite">
    <div class="flex items-center gap-2">
      <Smartphone class="h-5 w-5 text-primary" />
      <h2 class="font-bold text-foreground">{{ t('smsConsole.title') }}</h2>
      <span class="ml-auto text-xs text-muted-foreground">{{ t('smsConsole.autoRefresh') }}</span>
    </div>

    <ol class="mt-4 grid grid-cols-3 gap-2 text-xs sm:text-sm">
      <li v-for="(label, index) in stepLabels" :key="label" class="flex items-center gap-2">
        <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full font-bold"
          :class="index < activeStep ? 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400' : index === activeStep ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground'">
          {{ index < activeStep ? '✓' : index + 1 }}
        </span>
        <span :class="index <= activeStep ? 'font-medium text-foreground' : 'text-muted-foreground'">{{ label }}</span>
      </li>
    </ol>

    <div v-if="loading" class="mt-5 text-sm text-muted-foreground">{{ t('smsConsole.loading') }}</div>
    <div v-else-if="error && !activation" class="mt-5 flex flex-wrap items-center gap-3 text-sm text-destructive">
      <AlertCircle class="h-4 w-4" /><span>{{ error }}</span>
      <Button size="sm" variant="outline" @click="loadCurrentActivation">{{ t('smsConsole.retryLoad') }}</Button>
    </div>
    <div v-else class="mt-5 space-y-4">
      <div v-if="!activation" class="flex flex-wrap items-center justify-between gap-3 border-t pt-4">
        <p class="text-sm text-muted-foreground">{{ t('smsConsole.chooseHint') }}</p>
        <Button :disabled="acquiring" @click="handleAcquire">{{ acquiring ? t('smsConsole.choosing') : t('smsConsole.choose') }}</Button>
      </div>

      <div v-if="activation" class="border-t pt-4">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div class="min-w-0">
            <p class="text-xs text-muted-foreground">{{ countryDisplay }} · {{ activation.operator || t('smsConsole.autoOperator') }}</p>
            <div class="mt-1 flex flex-wrap items-center gap-2">
              <strong class="break-all font-mono text-2xl tracking-wide text-foreground select-all">{{ activation.phone || t('smsConsole.allocating') }}</strong>
              <Button v-if="activation.phone" size="sm" variant="outline" class="h-8 gap-1" @click="copyPhone">
                <Check v-if="copiedPhone" class="h-3.5 w-3.5" /><Copy v-else class="h-3.5 w-3.5" />{{ copiedPhone ? t('smsConsole.copied') : t('smsConsole.copyPhone') }}
              </Button>
            </div>
          </div>
          <div v-if="activation.status === 'pending'" class="flex items-center gap-2 text-sm text-amber-700 dark:text-amber-400">
            <Clock class="h-4 w-4" /><span class="font-mono tabular-nums">{{ countdownFormatted }}</span>
          </div>
        </div>
        <div v-if="activation.status === 'pending'" class="mt-3 flex flex-wrap items-center gap-3">
          <p class="min-w-0 flex-1 text-sm text-muted-foreground">{{ receiving ? t('smsConsole.reading') : t('smsConsole.submitNumberHint') }}</p>
          <Button v-if="!receiving" size="sm" :disabled="checking" @click="handleStartReceiving">{{ checking ? t('smsConsole.checking') : t('smsConsole.receive') }}</Button>
          <Button v-if="canRetry" size="sm" variant="outline" :disabled="loading || retrying" @click="handleRetry">
            <RefreshCw class="mr-1 h-3.5 w-3.5" :class="{ 'animate-spin': retrying }" />{{ retrying ? t('smsConsole.changing') : t('smsConsole.changeNumber') }}
          </Button>
        </div>
        <details v-if="activation.status === 'pending'" class="mt-3 text-xs text-muted-foreground">
          <summary class="cursor-pointer">{{ t('smsConsole.numberHelp') }}</summary>
          <p class="mt-2">{{ t('smsConsole.numberHelpBody', { country: countryDisplay }) }}</p>
        </details>
      </div>

      <div v-if="activation?.status === 'received' || activation?.status === 'finished'" class="flex flex-wrap items-center gap-3 rounded-xl bg-emerald-500/10 p-4">
        <div class="min-w-0 flex-1">
          <p class="text-xs font-medium text-emerald-700 dark:text-emerald-400">{{ t('smsConsole.received') }}</p>
          <strong class="break-all font-mono text-3xl tracking-widest text-foreground select-all">{{ activation.code }}</strong>
        </div>
        <Button size="sm" @click="copyCode"><Check v-if="copiedCode" class="mr-1 h-4 w-4" /><Copy v-else class="mr-1 h-4 w-4" />{{ copiedCode ? t('smsConsole.copied') : t('smsConsole.copyCode') }}</Button>
        <details v-if="activation.full_sms" class="w-full text-xs text-muted-foreground"><summary class="cursor-pointer">{{ t('smsConsole.smsText') }}</summary><p class="mt-2 break-all font-mono">{{ activation.full_sms }}</p></details>
      </div>
      <div v-else-if="activation?.status === 'timeout' || activation?.status === 'canceled'" class="flex flex-wrap items-center justify-between gap-3 rounded-xl bg-muted/30 p-4">
        <p class="text-sm">{{ t('smsConsole.expired') }}</p>
        <Button size="sm" :disabled="retrying" @click="handleRetry">{{ retrying ? t('smsConsole.changing') : t('smsConsole.newNumber') }}</Button>
      </div>
      <p v-if="error && activation" class="text-sm text-destructive">{{ error }}</p>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { Smartphone, RefreshCw, Copy, Check, Clock, AlertCircle } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { userOrderAPI, guestOrderAPI } from '@/api/order'
import { t } from '@/api/client'
import { toast } from '@/composables/useToast'

const props = defineProps<{
  order: any
  isGuest?: boolean
  guestEmail?: string
}>()

const emit = defineEmits<{
  (e: 'refresh'): void
}>()

const activation = ref<any>(null)
const loading = ref(false)
const acquiring = ref(false)
const checking = ref(false)
const receiving = ref(false)
const retrying = ref(false)
const error = ref('')
const copiedPhone = ref(false)
const copiedCode = ref(false)
const remainingSeconds = ref(0)

let pollTimer: any = null
let countdownTimer: any = null
let pollingRequest = false
const receiveIntentKey = computed(() => `sms-receive:${props.order?.order_no || ''}`)
const rememberReceiving = (enabled: boolean) => {
  try {
    if (enabled) sessionStorage.setItem(receiveIntentKey.value, '1')
    else sessionStorage.removeItem(receiveIntentKey.value)
  } catch {}
}
const wasReceiving = () => {
  try { return sessionStorage.getItem(receiveIntentKey.value) === '1' } catch { return false }
}

// userApi returns { data: ApiResponse }; the activation record is in data.data.
// Accept a direct activation object as well so this component stays resilient to
// API adapters that already unwrap the standard response envelope.
const readActivation = (response: any) => {
  const payload = response?.data?.data ?? response?.data ?? response
  if (!payload || typeof payload !== 'object' || Array.isArray(payload)) return null
  if (typeof payload.status !== 'string' && typeof payload.phone !== 'string') return null
  return payload
}

const countryDisplay = computed(() => {
  const c = activation.value?.country?.toLowerCase() || ''
  const map: Record<string, string> = {
    vietnam: '越南 (+84)',
    greece: '希腊 (+30)',
    england: '英国 (+44)',
    indonesia: '印度尼西亚 (+62)',
    usa: '美国 (+1)',
    philippines: '菲律宾 (+63)',
    thailand: '泰国 (+66)',
    malaysia: '马来西亚 (+60)',
  }
  return map[c] || c || '海外国际'
})

const countdownFormatted = computed(() => {
  const mins = Math.floor(remainingSeconds.value / 60)
  const secs = remainingSeconds.value % 60
  return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`
})

const stepLabels = computed(() => [t('smsConsole.chooseStep'), t('smsConsole.receiveStep'), t('smsConsole.completeStep')])
const activeStep = computed(() => {
  if (activation.value?.status === 'received' || activation.value?.status === 'finished') return 2
  if (activation.value?.phone) return 1
  return 0
})

const canRetry = computed(() => {
  return activation.value?.status === 'pending'
})

const getApiClient = () => {
  const orderNo = props.order?.order_no || ''
  if (props.isGuest) {
    const params = { email: props.guestEmail || props.order?.guest_email || '' }
    return {
      acquire: () => guestOrderAPI.acquireSmsNumber(orderNo, params),
      check: () => guestOrderAPI.checkSmsStatus(orderNo, params),
      retry: () => guestOrderAPI.retrySmsNumber(orderNo, params),
    }
  }
  return {
    acquire: () => userOrderAPI.acquireSmsNumber(orderNo),
    check: () => userOrderAPI.checkSmsStatus(orderNo),
    retry: () => userOrderAPI.retrySmsNumber(orderNo),
  }
}

const updateCountdown = () => {
  if (!activation.value?.expires_at) {
    remainingSeconds.value = 0
    return
  }
  const expTime = new Date(activation.value.expires_at).getTime()
  const now = Date.now()
  const diff = Math.max(0, Math.floor((expTime - now) / 1000))
  remainingSeconds.value = diff
  if (diff <= 0 && activation.value?.status === 'pending') {
    activation.value.status = 'timeout'
    stopPolling()
  }
}

const startCountdown = () => {
  clearInterval(countdownTimer)
  updateCountdown()
  countdownTimer = setInterval(updateCountdown, 1000)
}

const startPolling = () => {
  clearInterval(pollTimer)
  receiving.value = true
  rememberReceiving(true)
  pollTimer = setInterval(async () => {
    if (pollingRequest) return
    pollingRequest = true
    try {
      const api = getApiClient()
      const res: any = await api.check()
      const current = readActivation(res)
      if (current) {
        activation.value = current
        if (current.status === 'received' || current.status === 'finished') {
          stopPolling()
          toast.success('🎉 OpenAI 验证码接收成功！')
          emit('refresh')
        } else if (current.status === 'timeout' || current.status === 'canceled') {
          stopPolling()
        }
      }
    } catch (e) {
      console.warn('SMS check polling warning:', e)
    } finally {
      pollingRequest = false
    }
  }, 4000)
}

const stopPolling = () => {
  clearInterval(pollTimer)
  pollTimer = null
  receiving.value = false
  rememberReceiving(false)
}

const loadCurrentActivation = async () => {
  if (!props.order?.order_no) return
  loading.value = true
  error.value = ''
  try {
    const api = getApiClient()
    const statusRes: any = await api.check()
    const current = readActivation(statusRes)
    if (current) {
      activation.value = current
      if (activation.value?.status === 'pending') {
        startCountdown()
        if (wasReceiving()) startPolling()
      } else {
        rememberReceiving(false)
      }
    }
  } catch (err: any) {
    error.value = err?.message || '读取接码状态失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

const handleAcquire = async () => {
  if (acquiring.value) return
  rememberReceiving(false)
  acquiring.value = true
  error.value = ''
  try {
    const res: any = await getApiClient().acquire()
    const current = readActivation(res)
    if (current) {
      activation.value = current
      startCountdown()
    } else {
      error.value = t('common.api.requestFailed')
    }
  } catch (err: any) {
    error.value = err?.message || '选号失败，请稍后重试'
  } finally {
    acquiring.value = false
  }
}

const handleStartReceiving = async () => {
  if (checking.value || !activation.value?.phone) return
  checking.value = true
  try {
    const res: any = await getApiClient().check()
    const current = readActivation(res)
    if (current) activation.value = current
    if (activation.value?.status === 'received' || activation.value?.status === 'finished') {
      stopPolling()
      toast.success('🎉 OpenAI 验证码接收成功！')
      emit('refresh')
    } else if (activation.value?.status === 'pending') {
      startPolling()
    }
  } catch (err: any) {
    toast.error(err?.message || '读取短信失败，请重试')
  } finally {
    checking.value = false
  }
}

const handleRetry = async () => {
  if (retrying.value) return
  retrying.value = true
  stopPolling()
  try {
    const api = getApiClient()
    const res: any = await api.retry()
    const current = readActivation(res)
    if (current) {
      activation.value = current
      toast.success('已成功为您更换新号码！')
      startCountdown()
    } else {
      toast.error(t('common.api.requestFailed'))
    }
  } catch (err: any) {
    toast.error(err?.message || '更换号码失败，请重试')
  } finally {
    retrying.value = false
  }
}

const copyPhone = async () => {
  if (!activation.value?.phone) return
  try {
    await navigator.clipboard.writeText(activation.value.phone)
    copiedPhone.value = true
    toast.success('手机号已复制')
    setTimeout(() => { copiedPhone.value = false }, 2000)
  } catch {
    toast.error('复制失败，请手动长按复制')
  }
}

const copyCode = async () => {
  if (!activation.value?.code) return
  try {
    await navigator.clipboard.writeText(activation.value.code)
    copiedCode.value = true
    toast.success('验证码已复制')
    setTimeout(() => { copiedCode.value = false }, 2000)
  } catch {
    toast.error('复制失败，请手动长按复制')
  }
}

onMounted(() => {
  loadCurrentActivation()
})

onUnmounted(() => {
  clearInterval(pollTimer)
  clearInterval(countdownTimer)
})
</script>
