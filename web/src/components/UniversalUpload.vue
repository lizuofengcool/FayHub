<template>
  <div class="universal-upload">
    <el-upload
      :show-file-list="false"
      :before-upload="beforeUpload"
      :http-request="handleUpload"
      :accept="accept"
      :disabled="uploading"
      :multiple="multiple"
      :drag="drag"
    >
      <template v-if="drag">
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">
          拖拽文件到此处或 <em>点击上传</em>
        </div>
        <template v-if="tip">
          <div class="el-upload__tip">{{ tip }}</div>
        </template>
      </template>
      <template v-else>
        <slot>
          <el-button :type="buttonType" :loading="uploading" :disabled="uploading">
            <el-icon class="mr-1"><Upload /></el-icon>
            {{ uploading ? '上传中...' : buttonText }}
          </el-button>
        </slot>
      </template>
    </el-upload>

    <div v-if="fileList.length > 0" class="mt-3">
      <div v-for="(file, index) in fileList" :key="index" class="flex items-center justify-between py-2 px-3 bg-slate-50 rounded-lg mb-2">
        <div class="flex items-center gap-2 flex-1 min-w-0">
          <el-icon :size="18" color="#3b82f6"><Document /></el-icon>
          <span class="text-sm text-slate-700 truncate">{{ file.name }}</span>
          <span class="text-xs text-slate-400">({{ formatSize(file.size) }})</span>
        </div>
        <div class="flex items-center gap-2 ml-2">
          <n-tag v-if="file.status === 'uploading'" type="warning" size="small">上传中</n-tag>
          <n-tag v-else-if="file.status === 'done'" type="success" size="small">完成</n-tag>
          <n-tag v-else-if="file.status === 'error'" type="error" size="small">失败</n-tag>
          <el-button type="danger" link size="small" @click="removeFile(index)">
            <el-icon><Delete /></el-icon>
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useMessage } from 'naive-ui'
import { Upload, UploadFilled, Document, Delete } from '@element-plus/icons-vue'
import fileApi from '@/api/file'

const message = useMessage()

interface UploadFile {
  name: string
  size: number
  status: 'uploading' | 'done' | 'error'
  url?: string
  fileId?: number
}

const props = withDefaults(defineProps<{
  accept?: string
  maxSize?: number
  multiple?: boolean
  drag?: boolean
  buttonText?: string
  buttonType?: string
  tip?: string
}>(), {
  accept: '',
  maxSize: 50,
  multiple: false,
  drag: false,
  buttonText: '上传文件',
  buttonType: 'primary',
  tip: ''
})

const emit = defineEmits<{
  (e: 'success', result: { fileId: number; url: string; name: string; size: number }): void
  (e: 'error', error: Error): void
}>()

const uploading = ref(false)
const fileList = ref<UploadFile[]>([])

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function beforeUpload(file: File) {
  const maxBytes = props.maxSize * 1024 * 1024
  if (file.size > maxBytes) {
    message.error(`文件大小不能超过 ${props.maxSize}MB`)
    return false
  }
  return true
}

async function handleUpload(options: any) {
  uploading.value = true
  const fileItem: UploadFile = {
    name: options.file.name,
    size: options.file.size,
    status: 'uploading'
  }
  fileList.value.push(fileItem)

  try {
    const formData = new FormData()
    formData.append('file', options.file)
    const res = await fileApi.upload(formData)
    fileItem.status = 'done'
    fileItem.url = res.data.url
    fileItem.fileId = res.data.id
    emit('success', {
      fileId: res.data.id,
      url: res.data.url,
      name: res.data.original_name,
      size: res.data.file_size
    })
  } catch (e: any) {
    fileItem.status = 'error'
    message.error(e?.message || '上传失败')
    emit('error', e)
  } finally {
    uploading.value = false
  }
}

function removeFile(index: number) {
  fileList.value.splice(index, 1)
}

defineExpose({
  fileList,
  clearFiles: () => { fileList.value = [] }
})
</script>
