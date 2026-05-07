<template>
  <div class="channel-page">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold text-slate-800">渠道配置管理</h2>
        <p class="text-slate-500 mt-1 text-sm">管理微信公众号、小程序、支付等渠道配置</p>
      </div>
      <el-button type="primary" @click="openConfigDialog()">
        <el-icon class="mr-1"><Plus /></el-icon>
        新增配置
      </el-button>
    </div>

    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 border-b border-slate-100 flex gap-4 flex-wrap">
        <el-select v-model="filters.channel_type" placeholder="选择渠道类型" clearable style="width: 220px">
          <el-option label="全部渠道" :value="undefined" />
          <el-option label="微信公众号" value="wechat_mp" />
          <el-option label="微信小程序" value="wechat_mini" />
          <el-option label="微信支付" value="wechat_pay" />
          <el-option label="支付宝小程序" value="alipay_mini" />
          <el-option label="抖音小程序" value="douyin_mini" />
        </el-select>
        <el-select v-model="filters.status" placeholder="选择状态" clearable style="width: 150px">
          <el-option label="全部状态" :value="undefined" />
          <el-option label="启用" :value="1" />
          <el-option label="禁用" :value="0" />
        </el-select>
        <el-button type="primary" @click="fetchConfigs">
          <el-icon class="mr-1"><Search /></el-icon>
          搜索
        </el-button>
        <el-button @click="resetFilters">
          <el-icon class="mr-1"><RefreshLeft /></el-icon>
          重置
        </el-button>
      </div>

      <el-table v-loading="loading" :data="configs" stripe highlight-current-row class="w-full">
        <el-table-column prop="channel_name" label="配置名称" min-width="150" show-overflow-tooltip />
        <el-table-column prop="channel_type" label="渠道类型" width="150">
          <template #default="{ row }">
            <el-tag :type="getChannelTypeColor(row.channel_type)">
              {{ getChannelTypeName(row.channel_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="app_id" label="AppID" min-width="150" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openConfigDialog(row)">编辑</el-button>
            <el-button type="primary" link size="small" @click="copyConfig(row)">复制</el-button>
            <el-button type="danger" link size="small" @click="handleDeleteConfig(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <div class="p-4 flex justify-end">
        <el-pagination
          v-model:current-page="page"
          :total="total"
          :page-size="20"
          layout="total, prev, next"
          @current-change="fetchConfigs"
        />
      </div>
    </div>

    <el-dialog v-model="configDialogVisible" :title="configForm.id ? '编辑渠道配置' : '新增渠道配置'" width="700px">
      <el-form :model="configForm" label-width="140px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="渠道类型" required>
            <el-select v-model="configForm.channel_type" placeholder="请选择渠道类型" :disabled="!!configForm.id">
              <el-option label="微信公众号" value="wechat_mp" />
              <el-option label="微信小程序" value="wechat_mini" />
              <el-option label="微信支付" value="wechat_pay" />
              <el-option label="支付宝小程序" value="alipay_mini" />
              <el-option label="抖音小程序" value="douyin_mini" />
            </el-select>
          </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="配置名称" required>
              <el-input v-model="configForm.channel_name" placeholder="请输入配置名称" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">基础配置</el-divider>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="AppID">
              <el-input v-model="configForm.app_id" placeholder="请输入AppID" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="AppSecret">
              <el-input v-model="configForm.app_secret" type="password" show-password placeholder="请输入AppSecret" />
            </el-form-item>
          </el-col>
        </el-row>

        <template v-if="configForm.channel_type === 'wechat_pay'">
          <el-divider content-position="left">支付配置</el-divider>

          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="商户号">
                <el-input v-model="configForm.merchant_id" placeholder="请输入微信支付商户号" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="证书序列号">
                <el-input v-model="configForm.cert_serial_no" placeholder="请输入证书序列号" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-form-item label="支付公钥">
            <el-input v-model="configForm.pay_public_key" type="textarea" :rows="3" placeholder="请输入支付公钥" />
          </el-form-item>
          <el-form-item label="支付私钥">
            <el-input v-model="configForm.pay_private_key" type="textarea" :rows="3" placeholder="请输入支付私钥" />
          </el-form-item>
        </template>

        <template v-if="configForm.channel_type === 'wechat_mp'">
          <el-divider content-position="left">公众号配置</el-divider>

          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="Token">
                <el-input v-model="configForm.token" placeholder="请输入服务器Token" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="EncodingAESKey">
                <el-input v-model="configForm.encoding_aes_key" placeholder="请输入消息加密密钥" />
              </el-form-item>
            </el-col>
          </el-row>
        </template>

        <el-form-item label="状态">
          <el-radio-group v-model="configForm.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="configDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSaveConfig">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, RefreshLeft, CopyDocument } from '@element-plus/icons-vue'
import channelApi from '@/api/tenant-channel'
import type { TenantChannelConfig } from '@/api/tenant-channel'

const loading = ref(false)
const submitLoading = ref(false)
const configs = ref<TenantChannelConfig[]>([])
const total = ref(0)
const page = ref(1)
const configDialogVisible = ref(false)

const filters = reactive<{
  channel_type?: string
  status?: number
}>({
  channel_type: undefined,
  status: undefined,
})

const configForm = reactive<Partial<TenantChannelConfig>>({
  channel_type: '',
  channel_name: '',
  app_id: '',
  app_secret: '',
  merchant_id: '',
  pay_public_key: '',
  pay_private_key: '',
  cert_serial_no: '',
  token: '',
  encoding_aes_key: '',
  status: 1,
})

const channelTypeMap: Record<string, string> = {
  wechat_mp: '微信公众号',
  wechat_mini: '微信小程序',
  wechat_pay: '微信支付',
  alipay_mini: '支付宝小程序',
  douyin_mini: '抖音小程序',
  toutiao_mini: '头条小程序',
}

const channelTypeColorMap: Record<string, string> = {
  wechat_mp: 'success',
  wechat_mini: 'primary',
  wechat_pay: 'warning',
  alipay_mini: 'info',
  douyin_mini: 'danger',
  toutiao_mini: 'danger',
}

function getChannelTypeName(type: string) {
  return channelTypeMap[type] || type
}

function getChannelTypeColor(type: string) {
  return channelTypeColorMap[type] || 'info'
}

async function fetchConfigs() {
  loading.value = true
  try {
    const res = await channelApi.listConfigs({
      page: page.value,
      page_size: 20,
      channel_type: filters.channel_type,
      status: filters.status,
    })
    if (res.code === 0 || res.code === 200) {
      configs.value = res.data.list
      total.value = res.data.total
    }
  } catch (e: any) {
    ElMessage.error(e.message || '获取列表失败')
  } finally {
    loading.value = false
  }
}

function openConfigDialog(row?: TenantChannelConfig) {
  if (row) {
    Object.assign(configForm, { ...row })
  } else {
    Object.assign(configForm, {
      id: undefined,
      channel_type: '',
      channel_name: '',
      app_id: '',
      app_secret: '',
      merchant_id: '',
      pay_public_key: '',
      pay_private_key: '',
      cert_serial_no: '',
      token: '',
      encoding_aes_key: '',
      status: 1,
    })
  }
  configDialogVisible.value = true
}

function copyConfig(row: TenantChannelConfig) {
  const copyData = {
    ...row,
    id: undefined,
    channel_name: row.channel_name + ' (副本)',
  }
  Object.assign(configForm, copyData)
  configDialogVisible.value = true
}

async function handleSaveConfig() {
  if (!configForm.channel_type) {
    ElMessage.warning('请选择渠道类型')
    return
  }
  if (!configForm.channel_name) {
    ElMessage.warning('请输入配置名称')
    return
  }

  submitLoading.value = true
  try {
    let res
    if (configForm.id) {
      res = await channelApi.updateConfig(configForm.id, configForm)
    } else {
      res = await channelApi.createConfig(configForm)
    }
    if (res.code === 0 || res.code === 200) {
      ElMessage.success(configForm.id ? '更新成功' : '创建成功')
      configDialogVisible.value = false
      await fetchConfigs()
    }
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}

async function handleDeleteConfig(row: TenantChannelConfig) {
  await ElMessageBox.confirm('确定要删除这个配置吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })

  try {
    const res = await channelApi.deleteConfig(row.id)
    if (res.code === 0 || res.code === 200) {
      ElMessage.success('删除成功')
      await fetchConfigs()
    }
  } catch (e: any) {
    ElMessage.error(e.message || '删除失败')
  }
}

function resetFilters() {
  filters.channel_type = undefined
  filters.status = undefined
  page.value = 1
  fetchConfigs()
}

onMounted(() => {
  fetchConfigs()
})
</script>

<style scoped>
.channel-page {
  padding: 24px;
}
</style>
