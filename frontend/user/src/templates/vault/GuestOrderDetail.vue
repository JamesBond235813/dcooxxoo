<template>
  <div class="mx-auto w-full max-w-[1180px] px-6 pb-8">
    <div class="my-5 flex items-start justify-between gap-4">
      <div>
        <h1 class="mb-1.5 text-3xl font-extrabold">{{ t('guestOrderDetail.title') }}</h1>
        <p class="text-muted-foreground">{{ t('guestOrderDetail.subtitle') }}</p>
      </div>
      <RouterLink class="text-[13px] font-semibold text-muted-foreground transition-colors hover:text-primary" to="/guest/orders">{{ t('guestOrderDetail.backSearch') }}</RouterLink>
    </div>

    <!-- 游客验证 -->
    <div v-if="viewState === 'auth'" class="mb-[18px] rounded-xl border bg-card p-[22px]">
      <h2 class="mb-1.5 text-lg font-bold">{{ t('guestOrderDetail.authTitle') }}</h2>
      <p class="mb-3.5 text-[13px] text-muted-foreground">{{ t('guestOrderDetail.authHint') }}</p>
      <div class="grid gap-3 sm:grid-cols-2">
        <Input v-model="auth.email" type="email" name="guest-order-detail-email" autocomplete="off" class="h-11" :placeholder="t('guestOrders.emailPlaceholder')" />
      </div>
      <div v-if="authError" class="mt-3.5 rounded-sm bg-destructive/10 px-3 py-2.5 text-[13px] font-semibold text-destructive">{{ authError }}</div>
      <div class="mt-3.5 flex flex-wrap gap-3">
        <Button type="button" size="sm" class="rounded-full" @click="handleAuthSubmit">{{ t('guestOrderDetail.authSubmit') }}</Button>
        <Button type="button" variant="ghost" size="sm" class="rounded-full" @click="clearAuth">{{ t('guestOrderDetail.authClear') }}</Button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="viewState === 'loading'" class="rounded-xl border bg-card p-[22px]">
      <div class="mb-4 h-5 w-[35%] rounded bg-secondary"></div>
      <div class="h-[200px] rounded-md bg-secondary"></div>
    </div>

    <!-- 不存在 -->
    <div v-else-if="viewState === 'empty'" class="my-6 flex flex-col items-center gap-3 rounded-xl border border-dashed py-16 text-center text-muted-foreground">
      <AlertCircle class="h-10 w-10 opacity-60" />
      <p>{{ t('guestOrderDetail.notFound') }}</p>
    </div>

    <template v-else-if="viewState === 'detail' && order">
      <SmsActivationConsole v-if="isSmsOrder && isSmsEligible" :order="order" :is-guest="true" :guest-email="auth.email" @refresh="handleSmsRefresh" />
        <SessionProgress v-if="isSessionOrder" :order="order" :submitting="sessionSubmitting" class="mb-5" />
      <details :open="!(isSmsOrder && isSmsEligible)" class="space-y-4">
        <summary v-if="isSmsOrder" class="cursor-pointer rounded-xl border bg-card p-3 text-sm font-medium">{{ t('smsConsole.orderDetails') }}</summary>
      <section v-if="needsSession" class="mb-[18px] rounded-xl border border-blue-200 bg-blue-50/70 p-[22px] dark:border-blue-900 dark:bg-blue-950/30">
        <h2 class="font-bold text-foreground">提交 Session 完成充值</h2>
        <div class="mt-3 rounded-xl bg-background/80 p-4 text-sm text-muted-foreground">
          <p class="font-semibold text-foreground">获取 Session</p>
          <ol class="mt-2 list-decimal space-y-1 pl-5">
            <li>在电脑或手机浏览器登录 <span class="font-medium text-foreground">chatgpt.com</span>。</li>
            <li>保持登录，在同一浏览器打开 <a class="break-all text-primary underline" href="https://chatgpt.com/api/auth/session" target="_blank" rel="noopener noreferrer">https://chatgpt.com/api/auth/session</a>。</li>
            <li>复制页面返回的完整 JSON，粘贴到下方输入框；操作期间保持 ChatGPT 登录。</li>
          </ol>
        </div>
        <textarea v-model="sessionValue" rows="5" autocomplete="off" spellcheck="false" class="mt-3 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm text-foreground" placeholder="粘贴完整 Session JSON" :disabled="sessionSubmitting || sessionPreflighting" />
        <Button class="mt-3 rounded-full" size="sm" :disabled="sessionSubmitting || sessionPreflighting || !sessionJSONValid" @click="preflightSession">{{ sessionPreflighting ? '正在向 ZovoCard 检查账号…' : '检查账号信息' }}</Button>
        <div v-if="sessionReviewOpen" class="mt-4 rounded-xl border bg-background p-4">
          <h3 class="font-semibold text-foreground">请确认充值账号</h3>
          <SessionAccountSummary v-if="sessionAccountSummary" :account="sessionAccountSummary" />
          <p class="mt-1 text-xs text-muted-foreground">确认后系统会提交 Session 并开始自动处理。请确认这是你要充值的账号。</p>
          <div class="mt-3 flex gap-2"><Button size="sm" :disabled="sessionSubmitting || sessionPreflighting || !sessionAccountSummary" @click="submitSession">{{ sessionSubmitting ? '提交中…' : '确认无误，提交并继续充值' }}</Button><Button size="sm" variant="outline" :disabled="sessionSubmitting || sessionPreflighting" @click="sessionReviewOpen = false">返回修改</Button></div>
        </div>
      </section>
      <div class="mb-[18px] flex flex-wrap items-start justify-between gap-[18px] rounded-xl border bg-card p-[22px]">
        <div>
          <div class="text-[11px] uppercase tracking-[0.06em] text-muted-foreground">{{ t('orders.orderNo') }}</div>
          <div class="mt-1 font-bold">{{ order.order_no }}</div>
          <div class="mt-1.5 text-[13px] text-muted-foreground">{{ t('orderDetail.createdAtLabel') }}：{{ formatDate(order.created_at) }}</div>
        </div>
        <div>
          <div class="text-[11px] uppercase tracking-[0.06em] text-muted-foreground">{{ t('orderDetail.amountTotal') }}</div>
          <div class="mt-1 text-2xl font-extrabold tabular-nums">{{ formatMoney(order.total_amount, order.currency) }}</div>
        </div>
        <div class="flex flex-wrap items-center gap-2.5">
          <Badge :variant="statusVariant(order.status)" class="rounded-full">{{ statusLabel(order.status) }}</Badge>
          <Button v-if="order.status === 'pending_payment'" as-child size="sm" class="rounded-full">
            <RouterLink :to="`/pay?guest=1&order_no=${order.order_no}`">{{ t('orders.payNow') }}</RouterLink>
          </Button>
        </div>
      </div>

      <VaultOrderBody
        :order="order"
        variant="guest"
        :fulfillment-downloading="fulfillmentDownloading"
        @download="handleDownloadFulfillment"
      />
      </details>
    </template>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { AlertCircle } from 'lucide-vue-next'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import VaultOrderBody from './components/VaultOrderBody.vue'
import { useGuestOrderDetail } from '../../composables/useGuestOrderDetail'
import { hasPendingSession } from '../../utils/sessionOrder'
import { guestOrderAPI } from '../../api/order'
import { toast } from '../../composables/useToast'
import { watch } from 'vue'
import SessionAccountSummary from '../../components/SessionAccountSummary.vue'
import SessionProgress from '../../components/SessionProgress.vue'
import SmsActivationConsole from '../../components/SmsActivationConsole.vue'
import { isSmsActivationOrder, isSmsActiveStatus } from '../../utils/smsOrder'

const { t } = useI18n()

const {
  order, authError, auth, viewState, handleAuthSubmit, clearAuth,
  fulfillmentDownloading, handleDownloadFulfillment,
  statusLabel, statusVariant, formatDate, formatMoney,
} = useGuestOrderDetail()

const sessionValue = ref('')
const sessionSubmitting = ref(false)
const sessionPreflighting = ref(false)
const sessionReviewOpen = ref(false)
const sessionAccountSummary = ref<Record<string, any> | null>(null)
const sessionJSONValid = computed(() => { try { return JSON.parse(sessionValue.value.trim()) !== null } catch { return false } })
watch(sessionValue, () => { sessionAccountSummary.value = null; sessionReviewOpen.value = false })
const isSmsOrder = computed(() => isSmsActivationOrder(order.value))
const isSmsEligible = computed(() => isSmsActiveStatus(order.value))
const handleSmsRefresh = async () => { if (order.value?.order_no && auth.value?.email) await handleAuthSubmit() }
const isSessionOrder = computed(() => !!order.value && [order.value, ...(order.value.children || [])].some((entry: any) => (entry.items || []).some((item: any) => item.requires_session)))
const needsSession = computed(() => {
  return hasPendingSession(order.value)
})
let progressTimer: ReturnType<typeof setInterval> | undefined
let progressLoading = false
onMounted(() => {
  progressTimer = setInterval(async () => {
    if (progressLoading || !order.value?.order_no || !auth.value.email || needsSession.value || !['paid', 'fulfilling', 'partially_delivered'].includes(order.value.status)) return
    progressLoading = true
    try {
      const response = await guestOrderAPI.detail(order.value.order_no, { email: auth.value.email })
      order.value = response.data.data
    } catch {} finally { progressLoading = false }
  }, 4000)
})
onUnmounted(() => { if (progressTimer) clearInterval(progressTimer) })
const submitSession = async () => {
  if (!order.value?.order_no || !sessionValue.value.trim() || sessionSubmitting.value || !sessionAccountSummary.value) return
  sessionSubmitting.value = true
  try {
    await guestOrderAPI.submitSession(order.value.order_no, {
      email: auth.value.email,
      session: sessionValue.value.trim(),
    })
    sessionValue.value = ''
    toast.success('Session 已提交，系统正在继续处理订单')
    await handleAuthSubmit()
  } catch (error: any) {
    toast.error(error?.message || '提交失败，请稍后重试')
  } finally {
    sessionSubmitting.value = false
  }
}
const preflightSession = async () => {
  if (!order.value?.order_no || !sessionJSONValid.value || sessionPreflighting.value) return
  sessionPreflighting.value = true
  sessionAccountSummary.value = null
  try {
    const result = await guestOrderAPI.preflightSession(order.value.order_no, {
      email: auth.value.email,
      session: sessionValue.value.trim(),
    })
    sessionAccountSummary.value = result.data?.data || null
    if (!sessionAccountSummary.value) throw new Error('ZovoCard 未返回账号信息')
    sessionReviewOpen.value = true
  } catch (error: any) {
    toast.error(error?.message || 'ZovoCard 账号检查失败，请确认 Session 后重试')
  } finally {
    sessionPreflighting.value = false
  }
}
</script>
