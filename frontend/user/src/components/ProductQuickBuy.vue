<template>
  <Teleport to="body">
    <div v-if="visible" class="fixed inset-0 z-50 flex items-end justify-center bg-black/45 p-0 backdrop-blur-[2px] md:items-center md:p-6" @click.self="close">
      <section role="dialog" aria-modal="true" aria-label="确认订单" class="flex max-h-[94vh] w-full max-w-[940px] flex-col overflow-hidden rounded-t-2xl border bg-card text-card-foreground shadow-2xl md:max-h-[88vh] md:rounded-2xl">
        <header class="flex shrink-0 items-center justify-between border-b px-6 py-5">
          <h2 class="text-xl font-bold">确认订单</h2>
          <button type="button" class="text-muted-foreground hover:text-foreground" aria-label="关闭" @click="close"><X class="h-5 w-5" /></button>
        </header>
        <div class="grid min-h-0 flex-1 overflow-y-auto md:grid-cols-[1.15fr_0.9fr] md:overflow-hidden">
          <div class="grid grid-cols-2 gap-2 bg-emerald-50/30 px-4 py-4 dark:bg-emerald-950/10 md:block md:space-y-7 md:bg-transparent md:px-6 md:py-6 md:overflow-y-auto md:border-r">
            <div class="rounded-2xl border border-emerald-100 bg-white p-3 shadow-sm dark:border-emerald-900/50 dark:bg-card md:rounded-none md:border-0 md:bg-transparent md:p-0 md:shadow-none" :class="activeMobileTerm === 'conditions' ? 'col-span-2' : activeMobileTerm === 'warranty' ? 'hidden md:block' : ''">
              <h3 class="mb-3 hidden font-semibold text-emerald-800 dark:text-emerald-300 md:mb-4 md:block md:text-foreground">{{ conditionsTitle }}</h3>
              <button type="button" class="flex min-h-10 w-full items-center justify-between gap-1 font-semibold text-xs text-emerald-800 dark:text-emerald-300 md:hidden" :aria-expanded="activeMobileTerm === 'conditions'" @click="toggleMobileTerm('conditions')">
                <span>{{ conditionsTitle }}</span><ChevronDown class="h-4 w-4 transition-transform duration-300" :class="conditionsOpen && 'rotate-180'" />
              </button>
              <div class="terms-collapse" :class="conditionsOpen || activeMobileTerm === 'conditions' ? 'open' : 'closed'">
                <div class="min-h-0 overflow-hidden">
                  <ol class="space-y-2 text-xs leading-relaxed text-muted-foreground md:space-y-3 md:text-sm">
                    <li v-for="(line, index) in conditions" :key="index" class="flex gap-2 md:gap-3"><span class="font-mono text-emerald-700">{{ String(index + 1).padStart(2, '0') }}</span><span>{{ line }}</span></li>
                  </ol>
                </div>
              </div>
            </div>
            <div class="rounded-2xl border border-sky-100 bg-white p-3 shadow-sm dark:border-sky-900/50 dark:bg-card md:rounded-none md:border-0 md:border-t md:bg-transparent md:p-0 md:pt-6 md:shadow-none" :class="activeMobileTerm === 'warranty' ? 'col-span-2' : activeMobileTerm === 'conditions' ? 'hidden md:block' : ''">
              <h3 class="mb-3 hidden font-semibold text-sky-800 dark:text-sky-300 md:mb-4 md:block md:text-foreground">{{ warrantyTitle }}</h3>
              <button type="button" class="flex min-h-10 w-full items-center justify-between gap-1 font-semibold text-xs text-sky-800 dark:text-sky-300 md:hidden" :aria-expanded="activeMobileTerm === 'warranty'" @click="toggleMobileTerm('warranty')">
                <span>{{ warrantyTitle }}</span><ChevronDown class="h-4 w-4 transition-transform duration-300" :class="warrantyOpen && 'rotate-180'" />
              </button>
              <div class="terms-collapse" :class="warrantyOpen || activeMobileTerm === 'warranty' ? 'open' : 'closed'">
                <div class="min-h-0 overflow-hidden">
                  <ol class="space-y-2 text-xs leading-relaxed text-muted-foreground md:space-y-3 md:text-sm">
                    <li v-for="(line, index) in activeWarranty" :key="line" class="flex gap-2 md:gap-3"><span class="font-mono text-sky-700">{{ String(index + 1).padStart(2, '0') }}</span><span>{{ line }}</span></li>
                  </ol>
                </div>
              </div>
            </div>
          </div>
          <div class="space-y-5 px-6 py-6 md:overflow-y-auto">
            <div>
              <div class="flex items-center justify-between gap-3 md:block">
                <h3 class="text-lg font-semibold">订单信息</h3>
                <p class="min-w-0 truncate text-right text-sm text-muted-foreground md:mt-2 md:text-left">{{ productTitle }}</p>
              </div>
            </div>
            <div v-if="activeSkus.length > 1">
              <label class="mb-1 block text-sm font-medium">套餐规格</label>
              <select v-model.number="selectedSkuId" class="w-full rounded-lg border bg-background px-3 py-2.5 text-sm">
                <option v-for="sku in activeSkus" :key="sku.id" :value="normalizeSkuId(sku.id)">{{ skuDisplayText(sku) }}</option>
              </select>
            </div>
            <div v-if="effectiveMin !== 1 || (effectiveLimit !== null && effectiveLimit !== 1)" class="flex items-center justify-between gap-3 text-sm">
              <label>数量</label>
              <input v-model.number="quantity" type="number" :min="effectiveMin" :max="effectiveLimit ?? undefined" class="w-24 rounded-lg border bg-background px-3 py-2 text-right" />
            </div>
            <div>
              <label for="quick-buy-email" class="mb-2 block text-sm font-medium">联系邮箱（仅用来接收订单信息）</label>
              <input id="quick-buy-email" v-model.trim="guestEmail" type="email" autocomplete="email" placeholder="name@example.com" class="w-full rounded-lg border bg-background px-3 py-2.5 text-sm outline-none focus:border-primary" />
              <p v-if="!useAgentOrderAPI" class="mt-1.5 text-xs text-muted-foreground">创建后可使用此邮箱查询订单。</p>
            </div>
            <div>
              <span class="mb-2 flex items-center gap-1.5 text-sm font-medium md:block">
                <span>支付方式</span>
              </span>
              <div class="flex gap-2 overflow-x-auto pb-1" role="group" aria-label="支付方式">
                <button v-for="channel in paymentChannels" :key="channel.id" type="button" :aria-pressed="selectedChannelId === Number(channel.id)" @click="selectedChannelId = Number(channel.id)" class="flex min-w-0 flex-1 items-center justify-center gap-1.5 whitespace-nowrap rounded-xl border px-3 py-2.5 text-xs font-semibold transition-colors" :class="selectedChannelId === Number(channel.id) ? 'border-emerald-500 bg-emerald-50 text-emerald-800 ring-1 ring-emerald-500/30 dark:bg-emerald-950/30 dark:text-emerald-300' : 'bg-background text-muted-foreground hover:border-emerald-300'">
                  <img v-if="channel.icon" :src="channel.icon" alt="" class="shrink-0 object-contain" :class="String(channel.channel_type || '').toLowerCase() === 'alipay' ? 'h-5 w-5 md:h-4 md:w-4' : 'h-4 w-4'" />
                  <component :is="channelIcon(channel)" v-else class="h-4 w-4 shrink-0" />{{ channelDisplayName(channel) }}
                </button>
              </div>
              <p v-if="!paymentChannels.length" class="mt-1 text-xs text-destructive">当前没有可用支付方式</p>
            </div>
            <div class="flex items-center justify-between border-y py-4">
              <span class="text-sm font-medium">应付金额</span>
              <strong class="text-2xl">{{ formatPrice(serverQuote?.total_amount ?? calculatedTotalPrice, siteCurrency) }}</strong>
            </div>
            <p v-if="quoteLoading" class="text-xs text-muted-foreground">正在核对服务端价格…</p>
            <p v-if="purchaseWarning" class="text-xs text-destructive">{{ purchaseWarning }}</p>
            <div v-if="captchaEnabled" class="rounded-lg border bg-secondary/30 p-3">
              <TurnstileCaptcha v-if="captchaProvider === 'turnstile' && turnstileSiteKey" ref="turnstileCaptchaRef" v-model="turnstileToken" :site-key="turnstileSiteKey" />
              <ImageCaptcha v-else-if="captchaProvider === 'image'" v-model="imageCaptchaPayload" />
            </div>
            <p class="text-xs leading-relaxed text-muted-foreground">{{ submissionDescription }}</p>
            <Button class="h-11 w-full rounded-lg" :disabled="submitting || quoteLoading || !serverQuote || !canPurchase || !selectedChannelId || !guestEmail || (captchaEnabled && !captchaReady)" @click="handleFastCheckoutAndPay">
              <Loader2 v-if="submitting" class="mr-2 h-4 w-4 animate-spin" />{{ submitting ? '正在创建订单…' : '创建订单' }}
            </Button>
          </div>
        </div>
      </section>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '../stores/app'
import { useUserAuthStore } from '../stores/userAuth'
import { useLocalized } from '../composables/useProduct'
import { normalizeSkuId, buildSkuDisplayText } from '../utils/sku'
import { getAffiliateCode, getAffiliateVisitorKey } from '../utils/affiliate'
import { loadGuestOrderAuth, saveGuestOrderAuth } from '../utils/guestOrderAuth'
import { purchaseConditions, purchaseWarranty as subscriptionWarranty } from '../utils/purchaseTerms'
import { guestOrderAPI, userOrderAPI } from '../api/order'
import { Button } from '@/components/ui/button'
import { Banknote, ChevronDown, CreditCard, Landmark, Loader2, QrCode, Wallet, X } from 'lucide-vue-next'
import TurnstileCaptcha from './captcha/TurnstileCaptcha.vue'
import ImageCaptcha from './captcha/ImageCaptcha.vue'
import { usePaymentChannelsForDevice } from '../composables/usePaymentChannelsForDevice'

const props = withDefaults(defineProps<{ product: any; visible: boolean; initialSkuId?: number; initialQuantity?: number; agentMode?: boolean }>(), { agentMode: false })
const emit = defineEmits<{ 'update:visible': [value: boolean] }>()
const router = useRouter()
const { t } = useI18n()
const appStore = useAppStore()
const userAuthStore = useUserAuthStore()
const useAgentOrderAPI = computed(() => props.agentMode && userAuthStore.isAuthenticated)
const { getLocalizedText, formatPrice, siteCurrency } = useLocalized()
const selectedSkuId = ref(0)
const quantity = ref(1)
const guestEmail = ref('')
const selectedChannelId = ref(0)
const availableChannels = ref<any[]>([])
const serverQuote = ref<any>(null)
const quoteLoading = ref(false)
const purchaseWarning = ref('')
const submitting = ref(false)
const turnstileToken = ref('')
const turnstileCaptchaRef = ref<InstanceType<typeof TurnstileCaptcha> | null>(null)
const imageCaptchaPayload = ref<any>({})
const conditionsOpen = ref(true)
const warrantyOpen = ref(true)
const activeMobileTerm = ref<'conditions' | 'warranty' | null>(null)
let termsCollapseTimer: ReturnType<typeof setTimeout> | undefined
let quoteSerial = 0

const productTitle = computed(() => getLocalizedText(props.product?.title))
const isSmsActivation = computed(() => String(props.product?.slug || '') === 'openai-sms-activate')
const isGoPlan = computed(() => String(props.product?.slug || '') === 'upstream-1-1109423947-1791826450925')
const conditionsTitle = computed(() => isSmsActivation.value ? t('homeMobile.smsServiceLabel') : '使用条件')
const warrantyTitle = computed(() => isSmsActivation.value ? t('homeMobile.smsWarranty') : '质保订阅30天')
const conditions = computed(() => isSmsActivation.value
  ? [t('homeMobile.smsCondition1'), t('homeMobile.smsCondition2')]
  : isGoPlan.value ? [t('homeMobile.goCondition1'), t('homeMobile.goCondition2'), t('homeMobile.goCondition3')]
  : /pro/i.test(String(productTitle.value)) ? purchaseConditions.pro : purchaseConditions.plus)
const activeWarranty = computed(() => isSmsActivation.value
  ? [t('homeMobile.smsGuarantee1'), t('homeMobile.smsGuarantee2'), t('homeMobile.smsGuarantee3')]
  : subscriptionWarranty)
const submissionDescription = computed(() => isSmsActivation.value
  ? t('homeMobile.smsCardRequirement')
  : '付款确认后将进入安全开通页面。全程无需领取或输入卡密，请保存订单号。')
const activeSkus = computed(() => (Array.isArray(props.product?.skus) ? props.product.skus : []).filter((sku: any) => sku.is_active))
const selectedSku = computed(() => activeSkus.value.find((sku: any) => normalizeSkuId(sku.id) === selectedSkuId.value))
const skuDisplayText = (sku: any) => buildSkuDisplayText({ skuCode: sku?.sku_code, specValues: sku?.spec_values, fallback: '默认规格', locale: appStore.locale })
const effectiveMin = computed(() => Math.max(1, Number(props.product?.min_purchase_quantity || 1)))
const effectiveLimit = computed(() => Number(props.product?.max_purchase_quantity || 0) || null)
const canPurchase = computed(() => !!props.product && !!selectedSku.value && quantity.value >= effectiveMin.value && (!effectiveLimit.value || quantity.value <= effectiveLimit.value))
const calculatedTotalPrice = computed(() => (Number(selectedSku.value?.price_amount || 0) * quantity.value).toFixed(2))
const captchaProvider = computed(() => String(appStore.config?.captcha?.provider || 'none'))
const captchaEnabled = computed(() => captchaProvider.value !== 'none' && Boolean(appStore.config?.captcha?.scenes?.guest_create_order))
const turnstileSiteKey = computed(() => String(appStore.config?.captcha?.turnstile?.site_key || ''))
const captchaReady = computed(() => captchaProvider.value === 'turnstile' ? Boolean(turnstileToken.value) : Boolean(imageCaptchaPayload.value?.captcha_id && imageCaptchaPayload.value?.captcha_code))
const configuredPaymentChannels = computed(() => {
  const allowed = Array.isArray(props.product?.payment_channel_ids) ? props.product.payment_channel_ids.map(Number) : []
  const list = availableChannels.value.filter((channel: any) => {
    if (channel.is_active === false || channel.enabled === false || channel.status === 'disabled') return false
    if (allowed.length && !allowed.includes(Number(channel.id))) return false
    if (String(channel.provider_type || '').toLowerCase() === 'epay') return ['alipay', 'wechat', 'wxpay', 'qqpay'].includes(String(channel.channel_type || '').toLowerCase())
    return true
  })
  return list.sort((a: any, b: any) => Number(String(b.channel_type).toLowerCase() === 'alipay') - Number(String(a.channel_type).toLowerCase() === 'alipay'))
})
const paymentChannels = usePaymentChannelsForDevice(configuredPaymentChannels)
const channelDisplayName = (channel: any) => String(channel?.name || channel?.title || channel?.channel_type || '支付方式')
const channelIcon = (channel: any) => {
  const type = String(channel?.channel_type || '').toLowerCase()
  if (['alipay', 'wechat', 'wxpay', 'qqpay'].includes(type)) return QrCode
  if (['usdt', 'bepusdt', 'crypto'].includes(type)) return Wallet
  if (['bank', 'bank_transfer'].includes(type)) return Landmark
  if (['card', 'stripe', 'paypal'].includes(type)) return CreditCard
  return Banknote
}
const close = () => emit('update:visible', false)
const toggleMobileTerm = (term: 'conditions' | 'warranty') => {
  if (termsCollapseTimer) clearTimeout(termsCollapseTimer)
  activeMobileTerm.value = activeMobileTerm.value === term ? null : term
  conditionsOpen.value = activeMobileTerm.value === 'conditions'
  warrantyOpen.value = activeMobileTerm.value === 'warranty'
}
const orderItems = () => [{ product_id: Number(props.product?.id), sku_id: selectedSkuId.value || undefined, quantity: Number(quantity.value), fulfillment_type: props.product?.fulfillment_type || undefined, agent_wholesale: Boolean(props.product?.agent_wholesale) || undefined }]
const orderPayload = () => ({ items: orderItems(), email: guestEmail.value.trim(), affiliate_code: getAffiliateCode() || undefined, affiliate_visitor_key: getAffiliateVisitorKey() || undefined })

async function loadConfirmation() {
  if (!props.visible || !props.product) return
  if (!appStore.config) await appStore.loadConfig()
  availableChannels.value = Array.isArray(appStore.config?.payment_channels) ? [...appStore.config.payment_channels] : []
  if (paymentChannels.value.length && !paymentChannels.value.some((item: any) => Number(item.id) === selectedChannelId.value)) selectedChannelId.value = Number(paymentChannels.value[0].id)
  const serial = ++quoteSerial
  quoteLoading.value = true
  serverQuote.value = null
  try {
    const payload = orderPayload()
    const response = useAgentOrderAPI.value ? await userOrderAPI.preview(payload) : await guestOrderAPI.preview({ ...payload, email: payload.email || 'preview@example.com' })
    if (serial !== quoteSerial || !props.visible) return
    serverQuote.value = response.data.data
    if (useAgentOrderAPI.value) {
      const channels = await userOrderAPI.getPaymentChannels({ amount: serverQuote.value?.total_amount, items: orderItems() })
      if (serial !== quoteSerial || !props.visible) return
      availableChannels.value = Array.isArray(channels.data.data) ? channels.data.data : availableChannels.value
    }
    if (!paymentChannels.value.some((item: any) => Number(item.id) === selectedChannelId.value)) selectedChannelId.value = Number(paymentChannels.value[0]?.id || 0)
  } catch (error: any) {
    if (serial === quoteSerial) purchaseWarning.value = error?.message || '订单价格核对失败'
  } finally {
    if (serial === quoteSerial) quoteLoading.value = false
  }
}
watch(() => [props.product, props.visible], () => {
  if (!props.visible || !props.product) return
  conditionsOpen.value = true
  warrantyOpen.value = true
  activeMobileTerm.value = null
  if (termsCollapseTimer) clearTimeout(termsCollapseTimer)
  if (window.matchMedia('(max-width: 767px)').matches) {
    termsCollapseTimer = setTimeout(() => {
      conditionsOpen.value = false
      warrantyOpen.value = false
    }, 1000)
  }
  purchaseWarning.value = ''
  quantity.value = Math.max(effectiveMin.value, Number(props.initialQuantity || 1))
  selectedSkuId.value = activeSkus.value.some((sku: any) => normalizeSkuId(sku.id) === props.initialSkuId) ? Number(props.initialSkuId) : normalizeSkuId(activeSkus.value[0]?.id)
  guestEmail.value = loadGuestOrderAuth().email || ''
  void loadConfirmation()
}, { immediate: true })
watch(() => props.visible, (isVisible) => {
  if (!isVisible && termsCollapseTimer) clearTimeout(termsCollapseTimer)
})
onBeforeUnmount(() => { if (termsCollapseTimer) clearTimeout(termsCollapseTimer) })
watch([selectedSkuId, quantity], () => { if (props.visible) void loadConfirmation() })

async function handleFastCheckoutAndPay() {
  if (!canPurchase.value || submitting.value) return
  purchaseWarning.value = ''
  const email = guestEmail.value.trim()
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) { purchaseWarning.value = '请填写正确的接收订单信息的邮箱'; return }
  if (!selectedChannelId.value) { purchaseWarning.value = '请选择支付方式'; return }
  if (captchaEnabled.value && !captchaReady.value) { purchaseWarning.value = '请完成 Cloudflare 验证'; return }
  submitting.value = true
  try {
    const payload = orderPayload()
    const preview = useAgentOrderAPI.value ? await userOrderAPI.preview(payload) : await guestOrderAPI.preview(payload)
    if (!serverQuote.value || String(preview.data.data?.total_amount) !== String(serverQuote.value.total_amount)) {
      serverQuote.value = preview.data.data
      purchaseWarning.value = '商品价格刚刚发生变化，请核对新金额后再提交'
      return
    }
    const captchaPayload = captchaEnabled.value ? { turnstile_token: turnstileToken.value, ...imageCaptchaPayload.value } : undefined
    const request = { ...payload, channel_id: selectedChannelId.value, captcha_payload: captchaPayload }
    const response = useAgentOrderAPI.value ? await userOrderAPI.createAndPay(request) : await guestOrderAPI.createAndPay(request)
    const result = response.data.data
    if (!result?.order_no) throw new Error('订单创建失败')
    if (!useAgentOrderAPI.value) saveGuestOrderAuth({ email })
    close()
    router.push({ path: '/checkout', query: { order_no: result.order_no, ...(useAgentOrderAPI.value ? { agent: '1' } : { guest: '1' }) } })
  } catch (error: any) {
    purchaseWarning.value = error?.message || '订单创建失败，请重试'
    turnstileToken.value = ''
    turnstileCaptchaRef.value?.reset()
  } finally { submitting.value = false }
}
</script>

<style scoped>
.terms-collapse {
  display: grid;
  grid-template-rows: 0fr;
  opacity: 0.35;
  transition: grid-template-rows 420ms cubic-bezier(0.2, 0.75, 0.25, 1), opacity 300ms ease;
}
.terms-collapse.open { grid-template-rows: 1fr; opacity: 1; }
@media (min-width: 768px) {
  .terms-collapse,
  .terms-collapse.closed { grid-template-rows: 1fr; opacity: 1; }
}
/* Mobile: slide up from bottom */
.quick-buy-enter-active {
  transition: transform 280ms cubic-bezier(0.32, 0.72, 0, 1), opacity 280ms ease-out;
}
.quick-buy-leave-active {
  transition: transform 200ms ease-in, opacity 200ms ease-in;
}
.quick-buy-enter-from {
  transform: translateY(100%);
  opacity: 1;
}
.quick-buy-enter-to {
  transform: translateY(0);
  opacity: 1;
}
.quick-buy-leave-from {
  transform: translateY(0);
  opacity: 1;
}
.quick-buy-leave-to {
  transform: translateY(100%);
  opacity: 1;
}

/* Desktop: centered scale + fade */
@media (min-width: 768px) {
  .quick-buy-enter-active {
    transition: transform 200ms ease-out, opacity 200ms ease-out;
  }
  .quick-buy-leave-active {
    transition: transform 150ms ease-in, opacity 150ms ease-in;
  }
  .quick-buy-enter-from {
    transform: scale(0.96);
    opacity: 0;
  }
  .quick-buy-enter-to {
    transform: scale(1);
    opacity: 1;
  }
  .quick-buy-leave-from {
    transform: scale(1);
    opacity: 1;
  }
  .quick-buy-leave-to {
    transform: scale(0.96);
    opacity: 0;
  }
}
</style>
