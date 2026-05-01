<template>
  <div class="flex h-screen w-screen overflow-hidden text-slate-800 relative z-10">
    
    <!-- 移动端遮罩 -->
    <div 
      v-if="sidebarOpen && isMobile" 
      class="fixed inset-0 bg-black/40 z-20 transition-opacity"
      @click="sidebarOpen = false"
    ></div>

    <!-- 左侧菜单栏 (毛玻璃效果) -->
    <aside 
      class="glass-sidebar w-64 flex flex-col z-20 transition-transform duration-300"
      :class="{ 
        '-translate-x-full': isMobile && !sidebarOpen,
        'fixed inset-y-0 left-0': isMobile,
        'translate-x-0': isMobile && sidebarOpen
      }"
    >
      <!-- Logo 区域 -->
      <div class="h-16 flex items-center px-6 border-b border-slate-100/50">
        <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center text-white font-bold text-lg shadow-md shadow-indigo-500/30 mr-3">
          F
        </div>
        <h1 class="text-xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-slate-800 to-slate-600 tracking-tight">FayHub</h1>
      </div>

      <!-- 导航菜单 -->
      <div class="flex-1 py-6 px-3 space-y-1 overflow-y-auto">
        <template v-for="menu in visibleMenuItems" :key="menu.id">
          <div v-if="menu.children && menu.children.length > 0" class="mb-3">
            <div class="text-xs font-bold text-slate-400 mb-2 px-3 uppercase tracking-wider">
              {{ menu.title }}
            </div>
            <template v-for="child in menu.children" :key="child.id">
              <div 
                class="menu-item flex items-center px-4 py-2.5 rounded-xl cursor-pointer font-medium text-sm"
                :class="{ active: isMenuActive(child) }"
                @click="navigateTo(child)"
              >
                <el-icon class="mr-3 text-lg" v-if="child.icon && iconMap[child.icon]">
                  <component :is="iconMap[child.icon]" />
                </el-icon>
                {{ child.title }}
                <el-icon v-if="child.layout === 'fullscreen'" class="ml-auto text-xs text-slate-400"><FullScreen /></el-icon>
              </div>
            </template>
          </div>
          <div 
            v-else-if="!menu.children || menu.children.length === 0"
            class="menu-item flex items-center px-4 py-2.5 rounded-xl cursor-pointer font-medium text-sm"
            :class="{ active: isMenuActive(menu) }"
            @click="navigateTo(menu)"
          >
            <el-icon class="mr-3 text-lg" v-if="menu.icon && iconMap[menu.icon]">
              <component :is="iconMap[menu.icon]" />
            </el-icon>
            {{ menu.title }}
          </div>
        </template>
      </div>

      <!-- 用户信息 -->
      <div class="p-4 border-t border-slate-100/50">
        <div class="flex items-center">
          <img :src="displayAvatar" :alt="userInfo.username" class="w-8 h-8 rounded-full border-2 border-slate-100" @error="handleAvatarError">
          <div class="ml-3 flex-1">
            <p class="text-sm font-semibold text-slate-700">{{ userInfo.username }}</p>
            <p class="text-xs text-slate-500">{{ displayRole }}</p>
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
      <header class="glass-header h-16 flex items-center justify-between px-4 md:px-8 z-10 sticky top-0">
        <!-- 面包屑 -->
        <div class="flex items-center text-sm font-medium text-slate-500">
          <el-button text class="md:hidden mr-2 !p-1" @click="sidebarOpen = !sidebarOpen">
            <el-icon class="text-xl"><Menu /></el-icon>
          </el-button>
          <span class="hover:text-indigo-600 cursor-pointer transition-colors hidden sm:inline" @click="$router.push('/dashboard')">首页</span>
          <el-icon class="mx-2 text-slate-400 hidden sm:inline"><ArrowRight /></el-icon>
          <span class="text-slate-800 font-semibold">{{ currentPageTitle }}</span>
        </div>

        <!-- 右侧工具栏 -->
        <div class="flex items-center space-x-2 md:space-x-5">
          <el-button text class="hidden sm:inline-flex">
            <el-icon><Search /></el-icon>
          </el-button>
          <el-button text class="hidden sm:inline-flex">
            <el-icon><Bell /></el-icon>
          </el-button>
          <div class="h-5 w-px bg-slate-200 hidden sm:block"></div>
          <el-dropdown trigger="hover" @command="handleUserCommand">
            <div class="flex items-center cursor-pointer group">
              <img :src="displayAvatar" alt="Avatar" class="w-8 h-8 rounded-full border-2 border-slate-100 group-hover:border-indigo-200 transition-all" @error="handleAvatarError">
              <span class="ml-2 text-sm font-semibold text-slate-700 group-hover:text-indigo-600 transition-colors">{{ userInfo.username }}</span>
              <el-icon class="ml-1 text-slate-400 group-hover:text-indigo-600"><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">
                  <el-icon><User /></el-icon>个人中心
                </el-dropdown-item>
                <el-dropdown-item command="settings">
                  <el-icon><Setting /></el-icon>系统设置
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <el-icon><SwitchButton /></el-icon>退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <!-- 核心页面内容 -->
      <div class="flex-1 overflow-y-auto p-4 md:p-8">
        <router-view />
      </div>
    </main>

    <PluginDevTools v-if="showDevTools" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Monitor, OfficeBuilding, User, Lock, Setting, SwitchButton, UserFilled, 
  Box, Menu, Connection, ArrowRight, Search, Bell, ArrowDown,
  Shop, DataAnalysis, Grid, Key, List, Management, Tickets, CreditCard, Wallet,
  FullScreen, Folder, Upload, Document
} from '@element-plus/icons-vue'
import menuApi, { type Menu as MenuType } from '@/api/menu'
import PluginDevTools from '@/plugin/PluginDevTools.vue'

const showDevTools = import.meta.env.DEV

const router = useRouter()
const route = useRoute()

const isMobile = ref(false)
const sidebarOpen = ref(false)

function checkMobile() {
  isMobile.value = window.innerWidth < 768
  if (!isMobile.value) {
    sidebarOpen.value = false
  }
}

const registeredPluginRoutes = ref<Set<string>>(new Set())

const userInfo = ref({
  id: 1,
  username: 'admin',
  role: '系统管理员',
  avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=admin&backgroundColor=e2e8f0'
})

const menuItems = ref<MenuType[]>([])

// 过滤菜单："插件应用"无子菜单时隐藏（空容器不显示）
const visibleMenuItems = computed(() => {
  return menuItems.value.filter(menu => {
    if (menu.path === '/plugin-apps' && (!menu.children || menu.children.length === 0)) {
      return false
    }
    return true
  })
})

const iconMap: Record<string, any> = {
  'Monitor': Monitor,
  'OfficeBuilding': OfficeBuilding,
  'User': User,
  'Lock': Lock,
  'Setting': Setting,
  'SwitchButton': SwitchButton,
  'UserFilled': UserFilled,
  'Box': Box,
  'Menu': Menu,
  'Connection': Connection,
  'Search': Search,
  'Bell': Bell,
  'ArrowDown': ArrowDown,
  'ArrowRight': ArrowRight,
  'Shop': Shop,
  'DataAnalysis': DataAnalysis,
  'Grid': Grid,
  'Key': Key,
  'List': List,
  'Management': Management,
  'Tickets': Tickets,
  'CreditCard': CreditCard,
  'Wallet': Wallet,
  'dashboard': Monitor,
  'setting': Setting,
  'user': User,
  'role': Lock,
  'menu': Menu,
  'api': Connection,
  'tenant': OfficeBuilding,
  'Folder': Folder,
  'folder': Folder,
  'Upload': Upload,
  'Document': Document
}

const roleMap: Record<string, string> = {
  'super_admin': '超级管理员',
  'tenant_admin': '租户管理员',
  'user': '普通用户'
}

function navigateTo(menu: MenuType) {
  if (menu.layout === 'fullscreen' && menu.path) {
    const pluginPath = menu.path.replace(/^\/plugin-apps\//, '')
    router.push(`/plugin-fullscreen/${pluginPath}`)
  } else {
    router.push(menu.path)
  }
  if (isMobile.value) {
    sidebarOpen.value = false
  }
}

function isMenuActive(menu: MenuType): boolean {
  if (menu.layout === 'fullscreen') {
    const pluginPath = menu.path.replace(/^\/plugin-apps\//, '')
    return route.path === `/plugin-fullscreen/${pluginPath}`
  }
  return route.path === menu.path
}

const displayAvatar = computed(() => {
  return userInfo.value.avatar || 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 36 36"><circle cx="18" cy="18" r="18" fill="#e2e8f0"/><text x="18" y="23" text-anchor="middle" fill="#64748b" font-size="16" font-family="sans-serif">' + (userInfo.value.username?.[0] || 'U') + '</text></svg>')
})

const displayRole = computed(() => {
  return roleMap[userInfo.value.role] || userInfo.value.role || '普通用户'
})

// 动态页面标题
const currentPageTitle = computed(() => {
  const routeMeta = route.meta as { title?: string }
  return routeMeta.title || '系统管理'
})

const handleAvatarError = (e: Event) => {
  const img = e.target as HTMLImageElement
  img.src = 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 36 36"><circle cx="18" cy="18" r="18" fill="#e2e8f0"/><text x="18" y="23" text-anchor="middle" fill="#64748b" font-size="16" font-family="sans-serif">' + (userInfo.value.username?.[0] || 'U') + '</text></svg>')
}

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    document.cookie = 'fayhub_token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;'
    localStorage.removeItem('userInfo')
    ElMessage.success('已安全退出')
    router.push('/')
  } catch {}
}

const handleUserCommand = (command: string) => {
  switch (command) {
    case 'profile':
      router.push('/dashboard')
      break
    case 'settings':
      router.push('/system/settings')
      break
    case 'logout':
      handleLogout()
      break
  }
}

async function fetchMenus() {
  try {
    const res = await menuApi.getMenuTree()
    menuItems.value = res.data || []
    registerPluginRoutes(menuItems.value)
  } catch (err: any) {
    console.error('获取菜单失败:', err)
  }
}

function registerPluginRoutes(menus: MenuType[]) {
  for (const menu of menus) {
    if (menu.children && menu.children.length > 0) {
      for (const child of menu.children) {
        if (child.path && child.component && !registeredPluginRoutes.value.has(child.path)) {
          const existingRoute = router.getRoutes().find(r => r.path === child.path)
          if (!existingRoute) {
            router.addRoute('admin', {
              path: child.path,
              name: `plugin-${child.path.replace(/\//g, '-')}`,
              component: () => import('@/views/PluginPage.vue'),
              meta: {
                requiresAuth: true,
                title: child.title,
                pluginId: child.component
              },
              props: { pluginId: child.component }
            })
          }
          registeredPluginRoutes.value.add(child.path)
        }
      }
    }
  }
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  const storedUser = localStorage.getItem('userInfo')
  if (storedUser) {
    try {
      const parsed = JSON.parse(storedUser)
      userInfo.value = { ...userInfo.value, ...parsed }
    } catch {}
  }
  
  fetchMenus()
})

watch(() => route.path, () => {
  const refreshFlag = localStorage.getItem('menu_refresh_needed')
  if (refreshFlag === 'true') {
    localStorage.removeItem('menu_refresh_needed')
    fetchMenus()
  }
})

let menuRefreshTimer: ReturnType<typeof setInterval> | null = null
onMounted(() => {
  menuRefreshTimer = setInterval(() => {
    const refreshFlag = localStorage.getItem('menu_refresh_needed')
    if (refreshFlag === 'true') {
      localStorage.removeItem('menu_refresh_needed')
      fetchMenus()
    }
  }, 1000)
})
onBeforeUnmount(() => {
  if (menuRefreshTimer) {
    clearInterval(menuRefreshTimer)
    menuRefreshTimer = null
  }
  window.removeEventListener('resize', checkMobile)
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
