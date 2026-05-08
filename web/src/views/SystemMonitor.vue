<template>
  <div class="system-monitor-page">
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 pb-3 flex items-center justify-between">
        <div>
          <h2 class="text-lg font-bold text-slate-800">系统监控</h2>
          <p class="text-slate-400 text-xs mt-0.5">实时监控系统运行状态、性能指标和告警信息</p>
        </div>
        <div class="flex items-center gap-3">
          <el-tag :type="autoRefresh ? 'success' : 'info'" class="cursor-pointer" @click="autoRefresh = !autoRefresh">
            {{ autoRefresh ? '自动刷新中' : '已暂停' }}
          </el-tag>
          <el-button @click="fetchStats" :loading="loading">
            <el-icon class="mr-1"><Refresh /></el-icon> 刷新
          </el-button>
        </div>
      </div>

      <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4 px-4 mb-4">
      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-4">
        <div class="text-xs text-slate-500 mb-1">运行时间</div>
        <div class="text-xl font-bold text-slate-800">{{ formatUptime(stats.uptime_seconds) }}</div>
        <div class="text-xs text-green-500 mt-1 flex items-center gap-1">
          <span class="w-2 h-2 rounded-full bg-green-500 inline-block"></span> 正常运行
        </div>
      </div>
      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-4">
        <div class="text-xs text-slate-500 mb-1">总请求数</div>
        <div class="text-xl font-bold text-slate-800">{{ formatNumber(stats.total_requests) }}</div>
        <div class="text-xs text-slate-400 mt-1">历史累计</div>
      </div>
      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-4">
        <div class="text-xs text-slate-500 mb-1">活跃请求</div>
        <div class="text-xl font-bold text-blue-600">{{ stats.active_requests }}</div>
        <div class="text-xs text-slate-400 mt-1">当前处理中</div>
      </div>
      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-4" :class="{ 'border-red-200 bg-red-50': errorRate > alertThresholds.error_rate }">
        <div class="text-xs text-slate-500 mb-1">错误率</div>
        <div class="text-xl font-bold" :class="errorRate > alertThresholds.error_rate ? 'text-red-600' : 'text-slate-800'">
          {{ errorRate.toFixed(1) }}%
        </div>
        <div class="text-xs mt-1" :class="errorRate > alertThresholds.error_rate ? 'text-red-500' : 'text-slate-400'">
          {{ errorRate > alertThresholds.error_rate ? '⚠ 超过阈值' : '正常范围' }}
        </div>
      </div>
      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-4">
        <div class="text-xs text-slate-500 mb-1">Goroutines</div>
        <div class="text-xl font-bold text-slate-800">{{ stats.goroutine_count }}</div>
        <div class="text-xs text-slate-400 mt-1">并发协程数</div>
      </div>
      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-4" :class="{ 'border-amber-200 bg-amber-50': stats.memory_alloc_mb > alertThresholds.memory_mb }">
        <div class="text-xs text-slate-500 mb-1">内存使用</div>
        <div class="text-xl font-bold" :class="stats.memory_alloc_mb > alertThresholds.memory_mb ? 'text-amber-600' : 'text-slate-800'">
          {{ (stats.memory_alloc_mb || 0).toFixed(1) }} MB
        </div>
        <div class="text-xs text-slate-400 mt-1">系统: {{ (stats.memory_sys_mb || 0).toFixed(1) }} MB</div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-6">
        <h3 class="text-lg font-semibold text-slate-800 mb-4 flex items-center gap-2">
          <el-icon class="text-blue-500"><TrendCharts /></el-icon> 请求趋势
        </h3>
        <div ref="requestChartRef" style="height: 280px;"></div>
      </div>
      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-6">
        <h3 class="text-lg font-semibold text-slate-800 mb-4 flex items-center gap-2">
          <el-icon class="text-purple-500"><Cpu /></el-icon> 资源使用
        </h3>
        <div ref="resourceChartRef" style="height: 280px;"></div>
      </div>
    </div>

    <div class="bg-white rounded-2xl shadow-sm border border-slate-100 overflow-hidden">
      <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
        <h3 class="text-lg font-semibold text-slate-800 flex items-center gap-2">
          <el-icon class="text-green-500"><List /></el-icon> API 端点性能
        </h3>
        <el-input v-model="apiSearch" placeholder="搜索端点..." size="small" clearable class="!w-56" />
      </div>
      <el-table :data="filteredApiMetrics" stripe class="w-full" empty-text="暂无API调用记录" max-height="400">
        <el-table-column prop="method" label="方法" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="methodTagType(row.method)" size="small">{{ row.method }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路径" min-width="280">
          <template #default="{ row }">
            <span class="font-mono text-sm">{{ row.path }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="total_count" label="调用次数" width="110" align="center" sortable />
        <el-table-column label="错误率" width="100" align="center">
          <template #default="{ row }">
            <span :class="row.total_count > 0 && (row.error_count / row.total_count * 100) > 10 ? 'text-red-500 font-semibold' : 'text-slate-600'">
              {{ row.total_count > 0 ? (row.error_count / row.total_count * 100).toFixed(1) + '%' : '-' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="平均耗时" width="110" align="center" sortable>
          <template #default="{ row }">
            <span :class="row.avg_dur_ms > alertThresholds.response_time_ms ? 'text-amber-500 font-semibold' : 'text-slate-600'">
              {{ row.avg_dur_ms }} ms
            </span>
          </template>
        </el-table-column>
        <el-table-column label="最小/最大" width="130" align="center">
          <template #default="{ row }">
            <span class="text-xs text-slate-500">{{ row.min_dur_ms }} / {{ row.max_dur_ms }} ms</span>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-6">
      <h3 class="text-lg font-semibold text-slate-800 mb-4 flex items-center gap-2">
        <el-icon class="text-red-500"><Bell /></el-icon> 告警配置
      </h3>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div>
          <label class="text-sm font-medium text-slate-700 block mb-2">错误率阈值 (%)</label>
          <el-input-number v-model="alertThresholds.error_rate" :min="1" :max="100" :step="1" class="!w-full" />
          <p class="text-xs text-slate-400 mt-1">当错误率超过此值时触发告警</p>
        </div>
        <div>
          <label class="text-sm font-medium text-slate-700 block mb-2">响应时间阈值 (ms)</label>
          <el-input-number v-model="alertThresholds.response_time_ms" :min="100" :max="10000" :step="100" class="!w-full" />
          <p class="text-xs text-slate-400 mt-1">当平均响应时间超过此值时触发告警</p>
        </div>
        <div>
          <label class="text-sm font-medium text-slate-700 block mb-2">内存阈值 (MB)</label>
          <el-input-number v-model="alertThresholds.memory_mb" :min="100" :max="8192" :step="100" class="!w-full" />
          <p class="text-xs text-slate-400 mt-1">当内存使用超过此值时触发告警</p>
        </div>
      </div>
      <div class="mt-4 flex items-center gap-3">
        <el-button type="primary" @click="saveAlertConfig" :loading="savingConfig">保存配置</el-button>
        <el-button @click="testAlert">测试告警通知</el-button>
      </div>
    </div>

    <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-6">
      <h3 class="text-lg font-semibold text-slate-800 mb-4 flex items-center gap-2">
        <el-icon class="text-slate-500"><InfoFilled /></el-icon> GC 统计
      </h3>
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
        <div>
          <span class="text-slate-500">GC 次数</span>
          <p class="text-lg font-semibold text-slate-800">{{ stats.gc_count || 0 }}</p>
        </div>
        <div>
          <span class="text-slate-500">GC 总暂停</span>
          <p class="text-lg font-semibold text-slate-800">{{ (stats.gc_pause_total_ms || 0).toFixed(2) }} ms</p>
        </div>
        <div>
          <span class="text-slate-500">堆内存分配</span>
          <p class="text-lg font-semibold text-slate-800">{{ (stats.memory_alloc_mb || 0).toFixed(1) }} MB</p>
        </div>
        <div>
          <span class="text-slate-500">系统内存</span>
          <p class="text-lg font-semibold text-slate-800">{{ (stats.memory_sys_mb || 0).toFixed(1) }} MB</p>
        </div>
      </div>
    </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { ElMessage, ElNotification } from 'element-plus'
import { Refresh, TrendCharts, Cpu, List, Bell, InfoFilled } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import request from '@/api/request'

interface ApiMetric {
  method: string
  path: string
  total_count: number
  error_count: number
  avg_dur_ms: number
  min_dur_ms: number
  max_dur_ms: number
}

interface SystemStats {
  uptime_seconds: number
  total_requests: number
  active_requests: number
  error_requests: number
  goroutine_count: number
  memory_alloc_mb: number
  memory_sys_mb: number
  gc_pause_total_ms: number
  gc_count: number
  api_metrics: ApiMetric[]
}

const stats = reactive<SystemStats>({
  uptime_seconds: 0,
  total_requests: 0,
  active_requests: 0,
  error_requests: 0,
  goroutine_count: 0,
  memory_alloc_mb: 0,
  memory_sys_mb: 0,
  gc_pause_total_ms: 0,
  gc_count: 0,
  api_metrics: []
})

const loading = ref(false)
const savingConfig = ref(false)
const autoRefresh = ref(true)
const apiSearch = ref('')
const requestChartRef = ref<HTMLElement>()
const resourceChartRef = ref<HTMLElement>()
let requestChart: echarts.ECharts | null = null
let resourceChart: echarts.ECharts | null = null
let refreshTimer: ReturnType<typeof setInterval> | null = null

const requestHistory = ref<{ time: string; total: number; errors: number }[]>([])
const resourceHistory = ref<{ time: string; memory: number; goroutines: number }[]>([])

const alertThresholds = reactive({
  error_rate: 10,
  response_time_ms: 2000,
  memory_mb: 1024
})

const errorRate = computed(() => {
  if (stats.total_requests === 0) return 0
  return (stats.error_requests / stats.total_requests) * 100
})

const filteredApiMetrics = computed(() => {
  if (!apiSearch.value) return stats.api_metrics || []
  const q = apiSearch.value.toLowerCase()
  return (stats.api_metrics || []).filter(m =>
    m.path.toLowerCase().includes(q) || m.method.toLowerCase().includes(q)
  )
})

onMounted(() => {
  fetchStats()
  loadAlertConfig()
  if (autoRefresh.value) startAutoRefresh()
  nextTick(() => {
    initCharts()
  })
})

onBeforeUnmount(() => {
  stopAutoRefresh()
  requestChart?.dispose()
  resourceChart?.dispose()
})

watch(autoRefresh, (val) => {
  if (val) startAutoRefresh()
  else stopAutoRefresh()
})

function startAutoRefresh() {
  stopAutoRefresh()
  refreshTimer = setInterval(fetchStats, 5000)
}

function stopAutoRefresh() {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

async function fetchStats() {
  try {
    const res = await request.get('/stats')
    if (res?.data) {
      const data = res.data
      Object.assign(stats, data)

      const now = new Date().toLocaleTimeString()
      requestHistory.value.push({
        time: now,
        total: data.total_requests || 0,
        errors: data.error_requests || 0
      })
      resourceHistory.value.push({
        time: now,
        memory: data.memory_alloc_mb || 0,
        goroutines: data.goroutine_count || 0
      })

      if (requestHistory.value.length > 60) requestHistory.value.shift()
      if (resourceHistory.value.length > 60) resourceHistory.value.shift()

      updateCharts()
      checkAlerts(data)
    }
  } catch (e) {
    console.error('刷新监控数据失败', e)
  } finally {
    loading.value = false
  }
}

function checkAlerts(data: any) {
  const errRate = data.total_requests > 0 ? (data.error_requests / data.total_requests) * 100 : 0
  if (errRate > alertThresholds.error_rate) {
    ElNotification({
      title: '告警：错误率过高',
      message: `当前错误率 ${errRate.toFixed(1)}%，超过阈值 ${alertThresholds.error_rate}%`,
      type: 'error',
      duration: 10000
    })
  }
  if (data.memory_alloc_mb > alertThresholds.memory_mb) {
    ElNotification({
      title: '告警：内存使用过高',
      message: `当前内存 ${data.memory_alloc_mb.toFixed(1)} MB，超过阈值 ${alertThresholds.memory_mb} MB`,
      type: 'warning',
      duration: 10000
    })
  }
}

function initCharts() {
  if (requestChartRef.value) {
    requestChart = echarts.init(requestChartRef.value)
  }
  if (resourceChartRef.value) {
    resourceChart = echarts.init(resourceChartRef.value)
  }
  updateCharts()
}

function updateCharts() {
  if (requestChart) {
    requestChart.setOption({
      tooltip: { trigger: 'axis' },
      legend: { data: ['总请求', '错误请求'], bottom: 0 },
      grid: { left: 50, right: 20, top: 10, bottom: 30 },
      xAxis: {
        type: 'category',
        data: requestHistory.value.map(h => h.time),
        axisLabel: { fontSize: 10, rotate: 45 }
      },
      yAxis: { type: 'value', minInterval: 1 },
      series: [
        {
          name: '总请求',
          type: 'line',
          data: requestHistory.value.map(h => h.total),
          smooth: true,
          lineStyle: { color: '#3b82f6' },
          itemStyle: { color: '#3b82f6' }
        },
        {
          name: '错误请求',
          type: 'line',
          data: requestHistory.value.map(h => h.errors),
          smooth: true,
          lineStyle: { color: '#ef4444' },
          itemStyle: { color: '#ef4444' }
        }
      ]
    })
  }

  if (resourceChart) {
    resourceChart.setOption({
      tooltip: { trigger: 'axis' },
      legend: { data: ['内存 (MB)', 'Goroutines'], bottom: 0 },
      grid: { left: 50, right: 50, top: 10, bottom: 30 },
      xAxis: {
        type: 'category',
        data: resourceHistory.value.map(h => h.time),
        axisLabel: { fontSize: 10, rotate: 45 }
      },
      yAxis: [
        { type: 'value', name: 'MB' },
        { type: 'value', name: '数量' }
      ],
      series: [
        {
          name: '内存 (MB)',
          type: 'line',
          data: resourceHistory.value.map(h => h.memory),
          smooth: true,
          lineStyle: { color: '#8b5cf6' },
          itemStyle: { color: '#8b5cf6' }
        },
        {
          name: 'Goroutines',
          type: 'line',
          yAxisIndex: 1,
          data: resourceHistory.value.map(h => h.goroutines),
          smooth: true,
          lineStyle: { color: '#10b981' },
          itemStyle: { color: '#10b981' }
        }
      ]
    })
  }
}

function loadAlertConfig() {
  try {
    const saved = localStorage.getItem('fayhub_alert_config')
    if (saved) {
      const config = JSON.parse(saved)
      Object.assign(alertThresholds, config)
    }
  } catch (e) { console.error('loadAlertConfig failed:', e); }
}

async function saveAlertConfig() {
  savingConfig.value = true
  try {
    localStorage.setItem('fayhub_alert_config', JSON.stringify(alertThresholds))
    ElMessage.success('告警配置已保存')
  } catch {
    ElMessage.error('保存失败')
  } finally {
    savingConfig.value = false
  }
}

function testAlert() {
  ElNotification({
    title: '测试告警',
    message: '这是一条测试告警通知，如果您能看到此消息，说明告警功能正常。',
    type: 'info',
    duration: 5000
  })
}

function formatUptime(seconds: number): string {
  if (!seconds || seconds < 0) return '-'
  const d = Math.floor(seconds / 86400)
  const h = Math.floor((seconds % 86400) / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (d > 0) return `${d}d ${h}h`
  if (h > 0) return `${h}h ${m}m`
  return `${m}m`
}

function formatNumber(n: number): string {
  if (!n) return '0'
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'K'
  return n.toString()
}

function methodTagType(method: string): 'success' | 'warning' | 'danger' | 'info' | '' {
  switch (method) {
    case 'GET': return 'success'
    case 'POST': return ''
    case 'PUT': return 'warning'
    case 'DELETE': return 'danger'
    default: return 'info'
  }
}
</script>

<style scoped>
:deep(.el-input__wrapper) {
  height: 32px;
}

:deep(.el-select .el-input__wrapper) {
  height: 32px;
}

:deep(.el-button) {
  height: 32px;
  padding: 8px 12px;
}

:deep(.el-input-number .el-input__wrapper) {
  height: 32px;
}
</style>
