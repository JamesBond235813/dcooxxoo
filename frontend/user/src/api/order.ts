import { userApi } from './client'
import type { CreatePaymentPayload } from './types'

type GuestAuthInput = {
    email?: string
    [key: string]: any
}

function encodeGuestAuthorization(email: string): string {
    const bytes = new TextEncoder().encode(email.trim().toLowerCase())
    let binary = ''
    for (const byte of bytes) {
        binary += String.fromCharCode(byte)
    }
    const token = btoa(binary)
        .replace(/\+/g, '-')
        .replace(/\//g, '_')
        .replace(/=+$/g, '')
    return `Guest ${token}`
}

function splitGuestAuth(input: GuestAuthInput = {}) {
    const { email = '', ...payload } = input
    return {
        payload,
        headers: {
            Authorization: encodeGuestAuthorization(String(email)),
        },
    }
}

function withGuestAuth(input: GuestAuthInput, options: Record<string, any> = {}) {
    const { payload, headers } = splitGuestAuth(input)
    return {
        payload,
        options: {
            ...options,
            headers: { ...(options.headers || {}), ...headers },
        },
    }
}

export const userOrderAPI = {
    preview: (data: any) => userApi.post('/orders/preview', data),
    getPaymentChannels: (data: any) => userApi.post('/order/payment-channels', data),
    create: (data: any) => userApi.post('/orders', data),
    createAndPay: (data: any) => userApi.post('/orders/create-and-pay', data),
    list: (params?: any) => userApi.get('/orders', { params }),
    stats: (params?: any) => userApi.get('/orders/stats', { params }),
    detail: (orderNo: string, options?: any) => userApi.get(`/orders/${encodeURIComponent(orderNo)}`, { ...options, params: { ...(options?.params || {}), progress: 1 } }),
    cancel: (orderNo: string) => userApi.post(`/orders/${encodeURIComponent(orderNo)}/cancel`),
    submitSession: (orderNo: string, session: string) =>
        userApi.post(`/orders/${encodeURIComponent(orderNo)}/session`, { session }),
    preflightSession: (orderNo: string, session: string) =>
        userApi.post(`/orders/${encodeURIComponent(orderNo)}/session/preflight`, { session }),
    downloadFulfillment: (orderNo: string) => userApi.get(`/orders/${encodeURIComponent(orderNo)}/fulfillment/download`, { blob: true }),
    acquireSmsNumber: (orderNo: string) =>
        userApi.post(`/orders/${encodeURIComponent(orderNo)}/sms/acquire`),
    checkSmsStatus: (orderNo: string) =>
        userApi.get(`/orders/${encodeURIComponent(orderNo)}/sms/status`),
    retrySmsNumber: (orderNo: string) =>
        userApi.post(`/orders/${encodeURIComponent(orderNo)}/sms/retry`),
}

export const guestOrderAPI = {
	resolveSmsReturn: (paymentID: string, ticket: string) =>
		userApi.get(`/guest/payments/sms-return/${encodeURIComponent(paymentID)}/${encodeURIComponent(ticket)}`, { silentBusinessError: true }),
    preview: (data: any) => userApi.post('/guest/orders/preview', data),
    create: (data: any) => userApi.post('/guest/orders', data),
    createAndPay: (data: any) => userApi.post('/guest/orders/create-and-pay', data),
    list: (params: GuestAuthInput) => {
        const request = withGuestAuth(params)
        return userApi.get('/guest/orders', { ...request.options, params: request.payload })
    },
    detail: (orderNo: string, params: GuestAuthInput, options?: any) => {
        const request = withGuestAuth(params, options)
        return userApi.get(`/guest/orders/${encodeURIComponent(orderNo)}`, { ...request.options, params: { ...request.payload, progress: 1 } })
    },
    submitSession: (orderNo: string, data: GuestAuthInput & { session: string }) => {
        const request = withGuestAuth(data)
        return userApi.post(`/guest/orders/${encodeURIComponent(orderNo)}/session`, request.payload, request.options)
    },
    preflightSession: (orderNo: string, data: GuestAuthInput & { session: string }) => {
        const request = withGuestAuth(data)
        return userApi.post(`/guest/orders/${encodeURIComponent(orderNo)}/session/preflight`, request.payload, request.options)
    },
    downloadFulfillment: (orderNo: string, params: GuestAuthInput) => {
        const request = withGuestAuth(params, { blob: true })
        return userApi.get(`/guest/orders/${encodeURIComponent(orderNo)}/fulfillment/download`, { ...request.options, params: request.payload })
    },
    createPayment: (data: GuestAuthInput) => {
        const request = withGuestAuth(data)
        return userApi.post('/guest/payments', request.payload, request.options)
    },
    capturePayment: (id: number, data: GuestAuthInput) => {
        const request = withGuestAuth(data)
        return userApi.post(`/guest/payments/${id}/capture`, request.payload, request.options)
    },
    latestPayment: (params: GuestAuthInput) => {
        const request = withGuestAuth(params, { silentBusinessError: true })
        return userApi.get('/guest/payments/latest', { ...request.options, params: request.payload })
    },
    submitManualCryptoProof: (id: number, data: GuestAuthInput & { transaction_hash: string }) => {
        const request = withGuestAuth(data)
        return userApi.post(`/guest/payments/${id}/manual-proof`, request.payload, request.options)
    },
    acquireSmsNumber: (orderNo: string, params: GuestAuthInput) => {
        const request = withGuestAuth(params)
        return userApi.post(`/guest/orders/${encodeURIComponent(orderNo)}/sms/acquire`, request.payload, request.options)
    },
    checkSmsStatus: (orderNo: string, params: GuestAuthInput) => {
        const request = withGuestAuth(params)
        return userApi.get(`/guest/orders/${encodeURIComponent(orderNo)}/sms/status`, { ...request.options, params: request.payload })
    },
    retrySmsNumber: (orderNo: string, params: GuestAuthInput) => {
        const request = withGuestAuth(params)
        return userApi.post(`/guest/orders/${encodeURIComponent(orderNo)}/sms/retry`, request.payload, request.options)
    },
}

export const paymentAPI = {
    create: (data: CreatePaymentPayload) => userApi.post('/payments', data),
    capture: (id: number) => userApi.post(`/payments/${id}/capture`),
    latest: (params: any) => userApi.get('/payments/latest', { params, silentBusinessError: true }),
    submitManualCryptoProof: (id: number, transactionHash: string) =>
        userApi.post(`/payments/${id}/manual-proof`, { transaction_hash: transactionHash }),
}
