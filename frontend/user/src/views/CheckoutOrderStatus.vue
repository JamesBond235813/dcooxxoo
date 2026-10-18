<template>
  <main class="min-h-screen bg-background px-4 pb-20 pt-28 text-foreground">
    <div class="mx-auto max-w-[960px]">
      <h1 class="text-3xl font-bold">订单详情</h1>
      <p class="mt-2 break-all text-sm text-muted-foreground">{{ orderNo }}</p>
      <div v-if="loading" class="mt-8 rounded-2xl border bg-card p-8 text-muted-foreground">正在读取订单…</div>
      <div v-else-if="error" class="mt-8 rounded-2xl border bg-card p-8">
        <p class="text-destructive">{{ error }}</p>
        <div v-if="isGuest" class="mt-4 grid gap-3 sm:grid-cols-2">
          <input v-model.trim="guestEmail" type="email" placeholder="接收订单信息的邮箱" class="rounded-lg border bg-background p-3" />
        </div>
        <Button class="mt-4" @click="loadOrder">重新读取</Button>
      </div>
      <article v-else-if="order" class="mt-8 rounded-2xl border bg-card p-6 shadow-sm md:p-8">
        <div v-if="paymentSuccessRedirecting" role="status" class="mb-5 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm font-semibold text-emerald-800 dark:border-emerald-900 dark:bg-emerald-950/30 dark:text-emerald-300">
          支付成功，正在前往 Session 提交页面…
        </div>
        <div class="flex flex-wrap items-center justify-between gap-4">
          <h2 class="font-semibold">{{ productTitle }}</h2>
          <strong class="text-2xl">{{ amountText }}</strong>
        </div>
        <div class="mt-7 flex flex-wrap items-center gap-2 text-xs font-medium sm:text-sm">
          <template v-for="(step, index) in steps" :key="step">
            <span :class="index <= currentStep ? 'text-emerald-700' : 'text-muted-foreground'">{{ index + 1 }}. {{ step }}</span>
            <span v-if="index < steps.length - 1" class="mx-1 text-muted-foreground">→</span>
          </template>
        </div>
        <div class="mt-6 flex flex-wrap gap-2 text-xs">
          <span class="rounded bg-secondary px-2 py-1">{{ statusText }}</span>
          <span v-if="order.status === 'pending_payment'" class="rounded bg-emerald-50 px-2 py-1 text-emerald-700">准备开通</span>
        </div>
        <p class="mt-3 text-xs text-muted-foreground">创建于 {{ new Date(order.created_at).toLocaleString() }}</p>
        <div class="mt-5 flex flex-wrap gap-3">
          <Button v-if="order.status === 'pending_payment'" :disabled="openingPayment" @click="goPay">{{ openingPayment ? '正在前往支付…' : '前往支付' }}</Button>
          <Button v-else-if="canContinue" @click="goContinue">继续完成充值</Button>
          <Button variant="outline" :disabled="loading" @click="loadOrder">刷新进度</Button>
          <span v-if="paymentChannel" class="hidden items-center gap-2 rounded-lg border bg-card px-3 py-2 text-sm text-muted-foreground md:inline-flex" :aria-label="`支付方式：${paymentChannel.name}`">
            <img v-if="paymentChannel.icon" :src="getImageUrl(paymentChannel.icon)" alt="" class="h-5 w-5 object-contain" />
            <span v-else class="flex h-5 w-5 items-center justify-center rounded-full bg-sky-100 text-[10px] font-bold text-sky-700">{{ paymentChannelFallback }}</span>
            <span>{{ paymentChannel.name }}</span>
          </span>
        </div>
        <p v-if="order.status === 'pending_payment'" class="mt-4 text-xs leading-relaxed text-muted-foreground">{{ paymentAmountNotice }} 如已完成付款，请勿重复支付；系统会自动核对支付结果，也可点击“刷新进度”。</p>
      </article>
      <RouterLink :to="orderQueryBackPath" class="mt-5 inline-block text-sm text-emerald-700 underline md:hidden">返回订单查询</RouterLink>
      <RouterLink :to="orderQueryBackPath" class="mt-5 hidden text-sm text-emerald-700 underline md:inline-block">返回订单查询</RouterLink>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { guestOrderAPI, paymentAPI, userOrderAPI } from '../api/order'
import { loadGuestOrderAuth, saveGuestOrderAuth } from '../utils/guestOrderAuth'
import { useLocalized } from '../composables/useProduct'
import { isRedirectPaymentInteractionMode } from '../utils/paymentResumePolicy'
import { isPaymentInteractionModeForDevice } from '../composables/usePaymentChannelsForDevice'
import { useAppStore } from '../stores/app'
import { getImageUrl } from '../utils/image'

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const { getLocalizedText, formatPrice } = useLocalized()
const orderNo = computed(() => String(route.query.order_no || ''))
const isGuest = computed(() => route.query.agent !== '1')
const savedAuth = loadGuestOrderAuth()
const guestEmail = ref(savedAuth.email)
const order = ref<any>(null)
const latestPayment = ref<any>(null)
const paymentChannel = computed(() => {
  const payment = latestPayment.value
  const channels = Array.isArray(appStore.config?.payment_channels) ? appStore.config.payment_channels : []
  const configured = channels.find((channel: any) => Number(channel.id) === Number(payment?.channel_id))
  const name = String(configured?.name || payment?.channel_name || payment?.display_channel_type || payment?.channel_type || '').trim()
  if (!name) return null
  return { name, icon: configured?.icon || payment?.channel_icon || '' }
})
const paymentChannelFallback = computed(() => paymentChannel.value?.name?.slice(0, 1) || '付')
const orderQueryBackPath = '/agent?tab=orders'
const loading = ref(false)
const openingPayment = ref(false)
const paymentSuccessRedirecting = ref(false)
const error = ref('')
let orderLoadInFlight = false
let paymentStatusPollTimer: number | null = null
let paymentSuccessRedirectTimer: number | null = null
let paymentStatusPollCount = 0
const steps = ['支付', '准备开通', '资格与开通', '完成']
const productTitle = computed(() => getLocalizedText(order.value?.items?.[0]?.title_snapshot || order.value?.items?.[0]?.title || order.value?.items?.[0]?.product_title) || '订单商品')
const amountText = computed(() => {
  const currency = String(order.value?.currency || 'CNY').toUpperCase()
  const formatted = formatPrice(order.value?.total_amount || '0.00', currency)
  return currency === 'CNY' ? `￥${formatted.replace(/CNY|CN¥|¥|￥/gi, '').trim()}` : formatted
})
const paymentAmountNotice = computed(() => {
  const payment = latestPayment.value
  if (!payment || !['customer_surcharge', 'legacy_customer_surcharge'].includes(String(payment.fee_policy || '').toLowerCase())) {
    return '前往支付将直接打开支付渠道；如渠道收取手续费，最终金额以收银台显示为准。'
  }
  const amount = payment.payable_amount ?? payment.amount ?? payment.online_pay_amount
  const currency = String(payment.currency || order.value?.currency || 'CNY').toUpperCase()
  const formatted = formatPrice(String(amount ?? ''), currency)
  const display = currency === 'CNY' ? `￥${formatted.replace(/CNY|CN¥|¥|￥/gi, '').trim()}` : formatted
  return amount !== undefined && amount !== null && amount !== ''
    ? `本次支付含渠道费用，收银台应付金额约为 ${display}。`
    : '本次支付可能包含渠道费用，最终金额以收银台显示为准。'
})
const currentStep = computed(() => order.value?.status === 'pending_payment' ? 0 : ['completed', 'delivered'].includes(order.value?.status) ? 3 : order.value?.status === 'paid' ? 1 : 2)
const statusText = computed(() => ({ pending_payment: '待付款', paid: '已支付', fulfilling: '履约处理中', partially_delivered: '部分履约', delivered: '已交付', completed: '已完成', canceled: '已取消', refunded: '已退款', partially_refunded: '部分退款', expired: '已过期' } as Record<string,string>)[order.value?.status] || '订单处理中')
const canContinue = computed(() => !['canceled', 'refunded', 'completed', 'delivered'].includes(String(order.value?.status || '')))
const hasPendingSession = (value: any) => [value, ...(value?.children || [])].some((entry: any) =>
  (entry?.items || []).some((item: any) => item.requires_session && !item.session_submitted),
)
function stopPaymentStatusPolling() {
  if (paymentStatusPollTimer) window.clearInterval(paymentStatusPollTimer)
  paymentStatusPollTimer = null
}
function scheduleSessionRedirect() {
  if (paymentSuccessRedirectTimer || paymentSuccessRedirecting.value) return
  paymentSuccessRedirecting.value = true
  stopPaymentStatusPolling()
  const path = isGuest.value
    ? `/guest/orders/${encodeURIComponent(orderNo.value)}/session`
    : `/orders/${encodeURIComponent(orderNo.value)}/session`
  paymentSuccessRedirectTimer = window.setTimeout(() => {
    void router.replace({ path, ...(isGuest.value ? {} : { query: { from: 'agent' } }) })
  }, 1200)
}
function syncPaymentStatus() {
  if (order.value?.status === 'pending_payment') {
    startPaymentStatusPolling()
    return
  }
  stopPaymentStatusPolling()
  if (['paid', 'fulfilling', 'partially_delivered'].includes(String(order.value?.status || '')) && hasPendingSession(order.value)) {
    scheduleSessionRedirect()
  }
}
function startPaymentStatusPolling() {
  if (paymentStatusPollTimer || order.value?.status !== 'pending_payment') return
  paymentStatusPollCount = 0
  paymentStatusPollTimer = window.setInterval(() => {
    if (document.visibilityState !== 'visible') return
    paymentStatusPollCount += 1
    if (paymentStatusPollCount > 60) {
      stopPaymentStatusPolling()
      return
    }
    void loadOrder({ silent: true })
  }, 3000)
}
async function loadOrder(options?: { silent?: boolean }) {
  if (orderLoadInFlight) return
  if (!orderNo.value) { error.value = '缺少订单号'; return }
  orderLoadInFlight = true
  if (!options?.silent) {
    loading.value = true
    error.value = ''
  }
  try {
    if (isGuest.value) {
      if (!guestEmail.value) throw new Error('请填写邮箱')
      saveGuestOrderAuth({ email: guestEmail.value })
      const response = await guestOrderAPI.detail(orderNo.value, { email: guestEmail.value })
      order.value = response.data.data
      syncPaymentStatus()
      try {
        const paymentResponse = await guestOrderAPI.latestPayment({ order_no: orderNo.value, email: guestEmail.value.trim() })
        latestPayment.value = paymentResponse.data.data
      } catch { latestPayment.value = null }
    } else {
      const response = await userOrderAPI.detail(orderNo.value)
      order.value = response.data.data
      syncPaymentStatus()
      try {
        const paymentResponse = await paymentAPI.latest({ order_no: orderNo.value })
        latestPayment.value = paymentResponse.data.data
      } catch { latestPayment.value = null }
    }
  } catch (cause: any) {
    if (!options?.silent) error.value = cause?.message || '读取订单失败'
  } finally {
    orderLoadInFlight = false
    if (!options?.silent) loading.value = false
  }
}
async function goPay() {
  if (openingPayment.value) return
  openingPayment.value = true
  try {
    const response = isGuest.value
      ? await guestOrderAPI.latestPayment({ order_no: orderNo.value, email: guestEmail.value.trim() })
      : await paymentAPI.latest({ order_no: orderNo.value })
    const payment = response.data.data
    const payUrl = String(payment?.pay_url || '').trim()
    if (payUrl && isRedirectPaymentInteractionMode(payment?.interaction_mode) && isPaymentInteractionModeForDevice(payment?.interaction_mode)) {
      const destination = new URL(payUrl, window.location.origin)
      if (destination.protocol === 'https:' || destination.protocol === 'http:') {
        window.location.assign(destination.href)
        return
      }
    }
  } catch {
    // Fall back to the payment page so it can refresh or create the payment normally.
  } finally {
    openingPayment.value = false
  }
  router.push({ path: '/pay', query: { order_no: orderNo.value, ...(isGuest.value ? { guest: '1' } : { agent: '1' }) } })
}
function goContinue() {
  void router.push(isGuest.value
    ? `/guest/orders/${encodeURIComponent(orderNo.value)}/session`
    : { path: `/orders/${encodeURIComponent(orderNo.value)}/session`, query: { from: 'agent' } })
}
function refreshPaymentStateOnResume() {
  if (document.visibilityState === 'visible') void loadOrder({ silent: true })
}
onMounted(() => {
  document.addEventListener('visibilitychange', refreshPaymentStateOnResume)
  window.addEventListener('pageshow', refreshPaymentStateOnResume)
  if (!Array.isArray(appStore.config?.payment_channels)) void appStore.loadConfig()
  void loadOrder()
})
onUnmounted(() => {
  document.removeEventListener('visibilitychange', refreshPaymentStateOnResume)
  window.removeEventListener('pageshow', refreshPaymentStateOnResume)
  stopPaymentStatusPolling()
  if (paymentSuccessRedirectTimer) window.clearTimeout(paymentSuccessRedirectTimer)
})
</script>
