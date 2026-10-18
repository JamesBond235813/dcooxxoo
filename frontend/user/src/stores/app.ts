import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'
import { configAPI } from '../api'
import { applyCustomScripts } from '../utils/customScripts'
import { getImageUrl } from '../utils/image'
import { getLocalizedText } from '../utils/resellerSiteConfig'
import { detectLocale, setI18nLocale } from '../i18n'
import { useHead } from '@unhead/vue'

const CONFIG_CACHE_KEY = 'site_config_cache_v2'

const getInitialCachedConfig = () => {
    try {
        const raw = localStorage.getItem(CONFIG_CACHE_KEY)
        return raw ? JSON.parse(raw) : null
    } catch {
        return null
    }
}

export const useAppStore = defineStore('app', () => {
    // 与 vue-i18n 复用同一套语言检测逻辑，避免首次访问时
    // UI 语言（vue-i18n）与商品多语言取值（appStore.locale）不一致
    const locale = ref(detectLocale())
    const config = ref<any>(getInitialCachedConfig())
    const hasInitialLoaded = ref(false)
    const loading = ref(false)
    // 服务器与客户端的时间偏移量（毫秒），serverTime = clientTime + offset
    const serverTimeOffset = ref(0)
    const siteIconSource = computed(() => {
        const siteIcon = String(config.value?.brand?.site_icon || '').trim()
        return siteIcon ? getImageUrl(siteIcon) : '/dj.svg'
    })
    const siteIconHref = ref(siteIconSource.value)
    let faviconRenderId = 0
    watch(siteIconSource, (source) => {
        const renderId = ++faviconRenderId
        if (!source || typeof window === 'undefined') {
            siteIconHref.value = source || '/dj.svg'
            return
        }

        const image = new Image()
        if (new URL(source, window.location.href).origin !== window.location.origin) {
            image.crossOrigin = 'anonymous'
        }
        image.onload = () => {
            if (renderId !== faviconRenderId) return
            try {
                const size = 128
                const canvas = document.createElement('canvas')
                canvas.width = size
                canvas.height = size
                const context = canvas.getContext('2d')
                if (!context) throw new Error('Canvas is unavailable')

                const side = Math.min(image.naturalWidth, image.naturalHeight)
                const left = (image.naturalWidth - side) / 2
                const top = (image.naturalHeight - side) / 2
                context.beginPath()
                context.arc(size / 2, size / 2, size / 2, 0, Math.PI * 2)
                context.clip()
                context.drawImage(image, left, top, side, side, 0, 0, size, size)
                siteIconHref.value = canvas.toDataURL('image/png')
            } catch {
                // Keep the configured icon if the browser cannot render it through canvas.
                siteIconHref.value = source
            }
        }
        image.onerror = () => {
            if (renderId === faviconRenderId) siteIconHref.value = source
        }
        image.src = source
    }, { immediate: true })
    const isResellerTenant = computed(() => {
        return String(config.value?.tenant?.mode || '').trim().toLowerCase() === 'reseller'
    })
    const canAccessResellerConsole = computed(() => !!config.value && !isResellerTenant.value)

    // 设置语言：同时驱动 vue-i18n（内部按需加载语言包）
    const setLocale = (newLocale: string) => {
        locale.value = newLocale
        localStorage.setItem('locale', newLocale)
        void setI18nLocale(newLocale)
    }

    // 全局 head 默认值（标题/描述/keywords/favicon/html lang）。
    // canonical / og / twitter 等页面级别字段交由各页面通过 usePageSeo 接管，
    // 避免与页面级 useHead 冲突或产生重复标签。
    useHead({
        htmlAttrs: { lang: computed(() => locale.value) },
        title: () => {
            const seo = config.value?.seo
            const lang = locale.value
            const localized = getLocalizedText(seo?.title, lang)
            if (localized) return String(localized).trim() || undefined
            const siteName = String(config.value?.brand?.site_name || '').trim()
            return siteName || undefined
        },
        link: () => [{ key: 'favicon', rel: 'icon', href: siteIconHref.value }],
        meta: () => {
            const seo = config.value?.seo
            if (!seo) return []
            const lang = locale.value
            const tags: Array<{ name?: string; property?: string; content: string }> = []
            const keywords = getLocalizedText(seo.keywords, lang)
            if (keywords) {
                tags.push({ name: 'keywords', content: keywords })
            }
            const description = getLocalizedText(seo.description, lang)
            if (description) {
                tags.push({ name: 'description', content: description })
            }
            return tags
        }
    })

    // 更新SEO信息 (向后兼容的方法)
    const applySEO = () => {
        // 由于 useHead 已经是响应式的，这里不再需要显式调用
        // 留空函数以防其他组件出错
    }

    // 加载全局配置
    const loadConfig = async (force = false) => {
        if (hasInitialLoaded.value && !force) {
            applySEO()
            applyCustomScripts(config.value?.scripts)
            return
        }
        if (!config.value) loading.value = true
        try {
            const requestTime = Date.now()
            const response = await configAPI.get()
            config.value = response.data.data
            hasInitialLoaded.value = true
            try {
                localStorage.setItem(CONFIG_CACHE_KEY, JSON.stringify(response.data.data))
                // 清理旧版缓存
                localStorage.removeItem('site_config_cache')
            } catch {
                // ignore localStorage quota errors
            }
            // 计算服务器与客户端的时间偏移量
            if (config.value?.server_time) {
                const responseTime = Date.now()
                const roundTripTime = responseTime - requestTime
                const estimatedServerNow = config.value.server_time + roundTripTime / 2
                serverTimeOffset.value = estimatedServerNow - responseTime
            }
            applySEO()
            applyCustomScripts(config.value?.scripts)
            // Print version to console
            if (config.value?.app_version) {
                console.log(
                    '%c Version %c ' + config.value.app_version + ' %c',
                    'background:#34c759;color:#fff;padding:2px 6px;border-radius:4px 0 0 4px;font-weight:bold;',
                    'background:#1d1d1f;color:#f5f5f7;padding:2px 6px;border-radius:0 4px 4px 0;',
                    'background:transparent;',
                )
            }
        } catch (error) {
            console.error('Failed to load config:', error)
        } finally {
            if (!force) loading.value = false
        }
    }

    // 获取校正后的服务器当前时间（毫秒时间戳）
    const getServerTime = () => Date.now() + serverTimeOffset.value

    // 获取校正后的服务器当前 Date 对象
    const getServerDate = () => new Date(getServerTime())

    return {
        locale,
        config,
        loading,
        serverTimeOffset,
        isResellerTenant,
        canAccessResellerConsole,
        setLocale,
        loadConfig,
        applySEO,
        getServerTime,
        getServerDate,
    }
})
