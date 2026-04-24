<template>
  <div class="flex h-screen w-screen overflow-hidden text-slate-800 relative z-10">
    
    <!-- 左侧菜单栏 -->
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
        <div class="menu-item flex items-center px-4 py-2.5 rounded-xl cursor-pointer text-slate-600 font-medium text-sm" @click="$router.push('/dashboard')">
          <el-icon class="mr-3 text-lg"><Monitor /></el-icon> 仪表盘
        </div>
        <div class="menu-item flex items-center px-4 py-2.5 rounded-xl cursor-pointer text-slate-600 font-medium text-sm" @click="$router.push('/tenants')">
          <el-icon class="mr-3 text-lg"><OfficeBuilding /></el-icon> 租户管理
        </div>
        <div class="menu-item flex items-center px-4 py-2.5 rounded-xl cursor-pointer text-slate-600 font-medium text-sm" @click="$router.push('/users')">
          <el-icon class="mr-3 text-lg"><User /></el-icon> 用户管理
        </div>
      </div>
    </aside>

    <!-- 右侧主内容区 -->
    <main class="flex-1 flex flex-col min-w-0">
      
      <!-- 顶部导航栏 -->
      <header class="glass-header h-16 flex items-center justify-between px-8 z-10 sticky top-0">
        <!-- 面包屑 -->
        <div class="flex items-center text-sm font-medium text-slate-500">
          <span class="hover:text-indigo-600 cursor-pointer transition-colors" @click="$router.push('/dashboard')">首页</span>
          <el-icon class="mx-2 text-slate-400"><ArrowRight /></el-icon>
          <span class="text-slate-800 font-semibold">页面未找到</span>
        </div>

        <!-- 右侧工具栏 -->
        <div class="flex items-center space-x-5">
          <div class="flex items-center cursor-pointer group">
            <img :src="userInfo.avatar" alt="Avatar" class="w-8 h-8 rounded-full border-2 border-slate-100 group-hover:border-indigo-200 transition-all">
            <span class="ml-2 text-sm font-semibold text-slate-700 group-hover:text-indigo-600 transition-colors">{{ userInfo.nickname }}</span>
          </div>
        </div>
      </header>

      <!-- 核心页面内容 -->
      <div class="flex-1 overflow-y-auto p-8">
        <div class="flex flex-col items-center justify-center h-full">
          <!-- 404 错误图标 -->
          <div class="w-32 h-32 rounded-full bg-gradient-to-br from-blue-100 to-indigo-100 flex items-center justify-center mb-8">
            <el-icon class="text-6xl text-blue-500"><Warning /></el-icon>
          </div>
          
          <!-- 错误信息 -->
          <h1 class="text-6xl font-bold text-slate-800 mb-4">404</h1>
          <h2 class="text-2xl font-semibold text-slate-600 mb-2">页面未找到</h2>
          <p class="text-slate-500 mb-8 text-center max-w-md">
            抱歉，您访问的页面不存在或已被移动。
            <br>
            请检查URL是否正确，或返回首页继续浏览。
          </p>
          
          <!-- 操作按钮 -->
          <div class="flex gap-4">
            <el-button type="primary" @click="$router.push('/dashboard')">
              <el-icon><HomeFilled /></el-icon>
              返回首页
            </el-button>
            <el-button @click="$router.go(-1)">
              <el-icon><ArrowLeft /></el-icon>
              返回上页
            </el-button>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

// 用户信息
const userInfo = ref({
  nickname: '超级管理员',
  avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=admin&backgroundColor=e2e8f0'
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
</style>