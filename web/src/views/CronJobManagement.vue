<template>
  <div class="cron-job-page">
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 pb-3 flex items-center justify-between">
        <div>
          <h2 class="text-lg font-bold text-slate-800">定时任务</h2>
          <p class="text-slate-400 text-xs mt-0.5">管理系统定时任务与调度策略</p>
        </div>
        <el-button type="primary" @click="openCreateDialog">
          <el-icon class="mr-1"><Plus /></el-icon>
          新增任务
        </el-button>
      </div>

      <el-table v-loading="loading" :data="jobs" stripe class="w-full">
        <el-table-column prop="name" label="任务名称" width="160" />
        <el-table-column prop="cron_expr" label="Cron表达式" width="150" />
        <el-table-column prop="command" label="执行命令" min-width="200" show-overflow-tooltip />
        <el-table-column prop="description" label="描述" width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <span>{{ row.description || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-switch
              :model-value="row.status === 1"
              @change="handleToggleStatus(row)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="last_run_at" label="上次执行" width="170">
          <template #default="{ row }">
            <span>{{ formatTime(row.last_run_at) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" align="center" fixed="right">
          <template #default="{ row }">
            <el-button size="small" link type="primary" @click="handleExecute(row)">
              执行
            </el-button>
            <el-button size="small" link type="primary" @click="openLogDialog(row)">
              日志
            </el-button>
            <el-button size="small" link type="primary" @click="openEditDialog(row)">
              编辑
            </el-button>
            <el-popconfirm
              title="确定删除该任务？"
              @confirm="handleDelete(row)"
            >
              <template #reference>
                <el-button size="small" link type="danger">删除</el-button>
              </template>
            </el-popconfirm>
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
          @current-change="fetchJobs"
          @size-change="fetchJobs"
        />
      </div>
    </div>

    <el-dialog v-model="formVisible" :title="isEdit ? '编辑任务' : '新增任务'" width="520px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入任务名称" />
        </el-form-item>
        <el-form-item label="Cron表达式" prop="cron_expr">
          <el-input v-model="form.cron_expr" placeholder="如: 0 0 2 * * * (每天凌晨2点)" />
        </el-form-item>
        <el-form-item label="执行命令" prop="command">
          <el-input v-model="form.command" placeholder="如: cleanup:logs" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="可选" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="logVisible" title="执行日志" width="700px">
      <el-table v-loading="logLoading" :data="logs" stripe max-height="400">
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'success' ? 'success' : row.status === 'running' ? 'warning' : 'danger'" size="small">
              {{ row.status === 'success' ? '成功' : row.status === 'running' ? '执行中' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="output" label="输出" min-width="200" show-overflow-tooltip />
        <el-table-column prop="started_at" label="开始时间" width="170">
          <template #default="{ row }">
            <span>{{ formatTime(row.started_at) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="ended_at" label="结束时间" width="170">
          <template #default="{ row }">
            <span>{{ formatTime(row.ended_at) }}</span>
          </template>
        </el-table-column>
      </el-table>
      <div class="mt-4 flex justify-end">
        <el-pagination
          v-model:current-page="logPage"
          v-model:page-size="logPageSize"
          :total="logTotal"
          :page-sizes="[10, 20]"
          layout="total, prev, pager, next"
          small
          @current-change="fetchLogs"
          @size-change="fetchLogs"
        />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import cronJobApi, { type CronJob, type CronJobLog } from '@/api/cronJob'

const loading = ref(false)
const jobs = ref<CronJob[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const formVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({
  id: 0,
  name: '',
  cron_expr: '',
  command: '',
  description: ''
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  cron_expr: [{ required: true, message: '请输入Cron表达式', trigger: 'blur' }],
  command: [{ required: true, message: '请输入执行命令', trigger: 'blur' }]
}

const logVisible = ref(false)
const logLoading = ref(false)
const logs = ref<CronJobLog[]>([])
const logPage = ref(1)
const logPageSize = ref(10)
const logTotal = ref(0)
const currentJobId = ref(0)

function formatTime(time: string | null) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

async function fetchJobs() {
  loading.value = true
  try {
    const res = await cronJobApi.list(page.value, pageSize.value)
    jobs.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch {
    // ignore
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  isEdit.value = false
  form.id = 0
  form.name = ''
  form.cron_expr = ''
  form.command = ''
  form.description = ''
  formVisible.value = true
}

function openEditDialog(row: CronJob) {
  isEdit.value = true
  form.id = row.id
  form.name = row.name
  form.cron_expr = row.cron_expr
  form.command = row.command
  form.description = row.description || ''
  formVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  try {
    if (isEdit.value) {
      await cronJobApi.update(form.id, form)
      ElMessage.success('更新成功')
    } else {
      await cronJobApi.create(form)
      ElMessage.success('创建成功')
    }
    formVisible.value = false
    fetchJobs()
  } catch {
    ElMessage.error('操作失败')
  }
}

async function handleDelete(row: CronJob) {
  try {
    await cronJobApi.delete(row.id)
    ElMessage.success('删除成功')
    fetchJobs()
  } catch {
    ElMessage.error('删除失败')
  }
}

async function handleToggleStatus(row: CronJob) {
  try {
    await cronJobApi.toggleStatus(row.id)
    ElMessage.success(row.status === 1 ? '已停用' : '已启用')
    fetchJobs()
  } catch {
    ElMessage.error('操作失败')
  }
}

async function handleExecute(row: CronJob) {
  try {
    await cronJobApi.executeOnce(row.id)
    ElMessage.success('已触发执行')
  } catch {
    ElMessage.error('执行失败')
  }
}

function openLogDialog(row: CronJob) {
  currentJobId.value = row.id
  logPage.value = 1
  logVisible.value = true
  fetchLogs()
}

async function fetchLogs() {
  logLoading.value = true
  try {
    const res = await cronJobApi.getLogs(currentJobId.value, logPage.value, logPageSize.value)
    logs.value = res.data?.list || []
    logTotal.value = res.data?.total || 0
  } catch {
    // ignore
  } finally {
    logLoading.value = false
  }
}

onMounted(() => {
  fetchJobs()
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
