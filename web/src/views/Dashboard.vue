<template>
  <div class="dashboard-page">
    <!-- 页面标题 -->
    <div class="mb-6 flex justify-between items-end">
      <div>
        <h2 class="text-2xl font-bold text-slate-800 tracking-tight">系统概览</h2>
        <p class="text-sm text-slate-500 mt-1">实时监控系统运行状态和关键指标</p>
      </div>
    </div>

    <!-- 统计卡片网格 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
      <!-- 租户统计 -->
      <div class="glass-card rounded-2xl p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-slate-500 mb-1">租户总数</p>
            <p class="text-3xl font-bold text-slate-800">{{ dashboardData.tenantCount }}</p>
          </div>
          <div class="w-12 h-12 rounded-xl bg-blue-50 flex items-center justify-center">
            <el-icon class="text-2xl text-blue-600"><OfficeBuilding /></el-icon>
          </div>
        </div>
        <div class="mt-4 flex items-center text-sm text-slate-500">
          <el-icon class="text-green-500 mr-1"><Top /></el-icon>
          <span class="text-green-500 font-medium">+{{ dashboardData.tenantGrowth }}%</span>
          <span class="ml-1">本月新增</span>
        </div>
      </div>

      <!-- 用户统计 -->
      <div class="glass-card rounded-2xl p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-slate-500 mb-1">用户总数</p>
            <p class="text-3xl font-bold text-slate-800">{{ dashboardData.userCount }}</p>
          </div>
          <div class="w-12 h-12 rounded-xl bg-green-50 flex items-center justify-center">
            <el-icon class="text-2xl text-green-600"><User /></el-icon>
          </div>
        </div>
        <div class="mt-4 flex items-center text-sm text-slate-500">
          <el-icon class="text-green-500 mr-1"><Top /></el-icon>
          <span class="text-green-500 font-medium">+{{ dashboardData.userGrowth }}%</span>
          <span class="ml-1">本月新增</span>
        </div>
      </div>

      <!-- 总请求数 -->
      <div class="glass-card rounded-2xl p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-slate-500 mb-1">总请求数</p>
            <p class="text-3xl font-bold text-slate-800">{{ dashboardData.totalRequests }}</p>
          </div>
          <div class="w-12 h-12 rounded-xl bg-orange-50 flex items-center justify-center">
            <el-icon class="text-2xl text-orange-600"><TrendCharts /></el-icon>
          </div>
        </div>
        <div class="mt-4 flex items-center text-sm text-slate-500">
          <div class="w-2 h-2 rounded-full bg-green-500 mr-2"></div>
          <span>今日操作 {{ dashboardData.activeUsers }}</span>
        </div>
      </div>

      <!-- 系统状态 -->
      <div class="glass-card rounded-2xl p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-slate-500 mb-1">运行时间</p>
            <p class="text-xl font-bold text-slate-800">{{ formatUptime(dashboardData.uptimeSeconds) }}</p>
          </div>
          <div class="w-12 h-12 rounded-xl bg-purple-50 flex items-center justify-center">
            <el-icon class="text-2xl text-purple-600"><Monitor /></el-icon>
          </div>
        </div>
        <div class="mt-4 flex items-center text-sm text-slate-500">
          <div class="w-2 h-2 rounded-full bg-green-500 mr-2"></div>
          <span>内存 {{ dashboardData.memoryAllocMb.toFixed(1) }}MB · 协程 {{ dashboardData.goroutineCount }}</span>
        </div>
      </div>
    </div>

    <!-- 快速操作区域 -->
    <div class="glass-card rounded-2xl p-6 mb-8">
      <h3 class="text-lg font-semibold text-slate-800 mb-4">快速操作</h3>
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <button class="flex flex-col items-center p-4 rounded-xl border border-slate-200 hover:border-blue-300 hover:bg-blue-50 transition-all group" @click="$router.push('/system/tenant')">
          <el-icon class="text-2xl text-blue-600 mb-2 group-hover:scale-110 transition-transform"><Plus /></el-icon>
          <span class="text-sm font-medium text-slate-700">新增租户</span>
        </button>
        <button class="flex flex-col items-center p-4 rounded-xl border border-slate-200 hover:border-green-300 hover:bg-green-50 transition-all group" @click="$router.push('/system/user')">
          <el-icon class="text-2xl text-green-600 mb-2 group-hover:scale-110 transition-transform"><UserFilled /></el-icon>
          <span class="text-sm font-medium text-slate-700">创建用户</span>
        </button>
        <button class="flex flex-col items-center p-4 rounded-xl border border-slate-200 hover:border-orange-300 hover:bg-orange-50 transition-all group" @click="$router.push('/system/settings')">
          <el-icon class="text-2xl text-orange-600 mb-2 group-hover:scale-110 transition-transform"><Setting /></el-icon>
          <span class="text-sm font-medium text-slate-700">系统设置</span>
        </button>
        <button class="flex flex-col items-center p-4 rounded-xl border border-slate-200 hover:border-purple-300 hover:bg-purple-50 transition-all group">
          <el-icon class="text-2xl text-purple-600 mb-2 group-hover:scale-110 transition-transform"><DataAnalysis /></el-icon>
          <span class="text-sm font-medium text-slate-700">查看报表</span>
        </button>
      </div>
    </div>

    <!-- 最近活动 -->
    <div class="glass-card rounded-2xl p-6">
      <h3 class="text-lg font-semibold text-slate-800 mb-4">最近活动</h3>
      <div class="space-y-4">
        <div v-for="activity in recentActivities" :key="activity.id" class="flex items-center p-3 rounded-lg hover:bg-slate-50/50 transition-colors">
          <div class="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center mr-3">
            <el-icon class="text-blue-600"><component :is="activityIconMap[activity.icon]" /></el-icon>
          </div>
          <div class="flex-1">
            <p class="text-sm font-medium text-slate-700">{{ activity.title }}</p>
            <p class="text-xs text-slate-500">{{ activity.time }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Monitor, OfficeBuilding, User, Setting, UserFilled, Plus, DataAnalysis, TrendCharts, Top } from '@element-plus/icons-vue'
import request from '@/api/request'

const activityIconMap: Record<string, any> = {
  'OfficeBuilding': OfficeBuilding,
  'UserFilled': UserFilled,
  'Setting': Setting
}

const loading = ref(false)

const dashboardData = ref({
  tenantCount: 0,
  userCount: 0,
  activeUsers: 0,
  systemStatus: '加载中',
  tenantGrowth: 0,
  userGrowth: 0,
  activeGrowth: 0,
  totalRequests: 0,
  uptimeSeconds: 0,
  memoryAllocMb: 0,
  goroutineCount: 0
})

const recentActivities = ref<any[]>([])

async function fetchDashboardData() {
  loading.value = true
  try {
    const statsRes = await request.get('/stats')
    if (statsRes.data) {
      const d = statsRes.data
      dashboardData.value.totalRequests = d.total_requests || 0
      dashboardData.value.uptimeSeconds = d.uptime_seconds || 0
      dashboardData.value.memoryAllocMb = d.memory_alloc_mb || 0
      dashboardData.value.goroutineCount = d.goroutine_count || 0
      dashboardData.value.systemStatus = '正常'
    }
  } catch {}

  try {
    const auditRes = await request.get('/audit/stats')
    if (auditRes.data) {
      dashboardData.value.activeUsers = auditRes.data.today || 0
    }
  } catch {}

  try {
    const tenantRes = await request.get('/tenants', { params: { page: 1, page_size: 1 } })
    if (tenantRes.data) {
      dashboardData.value.tenantCount = tenantRes.data.total || 0
    }
  } catch {}

  try {
    const userRes = await request.get('/users', { params: { page: 1, page_size: 1 } })
    if (userRes.data) {
      dashboardData.value.userCount = userRes.data.total || 0
    }
  } catch {}

  try {
    const auditRes = await request.get('/audit/logs', { params: { page: 1, page_size: 5 } })
    if (auditRes.data?.list) {
      recentActivities.value = auditRes.data.list.map((log: any) => ({
        id: log.id,
        icon: 'Setting',
        title: `${log.username || '系统'} ${actionLabel(log.action)} ${log.resource || ''}`,
        time: log.created_at || ''
      }))
    }
  } catch {}

  loading.value = false
}

function actionLabel(action: string): string {
  const map: Record<string, string> = {
    login: '登录了系统',
    logout: '退出了系统',
    create: '创建了',
    update: '更新了',
    delete: '删除了',
    enable: '启用了',
    disable: '禁用了'
  }
  return map[action] || action
}

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  if (days > 0) return `${days}天${hours}小时`
  const mins = Math.floor((seconds % 3600) / 60)
  return `${hours}小时${mins}分钟`
}

onMounted(() => {
  fetchDashboardData()
})
</script>

<style scoped>
.glass-card {
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.6);
  box-shadow: 
    0 2px 8px rgba(0, 0, 0, 0.04),
    inset 0 0 0 1px rgba(255, 255, 255, 0.5);
}
</style>
