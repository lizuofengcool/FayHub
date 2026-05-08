<template>
  <div class="settlement-page">
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 pb-3 flex items-center justify-between">
        <div>
          <h2 class="text-lg font-bold text-slate-800">结算管理</h2>
          <p class="text-slate-400 text-xs mt-0.5">管理平台分账配置与结算记录</p>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-4 gap-4 px-4 mb-4">
        <div class="bg-slate-50 rounded-xl border border-slate-100 p-4">
          <p class="text-sm text-slate-500">总交易额</p>
          <p class="text-2xl font-bold text-slate-800 mt-1">¥{{ formatAmount(stats.total_amount) }}</p>
        </div>
        <div class="bg-slate-50 rounded-xl border border-slate-100 p-4">
          <p class="text-sm text-slate-500">平台收入</p>
          <p class="text-2xl font-bold text-blue-600 mt-1">¥{{ formatAmount(stats.platform_amount) }}</p>
        </div>
        <div class="bg-slate-50 rounded-xl border border-slate-100 p-4">
          <p class="text-sm text-slate-500">租户收入</p>
          <p class="text-2xl font-bold text-green-600 mt-1">¥{{ formatAmount(stats.tenant_amount) }}</p>
        </div>
        <div class="bg-slate-50 rounded-xl border border-slate-100 p-4">
          <p class="text-sm text-slate-500">待结算</p>
          <p class="text-2xl font-bold text-orange-500 mt-1">{{ stats.pending_count || 0 }} 笔</p>
        </div>
      </div>

      <el-tabs v-model="activeTab" class="settlement-tabs">
      <el-tab-pane label="分账配置" name="config">
        <div class="bg-white rounded-2xl border border-slate-100 shadow-sm p-6">
          <div class="flex items-center justify-between mb-6">
            <h3 class="text-lg font-semibold text-slate-800">分账规则</h3>
            <el-button type="primary" @click="saveConfig" :loading="configSaving">
              <el-icon class="mr-1"><Check /></el-icon> 保存配置
            </el-button>
          </div>
          <el-form label-width="140px" style="max-width: 500px">
            <el-form-item label="平台分账比例">
              <el-input-number v-model="configForm.platform_rate" :min="0" :max="10000" :step="100" />
              <span class="ml-2 text-slate-500 text-sm">万分比（1000 = 10%）</span>
            </el-form-item>
            <el-form-item label="最小结算金额">
              <el-input-number v-model="configForm.min_amount" :min="0" :step="100" />
              <span class="ml-2 text-slate-500 text-sm">单位：分（100 = 1元）</span>
            </el-form-item>
            <el-form-item label="当前比例显示">
              <span class="text-lg font-bold text-blue-600">{{ (configForm.platform_rate / 100).toFixed(2) }}%</span>
            </el-form-item>
          </el-form>
        </div>
      </el-tab-pane>

      <el-tab-pane label="结算记录" name="records">
        <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
          <el-table v-loading="recordsLoading" :data="records" stripe class="w-full">
            <el-table-column prop="settlement_no" label="结算单号" min-width="160" />
            <el-table-column prop="order_no" label="关联订单" min-width="160" />
            <el-table-column label="订单金额" width="120" align="right">
              <template #default="{ row }">
                <span class="font-medium">¥{{ formatAmount(row.total_amount) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="平台收入" width="120" align="right">
              <template #default="{ row }">
                <span class="text-blue-600">¥{{ formatAmount(row.platform_amount) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="租户收入" width="120" align="right">
              <template #default="{ row }">
                <span class="text-green-600">¥{{ formatAmount(row.tenant_amount) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="分账比例" width="100" align="center">
              <template #default="{ row }">
                {{ (row.platform_rate / 100).toFixed(2) }}%
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="160" />
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button
                  v-if="row.status === 'pending'"
                  type="primary" link size="small"
                  @click="handleProcess(row)"
                >执行结算</el-button>
                <el-button
                  v-if="row.status === 'failed'"
                  type="warning" link size="small"
                  @click="handleProcess(row)"
                >重试</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div class="p-4 flex justify-end">
            <el-pagination
              v-model:current-page="recordsPage"
              v-model:page-size="recordsPageSize"
              :total="recordsTotal"
              layout="total, sizes, prev, pager, next"
              @size-change="fetchRecords"
              @current-change="fetchRecords"
            />
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Check } from '@element-plus/icons-vue'
import settlementApi, { type SettlementStats } from '@/api/settlement'

const activeTab = ref('config')

const configSaving = ref(false)
const configForm = reactive({
  platform_rate: 1000,
  min_amount: 100
})

const stats = ref<SettlementStats>({
  total_amount: 0,
  platform_amount: 0,
  tenant_amount: 0,
  pending_count: 0,
  settled_count: 0,
  failed_count: 0
})

const recordsLoading = ref(false)
const records = ref<any[]>([])
const recordsPage = ref(1)
const recordsPageSize = ref(10)
const recordsTotal = ref(0)

function formatAmount(amount: number | undefined | null): string {
  if (amount == null || isNaN(amount)) return '0.00'
  return (amount / 100).toFixed(2)
}

function statusTagType(status: string): string {
  switch (status) {
    case 'settled': return 'success'
    case 'pending': return 'warning'
    case 'failed': return 'danger'
    default: return 'info'
  }
}

function statusLabel(status: string): string {
  switch (status) {
    case 'settled': return '已结算'
    case 'pending': return '待结算'
    case 'failed': return '失败'
    default: return status
  }
}

async function fetchConfig() {
  try {
    const res = await settlementApi.getSettlementConfig()
    if (res.data) {
      configForm.platform_rate = res.data.platform_rate || 1000
      configForm.min_amount = res.data.min_amount || 100
    }
  } catch (e) { console.error('fetchConfig failed:', e); }
}

async function saveConfig() {
  configSaving.value = true
  try {
    await settlementApi.updateSettlementConfig({
      platform_rate: configForm.platform_rate,
      min_amount: configForm.min_amount
    })
    ElMessage.success('分账配置更新成功')
  } catch (err: any) {
    ElMessage.error(err.message || '保存失败')
  } finally {
    configSaving.value = false
  }
}

async function fetchStats() {
  try {
    const res = await settlementApi.getSettlementStats()
    if (res.data) {
      stats.value = {
        total_amount: res.data.total_amount ?? 0,
        platform_amount: res.data.platform_amount ?? 0,
        tenant_amount: res.data.tenant_amount ?? 0,
        pending_count: res.data.pending_count ?? 0,
        settled_count: res.data.settled_count ?? 0,
        failed_count: res.data.failed_count ?? 0
      }
    }
  } catch (e) { console.error('fetchStats failed:', e); }
}

async function fetchRecords() {
  recordsLoading.value = true
  try {
    const res = await settlementApi.listSettlements({
      page: recordsPage.value,
      page_size: recordsPageSize.value
    })
    records.value = res.data?.list || []
    recordsTotal.value = res.data?.total || 0
  } catch {
    records.value = []
    recordsTotal.value = 0
  } finally {
    recordsLoading.value = false
  }
}

async function handleProcess(row: any) {
  try {
    await ElMessageBox.confirm('确定要执行结算吗？', '确认', { type: 'warning' })
    await settlementApi.processSettlement(row.settlement_no)
    ElMessage.success('结算处理成功')
    fetchRecords()
    fetchStats()
  } catch (e) { console.error('handleProcess failed:', e); }
}

onMounted(() => {
  fetchConfig()
  fetchStats()
  fetchRecords()
})
</script>

<style scoped>
.settlement-tabs :deep(.el-tabs__header) {
  padding: 0 20px;
  margin-bottom: 0;
}
.settlement-tabs :deep(.el-tabs__content) {
  padding: 16px 20px 20px;
}

:deep(.el-input__wrapper) {
  height: 32px;
}

:deep(.el-select .el-input__wrapper) {
  height: 32px;
}

:deep(.el-input-number .el-input__wrapper) {
  height: 32px;
}

:deep(.el-button) {
  height: 32px;
  padding: 8px 12px;
}
</style>
