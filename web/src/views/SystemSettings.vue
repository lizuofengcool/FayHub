<template>
  <div class="p-6 space-y-6">
    <div class="flex items-center justify-between">
      <h2 class="text-2xl font-bold text-slate-800">系统设置</h2>
      <el-button type="primary" @click="saveSettings" :loading="saving">
        <el-icon class="mr-1"><Check /></el-icon> 保存设置
      </el-button>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-6">
        <div class="flex items-center mb-6">
          <el-icon class="text-2xl text-blue-500 mr-3"><Link /></el-icon>
          <h3 class="text-lg font-semibold text-slate-800">域名配置</h3>
        </div>
        <el-form label-position="top">
          <el-form-item label="管理后台">
            <el-input v-model="settings.domains.admin_url" placeholder="https://admin.yourdomain.com" />
          </el-form-item>
          <el-form-item label="API 服务">
            <el-input v-model="settings.domains.api_url" placeholder="https://api.yourdomain.com" />
          </el-form-item>
          <el-form-item label="主站 (WWW)">
            <el-input v-model="settings.domains.www_url" placeholder="https://www.yourdomain.com" />
          </el-form-item>
          <el-form-item label="插件市场">
            <el-input v-model="settings.domains.market_url" placeholder="https://market.yourdomain.com" />
          </el-form-item>
          <el-form-item label="开发者中心">
            <el-input v-model="settings.domains.dev_url" placeholder="https://dev.yourdomain.com" />
          </el-form-item>
          <el-form-item label="SSO 服务">
            <el-input v-model="settings.domains.sso_url" placeholder="https://sso.yourdomain.com" />
          </el-form-item>
        </el-form>
      </div>

      <div class="space-y-6">
        <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-6">
          <div class="flex items-center mb-6">
            <el-icon class="text-2xl text-green-500 mr-3"><Wallet /></el-icon>
            <h3 class="text-lg font-semibold text-slate-800">支付网关</h3>
          </div>
          <el-form label-position="top">
            <el-form-item label="回调基础URL">
              <el-input v-model="settings.payment.notify_base_url" placeholder="https://api.yourdomain.com" />
            </el-form-item>
            <el-form-item label="订单过期时间（分钟）">
              <el-input-number v-model="settings.payment.order_expire_min" :min="5" :max="120" />
            </el-form-item>
            <el-form-item label="微信支付网关">
              <el-input v-model="settings.payment.wechat_gateway_url" placeholder="https://api.mch.weixin.qq.com" />
            </el-form-item>
            <el-form-item label="支付宝网关">
              <el-input v-model="settings.payment.alipay_gateway_url" placeholder="https://openapi.alipay.com/gateway.do" />
            </el-form-item>
            <el-form-item label="支付宝沙箱网关">
              <el-input v-model="settings.payment.alipay_sandbox_url" placeholder="https://openapi.alipaydev.com/gateway.do" />
            </el-form-item>
          </el-form>
        </div>

        <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-6">
          <div class="flex items-center mb-6">
            <el-icon class="text-2xl text-orange-500 mr-3"><Lock /></el-icon>
            <h3 class="text-lg font-semibold text-slate-800">安全设置</h3>
          </div>
          <el-form label-position="top">
            <el-form-item label="最大登录尝试次数">
              <el-input-number v-model="settings.security.max_login_attempts" :min="1" :max="20" />
            </el-form-item>
            <el-form-item label="锁定时长（分钟）">
              <el-input-number v-model="settings.security.lock_duration_min" :min="1" :max="120" />
            </el-form-item>
          </el-form>
        </div>
      </div>
    </div>

    <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-6">
      <div class="flex items-center mb-4">
        <el-icon class="text-2xl text-slate-500 mr-3"><Monitor /></el-icon>
        <h3 class="text-lg font-semibold text-slate-800">服务器信息</h3>
        <el-tag class="ml-3" :type="settings.server.mode === 'release' ? 'success' : 'warning'">
          {{ settings.server.mode === 'release' ? '生产模式' : '开发模式' }}
        </el-tag>
      </div>
      <div class="grid grid-cols-2 gap-4 text-sm">
        <div class="text-slate-500">服务端口</div>
        <div class="text-slate-800 font-medium">{{ settings.server.port }}</div>
        <div class="text-slate-500">运行模式</div>
        <div class="text-slate-800 font-medium">{{ settings.server.mode }}</div>
      </div>
    </div>

    <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-6">
      <div class="flex items-center justify-between mb-4">
        <div class="flex items-center">
          <el-icon class="text-2xl text-purple-500 mr-3"><FolderOpened /></el-icon>
          <h3 class="text-lg font-semibold text-slate-800">数据备份</h3>
        </div>
        <el-button type="primary" @click="handleCreateBackup" :loading="backupCreating">
          <el-icon class="mr-1"><Plus /></el-icon> 创建备份
        </el-button>
        <el-upload
          :show-file-list="false"
          accept=".sql"
          :before-upload="handleRestoreBackup"
          class="ml-3"
        >
          <el-button type="warning" :loading="restoreLoading">
            <el-icon class="mr-1"><Upload /></el-icon> 恢复数据库
          </el-button>
        </el-upload>
      </div>
      <el-table :data="backups" stripe v-loading="backupLoading" class="w-full">
        <el-table-column prop="filename" label="文件名" min-width="240" />
        <el-table-column label="大小" width="120" align="right">
          <template #default="{ row }">{{ formatSize(row.file_size) }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'completed' ? 'success' : row.status === 'failed' ? 'danger' : 'warning'" size="small">
              {{ row.status === 'completed' ? '完成' : row.status === 'failed' ? '失败' : '进行中' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.status === 'completed'" type="primary" link size="small" @click="handleDownload(row)">下载</el-button>
            <el-button type="danger" link size="small" @click="handleDeleteBackup(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="bg-blue-50 border border-blue-200 rounded-xl p-4">
      <div class="flex items-start">
        <el-icon class="text-blue-500 text-xl mr-3 mt-0.5"><InfoFilled /></el-icon>
        <div>
          <p class="font-medium text-blue-800">配置说明</p>
          <p class="text-sm text-blue-700 mt-1">域名和支付网关配置修改后立即生效（运行时更新），重启服务后将恢复为配置文件中的值。如需永久修改，请同步更新 config.yaml 文件。</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Check, Link, Wallet, Lock, Monitor, InfoFilled, FolderOpened, Plus, Upload } from '@element-plus/icons-vue'
import request from '@/api/request'
import backupApi, { type BackupRecord } from '@/api/backup'

interface SystemSettings {
  domains: {
    admin_url: string
    market_url: string
    dev_url: string
    api_url: string
    sso_url: string
    www_url: string
  }
  payment: {
    notify_base_url: string
    order_expire_min: number
    wechat_gateway_url: string
    alipay_gateway_url: string
    alipay_sandbox_url: string
  }
  security: {
    max_login_attempts: number
    lock_duration_min: number
  }
  server: {
    port: number
    mode: string
  }
}

const saving = ref(false)
const settings = ref<SystemSettings>({
  domains: {
    admin_url: '',
    market_url: '',
    dev_url: '',
    api_url: '',
    sso_url: '',
    www_url: ''
  },
  payment: {
    notify_base_url: '',
    order_expire_min: 30,
    wechat_gateway_url: '',
    alipay_gateway_url: '',
    alipay_sandbox_url: ''
  },
  security: {
    max_login_attempts: 5,
    lock_duration_min: 15
  },
  server: {
    port: 8080,
    mode: 'debug'
  }
})

async function loadSettings() {
  try {
    const data = await request.get('/system/settings')
    if (data.data) {
      settings.value = { ...settings.value, ...data.data }
    }
  } catch {
    ElMessage.warning('加载系统设置失败，使用默认值')
  }
}

async function saveSettings() {
  saving.value = true
  try {
    await request.put('/system/settings', settings.value)
    ElMessage.success('系统设置更新成功')
  } catch (e: any) {
    ElMessage.error('保存失败: ' + (e.message || '未知错误'))
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadSettings()
  fetchBackups()
})

const backups = ref<BackupRecord[]>([])
const backupLoading = ref(false)
const backupCreating = ref(false)

function formatSize(bytes: number): string {
  if (!bytes || bytes <= 0) return '-'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1024 / 1024).toFixed(1) + ' MB'
}

async function fetchBackups() {
  backupLoading.value = true
  try {
    const res = await backupApi.listBackups()
    backups.value = res.data?.list || []
  } catch {} finally {
    backupLoading.value = false
  }
}

async function handleCreateBackup() {
  backupCreating.value = true
  try {
    await backupApi.createBackup()
    ElMessage.success('备份创建成功')
    fetchBackups()
  } catch (err: any) {
    ElMessage.error(err.message || '创建备份失败')
  } finally {
    backupCreating.value = false
  }
}

async function handleDownload(row: BackupRecord) {
  try {
    const res = await backupApi.downloadBackup(row.id)
    const blob = new Blob([res as any])
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = row.filename
    link.click()
    window.URL.revokeObjectURL(url)
  } catch (err: any) {
    ElMessage.error(err.message || '下载失败')
  }
}

async function handleDeleteBackup(row: BackupRecord) {
  try {
    await ElMessageBox.confirm('确定要删除此备份吗？', '确认删除', { type: 'warning' })
    await backupApi.deleteBackup(row.id)
    ElMessage.success('删除成功')
    fetchBackups()
  } catch {}
}

const restoreLoading = ref(false)

async function handleRestoreBackup(file: File) {
  try {
    await ElMessageBox.confirm(
      '恢复数据库将覆盖当前数据，此操作不可逆！确定要继续吗？',
      '危险操作',
      { confirmButtonText: '确定恢复', cancelButtonText: '取消', type: 'error' }
    )
  } catch {
    return false
  }

  restoreLoading.value = true
  try {
    await backupApi.restoreBackup(file)
    ElMessage.success('数据库恢复成功')
  } catch (err: any) {
    ElMessage.error(err.message || '恢复失败')
  } finally {
    restoreLoading.value = false
  }
  return false
}
</script>
