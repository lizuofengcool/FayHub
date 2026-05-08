<template>
  <div class="plugin-resource-page">
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 pb-3 flex items-center justify-between">
        <div>
          <h2 class="text-lg font-bold text-slate-800">插件资源监控</h2>
          <p class="text-slate-400 text-xs mt-0.5">实时监控插件调用性能、资源使用和告警信息</p>
        </div>
        <div class="flex items-center gap-3">
          <el-tag :type="autoRefresh ? 'success' : 'info'" class="cursor-pointer" @click="autoRefresh = !autoRefresh">
            {{ autoRefresh ? '自动刷新中' : '已暂停' }}
          </el-tag>
          <el-button @click="fetchAll" :loading="loading">
            <el-icon class="mr-1"><Refresh /></el-icon> 刷新
          </el-button>
        </div>
      </div>

      <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4 px-4 mb-4">
      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-4">
        <div class="text-xs text-slate-500 mb-1">总调用次数</div>
        <div class="text-xl font-bold text-slate-800">{{ formatNumber(summary.total_calls) }}</div>
        <div class="text-xs text-slate-400 mt-1">历史累计</div>
      </div>
      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-4">
        <div class="text-xs text-slate-500 mb-1">总错误数</div>
        <div class="text-xl font-bold text-red-600">{{ formatNumber(summary.total_errors) }}</div>
        <div class="text-xs text-slate-400 mt-1">历史累计</div>
      </div>
      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-4" :class="{ 'border-red-200 bg-red-50': summary.error_rate > 10 }">
        <div class="text-xs text-slate-500 mb-1">错误率</div>
        <div class="text-xl font-bold" :class="summary.error_rate > 10 ? 'text-red-600' : 'text-slate-800'">
          {{ (summary.error_rate || 0).toFixed(1) }}%
        </div>
        <div class="text-xs mt-1" :class="summary.error_rate > 10 ? 'text-red-500' : 'text-slate-400'">
          {{ summary.error_rate > 10 ? '⚠ 异常' : '正常' }}
        </div>
      </div>
      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-4">
        <div class="text-xs text-slate-500 mb-1">监控插件数</div>
        <div class="text-xl font-bold text-blue-600">{{ summary.total_plugins }}</div>
        <div class="text-xs text-slate-400 mt-1">活跃插件</div>
      </div>
      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-4">
        <div class="text-xs text-slate-500 mb-1">运行时插件</div>
        <div class="text-xl font-bold text-purple-600">{{ runtimeStats.length }}</div>
        <div class="text-xs text-slate-400 mt-1">内存中活跃</div>
      </div>
      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-4">
        <div class="text-xs text-slate-500 mb-1">告警数量</div>
        <div class="text-xl font-bold text-amber-600">{{ alerts.length }}</div>
        <div class="text-xs text-slate-400 mt-1">最近20条</div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-6">
        <h3 class="text-lg font-semibold text-slate-800 mb-4 flex items-center gap-2">
          <el-icon class="text-blue-500"><TrendCharts /></el-icon> 调用次数分布
        </h3>
        <div ref="callChartRef" style="height: 280px;"></div>
      </div>
      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-6">
        <h3 class="text-lg font-semibold text-slate-800 mb-4 flex items-center gap-2">
          <el-icon class="text-purple-500"><Cpu /></el-icon> 平均耗时对比
        </h3>
        <div ref="durationChartRef" style="height: 280px;"></div>
      </div>
    </div>

    <div class="bg-white rounded-2xl shadow-sm border border-slate-100 overflow-hidden">
      <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
        <h3 class="text-lg font-semibold text-slate-800 flex items-center gap-2">
          <el-icon class="text-green-500"><List /></el-icon> 插件运行时统计
        </h3>
        <el-input v-model="pluginSearch" placeholder="搜索插件..." size="small" clearable class="!w-56" />
      </div>
      <el-table :data="filteredRuntimeStats" stripe class="w-full" empty-text="暂无插件调用记录" max-height="500">
        <el-table-column prop="plugin_id" label="插件ID" min-width="180">
          <template #default="{ row }">
            <span class="font-mono text-sm">{{ row.plugin_id }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="call_count" label="调用次数" width="110" align="center" sortable />
        <el-table-column prop="error_count" label="错误数" width="90" align="center" sortable>
          <template #default="{ row }">
            <span :class="row.error_count > 0 ? 'text-red-500 font-semibold' : 'text-slate-600'">
              {{ row.error_count }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="错误率" width="90" align="center">
          <template #default="{ row }">
            <span :class="getErrorRate(row) > 10 ? 'text-red-500 font-semibold' : 'text-slate-600'">
              {{ getErrorRate(row).toFixed(1) }}%
            </span>
          </template>
        </el-table-column>
        <el-table-column label="平均耗时" width="110" align="center" sortable prop="avg_duration_ms">
          <template #default="{ row }">
            <span :class="row.avg_duration_ms > 3000 ? 'text-amber-500 font-semibold' : 'text-slate-600'">
              {{ row.avg_duration_ms }} ms
            </span>
          </template>
        </el-table-column>
        <el-table-column label="最大耗时" width="110" align="center" sortable prop="max_duration_ms">
          <template #default="{ row }">
            <span :class="row.max_duration_ms > 30000 ? 'text-red-500 font-semibold' : 'text-slate-600'">
              {{ row.max_duration_ms }} ms
            </span>
          </template>
        </el-table-column>
        <el-table-column label="内存(KB)" width="100" align="center" sortable prop="memory_usage_kb" />
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'warning'" size="small">
              {{ row.status === 'active' ? '活跃' : row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="最后调用" width="170" align="center">
          <template #default="{ row }">
            <span class="text-xs text-slate-500">{{ formatTime(row.last_call_at) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="danger" size="small" text @click="handleReset(row.plugin_id)">
              重置
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="bg-white rounded-2xl shadow-sm border border-slate-100 overflow-hidden">
      <div class="px-6 py-4 border-b border-slate-100">
        <h3 class="text-lg font-semibold text-slate-800 flex items-center gap-2">
          <el-icon class="text-red-500"><Bell /></el-icon> 最近告警
        </h3>
      </div>
      <el-table :data="alerts" stripe class="w-full" empty-text="暂无告警记录" max-height="350">
        <el-table-column prop="plugin_id" label="插件ID" width="200" />
        <el-table-column prop="event_type" label="类型" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="alertTagType(row.event_type)" size="small">{{ row.event_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="event_data" label="内容" min-width="300" />
        <el-table-column prop="created_at" label="时间" width="180" align="center">
          <template #default="{ row }">
            <span class="text-xs text-slate-500">{{ formatTime(row.created_at) }}</span>
          </template>
        </el-table-column>
      </el-table>
    </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, TrendCharts, Cpu, List, Bell } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import pluginMonitorApi, {
  type PluginRuntimeStats,
  type PluginAlert,
  type StatsSummary
} from '@/api/pluginMonitor'

const loading = ref(false)
const autoRefresh = ref(true)
const pluginSearch = ref('')
const callChartRef = ref<HTMLElement>()
const durationChartRef = ref<HTMLElement>()
let callChart: echarts.ECharts | null = null
let durationChart: echarts.ECharts | null = null
let refreshTimer: ReturnType<typeof setInterval> | null = null

const runtimeStats = ref<PluginRuntimeStats[]>([])
const alerts = ref<PluginAlert[]>([])
const summary = reactive<StatsSummary>({
  total_calls: 0,
  total_errors: 0,
  total_plugins: 0,
  error_rate: 0,
  updated_at: ''
})

const filteredRuntimeStats = computed(() => {
  if (!pluginSearch.value) return runtimeStats.value
  const q = pluginSearch.value.toLowerCase()
  return runtimeStats.value.filter(s => s.plugin_id.toLowerCase().includes(q))
})

onMounted(() => {
  fetchAll()
  if (autoRefresh.value) startAutoRefresh()
  nextTick(() => initCharts())
})

onBeforeUnmount(() => {
  stopAutoRefresh()
  callChart?.dispose()
  durationChart?.dispose()
})

watch(autoRefresh, (val) => {
  if (val) startAutoRefresh()
  else stopAutoRefresh()
})

function startAutoRefresh() {
  stopAutoRefresh()
  refreshTimer = setInterval(fetchAll, 5000)
}

function stopAutoRefresh() {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

async function fetchAll() {
  loading.value = true
  try {
    const [runtimeRes, dbRes, alertsRes] = await Promise.all([
      pluginMonitorApi.getRuntimeStats(),
      pluginMonitorApi.getDBStats(),
      pluginMonitorApi.getAlerts()
    ])

    if (runtimeRes?.data) {
      runtimeStats.value = Array.isArray(runtimeRes.data) ? runtimeRes.data : []
    }
    if (dbRes?.data) {
      Object.assign(summary, dbRes.data)
    }
    if (alertsRes?.data) {
      alerts.value = Array.isArray(alertsRes.data) ? alertsRes.data : []
    }

    updateCharts()
  } catch {
    // silent
  } finally {
    loading.value = false
  }
}

function initCharts() {
  if (callChartRef.value) {
    callChart = echarts.init(callChartRef.value)
  }
  if (durationChartRef.value) {
    durationChart = echarts.init(durationChartRef.value)
  }
  updateCharts()
}

function updateCharts() {
  if (!callChart || !durationChart) return

  const plugins = runtimeStats.value.slice(0, 15)
  const names = plugins.map(p => p.plugin_id.length > 20 ? p.plugin_id.slice(0, 20) + '...' : p.plugin_id)
  const calls = plugins.map(p => p.call_count)
  const durations = plugins.map(p => p.avg_duration_ms)

  callChart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: {
      type: 'category',
      data: names,
      axisLabel: { rotate: 45, fontSize: 11 }
    },
    yAxis: { type: 'value', name: '调用次数' },
    series: [{
      name: '调用次数',
      type: 'bar',
      data: calls,
      itemStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: '#3b82f6' },
          { offset: 1, color: '#93c5fd' }
        ])
      },
      borderRadius: [6, 6, 0, 0]
    }]
  })

  durationChart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: {
      type: 'category',
      data: names,
      axisLabel: { rotate: 45, fontSize: 11 }
    },
    yAxis: { type: 'value', name: '耗时(ms)' },
    series: [{
      name: '平均耗时',
      type: 'bar',
      data: durations,
      itemStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: '#8b5cf6' },
          { offset: 1, color: '#c4b5fd' }
        ])
      },
      borderRadius: [6, 6, 0, 0]
    }]
  })
}

async function handleReset(pluginId: string) {
  try {
    await pluginMonitorApi.resetStats(pluginId)
    ElMessage.success(`插件 ${pluginId} 统计已重置`)
    fetchAll()
  } catch {
    ElMessage.error('重置失败')
  }
}

function getErrorRate(row: PluginRuntimeStats): number {
  if (row.call_count === 0) return 0
  return (row.error_count / row.call_count) * 100
}

function formatNumber(n: number): string {
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'K'
  return String(n)
}

function formatTime(t: string): string {
  if (!t) return '-'
  try {
    return new Date(t).toLocaleString('zh-CN')
  } catch {
    return t
  }
}

function alertTagType(type: string): string {
  const map: Record<string, string> = {
    resource_alert: 'danger',
    high_latency: 'warning',
    high_memory: 'danger',
    high_cpu: 'warning',
    high_error_rate: 'danger',
    consecutive_errors: 'danger'
  }
  return map[type] || 'info'
}
</script>
