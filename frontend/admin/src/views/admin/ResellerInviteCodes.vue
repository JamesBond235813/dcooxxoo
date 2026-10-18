<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { adminAPI } from '@/api/admin'
import IdCell from '@/components/IdCell.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import TableSkeleton from '@/components/TableSkeleton.vue'
import ListPagination from '@/components/ListPagination.vue'
import { useListRefresh, type ListFetchOptions } from '@/composables/useListRefresh'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { confirmAction } from '@/utils/confirm'
import { notifyError, notifySuccess } from '@/utils/notify'
import { formatDate } from '@/utils/format'
import { Copy, Plus, RefreshCw, Trash2, Power, CheckCheck, Pencil } from 'lucide-vue-next'

interface InviteCodeItem {
  id: number
  code: string
  status: string
  memo: string
  used_count: number
  max_uses: number
  expires_at: string | null
  created_at: string
}

const loading = ref(true)
const { refreshing, refreshList } = useListRefresh()
const operatingId = ref<number | null>(null)
const rows = ref<InviteCodeItem[]>([])
const pagination = ref({
  page: 1,
  page_size: 20,
  total: 0,
  total_page: 1,
})

const filterStatus = ref('__all__')
const pageSizeOptions = [10, 20, 50, 100]

// Dialog for generating invite codes
const isCreateDialogOpen = ref(false)
const createSubmitting = ref(false)
const createForm = reactive({
  mode: 'single', // 'single' | 'batch'
  code: '',
  count: 5,
  max_uses: 1,
  memo: '',
  expires_at: '',
})
const isMemoDialogOpen = ref(false)
const memoSubmitting = ref(false)
const memoTarget = ref<InviteCodeItem | null>(null)
const memoDraft = ref('')

const fetchRows = async (page = 1, options: ListFetchOptions = {}) => {
  if (!options.preserveRows) loading.value = true
  try {
    const statusParam = filterStatus.value === '__all__' ? undefined : filterStatus.value
    const res = await adminAPI.getResellerInviteCodes({
      page,
      page_size: pagination.value.page_size,
      status: statusParam,
    })
    const payload = res.data?.data
    if (payload && Array.isArray(payload.items)) {
      rows.value = payload.items
      if (payload.pagination) {
        pagination.value = {
          page: Number(payload.pagination.page) || page,
          page_size: Number(payload.pagination.page_size) || pagination.value.page_size,
          total: Number(payload.pagination.total) || 0,
          total_page: Number(payload.pagination.total_page) || 1,
        }
      }
    } else if (Array.isArray(payload)) {
      rows.value = payload
      if (res.data?.pagination) {
        pagination.value = res.data.pagination
      }
    } else {
      rows.value = []
    }
  } catch (err: any) {
    if (!options.preserveRows) rows.value = []
    notifyError(err?.message || '获取邀请码列表失败')
  } finally {
    if (!options.preserveRows) loading.value = false
  }
}

const handleStatusChange = () => {
  fetchRows(1, { preserveRows: true })
}

const refreshCurrentPage = () => {
  refreshList(() => fetchRows(pagination.value.page, { preserveRows: true }))
}

const changePage = (page: number) => {
  if (page < 1 || page > pagination.value.total_page) return
  fetchRows(page)
}

const changePageSize = (size: number) => {
  if (size === pagination.value.page_size) return
  pagination.value.page_size = size
  fetchRows(1)
}

const openCreateDialog = () => {
  createForm.mode = 'single'
  createForm.code = ''
  createForm.count = 5
  createForm.max_uses = 1
  createForm.memo = ''
  createForm.expires_at = ''
  isCreateDialogOpen.value = true
}

const handleCreateSubmit = async () => {
  createSubmitting.value = true
  try {
    const payload: { count?: number; max_uses?: number; memo?: string; code?: string; expires_at?: string } = {
      max_uses: Number.isFinite(Number(createForm.max_uses)) ? Math.max(0, Math.trunc(Number(createForm.max_uses))) : 1,
      memo: createForm.memo.trim(),
    }
    if (createForm.expires_at) payload.expires_at = new Date(createForm.expires_at).toISOString()
    if (createForm.mode === 'single') {
      if (createForm.code.trim()) {
        payload.code = createForm.code.trim().toUpperCase()
      }
      payload.count = 1
    } else {
      payload.count = Math.max(1, Math.min(100, Number(createForm.count) || 5))
    }

    const res = await adminAPI.createResellerInviteCodes(payload)
    const countCreated = Number(res?.data?.data?.count) || 0
    const failedCount = Number(res?.data?.data?.failed_count) || 0
    if (failedCount > 0) notifyError(`请求生成 ${res?.data?.data?.requested_count || countCreated + failedCount} 个，成功 ${countCreated} 个，失败 ${failedCount} 个`)
    else notifySuccess(`成功生成 ${countCreated} 个邀请码`)
    isCreateDialogOpen.value = false
    fetchRows(1, { preserveRows: true })
  } catch (err: any) {
    notifyError(err?.message || '生成邀请码失败')
  } finally {
    createSubmitting.value = false
  }
}

const copiedCodes = ref<Record<string, boolean>>({})
const batchCopied = ref(false)
const isExpired = (item: InviteCodeItem) => Boolean(item.expires_at && new Date(item.expires_at).getTime() <= Date.now())

const copyCode = async (code: string) => {
  try {
    await navigator.clipboard.writeText(code)
    copiedCodes.value[code] = true
    notifySuccess(`邀请码 ${code} 已复制到剪贴板`)
    setTimeout(() => {
      copiedCodes.value[code] = false
    }, 2000)
  } catch {
    notifyError('复制失败，请手动复制')
  }
}

const copyActiveCodes = async () => {
  const activeCodes = rows.value.filter(r => r.status === 'active').map(r => r.code)
  if (activeCodes.length === 0) {
    notifyError('当前页暂无有效邀请码可复制')
    return
  }
  try {
    await navigator.clipboard.writeText(activeCodes.join('\n'))
    batchCopied.value = true
    notifySuccess(`已批量复制本页 ${activeCodes.length} 个有效邀请码到剪贴板`)
    setTimeout(() => {
      batchCopied.value = false
    }, 2000)
  } catch {
    notifyError('复制失败，请重试')
  }
}

const handleToggleStatus = async (item: InviteCodeItem) => {
  const isCurrentlyActive = item.status === 'active'
  const actionText = isCurrentlyActive ? '停用' : '启用'
  const confirmed = await confirmAction({
    title: `${actionText}邀请码`,
    description: `确定要${actionText}邀请码 [${item.code}] 吗？`,
    confirmText: actionText,
    variant: isCurrentlyActive ? 'destructive' : 'default',
  })
  if (!confirmed) return

  operatingId.value = item.id
  try {
    await adminAPI.toggleResellerInviteCode(item.id)
    notifySuccess(`邀请码已${actionText}`)
    refreshCurrentPage()
  } catch (err: any) {
    notifyError(err?.message || `${actionText}失败`)
  } finally {
    operatingId.value = null
  }
}

const handleDelete = async (item: InviteCodeItem) => {
  const confirmed = await confirmAction({
    title: '删除邀请码',
    description: `确定要彻底删除邀请码 [${item.code}] 吗？删除后将无法恢复。`,
    confirmText: '删除',
    variant: 'destructive',
  })
  if (!confirmed) return

  operatingId.value = item.id
  try {
    await adminAPI.deleteResellerInviteCode(item.id)
    notifySuccess('邀请码已删除')
    refreshCurrentPage()
  } catch (err: any) {
    notifyError(err?.message || '删除失败')
  } finally {
    operatingId.value = null
  }
}

const openMemoEditor = (item: InviteCodeItem) => {
  memoTarget.value = item
  memoDraft.value = item.memo || ''
  isMemoDialogOpen.value = true
}

const handleMemoSubmit = async () => {
  if (!memoTarget.value || memoSubmitting.value) return
  memoSubmitting.value = true
  try {
    await adminAPI.updateResellerInviteCodeMemo(memoTarget.value.id, memoDraft.value.trim())
    notifySuccess('备注说明已保存')
    isMemoDialogOpen.value = false
    await fetchRows(pagination.value.page, { preserveRows: true })
  } catch (err: any) {
    notifyError(err?.message || '保存备注说明失败')
  } finally {
    memoSubmitting.value = false
  }
}

onMounted(() => {
  fetchRows()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">代理邀请码管理</h1>
        <p class="text-xs text-muted-foreground mt-1">
          生成并管理专属代理注册邀请码。普通用户注册代理身份时必须校验该邀请码。
        </p>
      </div>

      <div class="flex items-center gap-2">
        <Button variant="outline" size="sm" :disabled="refreshing" @click="refreshCurrentPage">
          <RefreshCw class="w-4 h-4 mr-1.5" :class="{ 'animate-spin': refreshing }" />
          刷新
        </Button>
        <Button variant="outline" size="sm" @click="copyActiveCodes">
          <CheckCheck v-if="batchCopied" class="w-4 h-4 mr-1.5 text-sky-500 animate-in zoom-in-50 duration-150" />
          <Copy v-else class="w-4 h-4 mr-1.5" />
          {{ batchCopied ? '已复制' : '复制本页有效码' }}
        </Button>
        <Button size="sm" @click="openCreateDialog">
          <Plus class="w-4 h-4 mr-1.5" />
          生成邀请码
        </Button>
      </div>
    </div>

    <!-- Filter Bar -->
    <div class="flex items-center gap-4 bg-card p-4 rounded-xl border">
      <div class="flex items-center gap-2">
        <span class="text-xs text-muted-foreground whitespace-nowrap">状态筛选:</span>
        <Select v-model="filterStatus" @update:model-value="handleStatusChange">
          <SelectTrigger class="w-[140px] h-9 text-xs">
            <SelectValue placeholder="全部状态" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="__all__">全部状态</SelectItem>
            <SelectItem value="active">有效 (active)</SelectItem>
            <SelectItem value="exhausted">已耗尽 (exhausted)</SelectItem>
            <SelectItem value="disabled">已停用 (disabled)</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>

    <!-- Table -->
    <div class="rounded-xl border bg-card">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead class="w-[80px]">ID</TableHead>
            <TableHead class="min-w-[180px]">邀请码</TableHead>
            <TableHead class="min-w-[120px]">状态</TableHead>
            <TableHead class="min-w-[120px]">使用次数 / 上限</TableHead>
            <TableHead class="min-w-[150px]">备注说明</TableHead>
            <TableHead class="min-w-[160px]">有效期至</TableHead>
            <TableHead class="min-w-[160px]">创建时间</TableHead>
            <TableHead class="text-right w-[160px]">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableSkeleton v-if="loading" :col-span="8" :row-count="5" />
          <TableRow v-else-if="rows.length === 0">
            <TableCell colspan="8" class="h-32 text-center text-muted-foreground text-sm">
              暂无代理邀请码数据，点击右上角【生成邀请码】按钮立即创建
            </TableCell>
          </TableRow>
          <TableRow v-else v-for="item in rows" :key="item.id">
            <TableCell>
              <IdCell :value="item.id" />
            </TableCell>
            <TableCell>
              <div class="flex items-center gap-2">
                <span class="font-mono font-semibold tracking-wider text-sm select-all">
                  {{ item.code }}
                </span>
                <button
                  type="button"
                  @click="copyCode(item.code)"
                  class="transition-all p-1 rounded hover:bg-muted/80 inline-flex items-center justify-center"
                  :class="copiedCodes[item.code] ? 'text-sky-500 scale-110' : 'text-muted-foreground hover:text-foreground'"
                  :title="copiedCodes[item.code] ? '已复制' : '复制邀请码'"
                >
                  <CheckCheck v-if="copiedCodes[item.code]" class="w-3.5 h-3.5 text-sky-500 animate-in zoom-in-50 duration-150" />
                  <Copy v-else class="w-3.5 h-3.5" />
                </button>
              </div>
            </TableCell>
            <TableCell>
              <span
                v-if="item.status === 'active' && !isExpired(item)"
                class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300"
              >
                有效
              </span>
              <span
                v-else-if="isExpired(item)"
                class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300"
              >
                已过期
              </span>
              <span
                v-else-if="item.status === 'exhausted'"
                class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300"
              >
                已用尽
              </span>
              <span
                v-else
                class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300"
              >
                已停用
              </span>
            </TableCell>
            <TableCell>
              <span class="font-medium text-xs">
                {{ item.used_count }}
                <span class="text-muted-foreground font-normal">
                  / {{ item.max_uses === 0 ? '无限制' : item.max_uses }}
                </span>
              </span>
            </TableCell>
            <TableCell>
              <button type="button" class="group inline-flex max-w-[280px] items-center gap-1.5 text-left text-xs text-muted-foreground hover:text-foreground" @click="openMemoEditor(item)" title="编辑备注说明">
                {{ item.memo || '-' }}
                <Pencil class="h-3 w-3 shrink-0 opacity-50 transition group-hover:opacity-100" />
              </button>
            </TableCell>
            <TableCell>
              <span class="text-xs text-muted-foreground">
                {{ item.expires_at ? formatDate(item.expires_at) : '永不过期' }}
              </span>
            </TableCell>
            <TableCell>
              <span class="text-xs text-muted-foreground">
                {{ formatDate(item.created_at) }}
              </span>
            </TableCell>
            <TableCell class="text-right">
              <div class="flex items-center justify-end gap-1">
                <Button
                  variant="ghost"
                  size="sm"
                  class="h-8 px-2 text-xs"
                  :disabled="operatingId === item.id"
                  @click="handleToggleStatus(item)"
                >
                  <Power class="w-3.5 h-3.5 mr-1" />
                  {{ item.status === 'active' ? '停用' : '启用' }}
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  class="h-8 px-2 text-xs text-rose-600 hover:text-rose-700 hover:bg-rose-50 dark:hover:bg-rose-950"
                  :disabled="operatingId === item.id"
                  @click="handleDelete(item)"
                >
                  <Trash2 class="w-3.5 h-3.5" />
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>

    <!-- Pagination -->
    <ListPagination
      :page="pagination.page"
      :total-page="pagination.total_page"
      :total="pagination.total"
      :page-size="pagination.page_size"
      :page-size-options="pageSizeOptions"
      @change-page="changePage"
      @change-page-size="changePageSize"
    />

    <!-- Create / Batch Generate Dialog -->
    <Dialog :open="isCreateDialogOpen" @update:open="isCreateDialogOpen = $event">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>生成代理邀请码</DialogTitle>
          <DialogDescription>
            支持自定义单个邀请码或由系统随机批量生成。每个邀请码可设定最大使用次数。
          </DialogDescription>
        </DialogHeader>

        <form @submit.prevent="handleCreateSubmit" class="space-y-4 py-2">
          <!-- 生成模式选择 -->
          <div class="space-y-1.5">
            <label class="text-xs font-medium">生成方式</label>
            <div class="grid grid-cols-2 gap-2">
              <button
                type="button"
                @click="createForm.mode = 'single'"
                class="py-2 px-3 text-xs font-semibold rounded-lg border text-center transition"
                :class="createForm.mode === 'single' ? 'bg-primary text-primary-foreground border-primary' : 'bg-background hover:bg-muted'"
              >
                单个生成 (可自定义)
              </button>
              <button
                type="button"
                @click="createForm.mode = 'batch'"
                class="py-2 px-3 text-xs font-semibold rounded-lg border text-center transition"
                :class="createForm.mode === 'batch' ? 'bg-primary text-primary-foreground border-primary' : 'bg-background hover:bg-muted'"
              >
                批量生成 (随机码)
              </button>
            </div>
          </div>

          <!-- 单个模式：自定义码或留空自动生成 -->
          <div v-if="createForm.mode === 'single'" class="space-y-1.5">
            <label class="text-xs font-medium">自定义邀请码 (选填)</label>
            <Input
              v-model="createForm.code"
              placeholder="留空则由系统随机生成（如 ABCD-EFGH-JKMN-2345）"
              class="h-9 text-xs uppercase"
            />
          </div>

          <!-- 批量模式：生成数量 -->
          <div v-else class="space-y-1.5">
            <label class="text-xs font-medium">生成数量 (1 - 100 个)</label>
            <Input
              v-model.number="createForm.count"
              type="number"
              min="1"
              max="100"
              class="h-9 text-xs"
            />
          </div>

          <!-- 最大使用次数 -->
          <div class="space-y-1.5">
            <label class="text-xs font-medium flex items-center justify-between">
              <span>最大使用次数</span>
              <span class="text-[11px] text-muted-foreground font-normal">0 表示无限制使用</span>
            </label>
            <Input
              v-model.number="createForm.max_uses"
              type="number"
              min="0"
              placeholder="默认 1 次（单次核销失效）"
              class="h-9 text-xs"
            />
          </div>

          <!-- 备注说明 -->
          <div class="space-y-1.5">
            <label class="text-xs font-medium">有效期至 (选填)</label>
            <Input v-model="createForm.expires_at" type="datetime-local" class="h-9 text-xs" />
            <p class="text-[11px] text-muted-foreground">留空表示永不过期。</p>
          </div>

          <!-- 备注说明 -->
          <div class="space-y-1.5">
            <label class="text-xs font-medium">备注说明 (选填)</label>
            <Input
              v-model="createForm.memo"
              placeholder="例如：发给张总的专属代理码 / 官方首批渠道活动"
              class="h-9 text-xs"
            />
          </div>

          <DialogFooter class="pt-4">
            <Button type="button" variant="outline" @click="isCreateDialogOpen = false">
              取消
            </Button>
            <Button type="submit" :disabled="createSubmitting">
              {{ createSubmitting ? '生成中...' : '确认生成' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <Dialog :open="isMemoDialogOpen" @update:open="isMemoDialogOpen = $event">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>编辑备注说明</DialogTitle>
          <DialogDescription>邀请码：{{ memoTarget?.code }}</DialogDescription>
        </DialogHeader>
        <form class="space-y-4" @submit.prevent="handleMemoSubmit">
          <div class="space-y-1.5">
            <label class="text-xs font-medium">备注说明</label>
            <Input v-model="memoDraft" maxlength="255" placeholder="请输入备注说明（选填）" class="text-sm" />
            <p class="text-[11px] text-muted-foreground">最多 255 个字符，留空可清除备注。</p>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" :disabled="memoSubmitting" @click="isMemoDialogOpen = false">取消</Button>
            <Button type="submit" :disabled="memoSubmitting">{{ memoSubmitting ? '保存中…' : '保存备注' }}</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </div>
</template>
