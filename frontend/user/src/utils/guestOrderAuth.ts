export interface GuestOrderAuth {
  email: string
}

const GUEST_ORDER_AUTH_KEY = 'guest_order_auth'
const EMPTY_GUEST_ORDER_AUTH: GuestOrderAuth = {
  email: '',
}

// 游客查单凭据只在当前 SPA 生命周期的内存中短暂传递，不写入浏览器存储。
let volatileGuestOrderAuth: GuestOrderAuth | null = null

const normalizeGuestOrderAuth = (auth: Partial<GuestOrderAuth>): GuestOrderAuth => ({
  email: typeof auth.email === 'string' ? auth.email : '',
})

const cloneGuestOrderAuth = (auth: GuestOrderAuth): GuestOrderAuth => ({ ...auth })

const clearStoredGuestOrderAuth = () => {
  if (typeof window === 'undefined') return
  try {
    window.sessionStorage.removeItem(GUEST_ORDER_AUTH_KEY)
  } catch {
    // 浏览器存储不可用时无需额外处理，凭据仅保留在运行时内存。
  }
  try {
    window.localStorage.removeItem(GUEST_ORDER_AUTH_KEY)
  } catch {
    // 清理旧版本可能留下的长期存储凭据。
  }
}

export const loadGuestOrderAuth = (): GuestOrderAuth => {
  clearStoredGuestOrderAuth()
  return cloneGuestOrderAuth(volatileGuestOrderAuth || EMPTY_GUEST_ORDER_AUTH)
}

export const saveGuestOrderAuth = (auth: GuestOrderAuth) => {
  clearStoredGuestOrderAuth()
  if (typeof window === 'undefined') return
  volatileGuestOrderAuth = cloneGuestOrderAuth(normalizeGuestOrderAuth(auth))
}

export const clearGuestOrderAuth = () => {
  volatileGuestOrderAuth = null
  clearStoredGuestOrderAuth()
}
