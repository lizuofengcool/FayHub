<template>
  <div class="relative z-10 flex min-h-screen items-center justify-center p-4 sm:p-6 py-8">
    
    <!-- 核心容器 -->
    <div class="glass-card flex w-full max-w-[1080px] overflow-hidden rounded-[24px] fade-in-up">
      
      <!-- 左侧：品牌展示区 -->
      <div class="hidden lg:flex w-5/12 flex-col justify-between bg-slate-900 p-10 text-white relative overflow-hidden">
        <!-- 左侧点缀背景 -->
        <div class="absolute inset-0 opacity-20" style="background-image: radial-gradient(#4f46e5 1px, transparent 1px); background-size: 24px 24px;"></div>
        <div class="absolute -bottom-24 -left-24 w-96 h-96 bg-blue-500 rounded-full mix-blend-multiply filter blur-3xl opacity-30"></div>
        <div class="absolute -top-24 -right-24 w-96 h-96 bg-purple-500 rounded-full mix-blend-multiply filter blur-3xl opacity-30"></div>

        <!-- Logo -->
        <div class="relative z-10 flex items-center space-x-3">
          <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center font-bold text-xl shadow-lg shadow-indigo-500/40">
            F
          </div>
          <span class="text-2xl font-bold tracking-tight">FayHub</span>
        </div>

        <!-- Slogan -->
        <div class="relative z-10 my-auto fade-in-left" style="animation-delay: 0.2s">
          <h1 class="text-3xl font-extrabold leading-tight mb-5">
            构建 AI 原生<br/>
            <span class="text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-indigo-400">
              多租户 SaaS 生态底座
            </span>
          </h1>
          <p class="text-slate-400 text-base leading-relaxed max-w-sm">
            高性能架构 · 全链路多租户隔离 · 企业级安全管控。为您打造可扩展的一站式应用开发与运营平台。
          </p>
        </div>

        <!-- 底部声明 -->
        <div class="relative z-10 text-xs text-slate-500">
          &copy; 2026 <a target="_blank" href="https://github.com/lizuofengcool/FayHub">FayHub</a> Platform. All rights reserved.
        </div>
      </div>

      <!-- 右侧：登录表单区 -->
      <div class="w-full lg:w-7/12 flex flex-col justify-center px-8 sm:px-16 md:px-20 py-10 bg-white/40">
        
        <div class="w-full max-w-[340px] mx-auto">
          <!-- 移动端 Logo -->
          <div class="flex lg:hidden items-center justify-center space-x-3 mb-8">
            <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center text-white font-bold text-xl shadow-md">F</div>
            <span class="text-2xl font-bold text-slate-800">FayHub</span>
          </div>

          <div class="mb-6 text-center lg:text-left">
            <h2 class="text-2xl font-bold text-slate-800 tracking-tight">欢迎回来 👋</h2>
            <p class="text-slate-500 mt-2 text-sm">请输入您的管理员账号以继续访问系统</p>
          </div>

          <!-- 表单 -->
          <el-form 
            ref="loginFormRef"
            :model="loginForm"
            :rules="loginRules"
            class="space-y-4"
            @keyup.enter="handleLogin"
          >
            <!-- 账号 -->
            <el-form-item prop="username">
              <el-input 
                ref="usernameInput"
                v-model="loginForm.username" 
                placeholder="请输入登录账号" 
                clearable
                @keyup.enter="focusNextField('password')"
              >
                <template #prefix><i class="ri-user-3-line text-slate-400 text-lg"></i></template>
              </el-input>
            </el-form-item>

            <!-- 密码 -->
            <el-form-item prop="password">
              <el-input 
                ref="passwordInput"
                v-model="loginForm.password" 
                type="password" 
                placeholder="请输入登录密码" 
                show-password
                @keyup.enter="focusNextField('captcha')"
              >
                <template #prefix><i class="ri-lock-password-line text-slate-400 text-lg"></i></template>
              </el-input>
            </el-form-item>

            <!-- 验证码 -->
            <el-form-item prop="captcha">
              <div class="flex w-full space-x-3">
                <el-input 
                  ref="captchaInput"
                  v-model="loginForm.captcha" 
                  placeholder="验证码" 
                  class="flex-1"
                  @keyup.enter="handleLogin"
                >
                  <template #prefix><i class="ri-shield-check-line text-slate-400 text-lg"></i></template>
                </el-input>
                <div class="w-28 h-[48px] rounded-lg bg-slate-100 border border-slate-200 flex items-center justify-center cursor-pointer hover:bg-slate-200 transition-colors" title="点击刷新" @click="refreshCaptcha">
                  <span class="font-mono text-xl tracking-widest text-indigo-600 font-bold italic" style="text-shadow: 1px 1px 2px rgba(0,0,0,0.1); transform: skewX(-10deg);">{{ captchaText }}</span>
                </div>
              </div>
            </el-form-item>

            <!-- 记住我 & 忘记密码 -->
            <div class="flex items-center justify-between pb-2 pt-1">
              <el-checkbox v-model="loginForm.remember" class="text-slate-500">记住账号</el-checkbox>
              <a href="#" class="text-sm font-semibold text-indigo-600 hover:text-indigo-500 transition-colors">忘记密码?</a>
            </div>

            <!-- 登录按钮 -->
            <el-button type="primary" class="w-full btn-login" :loading="loading" @click="handleLogin">
              {{ loading ? '登 录 中...' : '立 即 登 录' }}
            </el-button>

          </el-form>

          <!-- 注册链接 -->
          <div class="mt-8 text-center">
            <span class="text-slate-500 text-sm">还没有账号？</span>
            <a href="javascript:void(0)" class="text-sm font-semibold text-indigo-600 hover:text-indigo-500 transition-colors ml-1" @click="showRegister = true">立即注册</a>
          </div>

          <!-- 第三方登录 -->
          <div class="mt-6">
            <div class="relative">
              <div class="absolute inset-0 flex items-center"><div class="w-full border-t border-slate-200"></div></div>
              <div class="relative flex justify-center text-sm">
                <span class="px-4 text-slate-400 bg-transparent">或使用其他方式</span>
              </div>
            </div>

            <div class="mt-5 grid grid-cols-3 gap-4">
              <button class="flex justify-center items-center py-2.5 px-4 border border-slate-200 rounded-xl hover:bg-slate-50 hover:border-slate-300 transition-all text-slate-600 hover:text-blue-500">
                <i class="ri-wechat-fill text-lg"></i>
              </button>
              <button class="flex justify-center items-center py-2.5 px-4 border border-slate-200 rounded-xl hover:bg-slate-50 hover:border-slate-300 transition-all text-slate-600 hover:text-blue-600">
                <i class="ri-dingding-fill text-lg"></i>
              </button>
              <button class="flex justify-center items-center py-2.5 px-4 border border-slate-200 rounded-xl hover:bg-slate-50 hover:border-slate-300 transition-all text-slate-600 hover:text-slate-900">
                <i class="ri-github-fill text-lg"></i>
              </button>
            </div>
          </div>

        </div>
      </div>
    </div>
  </div>
  
  <!-- 注册对话框 -->
  <el-dialog v-model="showRegister" title="注册新账号" width="500px" :close-on-click-modal="false">
    <el-form ref="registerFormRef" :model="registerForm" :rules="registerRules" label-width="80px">
      <el-form-item label="用户名" prop="username">
        <el-input v-model="registerForm.username" placeholder="请输入用户名" />
      </el-form-item>
      <el-form-item label="密码" prop="password">
        <el-input v-model="registerForm.password" type="password" placeholder="请输入密码" show-password />
      </el-form-item>
      <el-form-item label="邮箱" prop="email">
        <el-input v-model="registerForm.email" placeholder="请输入邮箱" />
      </el-form-item>
      <el-form-item label="手机号" prop="phone">
        <el-input v-model="registerForm.phone" placeholder="请输入手机号" />
      </el-form-item>
      <el-form-item label="真实姓名" prop="real_name">
        <el-input v-model="registerForm.real_name" placeholder="请输入真实姓名" />
      </el-form-item>
      <el-form-item label="验证码" prop="captcha">
        <div class="flex space-x-3">
          <el-input v-model="registerForm.captcha" placeholder="请输入验证码" class="flex-1" />
          <div class="w-28 h-[40px] rounded bg-slate-100 border border-slate-200 flex items-center justify-center cursor-pointer hover:bg-slate-200 transition-colors" @click="refreshCaptcha">
            <span class="font-mono text-lg tracking-widest text-indigo-600 font-bold italic">{{ captchaText }}</span>
          </div>
        </div>
      </el-form-item>
    </el-form>
    
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="showRegister = false">取消</el-button>
        <el-button type="primary" :loading="registerLoading" @click="handleRegister">注册</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance } from 'element-plus'
import authApi from '@/api/auth'

const router = useRouter()

const loginFormRef = ref<FormInstance>()
const registerFormRef = ref<FormInstance>()
const usernameInput = ref()
const passwordInput = ref()
const captchaInput = ref()
const loading = ref(false)
const registerLoading = ref(false)
const showRegister = ref(false)
const captchaText = ref('')

// 在 onMounted 中初始化验证码，避免在 ref 初始化时调用未定义的函数
import { onMounted } from 'vue'

onMounted(() => {
  captchaText.value = generateCaptcha()
})

const loginForm = reactive({
  username: '',
  password: '',
  captcha: '',
  remember: true
})

const registerForm = reactive({
  username: '',
  password: '',
  email: '',
  phone: '',
  real_name: '',
  captcha: '',
  tenant_id: 1 // 默认租户ID
})

const loginRules = {
  username: [{ required: true, message: '请输入登录账号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  captcha: [{ required: true, message: '请输入验证码', trigger: 'blur' }]
}

const registerRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度3-20个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度6-20个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' }
  ],
  real_name: [
    { required: true, message: '请输入真实姓名', trigger: 'blur' }
  ],
  captcha: [
    { required: true, message: '请输入验证码', trigger: 'blur' }
  ]
}

// 生成验证码函数
const generateCaptcha = () => {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789'
  let newCaptcha = ''
  for (let i = 0; i < 4; i++) {
    newCaptcha += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  return newCaptcha
}

// 刷新验证码
const refreshCaptcha = () => {
  captchaText.value = generateCaptcha()
  loginForm.captcha = '' // 清空输入框
  registerForm.captcha = '' // 清空注册输入框
}

// 自动跳转到下一个输入框
const focusNextField = (fieldName: string) => {
  switch (fieldName) {
    case 'password':
      if (passwordInput.value) {
        passwordInput.value.focus()
      }
      break
    case 'captcha':
      if (captchaInput.value) {
        captchaInput.value.focus()
      }
      break
    case 'login':
      handleLogin()
      break
  }
}

// 检查数据库连接状态
const checkDatabaseStatus = async () => {
  try {
    const response = await fetch('/api/health', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    })
    
    if (response.status === 200) {
      return { connected: true, message: '数据库连接正常' }
    } else if (response.status === 401) {
      return { connected: true, message: '服务运行中，需要登录' }
    } else {
      return { connected: false, message: '服务异常，请检查后端服务' }
    }
  } catch (error) {
    return { 
      connected: false, 
      message: '无法连接到后端服务，请检查：\n1. 后端服务是否启动\n2. 数据库服务是否运行\n3. 网络连接是否正常' 
    }
  }
}

const handleLogin = async () => {
  if (!loginFormRef.value) return
  
  // 防止重复点击
  if (loading.value) return
  
  // 先检查数据库连接状态
  const dbStatus = await checkDatabaseStatus()
  if (!dbStatus.connected) {
    ElMessage.error(`数据库连接失败\n${dbStatus.message}`)
    return
  }
  
  const valid = await loginFormRef.value.validate()
  if (!valid) return
  
  // 前端验证码验证
  if (loginForm.captcha !== captchaText.value) {
    ElMessage.error('验证码错误')
    return
  }
  
  loading.value = true
  
  try {
    const res = await authApi.login({
      username: loginForm.username,
      password: loginForm.password,
      captcha: loginForm.captcha
    })
    
    // 确保数据存在
    if (!res.data || !res.data.token) {
      throw new Error('登录响应数据异常')
    }
    
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('user', JSON.stringify({
      id: res.data.user_id || 0,
      username: res.data.username || '未知用户',
      nickname: res.data.username || '未知用户',
      role: res.data.role || 'user'
    }))
    
    ElMessage({
      message: `欢迎回来，${res.data.username}！`,
      type: 'success',
      duration: 2000
    })
    
    // 延迟跳转，确保消息显示完整
    setTimeout(() => {
      router.push('/dashboard')
    }, 500)
    
  } catch (error: any) {
    console.error('登录失败:', error)
    
    // 更精确的错误处理
    let errorMessage = '登录失败，请检查账号密码'
    
    if (error.response) {
      // HTTP错误
      if (error.response.status === 401) {
        errorMessage = '用户名或密码错误'
      } else if (error.response.status === 403) {
        errorMessage = '账号已被禁用'
      } else if (error.response.data && error.response.data.msg) {
        errorMessage = error.response.data.msg
      }
    } else if (error.message) {
      // 网络错误或API错误
      if (error.message.includes('Network Error') || error.message.includes('timeout')) {
        errorMessage = '网络连接失败，请检查网络设置'
      } else {
        errorMessage = error.message
      }
    }
    
    // 显示错误提示
    ElMessage.error(errorMessage)
    
    // 刷新验证码
    refreshCaptcha()
    
  } finally {
    // 延迟重置loading状态，避免闪烁
    setTimeout(() => {
      loading.value = false
    }, 300)
  }
}

// 处理注册
const handleRegister = async () => {
  if (!registerFormRef.value) return
  
  const valid = await registerFormRef.value.validate()
  if (!valid) return
  
  // 验证码验证
  if (registerForm.captcha !== captchaText.value) {
    ElMessage.error('验证码错误')
    return
  }
  
  registerLoading.value = true
  
  try {
    const res = await authApi.register({
      username: registerForm.username,
      password: registerForm.password,
      email: registerForm.email,
      phone: registerForm.phone,
      real_name: registerForm.real_name,
      tenant_id: registerForm.tenant_id
    })
    
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('user', JSON.stringify({
      id: res.data.user_id,
      username: res.data.username,
      nickname: res.data.username,
      role: res.data.role
    }))
    
    ElMessage({
      message: `注册成功，欢迎 ${res.data.username}！`,
      type: 'success',
      duration: 2000
    })
    
    showRegister.value = false
    router.push('/dashboard')
  } catch (error: any) {
    console.error('注册失败:', error)
    if (error.response && error.response.data) {
      ElMessage.error(error.response.data.msg || '注册失败')
    } else if (error.message) {
      ElMessage.error(error.message)
    } else {
      ElMessage.error('注册失败，请稍后重试')
    }
  } finally {
    registerLoading.value = false
  }
}
</script>
