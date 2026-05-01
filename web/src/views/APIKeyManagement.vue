<template>
  <div class="api-key-page">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold text-slate-800">API 密钥管理</h2>
        <p class="text-slate-500 mt-1 text-sm">管理用于外部调用的 API 密钥</p>
      </div>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon class="mr-1"><Plus /></el-icon> 新建密钥
      </el-button>
    </div>

    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <el-table v-loading="loading" :data="keys" stripe class="w-full">
        <el-table-column prop="name" label="名称" min-width="150">
          <template #default="{ row }">
            <div class="font-medium text-slate-800">{{ row.name }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="key_prefix" label="密钥前缀" width="140">
          <template #default="{ row }">
            <code class="bg-slate-100 px-2 py-1 rounded text-sm text-slate-700">{{ row.key_prefix }}****</code>
          </template>
        </el-table-column>
        <el-table-column label="频率限制" width="120" align="center">
          <template #default="{ row }">
            {{ row.rate_limit }}/小时
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_used_at" label="最后使用" width="160">
          <template #default="{ row }">
            {{ row.last_used_at || '未使用' }}
          </template>
        </el-table-column>
        <el-table-column prop="expires_at" label="过期时间" width="160">
          <template #default="{ row }">
            <span v-if="row.expires_at">{{ row.expires_at }}</span>
            <el-tag v-else type="info" size="small">永不过期</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160" />
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" title="新建 API 密钥" width="560px" :close-on-click-modal="false">
      <el-form :model="form" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入密钥名称，如：Market调用" />
        </el-form-item>
        <el-form-item label="频率限制">
          <el-input-number v-model="form.rate_limit" :min="100" :max="100000" :step="100" />
          <span class="ml-2 text-slate-500 text-sm">次/小时</span>
        </el-form-item>
        <el-form-item label="过期时间">
          <el-date-picker
            v-model="form.expires_at"
            type="datetime"
            placeholder="留空则永不过期"
            format="YYYY-MM-DD HH:mm"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="secretVisible" title="密钥创建成功" width="560px" :close-on-click-modal="false">
      <div class="bg-amber-50 border border-amber-200 rounded-xl p-4 mb-4">
        <div class="flex items-start">
          <el-icon class="text-amber-500 text-xl mr-3 mt-0.5"><WarningFilled /></el-icon>
          <div>
            <p class="font-medium text-amber-800">请立即保存密钥</p>
            <p class="text-sm text-amber-700 mt-1">密钥明文仅在创建时显示一次，关闭后无法再次查看。请妥善保管。</p>
          </div>
        </div>
      </div>
      <el-form label-width="100px">
        <el-form-item label="密钥名称">
          <span class="text-slate-800">{{ createdKey?.name }}</span>
        </el-form-item>
        <el-form-item label="密钥前缀">
          <code class="bg-slate-100 px-2 py-1 rounded text-sm">{{ createdKey?.key_prefix }}****</code>
        </el-form-item>
        <el-form-item label="完整密钥">
          <div class="flex items-center gap-2 w-full">
            <el-input :model-value="createdKey?.secret || ''" readonly class="flex-1" />
            <el-button @click="copySecret">复制</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button type="primary" @click="secretVisible = false">我已保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, WarningFilled } from '@element-plus/icons-vue'
import apiKeyApi, { type APIKey, type CreateAPIKeyRequest } from '@/api/apiKey'

const loading = ref(false)
const keys = ref<APIKey[]>([])

const dialogVisible = ref(false)
const submitLoading = ref(false)
const formRef = ref()
const form = reactive({
  name: '',
  rate_limit: 1000,
  expires_at: ''
})
const formRules = {
  name: [{ required: true, message: '请输入密钥名称', trigger: 'blur' }]
}

const secretVisible = ref(false)
const createdKey = ref<APIKey | null>(null)

async function fetchKeys() {
  loading.value = true
  try {
    const res = await apiKeyApi.listAPIKeys()
    keys.value = Array.isArray(res.data) ? res.data : []
  } catch (err: any) {
    ElMessage.error(err.message || '获取密钥列表失败')
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  form.name = ''
  form.rate_limit = 1000
  form.expires_at = ''
  dialogVisible.value = true
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
  } catch { return }

  submitLoading.value = true
  try {
    const data: CreateAPIKeyRequest = {
      name: form.name,
      rate_limit: form.rate_limit
    }
    if (form.expires_at) data.expires_at = form.expires_at

    const res = await apiKeyApi.createAPIKey(data)
    dialogVisible.value = false
    createdKey.value = res.data || null
    secretVisible.value = true
    fetchKeys()
  } catch (err: any) {
    ElMessage.error(err.message || '创建失败')
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: APIKey) {
  try {
    await ElMessageBox.confirm('删除后使用此密钥的所有 API 调用将立即失效，确定要删除吗？', '确认删除', { type: 'warning' })
    await apiKeyApi.deleteAPIKey(row.id)
    ElMessage.success('删除成功')
    fetchKeys()
  } catch {}
}

function copySecret() {
  if (createdKey.value?.secret) {
    navigator.clipboard.writeText(createdKey.value.secret).then(() => {
      ElMessage.success('已复制到剪贴板')
    }).catch(() => {
      ElMessage.error('复制失败，请手动复制')
    })
  }
}

onMounted(() => {
  fetchKeys()
})
</script>

<style scoped>
</style>
