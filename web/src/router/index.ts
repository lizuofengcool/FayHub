import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'
import menuApi from '@/api/menu'

const publicPaths = new Set([
  '/dashboard',
  '/profile',
  '/system/notifications',
  '/system/webhook',
  '/system/webhooks',
  '/system/audit',
  '/system/files',
  '/system/user',
  '/system/role',
  '/system/department',
  '/plugins/installed',
  '/payment/transactions'
])

let cachedAllowedPaths: Set<string> | null = null

async function getAllowedPaths(): Promise<Set<string>> {
  if (cachedAllowedPaths) return cachedAllowedPaths

  const userStore = useUserStore()
  if (userStore.isSuperAdmin) {
    cachedAllowedPaths = new Set(['*'])
    return cachedAllowedPaths
  }

  try {
    const res = await menuApi.getMenuTree()
    const menus = res.data || []
    const paths = new Set<string>()

    function collectPaths(menuList: any[]) {
      for (const m of menuList) {
        if (m.path) paths.add(m.path)
        if (m.children?.length) collectPaths(m.children)
      }
    }
    collectPaths(menus)
    cachedAllowedPaths = paths
    return paths
  } catch {
    cachedAllowedPaths = publicPaths
    return cachedAllowedPaths
  }
}

export function clearAllowedPathsCache() {
  cachedAllowedPaths = null
}

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'login',
      component: () => import('@/views/Login.vue'),
      meta: { title: '登录' }
    },
    {
      path: '/admin',
      component: () => import('@/layouts/AdminLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '/dashboard',
          name: 'dashboard',
          component: () => import('@/views/Dashboard.vue'),
          meta: { requiresAuth: true, title: '仪表盘' }
        },
        {
          path: '/profile',
          name: 'profile',
          component: () => import('@/views/UserProfile.vue'),
          meta: { requiresAuth: true, title: '个人中心' }
        },
        {
          path: '/system/tenant',
          name: 'tenants',
          component: () => import('@/views/TenantManagement.vue'),
          meta: { requiresAuth: true, title: '租户管理', roles: ['super_admin'] }
        },
        {
          path: '/system/user',
          name: 'users',
          component: () => import('@/views/UserManagement.vue'),
          meta: { requiresAuth: true, title: '用户管理' }
        },
        {
          path: '/system/role',
          name: 'roles',
          component: () => import('@/views/RoleManagement.vue'),
          meta: { requiresAuth: true, title: '角色权限' }
        },
        {
          path: '/system/department',
          name: 'departments',
          component: () => import('@/views/DepartmentManagement.vue'),
          meta: { requiresAuth: true, title: '部门管理' }
        },
        {
          path: '/system/menu',
          name: 'menus',
          component: () => import('@/views/MenuManagement.vue'),
          meta: { requiresAuth: true, title: '菜单管理', roles: ['super_admin'] }
        },
        {
          path: '/system/api',
          name: 'apis',
          component: () => import('@/views/ApiManagement.vue'),
          meta: { requiresAuth: true, title: 'API管理', roles: ['super_admin'] }
        },
        {
          path: '/system/settings',
          name: 'system-settings',
          component: () => import('@/views/SystemSettings.vue'),
          meta: { requiresAuth: true, title: '系统设置', roles: ['super_admin'] }
        },
        {
          path: '/plugins/installed',
          name: 'plugins-installed',
          component: () => import('@/views/InstalledPlugins.vue'),
          meta: { requiresAuth: true, title: '插件管理' }
        },
        {
          path: '/plugins/versions',
          name: 'plugins-versions',
          component: () => import('@/views/PluginVersionManagement.vue'),
          meta: { requiresAuth: true, title: '插件版本管理' }
        },
        {
          path: '/plugins/engine',
          name: 'plugins-engine',
          component: () => import('@/views/EngineStatus.vue'),
          meta: { requiresAuth: true, title: '插件引擎', roles: ['super_admin'] }
        },
        {
          path: '/system/webhook',
          name: 'webhook',
          component: () => import('@/views/WebhookManagement.vue'),
          meta: { requiresAuth: true, title: 'Webhook管理' }
        },
        {
          path: '/system/webhooks',
          name: 'webhooks',
          component: () => import('@/views/WebhookManagement.vue'),
          meta: { requiresAuth: true, title: 'Webhook管理' }
        },
        {
          path: '/system/audit',
          name: 'audit-logs',
          component: () => import('@/views/AuditLogs.vue'),
          meta: { requiresAuth: true, title: '审计日志' }
        },
        {
          path: '/system/backups',
          name: 'backups',
          component: () => import('@/views/BackupManagement.vue'),
          meta: { requiresAuth: true, title: '备份管理', roles: ['super_admin'] }
        },
        {
          path: '/system/monitor',
          name: 'system-monitor',
          component: () => import('@/views/SystemMonitor.vue'),
          meta: { requiresAuth: true, title: '系统监控', roles: ['super_admin'] }
        },
        {
          path: '/system/files',
          name: 'file-management',
          component: () => import('@/views/FileManagement.vue'),
          meta: { requiresAuth: true, title: '文件管理' }
        },
        {
          path: '/system/notifications',
          name: 'notification-center',
          component: () => import('@/views/NotificationCenter.vue'),
          meta: { requiresAuth: true, title: '通知中心' }
        },
        {
          path: '/system/api-keys',
          name: 'api-keys',
          component: () => import('@/views/APIKeyManagement.vue'),
          meta: { requiresAuth: true, title: 'API密钥管理', roles: ['super_admin'] }
        },
        {
          path: '/payment/settlement',
          name: 'settlement',
          component: () => import('@/views/SettlementManagement.vue'),
          meta: { requiresAuth: true, title: '结算管理', roles: ['super_admin'] }
        },
        {
          path: '/payment/config',
          name: 'payment-config',
          component: () => import('@/views/PaymentConfig.vue'),
          meta: { requiresAuth: true, title: '支付参数配置', roles: ['super_admin'] }
        },
        {
          path: '/payment/transactions',
          name: 'payment-transactions',
          component: () => import('@/views/PaymentTransactions.vue'),
          meta: { requiresAuth: true, title: '交易记录' }
        },
        {
          path: '/plugin-apps/:pluginPath(.*)',
          name: 'plugin-apps-dynamic',
          component: () => import('@/views/PluginPage.vue'),
          meta: { requiresAuth: true, title: '插件应用', layout: 'embedded' }
        }
      ]
    },
    {
      path: '/plugin-fullscreen/:pluginPath(.*)',
      name: 'plugin-fullscreen-dynamic',
      component: () => import('@/views/PluginFullscreen.vue'),
      meta: { requiresAuth: true, title: '插件应用', layout: 'fullscreen' }
    },
    {
      path: '/forbidden',
      name: 'forbidden',
      component: () => import('@/views/Forbidden.vue'),
      meta: { title: '无权限' }
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/NotFound.vue')
    }
  ]
})

function isTokenExpired(token: string): boolean {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    return payload.exp * 1000 < Date.now()
  } catch {
    return true
  }
}

const whiteList = ['/']

router.beforeEach(async (to, _from, next) => {
  const userStore = useUserStore()
  const token = userStore.token

  if (to.meta.requiresAuth) {
    if (!token || isTokenExpired(token)) {
      localStorage.removeItem('userInfo')
      next({ path: '/', query: { redirect: to.fullPath } })
      return
    }

    if (!userStore.userInfo) {
      try {
        await userStore.fetchCurrentUser()
      } catch {
        await userStore.logout()
        next({ path: '/', query: { redirect: to.fullPath } })
        return
      }
    }

    if (to.meta.roles && Array.isArray(to.meta.roles)) {
      const userRole = userStore.userInfo?.role
      if (!userRole || !(to.meta.roles as string[]).includes(userRole)) {
        next('/forbidden')
        return
      }
    }

    if (to.path !== '/dashboard' && to.path !== '/forbidden') {
      const allowed = await getAllowedPaths()
      if (!allowed.has('*') && !allowed.has(to.path)) {
        next('/forbidden')
        return
      }
    }

    next()
  } else if (whiteList.includes(to.path) && token && !isTokenExpired(token)) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
