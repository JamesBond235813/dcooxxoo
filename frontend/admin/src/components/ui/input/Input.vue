<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { computed, ref, useAttrs, watch } from "vue"
import { CalendarDays, ChevronLeft, ChevronRight } from "lucide-vue-next"
import { PopoverContent, PopoverPortal, PopoverRoot, PopoverTrigger } from "reka-ui"
import { cn } from "@/lib/utils"

defineOptions({ inheritAttrs: false })

const props = defineProps<{
  defaultValue?: string | number
  modelValue?: string | number
  class?: HTMLAttributes["class"]
}>()

const emits = defineEmits<{
  (e: "update:modelValue", payload: string | number): void
  (e: "change", payload: string): void
  (e: "input", payload: Event): void
}>()

const attrs = useAttrs()
const kind = computed(() => attrs.type === "date" || attrs.type === "datetime-local" ? attrs.type : "")
const isCalendar = computed(() => Boolean(kind.value))
const open = ref(false)
const localValue = ref<string | number>(props.modelValue ?? props.defaultValue ?? "")
const current = computed(() => String(props.modelValue ?? localValue.value ?? ""))
const cursor = ref(monthStart(parseDate(current.value) ?? new Date()))
const secondMonth = computed(() => new Date(cursor.value.getFullYear(), cursor.value.getMonth() + 1, 1))

watch(() => props.modelValue, (value) => {
  if (value !== undefined) localValue.value = value
})

watch(open, (value) => {
  if (value) cursor.value = monthStart(parseDate(current.value) ?? new Date())
})

function parseDate(value: string): Date | null {
  if (!value) return null
  const match = value.match(/^(\d{4})-(\d{2})-(\d{2})/)
  if (!match) return null
  const date = new Date(Number(match[1]), Number(match[2]) - 1, Number(match[3]))
  return Number.isNaN(date.getTime()) ? null : date
}

function monthStart(date: Date) { return new Date(date.getFullYear(), date.getMonth(), 1) }
function moveMonth(delta: number) { cursor.value = new Date(cursor.value.getFullYear(), cursor.value.getMonth() + delta, 1) }
function monthTitle(date: Date) { return `${date.getFullYear()} 年 ${date.getMonth() + 1} 月` }
function dateKey(date: Date) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`
}
function monthDays(date: Date) {
  const first = monthStart(date)
  const start = new Date(first)
  start.setDate(1 - first.getDay())
  return Array.from({ length: 42 }, (_, index) => {
    const day = new Date(start)
    day.setDate(start.getDate() + index)
    return { date: day, key: dateKey(day), inMonth: day.getMonth() === date.getMonth() }
  })
}
function isSelected(date: Date) { return current.value.slice(0, 10) === dateKey(date) }
function isDisabled(date: Date) {
  const key = dateKey(date)
  return (typeof attrs.min === "string" && key < attrs.min.slice(0, 10)) || (typeof attrs.max === "string" && key > attrs.max.slice(0, 10))
}
function selectDate(date: Date) {
  const key = dateKey(date)
  let value = key
  if (kind.value === "datetime-local") {
    const time = current.value.match(/T(\d{2}:\d{2})/)?.[1] ?? "00:00"
    value = `${key}T${time}`
  }
  setValue(value)
  if (kind.value === "date") open.value = false
}
function setValue(value: string) {
  localValue.value = value
  emits("update:modelValue", value)
  emits("change", value)
}
function handleInput(event: Event) {
  const value = (event.target as HTMLInputElement).value
  localValue.value = value
  emits("update:modelValue", value)
  emits("input", event)
}
function selectTime(event: Event) {
  const time = (event.target as HTMLInputElement).value
  const day = parseDate(current.value) ?? new Date()
  const value = `${dateKey(day)}T${time}`
  setValue(value)
  open.value = false
}
function displayValue() {
  if (!current.value) return String(attrs.placeholder ?? "选择日期")
  const date = parseDate(current.value)
  if (!date) return current.value
  const day = `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日`
  return kind.value === "datetime-local" ? `${day} ${current.value.match(/T(\d{2}:\d{2})/)?.[1] ?? "00:00"}` : day
}
const weekdays = ["日", "一", "二", "三", "四", "五", "六"]
</script>

<template>
  <input
    v-if="!isCalendar"
    :value="localValue"
    v-bind="attrs"
    :class="cn('block h-9 w-full min-w-0 max-w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50', props.class)"
    @input="handleInput"
    @change="emits('change', ($event.target as HTMLInputElement).value)"
  >
  <PopoverRoot v-else v-model:open="open">
    <div :class="cn('relative w-full min-w-0', props.class)">
      <PopoverTrigger as-child>
        <button
          type="button"
          :disabled="Boolean(attrs.disabled)"
          :aria-label="String(attrs['aria-label'] ?? attrs.placeholder ?? '选择日期')"
          class="flex h-9 w-full min-w-0 items-center gap-2 rounded-md border border-input bg-background px-3 text-left text-sm shadow-sm transition-colors hover:bg-accent/40 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
          :class="props.class"
        >
          <CalendarDays class="h-4 w-4 shrink-0 text-muted-foreground" />
          <span class="truncate" :class="!current && 'text-muted-foreground'">{{ displayValue() }}</span>
        </button>
      </PopoverTrigger>
      <input v-if="attrs.required" type="text" :value="current" required tabindex="-1" aria-hidden="true" class="pointer-events-none absolute h-px w-px opacity-0">
      <PopoverPortal>
        <PopoverContent align="start" :side-offset="8" class="z-[100] w-auto rounded-lg border border-border bg-popover p-0 text-popover-foreground shadow-xl outline-none data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95">
          <div class="grid grid-cols-1 divide-y divide-border sm:grid-cols-2 sm:divide-x sm:divide-y-0">
            <section v-for="(month, monthIndex) in [cursor, secondMonth]" :key="monthIndex" class="w-[280px] p-3 sm:w-[324px] sm:p-4">
              <header class="mb-2 flex h-8 items-center justify-between">
                <button type="button" class="inline-flex h-8 w-8 items-center justify-center rounded-md hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring" aria-label="上一个月" @click="moveMonth(-1)">
                  <ChevronLeft class="h-4 w-4" />
                </button>
                <div class="text-sm font-medium">{{ monthTitle(month) }}</div>
                <button type="button" class="inline-flex h-8 w-8 items-center justify-center rounded-md hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring" aria-label="下一个月" @click="moveMonth(1)">
                  <ChevronRight class="h-4 w-4" />
                </button>
              </header>
              <div class="grid grid-cols-7 text-center text-xs text-muted-foreground">
                <span v-for="day in weekdays" :key="day" class="flex h-8 items-center justify-center">{{ day }}</span>
              </div>
              <div class="grid grid-cols-7 text-center">
                <button
                  v-for="cell in monthDays(month)"
                  :key="cell.key"
                  type="button"
                  :disabled="isDisabled(cell.date)"
                  class="mx-auto flex h-9 w-9 items-center justify-center rounded-full text-[13px] transition-colors hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-30"
                  :class="[!cell.inMonth && 'text-muted-foreground/50', isSelected(cell.date) && 'bg-primary font-semibold text-primary-foreground hover:bg-primary/90']"
                  @click="selectDate(cell.date)"
                >{{ cell.date.getDate() }}</button>
              </div>
            </section>
          </div>
          <div v-if="kind === 'datetime-local'" class="flex items-center justify-between gap-3 border-t border-border px-4 py-3 text-sm">
            <span class="text-muted-foreground">时间</span>
            <input type="time" :value="current.match(/T(\d{2}:\d{2})/)?.[1] ?? '00:00'" class="h-8 rounded-md border border-input bg-background px-2 text-sm focus:outline-none focus:ring-1 focus:ring-ring" @change="selectTime">
          </div>
        </PopoverContent>
      </PopoverPortal>
    </div>
  </PopoverRoot>
</template>
