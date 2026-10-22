<template>
  <div class="min-h-screen bg-background text-foreground pt-24 pb-16">
    <div class="container mx-auto px-4">
      <BreadcrumbNav
        class="mb-6"
        :class="isSessionPage ? 'hidden md:block' : ''"
        :items="[
          { label: t('nav.home'), to: '/' },
          { label: orderListLabel, to: orderListPath },
          { label: t('orderDetail.title') },
        ]"
      />
      <BreadcrumbNav v-if="isSessionPage" class="mb-6 md:hidden" :items="sessionBreadcrumbItems" />
      <div class="mb-8" :class="isSessionPage ? 'hidden md:block' : ''">
        <h1 class="text-3xl font-black text-foreground mb-2">{{ t(isSessionPage ? 'orderDetail.sessionPageTitle' : 'orderDetail.title') }}</h1>
        <p class="text-muted-foreground text-sm">{{ t(isSessionPage ? 'orderDetail.sessionPageSubtitle' : 'orderDetail.subtitle') }}</p>
      </div>

      <!-- Loading Skeleton -->
      <div v-if="loading" class="space-y-6">
        <div class="rounded-2xl border bg-card shadow-sm p-6 space-y-4">
          <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
            <div class="space-y-2">
              <div class="h-3 w-20 rounded theme-skeleton"></div>
              <div class="h-5 w-56 rounded theme-skeleton"></div>
            </div>
            <div class="h-7 w-24 rounded-full theme-skeleton"></div>
          </div>
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4 pt-4 border-t">
            <div v-for="i in 4" :key="i" class="space-y-2">
              <div class="h-3 w-16 rounded theme-skeleton"></div>
              <div class="h-4 w-24 rounded theme-skeleton"></div>
            </div>
          </div>
        </div>
        <div class="rounded-2xl border bg-card shadow-sm p-6 space-y-4">
          <div class="h-5 w-28 rounded theme-skeleton"></div>
          <div v-for="i in 2" :key="i" class="flex gap-4">
            <div class="w-16 h-16 rounded-xl theme-skeleton shrink-0"></div>
            <div class="flex-1 space-y-2">
              <div class="h-4 w-3/4 rounded theme-skeleton"></div>
              <div class="h-3 w-1/2 rounded theme-skeleton"></div>
            </div>
            <div class="h-5 w-20 rounded theme-skeleton"></div>
          </div>
        </div>
      </div>

      <EmptyState
        v-else-if="!order"
        icon="alert"
        :title="t('orderDetail.notFound')"
        :action-label="t('errorBoundary.retry')"
        @action="debouncedLoadOrder()"
      />

      <div v-else class="space-y-6">
        <SmsActivationConsole v-if="isSmsOrder && isSmsEligible" :order="order" @refresh="debouncedLoadOrder" />
        <SessionProgress v-if="isSessionOrder" :order="order" :submitting="sessionSubmitting" class="mb-5" />
        <details :open="!(isSmsOrder && isSmsEligible)" class="space-y-4">
          <summary v-if="isSmsOrder" class="cursor-pointer rounded-xl border bg-card p-3 text-sm font-medium">{{ t('smsConsole.orderDetails') }}</summary>
        <section v-if="needsSession" class="rounded-2xl border border-blue-200 bg-blue-50/70 p-6 shadow-sm dark:border-blue-900 dark:bg-blue-950/30">
          <h2 class="text-lg font-bold text-foreground">提交 Session 完成充值</h2>
          <div class="mt-3 rounded-xl bg-background/80 p-4 text-sm text-muted-foreground">
            <p class="font-semibold text-foreground">获取 Session</p>
            <ol class="mt-2 list-decimal space-y-1 pl-5">
              <li>在电脑或手机浏览器登录 <span class="font-medium text-foreground">chatgpt.com</span>。</li>
              <li>保持登录，在同一浏览器打开 <a class="break-all text-primary underline" href="https://chatgpt.com/api/auth/session" target="_blank" rel="noopener noreferrer">https://chatgpt.com/api/auth/session</a>。</li>
              <li>复制页面返回的完整 JSON，粘贴到下方输入框；操作期间保持 ChatGPT 登录。</li>
            </ol>
          </div>
          <textarea v-model="sessionValue" rows="5" autocomplete="off" spellcheck="false"
            class="mt-4 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm text-foreground"
            placeholder="粘贴完整 Session JSON" :disabled="sessionSubmitting || sessionPreflighting" />
          <Button class="mt-3" :disabled="sessionSubmitting || sessionPreflighting || !sessionJSONValid" @click="preflightSession">
            {{ t(sessionPreflighting ? 'orderDetail.sessionChecking' : 'orderDetail.sessionCheckNow') }}
          </Button>
          <div v-if="sessionReviewOpen" class="mt-4 rounded-xl border bg-background p-4">
            <h3 class="font-semibold text-foreground">请确认充值账号</h3>
            <SessionAccountSummary v-if="sessionAccountSummary" :account="sessionAccountSummary" />
            <p class="mt-1 text-xs text-muted-foreground">确认后系统会提交 Session 并开始自动处理。请确认这是你要充值的账号。</p>
            <div class="mt-3 flex flex-wrap items-center gap-3">
              <label class="flex items-center gap-2 text-sm"><input v-model="accountConfirmed" type="checkbox" class="h-4 w-4 accent-primary" />{{ t('orderDetail.sessionAccountConfirmed') }}</label>
              <Button :disabled="sessionSubmitting || sessionPreflighting || !sessionAccountSummary || !accountConfirmed" @click="submitSession">{{ t(sessionSubmitting ? 'orderDetail.sessionStarting' : 'orderDetail.sessionStart') }}</Button>
              <Button variant="outline" :disabled="sessionSubmitting || sessionPreflighting" @click="sessionReviewOpen = false">返回修改</Button>
            </div>
          </div>
        </section>
        <section v-if="isSessionPage" class="rounded-2xl border bg-card p-4 shadow-sm">
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <div class="text-xs text-muted-foreground">{{ t('orders.orderNo') }}</div>
              <div class="mt-1 break-all text-sm font-semibold">{{ order.order_no }}</div>
              <div class="mt-1 text-xs text-muted-foreground md:hidden">{{ t('orderDetail.createdAtLabel') }}：{{ formatDate(order.created_at) }}</div>
              <div class="mt-1 hidden text-xs text-muted-foreground md:block">{{ t('orderDetail.createdAtLabel') }}：{{ formatDate(order.created_at) }}</div>
            </div>
            <Badge :variant="statusVariant(order.status)" size="sm">{{ statusLabel(order.status) }}</Badge>
          </div>
          <div v-if="sessionSummaryItems.length" class="mt-3 border-t pt-3 text-sm">
            <div class="text-xs text-muted-foreground">充值项目</div>
            <div v-for="(item, index) in sessionSummaryItems" :key="item.id || index" class="mt-1 flex justify-between gap-3">
              <span class="min-w-0 truncate">{{ getLocalizedText(item.title) }}</span>
              <span class="shrink-0 text-muted-foreground">× {{ item.quantity }}</span>
            </div>
          </div>
          <div class="mt-3 flex items-center justify-between border-t pt-3 text-sm">
            <span class="text-muted-foreground">实付金额</span>
            <span class="font-bold tabular-nums">{{ formatMoney(order.total_amount, order.currency) }}</span>
          </div>
          <div v-if="order.expires_at" class="mt-3 flex items-center justify-between gap-3 border-t pt-3 text-sm">
            <span class="shrink-0 text-muted-foreground">{{ t('orderDetail.expiresAtLabel') }}</span>
            <time class="text-right font-medium tabular-nums">{{ formatDate(order.expires_at) }}</time>
          </div>
        </section>
        <template v-else-if="!isAgentOrderView">
        <div class="rounded-2xl border bg-card shadow-sm p-6">
          <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
            <div>
              <div class="text-xs uppercase tracking-wider text-muted-foreground">{{ t('orders.orderNo') }}</div>
              <div class="text-sm font-semibold text-foreground mt-1">{{ order.order_no }}</div>
                <div class="text-xs text-muted-foreground mt-2">{{ t('orderDetail.createdAtLabel') }}：{{ formatDate(order.created_at) }}</div>
            </div>
            <div class="flex flex-col items-start md:items-end gap-2">
              <div class="text-xs uppercase tracking-wider text-muted-foreground">{{ t('orderDetail.amountTotal') }}</div>
              <div class="text-lg font-bold text-foreground">{{ formatMoney(order.total_amount,
                order.currency) }}</div>
            </div>
            <div class="flex items-center gap-3">
              <Badge :variant="statusVariant(order.status)" size="sm">
                {{ statusLabel(order.status) }}
              </Badge>
              <Button v-if="order.status === 'pending_payment'" as-child size="sm">
                <router-link :to="`/pay?order_no=${order.order_no}`">{{ t('orderDetail.payNow') }}</router-link>
              </Button>
              <Button v-if="order.status === 'pending_payment'" variant="destructive" size="sm" @click="cancelOrder">
                {{ t('orderDetail.cancel') }}
              </Button>
            </div>
          </div>
        </div>

        <div v-if="order.children && order.children.length > 0"
          class="rounded-2xl border bg-card shadow-sm p-6">
          <h2 class="text-lg font-bold mb-4">{{ t('orderDetail.childOrdersTitle') }}</h2>
          <div class="space-y-4">
            <div v-for="child in order.children" :key="child.id"
              class="border rounded-2xl p-4">
              <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-3">
                <div>
                  <div class="text-sm text-muted-foreground">{{ t('orderDetail.childOrderNo') }}：{{ child.order_no }}</div>
                  <div class="text-xs text-muted-foreground mt-1">{{ t('orderDetail.childOrderAmount') }}：{{
                    formatMoney(child.total_amount, child.currency || order.currency) }}</div>
                </div>
                <Badge :variant="statusVariant(resolvedChildStatus(child))" size="sm">
                  {{ statusLabel(resolvedChildStatus(child)) }}
                </Badge>
              </div>
              <div class="mt-4">
                <h3 class="text-sm font-semibold text-foreground mb-3">{{ t('orderDetail.childItemsTitle')
                  }}</h3>
                <div v-if="child.items && child.items.length" class="space-y-3">
                  <div v-for="(item, cidx) in child.items" :key="cidx"
                    class="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3 sm:gap-4 border-b border-gray-100 pb-3 text-sm text-muted-foreground dark:border-white/5">
                    <div class="flex min-w-0 items-start gap-3">
                      <div class="h-14 w-14 shrink-0 overflow-hidden rounded-xl border border-gray-200 bg-white transition-all duration-200 hover:-translate-y-0.5 hover:shadow-sm dark:border-white/10 dark:bg-black/30 sm:h-16 sm:w-16">
                        <SmartImage
                          :src="orderItemImage(item)"
                          :alt="getLocalizedText(item.title)"
                          img-class="h-full w-full object-cover"
                        />
                      </div>
                      <div class="min-w-0">
                        <div class="text-foreground font-medium">{{ getLocalizedText(item.title) }}</div>
                        <div class="text-xs text-muted-foreground">{{ t('orderDetail.quantityLabel') }}：{{ item.quantity }}</div>
                        <div v-if="orderItemSkuText(item)" class="text-xs text-muted-foreground mt-1">{{ t('orderDetail.itemSkuLabel') }}：{{ orderItemSkuText(item) }}</div>
                        <div class="text-xs text-muted-foreground mt-1">
                          {{ t('orderDetail.itemFulfillmentLabel') }}：{{ fulfillmentTypeLabelText(item.fulfillment_type) }}
                        </div>
                        <div v-if="item.tags && item.tags.length" class="mt-2 flex flex-wrap gap-2">
                          <Badge v-for="(tag, index) in item.tags" :key="index" variant="neutral" size="sm" class="rounded-full">{{ tag }}</Badge>
                        </div>
                        <div v-if="manualSubmissionRows(item.manual_form_submission, item.manual_form_schema_snapshot).length"
                          class="mt-3 rounded-xl border border-gray-200 bg-white p-3 text-xs text-gray-600 dark:border-white/10 dark:bg-black/30 dark:text-gray-300">
                          <div class="mb-2 font-semibold text-muted-foreground">{{ t('orderDetail.manualSubmissionTitle') }}</div>
                          <div v-for="row in manualSubmissionRows(item.manual_form_submission, item.manual_form_schema_snapshot)" :key="row.key" class="mb-1 last:mb-0">
                            <span class="text-foreground">{{ row.label }}</span>：{{ row.value }}
                          </div>
                        </div>
                      </div>
                    </div>
                    <div class="shrink-0 pl-[4.25rem] sm:pl-0 text-left sm:text-right text-sm text-muted-foreground space-y-1">
                      <div>{{ t('orderDetail.unitPriceLabel') }}：{{ formatMoney(item.original_unit_price, order.currency) }}
                      </div>
                      <div>{{ t('orderDetail.totalPriceLabel') }}：{{ formatMoney(item.original_total_price, order.currency) }}
                      </div>
                      <div v-if="hasDiscountAmount(item.coupon_discount_amount)">
                        {{ t('orderDetail.couponDiscountLabel') }}：{{ formatDiscountMoney(item.coupon_discount_amount,
                        order.currency) }}
                      </div>
                      <div v-if="hasDiscountAmount(item.promotion_discount_amount)">
                        {{ t('orderDetail.promotionDiscountLabel') }}：{{ formatDiscountMoney(item.promotion_discount_amount,
                        order.currency) }}
                      </div>
                      <div v-if="hasDiscountAmount(item.member_discount_amount)">
                        {{ t('orderDetail.memberDiscountLabel') }}：{{ formatDiscountMoney(item.member_discount_amount,
                        order.currency) }}
                      </div>
                      <div v-if="hasItemDiscount(item)" class="font-medium text-rose-600 dark:text-rose-300">
                        {{ t('orderDetail.itemDiscountTotalLabel') }}：{{ formatItemDiscountTotal(item, order.currency) }}
                      </div>
                      <div class="font-medium text-foreground">
                        {{ t('orderDetail.itemPaidAmountLabel') }}：{{ formatItemPaidAmount(item, order.currency) }}
                      </div>
                    </div>
                  </div>
                </div>
                <div v-else class="text-sm text-muted-foreground">{{ t('orderDetail.noItems') }}</div>
              </div>
              <div class="mt-4">
                <div class="flex items-center justify-between mb-3">
                  <h3 class="text-sm font-semibold text-foreground">{{
                    t('orderDetail.childFulfillmentTitle') }}</h3>
                  <button v-if="child.fulfillment?.status === 'delivered'"
                    class="inline-flex items-center gap-1 text-xs font-medium px-3 py-1.5 rounded-lg transition-colors shadow-sm"
                    :class="fulfillmentCopied ? 'bg-emerald-600 text-white' : 'bg-blue-600 text-white hover:bg-blue-700 active:bg-blue-800'"
                    @click="handleCopyFulfillment(child.fulfillment)">
                    <Copy v-if="!fulfillmentCopied" class="w-3.5 h-3.5" :stroke-width="2" />
                    <Check v-else class="w-3.5 h-3.5" :stroke-width="2" />
                    {{ fulfillmentCopied ? t('orderDetail.fulfillmentCopied') : t('orderDetail.fulfillmentCopy') }}
                  </button>
                </div>
                <div v-if="child.fulfillment">
                  <div class="text-sm text-muted-foreground">{{ t('orderDetail.fulfillmentType') }}：{{
                    fulfillmentTypeLabelText(child.fulfillment.type) }}</div>
                  <div class="text-sm text-muted-foreground">{{ t('orderDetail.fulfillmentStatus') }}：{{
                    fulfillmentStatusLabelText(child.fulfillment.status) }}</div>
                  <div v-if="isFulfillmentTruncated(child.fulfillment)" class="mt-3">
                    <div class="flex items-center justify-between mb-2">
                      <span class="text-sm text-muted-foreground">{{ t('orderDetail.fulfillmentTotalLines', { count: child.fulfillment.payload_line_count }) }}</span>
                      <button class="inline-flex items-center gap-1 text-xs font-medium px-3 py-1.5 rounded-lg bg-blue-600 text-white hover:bg-blue-700 active:bg-blue-800 transition-colors shadow-sm disabled:opacity-50"
                        :disabled="fulfillmentDownloading"
                        @click="handleDownloadFulfillment(child.order_no || order.order_no)">
                        <Download class="w-3.5 h-3.5" :stroke-width="2" />
                        {{ fulfillmentDownloading ? t('orderDetail.fulfillmentDownloading') : t('orderDetail.fulfillmentDownload') }}
                      </button>
                    </div>
                    <div class="mb-2 rounded-lg bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-800 px-3 py-2 text-xs text-amber-700 dark:text-amber-400">
                      {{ t('orderDetail.fulfillmentTruncatedHint') }}
                    </div>
                    <div class="border rounded-xl p-4 text-sm text-muted-foreground whitespace-pre-wrap break-all overflow-hidden max-h-48 overflow-y-auto">{{ child.fulfillment.payload }}</div>
                  </div>
                  <div v-else-if="fulfillmentDeliveryLines(child.fulfillment).length"
                    class="mt-3 border rounded-xl p-4 text-sm text-muted-foreground space-y-1 break-all overflow-hidden">
                    <div v-for="(line, index) in fulfillmentDeliveryLines(child.fulfillment)" :key="`child-fulfillment-${child.id}-${index}`">{{ line }}</div>
                  </div>
                  <div v-else-if="child.fulfillment.payload"
                    class="mt-3 border rounded-xl p-4 text-sm text-muted-foreground whitespace-pre-wrap break-all overflow-hidden">
                    {{ child.fulfillment.payload }}
                  </div>
                  <div v-if="child.fulfillment.status === 'delivered' && instructionBlocks(child.items).length"
                    class="mt-4 space-y-3">
                    <div v-for="(block, bi) in instructionBlocks(child.items)" :key="`child-inst-${child.id}-${bi}`"
                      class="rounded-xl border border-blue-200 bg-blue-50/50 dark:border-blue-900 dark:bg-blue-950/30 p-4">
                      <div class="flex items-center gap-2 mb-2 text-sm font-semibold text-blue-700 dark:text-blue-300">
                        <Info class="w-4 h-4" :stroke-width="2" />
                        {{ t('orderDetail.instructionsTitle') }}
                      </div>
                      <div class="prose prose-sm max-w-none dark:prose-invert text-muted-foreground break-words" v-html="block.html"></div>
                    </div>
                  </div>
                </div>
                <div v-else class="text-sm text-muted-foreground">{{ t('orderDetail.childFulfillmentEmpty') }}</div>
              </div>
            </div>
          </div>
        </div>

        <div v-if="order.fulfillment"
          class="rounded-2xl border bg-card shadow-sm p-6">
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-lg font-bold">{{ t('orderDetail.fulfillmentTitle') }}</h2>
            <div class="flex items-center gap-2">
              <button v-if="isFulfillmentTruncated(order.fulfillment)"
                class="inline-flex items-center gap-1 text-sm font-medium px-4 py-2 rounded-lg bg-blue-600 text-white hover:bg-blue-700 active:bg-blue-800 transition-colors shadow-sm disabled:opacity-50"
                :disabled="fulfillmentDownloading"
                @click="handleDownloadFulfillment(order.order_no)">
                <Download class="w-4 h-4" :stroke-width="2" />
                {{ fulfillmentDownloading ? t('orderDetail.fulfillmentDownloading') : t('orderDetail.fulfillmentDownload') }}
              </button>
              <button v-if="order.fulfillment.status === 'delivered' && !isFulfillmentTruncated(order.fulfillment)"
                class="inline-flex items-center gap-1 text-sm font-medium px-4 py-2 rounded-lg transition-colors shadow-sm"
                :class="fulfillmentCopied ? 'bg-emerald-600 text-white' : 'bg-blue-600 text-white hover:bg-blue-700 active:bg-blue-800'"
                @click="handleCopyFulfillment(order.fulfillment)">
                <Copy v-if="!fulfillmentCopied" class="w-4 h-4" :stroke-width="2" />
                <Check v-else class="w-4 h-4" :stroke-width="2" />
                {{ fulfillmentCopied ? t('orderDetail.fulfillmentCopied') : t('orderDetail.fulfillmentCopy') }}
              </button>
            </div>
          </div>
          <div class="text-sm text-muted-foreground">{{ t('orderDetail.fulfillmentType') }}：{{
            fulfillmentTypeLabelText(order.fulfillment.type) }}</div>
          <div class="text-sm text-muted-foreground">{{ t('orderDetail.fulfillmentStatus') }}：{{
            fulfillmentStatusLabelText(order.fulfillment.status) }}</div>
          <div v-if="isFulfillmentTruncated(order.fulfillment)" class="mt-4">
            <div class="text-sm text-muted-foreground mb-2">{{ t('orderDetail.fulfillmentTotalLines', { count: order.fulfillment.payload_line_count }) }}</div>
            <div class="mb-2 rounded-lg bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-800 px-3 py-2 text-xs text-amber-700 dark:text-amber-400">
              {{ t('orderDetail.fulfillmentTruncatedHint') }}
            </div>
            <div class="border rounded-xl p-4 text-sm text-muted-foreground whitespace-pre-wrap break-all overflow-hidden max-h-64 overflow-y-auto">{{ order.fulfillment.payload }}</div>
          </div>
          <div v-else-if="fulfillmentDeliveryLines(order.fulfillment).length"
            class="mt-4 border rounded-xl p-4 text-sm text-muted-foreground space-y-1 break-all overflow-hidden">
            <div v-for="(line, index) in fulfillmentDeliveryLines(order.fulfillment)" :key="`fulfillment-${order.order_no || 'order'}-${index}`">{{ line }}</div>
          </div>
          <div v-else
            class="mt-4 border rounded-xl p-4 text-sm text-muted-foreground whitespace-pre-wrap break-all overflow-hidden">
            {{ order.fulfillment.payload }}
          </div>
          <div v-if="order.fulfillment.status === 'delivered' && instructionBlocks(order.items).length"
            class="mt-4 space-y-3">
            <div v-for="(block, bi) in instructionBlocks(order.items)" :key="`order-inst-${bi}`"
              class="rounded-xl border border-blue-200 bg-blue-50/50 dark:border-blue-900 dark:bg-blue-950/30 p-4">
              <div class="flex items-center gap-2 mb-2 text-sm font-semibold text-blue-700 dark:text-blue-300">
                <Info class="w-4 h-4" :stroke-width="2" />
                {{ t('orderDetail.instructionsTitle') }}
              </div>
              <div class="prose prose-sm max-w-none dark:prose-invert text-muted-foreground break-words" v-html="block.html"></div>
            </div>
          </div>
        </div>

        <div class="rounded-2xl border bg-card shadow-sm p-6">
          <h2 class="text-lg font-bold mb-4">{{ t('orderDetail.itemsTitle') }}</h2>
          <div v-if="order.items && order.items.length > 0" class="space-y-4">
            <div v-for="(item, idx) in order.items" :key="idx"
              class="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3 sm:gap-4 border-b border-gray-100 pb-3 dark:border-white/5">
              <div class="flex min-w-0 items-start gap-3">
                <div class="h-14 w-14 shrink-0 overflow-hidden rounded-xl border border-gray-200 bg-white transition-all duration-200 hover:-translate-y-0.5 hover:shadow-sm dark:border-white/10 dark:bg-black/30 sm:h-16 sm:w-16">
                  <SmartImage
                    :src="orderItemImage(item)"
                    :alt="getLocalizedText(item.title)"
                    img-class="h-full w-full object-cover"
                  />
                </div>
                <div class="min-w-0">
                  <div class="text-foreground font-medium">{{ getLocalizedText(item.title) }}</div>
                  <div class="text-xs text-muted-foreground">{{ t('orderDetail.quantityLabel') }}：{{ item.quantity }}</div>
                  <div v-if="orderItemSkuText(item)" class="text-xs text-muted-foreground mt-1">{{ t('orderDetail.itemSkuLabel') }}：{{ orderItemSkuText(item) }}</div>
                  <div class="text-xs text-muted-foreground mt-1">
                    {{ t('orderDetail.itemFulfillmentLabel') }}：{{ fulfillmentTypeLabelText(item.fulfillment_type) }}
                  </div>
                  <div v-if="item.tags && item.tags.length" class="mt-2 flex flex-wrap gap-2">
                    <Badge v-for="(tag, index) in item.tags" :key="index" variant="neutral" size="sm" class="rounded-full">{{ tag }}</Badge>
                  </div>
                  <div v-if="manualSubmissionRows(item.manual_form_submission, item.manual_form_schema_snapshot).length"
                    class="mt-3 rounded-xl border border-gray-200 bg-white p-3 text-xs text-gray-600 dark:border-white/10 dark:bg-black/30 dark:text-gray-300">
                    <div class="mb-2 font-semibold text-muted-foreground">{{ t('orderDetail.manualSubmissionTitle') }}</div>
                    <div v-for="row in manualSubmissionRows(item.manual_form_submission, item.manual_form_schema_snapshot)" :key="row.key" class="mb-1 last:mb-0">
                      <span class="text-foreground">{{ row.label }}</span>：{{ row.value }}
                    </div>
                  </div>
                </div>
              </div>
              <div class="shrink-0 pl-[4.25rem] sm:pl-0 text-left sm:text-right text-sm text-muted-foreground space-y-1">
                <div>{{ t('orderDetail.unitPriceLabel') }}：{{ formatMoney(item.original_unit_price, order.currency) }}</div>
                <div>{{ t('orderDetail.totalPriceLabel') }}：{{ formatMoney(item.original_total_price, order.currency) }}</div>
                <div v-if="hasDiscountAmount(item.coupon_discount_amount)">
                  {{ t('orderDetail.couponDiscountLabel') }}：{{ formatDiscountMoney(item.coupon_discount_amount, order.currency)
                  }}
                </div>
                <div v-if="hasDiscountAmount(item.promotion_discount_amount)">
                  {{ t('orderDetail.promotionDiscountLabel') }}：{{ formatDiscountMoney(item.promotion_discount_amount,
                  order.currency) }}
                </div>
                <div v-if="hasDiscountAmount(item.wholesale_discount_amount)">
                  {{ t('orderDetail.wholesaleDiscountLabel') }}：{{ formatDiscountMoney(item.wholesale_discount_amount,
                  order.currency) }}
                </div>
                <div v-if="hasDiscountAmount(item.member_discount_amount)">
                  {{ t('orderDetail.memberDiscountLabel') }}：{{ formatDiscountMoney(item.member_discount_amount,
                  order.currency) }}
                </div>
                <div v-if="hasItemDiscount(item)" class="font-medium text-rose-600 dark:text-rose-300">
                  {{ t('orderDetail.itemDiscountTotalLabel') }}：{{ formatItemDiscountTotal(item, order.currency) }}
                </div>
                <div class="font-medium text-foreground">
                  {{ t('orderDetail.itemPaidAmountLabel') }}：{{ formatItemPaidAmount(item, order.currency) }}
                </div>
              </div>
            </div>
          </div>
          <div v-else class="text-sm text-muted-foreground">{{ t('orderDetail.noItems') }}</div>
        </div>

        <div class="rounded-2xl border bg-card shadow-sm p-6">
          <h2 class="text-lg font-bold mb-4">{{ t('orderDetail.amountTitle') }}</h2>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
            <div class="border rounded-xl p-4">
              <div class="text-xs text-muted-foreground">{{ t('orderDetail.amountOriginal') }}</div>
              <div class="text-foreground font-mono mt-1">{{ formatMoney(order.original_amount,
                order.currency) }}</div>
            </div>
            <div class="border rounded-xl p-4">
              <div class="text-xs text-muted-foreground">{{ t('orderDetail.amountDiscount') }}</div>
              <div
                class="font-mono mt-1"
                :class="hasDiscountAmount(order.discount_amount) ? 'text-rose-600 dark:text-rose-300' : 'text-foreground'"
              >
                {{ formatDiscountMoney(order.discount_amount, order.currency) }}
              </div>
            </div>
            <div class="border rounded-xl p-4">
              <div class="text-xs text-muted-foreground">{{ t('orderDetail.amountTotal') }}</div>
              <div class="text-foreground font-mono mt-1">{{ formatMoney(order.total_amount,
                order.currency) }}</div>
            </div>
            <div v-if="hasAmount(order.wallet_paid_amount)" class="border rounded-xl p-4">
              <div class="text-xs text-muted-foreground">{{ t('orderDetail.amountWalletPaid') }}</div>
              <div class="text-foreground font-mono mt-1">{{ formatMoney(order.wallet_paid_amount,
                order.currency) }}</div>
            </div>
            <div v-if="hasAmount(order.online_paid_amount)" class="border rounded-xl p-4">
              <div class="text-xs text-muted-foreground">{{ t('orderDetail.amountOnlinePaid') }}</div>
              <div class="text-foreground font-mono mt-1">{{ formatMoney(order.online_paid_amount,
                order.currency) }}</div>
            </div>
            <div v-if="hasAmount(order.refunded_amount)" class="border rounded-xl p-4">
              <div class="text-xs text-muted-foreground">{{ t('orderDetail.amountRefunded') }}</div>
              <div class="text-foreground font-mono mt-1">{{ formatMoney(order.refunded_amount,
                order.currency) }}</div>
            </div>
            <div v-if="hasDiscountAmount(order.member_discount_amount)" class="border border-amber-200 bg-amber-50/50 dark:border-amber-800 dark:bg-amber-950/30 rounded-xl p-4">
              <div class="text-xs text-amber-700 dark:text-amber-400">{{ t('orderDetail.amountMemberDiscount') }}</div>
              <div class="text-amber-700 dark:text-amber-400 font-mono mt-1">{{ formatDiscountMoney(order.member_discount_amount,
                order.currency) }}</div>
            </div>
            <div v-if="hasDiscountAmount(order.wholesale_discount_amount)" class="border border-emerald-200 bg-emerald-50/50 dark:border-emerald-800 dark:bg-emerald-950/30 rounded-xl p-4">
              <div class="text-xs text-emerald-700 dark:text-emerald-400">{{ t('orderDetail.amountWholesaleDiscount') }}</div>
              <div class="text-emerald-700 dark:text-emerald-400 font-mono mt-1">{{ formatDiscountMoney(order.wholesale_discount_amount,
                order.currency) }}</div>
            </div>
          </div>
        </div>

        <div v-if="showRefundRecordsCard" class="rounded-2xl border bg-card shadow-sm p-6">
          <h2 class="text-lg font-bold mb-4">{{ t('orderDetail.refundRecordsTitle') }}</h2>
          <div v-if="refundRecords.length > 0" class="overflow-x-auto rounded-xl border">
            <Table>
              <TableHeader>
                <TableRow class="bg-muted/50">
                  <TableHead class="px-4">{{ t('orderDetail.refundRecordTime') }}</TableHead>
                  <TableHead class="px-4">{{ t('orderDetail.refundRecordAmount') }}</TableHead>
                  <TableHead class="px-4">{{ t('orderDetail.refundRecordReason') }}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow
                  v-for="(record, idx) in refundRecords"
                  :key="`refund-record-row-${idx}`"
                >
                  <TableCell class="px-4 text-xs text-muted-foreground whitespace-nowrap">{{ formatDate(record.created_at) }}</TableCell>
                  <TableCell class="px-4 font-mono text-sm text-foreground whitespace-nowrap">{{ formatMoney(record.amount, record.currency || order.currency) }}</TableCell>
                  <TableCell class="px-4 text-xs text-muted-foreground whitespace-pre-wrap break-words">{{ refundReasonText(record.remark) }}</TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>
          <div v-else class="text-sm text-muted-foreground">{{ t('orderDetail.refundRecordsEmpty') }}</div>
        </div>

        <div v-if="showTimeCard" class="rounded-2xl border bg-card shadow-sm p-6">
          <h2 class="text-lg font-bold mb-4">{{ t('orderDetail.timeTitle') }}</h2>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
            <div class="border rounded-xl p-4">
              <div class="text-xs text-muted-foreground">{{ t('orderDetail.createdAtLabel') }}</div>
              <div class="text-foreground mt-1">{{ formatDate(order.created_at) }}</div>
            </div>
            <div v-if="order.paid_at" class="border rounded-xl p-4">
              <div class="text-xs text-muted-foreground">{{ t('orderDetail.paidAtLabel') }}</div>
              <div class="text-foreground mt-1">{{ formatDate(order.paid_at) }}</div>
            </div>
            <div v-if="order.expires_at" class="border rounded-xl p-4">
              <div class="text-xs text-muted-foreground">{{ t('orderDetail.expiresAtLabel') }}</div>
              <div class="text-foreground mt-1">{{ formatDate(order.expires_at) }}</div>
            </div>
            <div v-if="order.canceled_at" class="border rounded-xl p-4">
              <div class="text-xs text-muted-foreground">{{ t('orderDetail.canceledAtLabel') }}</div>
              <div class="text-foreground mt-1">{{ formatDate(order.canceled_at) }}</div>
            </div>
          </div>
        </div>
        </template>
        <section v-else class="rounded-2xl border bg-card p-5 shadow-sm md:p-6">
          <div class="flex flex-wrap items-start justify-between gap-3 border-b pb-4">
            <div class="min-w-0">
              <div class="text-xs text-muted-foreground">{{ t('orders.orderNo') }}</div>
              <div class="mt-1 break-all font-semibold">{{ order.order_no }}</div>
              <div class="mt-1 text-xs text-muted-foreground">{{ t('orderDetail.createdAtLabel') }}：{{ formatDate(order.created_at) }}</div>
            </div>
            <Badge :variant="statusVariant(order.status)" size="sm">{{ statusLabel(order.status) }}</Badge>
          </div>
          <div class="space-y-3 py-4">
            <div v-for="(item, index) in detailItems" :key="item.id || index" class="flex justify-between gap-4 text-sm">
              <span class="min-w-0">{{ getLocalizedText(item.title) }} <span class="text-muted-foreground">× {{ item.quantity }}</span></span>
              <span v-if="index === 0" class="shrink-0 text-muted-foreground">{{ formatMoney(order.total_amount, order.currency) }}</span>
            </div>
          </div>
          <div class="grid gap-3 border-t pt-4 text-sm sm:grid-cols-2">
            <div><span class="text-muted-foreground">实付金额：</span><strong>{{ formatMoney(order.total_amount, order.currency) }}</strong></div>
            <div v-if="order.expires_at"><span class="text-muted-foreground">{{ t('orderDetail.expiresAtLabel') }}：</span>{{ formatDate(order.expires_at) }}</div>
            <div v-if="order.paid_at"><span class="text-muted-foreground">{{ t('orderDetail.paidAtLabel') }}：</span>{{ formatDate(order.paid_at) }}</div>
            <div v-if="order.canceled_at"><span class="text-muted-foreground">{{ t('orderDetail.canceledAtLabel') }}：</span>{{ formatDate(order.canceled_at) }}</div>
          </div>
          <div v-if="fulfillmentInfo" class="mt-4 rounded-xl border border-emerald-200 bg-emerald-50/60 p-4 dark:border-emerald-900 dark:bg-emerald-950/20">
            <div class="flex items-center justify-between gap-3">
              <div class="font-semibold">履约结果</div>
              <button v-if="fulfillmentInfo.status === 'delivered'" class="text-sm text-primary" @click="handleCopyFulfillment(fulfillmentInfo)">
                {{ fulfillmentCopied ? t('orderDetail.fulfillmentCopied') : t('orderDetail.fulfillmentCopy') }}
              </button>
            </div>
            <div class="mt-1 text-sm text-muted-foreground">{{ fulfillmentStatusLabelText(fulfillmentInfo.status) }}</div>
            <pre v-if="fulfillmentInfo.payload" class="mt-3 whitespace-pre-wrap break-all text-sm">{{ fulfillmentInfo.payload }}</pre>
            <div v-else-if="['delivered', 'completed'].includes(order.status)" class="mt-2 text-sm text-muted-foreground">订单已履约完成。</div>
          </div>
          <RouterLink :to="orderListPath" class="mt-5 inline-flex text-sm font-medium text-primary hover:underline">返回代理拿货订单</RouterLink>
        </section>
        </details>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Copy, Check, Download, Info } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import EmptyState from '../components/EmptyState.vue'
import BreadcrumbNav from '../components/BreadcrumbNav.vue'
import SmartImage from '../components/SmartImage.vue'
import SessionAccountSummary from '../components/SessionAccountSummary.vue'
import SessionProgress from '../components/SessionProgress.vue'
import SmsActivationConsole from '../components/SmsActivationConsole.vue'
import { useOrderDetail } from '../composables/useOrderDetail'
import { hasPendingSession } from '../utils/sessionOrder'
import { isSmsActivationOrder, isSmsActiveStatus } from '../utils/smsOrder'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { userOrderAPI } from '../api/order'
import { toast } from '../composables/useToast'
import { useUserAuthStore } from '../stores/userAuth'

const { t } = useI18n()
const route = useRoute()
const userAuthStore = useUserAuthStore()
const isSessionPage = computed(() => route.name === 'order-session')
const isAgentOrderView = computed(() => route.query.from === 'agent' || userAuthStore.user?.role === 'agent')
const orderListPath = computed(() => isAgentOrderView.value ? '/agent?tab=orders' : '/me/orders')
const orderListLabel = computed(() => isAgentOrderView.value ? '代理拿货订单' : t('orders.title'))
const sessionBackPath = computed(() => userAuthStore.user?.role === 'agent' || route.query.from === 'agent' ? '/agent?tab=orders' : '/me/orders')
const sessionBreadcrumbItems = computed(() => [
  { label: userAuthStore.user?.role === 'agent' || route.query.from === 'agent' ? '代理拿货订单' : '我的订单', to: sessionBackPath.value },
  { label: '账号确认与充值' },
])
const sessionValue = ref('')
const accountConfirmed = ref(false)
const sessionSubmitting = ref(false)
const sessionPreflighting = ref(false)
const sessionReviewOpen = ref(false)
const sessionAccountSummary = ref<Record<string, any> | null>(null)
const sessionJSONValid = computed(() => { try { return JSON.parse(sessionValue.value.trim()) !== null } catch { return false } })
watch(sessionValue, () => {
  accountConfirmed.value = false
  sessionAccountSummary.value = null
  sessionReviewOpen.value = false
})
const isSessionOrder = computed(() => !!order.value && [order.value, ...(order.value.children || [])].some((entry: any) => (entry.items || []).some((item: any) => item.requires_session)))

const {
  loading, order, debouncedLoadOrder, cancelOrder,
  statusLabel, statusVariant, fulfillmentTypeLabelText, fulfillmentStatusLabelText,
  formatDate, getLocalizedText, formatMoney, formatDiscountMoney, hasDiscountAmount, hasAmount,
  refundReasonText, showRefundRecordsCard, refundRecords, showTimeCard,
  orderItemImage, orderItemSkuText, manualSubmissionRows,
  hasItemDiscount, formatItemDiscountTotal, formatItemPaidAmount, resolvedChildStatus,
  fulfillmentDeliveryLines, instructionBlocks, isFulfillmentTruncated,
  fulfillmentCopied, handleCopyFulfillment, fulfillmentDownloading, handleDownloadFulfillment,
} = useOrderDetail()

const isSmsOrder = computed(() => isSmsActivationOrder(order.value))
const isSmsEligible = computed(() => isSmsActiveStatus(order.value))

const needsSession = computed(() => {
  return hasPendingSession(order.value)
})
const sessionSummaryItems = computed(() => {
  const children = order.value?.children || []
  const childItems = children.flatMap((child: any) => child.items || [])
  return childItems.length ? childItems : (order.value?.items || [])
})
const detailItems = computed(() => {
  const children = order.value?.children || []
  const items = children.flatMap((child: any) => child.items || [])
  return items.length ? items : (order.value?.items || [])
})
const fulfillmentInfo = computed(() => {
  if (order.value?.fulfillment) return order.value.fulfillment
  return (order.value?.children || []).find((child: any) => child.fulfillment)?.fulfillment || null
})

let progressTimer: ReturnType<typeof setInterval> | undefined
let progressLoading = false
onMounted(() => {
  progressTimer = setInterval(async () => {
    if (progressLoading || !order.value?.order_no || needsSession.value || !['paid', 'fulfilling', 'partially_delivered'].includes(order.value.status)) return
    progressLoading = true
    try {
      const response = await userOrderAPI.detail(order.value.order_no)
      order.value = response.data.data
    } catch {} finally { progressLoading = false }
  }, 4000)
})
onUnmounted(() => { if (progressTimer) clearInterval(progressTimer) })

const submitSession = async () => {
  if (!order.value?.order_no || !sessionValue.value.trim() || sessionSubmitting.value || !sessionAccountSummary.value || !accountConfirmed.value) return
  sessionSubmitting.value = true
  try {
    await userOrderAPI.submitSession(order.value.order_no, sessionValue.value.trim())
    sessionValue.value = ''
    toast.success('Session 已提交，系统正在继续处理订单')
    await debouncedLoadOrder()
  } catch (error: any) {
    toast.error(error?.message || '提交失败，请稍后重试')
  } finally {
    sessionSubmitting.value = false
  }
}

const preflightSession = async () => {
  if (!order.value?.order_no || !sessionJSONValid.value || sessionPreflighting.value) return
  sessionPreflighting.value = true
  sessionAccountSummary.value = null
  try {
    const result = await userOrderAPI.preflightSession(order.value.order_no, sessionValue.value.trim())
    sessionAccountSummary.value = result.data?.data || null
    if (!sessionAccountSummary.value) throw new Error(t('accountPreflight.noAccount'))
    sessionReviewOpen.value = true
  } catch (error: any) {
    toast.error(error?.message || t('accountPreflight.failed'))
  } finally {
    sessionPreflighting.value = false
  }
}
</script>
