<template>
  <Teleport to="body">
    <div v-if="visible" class="fixed inset-0 z-[90] flex items-end justify-center bg-slate-950/35 px-3 pb-[max(1rem,env(safe-area-inset-bottom))] lg:hidden" @click.self="dismiss">
      <section role="dialog" aria-modal="true" :aria-label="t('installPrompt.title')" class="w-full max-w-md rounded-3xl border border-emerald-100 bg-card p-5 text-card-foreground shadow-2xl dark:border-emerald-900/50">
        <div class="flex items-start gap-3">
          <img src="/app-icon-192.png" alt="" class="h-14 w-14 shrink-0 rounded-2xl border border-emerald-100 object-cover shadow-sm" />
          <div class="min-w-0 flex-1">
            <p class="text-[11px] font-semibold tracking-wide text-emerald-600 dark:text-emerald-400">{{ t('installPrompt.eyebrow') }}</p>
            <h2 class="mt-1 text-lg font-bold">{{ t('installPrompt.title') }}</h2>
            <p class="mt-1 text-xs leading-relaxed text-muted-foreground">{{ t('installPrompt.description') }}</p>
          </div>
          <button type="button" class="rounded-full p-1 text-muted-foreground hover:bg-muted" :aria-label="t('installPrompt.later')" @click="dismiss"><X class="h-5 w-5" /></button>
        </div>

        <div v-if="showGuide" class="mt-4 rounded-2xl border border-emerald-100 bg-emerald-50/60 p-3 text-sm leading-relaxed text-foreground dark:border-emerald-900/50 dark:bg-emerald-950/20">
          <p class="font-semibold">{{ t('installPrompt.guideTitle') }}</p>
          <p class="mt-1 text-xs text-muted-foreground">{{ guideText }}</p>
        </div>

        <div class="mt-5 flex gap-2">
          <button type="button" class="h-11 flex-1 rounded-xl border bg-background text-sm font-semibold text-muted-foreground" @click="dismiss">{{ t('installPrompt.later') }}</button>
          <button type="button" class="flex h-11 flex-[1.5] items-center justify-center gap-2 rounded-xl bg-emerald-600 px-3 text-sm font-semibold text-white shadow-sm hover:bg-emerald-700" @click="install">
            <Download class="h-4 w-4" />{{ deferredPrompt ? t('installPrompt.install') : (showGuide ? t('installPrompt.gotIt') : t('installPrompt.download')) }}
          </button>
        </div>
      </section>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Download, X } from 'lucide-vue-next'

interface InstallPromptEvent extends Event {
  prompt: () => Promise<void>
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>
}

const { t } = useI18n()
const route = useRoute()
const visible = ref(false)
const showGuide = ref(false)
const deferredPrompt = ref<InstallPromptEvent | null>(null)
const sessionKey = 'dujiao-mobile-install-prompt-session-v1'
const installedKey = 'dujiao-mobile-app-installed-v1'
let showTimer: ReturnType<typeof setTimeout> | undefined

const isIOS = () => /iPad|iPhone|iPod/.test(navigator.userAgent)
const isSafari = () => /Safari/.test(navigator.userAgent) && !/CriOS|FxiOS|EdgiOS|OPiOS/.test(navigator.userAgent)
const isStandalone = () => window.matchMedia('(display-mode: standalone)').matches || Boolean((navigator as Navigator & { standalone?: boolean }).standalone)
const isMobile = () => window.matchMedia('(max-width: 767px) and (pointer: coarse)').matches
const wasShownThisSession = () => { try { return sessionStorage.getItem(sessionKey) === '1' } catch { return false } }
const isMarkedInstalled = () => { try { return localStorage.getItem(installedKey) === '1' } catch { return false } }
const markShownThisSession = () => { try { sessionStorage.setItem(sessionKey, '1') } catch { /* storage may be unavailable */ } }
const markInstalled = () => { try { localStorage.setItem(installedKey, '1') } catch { /* storage may be unavailable */ } }

const guideText = computed(() => isIOS()
  ? t(isSafari() ? 'installPrompt.iosGuide' : 'installPrompt.iosOtherBrowser')
  : t('installPrompt.androidGuide'))

const onBeforeInstallPrompt = (event: Event) => {
  if (!isMobile() || isStandalone() || isMarkedInstalled() || wasShownThisSession()) return
  event.preventDefault()
  deferredPrompt.value = event as InstallPromptEvent
}

const dismiss = () => {
  markShownThisSession()
  visible.value = false
}

const install = async () => {
  if (deferredPrompt.value) {
    const prompt = deferredPrompt.value
    deferredPrompt.value = null
    await prompt.prompt()
    const choice = await prompt.userChoice
    if (choice.outcome === 'accepted') dismiss()
    else showGuide.value = true
    return
  }
  if (showGuide.value) dismiss()
  else showGuide.value = true
}

const onAppInstalled = () => {
  markInstalled()
  dismiss()
}

const scheduleHomePrompt = () => {
  if (showTimer) clearTimeout(showTimer)
  if (route.path !== '/' || !isMobile() || isStandalone() || isMarkedInstalled() || wasShownThisSession()) return
  showTimer = setTimeout(() => {
    if (route.path === '/' && !isStandalone() && !isMarkedInstalled() && !wasShownThisSession()) {
      markShownThisSession()
      visible.value = true
    }
  }, 800)
}

onMounted(() => {
  window.addEventListener('beforeinstallprompt', onBeforeInstallPrompt)
  window.addEventListener('appinstalled', onAppInstalled)
  scheduleHomePrompt()
})

watch(() => route.path, (path) => {
  if (path === '/') scheduleHomePrompt()
  else {
    if (showTimer) clearTimeout(showTimer)
    visible.value = false
  }
})

onBeforeUnmount(() => {
  if (showTimer) clearTimeout(showTimer)
  window.removeEventListener('beforeinstallprompt', onBeforeInstallPrompt)
  window.removeEventListener('appinstalled', onAppInstalled)
})
</script>
