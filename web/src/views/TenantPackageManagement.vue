<template>
  <div class="package-page">
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 pb-3 flex items-center justify-between">
        <div>
          <h2 class="text-lg font-bold text-slate-800">套餐管理</h2>
          <p class="text-slate-400 text-xs mt-0.5">管理租户套餐，配置套餐配额和菜单权限</p>
        </div>
        <el-button type="primary" @click="openCreateDialog">
          <el-icon class="mr-1"><Plus /></el-icon> 新建套餐
        </el-button>
      </div>

      <div class="px-4 pb-4 border-b border-slate-100">
        <el-form :inline="true" :model="searchForm" class="search-form">
          <el-form-item label="套餐名称">
            <el-input v-model="searchForm.name" placeholder="请输入套餐名称" clearable @keyup.enter="fetchList" />
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="searchForm.status" placeholder="全部" clearable class="w-[120px]">
              <el-option :value="1" label="启用" />
              <el-option :value="0" label="禁用" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="fetchList">查询</el-button>
            <el-button @click="resetSearch">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <el-table v-loading="loading" :data="tableData" stripe class="w-full">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="套餐名称" min-width="140" />
        <el-table-column prop="code" label="套餐编码" min-width="120" />
        <el-table-column prop="max_users" label="最大用户数" width="110" align="center" />
        <el-table-column prop="max_storage_mb" label="存储(MB)" width="100" align="center" />
        <el-table-column prop="max_plugins" label="插件数" width="90" align="center" />
        <el-table-column prop="sort" label="排序" width="70" align="center" />
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="(row as any).status === 1 ? 'success' : 'danger'" size="small">
              {{ (row as any).status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" min-width="160" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEditDialog(row)">编辑</el-button>
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑套餐' : '新建套餐'" width="600px" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="套餐名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入套餐名称" />
        </el-form-item>
        <el-form-item label="套餐编码" prop="code">
          <el-input v-model="form.code" placeholder="请输入套餐编码" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" :max="999" />
        </el-form-item>
        <el-form-item label="最大用户数" prop="max_users">
          <el-input-number v-model="form.max_users" :min="1" :max="99999" />
        </el-form-item>
        <el-form-item label="存储配额(MB)" prop="max_storage_mb">
          <el-input-number v-model="form.max_storage_mb" :min="1" :max="999999" />
        </el-form-item>
        <el-form-item label="最大插件数" prop="max_plugins">
          <el-input-number v-model="form.max_plugins" :min="0" :max="999" />
        </el-form-item>
        <el-form-item label="菜单权限">
          <el-tree
            ref="menuTreeRef"
            :data="menuTree"
            show-checkbox
            node-key="id"
            :default-checked-keys="form.menu_ids"
            :props="{ label: 'title', children: 'children' }"
            class="max-h-[300px] overflow-auto border rounded p-2"
          />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="请输入备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  getTenantPackageList,
  getTenantPackage,
  createTenantPackage,
  updateTenantPackage,
  deleteTenantPackage,
  type TenantPackage,
  type TenantPackageForm
} from '@/api/tenantPackage'
import menuApi, { type Menu } from '@/api/menu'

const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref<number>(0)
const formRef = ref<FormInstance>()
const menuTreeRef = ref()
const tableData = ref<TenantPackage[]>([])
const menuTree = ref<Menu[]>([])

const searchForm = reactive({ name: '', status: undefined as number | undefined })

const pagination = reactive({
  page: 1,
  page_size: 10,
  total: 0
})

const form = reactive<TenantPackageForm>({
  name: '',
  code: '',
  status: 1,
  sort: 0,
  remark: '',
  max_users: 10,
  max_storage_mb: 1024,
  max_plugins: 5,
  menu_ids: []
})

const formRules = {
  name: [{ required: true, message: '请输入套餐名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入套餐编码', trigger: 'blur' }]
}

onMounted(() => {
  fetchList()
  fetchMenuTree()
})

async function fetchList() {
  loading.value = true
  try {
    const res = await getTenantPackageList({
      page: pagination.page,
      page_size: pagination.page_size,
      name: searchForm.name || undefined,
      status: searchForm.status
    })
    tableData.value = res.data.list || []
    pagination.total = res.data.total || 0
  } catch (err: any) {
    ElMessage.error(err.message || '获取套餐列表失败')
  } finally {
    loading.value = false
  }
}

async function fetchMenuTree() {
  try {
    const res = await menuApi.getMenuTree()
    menuTree.value = res.data || []
  } catch {
    // ignore
  }
}

function resetSearch() {
  searchForm.name = ''
  searchForm.status = undefined
  pagination.page = 1
  fetchList()
}

function openCreateDialog() {
  isEdit.value = false
  editId.value = 0
  Object.assign(form, {
    name: '',
    code: '',
    status: 1,
    sort: 0,
    remark: '',
    max_users: 10,
    max_storage_mb: 1024,
    max_plugins: 5,
    menu_ids: []
  })
  dialogVisible.value = true
}

async function openEditDialog(row: TenantPackage) {
  isEdit.value = true
  editId.value = row.id
  try {
    const res = await getTenantPackage(row.id)
    const pkg = res.data.package
    Object.assign(form, {
      name: pkg.name,
      code: pkg.code,
      status: pkg.status,
      sort: pkg.sort,
      remark: pkg.remark,
      max_users: pkg.max_users,
      max_storage_mb: pkg.max_storage_mb,
      max_plugins: pkg.max_plugins,
      menu_ids: res.data.menu_ids || []
    })
    dialogVisible.value = true
  } catch (err: any) {
    ElMessage.error(err.message || '获取套餐详情失败')
  }
}

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return

    const checkedKeys = menuTreeRef.value?.getCheckedKeys() || []
    const halfCheckedKeys = menuTreeRef.value?.getHalfCheckedKeys() || []
    form.menu_ids = [...checkedKeys, ...halfCheckedKeys]

    submitLoading.value = true
    try {
      if (isEdit.value) {
        await updateTenantPackage(editId.value, { ...form })
        ElMessage.success('更新成功')
      } else {
        await createTenantPackage({ ...form })
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      fetchList()
    } catch (err: any) {
      ElMessage.error(err.message || '操作失败')
    } finally {
      submitLoading.value = false
    }
  })
}

async function handleDelete(row: TenantPackage) {
  try {
    await ElMessageBox.confirm(`确认删除套餐「${row.name}」？`, '删除确认', {
      type: 'warning'
    })
    await deleteTenantPackage(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch {
    // cancelled
  }
}
</script>

<style scoped>
.search-form {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 0;
}

.search-form :deep(.el-form-item) {
  margin-bottom: 0;
  margin-right: 16px;
}

.search-form :deep(.el-form-item__label) {
  height: 32px;
  line-height: 32px;
  margin-bottom: 0;
}

.search-form :deep(.el-input__wrapper) {
  height: 32px;
}

.search-form :deep(.el-select .el-input__wrapper) {
  height: 32px;
}

.search-form :deep(.el-button) {
  height: 32px;
  padding: 8px 12px;
}

:deep(.el-input__wrapper) {
  height: 32px;
}

:deep(.el-select .el-input__wrapper) {
  height: 32px;
}

:deep(.el-input-number .el-input__wrapper) {
  height: 32px;
}
</style>
