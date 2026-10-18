<template>
  <nav class="lg:hidden fixed bottom-0 left-0 right-0 z-40 bg-card/90 backdrop-blur-xl border-t theme-safe-bottom">
    <div class="flex items-stretch h-14">
      <router-link
        v-for="item in navItems"
        :key="item.path"
        :to="item.path"
        class="flex-1 flex flex-col items-center justify-center gap-0.5 text-xs transition-colors min-h-[44px]"
        :class="isActive(item.path) ? 'text-primary font-semibold' : 'text-muted-foreground'"
      >
        <!-- Home -->
        <Home v-if="item.icon === 'home'" class="w-5 h-5" />
        <!-- Agent -->
        <Shield v-else-if="item.icon === 'agent'" class="w-5 h-5" />
        <!-- Guest orders -->
        <ClipboardList v-else-if="item.icon === 'orders'" class="w-5 h-5" />
        <span>{{ item.label }}</span>
      </router-link>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Home, Shield, ClipboardList } from 'lucide-vue-next'

const route = useRoute()
const { t } = useI18n()
const navItems = computed(() => {
  return [
    { path: '/', icon: 'home', label: t('bottomNav.home') },
    { path: '/agent', icon: 'agent', label: '代理中心' },
    { path: '/guest/orders', icon: 'orders', label: t('navbar.guestOrders') },
  ]
})

const isActive = (path: string) => {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}
</script>
