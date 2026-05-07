<template>
  <div class="subscription-page">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold text-slate-800">订阅管理</h2>
        <p class="text-slate-500 mt-1 text-sm">管理租户订阅套餐与计费</p>
      </div>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon class="mr-1"><Plus /></el-icon>
        新增订阅
      </el-button>
    </div>

    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <el-table v-loading="loading" :data="subs" stripe class="w-full">
        <el-table-column prop="tenant_id" label="租户ID" width="80" />
        <el-table-column prop="package_name" label="套餐" width="140" />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.status === 'active' ? 'success' : row.status === 'trial' ? 'warning' : row.status === 'expired' ? 'danger' : 'info'"
              size="small"
            >
              {{ statusMap[row.status] || row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="start_date" label="开始日期" width="120">
          <template #default="{ row }">
            <span>{{ formatDate(row.start_date) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="end_date" label="到期日期" width="120">
          <template #default="{ row }">
            <span :class="{ 'text-red-500': isExpiringSoon(row.end_date) }">
              {{ formatDate(row.end_date) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="auto_renew" label="自动续费" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.auto_renew ? 'success' : 'info'" size="small">
              {{ row.auto_renew ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="用量" width="160">
          <template #default="{ row }">
            <div class="text-xs">
              <div>用户: {{ row.current_users }}/{{ row.max_users }}</div>
              <div>存储: {{ formatStorage(row.current_storage) }}/{{ formatStorage(row.max_storage) }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" align="center" fixed="right">
          <template #default="{ row }">
            <el-button size="small" link type="primary" @click="openRenewDialog(row)">
              续费
            </el-button>
            <el-button size="small" link type="primary" @click="openInvoiceDialog(row)">
              账单
            </el-button>
            <el-button size="small" link type="primary" @click="openEditDialog(row)">
              编辑
            </el-button>
            <el-popconfirm
              v-if="row.status === 'active' || row.status === 'trial'"
              title="确定取消该订阅？"
              @confirm="handleCancel(row)"
            >
              <template #reference>
                <el-button size="small" link type="danger">取消</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <div class="p-4 flex justify-end">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @current-change="fetchSubs"
          @size-change="fetchSubs"
        />
      </div>
    </div>

    <el-dialog v-model="formVisible" :title="isEdit ? '编辑订阅' : '新增订阅'" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="租户ID" prop="tenant_id">
          <el-input-number v-model="form.tenant_id" :min="1" />
        </el-form-item>
        <el-form-item label="套餐ID" prop="package_id">
          <el-input-number v-model="form.package_id" :min="1" />
        </el-form-item>
        <el-form-item label="套餐名称" prop="package_name">
          <el-input v-model="form.package_name" placeholder="如: 企业版" />
        </el-form-item>
        <el-form-item label="开始日期" prop="start_date">
          <el-date-picker v-model="form.start_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="到期日期" prop="end_date">
          <el-date-picker v-model="form.end_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="最大用户数">
          <el-input-number v-model="form.max_users" :min="1" />
        </el-form-item>
        <el-form-item label="自动续费">
          <el-switch v-model="form.auto_renew" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="renewVisible" title="续费" width="400px">
      <el-form label-width="80px">
        <el-form-item label="续费月数">
          <el-input-number v-model="renewMonths" :min="1" :max="36" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="renewVisible = false">取消</el-button>
        <el-button type="primary" @click="handleRenew">确定续费</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="invoiceVisible" title="账单记录" width="700px">
      <el-table v-loading="invoiceLoading" :data="invoices" stripe max-height="400">
        <el-table-column prop="amount" label="金额" width="100">
          <template #default="{ row }">
            <span>{{ row.currency }} {{ row.amount }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="billing_period" label="计费周期" width="120" />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'paid' ? 'success' : row.status === 'pending' ? 'warning' : 'info'" size="small">
              {{ row.status === 'paid' ? '已支付' : row.status === 'pending' ? '待支付' : row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="due_date" label="到期日" width="120">
          <template #default="{ row }">
            <span>{{ formatDate(row.due_date) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="paid_at" label="支付时间" width="170">
          <template #default="{ row }">
            <span>{{ formatTime(row.paid_at) }}</span>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import subscriptionApi, { type Subscription, type SubscriptionInvoice } from '@/api/subscription'

const statusMap: Record<string, string> = {
  active: '活跃',
  trial: '试用',
  expired: '已过期',
  cancelled: '已取消'
}

const loading = ref(false)
const subs = ref<Subscription[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const formVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({
  id: 0,
  tenant_id: 1,
  package_id: 1,
  package_name: '',
  start_date: '',
  end_date: '',
  max_users: 10,
  auto_renew: 0 as number
})

const rules: FormRules = {
  tenant_id: [{ required: true, message: '请输入租户ID', trigger: 'blur' }],
  package_id: [{ required: true, message: '请输入套餐ID', trigger: 'blur' }],
  package_name: [{ required: true, message: '请输入套餐名称', trigger: 'blur' }],
  start_date: [{ required: true, message: '请选择开始日期', trigger: 'change' }],
  end_date: [{ required: true, message: '请选择到期日期', trigger: 'change' }]
}

const renewVisible = ref(false)
const renewMonths = ref(1)
const currentSubId = ref(0)

const invoiceVisible = ref(false)
const invoiceLoading = ref(false)
const invoices = ref<SubscriptionInvoice[]>([])

function formatDate(date: string) {
  if (!date) return '-'
  return date.slice(0, 10)
}

function formatTime(time: string | null) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

function formatStorage(bytes: number) {
  if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(1) + 'GB'
  if (bytes >= 1048576) return (bytes / 1048576).toFixed(1) + 'MB'
  return bytes + 'B'
}

function isExpiringSoon(date: string) {
  if (!date) return false
  const end = new Date(date)
  const now = new Date()
  const diff = end.getTime() - now.getTime()
  return diff > 0 && diff < 30 * 24 * 60 * 60 * 1000
}

async function fetchSubs() {
  loading.value = true
  try {
    const res = await subscriptionApi.list(page.value, pageSize.value)
    subs.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch {
    // ignore
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  isEdit.value = false
  form.id = 0
  form.tenant_id = 1
  form.package_id = 1
  form.package_name = ''
  form.start_date = ''
  form.end_date = ''
  form.max_users = 10
  form.auto_renew = 0
  formVisible.value = true
}

function openEditDialog(row: Subscription) {
  isEdit.value = true
  form.id = row.id
  form.tenant_id = row.tenant_id
  form.package_id = row.package_id
  form.package_name = row.package_name
  form.start_date = formatDate(row.start_date)
  form.end_date = formatDate(row.end_date)
  form.max_users = row.max_users
  form.auto_renew = row.auto_renew
  formVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  try {
    if (isEdit.value) {
      await subscriptionApi.update(form.id, form)
      ElMessage.success('更新成功')
    } else {
      await subscriptionApi.create(form)
      ElMessage.success('创建成功')
    }
    formVisible.value = false
    fetchSubs()
  } catch {
    ElMessage.error('操作失败')
  }
}

async function handleCancel(row: Subscription) {
  try {
    await subscriptionApi.cancel(row.id)
    ElMessage.success('已取消订阅')
    fetchSubs()
  } catch {
    ElMessage.error('操作失败')
  }
}

function openRenewDialog(row: Subscription) {
  currentSubId.value = row.id
  renewMonths.value = 1
  renewVisible.value = true
}

async function handleRenew() {
  try {
    await subscriptionApi.renew(currentSubId.value, renewMonths.value)
    ElMessage.success('续费成功')
    renewVisible.value = false
    fetchSubs()
  } catch {
    ElMessage.error('续费失败')
  }
}

async function openInvoiceDialog(row: Subscription) {
  currentSubId.value = row.id
  invoiceVisible.value = true
  invoiceLoading.value = true
  try {
    const res = await subscriptionApi.getInvoices(row.id)
    invoices.value = res.data?.list || []
  } catch {
    // ignore
  } finally {
    invoiceLoading.value = false
  }
}

onMounted(() => {
  fetchSubs()
})
</script>
