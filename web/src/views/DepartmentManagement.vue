<template>
  <div class="dept-page">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold text-slate-800">部门管理</h2>
        <p class="text-slate-500 mt-1 text-sm">管理组织架构，配置部门与用户归属</p>
      </div>
      <el-button type="primary" @click="openCreateDialog(0)">
        <el-icon class="mr-1"><Plus /></el-icon> 新建顶级部门
      </el-button>
    </div>

    <div class="flex gap-6">
      <div class="bg-white rounded-2xl border border-slate-100 shadow-sm p-6 flex-1 min-w-0">
        <h3 class="text-lg font-semibold text-slate-700 mb-4">组织架构</h3>
        <el-tree
          v-loading="loading"
          :data="treeData"
          :props="{ label: 'name', children: 'children' }"
          node-key="id"
          default-expand-all
          :expand-on-click-node="false"
          :highlight-current="true"
          @node-click="handleNodeClick"
        >
          <template #default="{ node, data }">
            <div class="flex items-center justify-between w-full py-1 px-2">
              <div class="flex items-center gap-2">
                <el-icon class="text-amber-500"><Folder /></el-icon>
                <span class="font-medium text-slate-700">{{ (data as Department).name }}</span>
                <el-tag v-if="(data as Department).status === 0" type="info" size="small">停用</el-tag>
              </div>
              <div class="flex items-center gap-1">
                <el-button type="primary" link size="small" @click.stop="openCreateDialog((data as Department).id)">添加子部门</el-button>
                <el-button type="success" link size="small" @click.stop="openUserDialog(data as Department)">成员</el-button>
                <el-button type="warning" link size="small" @click.stop="openEditDialog(data as Department)">编辑</el-button>
                <el-button type="danger" link size="small" @click.stop="handleDelete(data as Department)">删除</el-button>
              </div>
            </div>
          </template>
        </el-tree>

        <el-empty v-if="!loading && treeData.length === 0" description="暂无部门数据" />
      </div>

      <div v-if="selectedDept" class="bg-white rounded-2xl border border-slate-100 shadow-sm p-6 w-[360px] shrink-0">
        <h3 class="text-lg font-semibold text-slate-700 mb-4">{{ selectedDept.name }} — 部门成员</h3>
        <div v-if="deptUsers.length > 0">
          <div v-for="u in deptUsers" :key="u.id" class="flex items-center justify-between py-2 border-b border-slate-50 last:border-0">
            <div class="flex items-center gap-2">
              <el-avatar :size="28" class="bg-blue-100 text-blue-600 text-xs">{{ (u.real_name || u.username).charAt(0) }}</el-avatar>
              <div>
                <p class="text-sm font-medium text-slate-700">{{ u.real_name || u.username }}</p>
                <p class="text-xs text-slate-400">{{ u.username }}</p>
              </div>
            </div>
            <el-button type="danger" link size="small" @click="handleRemoveUser(u.id)">移除</el-button>
          </div>
        </div>
        <el-empty v-else description="暂无成员" :image-size="60" />
        <el-button type="primary" plain class="w-full mt-4" @click="openUserDialog(selectedDept)">
          <el-icon class="mr-1"><Plus /></el-icon> 添加成员
        </el-button>
      </div>
    </div>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑部门' : '新建部门'" width="480px" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="部门名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入部门名称" />
        </el-form-item>
        <el-form-item label="上级部门" prop="parent_id">
          <el-tree-select
            v-model="form.parent_id"
            :data="parentTreeOptions"
            :props="{ label: 'name', children: 'children', value: 'id' }"
            placeholder="顶级部门"
            clearable
            check-strictly
            class="w-full"
          />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" :max="9999" />
        </el-form-item>
        <el-form-item v-if="isEdit" label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">停用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确认</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="userDialogVisible" :title="`添加成员到「${userDialogDept?.name || ''}」`" width="560px" :close-on-click-modal="false">
      <div v-loading="userDialogLoading">
        <el-input v-model="userSearchKeyword" placeholder="搜索用户名/姓名" clearable class="mb-4" @keyup.enter="fetchAvailableUsers" />

        <el-table :data="availableUsers" max-height="360" stripe size="small" @selection-change="handleUserSelect">
          <el-table-column type="selection" width="40" />
          <el-table-column prop="username" label="用户名" min-width="100" />
          <el-table-column prop="real_name" label="姓名" min-width="100" />
          <el-table-column prop="email" label="邮箱" min-width="140" />
        </el-table>

        <div class="mt-3 flex justify-end">
          <el-pagination
            v-model:current-page="userPagination.page"
            v-model:page-size="userPagination.page_size"
            :total="userPagination.total"
            :page-sizes="[10, 20]"
            layout="total, prev, pager, next"
            small
            @size-change="fetchAvailableUsers"
            @current-change="fetchAvailableUsers"
          />
        </div>
      </div>
      <template #footer>
        <el-button @click="userDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="userAssignLoading" :disabled="selectedUserIds.length === 0" @click="handleAssignUsers">
          确认添加（{{ selectedUserIds.length }}人）
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { Plus, Folder } from '@element-plus/icons-vue'
import deptApi, { type Department, type CreateDeptParams, type UpdateDeptParams } from '@/api/department'
import userApi, { type User } from '@/api/user'

const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref<number>(0)
const formRef = ref<FormInstance>()
const treeData = ref<Department[]>([])
const selectedDept = ref<Department | null>(null)
const deptUsers = ref<User[]>([])

const form = reactive<CreateDeptParams & { status?: number }>({
  name: '',
  parent_id: 0,
  sort: 0,
  status: 1
})

const formRules = {
  name: [{ required: true, message: '请输入部门名称', trigger: 'blur' }]
}

const parentTreeOptions = computed(() => {
  return filterDeptFromTree(treeData.value, isEdit.value ? editId.value : 0)
})

const userDialogVisible = ref(false)
const userDialogLoading = ref(false)
const userAssignLoading = ref(false)
const userDialogDept = ref<Department | null>(null)
const userSearchKeyword = ref('')
const availableUsers = ref<User[]>([])
const selectedUserIds = ref<number[]>([])
const userPagination = reactive({ page: 1, page_size: 10, total: 0 })

function filterDeptFromTree(nodes: Department[], excludeId: number): Department[] {
  return nodes
    .filter(n => n.id !== excludeId)
    .map(n => ({
      ...n,
      children: n.children ? filterDeptFromTree(n.children, excludeId) : []
    }))
}

onMounted(() => { fetchTree() })

async function fetchTree() {
  loading.value = true
  try {
    const res = await deptApi.getTree()
    treeData.value = res.data || []
  } catch (err: any) {
    ElMessage.error(err.message || '获取部门树失败')
  } finally {
    loading.value = false
  }
}

function handleNodeClick(data: Department) {
  selectedDept.value = data
  fetchDeptUsers(data.id)
}

async function fetchDeptUsers(deptId: number) {
  try {
    const res = await userApi.getUserList({ page: 1, page_size: 100 })
    const allUsers = res.data?.list || []
    deptUsers.value = allUsers.filter((u: User) => {
      return (u as any).dept_id === deptId
    })
  } catch {
    deptUsers.value = []
  }
}

async function handleRemoveUser(userId: number) {
  if (!selectedDept.value) return
  try {
    await ElMessageBox.confirm('确定将该用户从本部门移除？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deptApi.removeUser(userId, selectedDept.value.id)
    ElMessage.success('移除成功')
    fetchDeptUsers(selectedDept.value.id)
  } catch {}
}

function openCreateDialog(parentId: number) {
  isEdit.value = false
  editId.value = 0
  Object.assign(form, { name: '', parent_id: parentId || 0, sort: 0, status: 1 })
  dialogVisible.value = true
}

function openEditDialog(dept: Department) {
  isEdit.value = true
  editId.value = dept.id
  Object.assign(form, {
    name: dept.name,
    parent_id: dept.parent_id,
    sort: dept.sort,
    status: dept.status
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
      const params: UpdateDeptParams = {
        name: form.name,
        parent_id: form.parent_id || undefined,
        sort: form.sort,
        status: form.status
      }
      await deptApi.update(editId.value, params)
      ElMessage.success('部门更新成功')
    } else {
      const params: CreateDeptParams = {
        name: form.name,
        parent_id: form.parent_id || undefined,
        sort: form.sort
      }
      await deptApi.create(params)
      ElMessage.success('部门创建成功')
    }
    dialogVisible.value = false
    fetchTree()
  } catch (err: any) {
    ElMessage.error(err.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(dept: Department) {
  try {
    await ElMessageBox.confirm(`确定要删除部门「${dept.name}」吗？存在子部门时无法删除。`, '警告', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'error'
    })
    await deptApi.delete(dept.id)
    ElMessage.success('删除成功')
    if (selectedDept.value?.id === dept.id) {
      selectedDept.value = null
      deptUsers.value = []
    }
    fetchTree()
  } catch {}
}

function openUserDialog(dept: Department) {
  userDialogDept.value = dept
  userDialogVisible.value = true
  userSearchKeyword.value = ''
  selectedUserIds.value = []
  userPagination.page = 1
  fetchAvailableUsers()
}

async function fetchAvailableUsers() {
  userDialogLoading.value = true
  try {
    const res = await userApi.getUserList({
      page: userPagination.page,
      page_size: userPagination.page_size
    })
    availableUsers.value = res.data?.list || []
    userPagination.total = res.data?.total || 0
  } catch (err: any) {
    ElMessage.error(err.message || '获取用户列表失败')
  } finally {
    userDialogLoading.value = false
  }
}

function handleUserSelect(rows: User[]) {
  selectedUserIds.value = rows.map(r => r.id)
}

async function handleAssignUsers() {
  if (!userDialogDept.value || selectedUserIds.value.length === 0) return
  userAssignLoading.value = true
  try {
    const deptId = userDialogDept.value.id
    await Promise.all(selectedUserIds.value.map(userId => deptApi.assignUser(userId, deptId)))
    ElMessage.success(`成功添加 ${selectedUserIds.value.length} 名成员`)
    userDialogVisible.value = false
    if (selectedDept.value?.id === deptId) {
      fetchDeptUsers(deptId)
    }
  } catch (err: any) {
    ElMessage.error(err.message || '添加成员失败')
  } finally {
    userAssignLoading.value = false
  }
}
</script>
