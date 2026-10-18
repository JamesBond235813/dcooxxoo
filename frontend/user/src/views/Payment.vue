<template>
  <div class="min-h-screen bg-background text-foreground pt-24 pb-16">
    <div class="container mx-auto px-4">
      <div class="mb-6 flex items-center justify-between">
        <div>
          <h1 class="text-3xl font-black text-foreground mb-2">{{ t('payment.title') }}</h1>
          <p class="text-muted-foreground text-sm">{{ t('payment.subtitle') }}</p>
        </div>
        <router-link :to="backLink"
          class="text-muted-foreground hover:text-foreground transition-colors text-sm">{{
            t('payment.backToOrders') }}</router-link>
      </div>

      <CheckoutSteps class="mb-8" current-step="payment" />

      <!-- Loading Skeleton -->
      <div v-if="loading" class="space-y-6">
        <div class="border bg-card text-card-foreground shadow-sm rounded-2xl p-6 space-y-4">
          <div class="flex items-center justify-between gap-4">
            <div class="space-y-2">
              <div class="h-5 w-40 rounded theme-skeleton"></div>
              <div class="h-3 w-56 rounded theme-skeleton"></div>
            </div>
            <div class="h-7 w-20 rounded-full theme-skeleton"></div>
          </div>
          <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 pt-4 border-t">
            <div class="lg:col-span-2 space-y-3">
              <div class="w-full max-w-[260px] aspect-square rounded-xl theme-skeleton mx-auto lg:mx-0"></div>
              <div class="h-3 w-48 rounded theme-skeleton"></div>
            </div>
            <div class="space-y-3">
              <div class="h-4 w-24 rounded theme-skeleton"></div>
              <div class="h-4 w-32 rounded theme-skeleton"></div>
              <div class="h-4 w-28 rounded theme-skeleton"></div>
            </div>
          </div>
        </div>
      </div>

      <div v-else-if="showGuestAuthForm"
        class="border bg-card text-card-foreground shadow-sm rounded-2xl p-6">
        <h2 class="text-lg font-bold mb-2">{{ t('payment.guestAuthTitle') }}</h2>
        <p class="text-xs text-muted-foreground mb-4">{{ t('payment.guestAuthHint') }}</p>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Input v-model="guestAuth.email" type="email"
            class="h-11"
            :placeholder="t('guestOrders.emailPlaceholder')" />
        </div>
        <Alert v-if="guestAuthError" variant="destructive" class="mt-4">
          <AlertDescription>{{ guestAuthError }}</AlertDescription>
        </Alert>
        <Button variant="secondary" class="mt-4 font-semibold" @click="handleGuestAuthSubmit">
          {{ t('payment.guestAuthSubmit') }}
        </Button>
      </div>

      <EmptyState
        v-else-if="!order"
        icon="alert"
        :title="t('payment.orderNotFound')"
        :action-label="t('payment.backToOrders')"
        :action-to="backLink"
      />

      <div v-else-if="showResultView" class="space-y-6">
        <div class="border bg-card text-card-foreground shadow-sm rounded-2xl p-6">
          <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
            <div>
              <h2 class="text-xl font-bold text-foreground">{{ paymentResultTitle }}</h2>
              <p class="text-sm text-muted-foreground mt-1">{{ paymentGuideTip }}</p>
              <div class="mt-2 text-xs text-muted-foreground">
                {{ t('payment.methodLabel') }}：{{ resultChannelName }}
              </div>
              <div v-if="cryptoNetworkInfo" class="mt-3 rounded-lg border-2 border-amber-500 bg-amber-50 px-3 py-2.5 text-amber-950 dark:bg-amber-950/40 dark:text-amber-100" role="alert">
                <div class="text-sm font-extrabold">转账网络：{{ cryptoNetworkInfo.label }}</div>
                <div class="mt-1 text-xs font-medium">{{ cryptoNetworkInfo.hint }} 发送到其他网络可能造成资产损失。</div>
              </div>
            </div>
            <div class="flex flex-wrap items-center gap-2">
              <Button as-child variant="default" class="font-semibold bg-emerald-600 hover:bg-emerald-700 text-white">
                <router-link to="/guest/orders">
                  前往游客查单
                </router-link>
              </Button>
              <Button variant="secondary" :disabled="loading" @click="handleRefresh">
                {{ t('payment.refreshStatus') }}
              </Button>
              <Button variant="secondary" @click="handleChangePaymentMethod">
                {{ t('payment.changeMethod') }}
              </Button>
            </div>
          </div>

          <div class="mt-6 grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div class="lg:col-span-2 space-y-4">
              <div v-if="showQRCode"
                class="bg-secondary border rounded-2xl p-4 sm:p-6 flex flex-col items-center justify-center text-center">
                <div class="text-sm text-muted-foreground mb-4">{{ paymentGuideTitle }}</div>
                <div class="w-full max-w-[280px] sm:max-w-[240px] aspect-square rounded-xl overflow-hidden bg-white p-2">
                  <img :src="qrImageUrl" alt="QR Code" class="w-full h-full object-contain" />
                </div>
                <div v-if="qrUsingPayLinkFallback" class="mt-3 text-xs text-muted-foreground">
                  {{ t('payment.qrFallbackHint') }}
                </div>
                <div v-if="hasCryptoPaymentDetails" class="mt-4 w-full max-w-xl space-y-2 rounded-xl border bg-white/5 p-3 text-left">
                  <div
                    v-for="item in cryptoPaymentDetails"
                    :key="item.key"
                    class="flex flex-col gap-1 border-b pb-2 last:border-b-0 last:pb-0 sm:flex-row sm:items-start sm:justify-between sm:gap-4"
                  >
                    <span class="shrink-0 text-xs text-muted-foreground">{{ item.label }}</span>
                    <span class="min-w-0 text-sm font-semibold text-foreground break-all sm:text-right">
                      {{ item.value }}
                      <span v-if="item.detail" class="ml-1 font-normal text-muted-foreground">({{ item.detail }})</span>
                    </span>
                  </div>
                  <div v-if="cryptoWalletAddress" class="flex flex-wrap items-center justify-end gap-2 pt-1">
                    <Button variant="secondary" size="sm" @click="handleCopyWalletAddress">
                      {{ t('payment.copyWalletAddress') }}
                    </Button>
                    <span v-if="walletAddressCopied" class="text-xs text-success">{{ t('payment.copied') }}</span>
                  </div>
                </div>
              </div>

              <div v-else class="bg-secondary border rounded-2xl p-6">
                <div class="text-sm text-muted-foreground mb-3">{{ t('payment.openPayLink') }}</div>
                <Button type="button" variant="default" class="font-bold" @click="handleOpenPayLink">
                  <ExternalLink class="h-4 w-4" aria-hidden="true" />
                  {{ t('payment.openPayLink') }}
                </Button>
                <div v-if="openedPayWindow" class="mt-3 text-xs text-success">
                  {{ payLinkOpenedTip }}
                </div>
                <div v-if="showTelegramPayHint" class="mt-3 text-xs text-muted-foreground">
                  {{ t('payment.telegramExternalHint') }}
                </div>
                <div class="mt-3 flex flex-wrap items-center gap-2">
                  <Button type="button" variant="outline" size="sm" class="font-semibold" @click="handleCopyPayLink">
                    <Copy class="h-4 w-4" aria-hidden="true" />
                    {{ t('payment.copyPayLink') }}
                  </Button>
                  <span v-if="copied" class="text-xs text-success">{{ t('payment.copied') }}</span>
                </div>
              </div>
            </div>

            <div v-if="isManualCryptoPayment" class="rounded-2xl border bg-card p-4 text-left">
              <div class="font-semibold">USDT 转账凭证（人工核账）</div>
              <p class="mt-1 text-xs text-muted-foreground">请确认网络、地址和金额后提交链上交易哈希。提交凭证不会自动确认付款，管理员核实到账后才会处理订单。</p>
              <div v-if="manualProofStatus === 'submitted'" class="mt-3 break-all text-sm">凭证状态：待管理员核验 · {{ paymentResult.manual_tx_hash }}</div>
              <div v-else-if="manualProofStatus === 'approved'" class="mt-3 text-sm text-success">凭证已核验，订单正在处理。</div>
              <div v-else class="mt-3 flex flex-col gap-2 sm:flex-row">
                <input v-model="manualTxHashInput" class="h-10 min-w-0 flex-1 rounded-md border bg-background px-3 text-sm" autocomplete="off" placeholder="交易哈希 / Transaction ID" />
                <Button :disabled="manualProofSubmitting || !manualTxHashInput.trim()" @click="handleSubmitManualCryptoProof">{{ manualProofSubmitting ? '提交中…' : '提交转账凭证' }}</Button>
              </div>
            </div>

            <div class="space-y-4">
              <div class="bg-secondary border rounded-2xl p-4">
                <div class="text-xs text-muted-foreground">{{ t('payment.orderNo') }}</div>
                <div class="text-sm font-semibold text-foreground mt-1">{{ order.order_no }}</div>
                <div class="mt-2 text-xs text-muted-foreground">{{ t('payment.orderStatus') }}：{{ statusLabel(order.status) }}</div>
              </div>
              <PaymentAmountBreakdown
                :order="order"
                :payment-result="paymentResult"
                :customer-fee-applied="customerFeeApplied"
                :customer-fee-amount-display="customerFeeAmountDisplay"
                :payable-amount-display="payableAmountDisplay"
                :wallet-paid-display="paymentWalletPaidDisplay"
                :online-pay-display="paymentOnlinePayDisplay"
                :show-countdown="showCountdown"
                :countdown-text="countdownText"
                :polling-active="pollingActive"
                :format-money="formatMoney"
                :format-discount-money="formatDiscountMoney"
                :has-discount-amount="hasDiscountAmount"
              />
              <div v-if="paymentResult.expires_at"
                class="bg-secondary border rounded-2xl p-4 text-xs text-muted-foreground">
                {{ t('payment.expiresAt') }}：{{ formatDate(paymentResult.expires_at) }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else-if="orderExpired || orderCanceled" class="space-y-6">
        <div class="border bg-card text-card-foreground shadow-sm rounded-2xl p-6">
          <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
            <div>
              <h2 class="text-xl font-bold text-foreground">
                {{ orderCanceled ? t('payment.orderCanceled') : t('payment.orderExpired') }}
              </h2>
              <p class="text-sm text-muted-foreground mt-1">{{ order.order_no }}</p>
            </div>
            <div class="flex flex-wrap items-center gap-2">
              <Button as-child variant="secondary" class="font-semibold">
                <router-link :to="backLink">
                  {{ t('payment.backToOrders') }}
                </router-link>
              </Button>
            </div>
          </div>
          <div class="mt-6 grid grid-cols-1 sm:grid-cols-3 gap-3 text-sm">
            <div class="bg-secondary border rounded-xl p-3">
              <div class="text-xs text-muted-foreground">{{ t('payment.orderNo') }}</div>
              <div class="text-foreground font-mono mt-1">{{ order.order_no }}</div>
            </div>
            <div class="bg-secondary border rounded-xl p-3">
              <div class="text-xs text-muted-foreground">{{ t('payment.orderStatus') }}</div>
              <div class="text-foreground font-mono mt-1">{{ statusLabel(order.status) }}</div>
            </div>
            <div class="bg-secondary border rounded-xl p-3">
              <div class="text-xs text-muted-foreground">{{ t('orderDetail.amountTotal') }}</div>
              <div class="text-foreground font-mono mt-1">{{ formatMoney(order.total_amount, order.currency) }}</div>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <div class="lg:col-span-2 space-y-6">
          <div class="border bg-card text-card-foreground shadow-sm rounded-2xl p-6">
            <h2 class="text-lg font-bold mb-4 text-foreground">{{ t('payment.orderInfo') }}</h2>
            <div class="flex flex-col md:flex-row md:items-start md:justify-between gap-4">
              <div>
                <div class="text-xs uppercase tracking-wider text-muted-foreground">{{ t('payment.orderNo') }}</div>
                <div class="text-sm font-semibold text-foreground mt-1">{{ order.order_no }}</div>
              <div class="text-xs text-muted-foreground mt-2">{{ t('orderDetail.createdAtLabel') }}：{{
                  formatDate(order.created_at) }}</div>
            </div>
            <div class="w-full md:w-auto md:min-w-[280px] bg-secondary border rounded-2xl p-4">
              <div class="text-xs uppercase tracking-wider text-muted-foreground md:text-right">{{ t('payment.payableAmountLabel') }}</div>
              <div class="mt-1 text-2xl font-bold text-foreground md:text-right">{{ payableAmountDisplay }}</div>
              <div class="mt-4 space-y-2 text-xs">
                <div class="flex items-center justify-between gap-4">
                  <span class="text-muted-foreground">{{ t('orderDetail.amountTotal') }}</span>
                  <span class="font-semibold text-foreground">{{ formatMoney(order.total_amount, order.currency) }}</span>
                </div>
                <div v-if="showBalanceOption && useBalance" class="flex items-center justify-between gap-4">
                  <span class="text-muted-foreground">{{ t('payment.walletDeductLabel') }}</span>
                  <span class="font-medium text-foreground">{{ expectedWalletPaidDisplay }}</span>
                </div>
                <div v-if="showBalanceOption && useBalance" class="flex items-center justify-between gap-4">
                  <span class="text-muted-foreground">{{ t('payment.onlinePayLabel') }}</span>
                  <span class="font-medium text-foreground">{{ expectedOnlinePayDisplay }}</span>
                </div>
              </div>
              <div class="mt-4 border-t pt-3 text-xs">
                <div class="flex items-center justify-between gap-4">
                  <span class="text-muted-foreground">{{ t('payment.orderStatus') }}</span>
                  <span class="font-medium text-foreground">{{ statusLabel(order.status) }}</span>
                </div>
              </div>
            </div>
            </div>
            <div class="mt-4 grid grid-cols-1 sm:grid-cols-3 gap-3 text-sm">
              <div class="bg-secondary border rounded-xl p-3">
                <div class="text-xs text-muted-foreground">{{ t('orderDetail.amountOriginal') }}</div>
                <div class="text-foreground font-mono mt-1">{{ formatMoney(order.original_amount,
                  order.currency) }}</div>
              </div>
              <div class="bg-secondary border rounded-xl p-3">
                <div class="text-xs text-muted-foreground">{{ t('orderDetail.amountDiscount') }}</div>
                <div
                  class="font-mono mt-1"
                  :class="hasDiscountAmount(order.discount_amount) ? 'text-rose-600 dark:text-rose-300' : 'text-foreground'"
                >
                  {{ formatDiscountMoney(order.discount_amount, order.currency) }}
                </div>
              </div>
              <div class="bg-secondary border rounded-xl p-3">
                <div class="text-xs text-muted-foreground">{{ t('orderDetail.promotionDiscountLabel') }}</div>
                <div
                  class="font-mono mt-1"
                  :class="hasDiscountAmount(order.promotion_discount_amount) ? 'text-rose-600 dark:text-rose-300' : 'text-foreground'"
                >
                  {{ formatDiscountMoney(order.promotion_discount_amount, order.currency) }}
                </div>
              </div>
              <div v-if="hasDiscountAmount(order.wholesale_discount_amount)" class="border-emerald-200 bg-emerald-50/50 dark:border-emerald-800 dark:bg-emerald-950/30 rounded-xl p-3">
                <div class="text-xs text-emerald-700 dark:text-emerald-400">{{ t('orderDetail.amountWholesaleDiscount') }}</div>
                <div class="text-emerald-700 dark:text-emerald-400 font-mono mt-1">
                  {{ formatDiscountMoney(order.wholesale_discount_amount, order.currency) }}
                </div>
              </div>
            </div>
            <div class="mt-3 text-sm text-muted-foreground">
              <span v-if="order.expires_at">{{ t('payment.expiresAt') }}：{{ formatDate(order.expires_at) }}</span>
            </div>
            <div v-if="pollingActive" class="mt-3 text-xs text-muted-foreground">{{ t('payment.pollingHint') }}</div>
          </div>

          <div v-if="orderItems.length"
            class="border bg-card text-card-foreground shadow-sm rounded-2xl p-6">
            <h2 class="text-lg font-bold mb-4 text-foreground">{{ t('payment.itemsTitle') }}</h2>
            <div class="space-y-3 text-sm text-muted-foreground">
              <div v-for="(item, idx) in orderItems" :key="idx"
                class="flex flex-col md:flex-row md:items-center md:justify-between gap-2 border-b pb-3">
                <div>
                  <div class="text-foreground font-medium">{{ getLocalizedText(item.title) }}</div>
                  <div v-if="item.requires_session" class="mt-1 text-xs text-muted-foreground">{{ t('orderDetail.paymentProductNote') }}</div>
                  <div class="text-xs text-muted-foreground mt-1">
                    {{ t('orderDetail.quantityLabel') }}：{{ item.quantity }} · {{ t('orderDetail.itemFulfillmentLabel') }}：{{
                      fulfillmentTypeLabelText(item.fulfillment_type) }}
                  </div>
                  <div v-if="orderItemSkuText(item)" class="text-xs text-muted-foreground mt-1">
                    {{ t('orderDetail.itemSkuLabel') }}：{{ orderItemSkuText(item) }}
                  </div>
                </div>
                <div class="text-xs text-muted-foreground">
                  {{ t('orderDetail.totalPriceLabel') }}：{{ formatMoney(item.total_price, order.currency) }}
                </div>
              </div>
            </div>
          </div>

          <div v-if="paymentResult"
            class="border bg-card text-card-foreground shadow-sm rounded-2xl p-6">
            <h2 class="text-lg font-bold mb-4 text-foreground">{{ t('payment.infoTitle') }}</h2>
            <div class="text-sm text-muted-foreground space-y-2">
              <div>{{ t('payment.methodLabel') }}：{{ resultChannelName }}</div>
              <div>{{ t('payment.interactionLabel') }}：{{ interactionLabel }}</div>
              <div v-if="paymentResult.expires_at">{{ t('payment.expiresAt') }}：{{ formatDate(paymentResult.expires_at)
                }}</div>
            </div>

            <div v-if="showQRCode" class="mt-6 grid grid-cols-1 md:grid-cols-2 gap-6 items-center">
              <div
                class="bg-secondary border rounded-xl p-4 flex items-center justify-center">
                <img :src="qrImageUrl" alt="QR Code" class="w-48 h-48 object-contain" />
              </div>
              <div class="text-sm text-muted-foreground space-y-3">
                <div class="text-foreground font-semibold">{{ paymentGuideTitle }}</div>
                <div>{{ paymentGuideTip }}</div>
                <div v-if="qrUsingPayLinkFallback" class="text-xs text-muted-foreground">
                  {{ t('payment.qrFallbackHint') }}
                </div>
                <div v-if="hasCryptoPaymentDetails" class="space-y-2 rounded-xl border bg-white/5 p-3">
                  <div
                    v-for="item in cryptoPaymentDetails"
                    :key="item.key"
                    class="flex flex-col gap-1 border-b pb-2 last:border-b-0 last:pb-0"
                  >
                    <span class="text-xs text-muted-foreground">{{ item.label }}</span>
                    <span class="min-w-0 font-semibold text-foreground break-all">
                      {{ item.value }}
                      <span v-if="item.detail" class="ml-1 font-normal text-muted-foreground">({{ item.detail }})</span>
                    </span>
                  </div>
                  <div v-if="cryptoWalletAddress" class="flex flex-wrap items-center gap-2 pt-1">
                    <Button variant="secondary" size="sm" class="font-bold" @click="handleCopyWalletAddress">
                      {{ t('payment.copyWalletAddress') }}
                    </Button>
                    <span v-if="walletAddressCopied" class="text-xs text-success">{{ t('payment.copied') }}</span>
                  </div>
                </div>
                <div v-if="paymentResult.pay_url" class="pt-2 flex flex-wrap items-center gap-2">
                  <Button type="button" variant="outline" size="sm" class="font-semibold" @click="handleCopyPayLink">
                    <Copy class="h-4 w-4" aria-hidden="true" />
                    {{ t('payment.copyPayLink') }}
                  </Button>
                  <span v-if="copied" class="text-xs text-success">{{ t('payment.copied') }}</span>
                </div>
              </div>
            </div>

            <div v-if="showPayLink" class="mt-6 flex flex-col md:flex-row md:items-center gap-3">
              <Button type="button" variant="default" class="font-bold" @click="handleOpenPayLink">
                <ExternalLink class="h-4 w-4" aria-hidden="true" />
                {{ t('payment.openPayLink') }}
              </Button>
              <Button type="button" variant="outline" class="font-semibold" @click="handleCopyPayLink">
                <Copy class="h-4 w-4" aria-hidden="true" />
                {{ t('payment.copyPayLink') }}
              </Button>
              <div v-if="showTelegramPayHint" class="text-xs text-muted-foreground">
                {{ t('payment.telegramExternalHint') }}
              </div>
            </div>
          </div>
        </div>

        <div class="h-fit rounded-2xl border bg-card text-card-foreground shadow-sm p-6 lg:sticky lg:top-24">
          <h2 class="text-lg font-bold mb-4 text-foreground">{{ t('payment.methodActionTitle') }}</h2>
          <div v-if="showCountdown" class="text-xs text-muted-foreground mb-3">
            {{ t('payment.countdownLabel') }}：<span class="font-mono">{{ countdownText }}</span>
          </div>
          <Alert
            v-if="paymentAlert"
            :variant="pageAlertVariant(paymentAlert.level)"
            :class="['mb-4', pageAlertToneClass(paymentAlert.level)]"
          >
            <AlertDescription>{{ paymentAlert.message }}</AlertDescription>
          </Alert>

          <div v-if="!configReady" class="mb-4 text-sm text-muted-foreground">
            {{ t('common.loading') }}
          </div>
          <template v-else>
            <div v-if="showBalanceOption" class="mb-4 rounded-xl border bg-secondary p-3">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <div class="text-xs text-muted-foreground">{{ t('payment.walletBalanceLabel') }}</div>
                  <div class="mt-1 text-sm font-semibold text-foreground">
                    {{ walletLoading ? t('common.loading') : walletBalanceDisplay }}
                  </div>
                </div>
                <label class="inline-flex items-center gap-2 text-xs text-muted-foreground">
                  <input v-model="useBalance" type="checkbox" class="h-4 w-4 accent-primary" :disabled="walletOnlyPayment" />
                  <span>{{ t('payment.useBalance') }}</span>
                </label>
              </div>
              <div v-if="walletOnlyPayment" class="mt-3 text-xs text-warning">
                {{ t('payment.walletOnlyHint') }}
              </div>
              <div v-if="useBalance" class="mt-3 space-y-1 text-xs text-muted-foreground">
                <div>{{ t('payment.walletDeductLabel') }}：{{ expectedWalletPaidDisplay }}</div>
                <div v-if="!walletOnlyPayment">{{ t('payment.onlinePayLabel') }}：{{ expectedOnlinePayDisplay }}</div>
                <div v-if="walletOnlyPayment && expectedOnlinePayCents > 0" class="text-warning">
                  {{ t('payment.walletInsufficientHint') }}
                </div>
              </div>
            </div>
            <div v-if="cachedPayment" class="mb-4 rounded-xl border-warning/40 bg-warning/10 p-3 text-sm space-y-2 text-warning">
              <div class="font-semibold">{{ t('payment.cachedTitle') }}</div>
              <div>{{ t('payment.cachedHint', { channel: cachedChannelName }) }}</div>
              <Button variant="secondary" size="sm" class="font-bold" @click="restoreCachedPayment">
                {{ t('payment.useCached') }}
              </Button>
            </div>
            <PaymentChannelSelector
              v-if="!walletOnlyPayment"
              :channels="channels"
              :model-value="selectedChannelId"
              :show-balance-option="showBalanceOption"
              :is-channel-disabled-for-amount="isChannelDisabledForAmount"
              :channel-amount-limit-hint="channelAmountLimitHint"
              @update:model-value="selectedChannelId = $event"
            />
          </template>

          <div
            v-if="configReady && !requiresOnlineChannel && !orderExpired && !orderCanceled"
            class="mb-4 rounded-lg border-success/40 bg-success/10 p-3 text-xs text-success"
          >
            {{ t('payment.walletPayOnly') }}
          </div>
          <div
            v-else-if="configReady && walletOnlyPayment && expectedOnlinePayCents > 0 && !orderExpired && !orderCanceled"
            class="mb-4 rounded-lg border-warning/40 bg-warning/10 p-3 text-xs text-warning"
          >
            {{ t('payment.walletInsufficientHint') }}
          </div>
          <div
            v-else-if="configReady && !walletOnlyPayment && requiresOnlineChannel && !orderExpired && !orderCanceled"
            class="mb-4 rounded-lg border-warning/40 bg-warning/10 p-3 text-xs text-warning"
          >
            {{ t('payment.selectChannelError') }}
          </div>

          <Button class="w-full font-semibold" :disabled="!configReady || !canSubmitPayment" @click="handlePayment">
            {{ submitting ? t('payment.submitting') : t('payment.submitButton') }}
          </Button>
          <Button variant="secondary" class="w-full mt-3 font-semibold" :disabled="loading" @click="handleRefresh">
            {{ t('payment.refreshStatus') }}
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Copy, ExternalLink } from 'lucide-vue-next'
import { pageAlertVariant, pageAlertToneClass } from '../utils/alerts'
import PaymentAmountBreakdown from '../components/payment/PaymentAmountBreakdown.vue'
import PaymentChannelSelector from '../components/payment/PaymentChannelSelector.vue'
import EmptyState from '../components/EmptyState.vue'
import CheckoutSteps from '../components/checkout/CheckoutSteps.vue'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { usePayment } from '../composables/usePayment'

const { t } = useI18n()

const {
  loading, submitting, order, paymentResult, selectedChannelId, copied, walletAddressCopied,
  openedPayWindow, cachedPayment, guestAuth, guestAuthError, walletLoading, useBalance,
  backLink, showGuestAuthForm, walletOnlyPayment, showBalanceOption, configReady, channels,
  cachedChannelName, resultChannelName, cryptoNetworkInfo, interactionLabel,
  paymentResultTitle, paymentGuideTitle, paymentGuideTip, showPayLink, showTelegramPayHint, payLinkOpenedTip,
  isManualCryptoPayment, manualProofStatus, manualTxHashInput, manualProofSubmitting,
  cryptoWalletAddress, cryptoPaymentDetails, hasCryptoPaymentDetails, qrUsingPayLinkFallback, showQRCode, qrImageUrl,
  orderExpired, orderCanceled, paymentAlert, countdownText, showCountdown, showResultView, pollingActive, orderItems,
  customerFeeApplied, customerFeeAmountDisplay, payableAmountDisplay, walletBalanceDisplay,
  expectedWalletPaidDisplay, expectedOnlinePayDisplay, expectedOnlinePayCents, requiresOnlineChannel,
  paymentWalletPaidDisplay, paymentOnlinePayDisplay, isChannelDisabledForAmount, channelAmountLimitHint, canSubmitPayment,
  formatDate, statusLabel, formatMoney, hasDiscountAmount, formatDiscountMoney, getLocalizedText, orderItemSkuText, fulfillmentTypeLabelText,
  handleCopyPayLink, handleCopyWalletAddress, handleOpenPayLink, restoreCachedPayment, handleChangePaymentMethod,
  handlePayment, handleGuestAuthSubmit, handleRefresh, handleSubmitManualCryptoProof,
} = usePayment()
</script>
