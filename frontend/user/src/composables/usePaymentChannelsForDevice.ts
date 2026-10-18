import { computed, type ComputedRef } from 'vue'

export function expectedAlipayInteractionMode(): 'wap' | 'page' {
  if (typeof navigator === 'undefined') return 'page'
  const mobile = /Android|iPhone|iPad|iPod|Mobile/i.test(navigator.userAgent)
    || (navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1)
  return mobile ? 'wap' : 'page'
}

export function isPaymentInteractionModeForDevice(mode: unknown): boolean {
  const normalized = String(mode || '').toLowerCase()
  return (normalized !== 'wap' && normalized !== 'page')
    || normalized === expectedAlipayInteractionMode()
}

/** Keep official Alipay website payments on desktop and WAP payments on mobile. */
export function usePaymentChannelsForDevice(channels: ComputedRef<any[]>) {
  return computed(() => {
    const available = channels.value
    const expectedMode = expectedAlipayInteractionMode()
    const isOfficialAlipay = (channel: any) =>
      String(channel?.provider_type || '').toLowerCase() === 'official'
      && String(channel?.channel_type || '').toLowerCase() === 'alipay'
    const matchesDevice = (channel: any) =>
      isOfficialAlipay(channel)
      && String(channel?.interaction_mode || '').toLowerCase() === expectedMode
    const deviceChannels = available.filter(matchesDevice)

    // Never fall back to the other device's Alipay cashier when the expected
    // channel is missing from a product's allowed channels.
    if (deviceChannels.length) return [deviceChannels[0]]
    return available.filter((channel: any) => !isOfficialAlipay(channel))
  })
}
