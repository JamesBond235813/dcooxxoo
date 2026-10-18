import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { guestOrderAPI } from '../api'
import { orderStatusVariant, orderStatusLabel } from '../utils/status'
import { debounceAsync } from '../utils/debounce'
import { amountToCents } from '../utils/money'
import { clearGuestOrderAuth, saveGuestOrderAuth } from '../utils/guestOrderAuth'
import { useRouter } from 'vue-router'

/**
 * 游客订单查询/列表逻辑（classic + vault 共用）。
 */
export function useGuestOrders() {
  const { t } = useI18n()
  const router = useRouter()

  const email = ref('')
  const orderNo = ref('')
  const loading = ref(false)
  const error = ref('')
  const orders = ref<any[]>([])
  const pagination = ref({
    page: 1,
    page_size: 20,
    total: 0,
    total_page: 1,
  })

  const persistAuth = () => {
    const payload = { email: email.value }
    saveGuestOrderAuth(payload)
  }

  const handleSearch = async () => {
    error.value = ''
    if (!email.value) {
      error.value = t('guestOrders.errors.missing')
      return
    }
    persistAuth()
    await debouncedLoadOrders(1)
  }

  const loadOrders = async (page: number) => {
    loading.value = true
    try {
      const response = await guestOrderAPI.list({
        email: email.value,
        order_no: orderNo.value || undefined,
        page,
        page_size: pagination.value.page_size,
      })
      orders.value = response.data.data || []
      pagination.value = response.data.pagination || pagination.value
      if (orderNo.value && orders.value.length === 0) {
        error.value = t('guestOrders.errors.notFound')
      }
    } catch (err: any) {
      orders.value = []
      error.value = err.message || t('guestOrders.errors.searchFailed')
    } finally {
      loading.value = false
    }
  }

  const debouncedLoadOrders = debounceAsync(loadOrders, 300)

  const emptyMessage = computed(() => {
    if (orderNo.value) {
      return t('guestOrders.emptyOrderNo')
    }
    return t('guestOrders.empty')
  })

  const changePage = (page: number) => {
    if (page < 1 || page > pagination.value.total_page) return
    debouncedLoadOrders(page)
  }

  const statusLabel = (status: string) => orderStatusLabel(t, status)
  const statusVariant = (status: string) => orderStatusVariant(status)

  // vault 模板：BadgeTone → vault pill 类名（classic 不使用）
  const statusPillClass = (status?: string) => {
    const map: Record<string, string> = {
      success: 'pill-done',
      warning: 'pill-low',
      danger: 'pill-sale',
      info: 'pill-stock',
      accent: 'pill-sale',
      neutral: 'pill-out',
    }
    return map[orderStatusVariant(status)] || 'pill-out'
  }

  const formatMoney = (amount?: string, currency?: string) => {
    if (amount === null || amount === undefined || amount === '') return '-'
    if (currency === null || currency === undefined || currency === '') {
      return String(amount)
    }
    return `${amount} ${currency}`
  }

  const formatDiscountMoney = (amount?: string, currency?: string) => {
    return hasDiscountAmount(amount) ? `-${formatMoney(amount, currency)}` : formatMoney(amount, currency)
  }

  const hasDiscountAmount = (amount?: string) => {
    if (amount === null || amount === undefined || amount === '') return false
    const valueCents = amountToCents(amount)
    return valueCents !== null && valueCents > 0
  }

  const hasDiscount = (order: any) => {
    if (!order) return false
    return hasDiscountAmount(order.discount_amount) || hasDiscountAmount(order.promotion_discount_amount)
  }

  const formatDate = (raw?: string) => {
    if (!raw) return ''
    const date = new Date(raw)
    if (Number.isNaN(date.getTime())) return raw
    return date.toLocaleString()
  }

  onMounted(() => {
    // 清除旧版本写入 session/local storage 的凭据，查单页不恢复历史账号或密码。
    clearGuestOrderAuth()
  })

  onUnmounted(() => {
    debouncedLoadOrders.cancel()
    // 只在本次查单继续前往订单详情或游客支付时短暂保留进程内凭据。
    const nextRoute = router.currentRoute.value
    const continuesGuestFlow = nextRoute.name === 'guest-order-detail'
      || (nextRoute.name === 'payment' && String(nextRoute.query.guest || '') === '1')
    if (!continuesGuestFlow) clearGuestOrderAuth()
  })

  return {
    email,
    orderNo,
    loading,
    error,
    orders,
    pagination,
    handleSearch,
    emptyMessage,
    changePage,
    statusLabel,
    statusVariant,
    statusPillClass,
    formatMoney,
    formatDiscountMoney,
    hasDiscountAmount,
    hasDiscount,
    formatDate,
  }
}
