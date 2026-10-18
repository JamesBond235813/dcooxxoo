import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { guestOrderAPI } from '../api'
import { debounceAsync } from '../utils/debounce'
import { clearGuestOrderAuth, loadGuestOrderAuth, saveGuestOrderAuth } from '../utils/guestOrderAuth'
import { resolveGuestOrderDetailViewState } from '../utils/guestOrderDetailState'
import { useOrderDisplayHelpers } from './useOrderDisplayHelpers'

/**
 * 游客订单详情逻辑（classic + vault 共用）。
 */
export function useGuestOrderDetail() {
  const route = useRoute()
  const router = useRouter()
  const { t } = useI18n()

  const loading = ref(true)
  const order = ref<any>(null)
  const authError = ref('')
  const auth = ref({
    email: '',
  })
  const fulfillmentDownloading = ref(false)

  const helpers = useOrderDisplayHelpers(order)

  const handleDownloadFulfillment = async (orderNo: string) => {
    if (fulfillmentDownloading.value) return
    fulfillmentDownloading.value = true
    try {
      const res = await guestOrderAPI.downloadFulfillment(orderNo, {
        email: auth.value.email,
      })
      const blob = new Blob([res.data], { type: 'text/plain; charset=utf-8' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `fulfillment-${orderNo}.txt`
      a.click()
      URL.revokeObjectURL(url)
    } catch {} finally {
      fulfillmentDownloading.value = false
    }
  }

  const loadSavedAuth = () => {
    auth.value = loadGuestOrderAuth()
  }

  const hasAuth = computed(() => Boolean(auth.value.email))
  const showAuthForm = computed(() => !hasAuth.value || authError.value !== '')
  const viewState = computed(() => resolveGuestOrderDetailViewState({
    loading: loading.value,
    order: order.value,
    showAuthForm: showAuthForm.value,
  }))

  const loadOrder = async () => {
    loading.value = true
    try {
      if (!hasAuth.value) {
        order.value = null
        authError.value = t('guestOrderDetail.authRequired')
        return
      }
      const response = await guestOrderAPI.detail(String(route.params.order_no || '').trim(), {
        email: auth.value.email,
      })
      order.value = response.data.data
      authError.value = ''
    } catch (error) {
      order.value = null
      authError.value = t('guestOrderDetail.authInvalid')
    } finally {
      loading.value = false
    }
  }

  const debouncedLoadOrder = debounceAsync(loadOrder, 300)

  const persistAuth = () => {
    saveGuestOrderAuth({
      email: auth.value.email,
    })
  }

  const handleAuthSubmit = async () => {
    authError.value = ''
    if (!hasAuth.value) {
      authError.value = t('guestOrderDetail.authRequired')
      return
    }
    persistAuth()
    loading.value = true
    await debouncedLoadOrder()
  }

  const clearAuth = () => {
    clearGuestOrderAuth()
    auth.value = { email: '' }
    order.value = null
    authError.value = t('guestOrderDetail.authRequired')
  }

  onMounted(() => {
    if (!route.params.order_no) {
      router.push('/guest/orders')
      return
    }
    loadSavedAuth()
    const returnValue = (key: string) => {
      const match = Object.entries(route.query).find(([name]) => name.replace(/^(amp;)+/, '') === key)
      return String(Array.isArray(match?.[1]) ? match?.[1][0] || '' : match?.[1] || '').trim()
    }
    const paymentID = returnValue('return_payment')
    const ticket = returnValue('return_ticket')
    if (!paymentID || !ticket) {
      void loadOrder()
      return
    }
    void (async () => {
      try {
        for (let attempt = 0; attempt < 30; attempt++) {
          const response = await guestOrderAPI.resolveSmsReturn(paymentID, ticket)
          const result = response.data.data
          if (result?.ready) {
            if (String(result.order_no) !== String(route.params.order_no)) break
            auth.value = { email: String(result.email || '') }
            saveGuestOrderAuth(auth.value)
            await router.replace({ path: route.path })
            await loadOrder()
            return
          }
          await new Promise(resolve => setTimeout(resolve, 2000))
        }
      } catch {
        // 凭据失效时仍可手动填写订单邮箱查询。
      }
      await router.replace({ path: route.path })
      await loadOrder()
    })()
  })

  onUnmounted(() => {
    debouncedLoadOrder.cancel()
    const nextRoute = router.currentRoute.value
    const continuesGuestPayment = nextRoute.name === 'payment' && String(nextRoute.query.guest || '') === '1'
    if (!continuesGuestPayment) clearGuestOrderAuth()
  })

  return {
    loading,
    order,
    authError,
    auth,
    showAuthForm,
    viewState,
    handleAuthSubmit,
    clearAuth,
    fulfillmentDownloading,
    handleDownloadFulfillment,
    ...helpers,
  }
}
