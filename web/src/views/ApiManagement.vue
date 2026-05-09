﻿<template>
  <div class="api-page">
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 pb-3 flex items-center justify-between">
        <div>
          <h2 class="text-lg font-bold text-slate-800">API管理</h2>
          <p class="text-slate-400 text-xs mt-0.5">管理系统API接口权限，用于RBAC权限校验</p>
        </div>
        <el-button type="default" @click="openCreateDialog">
          <el-icon class="mr-1"><Plus /></el-icon> 新建API
        </el-button>
      </div>

      <div class="px-4 pb-4 border-b border-slate-100">
        <el-form :inline="true" :model="searchForm" class="search-form">
          <el-form-item label="关键词">
            <el-input v-model="searchForm.keyword" placeholder="API名称/路径" clearable @keyup.enter="fetchList" />
          </el-form-item>
          <el-form-item label="请求方法">
            <el-select v-model="searchForm.method" placeholder="全部" clearable style="width: 120px">
              <el-option label="GET" value="GET" />
              <el-option label="POST" value="POST" />
              <el-option label="PUT" value="PUT" />
              <el-option label="DELETE" value="DELETE" />
              <el-option label="PATCH" value="PATCH" />
            </el-select>
          </el-form-item>
          <el-form-item label="分组">
            <el-input v-model="searchForm.group" placeholder="API分组" clearable style="width: 120px" />
          </el-form-item>
          <el-form-item>
            <el-button type="default" @click="fetchList">查询</el-button>
            <el-button @click="resetSearch">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <el-table v-loading="loading" :data="tableData" stripe class="w-full">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="method" label="请求方法" width="110" align="center">
          <template #default="{ row }">
            <n-tag size="small" :type="methodTagType(row.method)">{{ row.method }}</n-tag>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路径" min-width="200">
          <template #default="{ row }">
            <span class="font-mono text-sm">{{ row.path }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="group" label="分组" width="120" />
        <el-table-column prop="description" label="描述" min-width="160" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="default" link size="small" @click="openEditDialog(row)">编辑</el-button>
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑API' : '新建API'" width="560px" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-form-item label="请求方法" prop="method">
          <el-select v-model="form.method" placeholder="请选择" style="width: 100%" :disabled="isEdit">
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="DELETE" value="DELETE" />
            <el-option label="PATCH" value="PATCH" />
          </el-select>
        </el-form-item>
        <el-form-item label="路径" prop="path">
          <el-input v-model="form.path" placeholder="如 /api/users" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="分组" prop="group">
          <el-input v-model="form.group" placeholder="如：用户管理" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="请输入描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="default" :loading="submitLoading" @click="handleSubmit">确认</el-button>
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
import apiApi, { type ApiItem, type CreateApiParams, type UpdateApiParams } from '@/api/api'

const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref<number>(0)
const formRef = ref<FormInstance>()
const tableData = ref<ApiItem[]>([])

const searchForm = reactive({
  keyword: '',
  method: undefined as string | undefined,
  group: ''
})

const pagination = reactive({
  page: 1,
  page_size: 10,
  total: 0
})

const form = reactive<CreateApiParams>({
  path: '',
  method: '',
  group: '',
  description: ''
})

const formRules = {
  method: [{ required: true, message: '请选择请求方法', trigger: 'change' }],
  path: [{ required: true, message: '请输入路径', trigger: 'blur' }]
}

function methodTagType(method: string) {
  const map: Record<string, string> = { GET: 'success', POST: 'primary', PUT: 'warning', DELETE: 'danger', PATCH: 'info' }
  return map[method.toUpperCase()] || 'info'
}

onMounted(() => { fetchList() })

async function fetchList() {
  loading.value = true
  try {
    const res = await apiApi.getApiList({
      page: pagination.page,
      page_size: pagination.page_size,
      keyword: searchForm.keyword || undefined,
      method: searchForm.method,
      group: searchForm.group || undefined
    })
    tableData.value = res.data.list || []
    pagination.total = res.data.total || 0
  } catch (err: any) {
    message.error(err.message || '获取API列表失败')
  } finally {
    loading.value = false
  }
}

function resetSearch() {
  searchForm.keyword = ''
  searchForm.method = undefined
  searchForm.group = ''
  pagination.page = 1
  fetchList()
}

function openCreateDialog() {
  isEdit.value = false
  editId.value = 0
  Object.assign(form, { path: '', method: '', group: '', description: '' })
  dialogVisible.value = true
}

function openEditDialog(row: ApiItem) {
  isEdit.value = true
  editId.value = row.id
  Object.assign(form, { path: row.path, method: row.method, group: row.group, description: row.description })
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (isEdit.value) {
      const params: UpdateApiParams = { group: form.group, description: form.description }
      await apiApi.updateApi(editId.value, params)
      message.success('API更新成功')
    } else {
      await apiApi.createApi(form)
      message.success('API创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (err: any) {
    message.error(err.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: ApiItem) {
  try {
    await dialog.warning(`确定要删除API「${row.path}」吗？`, '警告', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'error'
    })
    await apiApi.deleteApi(row.id)
    message.success('删除成功')
    fetchList()
  } catch (e) { console.error('handleDelete failed:', e); }
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
