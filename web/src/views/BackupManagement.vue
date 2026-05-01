<template>
  <div class="p-6 space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-2xl font-bold text-slate-800">备份管理</h2>
        <p class="text-sm text-slate-500 mt-1">管理数据库备份，支持创建、下载、恢复和删除</p>
      </div>
      <div class="flex items-center gap-3">
        <el-button @click="showRestoreDialog = true">
          <el-icon class="mr-1"><Upload /></el-icon> 恢复备份
        </el-button>
        <el-button type="primary" @click="createBackup" :loading="creating">
          <el-icon class="mr-1"><Plus /></el-icon> 创建备份
        </el-button>
      </div>
    </div>

    <div class="bg-white rounded-2xl shadow-sm border border-slate-100 overflow-hidden">
      <el-table :data="backups" v-loading="loading" stripe class="w-full" empty-text="暂无备份记录">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="filename" label="文件名" min-width="280">
          <template #default="{ row }">
            <div class="flex items-center gap-2">
              <el-icon class="text-blue-500"><Document /></el-icon>
              <span class="font-mono text-sm">{{ row.filename }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="file_size" label="文件大小" width="120" align="center">
          <template #default="{ row }">
            <span class="text-sm">{{ formatFileSize(row.file_size) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="110" align="center">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" align="center">
          <template #default="{ row }">
            <span class="text-sm text-slate-500">{{ formatTime(row.created_at) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template #default="{ row }">
            <div class="flex items-center justify-center gap-1">
              <el-button text type="primary" size="small" @click="downloadBackup(row)" :disabled="row.status !== 'completed'">
                <el-icon><Download /></el-icon>
              </el-button>
              <el-popconfirm title="确定删除此备份？删除后不可恢复" @confirm="deleteBackup(row)">
                <template #reference>
                  <el-button text type="danger" size="small">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </template>
              </el-popconfirm>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="flex items-center justify-between px-6 py-4 border-t border-slate-100">
        <span class="text-sm text-slate-500">共 {{ total }} 条记录</span>
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="prev, pager, next"
          @current-change="fetchBackups"
          small
        />
      </div>
    </div>

    <el-dialog v-model="showRestoreDialog" title="恢复数据库" width="480px" :close-on-click-modal="false">
      <div class="space-y-4">
        <el-alert
          title="警告：恢复操作将覆盖当前数据库所有数据，请谨慎操作！"
          type="warning"
          :closable="false"
          show-icon
        />
        <el-upload
          ref="uploadRef"
          drag
          :auto-upload="false"
          :limit="1"
          accept=".sql"
          :on-change="handleFileChange"
          :on-remove="handleFileRemove"
        >
          <el-icon class="text-4xl text-slate-400 mb-3"><UploadFilled /></el-icon>
          <div class="text-sm text-slate-600">将 .sql 备份文件拖到此处，或点击上传</div>
          <template #tip>
            <div class="text-xs text-slate-400 mt-2">仅支持 .sql 格式的备份文件</div>
          </template>
        </el-upload>
      </div>
      <template #footer>
        <el-button @click="showRestoreDialog = false">取消</el-button>
        <el-button type="danger" @click="restoreBackup" :loading="restoring" :disabled="!uploadFile">
          确认恢复
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Upload, Document, Download, Delete, UploadFilled } from '@element-plus/icons-vue'
import request from '@/api/request'

interface BackupRecord {
  id: number
  filename: string
  file_size: number
  status: string
  created_at: string
}

const backups = ref<BackupRecord[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)
const creating = ref(false)
const restoring = ref(false)
const showRestoreDialog = ref(false)
const uploadFile = ref<File | null>(null)
const uploadRef = ref()

onMounted(() => {
  fetchBackups()
})

async function fetchBackups() {
  loading.value = true
  try {
    const res = await request.get('/backups')
    if (res?.data) {
      backups.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.message || '获取备份列表失败')
  } finally {
    loading.value = false
  }
}

async function createBackup() {
  creating.value = true
  try {
    const res = await request.post('/backups')
    if (res?.data) {
      ElMessage.success('备份创建成功')
      fetchBackups()
    }
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.message || '创建备份失败')
  } finally {
    creating.value = false
  }
}

function downloadBackup(row: BackupRecord) {
  const token = localStorage.getItem('fayhub_token') || ''
  const link = document.createElement('a')
  link.href = `/api/backups/${row.id}/download?token=${token}`
  link.download = row.filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

async function deleteBackup(row: BackupRecord) {
  try {
    const res = await request.delete(`/backups/${row.id}`)
    if (res?.data) {
      ElMessage.success('备份删除成功')
      fetchBackups()
    }
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.message || '删除备份失败')
  }
}

function handleFileChange(file: any) {
  uploadFile.value = file.raw
}

function handleFileRemove() {
  uploadFile.value = null
}

async function restoreBackup() {
  if (!uploadFile.value) {
    ElMessage.warning('请先选择备份文件')
    return
  }

  restoring.value = true
  try {
    const formData = new FormData()
    formData.append('file', uploadFile.value)

    const res = await request.post('/backups/restore', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })

    if (res?.data) {
      ElMessage.success('数据库恢复成功')
      showRestoreDialog.value = false
      uploadFile.value = null
      uploadRef.value?.clearFiles()
      fetchBackups()
    }
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.message || '恢复数据库失败')
  } finally {
    restoring.value = false
  }
}

function formatFileSize(bytes: number): string {
  if (!bytes || bytes === 0) return '-'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let size = bytes
  while (size >= 1024 && i < units.length - 1) {
    size /= 1024
    i++
  }
  return `${size.toFixed(1)} ${units[i]}`
}

function formatTime(dateStr: string): string {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

function statusType(status: string): 'success' | 'warning' | 'danger' | 'info' {
  switch (status) {
    case 'completed': return 'success'
    case 'pending': return 'warning'
    case 'failed': return 'danger'
    default: return 'info'
  }
}

function statusLabel(status: string): string {
  switch (status) {
    case 'completed': return '已完成'
    case 'pending': return '进行中'
    case 'failed': return '失败'
    default: return status
  }
}
</script>
