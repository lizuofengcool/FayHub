<template>
  <div class="online-user-page">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold text-slate-800">在线用户</h2>
        <p class="text-slate-500 mt-1 text-sm">
          当前在线 <span class="text-blue-600 font-semibold">{{ onlineCount }}</span> 人
        </p>
      </div>
      <el-button @click="fetchUsers">
        <el-icon class="mr-1"><Refresh /></el-icon>
        刷新
      </el-button>
    </div>

    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <el-table v-loading="loading" :data="users" stripe class="w-full">
        <el-table-column prop="username" label="用户名" width="140">
          <template #default="{ row }">
            <el-link type="primary" @click="viewLoginLogs(row)">
              {{ row.username }}
            </el-link>
          </template>
        </el-table-column>
        <el-table-column prop="nickname" label="昵称" width="140">
          <template #default="{ row }">
            <span>{{ row.nickname || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="department" label="部门" width="140">
          <template #default="{ row }">
            <span>{{ row.department || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="role" label="角色" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="row.role === 'super_admin' ? 'danger' : row.role === 'admin' ? 'warning' : 'info'">
              {{ row.role === 'super_admin' ? '超管' : row.role === 'admin' ? '管理员' : row.role }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ip" label="IP地址" width="130" />
        <el-table-column prop="location" label="登录地点" width="130">
          <template #default="{ row }">
            <span>{{ row.location || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作系统" width="120">
          <template #default="{ row }">
            <span>{{ parseOS(row.user_agent) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="浏览器" width="120">
          <template #default="{ row }">
            <span>{{ parseBrowser(row.user_agent) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="login_at" label="登录时间" width="170">
          <template #default="{ row }">
            <span>{{ formatTime(row.login_at) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="last_seen" label="最后活跃" width="170">
          <template #default="{ row }">
            <span>{{ formatTime(row.last_seen) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" align="center" fixed="right">
          <template #default="{ row }">
            <el-popconfirm
              title="确定要强制该用户下线吗？"
              confirm-button-text="确定"
              cancel-button-text="取消"
              @confirm="handleForceLogout(row)"
            >
              <template #reference>
                <el-button type="danger" size="small" link>
                  <el-icon><SwitchButton /></el-icon>
                  强制下线
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="users.length === 0 && !loading" class="py-16 text-center text-slate-400">
        <el-icon :size="48" class="mb-3"><User /></el-icon>
        <p>暂无在线用户</p>
      </div>
    </div>

    <!-- 用户登录记录对话框 -->
    <el-dialog v-model="loginLogDialogVisible" :title="`${selectedUser?.username || ''} 的登录记录`" width="1000px">
      <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
        <div class="p-4 border-b border-slate-100 flex gap-3 flex-wrap">
          <el-select v-model="loginLogFilters.login_status" placeholder="登录状态" clearable style="width: 120px">
            <el-option label="成功" value="success" />
            <el-option label="失败" value="failed" />
          </el-select>
          <el-input v-model="loginLogFilters.login_ip" placeholder="IP地址" clearable style="width: 150px" />
          <el-date-picker
            v-model="loginLogDateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            format="YYYY-MM-DD HH:mm"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            style="width: 360px"
          />
          <el-button type="primary" @click="handleLoginLogSearch">查询</el-button>
          <el-button @click="resetLoginLogFilters">重置</el-button>
        </div>

        <el-table v-loading="loginLogLoading" :data="loginLogs" stripe class="w-full">
          <el-table-column prop="login_status" label="状态" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="row.login_status === 'success' ? 'success' : 'danger'" size="small">
                {{ row.login_status === 'success' ? '成功' : '失败' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="login_ip" label="登录IP" width="140" />
          <el-table-column prop="location" label="登录地点" width="140">
            <template #default="{ row }">
              <span>{{ row.location || '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="browser" label="浏览器" width="100" />
          <el-table-column prop="os" label="操作系统" width="100" />
          <el-table-column prop="msg" label="提示信息" min-width="180" show-overflow-tooltip />
          <el-table-column prop="login_time" label="登录时间" width="170" />
        </el-table>

        <div class="p-4 flex justify-end">
          <el-pagination
            v-model:current-page="loginLogPage"
            v-model:page-size="loginLogPageSize"
            :total="loginLogTotal"
            :page-sizes="[20, 50, 100]"
            layout="total, sizes, prev, pager, next"
            @current-change="fetchLoginLogs"
            @size-change="fetchLoginLogs"
          />
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import onlineUserApi, { type OnlineUser } from '@/api/onlineUser'
import loginLogApi from '@/api/loginLog'

const loading = ref(false)
const users = ref<OnlineUser[]>([])
const onlineCount = ref(0)
let timer: ReturnType<typeof setInterval> | null = null

// 登录记录相关
const loginLogDialogVisible = ref(false)
const selectedUser = ref<OnlineUser | null>(null)
const loginLogLoading = ref(false)
const loginLogs = ref<any[]>([])
const loginLogPage = ref(1)
const loginLogPageSize = ref(20)
const loginLogTotal = ref(0)
const loginLogDateRange = ref<string[]>([])
const loginLogFilters = reactive({
  login_status: '',
  login_ip: ''
})

function formatTime(time: string) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

function parseOS(userAgent: string): string {
  if (!userAgent) return '-'
  if (userAgent.includes('Windows')) return 'Windows'
  if (userAgent.includes('Mac')) return 'macOS'
  if (userAgent.includes('Linux')) return 'Linux'
  if (userAgent.includes('Android')) return 'Android'
  if (userAgent.includes('iPhone') || userAgent.includes('iOS')) return 'iOS'
  return '未知'
}

function parseBrowser(userAgent: string): string {
  if (!userAgent) return '-'
  if (userAgent.includes('Chrome')) return 'Chrome'
  if (userAgent.includes('Firefox')) return 'Firefox'
  if (userAgent.includes('Safari')) return 'Safari'
  if (userAgent.includes('Edge')) return 'Edge'
  if (userAgent.includes('Opera')) return 'Opera'
  return '未知'
}

async function fetchUsers() {
  loading.value = true
  try {
    const [userRes, countRes] = await Promise.all([
      onlineUserApi.getOnlineUsers(),
      onlineUserApi.getOnlineCount()
    ])
    users.value = userRes.data || []
    onlineCount.value = countRes.data?.count || 0
  } catch {
    // ignore
  } finally {
    loading.value = false
  }
}

async function handleForceLogout(user: OnlineUser) {
  try {
    await onlineUserApi.forceLogout(user.user_id)
    ElMessage.success(`已将 ${user.username} 强制下线`)
    fetchUsers()
  } catch {
    ElMessage.error('操作失败')
  }
}

function viewLoginLogs(user: OnlineUser) {
  selectedUser.value = user
  loginLogPage.value = 1
  loginLogDialogVisible.value = true
  fetchLoginLogs()
}

async function fetchLoginLogs() {
  if (!selectedUser.value) return
  loginLogLoading.value = true
  try {
    const params: any = {
      page: loginLogPage.value,
      page_size: loginLogPageSize.value,
      username: selectedUser.value.username,
      login_status: loginLogFilters.login_status || undefined,
      login_ip: loginLogFilters.login_ip || undefined
    }
    if (loginLogDateRange.value && loginLogDateRange.value.length === 2) {
      params.start_time = loginLogDateRange.value[0]
      params.end_time = loginLogDateRange.value[1]
    }
    const res = await loginLogApi.listLogs(params)
    loginLogs.value = res.data?.list || []
    loginLogTotal.value = res.data?.total || 0
  } catch (err: any) {
    ElMessage.error(err.message || '获取登录日志失败')
  } finally {
    loginLogLoading.value = false
  }
}

function handleLoginLogSearch() {
  loginLogPage.value = 1
  fetchLoginLogs()
}

function resetLoginLogFilters() {
  loginLogFilters.login_status = ''
  loginLogFilters.login_ip = ''
  loginLogDateRange.value = []
  loginLogPage.value = 1
  fetchLoginLogs()
}

onMounted(() => {
  fetchUsers()
  timer = setInterval(fetchUsers, 30000)
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
})
</script>
