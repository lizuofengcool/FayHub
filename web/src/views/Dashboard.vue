<template>
  <div class="flex h-screen w-screen overflow-hidden text-slate-800 relative z-10">
    
    <!-- 左侧菜单栏 (毛玻璃效果) -->
    <aside class="glass-sidebar w-64 flex flex-col z-20">
      <!-- Logo 区域 -->
      <div class="h-16 flex items-center px-6 border-b border-slate-100/50">
        <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center text-white font-bold text-lg shadow-md shadow-indigo-500/30 mr-3">
          F
        </div>
        <h1 class="text-xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-slate-800 to-slate-600 tracking-tight">FayHub</h1>
      </div>

      <!-- 导航菜单 -->
      <div class="flex-1 py-6 px-3 space-y-1 overflow-y-auto">
        <div class="text-xs font-bold text-slate-400 mb-2 px-3 uppercase tracking-wider">系统管理</div>
        <div class="menu-item active flex items-center px-4 py-2.5 rounded-xl cursor-pointer font-medium text-sm">
          <el-icon class="mr-3 text-lg"><Monitor /></el-icon> 仪表盘
        </div>
        <div class="menu-item flex items-center px-4 py-2.5 rounded-xl cursor-pointer text-slate-600 font-medium text-sm" @click="$router.push('/tenants')">
          <el-icon class="mr-3 text-lg"><OfficeBuilding /></el-icon> 租户管理
        </div>
        <div class="menu-item flex items-center px-4 py-2.5 rounded-xl cursor-pointer text-slate-600 font-medium text-sm" @click="$router.push('/users')">
          <el-icon class="mr-3 text-lg"><User /></el-icon> 用户管理
        </div>
        <div class="menu-item flex items-center px-4 py-2.5 rounded-xl cursor-pointer text-slate-600 font-medium text-sm">
          <el-icon class="mr-3 text-lg"><Lock /></el-icon> 角色权限
        </div>
        <div class="menu-item flex items-center px-4 py-2.5 rounded-xl cursor-pointer text-slate-600 font-medium text-sm">
          <el-icon class="mr-3 text-lg"><Setting /></el-icon> 系统设置
        </div>
      </div>

      <!-- 用户信息 -->
      <div class="p-4 border-t border-slate-100/50">
        <div class="flex items-center">
          <img :src="userInfo.avatar" :alt="userInfo.nickname" class="w-8 h-8 rounded-full border-2 border-slate-100">
          <div class="ml-3 flex-1">
            <p class="text-sm font-semibold text-slate-700">{{ userInfo.nickname }}</p>
            <p class="text-xs text-slate-500">{{ userInfo.role }}</p>
          </div>
          <el-button text @click="handleLogout">
            <el-icon><SwitchButton /></el-icon>
          </el-button>
        </div>
      </div>
    </aside>

    <!-- 右侧主内容区 -->
    <main class="flex-1 flex flex-col min-w-0">
      
      <!-- 顶部导航栏 (毛玻璃效果) -->
      <header class="glass-header h-16 flex items-center justify-between px-8 z-10 sticky top-0">
        <!-- 面包屑 -->
        <div class="flex items-center text-sm font-medium text-slate-500">
          <span class="hover:text-indigo-600 cursor-pointer transition-colors">首页</span>
          <el-icon class="mx-2 text-slate-400"><ArrowRight /></el-icon>
          <span class="text-slate-800 font-semibold">系统概览</span>
        </div>

        <!-- 右侧工具栏 -->
        <div class="flex items-center space-x-5">
          <el-button text>
            <el-icon><Search /></el-icon>
          </el-button>
          <el-button text>
            <el-icon><Bell /></el-icon>
          </el-button>
          <div class="h-5 w-px bg-slate-200"></div>
          <div class="flex items-center cursor-pointer group">
            <img :src="userInfo.avatar" alt="Avatar" class="w-8 h-8 rounded-full border-2 border-slate-100 group-hover:border-indigo-200 transition-all">
            <span class="ml-2 text-sm font-semibold text-slate-700 group-hover:text-indigo-600 transition-colors">{{ userInfo.nickname }}</span>
            <el-icon class="ml-1 text-slate-400 group-hover:text-indigo-600"><ArrowDown /></el-icon>
          </div>
        </div>
      </header>

      <!-- 核心页面内容 -->
      <div class="flex-1 overflow-y-auto p-8">
        
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
          <div class="glass-card rounded-2xl p-6 fade-in-up" style="animation-delay: 0.1s">
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
          <div class="glass-card rounded-2xl p-6 fade-in-up" style="animation-delay: 0.2s">
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
          <div class="glass-card rounded-2xl p-6 fade-in-up" style="animation-delay: 0.3s">
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
          <div class="glass-card rounded-2xl p-6 fade-in-up" style="animation-delay: 0.4s">
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
        <div class="glass-card rounded-2xl p-6 mb-8 fade-in-up" style="animation-delay: 0.5s">
          <h3 class="text-lg font-semibold text-slate-800 mb-4">快速操作</h3>
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <button class="flex flex-col items-center p-4 rounded-xl border border-slate-200 hover:border-blue-300 hover:bg-blue-50 transition-all group" @click="$router.push('/tenants')">
              <el-icon class="text-2xl text-blue-600 mb-2 group-hover:scale-110 transition-transform"><Plus /></el-icon>
              <span class="text-sm font-medium text-slate-700">新增租户</span>
            </button>
            <button class="flex flex-col items-center p-4 rounded-xl border border-slate-200 hover:border-green-300 hover:bg-green-50 transition-all group" @click="$router.push('/users')">
              <el-icon class="text-2xl text-green-600 mb-2 group-hover:scale-110 transition-transform"><UserFilled /></el-icon>
              <span class="text-sm font-medium text-slate-700">创建用户</span>
            </button>
            <button class="flex flex-col items-center p-4 rounded-xl border border-slate-200 hover:border-orange-300 hover:bg-orange-50 transition-all group">
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
        <div class="glass-card rounded-2xl p-6 fade-in-up" style="animation-delay: 0.6s">
          <h3 class="text-lg font-semibold text-slate-800 mb-4">最近活动</h3>
          <div class="space-y-4">
            <div v-for="activity in recentActivities" :key="activity.id" class="flex items-center p-3 rounded-lg hover:bg-slate-50/50 transition-colors">
              <div class="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center mr-3">
                <el-icon class="text-blue-600">{{ activity.icon }}</el-icon>
              </div>
              <div class="flex-1">
                <p class="text-sm font-medium text-slate-700">{{ activity.title }}</p>
                <p class="text-xs text-slate-500">{{ activity.time }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

const router = useRouter()

// 用户信息
const userInfo = ref({
  id: 1,
  username: 'admin',
  nickname: '超级管理员',
  role: '系统管理员',
  avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=admin&backgroundColor=e2e8f0'
})

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

// 退出登录
const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    ElMessage.success('已安全退出')
    router.push('/')
  } catch {
    // 用户取消操作
  }
}

onMounted(() => {
  // 从本地存储加载用户信息
  const storedUser = localStorage.getItem('user')
  if (storedUser) {
    userInfo.value = { ...userInfo.value, ...JSON.parse(storedUser) }
  }
})
</script>

<style scoped>
.glass-sidebar {
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-right: 1px solid rgba(255, 255, 255, 0.6);
  box-shadow: 
    2px 0 8px rgba(0, 0, 0, 0.02),
    inset 0 0 0 1px rgba(255, 255, 255, 0.5);
}

.glass-header {
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.6);
  box-shadow: 
    0 2px 8px rgba(0, 0, 0, 0.03),
    inset 0 0 0 1px rgba(255, 255, 255, 0.5);
}

.menu-item {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border-radius: 10px;
  margin: 2px 8px;
}

.menu-item:hover:not(.active) {
  background-color: rgba(241, 245, 249, 0.8);
  color: #334155;
  transform: translateX(4px);
}

.menu-item.active {
  background: linear-gradient(135deg, #4f46e5, #3b82f6);
  color: white;
  box-shadow: 0 4px 12px rgba(79, 70, 229, 0.25);
}
</style>