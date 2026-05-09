﻿<template>
  <div class="payment-config-page">
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 pb-3 flex items-center justify-between">
        <div>
          <h2 class="text-lg font-bold text-slate-800">支付参数配置</h2>
          <p class="text-slate-400 text-xs mt-0.5">管理微信支付与支付宝的商户参数配置</p>
        </div>
        <el-button type="primary" @click="saveConfig" :loading="saving">
          <el-icon class="mr-1"><Check /></el-icon> 保存配置
        </el-button>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 px-4 pb-4">
      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-6">
        <div class="flex items-center mb-6">
          <el-icon class="text-2xl text-green-500 mr-3"><ChatDotRound /></el-icon>
          <h3 class="text-lg font-semibold text-slate-800">微信支付</h3>
          <el-switch v-model="config.wechat.enabled" class="ml-auto" />
        </div>
        <el-form label-position="top" :disabled="!config.wechat.enabled">
          <el-form-item label="商户号 (MchID)">
            <el-input v-model="config.wechat.mch_id" placeholder="请输入微信支付商户号" />
          </el-form-item>
          <el-form-item label="应用ID (AppID)">
            <el-input v-model="config.wechat.app_id" placeholder="请输入关联的公众号/小程序AppID" />
          </el-form-item>
          <el-form-item label="API密钥">
            <el-input v-model="config.wechat.api_key" type="password" show-password placeholder="请输入API密钥" />
          </el-form-item>
          <el-form-item label="API证书序列号">
            <el-input v-model="config.wechat.serial_no" placeholder="请输入证书序列号" />
          </el-form-item>
          <el-form-item label="回调通知URL">
            <el-input v-model="config.wechat.notify_url" placeholder="https://your-domain.com/api/payment/wechat/notify" />
          </el-form-item>
        </el-form>
      </div>

      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-6">
        <div class="flex items-center mb-6">
          <el-icon class="text-2xl text-blue-500 mr-3"><Wallet /></el-icon>
          <h3 class="text-lg font-semibold text-slate-800">支付宝</h3>
          <el-switch v-model="config.alipay.enabled" class="ml-auto" />
        </div>
        <el-form label-position="top" :disabled="!config.alipay.enabled">
          <el-form-item label="应用ID (AppID)">
            <el-input v-model="config.alipay.app_id" placeholder="请输入支付宝应用AppID" />
          </el-form-item>
          <el-form-item label="应用私钥">
            <el-input v-model="config.alipay.private_key" type="password" show-password placeholder="请输入应用私钥" />
          </el-form-item>
          <el-form-item label="支付宝公钥">
            <el-input v-model="config.alipay.public_key" type="textarea" :rows="3" placeholder="请输入支付宝公钥" />
          </el-form-item>
          <el-form-item label="回调通知URL">
            <el-input v-model="config.alipay.notify_url" placeholder="https://your-domain.com/api/payment/alipay/notify" />
          </el-form-item>
          <el-form-item label="沙箱模式">
            <el-switch v-model="config.alipay.sandbox" />
          </el-form-item>
        </el-form>
      </div>
    </div>

    <div class="bg-amber-50 border border-amber-200 rounded-xl p-4">
      <div class="flex items-start">
        <el-icon class="text-amber-500 text-xl mr-3 mt-0.5"><WarningFilled /></el-icon>
        <div>
          <p class="font-medium text-amber-800">安全提示</p>
          <p class="text-sm text-amber-700 mt-1">支付密钥属于敏感信息，请勿泄露。修改配置后需重启服务或刷新缓存方可生效。建议在生产环境使用环境变量管理密钥。</p>
        </div>
      </div>
    </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import { Check, ChatDotRound, Wallet, WarningFilled } from '@element-plus/icons-vue'
import request from '@/api/request'

interface PaymentConfig {
  wechat: {
    enabled: boolean
    mch_id: string
    app_id: string
    api_key: string
    serial_no: string
    notify_url: string
  }
  alipay: {
    enabled: boolean
    app_id: string
    private_key: string
    public_key: string
    notify_url: string
    sandbox: boolean
  }
}

const saving = ref(false)
const config = ref<PaymentConfig>({
  wechat: {
    enabled: false,
    mch_id: '',
    app_id: '',
    api_key: '',
    serial_no: '',
    notify_url: ''
  },
  alipay: {
    enabled: false,
    app_id: '',
    private_key: '',
    public_key: '',
    notify_url: '',
    sandbox: false
  }
})

async function loadConfig() {
  try {
    const data = await request.get('/payment/config')
    if (data.data) {
      config.value = { ...config.value, ...data.data }
    }
  } catch (e) {
    console.error('加载支付配置失败:', e)
  }
}

async function saveConfig() {
  saving.value = true
  try {
    await request.put('/payment/config', config.value)
    message.success('支付配置保存成功')
  } catch (e: any) {
    message.error('保存失败: ' + (e.message || '未知错误'))
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadConfig()
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
