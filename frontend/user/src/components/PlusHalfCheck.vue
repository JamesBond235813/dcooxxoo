<template>
  <div class="inline-block">
    <button
      type="button"
      :class="[
        'breath-btn inline-flex items-center gap-1.5 sm:gap-2 rounded-full bg-emerald-600 px-3 py-1.5 sm:px-4 sm:py-2 text-xs sm:text-sm font-bold text-white hover:bg-emerald-700 active:scale-95 transition-all cursor-pointer select-none',
        $attrs.class
      ]"
      @click="openModal"
    >
      <Sparkles class="h-3.5 w-3.5 sm:h-4 sm:w-4 text-emerald-200 animate-pulse shrink-0" />
      <span>{{ t('plusHalf.checkButton') }}</span>
    </button>

    <Teleport to="body">
      <Transition
        enter-active-class="transition-opacity duration-200 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-opacity duration-150 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div
          v-if="open"
          class="fixed inset-0 z-[9999] flex items-start sm:items-center justify-center p-3 sm:p-6 overflow-y-auto pt-5 sm:pt-6 pb-20 sm:pb-6"
        >
          <!-- Backdrop with blur -->
          <div class="fixed inset-0 bg-black/65 backdrop-blur-xs" aria-hidden="true" @click="close" />

          <!-- Modal Card -->
          <div
            :class="[
              'relative z-10 w-full max-w-lg rounded-2xl bg-card p-5 sm:p-7 text-card-foreground shadow-2xl border border-border/80 transition-transform duration-200',
              isKeyboardActive ? '-translate-y-3 sm:translate-y-0' : ''
            ]"
            role="dialog"
            aria-modal="true"
            :aria-label="t('plusHalf.title')"
          >
            <!-- Header -->
            <div class="flex items-start justify-between gap-4 pb-3 border-b border-border/50">
              <div class="flex items-center gap-3">
                <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-emerald-100 dark:bg-emerald-950/80 text-emerald-700 dark:text-emerald-300">
                  <Sparkles class="h-5 w-5" />
                </div>
                <div>
                  <h2 class="text-base sm:text-lg font-bold leading-tight text-foreground">{{ t('plusHalf.title') }}</h2>
                  <p class="text-xs text-muted-foreground mt-0.5 leading-relaxed">{{ t('plusHalf.description') }}</p>
                </div>
              </div>
              <button
                type="button"
                class="grid h-8 w-8 shrink-0 place-items-center rounded-full text-muted-foreground hover:bg-accent hover:text-foreground transition"
                :aria-label="t('plusHalf.close')"
                @click="close"
              >
                <X class="h-4 w-4" />
              </button>
            </div>

            <!-- Body -->
            <div class="mt-4 space-y-3">
              <div class="flex items-center justify-between text-xs text-muted-foreground">
                <span class="font-medium text-foreground">ChatGPT Session JSON</span>
                <span class="inline-flex items-center gap-1 text-[11px] text-emerald-600 dark:text-emerald-400">
                  <ShieldCheck class="h-3.5 w-3.5" /> 官方实时预检 · 安全脱敏
                </span>
              </div>

              <textarea
                v-model="session"
                rows="4"
                autocomplete="off"
                spellcheck="false"
                class="w-full rounded-xl border border-input bg-background/50 p-3 text-xs sm:text-sm font-mono placeholder:text-muted-foreground/60 focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 focus:outline-none transition resize-none"
                :placeholder="t('plusHalf.sessionPlaceholder')"
                :disabled="loading"
                @focus="isKeyboardActive = true"
                @blur="isKeyboardActive = false"
              />

              <!-- Session Tutorial Guide -->
              <div class="rounded-xl border border-emerald-500/25 bg-emerald-500/5 dark:bg-emerald-950/20 p-3 text-xs space-y-1.5">
                <div class="flex items-center justify-between font-semibold text-foreground">
                  <span class="inline-flex items-center gap-1.5 text-emerald-800 dark:text-emerald-300">
                    <HelpCircle class="h-3.5 w-3.5 text-emerald-600 dark:text-emerald-400 shrink-0" />
                    如何获取 ChatGPT Session？
                  </span>
                  <button
                    type="button"
                    class="inline-flex items-center gap-1 rounded px-2 py-0.5 text-[11px] font-medium text-emerald-700 dark:text-emerald-300 bg-emerald-100/70 dark:bg-emerald-900/40 hover:bg-emerald-200/60 transition cursor-pointer"
                    @click="copySessionUrl"
                  >
                    <Check v-if="copied" class="h-3 w-3 text-emerald-600" />
                    <Copy v-else class="h-3 w-3" />
                    <span>{{ copied ? '已复制网址' : '复制 Session 网址' }}</span>
                  </button>
                </div>
                <ol class="list-decimal list-inside space-y-1 text-muted-foreground text-[11px] leading-relaxed">
                  <li>浏览器登录官方 <span class="font-mono text-foreground font-medium">chatgpt.com</span></li>
                  <li>同一浏览器打开 <span class="font-mono text-emerald-700 dark:text-emerald-300 font-medium">chatgpt.com/api/auth/session</span></li>
                  <li>全选页面内容并复制，粘贴到上方输入框即可</li>
                </ol>
              </div>

              <!-- Error Alert -->
              <div v-if="error" class="flex items-start gap-2.5 rounded-xl border border-destructive/40 bg-destructive/10 p-3 text-xs text-destructive" role="alert">
                <AlertCircle class="h-4 w-4 shrink-0 mt-0.5" />
                <div class="leading-relaxed">{{ error }}</div>
              </div>

              <!-- Ineligible Alert -->
              <div v-if="checked && !eligible && !error" class="flex items-start gap-2.5 rounded-xl border border-amber-500/40 bg-amber-500/10 p-3.5 text-xs text-amber-700 dark:text-amber-300" role="status">
                <AlertCircle class="h-4 w-4 shrink-0 mt-0.5" />
                <div class="space-y-1 leading-relaxed">
                  <div class="font-semibold">{{ t('plusHalf.notEligible') }}</div>
                  <div class="text-[11px] opacity-85">{{ t('plusHalf.notEligibleDesc') }}</div>
                </div>
              </div>

              <!-- Eligible Card -->
              <div v-if="eligible" class="rounded-xl border border-emerald-500/40 bg-emerald-500/10 p-4 text-xs sm:text-sm text-foreground space-y-2.5" role="status">
                <div class="flex items-center gap-2 text-emerald-700 dark:text-emerald-300 font-bold">
                  <CheckCircle2 class="h-4 w-4 shrink-0" />
                  <span class="truncate">{{ t('plusHalf.eligible', { email: accountEmail }) }}</span>
                </div>
                <div class="flex items-baseline justify-between pt-2 border-t border-emerald-500/20">
                  <span class="text-xs text-muted-foreground">特惠专享价</span>
                  <span class="text-lg font-extrabold text-emerald-600 dark:text-emerald-400">¥{{ price || '66.00' }} <span class="text-xs font-normal text-muted-foreground">/ 月</span></span>
                </div>
              </div>
            </div>

            <!-- Footer Buttons -->
            <div class="mt-5 flex items-center justify-end gap-3 pt-3 border-t border-border/50">
              <button
                type="button"
                class="rounded-xl border border-border px-4 py-2 text-sm font-medium hover:bg-accent transition"
                @click="close"
              >
                {{ t('plusHalf.close') }}
              </button>

              <button
                v-if="!eligible"
                type="button"
                class="inline-flex items-center justify-center gap-2 rounded-xl bg-emerald-600 px-5 py-2 text-sm font-bold text-white shadow-sm hover:bg-emerald-700 disabled:opacity-50 transition active:scale-95"
                :disabled="loading || !session.trim()"
                @click="check"
              >
                <Loader2 v-if="loading" class="h-4 w-4 animate-spin" />
                <span>{{ loading ? t('plusHalf.checking') : t('plusHalf.checkNow') }}</span>
              </button>

              <button
                v-if="eligible"
                type="button"
                class="inline-flex items-center justify-center gap-2 rounded-xl bg-emerald-600 px-6 py-2 text-sm font-extrabold text-white shadow-md hover:bg-emerald-700 hover:shadow-lg disabled:opacity-50 transition active:scale-95"
                :disabled="loading || !offer"
                @click="buy"
              >
                <span>{{ t('plusHalf.buyNow') }} (¥{{ price || '66.00' }})</span>
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { Sparkles, X, Loader2, CheckCircle2, AlertCircle, ShieldCheck, HelpCircle, Copy, Check } from 'lucide-vue-next'
import { guestOrderAPI, productAPI } from '../api'
import { useBuyNowStore } from '../stores/buyNow'

defineOptions({
  inheritAttrs: false,
})

const { t } = useI18n()
const router = useRouter()
const buyNowStore = useBuyNowStore()
const open = ref(false)
const session = ref('')
const loading = ref(false)
const error = ref('')
const checked = ref(false)
const eligible = ref(false)
const accountEmail = ref('')
const isKeyboardActive = ref(false)
const copied = ref(false)

const copySessionUrl = async () => {
  try {
    await navigator.clipboard.writeText('https://chatgpt.com/api/auth/session')
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    // fallback
    const input = document.createElement('input')
    input.value = 'https://chatgpt.com/api/auth/session'
    document.body.appendChild(input)
    input.select()
    document.execCommand('copy')
    document.body.removeChild(input)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  }
}
const offer = ref<any>(null)
const offerSku = ref<any>(null)
const price = ref('')

watch(session, () => {
  eligible.value = false
  checked.value = false
  offer.value = null
  offerSku.value = null
  price.value = ''
  error.value = ''
})

const openModal = () => {
  open.value = true
}

const close = () => {
  open.value = false
  session.value = ''
  error.value = ''
  eligible.value = false
  checked.value = false
}

const onKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape' && open.value) {
    close()
  }
}

onMounted(() => {
  window.addEventListener('keydown', onKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown)
})

const check = async () => {
  if (!session.value.trim()) return
  loading.value = true
  error.value = ''
  eligible.value = false
  checked.value = false
  try {
    const response = await guestOrderAPI.checkPlusHalfEligibility(session.value.trim())
    const result = response.data.data
    checked.value = true
    if (result?.eligible !== true) return
    const detail = await productAPI.detail('chatgpt-plus-half')
    const product = detail.data.data
    const sku = (product?.skus || []).find((row: any) => row.id === result.sku_id && row.sku_code === 'plus_half' && row.is_active)
    if (!product || !sku || product.id !== result.product_id) throw new Error(t('plusHalf.unavailable'))
    accountEmail.value = String(result.account_email || '')
    offer.value = product
    offerSku.value = sku
    price.value = String(sku.price_amount || product.price_amount || '')
    eligible.value = true
  } catch (cause: any) {
    error.value = cause?.message || t('plusHalf.checkFailed')
  } finally {
    loading.value = false
  }
}

const buy = () => {
  if (!eligible.value || !offer.value || !offerSku.value || !session.value.trim()) return
  buyNowStore.setItem({
    productId: offer.value.id,
    skuId: offerSku.value.id,
    skuCode: 'plus_half',
    slug: offer.value.slug,
    title: offer.value.title,
    priceAmount: price.value,
    quantity: 1,
    purchaseType: offer.value.purchase_type,
    fulfillmentType: offer.value.fulfillment_type,
    manualFormSchema: offer.value.manual_form_schema || {},
    paymentChannelIds: offer.value.payment_channel_ids,
  })
  buyNowStore.setPlusHalfSession(session.value.trim())
  close()
  void router.push({ path: '/checkout', query: { mode: 'buynow' } })
}
</script>

<style scoped>
@keyframes breath {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.55), 0 2px 5px 0 rgba(0, 0, 0, 0.12);
    transform: scale(1);
  }
  50% {
    box-shadow: 0 0 0 7px rgba(16, 185, 129, 0), 0 4px 14px 0 rgba(16, 185, 129, 0.35);
    transform: scale(1.025);
  }
}

.breath-btn {
  animation: breath 2.4s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}
</style>
