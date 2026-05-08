<template>
  <div class="profile-page">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold text-slate-800">个人中心</h2>
        <p class="text-slate-500 mt-1 text-sm">管理您的账户信息和安全设置</p>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div class="lg:col-span-1">
        <div class="bg-white rounded-2xl border border-slate-100 shadow-sm p-6 text-center">
          <el-avatar :size="96" :src="profile.avatar || undefined" class="mb-4">
            <span class="text-3xl">{{ profile.real_name?.charAt(0) || profile.username?.charAt(0) || 'U' }}</span>
          </el-avatar>
          <h3 class="text-lg font-semibold text-slate-800">{{ profile.real_name || profile.username }}</h3>
          <p class="text-slate-500 text-sm mt-1">{{ roleLabel(profile.role) }}</p>
          <p class="text-slate-400 text-xs mt-1">{{ profile.email }}</p>
          <el-divider />
          <div class="text-left space-y-3">
            <div class="flex items-center text-sm">
              <el-icon class="mr-2 text-slate-400"><User /></el-icon>
              <span class="text-slate-500">用户名：</span>
              <span class="ml-auto text-slate-800">{{ profile.username }}</span>
            </div>
            <div class="flex items-center text-sm">
              <el-icon class="mr-2 text-slate-400"><Phone /></el-icon>
              <span class="text-slate-500">手机：</span>
              <span class="ml-auto text-slate-800">{{ profile.phone || '未设置' }}</span>
            </div>
            <div class="flex items-center text-sm">
              <el-icon class="mr-2 text-slate-400"><Clock /></el-icon>
              <span class="text-slate-500">注册时间：</span>
              <span class="ml-auto text-slate-800">{{ formatDate(profile.created_at) }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="lg:col-span-2 space-y-6">
        <div class="bg-white rounded-2xl border border-slate-100 shadow-sm p-6">
          <h3 class="text-lg font-semibold text-slate-800 mb-4">基本信息</h3>
          <el-form :model="infoForm" label-width="80px" style="max-width: 500px">
            <el-form-item label="真实姓名">
              <el-input v-model="infoForm.real_name" placeholder="请输入真实姓名" />
            </el-form-item>
            <el-form-item label="邮箱">
              <el-input v-model="infoForm.email" placeholder="请输入邮箱" />
            </el-form-item>
            <el-form-item label="手机号">
              <el-input v-model="infoForm.phone" placeholder="请输入手机号" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleUpdateInfo" :loading="infoSaving">保存修改</el-button>
            </el-form-item>
          </el-form>
        </div>

        <div class="bg-white rounded-2xl border border-slate-100 shadow-sm p-6">
          <h3 class="text-lg font-semibold text-slate-800 mb-4">修改密码</h3>
          <el-form :model="pwdForm" :rules="pwdRules" ref="pwdFormRef" label-width="100px" style="max-width: 500px">
            <el-form-item label="当前密码" prop="old_password">
              <el-input v-model="pwdForm.old_password" type="password" show-password placeholder="请输入当前密码" />
            </el-form-item>
            <el-form-item label="新密码" prop="new_password">
              <el-input v-model="pwdForm.new_password" type="password" show-password placeholder="请输入新密码（至少6位）" />
            </el-form-item>
            <el-form-item label="确认密码" prop="confirm_password">
              <el-input v-model="pwdForm.confirm_password" type="password" show-password placeholder="请再次输入新密码" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleChangePassword" :loading="pwdSaving">修改密码</el-button>
            </el-form-item>
          </el-form>
        </div>

        <div class="bg-white rounded-2xl border border-slate-100 shadow-sm p-6">
          <h3 class="text-lg font-semibold text-slate-800 mb-4">更换头像</h3>
          <div class="flex items-center gap-6">
            <el-avatar :size="64" :src="profile.avatar || undefined">
              <span class="text-xl">{{ profile.real_name?.charAt(0) || 'U' }}</span>
            </el-avatar>
            <el-upload
              :show-file-list="false"
              :before-upload="beforeAvatarUpload"
              :http-request="handleAvatarUpload"
              accept="image/*"
            >
              <el-button type="primary">上传头像</el-button>
            </el-upload>
            <span class="text-slate-400 text-sm">支持 JPG/PNG，不超过 2MB</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { User, Phone, Clock } from '@element-plus/icons-vue'
import type { FormInstance, FormRules, UploadRequestOptions } from 'element-plus'
import userApi, { type User as UserType } from '@/api/user'
import fileApi from '@/api/file'

const profile = ref<Partial<UserType>>({})
const infoSaving = ref(false)
const pwdSaving = ref(false)
const pwdFormRef = ref<FormInstance>()

const infoForm = reactive({
  real_name: '',
  email: '',
  phone: ''
})

const pwdForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const pwdRules: FormRules = {
  old_password: [{ required: true, message: '请输入当前密码', trigger: 'blur' }],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' }
  ],
  confirm_password: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        if (value !== pwdForm.new_password) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

function roleLabel(role?: string): string {
  switch (role) {
    case 'super_admin': return '超级管理员'
    case 'platform_admin': return '平台管理员'
    case 'tenant_admin': return '租户管理员'
    default: return role || '普通用户'
  }
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return '-'
  return dateStr.slice(0, 10)
}

async function fetchProfile() {
  try {
    const res = await userApi.getProfile()
    if (res.data) {
      profile.value = res.data
      infoForm.real_name = res.data.real_name || ''
      infoForm.email = res.data.email || ''
      infoForm.phone = res.data.phone || ''
    }
  } catch (e) { console.error('fetchProfile failed:', e); }
}

async function handleUpdateInfo() {
  infoSaving.value = true
  try {
    const userId = profile.value.id
    if (!userId) return
    await userApi.updateUser(userId, {
      real_name: infoForm.real_name,
      email: infoForm.email,
      phone: infoForm.phone
    })
    ElMessage.success('信息更新成功')
    fetchProfile()
  } catch (err: any) {
    ElMessage.error(err.message || '更新失败')
  } finally {
    infoSaving.value = false
  }
}

async function handleChangePassword() {
  try {
    await pwdFormRef.value?.validate()
  } catch { return }

  pwdSaving.value = true
  try {
    await userApi.changePassword({
      old_password: pwdForm.old_password,
      new_password: pwdForm.new_password
    })
    ElMessage.success('密码修改成功，请重新登录')
    pwdForm.old_password = ''
    pwdForm.new_password = ''
    pwdForm.confirm_password = ''
  } catch (err: any) {
    ElMessage.error(err.message || '密码修改失败')
  } finally {
    pwdSaving.value = false
  }
}

function beforeAvatarUpload(file: File) {
  const isImage = file.type.startsWith('image/')
  const isLt2M = file.size / 1024 / 1024 < 2
  if (!isImage) {
    ElMessage.error('只能上传图片文件')
    return false
  }
  if (!isLt2M) {
    ElMessage.error('图片大小不能超过 2MB')
    return false
  }
  return true
}

async function handleAvatarUpload(options: UploadRequestOptions) {
  try {
    const res = await fileApi.upload(options.file)
    if (res.data?.url && profile.value.id) {
      await userApi.updateUser(profile.value.id, {
        real_name: profile.value.real_name,
        email: profile.value.email,
        phone: profile.value.phone
      })
      profile.value.avatar = res.data.url
      ElMessage.success('头像更新成功')
    }
  } catch (err: any) {
    ElMessage.error(err.message || '头像上传失败')
  }
}

onMounted(() => {
  fetchProfile()
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
