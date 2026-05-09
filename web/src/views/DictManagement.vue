﻿<template>
  <div class="dict-page">
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 pb-3 flex items-center justify-between">
        <div>
          <h2 class="text-lg font-bold text-slate-800">字典管理</h2>
          <p class="text-slate-400 text-xs mt-0.5">管理系统枚举值与下拉选项</p>
        </div>
        <el-button type="default" @click="openTypeDialog()">
          <el-icon class="mr-1"><Plus /></el-icon>
          新增字典类型
        </el-button>
      </div>

      <div class="p-6">
        <div class="mb-4">
          <el-input v-model="typeFilters.dict_name" placeholder="搜索字典名称" clearable @input="fetchTypes" />
        </div>
        <el-table v-loading="typeLoading" :data="types" stripe highlight-current-row @current-change="selectType" class="w-full">
          <el-table-column prop="dict_name" label="字典名称" min-width="120" show-overflow-tooltip />
          <el-table-column prop="dict_type" label="字典类型" min-width="120" show-overflow-tooltip />
          <el-table-column prop="status" label="状态" width="70" align="center">
            <template #default="{ row }">
              <n-tag :type="row.status === 1 ? 'success' : 'error'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</n-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100" fixed="right">
            <template #default="{ row }">
              <el-button type="default" link size="small" @click.stop="openTypeDialog(row)">编辑</el-button>
              <el-button type="error" link size="small" @click.stop="handleDeleteType(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="p-3 flex justify-end">
          <el-pagination
            v-model:current-page="typePage"
            :total="typeTotal"
            :page-size="20"
            layout="total, prev, next"
            @current-change="fetchTypes"
          />
        </div>
      </div>

      <div class="lg:col-span-2 bg-white rounded-2xl border border-slate-100 shadow-sm">
        <div class="p-4 border-b border-slate-100 flex items-center justify-between">
          <span class="font-medium text-slate-700">{{ currentType ? currentType.dict_name + ' - 字典数据' : '请选择左侧字典类型' }}</span>
          <el-button v-if="currentType" type="default" size="small" @click="openDataDialog()">
            <el-icon class="mr-1"><Plus /></el-icon>
            新增
          </el-button>
        </div>
        <el-table v-loading="dataLoading" :data="dataList" stripe class="w-full">
          <el-table-column prop="dict_label" label="标签" min-width="120" />
          <el-table-column prop="dict_value" label="值" min-width="100" />
          <el-table-column prop="sort" label="排序" width="70" align="center" />
          <el-table-column prop="status" label="状态" width="70" align="center">
            <template #default="{ row }">
              <n-tag :type="row.status === 1 ? 'success' : 'error'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</n-tag>
            </template>
          </el-table-column>
          <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip />
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button type="default" link size="small" @click="openDataDialog(row)">编辑</el-button>
              <el-button type="error" link size="small" @click="handleDeleteData(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="p-3 flex justify-end">
          <el-pagination
            v-model:current-page="dataPage"
            :total="dataTotal"
            :page-size="20"
            layout="total, prev, next"
            @current-change="fetchData"
          />
        </div>
      </div>
    </div>

    <el-dialog v-model="typeDialogVisible" :title="typeForm.id ? '编辑字典类型' : '新增字典类型'" width="500px">
      <el-form :model="typeForm" label-width="100px">
        <el-form-item label="字典名称" required>
          <el-input v-model="typeForm.dict_name" placeholder="请输入字典名称" />
        </el-form-item>
        <el-form-item label="字典类型" required>
          <el-input v-model="typeForm.dict_type" placeholder="请输入字典类型（英文）" :disabled="!!typeForm.id" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="typeForm.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="typeForm.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="typeDialogVisible = false">取消</el-button>
        <el-button type="default" :loading="submitLoading" @click="handleSaveType">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="dataDialogVisible" :title="dataForm.id ? '编辑字典数据' : '新增字典数据'" width="500px">
      <el-form :model="dataForm" label-width="100px">
        <el-form-item label="字典类型">
          <el-input :model-value="currentType?.dict_type" disabled />
        </el-form-item>
        <el-form-item label="标签" required>
          <el-input v-model="dataForm.dict_label" placeholder="请输入显示标签" />
        </el-form-item>
        <el-form-item label="值" required>
          <el-input v-model="dataForm.dict_value" placeholder="请输入实际值" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="dataForm.sort" :min="0" />
        </el-form-item>
        <el-form-item label="样式类型">
          <el-input v-model="dataForm.list_class" placeholder="如: default/primary/success/warning/danger" />
        </el-form-item>
        <el-form-item label="是否默认">
          <el-radio-group v-model="dataForm.is_default">
            <el-radio :value="1">是</el-radio>
            <el-radio :value="0">否</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="dataForm.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="dataForm.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dataDialogVisible = false">取消</el-button>
        <el-button type="default" :loading="submitLoading" @click="handleSaveData">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import { Plus } from '@element-plus/icons-vue'
import dictApi, { type DictType, type DictData } from '@/api/dict'

const typeLoading = ref(false)
const types = ref<DictType[]>([])
const typePage = ref(1)
const typeTotal = ref(0)
const typeFilters = reactive({ dict_name: '' })

const dataLoading = ref(false)
const dataList = ref<DictData[]>([])
const dataPage = ref(1)
const dataTotal = ref(0)

const currentType = ref<DictType | null>(null)

const typeDialogVisible = ref(false)
const dataDialogVisible = ref(false)
const submitLoading = ref(false)

const typeForm = reactive<Partial<DictType>>({
  id: undefined,
  dict_name: '',
  dict_type: '',
  status: 1,
  remark: ''
})

const dataForm = reactive<Partial<DictData>>({
  id: undefined,
  dict_type: '',
  dict_label: '',
  dict_value: '',
  sort: 0,
  list_class: '',
  is_default: 0,
  status: 1,
  remark: ''
})

async function fetchTypes() {
  typeLoading.value = true
  try {
    const res = await dictApi.listTypes({
      page: typePage.value,
      page_size: 20,
      dict_name: typeFilters.dict_name || undefined
    })
    types.value = res.data?.list || []
    typeTotal.value = res.data?.total || 0
  } catch (err: any) {
    message.error(err.message || '获取字典类型失败')
  } finally {
    typeLoading.value = false
  }
}

async function fetchData() {
  if (!currentType.value) return
  dataLoading.value = true
  try {
    const res = await dictApi.listData({
      page: dataPage.value,
      page_size: 20,
      dict_type: currentType.value.dict_type
    })
    dataList.value = res.data?.list || []
    dataTotal.value = res.data?.total || 0
  } catch (err: any) {
    message.error(err.message || '获取字典数据失败')
  } finally {
    dataLoading.value = false
  }
}

function selectType(row: DictType | null) {
  currentType.value = row
  dataPage.value = 1
  if (row) fetchData()
  else dataList.value = []
}

function openTypeDialog(row?: DictType) {
  if (row) {
    Object.assign(typeForm, { ...row })
  } else {
    Object.assign(typeForm, { id: undefined, dict_name: '', dict_type: '', status: 1, remark: '' })
  }
  typeDialogVisible.value = true
}

function openDataDialog(row?: DictData) {
  if (row) {
    Object.assign(dataForm, { ...row })
  } else {
    Object.assign(dataForm, { id: undefined, dict_type: currentType.value?.dict_type, dict_label: '', dict_value: '', sort: 0, list_class: '', is_default: 0, status: 1, remark: '' })
  }
  dataDialogVisible.value = true
}

async function handleSaveType() {
  submitLoading.value = true
  try {
    if (typeForm.id) {
      await dictApi.updateType(typeForm.id, typeForm)
    } else {
      await dictApi.createType(typeForm)
    }
    message.success(typeForm.id ? '更新成功' : '创建成功')
    typeDialogVisible.value = false
    fetchTypes()
  } catch (err: any) {
    message.error(err.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}

async function handleSaveData() {
  submitLoading.value = true
  try {
    dataForm.dict_type = currentType.value?.dict_type
    if (dataForm.id) {
      await dictApi.updateData(dataForm.id, dataForm)
    } else {
      await dictApi.createData(dataForm)
    }
    message.success(dataForm.id ? '更新成功' : '创建成功')
    dataDialogVisible.value = false
    fetchData()
  } catch (err: any) {
    message.error(err.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}

async function handleDeleteType(row: DictType) {
  try {
    await dialog.warning(`确定删除字典类型"${row.dict_name}"？关联的字典数据也会被删除。`, '确认删除', { type: 'warning' })
  } catch { return }
  try {
    await dictApi.deleteType(row.id)
    message.success('删除成功')
    if (currentType.value?.id === row.id) {
      currentType.value = null
      dataList.value = []
    }
    fetchTypes()
  } catch (err: any) {
    message.error(err.message || '删除失败')
  }
}

async function handleDeleteData(row: DictData) {
  try {
    await dialog.warning(`确定删除字典数据"${row.dict_label}"？`, '确认删除', { type: 'warning' })
  } catch { return }
  try {
    await dictApi.deleteData(row.id)
    message.success('删除成功')
    fetchData()
  } catch (err: any) {
    message.error(err.message || '删除失败')
  }
}

onMounted(() => {
  fetchTypes()
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
</style>
