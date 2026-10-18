<template>
  <section class="rounded-2xl border bg-card p-4 shadow-sm" aria-live="polite">
    <div class="flex items-center justify-between gap-3">
      <h2 class="font-semibold text-foreground">{{ t('fulfillmentProgress.title') }}</h2>
      <span class="text-xs text-muted-foreground">{{ t('fulfillmentProgress.autoRefresh') }}</span>
    </div>
    <ol class="mt-4 grid grid-cols-2 gap-x-3 gap-y-4 sm:grid-cols-5">
      <li v-for="(step, index) in steps" :key="step" class="flex items-start gap-2 text-xs sm:text-sm">
        <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full font-bold"
          :class="index < activeIndex ? 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400' : index === activeIndex ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground'">
          {{ index < activeIndex ? '✓' : index + 1 }}
        </span>
        <span :class="index <= activeIndex ? 'font-medium text-foreground' : 'text-muted-foreground'">{{ step }}</span>
      </li>
    </ol>
    <p class="mt-4 border-t pt-3 text-sm" :class="failed ? 'text-destructive' : 'text-muted-foreground'">
      {{ currentMessage }}
    </p>
    <ol v-if="visibleEvents.length" class="mt-3 space-y-2 border-l-2 border-primary/20 pl-3 text-xs text-muted-foreground">
      <li v-for="(event, index) in visibleEvents" :key="`${index}-${event.stage}-${event.status}`" class="flex flex-wrap justify-between gap-x-3">
        <span>{{ stageName(event.stage || event.status || '') }}</span>
        <time v-if="event.at">{{ formatEventTime(event.at) }}</time>
      </li>
    </ol>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{ order: any; submitting?: boolean }>()
const { t } = useI18n()
const entries = computed<any[]>(() => [props.order, ...(props.order?.children || [])].filter(Boolean))
const submitted = computed(() => entries.value.some(entry => (entry.items || []).some((item: any) => item.requires_session && item.session_submitted)))
const progress = computed(() => entries.value.map(entry => entry.upstream_progress).find(Boolean))
type ProgressEvent = { stage?: string; status?: string; at?: string }
const visibleEvents = computed<ProgressEvent[]>(() => {
  const events = Array.isArray(progress.value?.events) ? progress.value.events as ProgressEvent[] : []
  return events.filter((event, index) => index === 0 || event.stage !== events[index - 1]?.stage || event.status !== events[index - 1]?.status)
})
const stageName = (stage: string) => {
  const known = new Set(['queued', 'cdk_accepted', 'awaiting_card', 'funding_pending', 'dispatching', 'running', 'requires_action', 'pending', 'plus_paid', 'completed', 'delivered', 'declined', 'failed_precharge', 'cancelled', 'review'])
  return known.has(stage) ? t(`fulfillmentProgress.stages.${stage}`) : stage.replace(/_/g, ' ')
}
const formatEventTime = (value: string) => {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '' : date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
const delivered = computed(() => ['delivered', 'completed'].includes(props.order?.status || ''))
const failed = computed(() => ['failed', 'rejected', 'canceled', 'refunded', 'partially_refunded'].includes(progress.value?.status || ''))
const accepted = computed(() => !!progress.value && !['pending', 'failed', 'rejected'].includes(progress.value.status))
const activeIndex = computed(() => delivered.value ? 4 : failed.value ? 3 : accepted.value ? 3 : submitted.value ? 2 : 1)
const steps = computed(() => [
  t('fulfillmentProgress.paid'),
  t('fulfillmentProgress.account'),
  t('fulfillmentProgress.submitted'),
  t('fulfillmentProgress.processing'),
  t('fulfillmentProgress.completed'),
])
const currentMessage = computed(() => {
  if (props.submitting) return t('fulfillmentProgress.submitting')
  if (delivered.value) return t('fulfillmentProgress.completedMessage')
  if (failed.value) return t('fulfillmentProgress.failedMessage')
  if (!submitted.value) return t('fulfillmentProgress.awaitingSession')
  if (!progress.value || progress.value.status === 'pending') return t('fulfillmentProgress.queued')
  if (progress.value.stage) return t('fulfillmentProgress.stage', { stage: stageName(progress.value.stage) })
  return t('fulfillmentProgress.upstreamProcessing')
})
</script>
