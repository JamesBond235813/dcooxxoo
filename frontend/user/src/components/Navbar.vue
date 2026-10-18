<template>
  <nav
    class="fixed top-0 left-0 right-0 z-50 bg-card/80 border-b backdrop-blur-md transition-all"
    :class="scrolled ? 'py-2 shadow-lg' : 'py-4'"
    :style="{ transitionDuration: 'var(--ui-duration-normal)' }">
    <div class="container mx-auto px-4 flex items-center justify-between gap-4">
      <!-- Logo -->
      <div class="relative flex min-w-0 items-center gap-2 lg:hidden">
        <button v-if="userAuthStore.isAuthenticated" type="button" class="shrink-0 rounded-full focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary" :aria-label="t('navbar.logout')" :aria-expanded="showAvatarMenu" @click="showAvatarMenu = !showAvatarMenu">
          <img v-if="brandLogo" :src="brandLogo" :alt="brandSiteName" class="h-8 w-8 rounded-full object-cover" />
          <span v-else class="flex h-8 w-8 items-center justify-center rounded-full bg-primary/10 text-xs font-bold text-primary">{{ brandSiteName.slice(0, 1) || 'AI' }}</span>
        </button>
        <router-link v-else to="/" class="shrink-0" :aria-label="t('nav.home')">
          <img v-if="brandLogo" :src="brandLogo" :alt="brandSiteName" class="h-8 w-8 rounded-full object-cover" />
          <span v-else class="flex h-8 w-8 items-center justify-center rounded-full bg-primary/10 text-xs font-bold text-primary">{{ brandSiteName.slice(0, 1) || 'AI' }}</span>
        </router-link>
        <span class="truncate text-sm font-semibold">{{ brandSiteName }}</span>
        <div v-if="showAvatarMenu && userAuthStore.isAuthenticated" class="absolute left-0 top-full z-[80] mt-3 min-w-32 rounded-xl border bg-card p-1.5 shadow-xl">
          <router-link v-if="route.path.startsWith('/agent')" to="/agent/withdraw" class="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-sm text-foreground hover:bg-accent" @click="showAvatarMenu = false"><Wallet class="h-4 w-4" />{{ t('agentWithdraw.menu') }}</router-link>
          <button type="button" class="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-sm text-destructive hover:bg-destructive/10" @click="handleAvatarLogout"><LogOut class="h-4 w-4" />{{ t('navbar.logout') }}</button>
        </div>
      </div>
      <router-link to="/" class="theme-wordmark group relative hidden lg:flex" :title="brandSiteName">
        <div
          v-if="brandLogo"
          class="w-7 h-7 sm:w-8 sm:h-8 rounded-full shrink-0 overflow-hidden flex items-center justify-center"
        >
          <img
            :src="brandLogo"
            :alt="brandSiteName"
            class="w-full h-full object-cover"
          />
        </div>
        <span v-if="brandSiteName" class="theme-wordmark-text">{{ brandSiteName }}</span>
      </router-link>

      <!-- Right Side Actions (Home, Agent Center, Guest Orders, User/Logout, Theme Switcher) -->
      <div class="flex items-center shrink-0 space-x-1 lg:space-x-2 ml-auto">
        <!-- 首页 -->
        <router-link
          to="/"
          :class="buttonVariants({ variant: 'ghost', size: 'sm' })"
          class="!hidden lg:!inline-flex gap-1.5 text-muted-foreground hover:text-foreground whitespace-nowrap"
          active-class="!text-primary !bg-primary/10">
          <Home class="w-4 h-4 shrink-0 opacity-70" />
          <span>首页</span>
        </router-link>

        <!-- 代理中心 -->
        <router-link
          to="/agent"
          :class="buttonVariants({ variant: 'ghost', size: 'sm' })"
          class="!hidden lg:!inline-flex gap-1.5 text-muted-foreground hover:text-foreground whitespace-nowrap"
          active-class="!text-primary !bg-primary/10">
          <Shield class="w-4 h-4 shrink-0 opacity-70" />
          <span>代理中心</span>
        </router-link>

        <!-- 游客查单 -->
        <router-link
          to="/guest/orders"
          :class="buttonVariants({ variant: 'ghost', size: 'sm' })"
          class="!hidden lg:!inline-flex gap-1.5 text-muted-foreground hover:text-foreground whitespace-nowrap"
          active-class="!text-primary !bg-primary/10">
          <ClipboardList class="w-4 h-4 shrink-0 opacity-70" />
          <span>{{ t('navbar.guestOrders') }}</span>
        </router-link>

        <Button v-if="userAuthStore.isAuthenticated" variant="ghost" size="sm"
          class="!hidden lg:!inline-flex gap-1.5 whitespace-nowrap text-destructive hover:text-destructive hover:bg-destructive/10"
          @click="userAuthStore.logout()">
          <LogOut class="w-4 h-4 shrink-0 opacity-70" />
          <span>{{ t('navbar.logout') }}</span>
        </Button>

        <!-- Theme Switcher -->
        <Button variant="ghost" size="icon" class="text-muted-foreground" @click="toggleTheme">
          <Sun v-if="theme === 'dark'" class="w-4 h-4" />
          <Moon v-else class="w-4 h-4" />
        </Button>

      </div>
    </div>

  </nav>

  <div v-if="showAvatarMenu" class="fixed inset-0 z-40 lg:hidden" @click="showAvatarMenu = false"></div>

</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useAppStore } from '../stores/app'
import { useCartStore } from '../stores/cart'
import { useUserAuthStore } from '../stores/userAuth'
import { useTheme } from '../utils/theme'
import { getImageUrl } from '../utils/image'
import {
  Home, Shield, Sun, Moon, ClipboardList, LogOut, Wallet,
} from 'lucide-vue-next'
import { Button, buttonVariants } from '@/components/ui/button'

const { t } = useI18n()
const route = useRoute()
const appStore = useAppStore()
const cartStore = useCartStore()
const userAuthStore = useUserAuthStore()
const { theme, toggleTheme } = useTheme()
const showAvatarMenu = ref(false)
const scrolled = ref(false)
const cartBounce = ref(false)

const cartCount = computed(() => cartStore.totalItems)

const brandSiteName = computed(() => {
  return String(appStore.config?.brand?.site_name || '').trim()
})

const brandLogo = computed(() => {
  const raw = String(appStore.config?.brand?.site_logo || '').trim()
  return raw ? getImageUrl(raw) : ''
})

const handleAvatarLogout = () => {
  showAvatarMenu.value = false
  userAuthStore.logout(route.path.startsWith('/agent') ? '/agent' : '/')
}

const handleScroll = () => {
  scrolled.value = window.scrollY > 20
}

// Cart badge bounce animation on count change
watch(cartCount, (newVal, oldVal) => {
  if (newVal > oldVal) {
    cartBounce.value = true
    setTimeout(() => { cartBounce.value = false }, 400)
  }
})
watch(() => route.fullPath, () => { showAvatarMenu.value = false })
watch(() => userAuthStore.isAuthenticated, () => { showAvatarMenu.value = false })

onMounted(() => {
  window.addEventListener('scroll', handleScroll, { passive: true })
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>

<style scoped>
.scrollbar-hide {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
.scrollbar-hide::-webkit-scrollbar {
  display: none;
}
</style>
