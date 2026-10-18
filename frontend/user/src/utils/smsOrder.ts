const smsEligibleStatuses = new Set(['paid', 'fulfilling', 'partially_delivered', 'delivered', 'completed'])

const orderEntries = (order: any) => [order, ...(Array.isArray(order?.children) ? order.children : [])]

export function isSmsActivationOrder(order: any): boolean {
  if (!order) return false
  return orderEntries(order).some((entry: any) =>
    (Array.isArray(entry?.items) ? entry.items : []).some((item: any) => {
      if (item.requires_sms_activation) return true
      if (item.manual_form_schema_snapshot?.type === 'sms_activation' || item.manual_form_schema_snapshot?.service === 'openai') return true
      const title = JSON.stringify(item.title || '')
      return title.includes('接码') || title.includes('收码') || title.toLowerCase().includes('sms')
    })
  )
}

export function isSmsActiveStatus(order: any): boolean {
  return !!order && smsEligibleStatuses.has(String(order.status || ''))
}
