﻿<template>
  <div class="file-page">
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 pb-3 flex items-center justify-between">
        <div>
          <h2 class="text-lg font-bold text-slate-800">文件管理</h2>
          <p class="text-slate-400 text-xs mt-0.5">上传、浏览和管理系统文件</p>
        </div>
        <el-upload
          :show-file-list="false"
          :before-upload="beforeUpload"
          :http-request="handleUpload"
          :multiple="true"
        >
          <el-button type="default">
            <el-icon class="mr-1"><Upload /></el-icon>
            上传文件
          </el-button>
        </el-upload>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4 px-4">
      <div class="bg-white rounded-xl border border-slate-100 p-4 shadow-sm">
        <p class="text-sm text-slate-500">文件总数</p>
        <p class="text-2xl font-bold text-slate-800 mt-1">{{ total }}</p>
      </div>
      <div class="bg-white rounded-xl border border-slate-100 p-4 shadow-sm">
        <p class="text-sm text-slate-500">总大小</p>
        <p class="text-2xl font-bold text-blue-600 mt-1">{{ formatSize(totalSize) }}</p>
      </div>
      <div class="bg-white rounded-xl border border-slate-100 p-4 shadow-sm">
        <p class="text-sm text-slate-500">存储驱动</p>
        <p class="text-2xl font-bold text-green-600 mt-1">本地存储</p>
      </div>
      </div>

    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 flex gap-3 flex-wrap">
        <el-input
          v-model="keyword"
          placeholder="搜索文件名"
          clearable
          style="width: 240px"
          @keyup.enter="fetchFiles"
          @clear="fetchFiles"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-select v-model="mimeTypeFilter" placeholder="文件类型" clearable style="width: 150px" @change="fetchFiles">
          <el-option label="图片" value="image/" />
          <el-option label="PDF" value="application/pdf" />
          <el-option label="Word" value="application/msword" />
          <el-option label="Excel" value="application/vnd.ms-excel" />
          <el-option label="压缩包" value="application/zip" />
        </el-select>
        <el-button type="default" @click="fetchFiles">查询</el-button>
      </div>

      <el-table v-loading="loading" :data="files" stripe class="w-full">
        <el-table-column prop="original_name" label="文件名" min-width="220">
          <template #default="{ row }">
            <div class="flex items-center gap-2">
              <el-icon :size="20" :color="fileIconColor(row.mime_type)">
                <component :is="fileIcon(row.mime_type)" />
              </el-icon>
              <span class="text-slate-800 truncate max-w-[200px]" :title="row.original_name">{{ row.original_name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="file_size" label="大小" width="110" align="center">
          <template #default="{ row }">
            {{ formatSize(row.file_size) }}
          </template>
        </el-table-column>
        <el-table-column prop="mime_type" label="类型" width="140">
          <template #default="{ row }">
            <n-tag size="small" :type="mimeTypeTag(row.mime_type)">{{ row.mime_type || '未知' }}</n-tag>
          </template>
        </el-table-column>
        <el-table-column prop="storage_driver" label="存储" width="90" align="center">
          <template #default="{ row }">
            <n-tag size="small" type="default">{{ row.storage_driver }}</n-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="上传时间" width="170" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="default" link size="small" @click="handleDownload(row)">下载</el-button>
            <el-button type="error" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="p-4 flex justify-end">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="fetchFiles"
        />
      </div>
    </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
const message = useMessage()
import { Upload, Search, Document, Picture, Folder, VideoPlay } from '@element-plus/icons-vue'
import fileApi, { type FileRecord } from '@/api/file'

const loading = ref(false)
const files = ref<FileRecord[]>([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const totalSize = ref(0)
const keyword = ref('')
const mimeTypeFilter = ref('')

async function fetchFiles() {
  loading.value = true
  try {
    const res = await fileApi.list({
      page: page.value,
      page_size: pageSize.value,
      keyword: keyword.value || undefined,
      mime_type: mimeTypeFilter.value || undefined
    })
    files.value = res.data?.list || []
    total.value = res.data?.total || 0
    totalSize.value = res.data?.total_size || 0
  } catch (err: any) {
    message.error(err.message || '获取文件列表失败')
  } finally {
    loading.value = false
  }
}

function beforeUpload(file: File) {
  const maxSize = 10 * 1024 * 1024
  if (file.size > maxSize) {
    message.error('文件大小不能超过10MB')
    return false
  }
  return true
}

async function handleUpload(options: any) {
  try {
    await fileApi.upload(options.file)
    message.success('上传成功')
    fetchFiles()
  } catch (err: any) {
    message.error(err.message || '上传失败')
  }
}

async function handleDownload(row: FileRecord) {
  try {
    const res = await fileApi.download(row.id)
    const blob = new Blob([res as any])
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = row.original_name
    link.click()
    window.URL.revokeObjectURL(url)
  } catch (err: any) {
    message.error(err.message || '下载失败')
  }
}

async function handleDelete(row: FileRecord) {
  try {
    await dialog.warning(`确定要删除文件"${row.original_name}"吗？`, '确认删除', { type: 'warning' })
    await fileApi.delete(row.id)
    message.success('删除成功')
    fetchFiles()
  } catch (e) { console.error('handleDelete failed:', e); }
}

function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + units[i]
}

function fileIcon(mimeType: string) {
  if (mimeType?.startsWith('image/')) return Picture
  if (mimeType?.startsWith('video/')) return VideoPlay
  if (mimeType?.startsWith('application/')) return Document
  return Folder
}

function fileIconColor(mimeType: string): string {
  if (mimeType?.startsWith('image/')) return '#67c23a'
  if (mimeType?.startsWith('video/')) return '#e6a23c'
  if (mimeType?.includes('pdf')) return '#f56c6c'
  if (mimeType?.includes('zip')) return '#909399'
  return '#409eff'
}

function mimeTypeTag(mimeType: string): string {
  if (mimeType?.startsWith('image/')) return 'success'
  if (mimeType?.startsWith('video/')) return 'warning'
  if (mimeType?.includes('pdf')) return 'danger'
  return 'info'
}

onMounted(() => {
  fetchFiles()
})
</script>

<style scoped>
:deep(.el-input__wrapper) {
  height: 32px;
}

:deep(.el-select .el-input__wrapper) {
  height: 32px;
}

:deep(.el-button) {
  height: 32px;
  padding: 8px 12px;
}
</style>
