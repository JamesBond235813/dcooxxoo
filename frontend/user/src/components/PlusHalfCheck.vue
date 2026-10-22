<template>
  <div>
    <button type="button" class="rounded-full bg-emerald-600 px-5 py-2.5 text-sm font-bold text-white hover:bg-emerald-700" @click="open = true">{{ t('plusHalf.checkButton') }}</button>
    <div v-if="open" class="fixed inset-0 z-[100] flex items-center justify-center bg-black/60 p-4" role="dialog" aria-modal="true" :aria-label="t('plusHalf.title')">
      <div class="w-full max-w-lg rounded-2xl bg-background p-6 text-foreground shadow-xl">
        <div class="flex items-start justify-between gap-4">
          <h2 class="text-lg font-bold">{{ t('plusHalf.title') }}</h2>
          <button type="button" :aria-label="t('plusHalf.close')" @click="close">×</button>
        </div>
        <p class="mt-2 text-sm text-muted-foreground">{{ t('plusHalf.description') }}</p>
        <textarea v-model="session" rows="6" autocomplete="off" spellcheck="false" class="mt-4 w-full rounded-lg border bg-background p-3 text-sm" :placeholder="t('plusHalf.sessionPlaceholder')" :disabled="loading" />
        <p v-if="error" class="mt-3 text-sm text-destructive" role="alert">{{ error }}</p>
        <p v-if="checked && !eligible" class="mt-3 text-sm text-amber-600" role="status">{{ t('plusHalf.notEligible') }}</p>
        <div v-if="eligible" class="mt-3 rounded-lg border border-emerald-400/40 bg-emerald-500/10 p-3 text-sm" role="status">
          <p>{{ t('plusHalf.eligible', { email: accountEmail }) }}</p>
          <p v-if="price" class="mt-1 font-semibold">{{ t('plusHalf.price', { price }) }}</p>
        </div>
        <div class="mt-5 flex flex-wrap gap-3">
          <button type="button" class="rounded-lg bg-emerald-600 px-5 py-2 text-sm font-bold text-white disabled:opacity-50" :disabled="loading || !session.trim()" @click="check">{{ loading ? t('plusHalf.checking') : t('plusHalf.checkNow') }}</button>
          <button v-if="eligible" type="button" class="rounded-lg border px-5 py-2 text-sm font-bold disabled:opacity-50" :disabled="loading || !offer" @click="buy">{{ t('plusHalf.buyNow') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { guestOrderAPI, productAPI } from '../api'
import { useBuyNowStore } from '../stores/buyNow'

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
const offer = ref<any>(null)
const offerSku = ref<any>(null)
const price = ref('')

watch(session, () => { eligible.value = false; checked.value = false; offer.value = null; offerSku.value = null; price.value = '' })

const close = () => { open.value = false; session.value = ''; error.value = ''; eligible.value = false }

const check = async () => {
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
