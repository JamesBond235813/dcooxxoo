import type { RouteLocationNormalizedLoaded, RouteLocationNormalized } from 'vue-router'
import { affiliateAPI } from './api'
import { useAppStore } from './stores/app'

const AFFILIATE_STORAGE_KEY = 'dj_affiliate_attribution'
const AFFILIATE_VISITOR_KEY = 'dj_affiliate_visitor_key'
const AGENT_SESSION_KEY = 'dj_agent_code'
const AFFILIATE_TTL_MS = 30 * 24 * 60 * 60 * 1000

let agentContextInitialized = false

type AffiliateAttribution = {
    code: string
    expires_at: number
}

const normalizeAffiliateCode = (raw: string) => {
    const value = String(raw || '').trim().toUpperCase()
    if (!/^[A-Z0-9]{4,32}$/.test(value)) {
        return ''
    }
    return value
}

const readAttribution = (): AffiliateAttribution | null => {
    if (typeof window === 'undefined') return null
    const raw = localStorage.getItem(AFFILIATE_STORAGE_KEY)
    if (!raw) return null
    try {
        const parsed = JSON.parse(raw) as Partial<AffiliateAttribution>
        const code = normalizeAffiliateCode(String(parsed.code || ''))
        const expiresAt = Number(parsed.expires_at || 0)
        const appStore = useAppStore()
        if (!code || !Number.isFinite(expiresAt) || expiresAt <= appStore.getServerTime()) {
            localStorage.removeItem(AFFILIATE_STORAGE_KEY)
            return null
        }
        return {
            code,
            expires_at: expiresAt,
        }
    } catch {
        localStorage.removeItem(AFFILIATE_STORAGE_KEY)
        return null
    }
}

const writeAttribution = (code: string) => {
    if (typeof window === 'undefined') return
    const payload: AffiliateAttribution = {
        code,
        expires_at: useAppStore().getServerTime() + AFFILIATE_TTL_MS,
    }
    localStorage.setItem(AFFILIATE_STORAGE_KEY, JSON.stringify(payload))
}

const ensureVisitorKey = () => {
    if (typeof window === 'undefined') return ''
    const existing = String(localStorage.getItem(AFFILIATE_VISITOR_KEY) || '').trim()
    if (existing) return existing
    const next = `${Date.now().toString(36)}${Math.random().toString(36).slice(2, 12)}`
    localStorage.setItem(AFFILIATE_VISITOR_KEY, next)
    return next
}

const resolveQueryCode = (route: RouteLocationNormalizedLoaded | RouteLocationNormalized) => {
    const raw = Array.isArray(route.query.aff) ? route.query.aff[0] : route.query.aff
    if (!raw) return ''
    return normalizeAffiliateCode(String(raw))
}

export const getAffiliateCode = () => {
    const attribution = readAttribution()
    return attribution?.code || ''
}

export const getAffiliateVisitorKey = () => {
    if (typeof window === 'undefined') return ''
    const existing = String(localStorage.getItem(AFFILIATE_VISITOR_KEY) || '').trim()
    if (existing) return existing
    return ensureVisitorKey()
}

export const clearAffiliateCode = () => {
    if (typeof window === 'undefined') return
    localStorage.removeItem(AFFILIATE_STORAGE_KEY)
}

const clearLegacyAgentContext = () => {
    if (typeof window === 'undefined') return
    // Older builds stored the price context in shared localStorage and cookies.
    // Clear those once per page load so stale values cannot affect this tab.
    localStorage.removeItem('dj_agent_code')
    if (typeof document !== 'undefined') {
        document.cookie = 'dj_agent=; path=/; max-age=0; SameSite=Lax'
        document.cookie = 'agent_code=; path=/; max-age=0; SameSite=Lax'
    }
}

export const captureAffiliateFromRoute = async (route: RouteLocationNormalizedLoaded | RouteLocationNormalized) => {
    if (typeof window === 'undefined') return

    // Store reseller price context per tab. On the first route of a fresh page
    // load, an untagged URL must start at main-site pricing even if this tab was
    // duplicated from another tab or still has state from an older build.
    const rawAgent = route.query.agent || route.query.reseller || route.query.ref || route.query.aff
    const agentVal = Array.isArray(rawAgent) ? rawAgent[0] : rawAgent
    const hasAgentParam = agentVal !== undefined && agentVal !== null
    if (hasAgentParam) {
        const agentCode = normalizeAffiliateCode(String(agentVal))
        if (agentCode) {
            sessionStorage.setItem(AGENT_SESSION_KEY, agentCode)
        } else {
            sessionStorage.removeItem(AGENT_SESSION_KEY)
        }
    } else if (!agentContextInitialized) {
        sessionStorage.removeItem(AGENT_SESSION_KEY)
    }
    if (!agentContextInitialized) clearLegacyAgentContext()
    agentContextInitialized = true

    const code = resolveQueryCode(route)
    if (!code) return

    writeAttribution(code)

    try {
        await affiliateAPI.trackClick({
            affiliate_code: code,
            visitor_key: ensureVisitorKey(),
            landing_path: route.fullPath || route.path || '/',
            referrer: typeof document !== 'undefined' ? document.referrer : '',
        })
    } catch {
    }
}
