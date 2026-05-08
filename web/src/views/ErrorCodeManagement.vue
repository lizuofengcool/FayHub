<template>
  <div class="error-code-page">
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 pb-3 flex items-center justify-between">
        <div>
          <h2 class="text-lg font-bold text-slate-800">错误码管理</h2>
          <p class="text-slate-400 text-xs mt-0.5">在线管理系统错误码，支持动态修改提示文案</p>
        </div>
        <div class="flex gap-2">
          <el-button @click="handleRefreshCache">
            <el-icon class="mr-1"><Refresh /></el-icon>
            刷新缓存
          </el-button>
          <el-button type="primary" @click="openDialog()">
            <el-icon class="mr-1"><Plus /></el-icon>
            新增错误码
          </el-button>
        </div>
      </div>

      <div class="p-6">
      <div class="p-4 border-b border-slate-100 flex gap-3 flex-wrap">
        <el-input v-model="filters.name" placeholder="错误码名称" clearable style="width: 180px" />
        <el-input v-model="filters.code" placeholder="错误码" clearable style="width: 120px" />
        <el-select v-model="filters.status" placeholder="状态" clearable style="width: 100px">
          <el-option label="启用" :value="1" />
          <el-option label="禁用" :value="0" />
        </el-select>
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>

      <el-table v-loading="loading" :data="codes" stripe class="w-full">
        <el-table-column prop="code" label="错误码" width="100" />
        <el-table-column prop="name" label="名称" width="180" show-overflow-tooltip />
        <el-table-column prop="msg" label="提示信息" min-width="250" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDialog(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="p-4 flex justify-end">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @current-change="fetchCodes"
          @size-change="fetchCodes"
        />
      </div>
      </div>
    </div>

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑错误码' : '新增错误码'" width="500px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="错误码" required>
          <el-input-number v-model="form.code" :min="40000" :max="59999" :disabled="!!form.id" style="width: 100%" />
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="如：用户不存在" />
        </el-form-item>
        <el-form-item label="提示信息" required>
          <el-input v-model="form.msg" type="textarea" :rows="2" placeholder="用户看到的提示文案" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSave">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh } from '@element-plus/icons-vue'
import errorCodeApi, { type ErrorCodeItem } from '@/api/errorCode'

const loading = ref(false)
const codes = ref<ErrorCodeItem[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const filters = reactive({
  name: '',
  code: '',
  status: undefined as number | undefined
})

const dialogVisible = ref(false)
const submitLoading = ref(false)
const form = reactive<Partial<ErrorCodeItem>>({
  id: undefined,
  code: undefined,
  name: '',
  msg: '',
  status: 1
})

async function fetchCodes() {
  loading.value = true
  try {
    const params: any = {
      page: page.value,
      page_size: pageSize.value,
      name: filters.name || undefined,
      status: filters.status
    }
    if (filters.code) {
      const codeNum = parseInt(filters.code)
      if (!isNaN(codeNum)) params.code = codeNum
    }
    const res = await errorCodeApi.list(params)
    codes.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (err: any) {
    ElMessage.error(err.message || '获取错误码失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  page.value = 1
  fetchCodes()
}

function resetFilters() {
  filters.name = ''
  filters.code = ''
  filters.status = undefined
  page.value = 1
  fetchCodes()
}

function openDialog(row?: ErrorCodeItem) {
  if (row) {
    Object.assign(form, { ...row })
  } else {
    Object.assign(form, { id: undefined, code: undefined, name: '', msg: '', status: 1 })
  }
  dialogVisible.value = true
}

async function handleSave() {
  submitLoading.value = true
  try {
    if (form.id) {
      await errorCodeApi.update(form.id, form)
    } else {
      await errorCodeApi.create(form)
    }
    ElMessage.success(form.id ? '更新成功' : '创建成功')
    dialogVisible.value = false
    fetchCodes()
  } catch (err: any) {
    ElMessage.error(err.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: ErrorCodeItem) {
  try {
    await ElMessageBox.confirm(`确定删除错误码 ${row.code}（${row.name}）？`, '确认删除', { type: 'warning' })
  } catch { return }
  try {
    await errorCodeApi.delete(row.id)
    ElMessage.success('删除成功')
    fetchCodes()
  } catch (err: any) {
    ElMessage.error(err.message || '删除失败')
  }
}

async function handleRefreshCache() {
  try {
    await errorCodeApi.refreshCache()
    ElMessage.success('缓存刷新成功')
  } catch (err: any) {
    ElMessage.error(err.message || '刷新缓存失败')
  }
}

onMounted(() => {
  fetchCodes()
})
</script>

<style scoped>
:deep(.el-input__wrapper) {
  height: 32px;
}

:deep(.el-select .el-input__wrapper) {
  height: 32px;
}

:deep(.el-input-number .el-input__wrapper) {
  height: 32px;
}

:deep(.el-button) {
  height: 32px;
  padding: 8px 12px;
}
</style>
