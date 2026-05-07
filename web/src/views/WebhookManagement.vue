<template>
  <div class="webhook-page">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold text-slate-800">Webhook 管理</h2>
        <p class="text-slate-500 mt-1 text-sm">管理系统事件订阅与消息推送</p>
      </div>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon class="mr-1"><Plus /></el-icon>
        新建订阅
      </el-button>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <div class="bg-white rounded-xl border border-slate-100 p-4 shadow-sm">
        <p class="text-sm text-slate-500">总投递数</p>
        <p class="text-2xl font-bold text-slate-800 mt-1">{{ stats.total_deliveries }}</p>
      </div>
      <div class="bg-white rounded-xl border border-slate-100 p-4 shadow-sm">
        <p class="text-sm text-slate-500">成功数</p>
        <p class="text-2xl font-bold text-green-600 mt-1">{{ stats.success_count }}</p>
      </div>
      <div class="bg-white rounded-xl border border-slate-100 p-4 shadow-sm">
        <p class="text-sm text-slate-500">失败数</p>
        <p class="text-2xl font-bold text-red-500 mt-1">{{ stats.failed_count }}</p>
      </div>
      <div class="bg-white rounded-xl border border-slate-100 p-4 shadow-sm">
        <p class="text-sm text-slate-500">成功率</p>
        <p class="text-2xl font-bold text-blue-600 mt-1">{{ successRateDisplay }}</p>
      </div>
    </div>

    <el-tabs v-model="activeTab">
      <el-tab-pane label="订阅列表" name="subscriptions">
        <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
          <el-table v-loading="subLoading" :data="subscriptions" stripe class="w-full">
            <el-table-column prop="name" label="名称" min-width="150">
              <template #default="{ row }">
                <div>
                  <p class="font-medium text-slate-800">{{ row.name }}</p>
                  <p class="text-xs text-slate-400 truncate max-w-[200px]">{{ row.url }}</p>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="events" label="事件" min-width="200">
              <template #default="{ row }">
                <el-tag v-for="evt in (row.events || []).slice(0, 3)" :key="evt" size="small" class="mr-1 mb-1">
                  {{ evt }}
                </el-tag>
                <el-tag v-if="(row.events || []).length > 3" size="small" type="info" class="mr-1">
                  +{{ row.events.length - 3 }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="is_active" label="状态" width="90" align="center">
              <template #default="{ row }">
                <el-tag :type="row.is_active ? 'success' : 'info'" size="small">
                  {{ row.is_active ? '启用' : '停用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="160" />
            <el-table-column label="操作" width="260" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openEditDialog(row)">编辑</el-button>
                <el-button type="success" link size="small" @click="handleTest(row)">测试</el-button>
                <el-button
                  v-if="row.is_active"
                  type="warning" link size="small"
                  @click="toggleActive(row, false)"
                >停用</el-button>
                <el-button
                  v-if="!row.is_active"
                  type="success" link size="small"
                  @click="toggleActive(row, true)"
                >启用</el-button>
                <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>

      <el-tab-pane label="投递记录" name="deliveries">
        <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
          <div class="p-4 flex gap-3 flex-wrap">
            <el-select v-model="deliveryFilter.status" placeholder="状态筛选" clearable size="default" style="width: 140px">
              <el-option label="成功" value="success" />
              <el-option label="失败" value="failed" />
              <el-option label="重试中" value="retrying" />
              <el-option label="待投递" value="pending" />
            </el-select>
            <el-button type="primary" @click="fetchDeliveries">查询</el-button>
          </div>
          <el-table v-loading="deliveryLoading" :data="deliveries" stripe class="w-full">
            <el-table-column prop="event" label="事件" width="160" />
            <el-table-column prop="status" label="状态" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="deliveryStatusType(row.status)" size="small">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="status_code" label="HTTP状态码" width="110" align="center" />
            <el-table-column prop="attempts" label="重试次数" width="100" align="center" />
            <el-table-column prop="delivered_at" label="投递时间" width="160" />
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="viewDeliveryDetail(row)">详情</el-button>
                <el-button
                  v-if="row.status === 'failed'"
                  type="warning" link size="small"
                  @click="handleRedeliver(row)"
                >重发</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div class="p-4 flex justify-end">
            <el-pagination
              v-model:current-page="deliveryPage"
              :page-size="10"
              :total="deliveryTotal"
              layout="total, prev, pager, next"
              @current-change="fetchDeliveries"
            />
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑订阅' : '新建订阅'" width="560px">
      <el-form :model="form" :rules="formRules" ref="formRef" label-width="90px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入订阅名称" />
        </el-form-item>
        <el-form-item label="URL" prop="url">
          <el-input v-model="form.url" placeholder="请输入回调URL" />
        </el-form-item>
        <el-form-item label="Secret">
          <el-input v-model="form.secret" placeholder="签名密钥（可选）" show-password />
        </el-form-item>
        <el-form-item label="事件" prop="events">
          <el-select v-model="form.events" multiple placeholder="请选择订阅事件" class="w-full">
            <el-option label="插件安装" value="plugin.installed" />
            <el-option label="插件卸载" value="plugin.uninstalled" />
            <el-option label="插件启用" value="plugin.enabled" />
            <el-option label="插件禁用" value="plugin.disabled" />
            <el-option label="插件升级" value="plugin.upgraded" />
            <el-option label="用户创建" value="user.created" />
            <el-option label="用户删除" value="user.deleted" />
            <el-option label="租户创建" value="tenant.created" />
            <el-option label="支付完成" value="payment.completed" />
            <el-option label="订单创建" value="order.created" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="描述信息（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" title="投递详情" width="640px">
      <div v-if="currentDelivery">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="事件">{{ currentDelivery.event }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="deliveryStatusType(currentDelivery.status)" size="small">{{ currentDelivery.status }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="HTTP状态码">{{ currentDelivery.status_code }}</el-descriptions-item>
          <el-descriptions-item label="重试次数">{{ currentDelivery.attempts }}</el-descriptions-item>
          <el-descriptions-item label="投递时间" :span="2">{{ currentDelivery.delivered_at }}</el-descriptions-item>
        </el-descriptions>
        <div class="mt-4">
          <p class="text-sm font-medium text-slate-700 mb-2">请求载荷</p>
          <pre class="bg-slate-50 rounded-lg p-3 text-xs overflow-auto max-h-48">{{ JSON.stringify(currentDelivery.payload, null, 2) }}</pre>
        </div>
        <div v-if="currentDelivery.response_body" class="mt-4">
          <p class="text-sm font-medium text-slate-700 mb-2">响应内容</p>
          <pre class="bg-slate-50 rounded-lg p-3 text-xs overflow-auto max-h-48">{{ currentDelivery.response_body }}</pre>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import webhookApi, { type WebhookSubscription, type WebhookDelivery, type WebhookStats } from '@/api/webhook'

function getErrMsg(err: unknown, fallback: string): string {
  if (err instanceof Error) return err.message || fallback
  return fallback
}

const activeTab = ref('subscriptions')
const subLoading = ref(false)
const subscriptions = ref<WebhookSubscription[]>([])
const stats = ref<WebhookStats>({ total_deliveries: 0, success_count: 0, failed_count: 0, pending_count: 0, success_rate: 0 })

const successRateDisplay = computed(() => {
  if (stats.value.total_deliveries === 0) return '-'
  return (stats.value.success_rate * 100).toFixed(1) + '%'
})

const deliveryLoading = ref(false)
const deliveries = ref<WebhookDelivery[]>([])
const deliveryPage = ref(1)
const deliveryTotal = ref(0)
const deliveryFilter = reactive({ status: '' })

const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(0)
const submitLoading = ref(false)
const formRef = ref()
const form = reactive({
  name: '',
  url: '',
  secret: '',
  events: [] as string[],
  description: ''
})
const formRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  url: [{ required: true, message: '请输入URL', trigger: 'blur' }, { type: 'url' as const, message: '请输入有效的URL', trigger: 'blur' }],
  events: [{ required: true, message: '请选择事件', trigger: 'change' }]
}

const detailVisible = ref(false)
const currentDelivery = ref<WebhookDelivery | null>(null)

async function fetchSubscriptions() {
  subLoading.value = true
  try {
    const res = await webhookApi.listSubscriptions()
    subscriptions.value = res.data?.list || []
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '获取订阅列表失败'))
  } finally {
    subLoading.value = false
  }
}

async function fetchStats() {
  try {
    const res = await webhookApi.getDeliveryStats()
    const raw = res.data || {}
    const delivered = Number(raw.delivered || 0)
    const failed = Number(raw.failed || 0)
    const pending = Number(raw.pending || 0)
    const retrying = Number(raw.retrying || 0)
    const total = delivered + failed + pending + retrying
    const rate = total > 0 ? delivered / total : 0
    stats.value = {
      total_deliveries: total,
      success_count: delivered,
      failed_count: failed,
      pending_count: pending + retrying,
      success_rate: rate
    }
  } catch (e) { console.error('fetchStats failed:', e); }
}

async function fetchDeliveries() {
  deliveryLoading.value = true
  try {
    const res = await webhookApi.listDeliveries({
      page: deliveryPage.value,
      page_size: 10,
      status: deliveryFilter.status || undefined
    })
    deliveries.value = res.data?.list || []
    deliveryTotal.value = res.data?.total || 0
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '获取投递记录失败'))
  } finally {
    deliveryLoading.value = false
  }
}

function openCreateDialog() {
  isEdit.value = false
  editId.value = 0
  form.name = ''
  form.url = ''
  form.secret = ''
  form.events = []
  form.description = ''
  dialogVisible.value = true
}

function openEditDialog(row: WebhookSubscription) {
  isEdit.value = true
  editId.value = row.id
  form.name = row.name
  form.url = row.url
  form.secret = ''
  form.events = [...(row.events || [])]
  form.description = row.description || ''
  dialogVisible.value = true
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
  } catch { return }

  submitLoading.value = true
  try {
    if (isEdit.value) {
      await webhookApi.updateSubscription(editId.value, {
        name: form.name,
        url: form.url,
        secret: form.secret || undefined,
        events: form.events,
        description: form.description || undefined
      })
      ElMessage.success('更新成功')
    } else {
      await webhookApi.createSubscription({
        name: form.name,
        url: form.url,
        secret: form.secret || undefined,
        events: form.events,
        description: form.description || undefined
      })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchSubscriptions()
    fetchStats()
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '操作失败'))
  } finally {
    submitLoading.value = false
  }
}

async function toggleActive(row: WebhookSubscription, active: boolean) {
  try {
    await webhookApi.updateSubscription(row.id, { is_active: active })
    ElMessage.success(active ? '已启用' : '已停用')
    fetchSubscriptions()
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '操作失败'))
  }
}

async function handleDelete(row: WebhookSubscription) {
  try {
    await ElMessageBox.confirm('确定要删除此订阅吗？', '确认删除', { type: 'warning' })
    await webhookApi.deleteSubscription(row.id)
    ElMessage.success('删除成功')
    fetchSubscriptions()
    fetchStats()
  } catch (e) { console.error('handleDeleteSubscription failed:', e); }
}

async function handleTest(row: WebhookSubscription) {
  try {
    await ElMessageBox.confirm(`将向 ${row.url} 发送一条测试消息，是否继续？`, '测试投递', { type: 'info', confirmButtonText: '发送', cancelButtonText: '取消' })
    const res = await webhookApi.testDelivery(row.id)
    if (res.code === 200) {
      ElMessage.success('测试投递已触发，请查看投递记录')
      fetchDeliveries()
      fetchStats()
    }
  } catch (e) { console.error('handleTest failed:', e); }
}

async function handleRedeliver(row: WebhookDelivery) {
  try {
    await webhookApi.redeliver(row.id)
    ElMessage.success('已重新投递')
    fetchDeliveries()
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '重发失败'))
  }
}

function viewDeliveryDetail(row: WebhookDelivery) {
  currentDelivery.value = row
  detailVisible.value = true
}

function deliveryStatusType(status: string) {
  switch (status) {
    case 'success': return 'success'
    case 'failed': return 'danger'
    case 'retrying': return 'warning'
    default: return 'info'
  }
}

onMounted(() => {
  fetchSubscriptions()
  fetchStats()
  fetchDeliveries()
})
</script>

<style scoped>
</style>
