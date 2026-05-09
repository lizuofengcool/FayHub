<template>
  <div class="dashboard-page">
    <div class="dashboard-header">
      <div class="flex items-center gap-4">
        <div class="header-icon">
          <el-icon class="text-2xl text-white"><DataAnalysis /></el-icon>
        </div>
        <div>
          <h2 class="text-xl font-bold text-slate-800">系统概览</h2>
          <p class="text-sm text-slate-400">实时监控系统运行状态和关键指标</p>
        </div>
      </div>
      <div class="header-meta">
        <span class="text-sm text-slate-400">{{ currentDate }}</span>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5 mb-6">
      <div class="stat-card group">
        <div class="flex items-center justify-between">
          <div>
            <p class="stat-label">租户总数</p>
            <p class="stat-value">{{ stats.tenant_count }}</p>
          </div>
          <div class="stat-icon bg-gradient-to-br from-indigo-400 to-indigo-600">
            <el-icon class="text-xl text-white"><OfficeBuilding /></el-icon>
          </div>
        </div>
        <div class="stat-footer">
          <el-icon class="text-emerald-500 mr-1 text-xs"><Top /></el-icon>
          <span class="text-emerald-500 font-medium text-xs">{{ formatGrowth(stats.tenant_growth) }}</span>
          <span class="text-slate-400 text-xs ml-1">本月新增</span>
        </div>
      </div>

      <div class="stat-card group">
        <div class="flex items-center justify-between">
          <div>
            <p class="stat-label">用户总数</p>
            <p class="stat-value">{{ stats.user_count }}</p>
          </div>
          <div class="stat-icon bg-gradient-to-br from-emerald-400 to-emerald-600">
            <el-icon class="text-xl text-white"><User /></el-icon>
          </div>
        </div>
        <div class="stat-footer">
          <el-icon class="text-emerald-500 mr-1 text-xs"><Top /></el-icon>
          <span class="text-emerald-500 font-medium text-xs">{{ formatGrowth(stats.user_growth) }}</span>
          <span class="text-slate-400 text-xs ml-1">本月新增</span>
        </div>
      </div>

      <div class="stat-card group">
        <div class="flex items-center justify-between">
          <div>
            <p class="stat-label">插件总数</p>
            <p class="stat-value">{{ stats.plugin_count }}</p>
          </div>
          <div class="stat-icon bg-gradient-to-br from-amber-400 to-amber-600">
            <el-icon class="text-xl text-white"><Grid /></el-icon>
          </div>
        </div>
        <div class="stat-footer">
          <div class="w-1.5 h-1.5 rounded-full bg-emerald-500 mr-1.5"></div>
          <span class="text-slate-400 text-xs">已激活 {{ stats.active_plugin_count }} 个</span>
        </div>
      </div>

      <div class="stat-card group">
        <div class="flex items-center justify-between">
          <div>
            <p class="stat-label">运行时间</p>
            <p class="stat-value text-2xl">{{ formatUptime(stats.uptime_seconds) }}</p>
          </div>
          <div class="stat-icon bg-gradient-to-br from-violet-400 to-violet-600">
            <el-icon class="text-xl text-white"><Monitor /></el-icon>
          </div>
        </div>
        <div class="stat-footer">
          <div class="w-1.5 h-1.5 rounded-full bg-emerald-500 mr-1.5"></div>
          <span class="text-slate-400 text-xs">内存 {{ (stats.memory_alloc_mb || 0).toFixed(1) }}MB · 协程 {{ stats.goroutine_count }}</span>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5 mb-6">
      <div class="stat-card group">
        <div class="flex items-center justify-between">
          <div>
            <p class="stat-label">今日订单</p>
            <p class="stat-value">{{ stats.order_today }}</p>
          </div>
          <div class="stat-icon bg-gradient-to-br from-cyan-400 to-cyan-600">
            <el-icon class="text-xl text-white"><Document /></el-icon>
          </div>
        </div>
        <div class="stat-footer">
          <span class="text-slate-400 text-xs">本月累计 {{ stats.order_month }} 笔</span>
        </div>
      </div>

      <div class="stat-card group">
        <div class="flex items-center justify-between">
          <div>
            <p class="stat-label">今日收入</p>
            <p class="stat-value">&yen;{{ (stats.payment_today || 0).toFixed(2) }}</p>
          </div>
          <div class="stat-icon bg-gradient-to-br from-teal-400 to-teal-600">
            <el-icon class="text-xl text-white"><Money /></el-icon>
          </div>
        </div>
        <div class="stat-footer">
          <span class="text-slate-400 text-xs">本月累计 &yen;{{ (stats.payment_month || 0).toFixed(2) }}</span>
        </div>
      </div>

      <div class="stat-card group">
        <div class="flex items-center justify-between">
          <div>
            <p class="stat-label">总请求数</p>
            <p class="stat-value">{{ stats.total_requests }}</p>
          </div>
          <div class="stat-icon bg-gradient-to-br from-rose-400 to-rose-600">
            <el-icon class="text-xl text-white"><TrendCharts /></el-icon>
          </div>
        </div>
        <div class="stat-footer">
          <div class="w-1.5 h-1.5 rounded-full bg-red-500 mr-1.5"></div>
          <span class="text-slate-400 text-xs">错误 {{ stats.error_requests }} 次</span>
        </div>
      </div>

      <div class="stat-card group">
        <div class="flex items-center justify-between">
          <div>
            <p class="stat-label">系统状态</p>
            <p class="stat-value text-2xl text-emerald-600">运行正常</p>
          </div>
          <div class="stat-icon bg-gradient-to-br from-blue-400 to-blue-600">
            <el-icon class="text-xl text-white"><CircleCheck /></el-icon>
          </div>
        </div>
        <div class="stat-footer">
          <span class="text-slate-400 text-xs">Go {{ stats.goroutine_count }} 协程</span>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-5 mb-6">
      <div class="glass-card rounded-xl p-6">
        <h3 class="text-base font-semibold text-slate-800 mb-4">请求趋势（近7天）</h3>
        <div style="height: 280px; display: flex; align-items: center; justify-content: center; color: #94a3b8;">图表区域（暂时隐藏）</div>
      </div>

      <div class="glass-card rounded-xl p-6">
        <h3 class="text-base font-semibold text-slate-800 mb-4">租户套餐分布</h3>
        <div style="height: 280px; display: flex; align-items: center; justify-content: center; color: #94a3b8;">图表区域（暂时隐藏）</div>
      </div>
    </div>

    <div class="glass-card rounded-xl p-6 mb-6">
      <h3 class="text-base font-semibold text-slate-800 mb-4">快速操作</h3>
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <button class="quick-action-btn group" @click="$router.push('/system/tenant')">
          <div class="quick-action-icon bg-indigo-50 group-hover:bg-indigo-100">
            <el-icon class="text-lg text-indigo-600"><Plus /></el-icon>
          </div>
          <span class="text-sm font-medium text-slate-600 group-hover:text-indigo-600">新增租户</span>
        </button>
        <button class="quick-action-btn group" @click="$router.push('/system/user')">
          <div class="quick-action-icon bg-emerald-50 group-hover:bg-emerald-100">
            <el-icon class="text-lg text-emerald-600"><UserFilled /></el-icon>
          </div>
          <span class="text-sm font-medium text-slate-600 group-hover:text-emerald-600">创建用户</span>
        </button>
        <button class="quick-action-btn group" @click="$router.push('/system/settings')">
          <div class="quick-action-icon bg-amber-50 group-hover:bg-amber-100">
            <el-icon class="text-lg text-amber-600"><Setting /></el-icon>
          </div>
          <span class="text-sm font-medium text-slate-600 group-hover:text-amber-600">系统设置</span>
        </button>
        <button class="quick-action-btn group" @click="$router.push('/plugins/installed')">
          <div class="quick-action-icon bg-violet-50 group-hover:bg-violet-100">
            <el-icon class="text-lg text-violet-600"><DataAnalysis /></el-icon>
          </div>
          <span class="text-sm font-medium text-slate-600 group-hover:text-violet-600">插件管理</span>
        </button>
      </div>
    </div>

    <div class="glass-card rounded-xl p-6">
      <h3 class="text-base font-semibold text-slate-800 mb-4">最近活动</h3>
      <div class="space-y-4">
        <div v-for="activity in stats.recent_activities" :key="activity.id" class="flex items-center p-3 rounded-lg hover:bg-slate-50/50 transition-colors">
          <div class="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center mr-3">
            <el-icon class="text-blue-600"><component :is="getActivityIcon(activity.icon)" /></el-icon>
          </div>
          <div class="flex-1">
            <p class="text-sm font-medium text-slate-700">{{ activity.title }}</p>
            <p class="text-xs text-slate-500">{{ activity.time }}</p>
          </div>
        </div>
        <div v-if="!stats.recent_activities || stats.recent_activities.length === 0" class="text-center text-slate-400 py-8">
          暂无活动记录
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import {
  Monitor, OfficeBuilding, User, Setting, UserFilled, Plus,
  DataAnalysis, TrendCharts, Top, Grid, Document, Money,
  CircleCheck, Download, Delete
} from '@element-plus/icons-vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart, BarChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, LegendComponent, GridComponent } from 'echarts/components'

use([CanvasRenderer, LineChart, PieChart, BarChart, TitleComponent, TooltipComponent, LegendComponent, GridComponent])
import request from '@/api/request'

const loading = ref(false)

const stats = ref({
  tenant_count: 0,
  tenant_growth: 0,
  user_count: 0,
  user_growth: 0,
  plugin_count: 0,
  active_plugin_count: 0,
  total_requests: 0,
  error_requests: 0,
  uptime_seconds: 0,
  memory_alloc_mb: 0,
  goroutine_count: 0,
  payment_today: 0,
  payment_month: 0,
  order_today: 0,
  order_month: 0,
  request_trend: [] as { date: string; count: number }[],
  tenant_distribution: [] as { name: string; value: number }[],
  recent_activities: [] as { id: number; title: string; time: string; icon: string }[]
})

const iconMap: Record<string, any> = {
  UserFilled, Setting, Plus, Download, Delete, Money, OfficeBuilding
}

const currentDate = computed(() => {
  const now = new Date()
  const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  return `${now.getFullYear()}年${now.getMonth() + 1}月${now.getDate()}日 ${weekdays[now.getDay()]}`
})

function getActivityIcon(iconName: string) {
  return iconMap[iconName] || Setting
}

async function fetchDashboardData() {
  loading.value = true
  try {
    const res = await request.get('/stats')
    if (res.data) {
      stats.value = { ...stats.value, ...res.data }
    }
  } catch (e) {
    console.error('dashboard stats failed:', e)
  }
  loading.value = false
}

function formatGrowth(val: number): string {
  if (val === 0) return '0%'
  if (val > 0) return `+${val.toFixed(1)}%`
  return `${val.toFixed(1)}%`
}

function formatUptime(seconds: number): string {
  if (!seconds || seconds <= 0) return '--'
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  if (days > 0) return `${days}天${hours}小时`
  const mins = Math.floor((seconds % 3600) / 60)
  return `${hours}小时${mins}分钟`
}

const pieColors = ['#6366f1', '#8b5cf6', '#a78bfa', '#c4b5fd', '#94a3b8', '#64748b', '#475569']

const requestTrendOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
  xAxis: {
    type: 'category',
    boundaryGap: false,
    data: (stats.value.request_trend || []).map((p: any) => p.date),
    axisLine: { lineStyle: { color: '#cbd5e1' } },
    axisLabel: { color: '#94a3b8' }
  },
  yAxis: {
    type: 'value',
    splitLine: { lineStyle: { color: '#f1f5f9' } },
    axisLabel: { color: '#94a3b8' }
  },
  series: [{
    data: (stats.value.request_trend || []).map((p: any) => p.count),
    type: 'line',
    smooth: true,
    areaStyle: {
      color: {
        type: 'linear',
        x: 0, y: 0, x2: 0, y2: 1,
        colorStops: [
          { offset: 0, color: 'rgba(99, 102, 241, 0.3)' },
          { offset: 1, color: 'rgba(99, 102, 241, 0.02)' }
        ]
      }
    },
    lineStyle: { color: '#6366f1', width: 2 },
    itemStyle: { color: '#6366f1' }
  }]
}))

const tenantPieOption = computed(() => ({
  tooltip: { trigger: 'item' },
  legend: { bottom: '0%', textStyle: { color: '#64748b' } },
  series: [{
    name: '租户分布',
    type: 'pie',
    radius: ['45%', '75%'],
    center: ['50%', '45%'],
    avoidLabelOverlap: false,
    itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
    label: { show: false },
    emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } },
    data: (stats.value.tenant_distribution || []).map((item: any, idx: number) => ({
      ...item,
      itemStyle: { color: pieColors[idx % pieColors.length] }
    }))
  }]
}))

onMounted(() => {
  fetchDashboardData()
})
</script>

<style scoped>
.dashboard-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  padding: 18px 24px;
  background: var(--card-bg, #fff);
  border: 1px solid var(--border-color, #e2e8f0);
  border-radius: 12px;
  box-shadow: var(--card-shadow, 0 1px 2px rgba(0,0,0,0.04));
}

.header-icon {
  width: 42px;
  height: 42px;
  border-radius: 10px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.header-meta {
  flex-shrink: 0;
}

.stat-card {
  background: var(--card-bg, #fff);
  border: 1px solid var(--border-color, #e2e8f0);
  border-radius: 12px;
  padding: 18px 20px;
  transition: all 0.2s ease;
  box-shadow: var(--card-shadow, 0 1px 2px rgba(0,0,0,0.04));
}

.stat-card:hover {
  border-color: var(--primary, #4f46e5);
  box-shadow: 0 4px 20px rgba(79, 70, 229, 0.08);
  transform: translateY(-1px);
}

.stat-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-muted, #94a3b8);
  margin-bottom: 4px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-title, #0f172a);
  line-height: 1.2;
}

.stat-icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.stat-footer {
  margin-top: 14px;
  display: flex;
  align-items: center;
}

.glass-card {
  background: var(--card-bg, #fff);
  border: 1px solid var(--border-color, #e2e8f0);
  border-radius: 12px;
  box-shadow: var(--card-shadow, 0 1px 2px rgba(0,0,0,0.04));
}

.quick-action-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px 12px;
  border-radius: 10px;
  border: 1px solid var(--border-color, #e2e8f0);
  background: var(--card-bg, #fff);
  cursor: pointer;
  transition: all 0.2s ease;
}

.quick-action-btn:hover {
  border-color: var(--primary, #4f46e5);
  box-shadow: 0 2px 12px rgba(79, 70, 229, 0.06);
  transform: translateY(-1px);
}

.quick-action-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s ease;
}
</style>
