import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

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
          path: '/system/audit',
          name: 'audit-logs',
          component: () => import('@/views/AuditLogs.vue'),
          meta: { requiresAuth: true, title: '审计日志' }
        },
        {
          path: '/system/files',
          name: 'file-management',
          component: () => import('@/views/FileManagement.vue'),
          meta: { requiresAuth: true, title: '文件管理' }
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
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/NotFound.vue')
    }
  ]
})

function getTokenFromCookie(): string {
  const match = document.cookie.match(/(?:^|;\s*)fayhub_token=([^;]*)/)
  return match ? decodeURIComponent(match[1]) : ''
}

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
  const token = getTokenFromCookie()

  if (to.meta.requiresAuth) {
    if (!token || isTokenExpired(token)) {
      document.cookie = 'fayhub_token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;'
      localStorage.removeItem('userInfo')
      next({ path: '/', query: { redirect: to.fullPath } })
      return
    }

    const userStore = useUserStore()
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
        next('/dashboard')
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
