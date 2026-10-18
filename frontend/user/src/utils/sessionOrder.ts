const sessionEligibleStatuses = new Set(['paid', 'fulfilling', 'partially_delivered'])

const sessionEntries = (order: any) => [order, ...(Array.isArray(order?.children) ? order.children : [])]

export function hasSessionOrder(order: any): boolean {
  if (!order) return false
  return sessionEntries(order).some((entry: any) => (Array.isArray(entry?.items) ? entry.items : []).some((item: any) => item.requires_session))
}

export function hasPendingSession(order: any): boolean {
  if (!order || !sessionEligibleStatuses.has(String(order.status || ''))) return false
  return sessionEntries(order).some((entry: any) => (Array.isArray(entry.items) ? entry.items : []).some((item: any) =>
    item.requires_session && !item.session_submitted,
  ))
}

// 已付款的 Session 商品尚未完成履约时始终提供回到进度/提交页的入口；
// session_submitted 用来决定页面展示提交表单还是当前处理进度。
export function hasSessionFollowUp(order: any): boolean {
  return !!order && sessionEligibleStatuses.has(String(order.status || '')) && hasSessionOrder(order)
}
