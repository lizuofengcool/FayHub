﻿<template>
  <div class="notification-channel-page">
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 pb-3 flex items-center justify-between">
        <div>
          <h2 class="text-lg font-bold text-slate-800">通知渠道</h2>
          <p class="text-slate-400 text-xs mt-0.5">管理短信、邮件等通知渠道与模板</p>
        </div>
      </div>

      <el-tabs v-model="activeTab" class="notification-tabs">
      <el-tab-pane label="渠道配置" name="channels">
        <div class="mb-4 flex justify-end">
          <el-button type="default" @click="openChannelDialog()">
            <el-icon class="mr-1"><Plus /></el-icon>
            新增渠道
          </el-button>
        </div>

        <el-table v-loading="channelLoading" :data="channels" stripe>
          <el-table-column prop="name" label="渠道名称" width="160" />
          <el-table-column prop="type" label="类型" width="100" align="center">
            <template #default="{ row }">
              <n-tag :type="row.type === 'email' ? 'info' : row.type === 'sms' ? 'success' : 'info'" size="small">
                {{ row.type === 'email' ? '邮件' : row.type === 'sms' ? '短信' : row.type }}
              </n-tag>
            </template>
          </el-table-column>
          <el-table-column prop="provider" label="服务商" width="140" />
          <el-table-column prop="config" label="配置" min-width="200" show-overflow-tooltip>
            <template #default="{ row }">
              <span class="text-xs text-slate-500">{{ row.config }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="80" align="center">
            <template #default="{ row }">
              <el-switch
                :model-value="row.status === 1"
                @change="handleToggleChannel(row)"
              />
            </template>
          </el-table-column>
          <el-table-column prop="is_default" label="默认" width="70" align="center">
            <template #default="{ row }">
              <n-tag :type="row.is_default ? 'success' : 'info'" size="small">
                {{ row.is_default ? '是' : '否' }}
              </n-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="140" align="center">
            <template #default="{ row }">
              <el-button size="small" link type="default" @click="openChannelDialog(row)">
                编辑
              </el-button>
              <n-popconfirm title="确定删除？" @confirm="handleDeleteChannel(row)">
                <template #trigger>
                  <el-button size="small" link type="error">删除</el-button>
                </template>
              </n-popconfirm>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="消息模板" name="templates">
        <div class="mb-4 flex justify-end">
          <el-button type="default" @click="openTemplateDialog()">
            <el-icon class="mr-1"><Plus /></el-icon>
            新增模板
          </el-button>
        </div>

        <el-table v-loading="templateLoading" :data="templates" stripe>
          <el-table-column prop="name" label="模板名称" width="160" />
          <el-table-column prop="code" label="模板编码" width="140" />
          <el-table-column prop="channel_id" label="渠道ID" width="80" />
          <el-table-column prop="subject" label="标题" width="180" show-overflow-tooltip />
          <el-table-column prop="content" label="内容" min-width="200" show-overflow-tooltip>
            <template #default="{ row }">
              <span class="text-xs text-slate-500">{{ row.content }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="80" align="center">
            <template #default="{ row }">
              <el-switch
                :model-value="row.status === 1"
                @change="handleToggleTemplate(row)"
              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="140" align="center">
            <template #default="{ row }">
              <el-button size="small" link type="default" @click="openTemplateDialog(row)">
                编辑
              </el-button>
              <n-popconfirm title="确定删除？" @confirm="handleDeleteTemplate(row)">
                <template #trigger>
                  <el-button size="small" link type="error">删除</el-button>
                </template>
              </n-popconfirm>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="发送记录" name="records">
        <el-table v-loading="recordLoading" :data="records" stripe>
          <el-table-column prop="recipient" label="接收人" width="200" />
          <el-table-column prop="subject" label="标题" width="180" show-overflow-tooltip />
          <el-table-column prop="status" label="状态" width="100" align="center">
            <template #default="{ row }">
              <n-tag
                :type="row.status === 'success' ? 'success' : row.status === 'pending' ? 'warning' : 'error'"
                size="small"
              >
                {{ row.status === 'success' ? '成功' : row.status === 'pending' ? '待发送' : '失败' }}
              </n-tag>
            </template>
          </el-table-column>
          <el-table-column prop="sent_at" label="发送时间" width="170">
            <template #default="{ row }">
              <span>{{ formatTime(row.sent_at) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="error" label="错误信息" min-width="150" show-overflow-tooltip>
            <template #default="{ row }">
              <span class="text-red-500 text-xs">{{ row.error || '-' }}</span>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>
    </div>

    <el-dialog v-model="channelVisible" :title="isChannelEdit ? '编辑渠道' : '新增渠道'" width="500px">
      <el-form ref="channelFormRef" :model="channelForm" :rules="channelRules" label-width="100px">
        <el-form-item label="渠道名称" prop="name">
          <el-input v-model="channelForm.name" placeholder="如: 阿里云邮件" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="channelForm.type" style="width: 100%">
            <el-option label="邮件" value="email" />
            <el-option label="短信" value="sms" />
          </el-select>
        </el-form-item>
        <el-form-item label="服务商" prop="provider">
          <el-input v-model="channelForm.provider" placeholder="如: aliyun" />
        </el-form-item>
        <el-form-item label="配置JSON" prop="config">
          <el-input v-model="channelForm.config" type="textarea" :rows="4" placeholder='{"access_key":"xxx","secret":"xxx"}' />
        </el-form-item>
        <el-form-item label="设为默认">
          <el-switch v-model="channelForm.is_default" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="channelVisible = false">取消</el-button>
        <el-button type="default" @click="handleChannelSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="templateVisible" :title="isTemplateEdit ? '编辑模板' : '新增模板'" width="500px">
      <el-form ref="templateFormRef" :model="templateForm" :rules="templateRules" label-width="100px">
        <el-form-item label="模板名称" prop="name">
          <el-input v-model="templateForm.name" placeholder="如: 注册验证码" />
        </el-form-item>
        <el-form-item label="模板编码" prop="code">
          <el-input v-model="templateForm.code" placeholder="如: register_code" />
        </el-form-item>
        <el-form-item label="渠道ID" prop="channel_id">
          <el-input-number v-model="templateForm.channel_id" :min="1" />
        </el-form-item>
        <el-form-item label="标题" prop="subject">
          <el-input v-model="templateForm.subject" placeholder="如: 验证码通知" />
        </el-form-item>
        <el-form-item label="内容" prop="content">
          <el-input v-model="templateForm.content" type="textarea" :rows="4" placeholder="如: 您的验证码是{{code}}" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="templateVisible = false">取消</el-button>
        <el-button type="default" @click="handleTemplateSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
const message = useMessage()
import type { FormInstance, FormRules } from 'element-plus'
import notificationApi, {
  type NotificationChannel,
  type NotificationTemplate,
  type NotificationRecord
} from '@/api/notification'

const activeTab = ref('channels')

const channelLoading = ref(false)
const channels = ref<NotificationChannel[]>([])

const channelVisible = ref(false)
const isChannelEdit = ref(false)
const channelFormRef = ref<FormInstance>()
const channelForm = reactive({
  id: 0,
  name: '',
  type: 'email',
  provider: '',
  config: '',
  is_default: 0 as number
})

const channelRules: FormRules = {
  name: [{ required: true, message: '请输入渠道名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  provider: [{ required: true, message: '请输入服务商', trigger: 'blur' }],
  config: [{ required: true, message: '请输入配置', trigger: 'blur' }]
}

const templateLoading = ref(false)
const templates = ref<NotificationTemplate[]>([])

const templateVisible = ref(false)
const isTemplateEdit = ref(false)
const templateFormRef = ref<FormInstance>()
const templateForm = reactive({
  id: 0,
  name: '',
  code: '',
  channel_id: 1,
  subject: '',
  content: ''
})

const templateRules: FormRules = {
  name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入模板编码', trigger: 'blur' }],
  channel_id: [{ required: true, message: '请输入渠道ID', trigger: 'blur' }],
  content: [{ required: true, message: '请输入内容', trigger: 'blur' }]
}

const recordLoading = ref(false)
const records = ref<NotificationRecord[]>([])

function formatTime(time: string | null) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

async function fetchChannels() {
  channelLoading.value = true
  try {
    const res = await notificationApi.listChannels()
    channels.value = res.data?.list || []
  } catch {
    // ignore
  } finally {
    channelLoading.value = false
  }
}

async function fetchTemplates() {
  templateLoading.value = true
  try {
    const res = await notificationApi.listTemplates()
    templates.value = res.data?.list || []
  } catch {
    // ignore
  } finally {
    templateLoading.value = false
  }
}

async function fetchRecords() {
  recordLoading.value = true
  try {
    const res = await notificationApi.getRecords()
    records.value = res.data?.list || []
  } catch {
    // ignore
  } finally {
    recordLoading.value = false
  }
}

function openChannelDialog(row?: NotificationChannel) {
  if (row) {
    isChannelEdit.value = true
    channelForm.id = row.id
    channelForm.name = row.name
    channelForm.type = row.type
    channelForm.provider = row.provider
    channelForm.config = row.config
    channelForm.is_default = row.is_default
  } else {
    isChannelEdit.value = false
    channelForm.id = 0
    channelForm.name = ''
    channelForm.type = 'email'
    channelForm.provider = ''
    channelForm.config = ''
    channelForm.is_default = 0
  }
  channelVisible.value = true
}

async function handleChannelSubmit() {
  const valid = await channelFormRef.value?.validate().catch(() => false)
  if (!valid) return

  try {
    if (isChannelEdit.value) {
      await notificationApi.updateChannel(channelForm.id, channelForm)
      message.success('更新成功')
    } else {
      await notificationApi.createChannel(channelForm)
      message.success('创建成功')
    }
    channelVisible.value = false
    fetchChannels()
  } catch {
    message.error('操作失败')
  }
}

async function handleDeleteChannel(row: NotificationChannel) {
  try {
    await notificationApi.deleteChannel(row.id)
    message.success('删除成功')
    fetchChannels()
  } catch {
    message.error('删除失败')
  }
}

async function handleToggleChannel(row: NotificationChannel) {
  try {
    await notificationApi.updateChannel(row.id, { status: row.status === 1 ? 0 : 1 })
    message.success('状态已更新')
    fetchChannels()
  } catch {
    message.error('操作失败')
  }
}

function openTemplateDialog(row?: NotificationTemplate) {
  if (row) {
    isTemplateEdit.value = true
    templateForm.id = row.id
    templateForm.name = row.name
    templateForm.code = row.code
    templateForm.channel_id = row.channel_id
    templateForm.subject = row.subject
    templateForm.content = row.content
  } else {
    isTemplateEdit.value = false
    templateForm.id = 0
    templateForm.name = ''
    templateForm.code = ''
    templateForm.channel_id = 1
    templateForm.subject = ''
    templateForm.content = ''
  }
  templateVisible.value = true
}

async function handleTemplateSubmit() {
  const valid = await templateFormRef.value?.validate().catch(() => false)
  if (!valid) return

  try {
    if (isTemplateEdit.value) {
      await notificationApi.updateTemplate(templateForm.id, templateForm)
      message.success('更新成功')
    } else {
      await notificationApi.createTemplate(templateForm)
      message.success('创建成功')
    }
    templateVisible.value = false
    fetchTemplates()
  } catch {
    message.error('操作失败')
  }
}

async function handleDeleteTemplate(row: NotificationTemplate) {
  try {
    await notificationApi.deleteTemplate(row.id)
    message.success('删除成功')
    fetchTemplates()
  } catch {
    message.error('删除失败')
  }
}

async function handleToggleTemplate(row: NotificationTemplate) {
  try {
    await notificationApi.updateTemplate(row.id, { status: row.status === 1 ? 0 : 1 })
    message.success('状态已更新')
    fetchTemplates()
  } catch {
    message.error('操作失败')
  }
}

onMounted(() => {
  fetchChannels()
  fetchTemplates()
  fetchRecords()
})
</script>

<style scoped>
.notification-tabs :deep(.el-tabs__header) {
  padding: 0 20px;
  margin-bottom: 0;
}
.notification-tabs :deep(.el-tabs__content) {
  padding: 16px 20px 20px;
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

:deep(.el-button) {
  height: 32px;
  padding: 8px 12px;
}
</style>
