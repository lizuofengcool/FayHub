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
import { ElMessage } from 'element-plus'
import { Check, Link, Wallet, Lock, Monitor, InfoFilled } from '@element-plus/icons-vue'
import request from '@/api/request'

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
})
</script>

<style scoped>
:deep(.el-form-item__label) {
  height: 32px;
  line-height: 32px;
  margin-bottom: 0;
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
</style>
