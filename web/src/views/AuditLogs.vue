<template>
  <div class="audit-page">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold text-slate-800">审计日志</h2>
        <p class="text-slate-500 mt-1 text-sm">查看系统关键操作审计轨迹</p>
      </div>
      <el-button type="danger" @click="openCleanupDialog">
        <el-icon class="mr-1"><Delete /></el-icon>
        清理历史日志
      </el-button>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
      <div class="bg-white rounded-xl border border-slate-100 p-4 shadow-sm">
        <p class="text-sm text-slate-500">总记录数</p>
        <p class="text-2xl font-bold text-slate-800 mt-1">{{ stats.total }}</p>
      </div>
      <div class="bg-white rounded-xl border border-slate-100 p-4 shadow-sm">
        <p class="text-sm text-slate-500">今日操作</p>
        <p class="text-2xl font-bold text-blue-600 mt-1">{{ stats.today }}</p>
      </div>
      <div class="bg-white rounded-xl border border-slate-100 p-4 shadow-sm">
        <p class="text-sm text-slate-500">成功率</p>
        <p class="text-2xl font-bold text-green-600 mt-1">{{ (stats.success_rate * 100).toFixed(1) }}%</p>
      </div>
    </div>

    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 border-b border-slate-100 flex gap-3 flex-wrap">
        <el-select v-model="filters.action" placeholder="操作类型" clearable style="width: 140px">
          <el-option label="登录" value="login" />
          <el-option label="登出" value="logout" />
          <el-option label="创建" value="create" />
          <el-option label="更新" value="update" />
          <el-option label="删除" value="delete" />
          <el-option label="启用" value="enable" />
          <el-option label="禁用" value="disable" />
          <el-option label="升级" value="upgrade" />
          <el-option label="安装" value="install" />
          <el-option label="卸载" value="uninstall" />
          <el-option label="导出" value="export" />
          <el-option label="导入" value="import" />
        </el-select>
        <el-select v-model="filters.success" placeholder="结果" clearable style="width: 120px">
          <el-option label="成功" :value="true" />
          <el-option label="失败" :value="false" />
        </el-select>
        <el-input v-model="filters.path" placeholder="请求路径" clearable style="width: 200px" />
        <el-input v-model="filters.ip" placeholder="IP地址" clearable style="width: 150px" />
        <el-date-picker
          v-model="dateRange"
          type="datetimerange"
          range-separator="至"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          format="YYYY-MM-DD HH:mm"
          value-format="YYYY-MM-DDTHH:mm:ssZ"
          style="width: 360px"
        />
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>

      <el-table v-loading="loading" :data="logs" stripe class="w-full">
        <el-table-column prop="username" label="用户" width="100">
          <template #default="{ row }">
            <span>{{ row.username || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="action" label="操作" width="100">
          <template #default="{ row }">
            <el-tag :type="actionTagType(row.action)" size="small">{{ actionLabel(row.action) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="resource" label="资源" width="120" show-overflow-tooltip />
        <el-table-column prop="method" label="方法" width="80" align="center">
          <template #default="{ row }">
            <span class="text-xs font-mono px-1.5 py-0.5 rounded" :class="methodClass(row.method)">{{ row.method }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路径" min-width="200" show-overflow-tooltip />
        <el-table-column prop="status_code" label="状态码" width="90" align="center">
          <template #default="{ row }">
            <span :class="row.status_code < 400 ? 'text-green-600' : 'text-red-500'">{{ row.status_code }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="success" label="结果" width="70" align="center">
          <template #default="{ row }">
            <el-tag :type="row.success ? 'success' : 'danger'" size="small">{{ row.success ? '成功' : '失败' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="duration" label="耗时" width="80" align="center">
          <template #default="{ row }">
            <span class="text-xs">{{ row.duration }}ms</span>
          </template>
        </el-table-column>
        <el-table-column prop="ip" label="IP" width="120" />
        <el-table-column prop="created_at" label="时间" width="160" />
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="viewDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="p-4 flex justify-end">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @current-change="fetchLogs"
          @size-change="fetchLogs"
        />
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="审计日志详情" width="640px">
      <div v-if="currentLog">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="用户">{{ currentLog.username || '-' }}</el-descriptions-item>
          <el-descriptions-item label="操作">
            <el-tag :type="actionTagType(currentLog.action)" size="small">{{ actionLabel(currentLog.action) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="资源">{{ currentLog.resource }}</el-descriptions-item>
          <el-descriptions-item label="资源ID">{{ currentLog.resource_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="方法">
            <span class="font-mono">{{ currentLog.method }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="路径" :span="2">{{ currentLog.path }}</el-descriptions-item>
          <el-descriptions-item label="状态码">{{ currentLog.status_code }}</el-descriptions-item>
          <el-descriptions-item label="结果">
            <el-tag :type="currentLog.success ? 'success' : 'danger'" size="small">{{ currentLog.success ? '成功' : '失败' }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="耗时">{{ currentLog.duration }}ms</el-descriptions-item>
          <el-descriptions-item label="IP">{{ currentLog.ip }}</el-descriptions-item>
          <el-descriptions-item label="请求ID" :span="2">{{ currentLog.request_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="User-Agent" :span="2">{{ currentLog.user_agent || '-' }}</el-descriptions-item>
          <el-descriptions-item v-if="currentLog.error_msg" label="错误信息" :span="2">
            <span class="text-red-500">{{ currentLog.error_msg }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="时间" :span="2">{{ currentLog.created_at }}</el-descriptions-item>
        </el-descriptions>
        <div v-if="currentLog.detail && Object.keys(currentLog.detail).length > 0" class="mt-4">
          <p class="text-sm font-medium text-slate-700 mb-2">请求详情</p>
          <pre class="bg-slate-50 rounded-lg p-3 text-xs overflow-auto max-h-48">{{ JSON.stringify(currentLog.detail, null, 2) }}</pre>
        </div>
      </div>
    </el-dialog>

    <el-dialog v-model="cleanupVisible" title="清理历史日志" width="400px">
      <el-form label-width="100px">
        <el-form-item label="清理范围">
          <el-radio-group v-model="cleanupType">
            <el-radio value="30d">30天前</el-radio>
            <el-radio value="90d">90天前</el-radio>
            <el-radio value="180d">180天前</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="cleanupVisible = false">取消</el-button>
        <el-button type="danger" :loading="cleanupLoading" @click="handleCleanup">确定清理</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete } from '@element-plus/icons-vue'
import auditApi, { type AuditLog, type AuditStats } from '@/api/audit'

const loading = ref(false)
const logs = ref<AuditLog[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const dateRange = ref<string[]>([])

const stats = ref<AuditStats>({ total: 0, today: 0, success_rate: 0, top_actions: [], top_users: [] })

const filters = reactive({
  action: '',
  success: undefined as boolean | undefined,
  path: '',
  ip: ''
})

const detailVisible = ref(false)
const currentLog = ref<AuditLog | null>(null)

const cleanupVisible = ref(false)
const cleanupType = ref('90d')
const cleanupLoading = ref(false)

async function fetchLogs() {
  loading.value = true
  try {
    const params: any = {
      page: page.value,
      page_size: pageSize.value,
      action: filters.action || undefined,
      success: filters.success,
      path: filters.path || undefined,
      ip: filters.ip || undefined
    }
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_time = dateRange.value[0]
      params.end_time = dateRange.value[1]
    }
    const res = await auditApi.listLogs(params)
    logs.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (err: any) {
    ElMessage.error(err.message || '获取审计日志失败')
  } finally {
    loading.value = false
  }
}

async function fetchStats() {
  try {
    const res = await auditApi.getStats()
    stats.value = res.data || { total: 0, today: 0, success_rate: 0, top_actions: [], top_users: [] }
  } catch {}
}

function handleSearch() {
  page.value = 1
  fetchLogs()
}

function resetFilters() {
  filters.action = ''
  filters.success = undefined
  filters.path = ''
  filters.ip = ''
  dateRange.value = []
  page.value = 1
  fetchLogs()
}

function viewDetail(row: AuditLog) {
  currentLog.value = row
  detailVisible.value = true
}

function openCleanupDialog() {
  cleanupVisible.value = true
}

async function handleCleanup() {
  cleanupLoading.value = true
  try {
    const days = parseInt(cleanupType.value)
    const beforeTime = new Date(Date.now() - days * 24 * 60 * 60 * 1000).toISOString()
    await auditApi.cleanup({ before_time: beforeTime })
    ElMessage.success('清理完成')
    cleanupVisible.value = false
    fetchLogs()
    fetchStats()
  } catch (err: any) {
    ElMessage.error(err.message || '清理失败')
  } finally {
    cleanupLoading.value = false
  }
}

function actionTagType(action: string) {
  switch (action) {
    case 'login': case 'logout': return ''
    case 'create': case 'install': case 'enable': return 'success'
    case 'delete': case 'uninstall': case 'disable': return 'danger'
    case 'update': case 'upgrade': return 'primary'
    case 'rollback': return 'warning'
    default: return 'info'
  }
}

function actionLabel(action: string) {
  const map: Record<string, string> = {
    login: '登录', logout: '登出', create: '创建', update: '更新',
    delete: '删除', enable: '启用', disable: '禁用', upgrade: '升级',
    rollback: '回滚', install: '安装', uninstall: '卸载',
    export: '导出', import: '导入', config_change: '配置变更', permission_change: '权限变更'
  }
  return map[action] || action
}

function methodClass(method: string) {
  switch (method) {
    case 'GET': return 'bg-green-100 text-green-700'
    case 'POST': return 'bg-blue-100 text-blue-700'
    case 'PUT': return 'bg-yellow-100 text-yellow-700'
    case 'DELETE': return 'bg-red-100 text-red-700'
    case 'PATCH': return 'bg-purple-100 text-purple-700'
    default: return 'bg-slate-100 text-slate-700'
  }
}

onMounted(() => {
  fetchLogs()
  fetchStats()
})
</script>

<style scoped>
</style>
