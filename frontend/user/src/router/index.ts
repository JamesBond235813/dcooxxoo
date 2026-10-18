import { createRouter, createWebHistory } from 'vue-router'
import { useUserAuthStore } from '../stores/userAuth'
import { useAppStore } from '../stores/app'
import { useTelegramMiniAppStore } from '../stores/telegramMiniApp'
import { captureAffiliateFromRoute } from '../utils/affiliate'
import { templateView } from '../templates/registry'

type RouteComponentLoader = () => Promise<unknown>

const homeViewLoader: RouteComponentLoader = () => import('../views/Home.vue')
const productsViewLoader: RouteComponentLoader = () => import('../views/Products.vue')
const productDetailViewLoader: RouteComponentLoader = () => import('../views/ProductDetail.vue')
const cartViewLoader: RouteComponentLoader = () => import('../views/Cart.vue')
const checkoutViewLoader: RouteComponentLoader = () => import('../views/CheckoutEntry.vue')
const paymentViewLoader: RouteComponentLoader = () => import('../views/Payment.vue')
const blogViewLoader: RouteComponentLoader = () => import('../views/Blog.vue')
const noticeViewLoader: RouteComponentLoader = () => import('../views/Notice.vue')
const resellerLayoutLoader: RouteComponentLoader = () => import('../views/reseller/ResellerConsoleLayout.vue')
const colorDemoLoader: RouteComponentLoader = () => import('../views/ColorDemo.vue')

const routeWarmupLoaders: RouteComponentLoader[] = [
    productsViewLoader,
    productDetailViewLoader,
    cartViewLoader,
    checkoutViewLoader,
    paymentViewLoader,
    blogViewLoader,
    noticeViewLoader,
]

let hasScheduledRouteWarmup = false

const shouldWarmupRoutes = () => {
    if (typeof window === 'undefined' || typeof navigator === 'undefined') {
        return false
    }

    const connection = (navigator as Navigator & {
        connection?: {
            saveData?: boolean
            effectiveType?: string
        }
    }).connection

    if (connection?.saveData) {
        return false
    }

    return connection?.effectiveType !== 'slow-2g' && connection?.effectiveType !== '2g'
}

const scheduleIdleTask = (task: () => void) => {
    if (typeof window === 'undefined') {
        return
    }

    if ('requestIdleCallback' in window && typeof window.requestIdleCallback === 'function') {
        window.requestIdleCallback(task, { timeout: 1500 })
        return
    }

    window.setTimeout(task, 600)
}

const runRouteWarmupQueue = (loaders: RouteComponentLoader[]) => {
    const nextLoader = loaders.shift()
    if (!nextLoader || typeof window === 'undefined') {
        return
    }

    void nextLoader()
        .catch(() => undefined)
        .finally(() => {
            window.setTimeout(() => {
                scheduleIdleTask(() => runRouteWarmupQueue(loaders))
            }, 400)
        })
}

export const warmupCommonRoutes = () => {
    if (hasScheduledRouteWarmup || !shouldWarmupRoutes()) {
        return
    }

    hasScheduledRouteWarmup = true

    const startWarmup = () => {
        window.setTimeout(() => {
            scheduleIdleTask(() => runRouteWarmupQueue([...routeWarmupLoaders]))
        }, 1200)
    }

    if (document.readyState === 'complete') {
        startWarmup()
        return
    }

    window.addEventListener('load', startWarmup, { once: true })
}

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    scrollBehavior(to, _from, savedPosition) {
        if (savedPosition) {
            return savedPosition
        }
        if (to.hash) {
            return { el: to.hash, top: 80 }
        }
        return { top: 0 }
    },
    routes: [
        {
            path: '/',
            name: 'home',
            component: templateView('Home', homeViewLoader),
        },
        {
            path: '/color-demo',
            name: 'color-demo',
            component: colorDemoLoader,
        },
        {
            path: '/agent',
            name: 'agent-center',
            component: () => import('../views/AgentCenter.vue'),
        },
        {
            path: '/agent/withdraw',
            name: 'agent-withdraw',
            component: () => import('../views/AgentCenter.vue'),
        },
        {
            path: '/products',
            name: 'products',
            component: () => {
                const appStore = useAppStore()
                return appStore.config?.template_mode === 'list'
                    ? templateView('Home', homeViewLoader)()
                    : templateView('Products', productsViewLoader)()
            },
        },
        {
            path: '/categories/:slug',
            name: 'category-products',
            component: () => {
                const appStore = useAppStore()
                return appStore.config?.template_mode === 'list'
                    ? templateView('Home', homeViewLoader)()
                    : templateView('Products', productsViewLoader)()
            },
        },
        {
            path: '/products/:slug',
            name: 'product-detail',
            component: templateView('ProductDetail', productDetailViewLoader),
        },
        {
            path: '/cart',
            name: 'cart',
            component: templateView('Cart', cartViewLoader),
        },
        {
            path: '/checkout',
            name: 'checkout',
            component: checkoutViewLoader,
        },
        {
            path: '/pay',
            name: 'payment',
            component: templateView('Payment', paymentViewLoader),
        },
        { path: '/me/orders', redirect: '/guest/orders' },
        { path: '/me/reseller', redirect: '/agent' },
        { path: '/me', redirect: '/guest/orders' },
        { path: '/me/:pathMatch(.*)*', redirect: '/' },
        {
            path: '/reseller',
            component: resellerLayoutLoader,
            meta: { requiresUserAuth: true, resellerConsole: true },
            children: [
                { path: '', name: 'reseller-dashboard', component: () => import('../views/reseller/ResellerDashboard.vue') },
                { path: 'apply', name: 'reseller-apply', component: () => import('../views/reseller/ResellerApply.vue') },
                { path: 'domains', name: 'reseller-domains', component: () => import('../views/reseller/ResellerDomains.vue') },
                { path: 'site', name: 'reseller-site', component: () => import('../views/reseller/ResellerSiteConfig.vue') },
                { path: 'products', name: 'reseller-products', component: () => import('../views/reseller/ResellerProducts.vue') },
                { path: 'orders', name: 'reseller-orders', component: () => import('../views/reseller/ResellerOrders.vue') },
                { path: 'orders/:order_no', name: 'reseller-order-detail', component: () => import('../views/reseller/ResellerOrderDetail.vue') },
                { path: 'finance', name: 'reseller-finance', component: () => import('../views/reseller/ResellerFinance.vue') },
                { path: 'ledger', name: 'reseller-ledger', component: () => import('../views/reseller/ResellerLedger.vue') },
                { path: 'withdraws', name: 'reseller-withdraws', component: () => import('../views/reseller/ResellerWithdraws.vue') },
            ],
        },
        {
            path: '/orders/:order_no',
            name: 'order-detail',
            component: templateView('OrderDetail', () => import('../views/OrderDetail.vue')),
            meta: { requiresUserAuth: true },
        },
        {
            path: '/orders/:order_no/session',
            name: 'order-session',
            component: () => import('../views/OrderDetail.vue'),
            meta: { requiresUserAuth: true },
        },
        {
            path: '/recharge-orders/:recharge_no',
            name: 'recharge-order-detail',
            redirect: '/guest/orders',
        },
        {
            path: '/guest/orders',
            name: 'guest-orders',
            component: templateView('GuestOrders', () => import('../views/GuestOrders.vue')),
        },
        {
            path: '/guest/orders/:order_no',
            name: 'guest-order-detail',
            component: templateView('GuestOrderDetail', () => import('../views/GuestOrderDetail.vue')),
        },
        {
            path: '/guest/orders/:order_no/session',
            name: 'guest-order-session',
            component: () => import('../views/GuestOrderDetail.vue'),
        },
        {
            path: '/blog',
            name: 'blog',
            component: templateView('Blog', blogViewLoader),
        },
        {
            path: '/blog/:slug',
            name: 'blog-detail',
            component: templateView('BlogDetail', () => import('../views/BlogDetail.vue')),
        },
        {
            path: '/notice',
            name: 'notice',
            component: templateView('Notice', noticeViewLoader),
        },
        {
            path: '/about',
            name: 'about',
            component: templateView('About', () => import('../views/About.vue')),
        },
        {
            path: '/terms',
            name: 'terms',
            component: templateView('Legal', () => import('../views/Legal.vue')),
            props: { type: 'terms' }
        },
        {
            path: '/privacy',
            name: 'privacy',
            component: templateView('Legal', () => import('../views/Legal.vue')),
            props: { type: 'privacy' }
        },
        { path: '/auth', redirect: '/agent' },
        { path: '/auth/login', redirect: '/agent' },
        { path: '/auth/register', redirect: '/agent' },
        { path: '/auth/forgot', redirect: '/agent' },
        { path: '/auth/:pathMatch(.*)*', redirect: '/agent' },
        {
            path: '/:pathMatch(.*)*',
            name: 'not-found',
            component: templateView('NotFound', () => import('../views/NotFound.vue')),
        },
    ],
})

// Navigation Guard
router.beforeEach(async (to, _from, next) => {
    // 域名为 aiadmin 时直接跳转到独立管理后台 /admin
    if (typeof window !== 'undefined') {
        const host = window.location.hostname.toLowerCase()
        if (host.includes('aiadmin') && to.path === '/') {
            window.location.href = '/admin'
            return
        }
    }

    const userAuthStore = useUserAuthStore()
    const appStore = useAppStore()
    void captureAffiliateFromRoute(to)

    // Revalidate config in background or wait if not loaded
    if (!appStore.config) {
        await appStore.loadConfig()
    } else {
        void appStore.loadConfig()
    }

    if (to.meta.requiresUserAuth) {
        if (!userAuthStore.isAuthenticated) {
            next('/agent')
        } else if (to.meta.resellerConsole && !appStore.canAccessResellerConsole) {
            next('/agent')
        } else {
            next()
        }
    }
    else {
        next()
    }
})

// Update SEO on route change
router.afterEach(() => {
    const appStore = useAppStore()
    const telegramMiniAppStore = useTelegramMiniAppStore()
    appStore.applySEO()
    telegramMiniAppStore.syncRouteBackButton(router.currentRoute.value.path, () => {
        if (window.history.length > 1) {
            router.back()
            return
        }
        void router.push('/')
    })
})

export default router
