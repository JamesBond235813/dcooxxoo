<template>
  <div class="agent-center-page min-h-screen bg-background text-foreground pt-24 pb-16">
    <div class="container mx-auto px-4 max-w-6xl">

      <!-- ==================== NOT LOGGED IN / GUEST STATE: 开通代理账号 & 代理登录 表单 ==================== -->
      <div v-if="!userAuthStore.isAuthenticated" class="max-w-md mx-auto my-4 sm:my-8">
        <div class="rounded-2xl border bg-card p-6 sm:p-8 shadow-sm">

          <!-- Header -->
          <div class="text-left space-y-1 mb-6">
            <h1 class="text-2xl font-bold tracking-tight text-foreground">
              {{ authMode === 'register' ? '开通代理账号' : (authMode === 'login' ? '代理登录' : '找回代理密码') }}
            </h1>
            <p class="text-xs text-muted-foreground leading-relaxed">
              {{ authMode === 'register' ? '使用管理员发放的专属邀请码，填写注册邮箱与密码即可极速开通。' : (authMode === 'login' ? '登录您的代理账号以查看代理底价、社群信息与分销数据。' : '输入您的注册邮箱，获取邮件验证码以安全重置登录密码。') }}
            </p>
          </div>

          <!-- Segmented Tab Switcher -->
          <div v-if="authMode !== 'forgot'" class="grid grid-cols-2 p-1 bg-muted/60 rounded-xl mb-6 text-sm font-semibold">
            <button
              type="button"
              @click="authMode = 'register'; authError = ''"
              class="py-2 rounded-lg transition text-center"
              :class="authMode === 'register' ? 'bg-card text-foreground shadow-xs' : 'text-muted-foreground hover:text-foreground'"
            >
              开通账号
            </button>
            <button
              type="button"
              @click="authMode = 'login'; authError = ''"
              class="py-2 rounded-lg transition text-center"
              :class="authMode === 'login' ? 'bg-card text-foreground shadow-xs' : 'text-muted-foreground hover:text-foreground'"
            >
              代理登录
            </button>
          </div>

          <!-- Tab 1: 开通账号 Form -->
          <form v-if="authMode === 'register'" @submit.prevent="handleRegisterAgent" class="space-y-4">
            <!-- 邀请码 -->
            <div class="space-y-1.5">
              <label class="text-xs font-medium text-foreground flex items-center justify-between">
                <span>邀请码</span>
                <span class="text-[11px] text-amber-600 dark:text-amber-400 font-normal">必填，需向站长获取专属邀请码</span>
              </label>
              <input
                v-model="regForm.inviteCode"
                type="text"
                required
                placeholder="例如：XXXX-XXXX-XXXX-XXXX"
                class="w-full h-10 px-3.5 rounded-xl border bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 text-foreground placeholder:text-muted-foreground uppercase font-mono tracking-wider transition-all"
              />
            </div>

            <!-- 账号 (邮箱) -->
            <div class="space-y-1.5">
              <label class="text-xs font-medium text-foreground flex items-center justify-between">
                <span>账号 (邮箱)</span>
                <span class="text-[11px] text-muted-foreground font-normal">账号必须是有效邮箱</span>
              </label>
              <input
                v-model="regForm.email"
                type="email"
                required
                placeholder="请输入您的注册邮箱"
                class="w-full h-10 px-3.5 rounded-xl border bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 text-foreground placeholder:text-muted-foreground transition-all"
              />
            </div>

            <!-- 密码 -->
            <div class="space-y-1.5">
              <label class="text-xs font-medium text-foreground">密码</label>
              <input
                v-model="regForm.password"
                type="password"
                required
                placeholder="8-72位密码"
                class="w-full h-10 px-3.5 rounded-xl border bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 text-foreground placeholder:text-muted-foreground transition-all"
              />
            </div>

            <!-- 确认密码 -->
            <div class="space-y-1.5">
              <label class="text-xs font-medium text-foreground">二次确认密码</label>
              <input
                v-model="regForm.confirmPassword"
                type="password"
                required
                placeholder="再次输入密码"
                class="w-full h-10 px-3.5 rounded-xl border bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 text-foreground placeholder:text-muted-foreground transition-all"
              />
            </div>

            <!-- 代理联系方式（选填） -->
            <div class="pt-2 space-y-3">
              <div class="text-xs font-semibold text-foreground">代理联系方式（选填）</div>

              <div class="space-y-1.5">
                <label class="text-xs font-medium text-muted-foreground">QQ</label>
                <input
                  v-model="regForm.qq"
                  type="text"
                  placeholder="选填，您的联系 QQ"
                  class="w-full h-10 px-3.5 rounded-xl border bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 text-foreground placeholder:text-muted-foreground transition-all"
                />
              </div>

              <div class="space-y-1.5">
                <label class="text-xs font-medium text-muted-foreground">Telegram</label>
                <input
                  v-model="regForm.telegram"
                  type="text"
                  placeholder="@用户名 或 https://t.me/用户名"
                  class="w-full h-10 px-3.5 rounded-xl border bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 text-foreground placeholder:text-muted-foreground transition-all"
                />
              </div>

              <p class="text-[11px] text-muted-foreground leading-snug">
                绑定联系方式后若发卡等可在客服处核验。
              </p>
            </div>

            <label class="flex items-start gap-2 text-xs text-muted-foreground">
              <input v-model="regForm.agreementAccepted" type="checkbox" required class="mt-0.5" />
              <span>我已阅读并同意<router-link to="/terms" target="_blank" class="text-primary underline">服务条款</router-link>和<router-link to="/privacy" target="_blank" class="text-primary underline">隐私政策</router-link>。</span>
            </label>

            <!-- Error message if any -->
            <div v-if="authError" class="rounded-xl border border-rose-200 bg-rose-50 dark:border-rose-900/50 dark:bg-rose-950/30 p-3 text-xs text-rose-600 dark:text-rose-400">
              {{ authError }}
            </div>

            <!-- Submit button -->
            <button
              type="submit"
              :disabled="isSubmitting"
              class="w-full h-11 rounded-xl bg-gray-900 hover:bg-gray-800 text-white font-semibold text-sm transition active:scale-98 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-100 disabled:opacity-50 mt-4 shadow-sm"
            >
              {{ isSubmitting ? '正在开通...' : '确认开通' }}
            </button>

            <div class="text-center pt-2">
              <button
                type="button"
                @click="authMode = 'login'; authError = ''"
                class="text-xs text-muted-foreground hover:text-foreground transition"
              >
                已有代理账号？点击切换到代理登录
              </button>
            </div>
          </form>

          <!-- Tab 2: 代理登录 Form -->
          <form v-else-if="authMode === 'login'" @submit.prevent="handleLoginAgent" class="space-y-4">
            <!-- 账号 / 邮箱 -->
            <div class="space-y-1.5">
              <label class="text-xs font-medium text-foreground">账号 / 邮箱</label>
              <input
                v-model="loginForm.account"
                type="text"
                required
                placeholder="请输入注册邮箱账号"
                class="w-full h-10 px-3.5 rounded-xl border bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 text-foreground placeholder:text-muted-foreground transition-all"
              />
            </div>

            <!-- 密码 -->
            <div class="space-y-1.5">
              <div class="flex items-center justify-between">
                <label class="text-xs font-medium text-foreground">密码</label>
                <button
                  type="button"
                  @click="authMode = 'forgot'; authError = ''"
                  class="text-[11px] text-emerald-600 dark:text-emerald-400 hover:underline font-medium"
                >
                  忘记密码？
                </button>
              </div>
              <input
                v-model="loginForm.password"
                type="password"
                required
                placeholder="请输入密码"
                class="w-full h-10 px-3.5 rounded-xl border bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 text-foreground placeholder:text-muted-foreground transition-all"
              />
            </div>

            <!-- Error message if any -->
            <div v-if="authError" class="rounded-xl border border-rose-200 bg-rose-50 dark:border-rose-900/50 dark:bg-rose-950/30 p-3 text-xs text-rose-600 dark:text-rose-400">
              {{ authError }}
            </div>

            <!-- Submit button -->
            <button
              type="submit"
              :disabled="isSubmitting"
              class="w-full h-11 rounded-xl bg-gray-900 hover:bg-gray-800 text-white font-semibold text-sm transition active:scale-98 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-100 disabled:opacity-50 mt-4 shadow-sm"
            >
              {{ isSubmitting ? '登录中...' : '代理登录' }}
            </button>

            <div class="text-center pt-2">
              <button
                type="button"
                @click="authMode = 'register'; authError = ''"
                class="text-xs text-muted-foreground hover:text-foreground transition"
              >
                还没有代理账号？点击切换到开通账号
              </button>
            </div>
          </form>

          <!-- Tab 3: 找回密码 Form -->
          <form v-else-if="authMode === 'forgot'" @submit.prevent="handleResetAgentPassword" class="space-y-4">
            <!-- 注册邮箱 -->
            <div class="space-y-1.5">
              <label class="text-xs font-medium text-foreground">注册邮箱</label>
              <input
                v-model="forgotForm.email"
                type="email"
                required
                placeholder="请输入您注册时使用的邮箱"
                class="w-full h-10 px-3.5 rounded-xl border bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 text-foreground placeholder:text-muted-foreground transition-all"
              />
            </div>

            <!-- 邮箱验证码 -->
            <div class="space-y-1.5">
              <label class="text-xs font-medium text-foreground">邮箱验证码</label>
              <div class="flex gap-2">
                <input
                  v-model="forgotForm.code"
                  type="text"
                  required
                  placeholder="6位邮件验证码"
                  class="flex-1 h-10 px-3.5 rounded-xl border bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 text-foreground placeholder:text-muted-foreground transition-all"
                />
                <button
                  type="button"
                  :disabled="sendCodeCountdown > 0 || isSendingCode"
                  @click="handleSendResetCode"
                  class="px-4 h-10 rounded-xl border bg-card text-xs font-semibold text-foreground hover:bg-accent disabled:opacity-50 transition shrink-0"
                >
                  {{ sendCodeCountdown > 0 ? `${sendCodeCountdown}s 后重发` : (isSendingCode ? '发送中...' : '发送验证码') }}
                </button>
              </div>
            </div>

            <!-- 新密码 -->
            <div class="space-y-1.5">
              <label class="text-xs font-medium text-foreground">设置新密码</label>
              <input
                v-model="forgotForm.newPassword"
                type="password"
                required
                placeholder="8-72位新密码"
                class="w-full h-10 px-3.5 rounded-xl border bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 text-foreground placeholder:text-muted-foreground transition-all"
              />
            </div>

            <!-- 确认新密码 -->
            <div class="space-y-1.5">
              <label class="text-xs font-medium text-foreground">确认新密码</label>
              <input
                v-model="forgotForm.confirmPassword"
                type="password"
                required
                placeholder="请再次输入新密码"
                class="w-full h-10 px-3.5 rounded-xl border bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 text-foreground placeholder:text-muted-foreground transition-all"
              />
            </div>

            <!-- Error message if any -->
            <div v-if="authError" class="rounded-xl border border-rose-200 bg-rose-50 dark:border-rose-900/50 dark:bg-rose-950/30 p-3 text-xs text-rose-600 dark:text-rose-400">
              {{ authError }}
            </div>

            <!-- Submit button -->
            <button
              type="submit"
              :disabled="isSubmitting"
              class="w-full h-11 rounded-xl bg-gray-900 hover:bg-gray-800 text-white font-semibold text-sm transition active:scale-98 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-100 disabled:opacity-50 mt-4 shadow-sm"
            >
              {{ isSubmitting ? '正在重置...' : '确认重置密码' }}
            </button>

            <div class="text-center pt-2">
              <button
                type="button"
                @click="authMode = 'login'; authError = ''"
                class="text-xs text-muted-foreground hover:text-foreground transition"
              >
                记起密码了？返回代理登录
              </button>
            </div>
          </form>

        </div>
      </div>

      <!-- ==================== LOGGED IN BUT NOT AGENT UPGRADE PROMPT ==================== -->
      <div v-else-if="agentIdentityLoading" class="max-w-md mx-auto my-8 rounded-2xl border bg-card p-8 text-center text-sm text-muted-foreground" role="status">
        正在确认代理权限...
      </div>
      <div v-else-if="agentIdentityError" class="max-w-md mx-auto my-8 rounded-2xl border bg-card p-8 text-center space-y-4" role="alert">
        <p class="text-sm text-muted-foreground">代理权限暂时无法验证，请重试。</p>
        <button type="button" class="text-sm text-primary hover:underline" @click="retryAgentIdentity">重新验证</button>
      </div>
      <div v-else-if="!isAgent" class="max-w-md mx-auto my-8">
        <div class="rounded-2xl border bg-card p-6 sm:p-8 shadow-sm text-center space-y-4">
          <div class="w-12 h-12 rounded-full bg-emerald-100 text-emerald-600 dark:bg-emerald-950 dark:text-emerald-400 mx-auto flex items-center justify-center">
            <ShieldCheck class="w-6 h-6" />
          </div>
          <h2 class="text-xl font-bold">激活官方代理权限</h2>
          <p class="text-xs text-muted-foreground leading-relaxed">
            当前账号 ({{ userAuthStore.user?.email || userAuthStore.user?.nickname }}) 尚未开通代理权限。代理权限需要使用管理员发放的邀请码注册，不能通过普通账号登录后自行激活。
          </p>
          <div>
            <button
              type="button"
              @click="handleLogout"
              class="text-xs text-muted-foreground hover:text-foreground underline transition"
            >
              退出当前账号
            </button>
          </div>
        </div>
      </div>

      <!-- ==================== LOGGED IN AGENT DASHBOARD (对应截图2) ==================== -->
      <div v-else class="space-y-5 sm:space-y-8">
        <div v-if="isWithdrawRoute" class="space-y-3">
          <router-link to="/agent" class="text-sm font-medium text-emerald-700 dark:text-emerald-400">← {{ t('agentWithdraw.back') }}</router-link>
          <h1 class="text-xl font-bold text-foreground">{{ t('agentWithdraw.title') }}</h1>
          <div class="rounded-2xl border bg-card p-4 shadow-xs">
            <p class="text-xs text-muted-foreground">{{ t('agentWithdraw.available') }}</p>
            <p class="mt-1 text-2xl font-black text-emerald-600 dark:text-emerald-400">¥{{ stats.pendingEarnings }}</p>
          </div>
        </div>
        <!-- Top bar -->
        <div v-if="!isWithdrawRoute" class="flex items-center gap-3 pb-4 border-b">
          <span class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300">
            <span class="h-2 w-2 rounded-full bg-emerald-500"></span>
            官方核心代理
          </span>
          <span class="text-sm font-medium text-muted-foreground">
            {{ userAuthStore.user?.email || userAuthStore.user?.nickname }}
          </span>
        </div>

        <!-- Profit Dashboard (截图2统计卡片) -->
        <div v-if="!isWithdrawRoute" class="grid grid-cols-3 gap-2 sm:gap-4">
          <div class="min-w-0 rounded-xl border bg-card p-2.5 shadow-xs sm:rounded-2xl sm:p-5">
            <div class="text-[10px] font-medium text-muted-foreground sm:text-xs">今日收益 (元)</div>
            <div class="mt-1 break-all text-base font-black text-foreground sm:mt-2 sm:text-2xl">¥{{ stats.todayEarnings }}</div>
            <div class="mt-1 hidden text-[11px] text-muted-foreground sm:block">今日分销佣金</div>
          </div>

          <div class="min-w-0 rounded-xl border bg-card p-2.5 shadow-xs sm:rounded-2xl sm:p-5">
            <div class="text-[10px] font-medium text-muted-foreground sm:text-xs">累计佣金 (元)</div>
            <div class="mt-1 break-all text-base font-black text-emerald-600 dark:text-emerald-400 sm:mt-2 sm:text-2xl">¥{{ stats.totalEarnings }}</div>
            <div class="mt-1 hidden text-[11px] text-muted-foreground sm:block">平台累计分成收入</div>
          </div>

          <div class="min-w-0 rounded-xl border bg-card p-2.5 shadow-xs sm:rounded-2xl sm:p-5">
            <div class="text-[10px] font-medium text-muted-foreground sm:text-xs">待结算金额 (元)</div>
            <div class="mt-1 break-all text-base font-black text-foreground sm:mt-2 sm:text-2xl">¥{{ stats.pendingEarnings }}</div>
            <div class="mt-1 hidden text-[11px] text-muted-foreground sm:block">可申请提现</div>
          </div>
        </div>

        <!-- 手机端提现从头像菜单进入独立页面；桌面端保留原有标签。 -->
        <div v-if="!isWithdrawRoute" class="border-b">
          <nav class="grid grid-cols-4 gap-1 sm:flex sm:space-x-6" aria-label="Tabs">
            <button
              v-for="t in tabs"
              :key="t.key"
              type="button"
              @click="activeTab = t.key"
              class="min-w-0 justify-center whitespace-nowrap py-3 px-0.5 text-[11px] font-semibold border-b-2 transition-all flex items-center sm:gap-1.5 sm:px-1 sm:text-sm"
              :class="[
                activeTab === t.key
                  ? 'border-emerald-600 text-emerald-600 dark:border-emerald-400 dark:text-emerald-400'
                  : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border',
                t.key === 'withdraws' ? '!hidden sm:!flex' : '',
              ]"
            >
              {{ t.label }}
            </button>
          </nav>
        </div>

        <!-- Tab 1: 代理拿货 (按代理底价陈列) -->
        <div v-if="!isWithdrawRoute && activeTab === 'wholesale'" class="space-y-3 sm:space-y-6">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-lg font-bold text-foreground">代理拿货专区</h2>
              <p class="text-xs leading-relaxed text-muted-foreground"><span class="md:hidden">以核心代理底价直接开卡直充<br />付款后自动进入安全充值流程</span><span class="hidden md:inline">以核心代理底价直接开卡直充，付款后自动进入安全充值流程</span></p>
            </div>
            <span class="text-xs text-emerald-600 dark:text-emerald-400 font-semibold bg-emerald-50 dark:bg-emerald-950/60 px-2.5 py-1 rounded-full border border-emerald-500/20">
              已激活专属拿货底价
            </span>
          </div>

          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4 lg:grid-cols-4">
            <div
              v-for="item in visibleAgentProducts"
              :key="item.slug"
              class="flex flex-col justify-between rounded-2xl border bg-card p-3 shadow-xs transition-all hover:border-emerald-500/40 sm:p-5"
            >
              <div class="space-y-2 sm:space-y-3">
                <span class="inline-block rounded-md bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300 px-2.5 py-0.5 text-[11px] font-semibold">
                  官方直充 · 拿货
                </span>
                <div class="flex items-start justify-between gap-2 sm:block">
                  <h3 class="min-w-0 flex-1 text-base font-bold text-foreground">{{ item.title }}</h3>
                  <div class="shrink-0 text-right sm:hidden">
                    <div class="flex items-baseline justify-end gap-1">
                      <span class="text-[10px] font-medium text-muted-foreground">代理价</span>
                      <strong class="text-lg font-black text-emerald-600 dark:text-emerald-400">{{ item.agent ? `¥${item.agent}` : '未设置' }}</strong>
                    </div>
                    <div class="flex justify-end gap-1.5 text-[10px]">
                      <span class="text-muted-foreground line-through">零售价: ¥{{ item.retail || '未设置' }}</span>
                      <span class="font-semibold text-emerald-600 dark:text-emerald-400">{{ item.retail && item.agent ? `省 ¥${(Number(item.retail) - Number(item.agent)).toFixed(0)}` : '' }}</span>
                    </div>
                  </div>
                </div>
                <p class="text-xs text-muted-foreground leading-relaxed">{{ t(item.descriptionKey) }}</p>

                <div class="flex items-stretch gap-2 sm:hidden">
                  <div class="min-w-0 flex-1 rounded-xl bg-muted/60 p-2.5 text-[11px] text-muted-foreground">
                    <div class="font-semibold text-foreground">{{ t(item.requirementTitleKey) }}</div>
                    <p class="mt-1 leading-relaxed">{{ t(item.requirementKey) }}</p>
                  </div>
                  <button
                    type="button"
                    @click="handleWholesaleBuy(item.slug)"
                    :disabled="!item.agent || !currentAgentCode"
                    class="flex w-[76px] shrink-0 flex-col items-center justify-center rounded-xl bg-primary px-2 text-xs font-bold leading-5 text-primary-foreground shadow-xs transition hover:opacity-90 disabled:opacity-50"
                  >
                    <span>立即</span><span>拿货</span>
                  </button>
                </div>
                <div class="hidden rounded-lg bg-muted/60 p-2 text-[11px] text-muted-foreground sm:block">
                  {{ t(item.requirementTitleKey) }}：{{ t(item.requirementKey) }}
                </div>
              </div>

              <div class="mt-3 hidden space-y-2 border-t pt-3 sm:mt-5 sm:block sm:space-y-3 sm:pt-4">
                <div class="flex items-baseline justify-between">
                  <span class="text-xs text-muted-foreground line-through">普通零售: ¥{{ item.retail || '未设置' }}</span>
                  <span class="text-xs font-semibold text-emerald-600 dark:text-emerald-400 bg-emerald-500/10 px-2 py-0.5 rounded-full">
                    {{ item.retail && item.agent ? `省 ¥${(Number(item.retail) - Number(item.agent)).toFixed(0)}` : '价格待设置' }}
                  </span>
                </div>
                <div class="flex items-baseline gap-1">
                  <span class="text-xl font-black text-emerald-600 dark:text-emerald-400 sm:text-2xl">{{ item.agent ? `¥${item.agent}` : '拿货价未设置' }}</span>
                  <span class="text-xs text-muted-foreground">/ 份 (代理底价)</span>
                </div>

                <button
                  type="button"
                  @click="handleWholesaleBuy(item.slug)"
                  :disabled="!item.agent || !currentAgentCode"
                  class="w-full rounded-xl bg-gray-900 py-2 text-xs font-bold text-white shadow-xs transition-all hover:bg-gray-800 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-100 sm:py-2.5"
                >
                  立即拿货
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Tab 2: 专属销售链接与售价设置 -->
        <div v-else-if="activeTab === 'pricing'" class="space-y-3 sm:space-y-6">
          <!-- 专属销售链接卡片 -->
          <div class="space-y-3 rounded-2xl border border-emerald-100 bg-gradient-to-br from-emerald-50/80 via-card to-sky-50/50 p-4 shadow-xs dark:border-emerald-900/40 dark:from-emerald-950/20 dark:to-card sm:space-y-4 sm:border-border sm:bg-none sm:bg-card sm:p-6">
            <div class="flex items-start justify-between gap-2 sm:flex-row sm:items-center sm:gap-4">
              <div class="min-w-0">
                <h3 class="text-base font-bold text-foreground">专属轻量化销售链接</h3>
                <p class="hidden text-xs text-muted-foreground mt-1 sm:block">
                  将此链接发给客户，客户打开即展示您设置的最新销售价格。客户自助下单付款后，差价利润将自动结算到您的余额。
                </p>
              </div>
              <span class="inline-flex shrink-0 items-center gap-1 text-[10px] font-mono font-semibold px-2 py-1 rounded-full bg-emerald-100/70 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300 sm:text-[11px] sm:px-2.5 sm:rounded-md sm:bg-muted sm:text-foreground">
                代理代码: {{ currentAgentCode }}
              </span>
            </div>
            <p v-if="agentProfileError" class="text-xs text-rose-600" role="alert">{{ agentProfileError }}</p>

            <div class="flex flex-col gap-3 sm:flex-row sm:gap-2">
              <input
                :value="agentShareLink"
                readonly
                class="flex-1 h-10 px-3.5 rounded-xl border bg-muted/40 font-mono text-xs text-foreground focus:outline-none"
              />
              <div class="grid grid-cols-3 gap-1.5 sm:flex sm:flex-wrap sm:gap-2">
                <button
                  type="button"
                  @click="copyText(agentShareLink, '专属销售链接')"
                  :disabled="!agentShareLink"
                  class="justify-center px-1.5 h-10 rounded-xl bg-gray-900 text-white dark:bg-white dark:text-gray-900 text-[11px] font-semibold hover:bg-gray-800 transition shrink-0 shadow-xs flex items-center gap-1 sm:px-4 sm:text-xs sm:gap-1.5"
                >
                  <Copy class="w-3.5 h-3.5" />
                  <span>复制链接</span>
                </button>
                <button
                  type="button"
                  @click="openQrModal"
                  :disabled="!agentShareLink"
                  class="justify-center px-1.5 h-10 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white text-[11px] font-semibold transition shrink-0 shadow-xs flex items-center gap-1 sm:px-4 sm:text-xs sm:gap-1.5"
                >
                  <QrCode class="w-3.5 h-3.5" />
                  <span>推广二维码</span>
                </button>
                <button
                  type="button"
                  @click="copyPromotionText"
                  :disabled="!agentShareLink"
                  class="justify-center px-1.5 h-10 rounded-xl border border-input bg-background hover:bg-muted text-foreground text-[11px] font-semibold transition shrink-0 flex items-center gap-1 sm:px-4 sm:text-xs sm:gap-1.5"
                >
                  <Share2 class="w-3.5 h-3.5" />
                  <span>复制推广文案</span>
                </button>
              </div>
            </div>
          </div>

          <!-- 套餐售价设置卡片 -->
          <div class="space-y-3 rounded-2xl border bg-card p-3 shadow-xs sm:space-y-6 sm:p-6">
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b pb-4">
              <div>
                <h3 class="text-base font-bold text-foreground">商品自定义销售价设置</h3>
                <p class="text-xs text-muted-foreground mt-1">
                  每个商品单独设置和保存。保存成功后，该商品的新订单使用新售价；已创建的订单价格不变。
                </p>
              </div>
              <div class="flex items-center gap-2">
                <button
                  type="button"
                  @click="resetToSuggestedPrices"
                  class="px-3 py-1.5 rounded-lg border text-xs text-muted-foreground hover:text-foreground transition"
                >
                  恢复建议价（仅修改未保存内容）
                </button>
              </div>
            </div>

            <!-- 4 个产品的定价设置表单 -->
            <div v-if="pricingSettingsError" class="rounded-lg border border-rose-300 bg-rose-50 px-3 py-2 text-xs text-rose-700" role="alert">
              {{ pricingSettingsError }}
            </div>
            <div class="grid grid-cols-1 gap-3 sm:gap-4 md:grid-cols-2">
              <div
                v-for="prod in visiblePricingProducts"
                :key="prod.id"
                class="space-y-2 rounded-xl border bg-muted/20 p-3 sm:space-y-3 sm:p-4"
              >
                <div class="flex items-start justify-between">
                  <div>
                    <div class="font-bold text-sm text-foreground">{{ prod.title }}</div>
                    <div class="text-[11px] text-muted-foreground">{{ t(prod.descriptionKey) }}</div>
                  </div>
                  <span class="text-[11px] font-mono bg-muted px-2 py-0.5 rounded text-muted-foreground">
                    底价 ¥{{ prod.basePrice }}
                  </span>
                </div>

                <div class="grid grid-cols-2 gap-3 pt-1">
                  <div class="space-y-1">
                    <label class="text-[11px] font-medium text-muted-foreground">我的销售价 (元)</label>
                    <div class="relative">
                      <span class="absolute left-3 top-2.5 text-xs text-muted-foreground">¥</span>
                      <input
                        v-model="prod.customPrice"
                        type="number"
                        step="0.01"
                        :min="prod.basePrice"
                        :disabled="isSavingPrice(prod.id)"
                        :aria-label="prod.title + '销售价，人民币'"
                        class="w-full h-9 pl-7 pr-3 rounded-lg border bg-background text-sm font-bold font-mono focus:outline-none focus:ring-1 focus:ring-emerald-500"
                        @input="clearPriceFeedback(prod.id)"
                      />
                    </div>
                  </div>

                  <div class="space-y-1">
                    <label class="text-[11px] font-medium text-muted-foreground">每单预估利润</label>
                    <div
                      class="h-9 px-3 rounded-lg border flex items-center font-mono font-bold text-sm"
                      :class="calculateProfit(prod) >= 0 ? 'bg-emerald-50 text-emerald-600 dark:bg-emerald-950/40 dark:text-emerald-400 border-emerald-200 dark:border-emerald-800/50' : 'bg-rose-50 text-rose-600 border-rose-200'"
                    >
                      {{ isValidProductPrice(prod) ? '+¥' + calculateProfit(prod).toFixed(2) : '请输入有效售价' }}
                    </div>
                  </div>
                </div>

                <div class="flex flex-wrap items-center justify-between gap-2 text-[11px] text-muted-foreground pt-1">
                  <span>官网参考价：¥{{ Number(prod.suggestedPrice).toFixed(2) }}</span>
                  <span v-if="!isValidProductPrice(prod)" class="text-rose-500 font-medium">售价至少为拿货底价 ¥{{ Number(prod.basePrice).toFixed(2) }}</span>
                </div>
                <div class="flex flex-wrap items-center justify-between gap-2 border-t pt-3">
                  <span v-if="priceFeedback[prod.id]" class="text-xs" :class="priceFeedback[prod.id]?.kind === 'success' ? 'text-emerald-600' : 'text-rose-600'" role="status">
                    {{ priceFeedback[prod.id]?.message }}
                  </span>
                  <span v-else class="text-xs text-muted-foreground">
                    {{ pricingSettingsLoading ? '正在读取当前售价…' : (isProductPriceDirty(prod) ? '有未保存的售价修改' : '售价已保存') }}
                  </span>
                  <button
                    type="button"
                    :disabled="pricingSettingsLoading || !!pricingSettingsError || isSavingPrice(prod.id) || !isValidProductPrice(prod) || !isProductPriceDirty(prod)"
                    @click="saveProductPrice(prod)"
                    class="ml-auto min-w-24 rounded-lg bg-emerald-600 px-4 py-2 text-xs font-bold text-white transition hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    {{ isSavingPrice(prod.id) ? '保存中…' : '保存此商品售价' }}
                  </button>
                </div>
              </div>
            </div>

            <!-- 底部保存提醒与快捷操作 -->
            <div class="pt-3 border-t flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs text-muted-foreground">
              <div class="flex items-center gap-1.5">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
                <span>每个商品分别保存；请确认卡片显示“售价已保存”后再分享销售链接。</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Tab 3: 拿货订单 (代理自己代充的订单) -->
        <div v-else-if="activeTab === 'orders'" class="space-y-4">
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-base font-bold text-foreground">代理拿货订单</h3>
            </div>
            <button
              type="button"
              @click="refreshAgentOrders"
              class="text-xs text-emerald-600 dark:text-emerald-400 hover:underline"
            >
              刷新订单
            </button>
          </div>

          <div v-if="ordersLoading" class="py-12 text-center text-xs text-muted-foreground">
            正在加载订单数据...
          </div>
          <div v-else-if="ordersError" class="py-8 text-center text-sm text-rose-600" role="alert">{{ ordersError }}</div>
          <div v-else-if="agentOrders.length === 0" class="py-12 text-center rounded-2xl border border-dashed text-xs text-muted-foreground">
            暂无拿货订单记录
          </div>
          <div v-else class="space-y-3">
            <router-link
              v-for="ord in agentOrders"
              :key="ord.id || ord.order_no"
              :to="orderDestination(ord)"
              class="p-4 rounded-xl border bg-card flex flex-col sm:flex-row sm:items-center justify-between gap-2 sm:gap-3 text-xs transition hover:border-primary/50 hover:shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
            >
              <div class="space-y-1 md:flex-1">
                <div class="flex items-center justify-between gap-2 md:flex">
                  <div class="font-mono font-semibold text-foreground">订单号: {{ ord.order_no }}</div>
                  <time class="shrink-0 text-[10px] text-muted-foreground md:hidden">{{ formatOrderDate(ord.created_at) }}</time>
                  <time class="hidden shrink-0 text-[10px] text-muted-foreground md:inline">{{ formatOrderDate(ord.created_at) }}</time>
                </div>
                <div class="hidden text-muted-foreground sm:block">{{ ord.product_title || 'ChatGPT 充值' }}</div>
              </div>
              <div class="flex items-center gap-2 sm:gap-4">
                <span class="min-w-0 flex-1 truncate text-muted-foreground sm:hidden">{{ ord.product_title || 'ChatGPT 充值' }}</span>
                <span class="font-bold text-foreground">¥{{ ord.total_amount }}</span>
                <span class="px-2 py-0.5 rounded text-[11px] bg-secondary text-secondary-foreground font-medium">
                  {{ displayOrderStatus(ord.status) }}
                </span>
                <span class="font-medium text-primary">{{ isSessionPending(ord) ? '继续提交 Session' : hasSessionFollowUp(ord) ? '查看 Session 进度' : ord.status === 'pending_payment' ? '继续付款' : '查看详情' }}</span>
              </div>
            </router-link>
          </div>

        </div>

        <!-- Tab 4: 客户销售订单 (客户通过链接自助购买的订单) -->
        <div v-else-if="activeTab === 'sales'" class="space-y-4">
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-base font-bold text-foreground">客户专属链接订单</h3>
              <p class="hidden text-xs text-muted-foreground sm:block">客户通过您的专属链接下单并支付的记录，差价利润自动结算到您的余额中</p>
            </div>
            <button
              type="button"
              @click="loadSalesOrders"
              class="text-xs text-emerald-600 dark:text-emerald-400 hover:underline"
            >
              刷新销售订单
            </button>
          </div>

          <div v-if="salesLoading" class="py-12 text-center text-xs text-muted-foreground">
            正在加载客户销售订单...
          </div>
          <div v-else-if="salesError" class="py-8 text-center text-sm text-rose-600" role="alert">{{ salesError }}</div>
          <div v-else-if="customerSalesOrders.length === 0" class="py-12 text-center rounded-2xl border border-dashed text-xs text-muted-foreground">
            <span class="block text-sm font-semibold text-foreground sm:hidden">{{ t('mobileCheckout.noCustomerOrders') }}</span>
            <span class="mt-1 block sm:hidden">{{ t('mobileCheckout.shareHint') }}</span>
            <span class="hidden sm:inline">暂无客户销售订单。您可以将专属销售链接发给客户，客户自助购买后订单与利润明细将在此实时显示。</span>
          </div>
          <div v-else class="space-y-3">
            <div
              v-for="ord in customerSalesOrders"
              :key="ord.order_no || ord.id"
              class="p-4 rounded-xl border bg-card flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs"
            >
              <div class="space-y-1">
                <div class="flex items-center gap-2">
                  <span class="font-mono font-semibold text-foreground">订单号: {{ ord.order_no }}</span>
                  <span class="px-2 py-0.5 rounded text-[10px] bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300 font-medium">
                    客户自付
                  </span>
                </div>
                <div class="text-muted-foreground">客户邮箱: {{ ord.buyer_email || ord.email || '游客' }}</div>
              </div>
              <div class="flex items-center gap-4">
                <div class="text-right">
                  <div class="text-xs text-muted-foreground">客户支付 ¥{{ ord.total_amount }}</div>
                  <div class="font-bold text-emerald-600 dark:text-emerald-400">
                    利润 +¥{{ ord.reseller_profit_amount || ord.profit_amount || '0.00' }}
                  </div>
                </div>
                <span class="px-2.5 py-1 rounded text-[11px] bg-secondary text-secondary-foreground font-medium">
                  {{ displayOrderStatus(ord.status) }}
                </span>
                <RouterLink
                  v-if="ord.session_pending || hasSessionFollowUp(ord)"
                  :to="orderDestination(ord)"
                  class="rounded-lg border border-primary/20 bg-primary/5 px-3 py-2 font-semibold text-primary hover:bg-primary/10"
                >
                  {{ ord.session_pending || hasPendingSession(ord) ? '继续提交 Session' : '查看 Session 进度' }}
                </RouterLink>
              </div>
            </div>
          </div>
        </div>

        <!-- Tab 5: 佣金提现与结算 -->
        <div v-else-if="isWithdrawRoute || activeTab === 'withdraws'" class="space-y-6" :class="isWithdrawRoute ? '' : 'hidden sm:block'">
          <!-- 提现申请卡片 -->
          <div class="rounded-2xl border bg-card p-6 space-y-4 shadow-xs">
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b pb-4">
              <div>
                <h3 class="text-base font-bold text-foreground">申请提现佣金</h3>
                <p class="text-xs text-muted-foreground mt-0.5">
                  当前可提现金额: <span class="font-bold font-mono text-emerald-600 dark:text-emerald-400 text-sm">¥{{ stats.pendingEarnings }}</span>
                </p>
              </div>
              <a
                href="https://telegram.me/ooxxoodc"
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex items-center gap-1 text-xs text-blue-600 dark:text-blue-400 hover:underline"
              >
                <span>也可直接联系站长结算</span>
                <ExternalLink class="w-3.5 h-3.5" />
              </a>
            </div>

            <form @submit.prevent="handleApplyWithdraw" class="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div class="space-y-1">
                <label class="text-xs font-medium text-muted-foreground">提现渠道</label>
                <select
                  v-model="withdrawForm.channel"
                  class="w-full h-10 px-3 rounded-xl border bg-background text-xs focus:outline-none focus:ring-1 focus:ring-emerald-500"
                >
                  <option value="alipay">支付宝</option>
                  <option value="wechat">微信支付</option>
                  <option value="usdt">USDT-TRC20</option>
                </select>
              </div>

              <div class="space-y-1">
                <label class="text-xs font-medium text-muted-foreground">收款账号 / 姓名</label>
                <input
                  v-model="withdrawForm.account"
                  type="text"
                  required
                  placeholder="例如：支付宝账号或姓名"
                  class="w-full h-10 px-3.5 rounded-xl border bg-background text-xs focus:outline-none focus:ring-1 focus:ring-emerald-500"
                />
              </div>

              <div class="space-y-1">
                <label class="text-xs font-medium text-muted-foreground">提现金额 (元)</label>
                <div class="flex gap-2">
                  <input
                    v-model="withdrawForm.amount"
                    type="number"
                    step="0.01"
                    required
                    :placeholder="stats.pendingEarnings"
                    class="flex-1 h-10 px-3.5 rounded-xl border bg-background text-xs font-mono font-bold focus:outline-none focus:ring-1 focus:ring-emerald-500"
                  />
                  <button
                    type="submit"
                    :disabled="isSubmittingWithdraw"
                    class="px-4 h-10 rounded-xl bg-gray-900 text-white dark:bg-white dark:text-gray-900 text-xs font-bold hover:bg-gray-800 transition disabled:opacity-50 shrink-0"
                  >
                    {{ isSubmittingWithdraw ? '提交中...' : '提交申请' }}
                  </button>
                </div>
              </div>
            </form>
          </div>

          <!-- 提现历史列表 -->
          <div class="space-y-3">
            <h3 class="text-base font-bold text-foreground">提现申请历史</h3>
            <div v-if="withdrawsLoading" class="py-12 text-center text-xs text-muted-foreground">
              正在加载提现记录...
            </div>
            <div v-else-if="withdrawsError" class="py-8 text-center text-sm text-rose-600" role="alert">{{ withdrawsError }}</div>
            <div v-else-if="withdrawRecords.length === 0" class="py-12 text-center rounded-2xl border border-dashed text-xs text-muted-foreground">
              暂无提现申请记录
            </div>
            <div v-else class="space-y-2">
              <div
                v-for="w in withdrawRecords"
                :key="w.id"
                class="p-4 rounded-xl border bg-card flex items-center justify-between text-xs"
              >
                <div>
                  <div class="font-bold text-foreground">¥{{ w.amount }} ({{ w.channel || '在线提现' }})</div>
                  <div class="text-muted-foreground text-[11px] mt-0.5">账号: {{ w.account_info || w.target_account }} · {{ w.created_at }}</div>
                </div>
                <span
                  class="px-2 py-0.5 rounded text-[11px] font-medium"
                  :class="w.status === 'completed' ? 'bg-emerald-100 text-emerald-800' : w.status === 'rejected' ? 'bg-rose-100 text-rose-800' : 'bg-amber-100 text-amber-800'"
                >
                  {{ w.status === 'completed' ? '已到账' : w.status === 'rejected' ? '已拒绝' : '审核处理中' }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Community / Group Cards (截图2顶部卡片) -->
        <div v-if="!isWithdrawRoute" class="grid grid-cols-2 gap-2 md:gap-4">
          <!-- Telegram Group Card -->
          <a
            href="https://t.me/+PWfyt1g1bJQ3YWRh"
            target="_blank"
            rel="noopener noreferrer"
            class="flex flex-col justify-between rounded-2xl border bg-card p-3 shadow-xs transition hover:border-blue-300 hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 md:p-6"
            aria-label="加入 Telegram 官方交流群（在新标签页打开）"
          >
            <div class="flex items-center gap-2 md:hidden">
              <svg class="h-6 w-6 shrink-0 text-sky-500" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm5.894 8.221-1.97 9.28c-.145.658-.537.818-1.084.508l-3-2.21-1.446 1.394c-.14.18-.357.295-.6.295-.002 0-.003 0-.005 0l.213-3.054 5.56-5.022c.24-.213-.054-.334-.373-.121l-6.869 4.326-2.96-.924c-.64-.203-.658-.64.135-.954l11.566-4.458c.538-.196 1.006.128.832.941z" /></svg>
              <span class="text-xs font-semibold text-foreground">{{ t('mobileCheckout.telegramGroup') }}</span>
            </div>
            <div class="hidden items-start justify-between gap-4 md:flex">
              <div>
                <span class="inline-block px-2 py-0.5 rounded text-[11px] font-medium bg-blue-100 text-blue-700 dark:bg-blue-950 dark:text-blue-300 mb-2">
                  官方交流群
                </span>
                <h3 class="text-base font-bold text-foreground">Telegram 官方交流群</h3>
                <p class="text-xs text-muted-foreground mt-1">
                  https://t.me/+PWfyt1g1bJQ3YWRh
                </p>
              </div>
              <div class="w-10 h-10 rounded-full bg-blue-50 dark:bg-blue-950 flex items-center justify-center text-blue-500 shrink-0">
                <ExternalLink class="w-5 h-5" />
              </div>
            </div>
            <div class="mt-4 hidden border-t pt-3 md:block">
              <span class="inline-flex items-center gap-1.5 text-xs font-semibold text-blue-600 dark:text-blue-400">
                <span>立即加入 Telegram 交流群</span>
                <ExternalLink class="w-3.5 h-3.5" />
              </span>
            </div>
          </a>

          <!-- QQ Group Card (暂不设置，置灰禁用) -->
          <div class="flex flex-col justify-between rounded-2xl border bg-muted/40 p-3 shadow-xs opacity-60 cursor-not-allowed select-none md:p-6">
            <div class="flex items-center gap-2 md:hidden">
              <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-sky-500 text-[9px] font-black text-white" aria-hidden="true">QQ</span>
              <span class="text-xs font-semibold text-foreground">{{ t('mobileCheckout.qqGroup') }}</span>
              <span class="ml-auto text-[10px] text-muted-foreground">暂未开启</span>
            </div>
            <div class="hidden items-start justify-between gap-4 md:flex">
              <div>
                <span class="inline-block px-2 py-0.5 rounded text-[11px] font-medium bg-muted text-muted-foreground mb-2">
                  国内交流群 · 暂未开放
                </span>
                <h3 class="text-base font-bold text-muted-foreground">QQ 官方代理交流群</h3>
                <p class="text-xs text-muted-foreground mt-1">
                  官方QQ交流群：<span>暂未开启</span>
                </p>
              </div>
              <div class="w-10 h-10 rounded-full bg-muted flex items-center justify-center text-muted-foreground shrink-0">
                <Users class="w-5 h-5" />
              </div>
            </div>
            <div class="mt-4 hidden items-center justify-between border-t border-border/50 pt-3 md:flex">
              <span class="text-xs text-muted-foreground">群号：暂未设置</span>
              <button
                type="button"
                disabled
                class="inline-flex items-center gap-1 text-xs font-semibold text-muted-foreground cursor-not-allowed opacity-50"
              >
                <Copy class="w-3.5 h-3.5" />
                <span>暂不可用</span>
              </button>
            </div>
          </div>
        </div>
      </div>

    </div>

    <!-- Quick Buy Modal for wholesale purchase -->
    <ProductQuickBuy
      v-if="quickBuyProduct"
      :product="quickBuyProduct"
      :visible="quickBuyVisible"
      agent-mode
      @update:visible="quickBuyVisible = $event"
    />

    <!-- QR Code Share Modal -->
    <div
      v-if="qrModalVisible"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs animate-in fade-in duration-200"
      @click.self="qrModalVisible = false"
    >
      <div class="relative w-full max-w-sm rounded-3xl border bg-card p-6 shadow-2xl space-y-5 animate-in zoom-in-95 duration-200 text-center">
        <!-- Close Button -->
        <button
          type="button"
          @click="qrModalVisible = false"
          class="absolute top-4 right-4 w-8 h-8 rounded-full bg-muted flex items-center justify-center text-muted-foreground hover:text-foreground transition"
        >
          <X class="w-4 h-4" />
        </button>

        <div class="space-y-1 pt-1">
          <div class="inline-flex items-center justify-center w-10 h-10 rounded-full bg-emerald-50 dark:bg-emerald-950 text-emerald-600 dark:text-emerald-400 mb-1">
            <QrCode class="w-5 h-5" />
          </div>
          <h3 class="text-base font-bold text-foreground">专属推广二维码</h3>
          <p class="text-xs text-muted-foreground">扫码即可直接访问并享受您的专属销售价</p>
        </div>

        <!-- QR Image Display -->
        <div class="flex flex-col items-center justify-center p-4 bg-white dark:bg-white rounded-2xl border shadow-inner">
          <img
            v-if="qrCodeDataUrl"
            :src="qrCodeDataUrl"
            :alt="siteDisplayName + '专属推广二维码'"
            class="w-52 h-52 object-contain"
          />
          <div v-else class="w-52 h-52 flex items-center justify-center text-xs text-gray-400">
            生成中...
          </div>
          <p class="mt-2 text-[11px] leading-relaxed text-center text-gray-600">
            用户自己的账号 · 官方直充 · 无需提供账号密码
          </p>
        </div>

        <div v-if="qrPriceError" class="rounded-xl border border-rose-200 bg-rose-50 px-3 py-2 text-xs leading-relaxed text-rose-700" role="alert">
          {{ qrPriceError }}
          <button type="button" class="ml-1 font-semibold underline" @click="loadQrSharePrices">重新读取售价</button>
        </div>

        <!-- Action Buttons: Save Image & Share / Forward -->
        <div class="grid grid-cols-2 gap-3 pt-1">
          <button
            type="button"
            @click="downloadQrCode"
            :disabled="qrPriceLoading || !!qrPriceError"
            class="w-full py-2.5 rounded-xl bg-gray-900 hover:bg-gray-800 text-white dark:bg-white dark:text-gray-900 dark:hover:bg-gray-100 text-xs font-semibold transition flex items-center justify-center gap-1.5 shadow-xs disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Download class="w-3.5 h-3.5" />
            <span>{{ qrPriceLoading ? '正在读取售价…' : '保存推广海报' }}</span>
          </button>
          <button
            type="button"
            @click="shareQrCode"
            class="w-full py-2.5 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-semibold transition flex items-center justify-center gap-1.5 shadow-xs"
          >
            <Share2 class="w-3.5 h-3.5" />
            <span>转发分享</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { orderStatusLabel } from '../utils/status'
import { hasPendingSession, hasSessionFollowUp } from '../utils/sessionOrder'
import QRCode from 'qrcode'
import { api } from '../api/client'
import { useUserAuthStore } from '../stores/userAuth'
import { useAppStore } from '../stores/app'
import { productAPI } from '../api/product'
import { affiliateAPI } from '../api/affiliate'
import { userOrderAPI } from '../api/order'
import { resellerAPI } from '../api/reseller'
import ProductQuickBuy from '../components/ProductQuickBuy.vue'
import { getImageUrl } from '../utils/image'
import {
  ShieldCheck, Users, Copy, ExternalLink, QrCode, Share2, Download, X,
} from 'lucide-vue-next'

const userAuthStore = useUserAuthStore()
const appStore = useAppStore()
const route = useRoute()
const { t } = useI18n()
const displayOrderStatus = (status?: string) => orderStatusLabel(t, status)
const isSessionPending = (order: any) => Boolean(order?.session_pending || hasPendingSession(order))
const isWithdrawRoute = computed(() => route.name === 'agent-withdraw')

// Auth Form Mode ('register' | 'login' | 'forgot')
const authMode = ref<'register' | 'login' | 'forgot'>('login')
const isSubmitting = ref(false)
const isSendingCode = ref(false)
const sendCodeCountdown = ref(0)
const authError = ref('')

// Form Models
const regForm = ref({
  inviteCode: '',
  email: '',
  password: '',
  confirmPassword: '',
  agreementAccepted: false,
  qq: '',
  telegram: '',
})

const loginForm = ref({
  account: '',
  password: '',
})

const forgotForm = ref({
  email: '',
  code: '',
  newPassword: '',
  confirmPassword: '',
})

// Tab state for Logged In Agent
const activeTab = ref<'wholesale' | 'pricing' | 'orders' | 'sales' | 'withdraws'>('wholesale')

const tabs = [
  { key: 'wholesale' as const, label: '代理拿货' },
  { key: 'pricing' as const, label: '销售链接与定价' },
  { key: 'orders' as const, label: '拿货订单' },
  { key: 'sales' as const, label: '客户订单' },
  { key: 'withdraws' as const, label: '提现结算' },
]

// 4 Agent Products (wholesale view)
const agentProducts = ref([
  {
    slug: 'upstream-1-1109423947-1791826450925',
    title: 'ChatGPT Go · 1个月',
    descriptionKey: 'productCardCopy.goDescription',
    requirementTitleKey: 'productCardCopy.accountRequirement',
    requirementKey: 'productCardCopy.goRequirement',
    retail: '',
    agent: '',
  },
  {
    slug: 'chatgpt-plus-1m',
    title: 'ChatGPT Plus · 1个月',
    descriptionKey: 'productCardCopy.subscriptionWarranty',
    requirementTitleKey: 'productCardCopy.accountRequirement',
    requirementKey: 'productCardCopy.plusRequirement',
    retail: '',
    agent: '',
  },
  {
    slug: 'chatgpt-pro-5x-1m',
    title: 'ChatGPT Pro 5x · 1个月',
    descriptionKey: 'productCardCopy.subscriptionWarranty',
    requirementTitleKey: 'productCardCopy.accountRequirement',
    requirementKey: 'productCardCopy.pro5Requirement',
    retail: '',
    agent: '',
  },
  {
    slug: 'chatgpt-pro-20x-1m',
    title: 'ChatGPT Pro 20x · 1个月',
    descriptionKey: 'productCardCopy.subscriptionWarranty',
    requirementTitleKey: 'productCardCopy.accountRequirement',
    requirementKey: 'productCardCopy.pro20Requirement',
    retail: '',
    agent: '',
  },
  {
    slug: 'chatgpt-pro-50x-1m',
    title: 'ChatGPT Pro 50x · 1个月',
    descriptionKey: 'productCardCopy.subscriptionWarranty',
    requirementTitleKey: 'productCardCopy.accountRequirement',
    requirementKey: 'productCardCopy.pro50Requirement',
    retail: '',
    agent: '',
  },
  {
    slug: 'openai-sms-activate',
    title: 'OpenAI 接码台',
    descriptionKey: 'productCardCopy.smsDescription',
    requirementTitleKey: 'productCardCopy.deliveryRule',
    requirementKey: 'productCardCopy.smsRequirement',
    retail: '',
    agent: '',
  },
])

// 4 Pricing Products for Custom Reseller Pricing
const pricingProducts = ref([
  {
    id: 8,
    skuId: 8,
    slug: 'upstream-1-1109423947-1791826450925',
    title: 'ChatGPT Go · 1个月',
    descriptionKey: 'productCardCopy.goDescription',
    basePrice: 0,
    suggestedPrice: 0,
    customPrice: '',
  },
  {
    id: 1,
    skuId: 1,
    slug: 'chatgpt-plus-1m',
    title: 'ChatGPT Plus · 1个月',
    descriptionKey: 'productCardCopy.subscriptionWarranty',
    basePrice: 0,
    suggestedPrice: 0,
    customPrice: '',
  },
  {
    id: 2,
    skuId: 2,
    slug: 'chatgpt-pro-5x-1m',
    title: 'ChatGPT Pro 5x · 1个月',
    descriptionKey: 'productCardCopy.subscriptionWarranty',
    basePrice: 0,
    suggestedPrice: 0,
    customPrice: '',
  },
  {
    id: 3,
    skuId: 3,
    slug: 'chatgpt-pro-20x-1m',
    title: 'ChatGPT Pro 20x · 1个月',
    descriptionKey: 'productCardCopy.subscriptionWarranty',
    basePrice: 0,
    suggestedPrice: 0,
    customPrice: '',
  },
  {
    id: 4,
    skuId: 4,
    slug: 'chatgpt-pro-50x-1m',
    title: 'ChatGPT Pro 50x · 1个月',
    descriptionKey: 'productCardCopy.subscriptionWarranty',
    basePrice: 0,
    suggestedPrice: 0,
    customPrice: '',
  },
  {
    id: 20,
    skuId: 2001,
    slug: 'openai-sms-activate',
    title: 'OpenAI 接码台',
    descriptionKey: 'productCardCopy.smsDescription',
    basePrice: 0,
    suggestedPrice: 0,
    customPrice: '',
  },
])

const currentAgentCode = ref('')
const agentProfileError = ref('')
const savingPriceIds = ref<number[]>([])
const savedPriceById = ref<Record<number, string>>({})
const wholesalePricesBySlug = ref<Record<string, Record<number, string>>>({})
const pricingSettingsLoading = ref(false)
const pricingSettingsError = ref('')
const priceFeedback = ref<Record<number, { kind: 'success' | 'error'; message: string }>>({})

const priceToCents = (value: unknown): number | null => {
  const text = String(value ?? '').trim()
  if (!/^\d+(?:\.\d{1,2})?$/.test(text)) return null
  const amount = Number(text)
  if (!Number.isFinite(amount) || amount <= 0) return null
  return Math.round(amount * 100)
}

const isValidProductPrice = (product: any) => {
  const price = priceToCents(product.customPrice)
  const base = priceToCents(product.basePrice)
  return price !== null && base !== null && base > 0 && price >= base
}

const isProductPriceDirty = (product: any) => {
  const price = priceToCents(product.customPrice)
  if (price === null) return true
  return savedPriceById.value[product.id] !== (price / 100).toFixed(2)
}

const isSavingPrice = (productId: number) => savingPriceIds.value.includes(productId)

const clearPriceFeedback = (productId: number) => {
  if (!(productId in priceFeedback.value)) return
  const next = { ...priceFeedback.value }
  delete next[productId]
  priceFeedback.value = next
}

const calculateProfit = (p: any) => {
  const custom = parseFloat(p.customPrice) || 0
  return custom - p.basePrice
}

const resetToSuggestedPrices = () => {
  for (const p of pricingProducts.value) {
    p.customPrice = p.suggestedPrice.toFixed(2)
    clearPriceFeedback(p.id)
  }
}

const saveProductPrice = async (product: any) => {
  const priceCents = priceToCents(product.customPrice)
  const baseCents = priceToCents(product.basePrice)
  if (priceCents === null || baseCents === null) {
    priceFeedback.value = { ...priceFeedback.value, [product.id]: { kind: 'error', message: '请输入最多两位小数的有效售价。' } }
    return
  }
  if (priceCents < baseCents) {
    priceFeedback.value = { ...priceFeedback.value, [product.id]: { kind: 'error', message: '售价不能低于拿货底价 ¥' + (baseCents / 100).toFixed(2) + '。' } }
    return
  }

  savingPriceIds.value = [...savingPriceIds.value, product.id]
  clearPriceFeedback(product.id)
  try {
    const fixedPrice = (priceCents / 100).toFixed(2)
    await resellerAPI.updateProductSettings(product.id, {
      settings: [
        {
          sku_id: product.skuId,
          is_listed: true,
          pricing_mode: 'fixed_price',
          fixed_price_amount: fixedPrice,
          markup_percent: '0',
          fixed_markup_amount: '0',
          sort_order: 0,
        },
      ],
    })
    savedPriceById.value = { ...savedPriceById.value, [product.id]: fixedPrice }
    priceFeedback.value = { ...priceFeedback.value, [product.id]: { kind: 'success', message: '已保存；该商品的新订单将使用此售价。' } }
  } catch (err: any) {
    priceFeedback.value = { ...priceFeedback.value, [product.id]: { kind: 'error', message: err.message || '保存失败，请重试。' } }
  } finally {
    savingPriceIds.value = savingPriceIds.value.filter((id) => id !== product.id)
  }
}

// Stats
const stats = ref({
  todayEarnings: '0.00',
  totalEarnings: '0.00',
  pendingEarnings: '0.00',
  referralUsers: 0,
  referralOrders: 0,
})

// Orders
const orders = ref<any[]>([])
const ordersLoading = ref(false)
const ordersError = ref('')
const orderDestination = (order: any) => {
  const orderNo = String(order?.order_no || '')
  if (isSessionPending(order) || hasSessionFollowUp(order)) return { name: 'order-session', params: { order_no: orderNo }, query: { from: 'agent' } }
  return order?.status === 'pending_payment'
    ? { name: 'payment', query: { order_no: orderNo, from: 'agent' } }
    : { name: 'order-detail', params: { order_no: orderNo }, query: { from: 'agent' } }
}
const formatOrderDate = (value?: string) => {
  if (!value) return '时间未知'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '时间未知'
  return new Intl.DateTimeFormat('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', hour12: false }).format(date)
}

// Sales Orders (Clients)
const salesOrders = ref<any[]>([])
const wholesaleSalesOrders = computed(() => salesOrders.value.filter((order: any) => order.is_wholesale))
const agentOrders = computed(() => {
  const byOrderNo = new Map<string, any>()
  for (const order of orders.value) {
    const key = String(order?.order_no || order?.id || '')
    if (key) byOrderNo.set(key, order)
  }
  // The reseller snapshot is authoritative for wholesale orders and contains
  // the Session-follow-up state that may be missing from the generic list.
  for (const order of wholesaleSalesOrders.value) {
    const key = String(order?.order_no || order?.id || '')
    if (key) byOrderNo.set(key, { ...byOrderNo.get(key), ...order })
  }
  return [...byOrderNo.values()].sort((a, b) => {
    const aTime = new Date(a?.created_at || 0).getTime()
    const bTime = new Date(b?.created_at || 0).getTime()
    return (Number.isFinite(bTime) ? bTime : 0) - (Number.isFinite(aTime) ? aTime : 0)
  })
})
const customerSalesOrders = computed(() => salesOrders.value.filter((order: any) => !order.is_wholesale))
const salesLoading = ref(false)
const salesError = ref('')

// Withdraw
const withdrawRecords = ref<any[]>([])
const withdrawsLoading = ref(false)
const withdrawsError = ref('')
const isSubmittingWithdraw = ref(false)
const withdrawForm = ref({
  channel: 'alipay',
  account: '',
  amount: '',
})

// Quick Buy
const quickBuyProduct = ref<any>(null)
const quickBuyVisible = ref(false)
const fullProducts = ref<any[]>([])
const productCatalogLoaded = ref(false)
const activeProductSlugs = computed(() => new Set(fullProducts.value.map((product) => String(product.slug || ''))))
const visibleAgentProducts = computed(() => productCatalogLoaded.value ? agentProducts.value.filter((product) => activeProductSlugs.value.has(product.slug)) : [])
const visiblePricingProducts = computed(() => productCatalogLoaded.value ? pricingProducts.value.filter((product) => activeProductSlugs.value.has(product.slug)) : [])

// Agent verification
const agentIdentityActive = ref(false)
const agentIdentityLoading = ref(true)
const agentIdentityError = ref(false)
const isAgent = computed(() => {
  if (!userAuthStore.isAuthenticated) return false
  return agentIdentityActive.value
})

const refreshAgentIdentity = async () => {
  agentIdentityLoading.value = true
  agentIdentityError.value = false
  agentIdentityActive.value = false
  if (!userAuthStore.isAuthenticated) {
    agentIdentityLoading.value = false
    return
  }
  try {
    const response = await resellerAPI.managementProfile()
    agentIdentityActive.value = response.data?.data?.profile?.status === 'active'
  } catch {
    agentIdentityActive.value = false
    agentIdentityError.value = true
  } finally {
    agentIdentityLoading.value = false
  }
}

// Lightweight Share Link
const agentShareLink = computed(() => {
  if (!currentAgentCode.value) return ''
  const origin = window.location.origin
  return `${origin}/?agent=${encodeURIComponent(currentAgentCode.value)}`
})

// QR Code Modal State & Handlers
const qrModalVisible = ref(false)
const qrCodeDataUrl = ref('')
const qrSharePrices = ref<Array<{ id: number; title: string; price: string }>>([])
const qrPriceLoading = ref(false)
const qrPriceError = ref('')

const siteDisplayName = computed(() => {
  const name = String(appStore.config?.brand?.site_name || '').trim()
  return name || 'GPT订阅源头代充'
})

const siteIconUrl = computed(() => {
  const icon = String(appStore.config?.brand?.site_icon || '').trim()
  return getImageUrl(icon || '/dj.svg')
})

const loadBrandIcon = () => new Promise<HTMLImageElement | null>((resolve) => {
  const image = new Image()
  image.crossOrigin = 'anonymous'
  image.onload = () => resolve(image)
  image.onerror = () => resolve(null)
  image.src = siteIconUrl.value
})

const roundedRectPath = (
  context: CanvasRenderingContext2D,
  x: number,
  y: number,
  width: number,
  height: number,
  radius: number,
) => {
  const r = Math.min(radius, width / 2, height / 2)
  context.beginPath()
  context.moveTo(x + r, y)
  context.lineTo(x + width - r, y)
  context.quadraticCurveTo(x + width, y, x + width, y + r)
  context.lineTo(x + width, y + height - r)
  context.quadraticCurveTo(x + width, y + height, x + width - r, y + height)
  context.lineTo(x + r, y + height)
  context.quadraticCurveTo(x, y + height, x, y + height - r)
  context.lineTo(x, y + r)
  context.quadraticCurveTo(x, y, x + r, y)
  context.closePath()
}

const drawRoundedRect = (
  context: CanvasRenderingContext2D,
  x: number,
  y: number,
  width: number,
  height: number,
  radius: number,
  fill: string | CanvasGradient | CanvasPattern,
  stroke?: string,
) => {
  roundedRectPath(context, x, y, width, height, radius)
  context.fillStyle = fill
  context.fill()
  if (stroke) {
    context.strokeStyle = stroke
    context.lineWidth = 1.5
    context.stroke()
  }
}

const drawCircularSiteIcon = (
  context: CanvasRenderingContext2D,
  image: HTMLImageElement | null,
  centerX: number,
  centerY: number,
  diameter: number,
) => {
  const radius = diameter / 2
  context.save()
  context.beginPath()
  context.arc(centerX, centerY, radius, 0, Math.PI * 2)
  context.fillStyle = '#ffffff'
  context.fill()

  if (image) {
    // Cover the reserved circle and clip the source image itself to a circle so
    // none of the uploaded square canvas remains visible around the artwork.
    const innerDiameter = diameter * 0.91
    const cropSize = Math.min(image.naturalWidth, image.naturalHeight)
    const sourceX = (image.naturalWidth - cropSize) / 2
    const sourceY = (image.naturalHeight - cropSize) / 2
    context.save()
    context.beginPath()
    context.arc(centerX, centerY, innerDiameter / 2, 0, Math.PI * 2)
    context.clip()
    context.drawImage(
      image,
      sourceX,
      sourceY,
      cropSize,
      cropSize,
      centerX - innerDiameter / 2,
      centerY - innerDiameter / 2,
      innerDiameter,
      innerDiameter,
    )
    context.restore()
  } else {
    const gradient = context.createLinearGradient(centerX - radius, centerY - radius, centerX + radius, centerY + radius)
    gradient.addColorStop(0, '#10b981')
    gradient.addColorStop(1, '#0f766e')
    context.fillStyle = gradient
    context.fillRect(centerX - radius, centerY - radius, diameter, diameter)
    context.fillStyle = '#ffffff'
    context.font = `700 ${Math.round(diameter * 0.36)}px "PingFang SC", "Microsoft YaHei", sans-serif`
    context.textAlign = 'center'
    context.textBaseline = 'middle'
    context.fillText(siteDisplayName.value.slice(0, 1), centerX, centerY + 1)
  }

  context.restore()
  context.beginPath()
  context.arc(centerX, centerY, radius - diameter * 0.025, 0, Math.PI * 2)
  context.strokeStyle = '#ffffff'
  context.lineWidth = Math.max(2, diameter * 0.025)
  context.stroke()
  context.beginPath()
  context.arc(centerX, centerY, radius + diameter * 0.045, 0, Math.PI * 2)
  context.strokeStyle = '#0f766e'
  context.globalAlpha = 0.16
  context.lineWidth = Math.max(1, diameter * 0.012)
  context.stroke()
  context.globalAlpha = 1
}

const createBrandedQrCanvas = async (size: number, icon: HTMLImageElement | null) => {
  const canvas = document.createElement('canvas')
  await QRCode.toCanvas(canvas, agentShareLink.value, {
    width: size,
    margin: 4,
    errorCorrectionLevel: 'H',
    color: { dark: '#102a2c', light: '#ffffff' },
  })
  const context = canvas.getContext('2d')
  if (!context) throw new Error('无法绘制二维码')
  drawCircularSiteIcon(context, icon, canvas.width / 2, canvas.height / 2, canvas.width * 0.27)
  return canvas
}

const loadQrSharePrices = async () => {
  if (!currentAgentCode.value) {
    qrPriceError.value = '暂时无法确认代理代码，请刷新页面后重试。'
    return
  }

  qrPriceLoading.value = true
  qrPriceError.value = ''
  try {
    // Reproduce the exact public catalog request made by the QR destination URL.
    const response = await api.get('/public/products', {
      params: { page: 1, page_size: 100 },
      headers: { 'X-Agent-Code': currentAgentCode.value },
    })
    const publicProducts = response.data?.data
    if (!Array.isArray(publicProducts)) throw new Error('商品列表不可用')

    const activeGPTProducts = pricingProducts.value.filter((product) =>
      product.slug !== 'openai-sms-activate' && publicProducts.some((item: any) => item.slug === product.slug),
    )
    const prices = activeGPTProducts.map((product) => {
      const publicProduct = publicProducts.find((item: any) => item.slug === product.slug)
      const cents = priceToCents(publicProduct?.price_amount)
      if (cents === null) throw new Error(`无法读取 ${product.title} 的代理展示售价`)
      return {
        id: product.id,
        title: product.title,
        price: (cents / 100).toFixed(2),
      }
    })
    qrSharePrices.value = prices
  } catch (error: any) {
    qrSharePrices.value = []
    qrPriceError.value = error?.message || '读取此代理链接的展示售价失败，请重试。'
  } finally {
    qrPriceLoading.value = false
  }
}

const generateQrCode = async () => {
  try {
    const icon = await loadBrandIcon()
    const canvas = await createBrandedQrCanvas(420, icon)
    qrCodeDataUrl.value = canvas.toDataURL('image/png')
  } catch (e) {
    console.error('Failed to generate agent QR code', e)
  }
}

const openQrModal = async () => {
  qrModalVisible.value = true
  qrPriceError.value = ''
  qrSharePrices.value = []
  await Promise.all([generateQrCode(), loadQrSharePrices()])
}

const createPromotionPoster = async () => {
  if (qrSharePrices.value.length === 0 || qrSharePrices.value.length > 4) throw new Error('未能读取当前上架 GPT 商品代理售价')
  const canvas = document.createElement('canvas')
  canvas.width = 1080
  canvas.height = 1600
  const context = canvas.getContext('2d')
  if (!context) throw new Error('无法生成推广图片')

  const icon = await loadBrandIcon()
  const qrCanvas = await createBrandedQrCanvas(560, icon)
  const background = context.createLinearGradient(0, 0, canvas.width, canvas.height)
  background.addColorStop(0, '#eff8f2')
  background.addColorStop(1, '#e4f1eb')
  context.fillStyle = background
  context.fillRect(0, 0, canvas.width, canvas.height)
  context.fillStyle = 'rgba(16, 185, 129, .08)'
  context.beginPath()
  context.arc(1010, 98, 150, 0, Math.PI * 2)
  context.fill()
  context.beginPath()
  context.arc(58, 1320, 190, 0, Math.PI * 2)
  context.fill()

  context.shadowColor = 'rgba(18, 55, 43, .12)'
  context.shadowBlur = 36
  context.shadowOffsetY = 14
  drawRoundedRect(context, 42, 34, 996, 1532, 42, '#ffffff')
  context.shadowColor = 'transparent'
  context.shadowBlur = 0
  context.shadowOffsetY = 0

  const headerGradient = context.createLinearGradient(42, 34, 1038, 280)
  headerGradient.addColorStop(0, '#122a34')
  headerGradient.addColorStop(1, '#145746')
  drawRoundedRect(context, 42, 34, 996, 270, 42, headerGradient)
  context.fillStyle = headerGradient
  context.fillRect(42, 244, 996, 60)

  context.globalAlpha = 0.16
  for (const radius of [72, 104, 136]) {
    context.beginPath()
    context.arc(915, 166, radius, 0, Math.PI * 2)
    context.strokeStyle = '#91f3d4'
    context.lineWidth = 2
    context.stroke()
  }
  context.globalAlpha = 1
  context.beginPath()
  context.arc(915, 166, 45, 0, Math.PI * 2)
  context.fillStyle = 'rgba(130, 235, 203, .16)'
  context.fill()
  context.fillStyle = '#a2f4d9'
  context.font = '700 17px Arial, sans-serif'
  context.textAlign = 'left'
  context.fillText('AI SUBSCRIPTION', 211, 96)

  drawCircularSiteIcon(context, icon, 139, 163, 112)
  context.fillStyle = '#f4fff9'
  context.textAlign = 'left'
  context.textBaseline = 'alphabetic'
  context.font = '750 44px "PingFang SC", "Microsoft YaHei", sans-serif'
  context.fillText(siteDisplayName.value, 220, 151, 615)
  context.fillStyle = '#d4e9df'
  context.font = '600 26px "PingFang SC", "Microsoft YaHei", sans-serif'
  context.fillText('用户自己的账号 · 官方直充', 222, 192)

  drawRoundedRect(context, 132, 326, 816, 680, 30, '#ffffff', '#e1e9e4')
  context.fillStyle = '#233e38'
  context.textAlign = 'center'
  context.font = '700 26px "PingFang SC", "Microsoft YaHei", sans-serif'
  context.fillText('扫码进入专属销售页面', 540, 367)
  context.drawImage(qrCanvas, 260, 386, 560, 560)
  context.textAlign = 'left'
  context.fillStyle = '#15352f'
  context.font = '700 28px "PingFang SC", "Microsoft YaHei", sans-serif'
  context.fillText('GPT 热门订阅', 94, 1053)
  context.fillStyle = '#71847c'
  context.textAlign = 'right'
  context.font = '600 20px "PingFang SC", "Microsoft YaHei", sans-serif'
  context.fillText('此二维码对应的代理售价', 986, 1052)

  const productCards = [
    { x: 88, y: 1076 },
    { x: 548, y: 1076 },
    { x: 88, y: 1188 },
    { x: 548, y: 1188 },
  ]
  qrSharePrices.value.forEach((product, index) => {
    const card = productCards[index]
    if (!card) return
    const { x, y } = card
    const [productName, duration] = product.title.split('·').map((part) => part.trim())
    drawRoundedRect(context, x, y, 444, 98, 18, '#f5f9f6', '#e5eee8')
    context.textAlign = 'left'
    context.fillStyle = '#394b45'
    context.font = '700 22px "PingFang SC", "Microsoft YaHei", sans-serif'
    context.fillText(productName || product.title, x + 18, y + 40, 210)
    context.fillStyle = '#75877f'
    context.font = '600 18px "PingFang SC", "Microsoft YaHei", sans-serif'
    context.fillText(duration || '1个月', x + 18, y + 70, 210)

    const priceText = `¥${product.price}`
    context.textAlign = 'right'
    context.fillStyle = '#087858'
    context.font = '750 35px "PingFang SC", "Microsoft YaHei", sans-serif'
    context.fillStyle = '#82928b'
    context.font = '600 17px "PingFang SC", "Microsoft YaHei", sans-serif'
    context.fillText('/月', x + 426, y + 60)
    const unitWidth = context.measureText('/月').width
    context.fillStyle = '#087858'
    context.font = '750 35px "PingFang SC", "Microsoft YaHei", sans-serif'
    context.fillText(priceText, x + 426 - unitWidth - 8, y + 61)
  })

  drawRoundedRect(context, 88, 1320, 904, 164, 20, '#eaf6ee', '#d8eadf')
  context.beginPath()
  context.arc(144, 1402, 28, 0, Math.PI * 2)
  context.fillStyle = '#0b8a65'
  context.fill()
  context.strokeStyle = '#ffffff'
  context.lineWidth = 4
  context.lineCap = 'round'
  context.lineJoin = 'round'
  context.beginPath()
  context.moveTo(132, 1402)
  context.lineTo(141, 1411)
  context.lineTo(157, 1392)
  context.stroke()
  context.textAlign = 'left'
  context.fillStyle = '#174637'
  context.font = '700 27px "PingFang SC", "Microsoft YaHei", sans-serif'
  context.fillText('用户自己的账号 · 官方直充', 190, 1397)
  context.fillStyle = '#536f63'
  context.font = '600 22px "PingFang SC", "Microsoft YaHei", sans-serif'
  context.fillText('无需提供账号密码 · 保护隐私', 190, 1434)

  context.fillStyle = '#83958c'
  context.textAlign = 'left'
  context.font = '600 19px Arial, "PingFang SC", sans-serif'
  context.fillText(window.location.host, 94, 1530)
  context.textAlign = 'right'
  context.fillText('扫码了解套餐并开启订阅', 986, 1530)

  return canvas
}

const downloadQrCode = async () => {
  if (!qrCodeDataUrl.value || qrPriceLoading.value || qrPriceError.value) return
  try {
    const poster = await createPromotionPoster()
    const a = document.createElement('a')
    a.href = poster.toDataURL('image/png')
    const profile = userAuthStore.user || {}
    const userName = String(profile.nickname || profile.username || profile.email?.split('@')[0] || 'user').trim()
    const safeUserName = userName.replace(/[\\/:*?"<>|\u0000-\u001f]/g, '_') || 'user'
    const safeAgentCode = String(currentAgentCode.value || 'share').replace(/[^\p{L}\p{N}_-]/gu, '_')
    a.download = `${safeUserName}Agent${safeAgentCode}.png`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
  } catch (error: any) {
    qrPriceError.value = error?.message || '生成推广图片失败，请重试。'
  }
}

const shareQrCode = async () => {
  if (navigator.share) {
    try {
      await navigator.share({
        title: 'ChatGPT 官方稳定直充',
        text: `🔥【ChatGPT 官方稳定直充】专属优惠通道：`,
        url: agentShareLink.value,
      })
      return
    } catch (e) {
      // Fall back to copy
    }
  }
  // Fallback to copying promotion text
  copyPromotionText()
}

const copyPromotionText = () => {
  const text = `🔥【ChatGPT 官方稳定直充】
支持 ChatGPT Plus 1个月、ChatGPT Pro 5x / 20x / 50x 极速充值，独享质保！
立即点击链接开通：
${agentShareLink.value}`
  copyText(text, '推广文案')
}

// Send verify code countdown timer
const startCountdown = () => {
  sendCodeCountdown.value = 60
  const timer = setInterval(() => {
    sendCodeCountdown.value--
    if (sendCodeCountdown.value <= 0) {
      clearInterval(timer)
    }
  }, 1000)
}

// Handlers
const handleRegisterAgent = async () => {
  authError.value = ''
  if (!regForm.value.inviteCode.trim()) {
    authError.value = '请先填写站长发放的专属代理邀请码'
    return
  }
  if (!regForm.value.email.trim() || !regForm.value.password) {
    authError.value = '请填写完整账号邮箱与密码'
    return
  }
  if (!regForm.value.agreementAccepted) {
    authError.value = '请先阅读并同意服务条款和隐私政策'
    return
  }
  if (regForm.value.confirmPassword && regForm.value.password !== regForm.value.confirmPassword) {
    authError.value = '两次输入的密码不一致'
    return
  }

  isSubmitting.value = true
  try {
    agentIdentityLoading.value = true
    await userAuthStore.register({
      email: regForm.value.email.trim(),
      password: regForm.value.password,
      code: '',
      agreement_accepted: regForm.value.agreementAccepted,
      role: 'agent',
      invite_code: regForm.value.inviteCode.trim(),
    })

    await refreshAgentIdentity()
    await loadData()
  } catch (err: any) {
    authError.value = err.message || '注册开通失败，请检查填写内容'
  } finally {
    isSubmitting.value = false
  }
}

const handleSendResetCode = async () => {
  authError.value = ''
  if (!forgotForm.value.email.trim()) {
    authError.value = '请先输入接收重置验证码的邮箱'
    return
  }
  isSendingCode.value = true
  try {
    await userAuthStore.sendVerifyCode({
      email: forgotForm.value.email.trim(),
      purpose: 'reset',
    })
    startCountdown()
  } catch (err: any) {
    authError.value = err.message || '发送验证码失败，请稍后重试'
  } finally {
    isSendingCode.value = false
  }
}

const handleResetAgentPassword = async () => {
  authError.value = ''
  if (!forgotForm.value.email.trim() || !forgotForm.value.code.trim() || !forgotForm.value.newPassword) {
    authError.value = '请填写完整邮箱、验证码与新密码'
    return
  }
  if (forgotForm.value.newPassword !== forgotForm.value.confirmPassword) {
    authError.value = '两次输入的新密码不一致'
    return
  }
  isSubmitting.value = true
  try {
    await userAuthStore.forgotPassword({
      email: forgotForm.value.email.trim(),
      code: forgotForm.value.code.trim(),
      new_password: forgotForm.value.newPassword,
    })
    alert('密码重置成功！请使用新密码登录代理账号。')
    loginForm.value.account = forgotForm.value.email.trim()
    loginForm.value.password = ''
    authMode.value = 'login'
    forgotForm.value.code = ''
    forgotForm.value.newPassword = ''
    forgotForm.value.confirmPassword = ''
  } catch (err: any) {
    authError.value = err.message || '重置密码失败，请检查验证码是否正确或已过期'
  } finally {
    isSubmitting.value = false
  }
}

const handleLoginAgent = async () => {
  authError.value = ''
  if (!loginForm.value.account || !loginForm.value.password) {
    authError.value = '请输入账号/邮箱和密码'
    return
  }

  isSubmitting.value = true
  try {
    agentIdentityLoading.value = true
    await userAuthStore.login({
      email: loginForm.value.account.trim(),
      password: loginForm.value.password,
    })

    await refreshAgentIdentity()
    await loadData()
  } catch (err: any) {
    authError.value = err.message || '登录失败，请检查账号密码'
  } finally {
    isSubmitting.value = false
  }
}

const handleLogout = () => {
  userAuthStore.logout('/agent')
}

const copyText = async (text: string, label: string) => {
  try {
    await navigator.clipboard.writeText(text)
    alert(`${label}已复制到剪贴板！`)
  } catch (e) {
    alert(`复制失败，请手动长按复制: ${text}`)
  }
}

const handleWholesaleBuy = (slug: string) => {
  const found = fullProducts.value.find((p) => p.slug === slug)
  if (found) {
    const prices = wholesalePricesBySlug.value[slug] || {}
    const skus = Array.isArray(found.skus) ? found.skus : []
    const pricedSkus = skus.map((sku: any) => ({ ...sku, price_amount: prices[Number(sku.id)] || '' }))
    if (!pricedSkus.length || pricedSkus.some((sku: any) => !sku.price_amount)) {
      alert('平台尚未为该套餐设置代理拿货价，请联系管理员。')
      return
    }
    sessionStorage.setItem('dj_agent_code', currentAgentCode.value)
    quickBuyProduct.value = {
      ...found,
      price_amount: pricedSkus[0].price_amount,
      skus: pricedSkus,
      agent_wholesale: true,
      agent_wholesale_agent_code: currentAgentCode.value,
    }
    quickBuyVisible.value = true
  } else {
    productAPI.detail(slug).then((res: any) => {
      if (res.data?.data) {
        fullProducts.value = [...fullProducts.value, res.data.data]
        handleWholesaleBuy(slug)
      }
    })
  }
}

const loadOrders = async () => {
  ordersError.value = ''
  try {
    ordersLoading.value = true
    const ordRes = await userOrderAPI.list({ page: 1, page_size: 15 })
    if (ordRes.data?.data) {
      orders.value = ordRes.data.data.list || ordRes.data.data || []
    }
  } catch (e: any) {
    ordersError.value = e?.message || '订单读取失败，请点击“刷新订单”重试。'
  } finally {
    ordersLoading.value = false
  }
}

const loadSalesOrders = async () => {
  salesError.value = ''
  try {
    salesLoading.value = true
    const res = await resellerAPI.orders({ page: 1, page_size: 20 })
    if (res.data?.data) {
      salesOrders.value = res.data.data.list || res.data.data || []
    }
  } catch (e: any) {
    salesError.value = e?.message || '销售订单读取失败，请点击刷新后重试。'
  } finally {
    salesLoading.value = false
  }
}

// 代理拿货订单由 reseller 快照接口提供；个人订单接口受租户范围限制，
// 因此刷新拿货订单时必须同时更新两路数据，避免支付前的旧快照一直留在页面。
const refreshAgentOrders = async () => {
  await Promise.all([loadOrders(), loadSalesOrders()])
}

const loadWithdraws = async () => {
  withdrawsError.value = ''
  try {
    withdrawsLoading.value = true
    const res = await resellerAPI.withdraws({ page: 1, page_size: 20 })
    if (res.data?.data) {
      withdrawRecords.value = res.data.data.list || res.data.data || []
    }
  } catch (e: any) {
    withdrawsError.value = e?.message || '提现记录读取失败，请稍后重试。'
  } finally {
    withdrawsLoading.value = false
  }
}

const handleApplyWithdraw = async () => {
  if (!withdrawForm.value.account || !withdrawForm.value.amount) {
    alert('请填写完整的提现账号和金额')
    return
  }
  const amt = parseFloat(withdrawForm.value.amount)
  if (isNaN(amt) || amt <= 0) {
    alert('请输入有效的提现金额')
    return
  }
  isSubmittingWithdraw.value = true
  try {
    await resellerAPI.applyWithdraw({
      amount: amt.toFixed(2),
      currency: 'CNY',
      channel: withdrawForm.value.channel,
      account: withdrawForm.value.account,
    })
    alert('提现申请已提交，请等待管理员审核打款！')
    withdrawForm.value.amount = ''
    await loadWithdraws()
    await loadData()
  } catch (err: any) {
    alert(err.message || '提交提现申请失败，请联系客服处理')
  } finally {
    isSubmittingWithdraw.value = false
  }
}

const loadData = async () => {
  productCatalogLoaded.value = false
  try {
    const res = await productAPI.list({ page: 1, page_size: 20 }, { skipAgentCode: true })
    const pList = res.data?.data || []
    fullProducts.value = pList
    productCatalogLoaded.value = true
    for (const item of agentProducts.value) {
      const match = pList.find((p: any) => p.slug === item.slug)
      if (match && match.price_amount) {
        item.retail = parseFloat(match.price_amount).toFixed(2)
      }
    }
    for (const item of pricingProducts.value) {
      const match = pList.find((p: any) => p.slug === item.slug)
      if (match && match.price_amount) {
        item.suggestedPrice = parseFloat(match.price_amount)
      }
    }
  } catch (e) {
    fullProducts.value = []
    productCatalogLoaded.value = true
  }

  if (isAgent.value) {
    currentAgentCode.value = ''
    agentProfileError.value = ''
    pricingSettingsLoading.value = true
    pricingSettingsError.value = ''
    let pricingSettingsLoaded = false
    try {
      // 1. 获取代理专属资料
      const profRes = await resellerAPI.managementProfile()
      if (profRes.data?.data?.profile) {
        const p = profRes.data.data.profile
        if (p.agent_code) {
          currentAgentCode.value = p.agent_code
        }
      }
      if (!currentAgentCode.value) throw new Error('未能读取代理专属代码，请刷新页面或联系管理员。')

      // 2. 获取商品当前分销定价
      const settingsRes = await resellerAPI.productSettings()
      pricingSettingsLoaded = true
      if (settingsRes.data?.data) {
        const list = settingsRes.data.data.list || settingsRes.data.data || []
        for (const item of list) {
          const product = item.product || {}
          const match = pricingProducts.value.find((p) => p.slug === product.slug)
          const agentProduct = agentProducts.value.find((p) => p.slug === product.slug)
          if (!match) continue
          match.id = Number(product.id || match.id)
          const skuRows = Array.isArray(item.skus) ? item.skus : []
          const activeSkuRows = skuRows.filter((sku: any) => sku.is_active !== false)
          const priceMap: Record<number, string> = {}
          for (const sku of activeSkuRows) {
            const price = Number(sku.reseller_purchase_price_amount || 0)
            if (price > 0) priceMap[Number(sku.id)] = price.toFixed(2)
          }
          wholesalePricesBySlug.value = { ...wholesalePricesBySlug.value, [product.slug]: priceMap }
          const defaultSku = activeSkuRows[0]
          if (defaultSku) {
            match.skuId = Number(defaultSku.id)
            match.basePrice = Number(defaultSku.reseller_purchase_price_amount || 0)
            if (agentProduct) agentProduct.agent = match.basePrice > 0 ? match.basePrice.toFixed(2) : ''
          }
          const productPrice = Number(product.price_amount || match.suggestedPrice || 0)
          if (productPrice > 0) match.suggestedPrice = productPrice
          const configuredSetting = defaultSku?.setting || item.product_setting
          const configuredPrice = Number(configuredSetting?.fixed_price_amount || 0)
          if (configuredSetting?.pricing_mode === 'fixed_price' && configuredPrice > 0) {
            match.customPrice = configuredPrice.toFixed(2)
            savedPriceById.value = { ...savedPriceById.value, [match.id]: match.customPrice }
          } else if (configuredSetting?.effective_price_amount) {
            match.customPrice = Number(configuredSetting.effective_price_amount).toFixed(2)
            savedPriceById.value = { ...savedPriceById.value, [match.id]: match.customPrice }
          }
        }
      }

      // 3. 财务与数据看板
      const dashRes = await resellerAPI.dashboard()
      if (dashRes.data?.data) {
        const d = dashRes.data.data
        const cnyBalance = (Array.isArray(d.balances) ? d.balances : []).find((item: any) => item.currency === 'CNY')
        stats.value.pendingEarnings = Number(cnyBalance?.available_amount || 0).toFixed(2)

        // Dashboard returns balances only. Read order-profit ledger rows for the
        // lifetime and today's figures so the cards reflect accounting records.
        let page = 1
        let totalPages = 1
        let totalProfit = 0
        let todayProfit = 0
        const today = new Date().toLocaleDateString('en-CA')
        do {
          const ledgerRes = await resellerAPI.ledgerEntries({ page, page_size: 100, type: 'order_profit' })
          const rows = ledgerRes.data?.data || []
          for (const row of rows) {
            if (row.status === 'canceled' || row.currency !== 'CNY') continue
            const amount = Number(row.amount || 0)
            totalProfit += amount
            if (new Date(row.created_at).toLocaleDateString('en-CA') === today) todayProfit += amount
          }
          totalPages = Number(ledgerRes.data?.pagination?.total_page || 1)
          page += 1
        } while (page <= totalPages)
        stats.value.todayEarnings = todayProfit.toFixed(2)
        stats.value.totalEarnings = totalProfit.toFixed(2)
      } else {
        const affRes = await affiliateAPI.dashboard()
        if (affRes.data?.data) {
          const d = affRes.data.data
          stats.value.totalEarnings = d.total_commission || '0.00'
          stats.value.pendingEarnings = d.unsettled_commission || '0.00'
          stats.value.referralUsers = d.referral_count || 0
        }
      }
    } catch (e: any) {
      agentProfileError.value = '代理中心资料加载失败，部分功能可能暂不可用，请刷新后重试。'
      if (!pricingSettingsLoaded) {
        pricingSettingsError.value = '无法确认当前售价配置；为避免覆盖已有价格，请刷新页面后再保存。'
      }
    } finally {
      pricingSettingsLoading.value = false
    }

    await refreshAgentOrders()
    await loadWithdraws()
  }
}

const retryAgentIdentity = async () => {
  await refreshAgentIdentity()
  if (isAgent.value) await loadData()
}

onMounted(() => {
  if (route.query.tab === 'orders') activeTab.value = 'orders'
  void refreshAgentIdentity().finally(() => loadData())
})
</script>

<style scoped>
.scrollbar-none::-webkit-scrollbar {
  display: none;
}
.scrollbar-none {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
</style>
