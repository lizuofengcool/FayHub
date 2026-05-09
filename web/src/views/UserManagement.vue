﻿﻿<template>
  <div class="user-page">
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 pb-3 flex items-center justify-between">
        <div>
          <h2 class="text-lg font-bold text-slate-800">用户管理</h2>
          <p class="text-slate-400 text-xs mt-0.5">管理当前租户下的所有用户</p>
        </div>
        <el-button type="default" @click="openCreateDialog">
          <el-icon class="mr-1"><Plus /></el-icon> 新建用户
        </el-button>
      </div>

      <div class="px-4 pb-4 border-b border-slate-100">
        <el-form :inline="true" :model="searchForm" class="search-form">
          <el-form-item label="关键词">
            <el-input v-model="searchForm.keyword" placeholder="用户名/昵称/手机号" clearable @keyup.enter="fetchList" />
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 120px">
              <el-option label="启用" :value="1" />
              <el-option label="禁用" :value="0" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="default" @click="fetchList">查询</el-button>
            <el-button @click="resetSearch">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <el-table v-loading="loading" :data="tableData" stripe class="w-full">
        <el-table-column prop="id" label="ID" width="100" :show-overflow-tooltip="true" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="real_name" label="昵称" min-width="120" />
        <el-table-column prop="phone" label="手机号" min-width="120" />
        <el-table-column prop="email" label="邮箱" min-width="160" />
        <el-table-column prop="status" label="状态" width="90" align="center">
          <template #default="{ row }">
            <n-tag :type="row.status === 1 ? 'success' : 'error'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </n-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" min-width="180" />
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button type="default" link size="small" @click="openEditDialog(row)">编辑</el-button>
            <el-button type="default" link size="small" @click="openAssignRoleDialog(row)">分配角色</el-button>
            <el-button
              :type="row.status === 1 ? 'warning' : 'success'"
              link
              size="small"
              @click="toggleStatus(row)"
            >
              {{ row.status === 1 ? '禁用' : '启用' }}
            </el-button>
            <el-button type="error" link size="small" @click="handleDelete(row)">删除</el-button>
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑用户' : '新建用户'" width="500px" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="请输入用户名" :disabled="isEdit" />
        </el-form-item>
        <el-form-item v-if="!isEdit" label="密码" prop="password">
          <el-input v-model="form.password" type="password" placeholder="请输入密码" show-password />
        </el-form-item>
        <el-form-item label="真实姓名" prop="real_name">
          <el-input v-model="form.real_name" placeholder="请输入真实姓名" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="form.phone" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入邮箱" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="default" :loading="submitLoading" @click="handleSubmit">确认</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="assignRoleVisible" title="分配角色" width="500px" :close-on-click-modal="false">
      <div v-loading="assignRoleLoading">
        <p class="mb-4 text-sm text-slate-500">为用户「{{ assignTargetUser?.username }}」分配角色：</p>
        <el-checkbox-group v-model="selectedRoleIds">
          <div v-for="role in allRoles" :key="role.id" class="mb-3">
            <el-checkbox :value="role.id">
              <span class="font-medium">{{ role.name }}</span>
              <span class="text-slate-400 text-xs ml-2">{{ role.type === 1 ? '平台角色' : '租户角色' }}</span>
            </el-checkbox>
          </div>
        </el-checkbox-group>
        <n-empty v-if="allRoles.length === 0" description="暂无可用角色" />
      </div>
      <template #footer>
        <el-button @click="assignRoleVisible = false">取消</el-button>
        <el-button type="default" :loading="assignSubmitLoading" @click="handleAssignRole">确认分配</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useMessage, useDialog } from 'naive-ui'
import type { FormInstance } from 'element-plus'
const message = useMessage()
const dialog = useDialog()
import { Plus } from '@element-plus/icons-vue'
import userApi, { type User, type CreateUserParams, type UpdateUserParams } from '@/api/user'
import rbacApi, { type Role } from '@/api/rbac'

const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref<number>(0)
const formRef = ref<FormInstance>()
const tableData = ref<User[]>([])

const searchForm = reactive({
  keyword: '',
  status: undefined as number | undefined
})

const pagination = reactive({
  page: 1,
  page_size: 10,
  total: 0
})

const form = reactive<CreateUserParams>({
  username: '',
  password: '',
  real_name: '',
  phone: '',
  email: '',
  status: 1
})

const formRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度3-20个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ]
}

const assignRoleVisible = ref(false)
const assignRoleLoading = ref(false)
const assignSubmitLoading = ref(false)
const assignTargetUser = ref<User | null>(null)
const allRoles = ref<Role[]>([])
const selectedRoleIds = ref<number[]>([])

onMounted(() => {
  fetchList()
})

async function fetchList() {
  loading.value = true
  try {
    const res = await userApi.getUserList({
      page: pagination.page,
      page_size: pagination.page_size,
      keyword: searchForm.keyword || undefined,
      status: searchForm.status
    })
    tableData.value = res.data.list || []
    pagination.total = res.data.total || 0
  } catch (err: any) {
    message.error(err.message || '获取用户列表失败')
  } finally {
    loading.value = false
  }
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
    username: '',
    password: '',
    real_name: '',
    phone: '',
    email: '',
    status: 1
  })
  dialogVisible.value = true
}

function openEditDialog(row: User) {
  isEdit.value = true
  editId.value = row.id
  Object.assign(form, {
    username: row.username,
    password: '',
    real_name: row.real_name,
    phone: row.phone,
    email: row.email,
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
      const params: UpdateUserParams = {
        real_name: form.real_name,
        phone: form.phone,
        email: form.email,
        status: form.status
      }
      await userApi.updateUser(editId.value, params)
      message.success('用户更新成功')
    } else {
      await userApi.createUser(form)
      message.success('用户创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (err: any) {
    message.error(err.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}

async function toggleStatus(row: User) {
  const newStatus = row.status === 1 ? 0 : 1
  const action = newStatus === 1 ? '启用' : '禁用'
  try {
    await dialog.warning({
      title: '提示',
      content: `确定要${action}用户「${row.username}」吗？`,
      positiveText: '确定',
      negativeText: '取消',
    })
    await userApi.updateUser(row.id, { status: newStatus })
    message.success(`${action}成功`)
    fetchList()
  } catch (e) { console.error('handleStatusChange failed:', e); }
}

async function handleDelete(row: User) {
  try {
    await dialog.warning({
      title: '警告',
      content: `确定要删除用户「${row.username}」吗？此操作不可恢复！`,
      positiveText: '确定删除',
      negativeText: '取消',
    })
    await userApi.deleteUser(row.id)
    message.success('删除成功')
    fetchList()
  } catch (e) { console.error('handleDelete failed:', e); }
}

async function openAssignRoleDialog(row: User) {
  assignTargetUser.value = row
  assignRoleVisible.value = true
  assignRoleLoading.value = true

  try {
    const [rolesRes, userRolesRes] = await Promise.all([
      rbacApi.getRoleList({ page: 1, page_size: 100 }),
      rbacApi.getUserRoles(row.id)
    ])
    allRoles.value = rolesRes.data.list || []
    selectedRoleIds.value = (userRolesRes.data || []).map((r: Role) => r.id)
  } catch (err: any) {
    message.error(err.message || '获取角色信息失败')
  } finally {
    assignRoleLoading.value = false
  }
}

async function handleAssignRole() {
  if (!assignTargetUser.value) return
  assignSubmitLoading.value = true

  try {
    const currentRolesRes = await rbacApi.getUserRoles(assignTargetUser.value.id)
    const currentRoleIds = (currentRolesRes.data || []).map((r: Role) => r.id)

    const toAdd = selectedRoleIds.value.filter(id => !currentRoleIds.includes(id))
    const toRemove = currentRoleIds.filter(id => !selectedRoleIds.value.includes(id))

    const promises = [
      ...toAdd.map(roleId => rbacApi.assignRoleToUser({ user_id: assignTargetUser.value!.id, role_id: roleId })),
      ...toRemove.map(roleId => rbacApi.removeRoleFromUser({ user_id: assignTargetUser.value!.id, role_id: roleId }))
    ]

    await Promise.all(promises)
    message.success('角色分配成功')
    assignRoleVisible.value = false
  } catch (err: any) {
    message.error(err.message || '角色分配失败')
  } finally {
    assignSubmitLoading.value = false
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
</style>
