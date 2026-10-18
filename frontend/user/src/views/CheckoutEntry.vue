<template>
  <CheckoutOrderStatus v-if="route.query.order_no" />
  <main v-else-if="route.query.mode === 'buynow'" class="min-h-screen bg-background px-4 pt-28 text-center text-muted-foreground">
    <p v-if="loading">正在读取商品…</p>
    <p v-else-if="error">{{ error }}</p>
    <ProductQuickBuy v-if="quickBuyProduct" :visible="true" :product="quickBuyProduct" :initial-sku-id="buyNowStore.item?.skuId" :initial-quantity="buyNowStore.item?.quantity" @update:visible="onClose" />
  </main>
  <component :is="legacyCheckout" v-else />
</template>
<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getActiveTemplate } from '../templates/registry'
import { useBuyNowStore } from '../stores/buyNow'
import { productAPI } from '../api/product'
import ProductQuickBuy from '../components/ProductQuickBuy.vue'
import CheckoutOrderStatus from './CheckoutOrderStatus.vue'
const route = useRoute()
const router = useRouter()
const buyNowStore = useBuyNowStore()
const quickBuyProduct = ref<any>(null)
const loading = ref(false)
const error = ref('')
const legacyCheckout = computed(() => getActiveTemplate() === 'vault'
  ? defineAsyncComponent(() => import('../templates/vault/Checkout.vue'))
  : defineAsyncComponent(() => import('./Checkout.vue')))
const onClose = (visible: boolean) => { if (!visible) router.push('/products') }
onMounted(async () => {
  if (route.query.mode !== 'buynow' || route.query.order_no) return
  const item = buyNowStore.item
  if (!item?.slug) { error.value = '未找到待购买商品，请重新选购'; return }
  loading.value = true
  try {
    const response = await productAPI.detail(item.slug)
    quickBuyProduct.value = { ...response.data.data, agent_wholesale: Boolean(item.agentWholesale) }
  } catch (cause: any) { error.value = cause?.message || '读取商品失败' }
  finally { loading.value = false }
})
</script>
