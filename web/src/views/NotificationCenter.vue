<template>
  <div class="notification-page">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold text-slate-800">通知中心</h2>
        <p class="text-slate-500 mt-1 text-sm">查看和管理系统通知消息</p>
      </div>
      <div class="flex gap-3">
        <el-button @click="handleMarkAllRead" :disabled="unreadCount === 0">
          <el-icon class="mr-1"><Check /></el-icon> 全部已读
        </el-button>
        <el-button type="danger" @click="handleBatchDelete" :disabled="selectedIds.length === 0">
          <el-icon class="mr-1"><Delete /></el-icon> 删除选中
        </el-button>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
      <div class="bg-white rounded-xl border border-slate-100 p-4 shadow-sm">
        <p class="text-sm text-slate-500">未读通知</p>
        <p class="text-2xl font-bold text-blue-600 mt-1">{{ unreadCount }}</p>
      </div>
      <div class="bg-white rounded-xl border border-slate-100 p-4 shadow-sm">
        <p class="text-sm text-slate-500">总通知数</p>
        <p class="text-2xl font-bold text-slate-800 mt-1">{{ total }}</p>
      </div>
      <div class="bg-white rounded-xl border border-slate-100 p-4 shadow-sm">
        <p class="text-sm text-slate-500">当前筛选</p>
        <p class="text-2xl font-bold text-slate-800 mt-1">{{ notifications.length }}</p>
      </div>
    </div>

    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 border-b border-slate-100 flex gap-3 flex-wrap">
        <el-select v-model="filters.type" placeholder="通知类型" clearable style="width: 140px" @change="fetchList">
          <el-option label="系统" value="system" />
          <el-option label="安全" value="security" />
          <el-option label="插件" value="plugin" />
          <el-option label="支付" value="payment" />
          <el-option label="Webhook" value="webhook" />
          <el-option label="审计" value="audit" />
        </el-select>
        <el-select v-model="filters.category" placeholder="级别" clearable style="width: 120px" @change="fetchList">
          <el-option label="信息" value="info" />
          <el-option label="成功" value="success" />
          <el-option label="警告" value="warning" />
          <el-option label="错误" value="error" />
        </el-select>
        <el-select v-model="filters.is_read" placeholder="已读状态" clearable style="width: 120px" @change="fetchList">
          <el-option label="未读" :value="false" />
          <el-option label="已读" :value="true" />
        </el-select>
        <el-button type="primary" @click="fetchList">查询</el-button>
      </div>

      <el-table
        v-loading="loading"
        :data="notifications"
        stripe
        class="w-full"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="50" />
        <el-table-column label="状态" width="70" align="center">
          <template #default="{ row }">
            <span v-if="!row.is_read" class="inline-block w-2 h-2 rounded-full bg-blue-500"></span>
            <span v-else class="inline-block w-2 h-2 rounded-full bg-slate-300"></span>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="200">
          <template #default="{ row }">
            <div :class="{ 'font-semibold': !row.is_read }" class="text-slate-800">
              {{ row.title }}
            </div>
            <div class="text-xs text-slate-400 truncate max-w-[300px]">{{ row.content }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="typeTagMap[row.type] || 'info'" size="small">{{ typeLabelMap[row.type] || row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="category" label="级别" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="categoryTagMap[row.category] || 'info'" size="small">{{ categoryLabelMap[row.category] || row.category }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sender_name" label="发送者" width="100" />
        <el-table-column prop="created_at" label="时间" width="160" />
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button v-if="!row.is_read" type="primary" link size="small" @click="handleMarkRead(row)">标为已读</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="p-4 flex justify-end">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @size-change="fetchList"
          @current-change="fetchList"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Check, Delete } from '@element-plus/icons-vue'
import notificationApi, { type Notification } from '@/api/notification'

const loading = ref(false)
const notifications = ref<Notification[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const unreadCount = ref(0)
const selectedIds = ref<number[]>([])

const filters = reactive({
  type: '' as string,
  category: '' as string,
  is_read: undefined as boolean | undefined
})

const typeTagMap: Record<string, string> = {
  system: '', security: 'danger', plugin: 'success', payment: 'warning', webhook: 'info', audit: ''
}
const typeLabelMap: Record<string, string> = {
  system: '系统', security: '安全', plugin: '插件', payment: '支付', webhook: 'Webhook', audit: '审计'
}
const categoryTagMap: Record<string, string> = {
  info: 'info', success: 'success', warning: 'warning', error: 'danger'
}
const categoryLabelMap: Record<string, string> = {
  info: '信息', success: '成功', warning: '警告', error: '错误'
}

async function fetchList() {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: page.value,
      page_size: pageSize.value
    }
    if (filters.type) params.type = filters.type
    if (filters.category) params.category = filters.category
    if (filters.is_read !== undefined) params.is_read = filters.is_read

    const res = await notificationApi.listNotifications(params)
    notifications.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (err: any) {
    ElMessage.error(err.message || '获取通知列表失败')
  } finally {
    loading.value = false
  }
}

async function fetchUnreadCount() {
  try {
    const res = await notificationApi.getUnreadCount()
    unreadCount.value = res.data?.unread_count || 0
  } catch {}
}

function handleSelectionChange(rows: Notification[]) {
  selectedIds.value = rows.map(r => r.id)
}

async function handleMarkRead(row: Notification) {
  try {
    await notificationApi.markAsRead([row.id])
    ElMessage.success('已标记为已读')
    fetchList()
    fetchUnreadCount()
  } catch (err: any) {
    ElMessage.error(err.message || '操作失败')
  }
}

async function handleMarkAllRead() {
  try {
    await notificationApi.markAllAsRead()
    ElMessage.success('已全部标记为已读')
    fetchList()
    fetchUnreadCount()
  } catch (err: any) {
    ElMessage.error(err.message || '操作失败')
  }
}

async function handleDelete(row: Notification) {
  try {
    await ElMessageBox.confirm('确定要删除此通知吗？', '确认删除', { type: 'warning' })
    await notificationApi.deleteNotifications([row.id])
    ElMessage.success('删除成功')
    fetchList()
    fetchUnreadCount()
  } catch {}
}

async function handleBatchDelete() {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedIds.value.length} 条通知吗？`, '确认删除', { type: 'warning' })
    await notificationApi.deleteNotifications(selectedIds.value)
    ElMessage.success('删除成功')
    selectedIds.value = []
    fetchList()
    fetchUnreadCount()
  } catch {}
}

onMounted(() => {
  fetchList()
  fetchUnreadCount()
})
</script>

<style scoped>
</style>
