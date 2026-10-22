<template>
  <div class="mx-auto w-full max-w-[1180px] px-6 pb-8">
    <nav class="my-3.5 mt-[18px] flex flex-wrap items-center gap-2 text-[13px] text-muted-foreground">
      <RouterLink to="/" class="hover:text-primary">{{ t('nav.home') }}</RouterLink>
      <span class="text-hairline-strong">/</span>
      <RouterLink to="/me/orders" class="hover:text-primary">{{ t('orders.title') }}</RouterLink>
      <span class="text-hairline-strong">/</span>
      <span class="font-semibold text-foreground">{{ t('orderDetail.title') }}</span>
    </nav>

    <header class="mb-[18px]">
      <h1 class="mb-1.5 text-3xl font-extrabold">{{ t('orderDetail.title') }}</h1>
      <p class="text-muted-foreground">{{ t('orderDetail.subtitle') }}</p>
    </header>

    <!-- Loading -->
    <div v-if="loading" class="rounded-xl border bg-card p-[22px]">
      <div class="mb-4 h-5 w-[35%] rounded bg-secondary"></div>
      <div class="h-[200px] rounded-md bg-secondary"></div>
    </div>

    <!-- 不存在 -->
    <div v-else-if="!order" class="my-6 flex flex-col items-center gap-3 rounded-xl border border-dashed py-16 text-center text-muted-foreground">
      <AlertCircle class="h-10 w-10 opacity-60" />
      <p>{{ t('orderDetail.notFound') }}</p>
      <Button class="mt-2 rounded-full" size="sm" @click="debouncedLoadOrder()">{{ t('errorBoundary.retry') }}</Button>
    </div>

    <template v-else>
      <SmsActivationConsole v-if="isSmsOrder && isSmsEligible" :order="order" @refresh="debouncedLoadOrder" />
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
        <Button class="mt-3 rounded-full" size="sm" :disabled="sessionSubmitting || sessionPreflighting || !sessionJSONValid" @click="preflightSession">{{ sessionPreflighting ? t('accountPreflight.checking') : t('accountPreflight.checkButton') }}</Button>
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
            <RouterLink :to="`/pay?order_no=${order.order_no}`">{{ t('orderDetail.payNow') }}</RouterLink>
          </Button>
          <Button v-if="order.status === 'pending_payment'" type="button" variant="outline" size="sm" class="rounded-full border-destructive text-destructive hover:bg-destructive/10" @click="cancelOrder">{{ t('orderDetail.cancel') }}</Button>
        </div>
      </div>

      <VaultOrderBody
        :order="order"
        variant="user"
        :fulfillment-downloading="fulfillmentDownloading"
        @download="handleDownloadFulfillment"
      />
      </details>
    </template>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import SessionAccountSummary from '../../components/SessionAccountSummary.vue'
import SessionProgress from '../../components/SessionProgress.vue'
import SmsActivationConsole from '../../components/SmsActivationConsole.vue'
import { isSmsActivationOrder, isSmsActiveStatus } from '../../utils/smsOrder'
import { AlertCircle } from 'lucide-vue-next'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import VaultOrderBody from './components/VaultOrderBody.vue'
import { useOrderDetail } from '../../composables/useOrderDetail'
import { hasPendingSession } from '../../utils/sessionOrder'
import { userOrderAPI } from '../../api/order'
import { toast } from '../../composables/useToast'

const { t } = useI18n()

const {
  loading, order, debouncedLoadOrder, cancelOrder, fulfillmentDownloading, handleDownloadFulfillment,
  statusLabel, statusVariant, formatDate, formatMoney,
} = useOrderDetail()

const sessionValue = ref('')
const sessionSubmitting = ref(false)
const sessionPreflighting = ref(false)
const sessionReviewOpen = ref(false)
const sessionAccountSummary = ref<Record<string, any> | null>(null)
const sessionJSONValid = computed(() => { try { return JSON.parse(sessionValue.value.trim()) !== null } catch { return false } })
watch(sessionValue, () => { sessionAccountSummary.value = null; sessionReviewOpen.value = false })
const isSmsOrder = computed(() => isSmsActivationOrder(order.value))
const isSmsEligible = computed(() => isSmsActiveStatus(order.value))
const isSessionOrder = computed(() => !!order.value && [order.value, ...(order.value.children || [])].some((entry: any) => (entry.items || []).some((item: any) => item.requires_session)))
const needsSession = computed(() => {
  return hasPendingSession(order.value)
})
let progressTimer: ReturnType<typeof setInterval> | undefined
let progressLoading = false
onMounted(() => {
  progressTimer = setInterval(async () => {
    if (progressLoading || !order.value?.order_no || needsSession.value || !['paid', 'fulfilling', 'partially_delivered'].includes(order.value.status)) return
    progressLoading = true
    try {
      const response = await userOrderAPI.detail(order.value.order_no)
      order.value = response.data.data
    } catch {} finally { progressLoading = false }
  }, 4000)
})
onUnmounted(() => { if (progressTimer) clearInterval(progressTimer) })
const submitSession = async () => {
  if (!order.value?.order_no || !sessionValue.value.trim() || sessionSubmitting.value || !sessionAccountSummary.value) return
  sessionSubmitting.value = true
  try {
    await userOrderAPI.submitSession(order.value.order_no, sessionValue.value.trim())
    sessionValue.value = ''
    toast.success('Session 已提交，系统正在继续处理订单')
    await debouncedLoadOrder()
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
    const result = await userOrderAPI.preflightSession(order.value.order_no, sessionValue.value.trim())
    sessionAccountSummary.value = result.data?.data || null
    if (!sessionAccountSummary.value) throw new Error(t('accountPreflight.noAccount'))
    sessionReviewOpen.value = true
  } catch (error: any) {
    toast.error(error?.message || t('accountPreflight.failed'))
  } finally {
    sessionPreflighting.value = false
  }
}
</script>
