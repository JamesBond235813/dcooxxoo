<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RefreshCw, Save, Trash2, ArrowLeftRight, Wallet, CircleDollarSign, Clock3 } from 'lucide-vue-next'
import { adminAPI } from '@/api/admin'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { formatMoney } from '@/utils/format'

interface CashFlowRow {
  id: number; event_date: string; direction: 'in' | 'out' | 'transfer'; category: string; account: string; counter_account?: string
  amount: string; currency: string; exchange_rate_snapshot: string; cny_amount: string; counterparty?: string; reference?: string
  order_no?: string; note?: string; related_source?: string; related_source_id?: string
}
interface PendingSettlement { payment_id: number; order_id: number; order_no: string; channel: string; amount: string; currency: string; paid_at: string }
interface CashFlowOptions { orders: { order_no: string }[]; resellers: { id: number; email: string; display_name: string }[] }
interface ZovoMovement { id: number; event_time: string; balance_after?: string; income?: string; expense?: string; currency: string; nature: string }
interface ZovoBalance { event_time: string; balance: string; available?: string; currency: string }
interface Report { cash_in: string; cash_out: string; manual_cash_out?: string; zovo_cash_out?: string; net_cash_flow: string; internal_transfers: string; items: CashFlowRow[]; zovo_movements: ZovoMovement[]; zovo_balance?: ZovoBalance; zovo_synced_at?: string; pending_settlements: PendingSettlement[] }
const localDate = (d: Date) => new Intl.DateTimeFormat('en-CA', { timeZone: 'Africa/Accra', year: 'numeric', month: '2-digit', day: '2-digit' }).format(d)
const today = localDate(new Date())
const startDate = ref(today)
const endDate = ref(today)
const report = ref<Report | null>(null)
const loading = ref(false)
const saving = ref(false)
const syncingZovo = ref(false)
const error = ref('')
const editingID = ref<number | null>(null)
const accountOptions = [
  { value: 'alipay', label: '支付宝' },
  { value: 'wechat', label: '微信' },
  { value: 'xianyu', label: '咸鱼' },
  { value: 'bank_transfer', label: '银行转账' },
  { value: 'binance', label: '币安' },
]
const counterAccountOptions = [...accountOptions, { value: 'zovocard', label: 'ZovoCard 上游账户' }]
const cashFlowOptions = ref<CashFlowOptions>({ orders: [], resellers: [] })
const orderOptionsOpen = ref(false)
const counterpartyOptionsOpen = ref(false)
const orderOptions = computed(() => {
  const choices = cashFlowOptions.value.orders.map(row => row.order_no)
  if (form.order_no && !choices.includes(form.order_no)) choices.unshift(form.order_no)
  const keyword = form.order_no.trim().toLowerCase()
  return choices.filter(value => !keyword || value.toLowerCase().includes(keyword))
})
const resellerOptions = computed(() => cashFlowOptions.value.resellers.map(row => ({
  value: row.display_name ? `${row.display_name}（${row.email}）` : row.email,
  label: row.display_name ? `${row.display_name}（${row.email}）` : row.email,
})).filter(row => !form.counterparty.trim() || row.label.toLowerCase().includes(form.counterparty.trim().toLowerCase())))
const form = reactive({ event_date: today, direction: 'in', category: 'customer_receipt', account: 'alipay', counter_account: '', amount: '', currency: 'CNY', exchange_rate_snapshot: '', counterparty: '', reference: '', order_no: '', note: '' })
const legacyAccountOption = computed(() => editingID.value && form.account && !accountOptions.some(option => option.value === form.account) ? form.account : '')
const rows = computed(() => report.value?.items || [])
const pending = computed(() => report.value?.pending_settlements || [])
const money = (amount?: string | number, currency = 'CNY') => formatMoney(Number(amount || 0).toFixed(2), currency)
const zovoMoney = (amount?: string, currency = 'USD') => amount ? `${Number(amount).toFixed(2)} ${currency === 'USD' ? '美元' : currency === 'CNY' ? '人民币' : currency}` : '—'
const zovoTime = (value: string) => new Intl.DateTimeFormat('zh-CN', { timeZone: 'Africa/Accra', year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false }).format(new Date(value))
const categoryLabel = (key: string) => ({ customer_receipt: '客户收款', wallet_funding: '收到转款并增加余额', channel_settlement: '支付渠道结算', upstream_recharge: '上游充值', direct_cost: '交付直接成本', operating_expense: '平台费用', platform_development: '平台发展支出', reseller_payout: '支付代理收益', owner_funding: '站长垫资', account_transfer: '账户间转账', refund: '退款', other: '其他收付' }[key] || key)
const directionLabel = (v: string) => v === 'in' ? '流入' : v === 'out' ? '流出' : '内部转账'
const accountLabel = (value: string) => counterAccountOptions.find(option => option.value === value)?.label || value
function handleCategoryChange(value: unknown) {
  if (value === 'upstream_recharge') {
    form.direction = 'transfer'
    form.counter_account = 'zovocard'
  } else if (value === 'platform_development') {
    form.direction = 'out'
  }
}
function setRange(days: number) {
  const end = new Date(`${today}T12:00:00Z`)
  if (days === 0) end.setUTCDate(end.getUTCDate() - 1)
  const start = new Date(end); start.setUTCDate(start.getUTCDate() - days + 1)
  startDate.value = localDate(start); endDate.value = localDate(end)
}
async function load() {
  loading.value = true; error.value = ''
  try {
    const reportResponse = await adminAPI.getFinanceCashFlowReport({ start_date: startDate.value, end_date: endDate.value })
    report.value = reportResponse.data.data as Report
  } catch (e: any) { error.value = e?.message || '读取现金流量数据失败' }
  finally { loading.value = false }
}
async function loadCashFlowOptions() {
  try {
    const response = await adminAPI.getFinanceCashFlowOptions()
    cashFlowOptions.value = response.data.data as CashFlowOptions
  } catch { /* Option lists are optional; manual text entry remains available. */ }
}
function closeOrderOptions() { window.setTimeout(() => { orderOptionsOpen.value = false }, 150) }
function closeCounterpartyOptions() { window.setTimeout(() => { counterpartyOptionsOpen.value = false }, 150) }
async function syncZovo() {
  error.value = ''; syncingZovo.value = true
  try { await adminAPI.syncFinanceCashFlow(); await load() }
  catch (e: any) { error.value = e?.message || '上游资金同步失败' }
  finally { syncingZovo.value = false }
}
async function save() {
  error.value = ''
  const amount = Number(form.amount)
  if (!form.event_date || !Number.isFinite(amount) || amount <= 0 || !form.account.trim()) { error.value = '请填写日期、账户和大于 0 的金额'; return }
  if (form.direction === 'transfer' && (!form.counter_account.trim() || form.counter_account === form.account)) { error.value = '内部转账需要选择不同的转出和转入账户'; return }
  if (form.category === 'upstream_recharge' && (form.direction !== 'transfer' || form.counter_account !== 'zovocard')) { error.value = '上游充值需要登记为转入 ZovoCard 上游账户的内部转账'; return }
  saving.value = true
  try {
    const payload = { ...form, amount: amount.toFixed(6), exchange_rate_snapshot: form.exchange_rate_snapshot ? Number(form.exchange_rate_snapshot) : undefined }
    if (editingID.value) await adminAPI.updateFinanceCashFlow(editingID.value, payload)
    else await adminAPI.createFinanceCashFlow(payload)
    reset(); await load()
  } catch (e: any) { error.value = e?.message || '保存现金流水失败' }
  finally { saving.value = false }
}
function edit(row: CashFlowRow) {
  editingID.value = row.id
  Object.assign(form, { event_date: row.event_date, direction: row.direction, category: row.category, account: row.account, counter_account: row.counter_account || '', amount: String(row.amount), currency: row.currency, exchange_rate_snapshot: String(row.exchange_rate_snapshot), counterparty: row.counterparty || '', reference: row.reference || '', order_no: row.order_no || '', note: row.note || '' })
  window.scrollTo({ top: 0, behavior: 'smooth' })
}
function reset() { editingID.value = null; Object.assign(form, { event_date: endDate.value || today, direction: 'in', category: 'customer_receipt', account: 'alipay', counter_account: '', amount: '', currency: 'CNY', exchange_rate_snapshot: '', counterparty: '', reference: '', order_no: '', note: '' }) }
async function remove(id: number) {
  if (!window.confirm('确定删除这条现金流水登记？')) return
  try { await adminAPI.deleteFinanceCashFlow(id); await load() } catch (e: any) { error.value = e?.message || '删除失败' }
}
onMounted(() => { void load(); void loadCashFlowOptions() })
</script>

<template>
  <div class="space-y-6 p-4 md:p-6">
    <div class="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
      <div><h1 class="text-2xl font-semibold tracking-tight">收付实现制现金流量表</h1><p class="mt-1 text-sm text-muted-foreground">按实际收款、付款日期展示现金流；平台账户间转账单列，不重复计入现金流入和流出。</p></div>
      <div class="flex flex-wrap items-end gap-2">
        <Button size="sm" variant="outline" @click="setRange(1)">今天</Button><Button size="sm" variant="outline" @click="setRange(0)">昨天</Button><Button size="sm" variant="outline" @click="setRange(3)">近三天</Button>
        <label class="grid gap-1 text-xs"><span class="text-muted-foreground">开始日期</span><Input v-model="startDate" type="date" class="w-40" /></label>
        <label class="grid gap-1 text-xs"><span class="text-muted-foreground">结束日期</span><Input v-model="endDate" type="date" class="w-40" /></label>
        <Button variant="outline" :disabled="syncingZovo" @click="syncZovo"><RefreshCw class="mr-2 h-4 w-4" :class="syncingZovo && 'animate-spin'" />{{ syncingZovo ? '同步中…' : '同步上游资金' }}</Button>
        <Button variant="outline" :disabled="loading" @click="load">刷新本地数据</Button>
      </div>
    </div>

    <div v-if="error" class="rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive">{{ error }}</div>
    <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <Card><CardHeader class="pb-2"><CardTitle class="flex items-center gap-2 text-sm text-muted-foreground"><Wallet class="h-4 w-4" />现金流入</CardTitle></CardHeader><CardContent><div class="text-2xl font-semibold">{{ money(report?.cash_in) }}</div><p class="mt-1 text-xs text-muted-foreground">所选区间已登记并确认到账</p></CardContent></Card>
      <Card><CardHeader class="pb-2"><CardTitle class="flex items-center gap-2 text-sm text-muted-foreground"><CircleDollarSign class="h-4 w-4" />现金流出</CardTitle></CardHeader><CardContent><div class="text-2xl font-semibold">{{ money(report?.cash_out) }}</div><p class="mt-1 text-xs text-muted-foreground">包含手工支出、Zovo 已确认消费与费用；卡片余额转入不重复计支出</p></CardContent></Card>
      <Card><CardHeader class="pb-2"><CardTitle class="text-sm text-muted-foreground">净现金流</CardTitle></CardHeader><CardContent><div class="text-2xl font-semibold">{{ money(report?.net_cash_flow) }}</div><p class="mt-1 text-xs text-muted-foreground">现金流入减现金流出</p></CardContent></Card>
      <Card><CardHeader class="pb-2"><CardTitle class="flex items-center gap-2 text-sm text-muted-foreground"><ArrowLeftRight class="h-4 w-4" />内部转账</CardTitle></CardHeader><CardContent><div class="text-2xl font-semibold">{{ money(report?.internal_transfers) }}</div><p class="mt-1 text-xs text-muted-foreground">单列展示，不影响合并净现金流</p></CardContent></Card>
    </div>

    <Card><CardHeader><CardTitle>{{ editingID ? '修改现金流水' : '登记实际收付' }}</CardTitle></CardHeader><CardContent class="space-y-4">
      <p class="text-sm text-muted-foreground">线上支付成功但资金尚未结算到账的订单会显示在下方“待结算”中，不要在此重复登记。Zovo 已同步的开卡费、服务费及已结算卡片消费会自动计入现金流出；平台发展等其他支出请选择“现金流出”并登记，避免重复登记 Zovo 流水。收到客户转款并加余额时，登记为现金流入并关联钱包流水；此款是客户余额负债，不是订单收入。</p>
      <form class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4" @submit.prevent="save">
        <label class="grid gap-1 text-sm"><span>实际发生日期</span><Input v-model="form.event_date" type="date" required /></label>
        <label class="grid gap-1 text-sm"><span>方向</span><Select v-model="form.direction"><SelectTrigger><SelectValue /></SelectTrigger><SelectContent><SelectItem value="in">现金流入</SelectItem><SelectItem value="out">现金流出</SelectItem><SelectItem value="transfer">平台账户间转账</SelectItem></SelectContent></Select></label>
        <label class="grid gap-1 text-sm"><span>业务类型</span><Select v-model="form.category" @update:model-value="handleCategoryChange"><SelectTrigger><SelectValue /></SelectTrigger><SelectContent><SelectItem value="customer_receipt">客户收款</SelectItem><SelectItem value="wallet_funding">转款并增加客户余额</SelectItem><SelectItem value="channel_settlement">支付渠道结算</SelectItem><SelectItem value="upstream_recharge">上游充值</SelectItem><SelectItem value="direct_cost">交付直接成本</SelectItem><SelectItem value="operating_expense">平台运营费用</SelectItem><SelectItem value="platform_development">平台发展支出</SelectItem><SelectItem value="reseller_payout">支付代理收益</SelectItem><SelectItem value="owner_funding">站长垫资</SelectItem><SelectItem value="account_transfer">账户间转账</SelectItem><SelectItem value="refund">退款</SelectItem><SelectItem value="other">其他</SelectItem></SelectContent></Select><span v-if="form.category === 'upstream_recharge'" class="text-xs text-muted-foreground">按平台账户间转账登记，不计为费用</span></label>
        <label class="grid gap-1 text-sm"><span>{{ form.direction === 'transfer' ? '转出账户' : '账户' }}</span><Select v-model="form.account"><SelectTrigger><SelectValue placeholder="选择账户" /></SelectTrigger><SelectContent><SelectItem v-for="option in accountOptions" :key="option.value" :value="option.value">{{ option.label }}</SelectItem><SelectItem v-if="legacyAccountOption" :value="legacyAccountOption">{{ legacyAccountOption }}</SelectItem></SelectContent></Select></label>
        <label v-if="form.direction === 'transfer'" class="grid gap-1 text-sm"><span>转入账户</span><Select v-model="form.counter_account"><SelectTrigger><SelectValue placeholder="选择转入账户" /></SelectTrigger><SelectContent><SelectItem v-for="option in counterAccountOptions" :key="option.value" :value="option.value">{{ option.label }}</SelectItem><SelectItem v-if="editingID && form.counter_account && !counterAccountOptions.some(option => option.value === form.counter_account)" :value="form.counter_account">{{ form.counter_account }}</SelectItem></SelectContent></Select></label>
        <label class="grid gap-1 text-sm"><span>金额</span><Input v-model="form.amount" type="number" min="0.000001" step="0.000001" placeholder="0.00" required /></label>
        <label class="grid gap-1 text-sm"><span>币种</span><Select v-model="form.currency"><SelectTrigger><SelectValue /></SelectTrigger><SelectContent><SelectItem value="CNY">CNY</SelectItem><SelectItem value="USD">USD</SelectItem><SelectItem value="USDT">USDT</SelectItem></SelectContent></Select></label>
        <label class="grid gap-1 text-sm"><span>折算汇率（留空按后台配置）</span><Input v-model="form.exchange_rate_snapshot" type="number" min="0.000001" step="0.000001" placeholder="CNY/币种；USD、USDT 使用后台 USD/CNY 汇率" /></label>
        <label class="relative grid gap-1 text-sm"><span>对方（选填）</span><Input v-model="form.counterparty" placeholder="输入代理名称或邮箱搜索" autocomplete="off" @focus="counterpartyOptionsOpen = true" @input="counterpartyOptionsOpen = true" @blur="closeCounterpartyOptions" /><div v-if="counterpartyOptionsOpen && resellerOptions.length" class="absolute inset-x-0 top-full z-30 mt-1 max-h-60 overflow-y-auto rounded-md border bg-popover p-1 text-popover-foreground shadow-md"><button v-for="option in resellerOptions" :key="option.value" type="button" class="block w-full rounded-sm px-3 py-2 text-left hover:bg-accent" @mousedown.prevent="form.counterparty = option.value; counterpartyOptionsOpen = false">{{ option.label }}</button></div></label>
        <label class="relative grid gap-1 text-sm"><span>关联订单号（选填）</span><Input v-model="form.order_no" placeholder="搜索未关联的已成交订单" autocomplete="off" @focus="orderOptionsOpen = true" @input="orderOptionsOpen = true" @blur="closeOrderOptions" /><div v-if="orderOptionsOpen && orderOptions.length" class="absolute inset-x-0 top-full z-30 mt-1 max-h-60 overflow-y-auto rounded-md border bg-popover p-1 text-popover-foreground shadow-md"><button v-for="orderNo in orderOptions" :key="orderNo" type="button" class="block w-full rounded-sm px-3 py-2 text-left hover:bg-accent" @mousedown.prevent="form.order_no = orderNo; orderOptionsOpen = false">{{ orderNo }}</button></div></label>
        <label class="grid gap-1 text-sm"><span>外部流水号 / 钱包流水引用</span><Input v-model="form.reference" placeholder="用于对账和去重" /></label>
        <label class="grid gap-1 text-sm sm:col-span-2 xl:col-span-2"><span>备注</span><Input v-model="form.note" placeholder="收款渠道、用途或关联说明" /></label>
        <div class="flex items-end gap-2"><Button type="submit" :disabled="saving"><Save class="mr-2 h-4 w-4" />{{ editingID ? '保存修改' : '登记流水' }}</Button><Button v-if="editingID" type="button" variant="outline" @click="reset">取消</Button></div>
      </form>
    </CardContent></Card>

    <Card><CardHeader><CardTitle>现金流水明细</CardTitle></CardHeader><CardContent><div class="overflow-x-auto"><table class="w-full min-w-[1100px] text-left text-sm"><thead><tr class="border-b text-muted-foreground"><th class="py-3">日期</th><th>方向</th><th>业务类型</th><th>账户</th><th>对方 / 关联</th><th>原币金额</th><th>折算人民币</th><th>来源备注</th><th>操作</th></tr></thead><tbody>
      <tr v-for="row in rows" :key="row.id" class="border-b last:border-0"><td class="py-3">{{ row.event_date }}</td><td>{{ directionLabel(row.direction) }}</td><td>{{ categoryLabel(row.category) }}</td><td>{{ accountLabel(row.account) }}{{ row.counter_account ? ` → ${accountLabel(row.counter_account)}` : '' }}</td><td>{{ row.counterparty || row.order_no || row.reference || '—' }}</td><td>{{ money(row.amount, row.currency) }}</td><td>{{ money(row.cny_amount) }}<span v-if="row.currency !== 'CNY'" class="block text-xs text-muted-foreground">汇率 {{ row.exchange_rate_snapshot }}</span></td><td class="max-w-[220px] truncate">{{ row.note || row.related_source || '手工确认' }}</td><td class="whitespace-nowrap"><Button size="sm" variant="ghost" @click="edit(row)">编辑</Button><Button size="icon" variant="ghost" aria-label="删除" @click="remove(row.id)"><Trash2 class="h-4 w-4" /></Button></td></tr>
      <tr v-if="!rows.length"><td colspan="9" class="py-8 text-center text-muted-foreground">所选区间暂无已登记现金流水</td></tr>
    </tbody></table></div></CardContent></Card>

    <Card><CardHeader><CardTitle class="flex items-center gap-2"><Clock3 class="h-4 w-4" />支付成功、等待结算</CardTitle></CardHeader><CardContent><p class="mb-3 text-sm text-muted-foreground">这些支付已由渠道确认成功，但尚未登记实际结算到账。它们是待结算款项，不计入上方现金流入；渠道结算到账后，请登记一笔“支付渠道结算”，并填写关联订单或支付编号。</p><div class="overflow-x-auto"><table class="w-full min-w-[700px] text-left text-sm"><thead><tr class="border-b text-muted-foreground"><th class="py-3">支付时间</th><th>支付编号</th><th>订单号</th><th>渠道</th><th>已支付金额</th></tr></thead><tbody><tr v-for="row in pending" :key="row.payment_id" class="border-b last:border-0"><td class="py-3">{{ new Date(row.paid_at).toLocaleString() }}</td><td>{{ row.payment_id }}</td><td>{{ row.order_no }}</td><td>{{ row.channel }}</td><td>{{ money(row.amount, row.currency) }}</td></tr><tr v-if="!pending.length"><td colspan="5" class="py-8 text-center text-muted-foreground">所选区间没有待结算线上支付</td></tr></tbody></table></div></CardContent></Card>
    <Card>
      <CardHeader><CardTitle>上游资金明细</CardTitle></CardHeader>
      <CardContent class="space-y-4">
        <div class="flex flex-wrap items-center justify-between gap-2 text-sm text-muted-foreground">
          <span>{{ report?.zovo_synced_at ? `最近同步：${zovoTime(report.zovo_synced_at)}` : '尚未同步资金记录' }}</span>
          <span>明细读取本站已保存的数据</span>
        </div>
        <div v-if="report?.zovo_balance" class="flex flex-wrap gap-x-8 gap-y-3 rounded-lg border bg-muted/30 px-4 py-3 text-sm">
          <div><div class="text-xs text-muted-foreground">账户余额</div><div class="mt-1 font-semibold">{{ zovoMoney(report.zovo_balance.balance, report.zovo_balance.currency) }}</div></div>
          <div v-if="report.zovo_balance.available"><div class="text-xs text-muted-foreground">可用余额</div><div class="mt-1 font-semibold">{{ zovoMoney(report.zovo_balance.available, report.zovo_balance.currency) }}</div></div>
          <div><div class="text-xs text-muted-foreground">余额时间</div><div class="mt-1 font-medium">{{ zovoTime(report.zovo_balance.event_time) }}</div></div>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full min-w-[720px] text-left text-sm">
            <thead><tr class="border-b text-muted-foreground"><th class="py-3">时间</th><th>收入</th><th>支出</th><th>余额</th><th>性质</th></tr></thead>
            <tbody>
              <tr v-for="row in report?.zovo_movements || []" :key="row.id" class="border-b last:border-0">
                <td class="py-3">{{ zovoTime(row.event_time) }}</td>
                <td class="text-emerald-700">{{ zovoMoney(row.income, row.currency) }}</td>
                <td class="text-rose-700">{{ zovoMoney(row.expense, row.currency) }}</td>
                <td>{{ zovoMoney(row.balance_after, row.currency) }}</td>
                <td>{{ row.nature }}</td>
              </tr>
              <tr v-if="!report?.zovo_movements?.length"><td colspan="5" class="py-8 text-center text-muted-foreground">所选日期没有资金收支记录</td></tr>
            </tbody>
          </table>
        </div>
      </CardContent>
    </Card>
    <p class="text-xs text-muted-foreground">余额快照只显示最新一条；资金收支以账户账单为准，不重复列出卡片交易、卡密服务费等对应记录。</p>
  </div>
</template>
