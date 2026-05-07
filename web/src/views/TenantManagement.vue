<template>
  <div class="tenant-page">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold text-slate-800">租户管理</h2>
        <p class="text-slate-500 mt-1 text-sm">管理平台所有租户，仅超级管理员可操作</p>
      </div>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon class="mr-1"><Plus /></el-icon> 新建租户
      </el-button>
    </div>

    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 border-b border-slate-100">
        <div class="flex items-center justify-between">
          <el-radio-group v-model="activeTab" @change="onTabChange" size="default">
            <el-radio-button value="all">全部租户</el-radio-button>
            <el-radio-button value="recycle">回收站</el-radio-button>
          </el-radio-group>
          <el-form v-if="activeTab === 'all'" :inline="true" :model="searchForm" class="search-form">
            <el-form-item>
              <el-input v-model="searchForm.keyword" placeholder="租户名称/域名" clearable @keyup.enter="fetchList" />
            </el-form-item>
            <el-form-item>
              <el-select v-model="searchForm.status" placeholder="状态" clearable style="width: 100px">
                <el-option label="启用" :value="1" />
                <el-option label="禁用" :value="0" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="fetchList">查询</el-button>
              <el-button @click="resetSearch">重置</el-button>
            </el-form-item>
          </el-form>
        </div>
      </div>

      <el-table v-loading="loading" :data="tableData" stripe class="w-full">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="租户名称" min-width="140" />
        <el-table-column prop="domain" label="域名" min-width="160" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.status === 1" type="success" size="small">启用</el-tag>
            <el-tag v-else-if="row.status === 0" type="danger" size="small">禁用</el-tag>
            <el-tag v-else-if="row.status === 2" type="warning" size="small">回收站</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" min-width="160" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <template v-if="activeTab === 'all'">
              <el-button type="primary" link size="small" @click="openEditDialog(row)">编辑</el-button>
              <el-button
                :type="row.status === 1 ? 'warning' : 'success'"
                link
                size="small"
                @click="toggleStatus(row)"
              >
                {{ row.status === 1 ? '禁用' : '启用' }}
              </el-button>
              <el-button v-if="row.status === 1" type="success" link size="small" @click="handleImpersonate(row)">进入后台</el-button>
              <el-button type="warning" link size="small" @click="handleSoftDelete(row)">移入回收站</el-button>
            </template>
            <template v-if="activeTab === 'recycle'">
              <el-button type="success" link size="small" @click="handleRestore(row)">恢复</el-button>
              <el-button type="danger" link size="small" @click="handlePermanentDelete(row)">彻底删除</el-button>
            </template>
          </template>
        </el-table-column>
      </el-table>

      <div class="p-4 flex justify-end">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :total="pagination.total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchList"
          @current-change="fetchList"
        />
      </div>
    </div>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑租户' : '新建租户'" width="560px" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="租户名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入租户名称" />
        </el-form-item>
        <el-form-item label="域名" prop="domain">
          <el-input v-model="form.domain" placeholder="请输入租户域名，如 tenant.fayhub.com" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入租户描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确认</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="detailVisible" :title="`租户详情 — ${detailTenant?.name || ''}`" size="520px">
      <div v-if="detailTenant" class="space-y-6">
        <div>
          <h4 class="text-sm font-semibold text-slate-500 mb-3">基本信息</h4>
          <div class="grid grid-cols-2 gap-3 text-sm">
            <div><span class="text-slate-400">ID：</span>{{ detailTenant.id }}</div>
            <div><span class="text-slate-400">域名：</span>{{ detailTenant.domain }}</div>
            <div><span class="text-slate-400">状态：</span>
              <el-tag v-if="detailTenant.status === 1" type="success" size="small">启用</el-tag>
              <el-tag v-else-if="detailTenant.status === 0" type="danger" size="small">禁用</el-tag>
              <el-tag v-else-if="detailTenant.status === 2" type="warning" size="small">回收站</el-tag>
            </div>
            <div><span class="text-slate-400">创建时间：</span>{{ detailTenant.created_at?.slice(0, 10) }}</div>
            <div class="col-span-2"><span class="text-slate-400">描述：</span>{{ detailTenant.description || '-' }}</div>
          </div>
        </div>

        <el-divider />

        <div>
          <div class="flex items-center justify-between mb-3">
            <h4 class="text-sm font-semibold text-slate-500">资源配额与用量</h4>
            <el-button type="primary" link size="small" @click="syncUsage" :loading="syncLoading">同步用量</el-button>
          </div>
          <div v-if="detailQuota" class="space-y-4">
            <div>
              <div class="flex justify-between text-sm mb-1">
                <span>用户数</span>
                <span class="text-slate-500">{{ detailQuota.used_users }} / {{ detailQuota.max_users || '无限制' }}</span>
              </div>
              <el-progress :percentage="quotaPercent(detailQuota.used_users, detailQuota.max_users)" :color="progressColor(detailQuota.used_users, detailQuota.max_users)" />
            </div>
            <div>
              <div class="flex justify-between text-sm mb-1">
                <span>存储空间</span>
                <span class="text-slate-500">{{ detailQuota.used_storage_mb }}MB / {{ detailQuota.max_storage_mb ? detailQuota.max_storage_mb + 'MB' : '无限制' }}</span>
              </div>
              <el-progress :percentage="quotaPercent(detailQuota.used_storage_mb, detailQuota.max_storage_mb)" :color="progressColor(detailQuota.used_storage_mb, detailQuota.max_storage_mb)" />
            </div>
            <div>
              <div class="flex justify-between text-sm mb-1">
                <span>插件数</span>
                <span class="text-slate-500">{{ detailQuota.used_plugins }} / {{ detailQuota.max_plugins || '无限制' }}</span>
              </div>
              <el-progress :percentage="quotaPercent(detailQuota.used_plugins, detailQuota.max_plugins)" :color="progressColor(detailQuota.used_plugins, detailQuota.max_plugins)" />
            </div>
            <div>
              <div class="flex justify-between text-sm mb-1">
                <span>今日API调用</span>
                <span class="text-slate-500">{{ detailQuota.used_api_per_day }} / {{ detailQuota.max_api_per_day || '无限制' }}</span>
              </div>
              <el-progress :percentage="quotaPercent(detailQuota.used_api_per_day, detailQuota.max_api_per_day)" :color="progressColor(detailQuota.used_api_per_day, detailQuota.max_api_per_day)" />
            </div>
          </div>
          <el-empty v-else description="暂无配额数据" />
        </div>

        <el-divider />

        <div>
          <h4 class="text-sm font-semibold text-slate-500 mb-3">调整配额</h4>
          <el-form label-width="100px" size="small">
            <el-form-item label="最大用户数">
              <el-input-number v-model="quotaForm.max_users" :min="0" :step="1" />
            </el-form-item>
            <el-form-item label="最大存储(MB)">
              <el-input-number v-model="quotaForm.max_storage_mb" :min="0" :step="256" />
            </el-form-item>
            <el-form-item label="最大插件数">
              <el-input-number v-model="quotaForm.max_plugins" :min="0" :step="1" />
            </el-form-item>
            <el-form-item label="每日API上限">
              <el-input-number v-model="quotaForm.max_api_per_day" :min="0" :step="1000" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveQuota" :loading="quotaSaving">保存配额</el-button>
            </el-form-item>
          </el-form>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import tenantApi, { type Tenant, type CreateTenantParams, type UpdateTenantParams, type TenantQuota } from '@/api/tenant'

const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref<number>(0)
const formRef = ref<FormInstance>()
const tableData = ref<Tenant[]>([])
const activeTab = ref('all')

const searchForm = reactive({
  keyword: '',
  status: undefined as number | undefined
})

const pagination = reactive({
  page: 1,
  page_size: 10,
  total: 0
})

const form = reactive<CreateTenantParams>({
  name: '',
  domain: '',
  description: '',
  status: 1
})

const formRules = {
  name: [{ required: true, message: '请输入租户名称', trigger: 'blur' }],
  domain: [{ required: true, message: '请输入租户域名', trigger: 'blur' }]
}

onMounted(() => {
  fetchList()
})

async function fetchList() {
  loading.value = true
  try {
    const statusParam = activeTab.value === 'recycle' ? 2 : (searchForm.status !== undefined ? searchForm.status : undefined)
    const res = await tenantApi.getTenantList({
      page: pagination.page,
      page_size: pagination.page_size,
      keyword: searchForm.keyword || undefined,
      status: statusParam
    })
    tableData.value = res.data.list || []
    pagination.total = res.data.total || 0
  } catch (err: any) {
    ElMessage.error(err.message || '获取租户列表失败')
  } finally {
    loading.value = false
  }
}

function onTabChange() {
  pagination.page = 1
  fetchList()
}

function resetSearch() {
  searchForm.keyword = ''
  searchForm.status = undefined
  pagination.page = 1
  fetchList()
}

function openCreateDialog() {
  isEdit.value = false
  editId.value = 0
  Object.assign(form, {
    name: '',
    domain: '',
    description: '',
    status: 1
  })
  dialogVisible.value = true
}

function openEditDialog(row: Tenant) {
  isEdit.value = true
  editId.value = row.id
  Object.assign(form, {
    name: row.name,
    domain: row.domain,
    description: row.description,
    status: row.status
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (isEdit.value) {
      const params: UpdateTenantParams = {
        name: form.name,
        domain: form.domain,
        description: form.description
      }
      await tenantApi.updateTenant(editId.value, params)
      ElMessage.success('租户更新成功')
    } else {
      await tenantApi.createTenant(form)
      ElMessage.success('租户创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (err: any) {
    ElMessage.error(err.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}

async function toggleStatus(row: Tenant) {
  const newStatus = row.status === 1 ? 0 : 1
  const action = newStatus === 1 ? '启用' : '禁用'
  try {
    await ElMessageBox.confirm(`确定要${action}租户「${row.name}」吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await tenantApi.updateTenant(row.id, { status: newStatus })
    ElMessage.success(`${action}成功`)
    fetchList()
  } catch (e) { console.error('handleStatusChange failed:', e); }
}

async function handleSoftDelete(row: Tenant) {
  try {
    await ElMessageBox.confirm(
      `确定将租户「${row.name}」移入回收站吗？移入后该租户及其所有数据将被隐藏，可随时恢复。`,
      '移入回收站',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    await tenantApi.softDeleteTenant(row.id)
    ElMessage.success('已移入回收站')
    fetchList()
  } catch (e) { console.error('handleSoftDelete failed:', e); }
}

async function handleRestore(row: Tenant) {
  try {
    await ElMessageBox.confirm(
      `确定恢复租户「${row.name}」吗？恢复后该租户及其所有数据将恢复正常。`,
      '恢复租户',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'info' }
    )
    await tenantApi.restoreTenant(row.id)
    ElMessage.success('租户已恢复')
    fetchList()
  } catch (e) { console.error('handleRestore failed:', e); }
}

async function handlePermanentDelete(row: Tenant) {
  try {
    await ElMessageBox.confirm(
      `确定永久删除租户「${row.name}」吗？此操作将物理删除该租户及其所有关联数据，不可恢复！`,
      '永久删除',
      { confirmButtonText: '确定删除', cancelButtonText: '取消', type: 'error' }
    )
    await tenantApi.permanentDeleteTenant(row.id)
    ElMessage.success('租户已永久删除')
    fetchList()
  } catch (e) { console.error('handlePermanentDelete failed:', e); }
}

async function handleImpersonate(row: Tenant) {
  try {
    await ElMessageBox.confirm(
      `确定以管理员身份进入租户「${row.name}」的后台吗？`,
      '模拟登录',
      { confirmButtonText: '确定进入', cancelButtonText: '取消', type: 'info' }
    )
    const res = await tenantApi.impersonateTenant(row.id)
    const { token, tenant_id } = res.data
    localStorage.setItem('fayhub_token', token)
    localStorage.setItem('fayhub_impersonated_tenant', String(tenant_id))
    window.location.replace('/')
  } catch (e) { console.error('handleImpersonate failed:', e); }
}

const detailVisible = ref(false)
const detailTenant = ref<Tenant | null>(null)
const detailQuota = ref<TenantQuota | null>(null)
const syncLoading = ref(false)
const quotaSaving = ref(false)
const quotaForm = reactive({
  max_users: 0,
  max_storage_mb: 0,
  max_plugins: 0,
  max_api_per_day: 0
})

function quotaPercent(used: number, max: number): number {
  if (!max || max <= 0) return 0
  return Math.min(Math.round((used / max) * 100), 100)
}

function progressColor(used: number, max: number): string {
  const pct = quotaPercent(used, max)
  if (pct >= 90) return '#f56c6c'
  if (pct >= 70) return '#e6a23c'
  return '#409eff'
}

async function openDetail(row: Tenant) {
  detailTenant.value = row
  detailVisible.value = true
  detailQuota.value = null
  try {
    const res = await tenantApi.getTenantQuota(row.id)
    detailQuota.value = res.data || null
    if (res.data) {
      quotaForm.max_users = res.data.max_users
      quotaForm.max_storage_mb = res.data.max_storage_mb
      quotaForm.max_plugins = res.data.max_plugins
      quotaForm.max_api_per_day = res.data.max_api_per_day
    }
  } catch (e) { console.error('fetchQuota failed:', e); }
}

async function syncUsage() {
  if (!detailTenant.value) return
  syncLoading.value = true
  try {
    const res = await tenantApi.syncTenantUsage(detailTenant.value.id)
    detailQuota.value = res.data || null
    ElMessage.success('用量同步完成')
  } catch (err: any) {
    ElMessage.error(err.message || '同步失败')
  } finally {
    syncLoading.value = false
  }
}

async function saveQuota() {
  if (!detailTenant.value) return
  quotaSaving.value = true
  try {
    await tenantApi.updateTenantQuota(detailTenant.value.id, quotaForm)
    ElMessage.success('配额更新成功')
    const res = await tenantApi.getTenantQuota(detailTenant.value.id)
    detailQuota.value = res.data || null
  } catch (err: any) {
    ElMessage.error(err.message || '更新失败')
  } finally {
    quotaSaving.value = false
  }
}
</script>
