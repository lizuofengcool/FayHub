<template>
  <div class="role-page">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold text-slate-800">角色权限</h2>
        <p class="text-slate-500 mt-1 text-sm">管理系统角色，分配菜单和API权限</p>
      </div>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon class="mr-1"><Plus /></el-icon> 新建角色
      </el-button>
    </div>

    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 border-b border-slate-100">
        <el-form :inline="true" :model="searchForm" class="flex items-center">
          <el-form-item label="关键词">
            <el-input v-model="searchForm.keyword" placeholder="角色名称/编码" clearable @keyup.enter="fetchList" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="fetchList">查询</el-button>
            <el-button @click="resetSearch">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <el-table v-loading="loading" :data="tableData" stripe class="w-full">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="角色名称" min-width="140" />
        <el-table-column prop="type" label="类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="(row as any).type === 1 ? 'danger' : 'primary'" size="small">
              {{ (row as any).type === 1 ? '平台' : '租户' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200" />
        <el-table-column prop="data_scope" label="数据范围" width="130" align="center">
          <template #default="{ row }">
            <el-tag :type="dataScopeTagType((row as any).data_scope)" size="small">{{ dataScopeLabel((row as any).data_scope) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" min-width="160" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEditDialog(row)">编辑</el-button>
            <el-button type="success" link size="small" @click="openPermissionDialog(row)">权限配置</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑角色' : '新建角色'" width="500px" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="角色类型" prop="type">
          <el-radio-group v-model="form.type">
            <el-radio :value="1">平台角色</el-radio>
            <el-radio :value="2">租户角色</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入角色描述" />
        </el-form-item>
        <el-form-item label="数据范围" prop="data_scope">
          <el-select v-model="form.data_scope" placeholder="请选择数据范围" class="w-full">
            <el-option :value="1" label="全部数据" />
            <el-option :value="2" label="本部门数据" />
            <el-option :value="3" label="本部门及下级" />
            <el-option :value="4" label="仅本人数据" />
            <el-option :value="5" label="自定义部门" />
          </el-select>
          <p class="text-xs text-slate-400 mt-1">控制角色可查看的数据范围，影响列表和详情查询</p>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确认</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="permissionVisible" title="权限配置" width="700px" :close-on-click-modal="false">
      <div v-loading="permissionLoading">
        <p class="mb-4 text-sm text-slate-500">为角色「{{ permissionTargetRole?.name }}」配置权限：</p>

        <el-tabs v-model="permissionTab">
          <el-tab-pane label="菜单权限" name="menu">
            <el-tree
              ref="menuTreeRef"
              :data="menuTree"
              show-checkbox
              node-key="id"
              :default-checked-keys="checkedMenuIds"
              :props="{ label: 'title', children: 'children' }"
              class="max-h-[400px] overflow-auto"
            />
            <el-empty v-if="menuTree.length === 0" description="暂无菜单数据" />
          </el-tab-pane>
          <el-tab-pane label="API权限" name="api">
            <el-checkbox-group v-model="checkedApiIds">
              <div v-for="api in allApis" :key="api.id" class="mb-2">
                <el-checkbox :value="api.id">
                  <el-tag size="small" :type="methodTagType(api.method)" class="mr-2">{{ api.method }}</el-tag>
                  <span class="font-mono text-sm">{{ api.path }}</span>
                  <span class="text-slate-400 text-xs ml-2">{{ api.description }}</span>
                </el-checkbox>
              </div>
            </el-checkbox-group>
            <el-empty v-if="allApis.length === 0" description="暂无API数据" />
          </el-tab-pane>
        </el-tabs>
      </div>
      <template #footer>
        <el-button @click="permissionVisible = false">取消</el-button>
        <el-button type="primary" :loading="permissionSubmitLoading" @click="handleSavePermission">保存权限</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import rbacApi, { type Role, type CreateRoleParams, type UpdateRoleParams } from '@/api/rbac'
import menuApi, { type Menu } from '@/api/menu'
import apiApi, { type ApiItem } from '@/api/api'

const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref<number>(0)
const formRef = ref<FormInstance>()
const tableData = ref<Role[]>([])

const searchForm = reactive({ keyword: '' })

const pagination = reactive({
  page: 1,
  page_size: 10,
  total: 0
})

const form = reactive<CreateRoleParams>({
  name: '',
  type: 2,
  description: '',
  data_scope: 1
})

const formRules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择角色类型', trigger: 'change' }]
}

const permissionVisible = ref(false)
const permissionLoading = ref(false)
const permissionSubmitLoading = ref(false)
const permissionTargetRole = ref<Role | null>(null)
const permissionTab = ref('menu')

const menuTreeRef = ref()
const menuTree = ref<Menu[]>([])
const checkedMenuIds = ref<number[]>([])

const allApis = ref<ApiItem[]>([])
const checkedApiIds = ref<number[]>([])

function methodTagType(method: string) {
  const map: Record<string, string> = { GET: 'success', POST: 'primary', PUT: 'warning', DELETE: 'danger', PATCH: 'info' }
  return map[method.toUpperCase()] || 'info'
}

function dataScopeLabel(scope: number) {
  const map: Record<number, string> = { 1: '全部数据', 2: '本部门', 3: '本部门及下级', 4: '仅本人', 5: '自定义' }
  return map[scope] || '全部数据'
}

function dataScopeTagType(scope: number) {
  const map: Record<number, string> = { 1: '', 2: 'warning', 3: 'success', 4: 'info', 5: 'danger' }
  return map[scope] || ''
}

onMounted(() => { fetchList() })

async function fetchList() {
  loading.value = true
  try {
    const res = await rbacApi.getRoleList({
      page: pagination.page,
      page_size: pagination.page_size,
      keyword: searchForm.keyword || undefined
    })
    tableData.value = res.data.list || []
    pagination.total = res.data.total || 0
  } catch (err: any) {
    ElMessage.error(err.message || '获取角色列表失败')
  } finally {
    loading.value = false
  }
}

function resetSearch() {
  searchForm.keyword = ''
  pagination.page = 1
  fetchList()
}

function openCreateDialog() {
  isEdit.value = false
  editId.value = 0
  Object.assign(form, { name: '', type: 2, description: '', data_scope: 1 })
  dialogVisible.value = true
}

function openEditDialog(row: Role) {
  isEdit.value = true
  editId.value = row.id
  Object.assign(form, { name: row.name, type: (row as any).type || 2, description: row.description, data_scope: (row as any).data_scope || 1 })
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (isEdit.value) {
      const params: UpdateRoleParams = { name: form.name, description: form.description, data_scope: form.data_scope }
      await rbacApi.updateRole(editId.value, params)
      ElMessage.success('角色更新成功')
    } else {
      await rbacApi.createRole(form)
      ElMessage.success('角色创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (err: any) {
    ElMessage.error(err.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: Role) {
  try {
    await ElMessageBox.confirm(`确定要删除角色「${row.name}」吗？`, '警告', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'error'
    })
    await rbacApi.deleteRole(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e) { console.error('handleDelete failed:', e); }
}

async function openPermissionDialog(row: Role) {
  permissionTargetRole.value = row
  permissionVisible.value = true
  permissionLoading.value = true
  permissionTab.value = 'menu'

  try {
    const [menuTreeRes, roleMenusRes, apiListRes, roleApisRes] = await Promise.all([
      menuApi.getMenuTree(),
      menuApi.getRoleMenus(row.id),
      apiApi.getApiList({ page: 1, page_size: 500 }),
      apiApi.getRoleApis(row.id)
    ])

    menuTree.value = menuTreeRes.data || []
    checkedMenuIds.value = (roleMenusRes.data || []).map((m: Menu) => m.id)
    allApis.value = (apiListRes.data?.list || apiListRes.data || []) as ApiItem[]
    checkedApiIds.value = (roleApisRes.data || []).map((a: ApiItem) => a.id)
  } catch (err: any) {
    ElMessage.error(err.message || '获取权限数据失败')
  } finally {
    permissionLoading.value = false
  }
}

async function handleSavePermission() {
  if (!permissionTargetRole.value) return
  permissionSubmitLoading.value = true

  try {
    const roleId = permissionTargetRole.value.id
    const menuIds = menuTreeRef.value?.getCheckedKeys() || []
    await Promise.all([
      menuApi.assignRoleMenus({ role_id: roleId, menu_ids: menuIds }),
      apiApi.assignRoleApis({ role_id: roleId, api_ids: checkedApiIds.value })
    ])
    ElMessage.success('权限保存成功')
    permissionVisible.value = false
  } catch (err: any) {
    ElMessage.error(err.message || '权限保存失败')
  } finally {
    permissionSubmitLoading.value = false
  }
}
</script>
