<template>
  <div class="plus-half-promo min-h-screen bg-background text-foreground pt-20 pb-16 md:pt-24 md:pb-24">
    <div class="max-w-4xl mx-auto px-4 sm:px-6 space-y-8">
      
      <!-- Back Link -->
      <div class="flex items-center justify-between">
        <router-link
          to="/"
          class="inline-flex items-center gap-1.5 text-xs sm:text-sm font-medium text-muted-foreground hover:text-foreground transition"
        >
          <ArrowLeft class="h-4 w-4" />
          <span>返回商城首页</span>
        </router-link>
        <span class="inline-flex items-center gap-1 text-[11px] font-semibold text-emerald-600 dark:text-emerald-400 bg-emerald-500/10 px-2.5 py-1 rounded-full border border-emerald-500/20">
          <Sparkles class="h-3.5 w-3.5 animate-pulse" />
          官方特惠资格通道
        </span>
      </div>

      <!-- Hero Banner Card -->
      <div class="relative overflow-hidden rounded-3xl border border-emerald-200/70 dark:border-emerald-900/50 bg-[radial-gradient(ellipse_at_top_right,_rgba(255,255,255,0.96),_transparent_55%),linear-gradient(135deg,#eef8f1_0%,#e4f3ea_50%,#edf7f5_100%)] dark:bg-[linear-gradient(135deg,#13241c,#172924_58%,#152427)] p-6 sm:p-8 md:p-10 shadow-sm">
        <div class="flex flex-col md:flex-row items-start md:items-center justify-between gap-6">
          <div class="space-y-3 flex-1">
            <div class="inline-flex items-center gap-2 rounded-full border border-emerald-300/70 dark:border-emerald-800 bg-emerald-100/90 dark:bg-emerald-950/80 px-3 py-1 text-xs font-semibold text-emerald-800 dark:text-emerald-300">
              <span class="h-2 w-2 rounded-full bg-emerald-500 animate-pulse"></span>
              <span>限时专属福利 · 官方直充</span>
            </div>
            <h1 class="text-2xl sm:text-3xl md:text-4xl font-extrabold tracking-tight text-slate-900 dark:text-white leading-tight">
              ChatGPT Plus 官方半价资格预检
            </h1>
            <p class="text-xs sm:text-sm md:text-base text-slate-600 dark:text-slate-300 leading-relaxed max-w-xl">
              OpenAI 会对符合条件的账号发放半价促销资格。提交您个人账号的 Session 即可免费进行实时预检，通过即可享受专属半价直充！
            </p>
            <div class="flex flex-wrap items-baseline gap-3 pt-2">
              <span class="text-3xl sm:text-4xl font-black text-emerald-600 dark:text-emerald-400">¥66.00 <span class="text-sm font-normal text-muted-foreground">/ 月</span></span>
              <span class="text-sm text-muted-foreground line-through">官方原价 ¥168</span>
              <span class="rounded-full bg-emerald-600 text-white text-[11px] font-bold px-2.5 py-0.5 shadow-xs">立省 ¥102</span>
            </div>
          </div>

          <div class="shrink-0 flex md:flex-col items-center justify-center p-4 rounded-2xl bg-white/70 dark:bg-black/20 backdrop-blur-xs border border-emerald-200/50 dark:border-emerald-800/40 shadow-xs self-stretch md:self-auto">
            <img src="/openai.svg" alt="OpenAI" class="h-14 w-14 sm:h-16 sm:w-16 object-contain dark:invert" />
            <span class="mt-2 text-[11px] font-bold text-slate-700 dark:text-slate-300">官方正品 · 一号一卡</span>
          </div>
        </div>
      </div>

      <!-- Checker Workbench Card (Direct Interactive Area) -->
      <div id="check-area" class="rounded-3xl border border-border bg-card p-6 sm:p-8 shadow-sm space-y-6">
        <div class="flex items-center justify-between border-b border-border/50 pb-4">
          <div class="flex items-center gap-3">
            <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-100 dark:bg-emerald-950/80 text-emerald-700 dark:text-emerald-300">
              <Sparkles class="h-5 w-5" />
            </div>
            <div>
              <h2 class="text-lg font-bold text-foreground">在线资格检测工作台</h2>
              <p class="text-xs text-muted-foreground mt-0.5">官方实时接口返回，检测结果 100% 真实有效</p>
            </div>
          </div>
          <span class="inline-flex items-center gap-1 text-xs text-emerald-600 dark:text-emerald-400 font-medium">
            <ShieldCheck class="h-4 w-4" /> 全程加密脱敏
          </span>
        </div>

        <!-- Input Area -->
        <div class="space-y-3">
          <div class="flex items-center justify-between text-xs text-muted-foreground">
            <label for="session-input" class="font-medium text-foreground">ChatGPT Session JSON</label>
            <span class="text-[11px]">复制完整的 Session JSON</span>
          </div>

          <textarea
            id="session-input"
            v-model="session"
            rows="5"
            autocomplete="off"
            spellcheck="false"
            class="w-full rounded-2xl border border-input bg-background/50 p-3.5 text-xs sm:text-sm font-mono placeholder:text-muted-foreground/60 focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 focus:outline-none transition resize-none"
            placeholder="粘贴从 chatgpt.com/api/auth/session 复制的完整代码..."
            :disabled="loading"
          />

          <!-- How to get Session Guide -->
          <div class="rounded-2xl border border-emerald-500/25 bg-emerald-500/5 dark:bg-emerald-950/20 p-4 text-xs space-y-2">
            <div class="flex items-center justify-between font-semibold text-foreground">
              <span class="inline-flex items-center gap-1.5 text-emerald-800 dark:text-emerald-300">
                <HelpCircle class="h-4 w-4 text-emerald-600 dark:text-emerald-400 shrink-0" />
                如何获取 ChatGPT Session？（只需 10 秒）
              </span>
              <button
                type="button"
                class="inline-flex items-center gap-1 rounded-lg px-2.5 py-1 text-xs font-semibold text-emerald-700 dark:text-emerald-300 bg-emerald-100/70 dark:bg-emerald-900/40 hover:bg-emerald-200/60 transition cursor-pointer"
                @click="copySessionUrl"
              >
                <Check v-if="copied" class="h-3.5 w-3.5 text-emerald-600" />
                <Copy v-else class="h-3.5 w-3.5" />
                <span>{{ copied ? '已复制网址' : '一键复制 Session 网址' }}</span>
              </button>
            </div>
            <ol class="list-decimal list-inside space-y-1.5 text-muted-foreground text-xs leading-relaxed pt-1">
              <li>在电脑或手机浏览器打开并登录官方 <a href="https://chatgpt.com" target="_blank" rel="noreferrer" class="text-emerald-600 underline font-medium">chatgpt.com</a></li>
              <li>在同一浏览器新标签页访问 <a href="https://chatgpt.com/api/auth/session" target="_blank" rel="noreferrer" class="text-emerald-600 font-mono font-medium underline">chatgpt.com/api/auth/session</a></li>
              <li>全选页面中的所有代码并复制，粘贴到上方输入框点击「开始检测」即可</li>
            </ol>
          </div>

          <!-- Error Alert -->
          <div v-if="error" class="flex items-start gap-2.5 rounded-2xl border border-destructive/40 bg-destructive/10 p-4 text-xs sm:text-sm text-destructive" role="alert">
            <AlertCircle class="h-4 w-4 shrink-0 mt-0.5" />
            <div class="leading-relaxed">{{ error }}</div>
          </div>

          <!-- Ineligible Alert -->
          <div v-if="checked && !eligible && !error" class="flex items-start gap-3 rounded-2xl border border-amber-500/40 bg-amber-500/10 p-4 text-xs sm:text-sm text-amber-700 dark:text-amber-300" role="status">
            <AlertCircle class="h-4 w-4 shrink-0 mt-0.5" />
            <div class="space-y-1 leading-relaxed">
              <div class="font-bold text-sm">当前未确认具备半价资格</div>
              <div>该账号依照OpenAI规则未授予或已使用过半价资格，您仍可正常选购常规 Plus 订阅服务。</div>
              <div class="pt-2">
                <router-link to="/products" class="inline-flex items-center gap-1 font-semibold text-emerald-700 dark:text-emerald-400 hover:underline">
                  前往选购常规 ChatGPT Plus 1个月套餐 &rarr;
                </router-link>
              </div>
            </div>
          </div>

          <!-- Eligible Card -->
          <div v-if="eligible" class="rounded-2xl border border-emerald-500/40 bg-emerald-500/10 p-5 text-sm text-foreground space-y-3" role="status">
            <div class="flex items-center gap-2 text-emerald-700 dark:text-emerald-300 font-bold text-base">
              <CheckCircle2 class="h-5 w-5 shrink-0" />
              <span>恭喜！账号 {{ accountEmail }} 已确认具备官方半价资格</span>
            </div>
            <div class="flex items-baseline justify-between pt-2 border-t border-emerald-500/20">
              <span class="text-xs text-muted-foreground">专属特惠结账价</span>
              <span class="text-2xl font-black text-emerald-600 dark:text-emerald-400">¥{{ price || '66.00' }} <span class="text-xs font-normal text-muted-foreground">/ 月</span></span>
            </div>
            <div class="pt-1">
              <button
                type="button"
                class="w-full inline-flex items-center justify-center gap-2 rounded-2xl bg-emerald-600 px-6 py-3.5 text-base font-extrabold text-white shadow-md hover:bg-emerald-700 hover:shadow-lg transition active:scale-98 animate-pulse"
                :disabled="loading || !offer"
                @click="buy"
              >
                <span>以 ¥{{ price || '66.00' }} 立即下单直充</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Action Button -->
        <div v-if="!eligible" class="pt-2">
          <button
            type="button"
            class="w-full inline-flex items-center justify-center gap-2 rounded-2xl bg-emerald-600 px-6 py-3.5 text-sm sm:text-base font-bold text-white shadow-md hover:bg-emerald-700 disabled:opacity-50 transition active:scale-98"
            :disabled="loading || !session.trim()"
            @click="check"
          >
            <Loader2 v-if="loading" class="h-4 w-4 animate-spin" />
            <span>{{ loading ? '正在请求官方接口实时预检中…' : '立即免费检测半价资格' }}</span>
          </button>
        </div>
      </div>

      <!-- Feature Guarantee Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div class="rounded-2xl border border-border bg-card p-5 space-y-2">
          <div class="flex items-center gap-2.5 font-bold text-foreground">
            <div class="p-2 rounded-xl bg-emerald-500/10 text-emerald-600">
              <ShieldCheck class="h-5 w-5" />
            </div>
            <span>官方正规直充 · 绝无封号风险</span>
          </div>
          <p class="text-xs text-muted-foreground leading-relaxed">
            走正规海外 Visa/MasterCard 银行卡直充，非共享账号，独立开通，稳定享有 30 天质保。
          </p>
        </div>

        <div class="rounded-2xl border border-border bg-card p-5 space-y-2">
          <div class="flex items-center gap-2.5 font-bold text-foreground">
            <div class="p-2 rounded-xl bg-emerald-500/10 text-emerald-600">
              <Sparkles class="h-5 w-5" />
            </div>
            <span>解锁全部旗舰大模型</span>
          </div>
          <p class="text-xs text-muted-foreground leading-relaxed">
            畅享 GPT-4o 顶阶推理、OpenAI o1 深度思考模型、高级实时语音、DALL·E 3 作图及 Canvas 交互。
          </p>
        </div>

        <div class="rounded-2xl border border-border bg-card p-5 space-y-2">
          <div class="flex items-center gap-2.5 font-bold text-foreground">
            <div class="p-2 rounded-xl bg-emerald-500/10 text-emerald-600">
              <Zap class="h-5 w-5" />
            </div>
            <span>秒级自动化充值履约</span>
          </div>
          <p class="text-xs text-muted-foreground leading-relaxed">
            支付成功后，系统即刻通过官方 API 触发代充升级，全流程自动化处理，最快几十秒即刻生效。
          </p>
        </div>

        <div class="rounded-2xl border border-border bg-card p-5 space-y-2">
          <div class="flex items-center gap-2.5 font-bold text-foreground">
            <div class="p-2 rounded-xl bg-emerald-500/10 text-emerald-600">
              <Lock class="h-5 w-5" />
            </div>
            <span>隐私保护 · 安全脱敏</span>
          </div>
          <p class="text-xs text-muted-foreground leading-relaxed">
            Session 凭证仅在履约时用于调用官方充值通道，全程密文存储脱敏传输，绝不留存或泄露任何用户对话。
          </p>
        </div>
      </div>

      <!-- FAQ Section -->
      <div class="rounded-3xl border border-border bg-card p-6 sm:p-8 space-y-4">
        <h3 class="text-lg font-bold text-foreground flex items-center gap-2">
          <HelpCircle class="h-5 w-5 text-emerald-600" />
          常见问答 FAQ
        </h3>
        <div class="divide-y divide-border/50 text-xs sm:text-sm space-y-3">
          <div class="pt-3 space-y-1">
            <p class="font-bold text-foreground">Q: 什么是 Plus 半价资格？</p>
            <p class="text-muted-foreground leading-relaxed">A: OpenAI 会依据用户活跃度、注册地区等因素随机发放半价促销资格。拥有该资格的用户可以用专属优惠价完成订阅。本页面通过官方预检接口判断您的账号是否命中该特权。</p>
          </div>
          <div class="pt-3 space-y-1">
            <p class="font-bold text-foreground">Q: 检测未通过怎么办？</p>
            <p class="text-muted-foreground leading-relaxed">A: 若未通过检测，说明该账号暂无优惠资格，您可以直接在我们的商城选购「ChatGPT Plus · 1个月」常规官方直充套餐，依然享受稳定一号一卡质保服务。</p>
          </div>
          <div class="pt-3 space-y-1">
            <p class="font-bold text-foreground">Q: 充值后支持退款吗？</p>
            <p class="text-muted-foreground leading-relaxed">A: 订单支付后系统将实时对接 OpenAI 官方进行扣费开通。因虚拟订阅资产开通后即时消耗，官方直充成功后不支持无故退款；若因不可抗力导致充值未成功，将全额退还您的支付款项。</p>
          </div>
        </div>
      </div>

      <!-- Footer CTA Banner -->
      <div class="text-center py-4 space-y-3">
        <p class="text-xs text-muted-foreground">GPT 源头代充 · 稳定 · 高效 · 长期服务</p>
        <div class="flex items-center justify-center gap-4">
          <router-link to="/" class="text-xs text-emerald-600 font-semibold hover:underline">
            进入商城选购更多商品 &rarr;
          </router-link>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import {
  Sparkles,
  ShieldCheck,
  Zap,
  Lock,
  ArrowLeft,
  Loader2,
  CheckCircle2,
  AlertCircle,
  HelpCircle,
  Copy,
  Check,
} from 'lucide-vue-next'
import { guestOrderAPI, productAPI } from '../api'
import { useBuyNowStore } from '../stores/buyNow'

const router = useRouter()
const buyNowStore = useBuyNowStore()

const session = ref('')
const loading = ref(false)
const error = ref('')
const checked = ref(false)
const eligible = ref(false)
const accountEmail = ref('')
const offer = ref<any>(null)
const offerSku = ref<any>(null)
const price = ref('')
const copied = ref(false)

watch(session, () => {
  eligible.value = false
  checked.value = false
  offer.value = null
  offerSku.value = null
  price.value = ''
  error.value = ''
})

const copySessionUrl = async () => {
  try {
    await navigator.clipboard.writeText('https://chatgpt.com/api/auth/session')
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    const input = document.createElement('input')
    input.value = 'https://chatgpt.com/api/auth/session'
    document.body.appendChild(input)
    input.select()
    document.execCommand('copy')
    document.body.removeChild(input)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  }
}

const check = async () => {
  if (!session.value.trim()) return
  loading.value = true
  error.value = ''
  eligible.value = false
  checked.value = false
  try {
    const response = await guestOrderAPI.checkPlusHalfEligibility(session.value.trim())
    const result = response.data.data
    checked.value = true
    if (result?.eligible !== true) return
    const detail = await productAPI.detail('chatgpt-plus-half')
    const product = detail.data.data
    const sku = (product?.skus || []).find((row: any) => row.id === result.sku_id && row.sku_code === 'plus_half' && row.is_active)
    if (!product || !sku || product.id !== result.product_id) {
      throw new Error('半价套餐暂不可售，请稍后再试')
    }
    accountEmail.value = String(result.account_email || '')
    offer.value = product
    offerSku.value = sku
    price.value = String(sku.price_amount || product.price_amount || '66.00')
    eligible.value = true
  } catch (cause: any) {
    error.value = cause?.message || '资格检测失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

const buy = () => {
  if (!eligible.value || !offer.value || !offerSku.value || !session.value.trim()) return
  buyNowStore.setItem({
    productId: offer.value.id,
    skuId: offerSku.value.id,
    skuCode: 'plus_half',
    slug: offer.value.slug,
    title: offer.value.title,
    priceAmount: price.value,
    quantity: 1,
    purchaseType: offer.value.purchase_type,
    fulfillmentType: offer.value.fulfillment_type,
    manualFormSchema: offer.value.manual_form_schema || {},
    paymentChannelIds: offer.value.payment_channel_ids,
  })
  buyNowStore.setPlusHalfSession(session.value.trim())
  void router.push({ path: '/checkout', query: { mode: 'buynow' } })
}
</script>
