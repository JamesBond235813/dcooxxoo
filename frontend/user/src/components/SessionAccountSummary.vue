<template>
  <div class="mt-3 rounded-xl border border-emerald-200 bg-emerald-50/70 p-4 text-sm dark:border-emerald-900 dark:bg-emerald-950/20">
    <div class="font-semibold text-foreground">ZovoCard 账号检查结果</div>
    <div class="mt-3 grid grid-cols-1 gap-x-6 gap-y-2 sm:grid-cols-2">
      <div v-for="row in rows" :key="row.label" class="flex items-start justify-between gap-3">
        <span class="shrink-0 text-muted-foreground">{{ row.label }}</span>
        <span class="text-right font-medium text-foreground break-words">{{ row.value || '上游未提供' }}</span>
      </div>
    </div>
    <p class="mt-3 text-xs text-muted-foreground">以上账号和套餐信息来自 ZovoCard 实时预检；确认账号正确后，系统才会继续充值。</p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ account: Record<string, any> }>()

const text = (value: unknown) => {
  if (value === null || value === undefined || value === '') return ''
  if (typeof value === 'boolean') return value ? '是' : '否'
  return String(value)
}

const formatDate = (value: unknown) => {
  if (!value) return ''
  const date = new Date(String(value))
  return Number.isNaN(date.getTime()) ? String(value) : date.toLocaleString()
}

const planText = (value: unknown) => {
  const plan = text(value)
  if (!plan) return ''
  const labels: Record<string, string> = { free: '免费版', plus: 'Plus', pro: 'Pro', pro_5x: 'Pro 5x', pro_20x: 'Pro 20x', pro_50x: 'Pro 50x' }
  return labels[plan.toLowerCase()] || plan
}

const subscriptionStatus = computed(() => {
  if ((props.account.subscription_recovery_required ?? props.account.subscriptionRecoveryRequired) === true) return '需要先处理现有订阅状态'
  if ((props.account.subscription_is_delinquent ?? props.account.subscriptionIsDelinquent) === true) return '订阅付款异常'
  const active = props.account.subscription_has_active ?? props.account.subscriptionHasActive
  if (active === true) return '订阅有效'
  if (active === false || String(props.account.currentPlan || props.account.current_plan || '').toLowerCase() === 'free') return '已到期 / 免费版'
  return ''
})

const remaining = computed(() => {
  const until = props.account.subscription_active_until || props.account.subscriptionActiveUntil
  if (!until) return ''
  const milliseconds = new Date(String(until)).getTime() - Date.now()
  if (!Number.isFinite(milliseconds)) return ''
  if (milliseconds <= 0) return '已到期'
  const days = Math.ceil(milliseconds / 86400000)
  return `约 ${days} 天`
})

const formatRecord = (value: unknown) => {
  if (!value) return ''
  if (typeof value === 'string' || typeof value === 'number') return String(value)
  if (typeof value !== 'object') return ''
  const data = value as Record<string, unknown>
  const amount = data.amountMajor ?? data.amount_major ?? data.amount ?? data.amountMinor ?? data.amount_minor
  return [amount, data.currency, data.status, data.brand || data.name || data.type]
    .map(text).filter(Boolean).join(' · ')
}

const paymentTime = computed(() => {
  const payment = props.account.last_payment || props.account.lastPayment
  if (!payment || typeof payment !== 'object') return ''
  const data = payment as Record<string, unknown>
  return formatDate(data.paidAt || data.paid_at || data.created || data.created_at)
})

const renewalText = computed(() => {
  const renews = props.account.subscription_will_renew ?? props.account.subscriptionWillRenew
  if (renews === true) return '到期后自动续费'
  if (renews === false) return '到期后不续费'
  return ''
})

const paymentMethodText = computed(() => {
  const method = props.account.payment_method || props.account.paymentMethod
  if (!method) return ''
  if (typeof method === 'string') return method
  const data = method as Record<string, unknown>
  const brand = data.brand || data.type || data.name
  const last4 = data.last4 || data.last_4
  return [brand, last4 ? `**** ${String(last4)}` : ''].filter(Boolean).map(text).join(' ')
})

const rows = computed(() => [
  { label: '账号', value: text(props.account.email || props.account.account_email) },
  { label: '当前套餐', value: planText(props.account.currentPlan || props.account.current_plan || props.account.plan) },
  { label: '订阅状态', value: subscriptionStatus.value },
  { label: '套餐到期', value: formatDate(props.account.subscription_active_until || props.account.subscriptionActiveUntil) },
  { label: '剩余', value: remaining.value },
  { label: '自动续费', value: renewalText.value },
  { label: '最近付款时间', value: paymentTime.value },
  { label: '最近付款', value: formatRecord(props.account.last_payment || props.account.lastPayment) },
  { label: '支付方式', value: paymentMethodText.value },
])
</script>
