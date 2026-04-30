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

      <!-- 活跃度 -->
      <div class="glass-card rounded-2xl p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-slate-500 mb-1">今日活跃</p>
            <p class="text-3xl font-bold text-slate-800">{{ dashboardData.activeUsers }}</p>
          </div>
          <div class="w-12 h-12 rounded-xl bg-orange-50 flex items-center justify-center">
            <el-icon class="text-2xl text-orange-600"><TrendCharts /></el-icon>
          </div>
        </div>
        <div class="mt-4 flex items-center text-sm text-slate-500">
          <el-icon class="text-green-500 mr-1"><Top /></el-icon>
          <span class="text-green-500 font-medium">+{{ dashboardData.activeGrowth }}%</span>
          <span class="ml-1">较昨日</span>
        </div>
      </div>

      <!-- 系统状态 -->
      <div class="glass-card rounded-2xl p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-slate-500 mb-1">系统状态</p>
            <p class="text-3xl font-bold text-slate-800">{{ dashboardData.systemStatus }}</p>
          </div>
          <div class="w-12 h-12 rounded-xl bg-purple-50 flex items-center justify-center">
            <el-icon class="text-2xl text-purple-600"><Monitor /></el-icon>
          </div>
        </div>
        <div class="mt-4 flex items-center text-sm text-slate-500">
          <div class="w-2 h-2 rounded-full bg-green-500 mr-2"></div>
          <span>运行稳定</span>
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
import { ref } from 'vue'
import { Monitor, OfficeBuilding, User, Setting, UserFilled, Plus, DataAnalysis, TrendCharts, Top } from '@element-plus/icons-vue'

const activityIconMap: Record<string, any> = {
  'OfficeBuilding': OfficeBuilding,
  'UserFilled': UserFilled,
  'Setting': Setting
}

// 仪表盘数据
const dashboardData = ref({
  tenantCount: 128,
  userCount: 2847,
  activeUsers: 892,
  systemStatus: '正常',
  tenantGrowth: 12,
  userGrowth: 8,
  activeGrowth: 5
})

// 最近活动
const recentActivities = ref([
  {
    id: 1,
    icon: 'OfficeBuilding',
    title: '新增租户 "星辉科技"',
    time: '5分钟前'
  },
  {
    id: 2,
    icon: 'UserFilled',
    title: '用户 "张三" 登录系统',
    time: '15分钟前'
  },
  {
    id: 3,
    icon: 'Setting',
    title: '系统配置已更新',
    time: '1小时前'
  }
])
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
