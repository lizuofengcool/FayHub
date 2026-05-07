<template>
  <div class="login-log-page">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold text-slate-800">登录日志</h2>
        <p class="text-slate-500 mt-1 text-sm">查看系统登录记录与安全审计</p>
      </div>
      <el-button type="danger" @click="openCleanupDialog">
        <el-icon class="mr-1"><Delete /></el-icon>
        清理历史日志
      </el-button>
    </div>

    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 border-b border-slate-100 flex gap-3 flex-wrap">
        <el-input v-model="filters.username" placeholder="用户名" clearable style="width: 150px" />
        <el-select v-model="filters.login_status" placeholder="登录状态" clearable style="width: 120px">
          <el-option label="成功" value="success" />
          <el-option label="失败" value="failed" />
        </el-select>
        <el-input v-model="filters.login_ip" placeholder="IP地址" clearable style="width: 150px" />
        <el-date-picker
          v-model="dateRange"
          type="datetimerange"
          range-separator="至"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          format="YYYY-MM-DD HH:mm"
          value-format="YYYY-MM-DDTHH:mm:ssZ"
          style="width: 360px"
        />
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>

      <el-table v-loading="loading" :data="logs" stripe class="w-full">
        <el-table-column prop="username" label="用户名" width="120">
          <template #default="{ row }">
            <span>{{ row.username || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="login_status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.login_status === 'success' ? 'success' : 'danger'" size="small">
              {{ row.login_status === 'success' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="login_ip" label="登录IP" width="140" />
        <el-table-column prop="browser" label="浏览器" width="100" />
        <el-table-column prop="os" label="操作系统" width="100" />
        <el-table-column prop="msg" label="提示信息" min-width="180" show-overflow-tooltip />
        <el-table-column prop="login_time" label="登录时间" width="170" />
      </el-table>

      <div class="p-4 flex justify-end">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @current-change="fetchLogs"
          @size-change="fetchLogs"
        />
      </div>
    </div>

    <el-dialog v-model="cleanupVisible" title="清理历史日志" width="400px">
      <el-form label-width="100px">
        <el-form-item label="清理范围">
          <el-radio-group v-model="cleanupDays">
            <el-radio :value="30">30天前</el-radio>
            <el-radio :value="90">90天前</el-radio>
            <el-radio :value="180">180天前</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="cleanupVisible = false">取消</el-button>
        <el-button type="danger" :loading="cleanupLoading" @click="handleCleanup">确定清理</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete } from '@element-plus/icons-vue'
import loginLogApi, { type LoginLog } from '@/api/loginLog'

const loading = ref(false)
const logs = ref<LoginLog[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const dateRange = ref<string[]>([])

const filters = reactive({
  username: '',
  login_status: '',
  login_ip: ''
})

const cleanupVisible = ref(false)
const cleanupDays = ref(90)
const cleanupLoading = ref(false)

async function fetchLogs() {
  loading.value = true
  try {
    const params: any = {
      page: page.value,
      page_size: pageSize.value,
      username: filters.username || undefined,
      login_status: filters.login_status || undefined,
      login_ip: filters.login_ip || undefined
    }
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_time = dateRange.value[0]
      params.end_time = dateRange.value[1]
    }
    const res = await loginLogApi.listLogs(params)
    logs.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (err: any) {
    ElMessage.error(err.message || '获取登录日志失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  page.value = 1
  fetchLogs()
}

function resetFilters() {
  filters.username = ''
  filters.login_status = ''
  filters.login_ip = ''
  dateRange.value = []
  page.value = 1
  fetchLogs()
}

function openCleanupDialog() {
  cleanupVisible.value = true
}

async function handleCleanup() {
  try {
    await ElMessageBox.confirm(
      `确定清理${cleanupDays.value}天前的登录日志？此操作不可恢复。`,
      '确认清理',
      { confirmButtonText: '确定清理', cancelButtonText: '取消', type: 'warning' }
    )
  } catch {
    return
  }
  cleanupLoading.value = true
  try {
    await loginLogApi.cleanup(cleanupDays.value)
    ElMessage.success('清理完成')
    cleanupVisible.value = false
    fetchLogs()
  } catch (err: any) {
    ElMessage.error(err.message || '清理失败')
  } finally {
    cleanupLoading.value = false
  }
}

onMounted(() => {
  fetchLogs()
})
</script>

<style scoped>
</style>
