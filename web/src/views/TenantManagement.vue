<template>
  <div class="flex h-screen w-screen overflow-hidden text-slate-800 relative z-10">
    
    <!-- 左侧菜单栏 -->
    <aside class="glass-sidebar w-64 flex flex-col z-20">
      <!-- Logo 区域 -->
      <div class="h-16 flex items-center px-6 border-b border-slate-100/50">
        <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center text-white font-bold text-lg shadow-md shadow-indigo-500/30 mr-3">
          F
        </div>
        <h1 class="text-xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-slate-800 to-slate-600 tracking-tight">FayHub</h1>
      </div>

      <!-- 导航菜单 -->
      <div class="flex-1 py-6 px-3 space-y-1 overflow-y-auto">
        <div class="text-xs font-bold text-slate-400 mb-2 px-3 uppercase tracking-wider">系统管理</div>
        <div class="menu-item flex items-center px-4 py-2.5 rounded-xl cursor-pointer text-slate-600 font-medium text-sm" @click="$router.push('/dashboard')">
          <el-icon class="mr-3 text-lg"><Monitor /></el-icon> 仪表盘
        </div>
        <div class="menu-item active flex items-center px-4 py-2.5 rounded-xl cursor-pointer font-medium text-sm">
          <el-icon class="mr-3 text-lg"><OfficeBuilding /></el-icon> 租户管理
        </div>
        <div class="menu-item flex items-center px-4 py-2.5 rounded-xl cursor-pointer text-slate-600 font-medium text-sm" @click="$router.push('/users')">
          <el-icon class="mr-3 text-lg"><User /></el-icon> 用户管理
        </div>
        <div class="menu-item flex items-center px-4 py-2.5 rounded-xl cursor-pointer text-slate-600 font-medium text-sm">
          <el-icon class="mr-3 text-lg"><Lock /></el-icon> 角色权限
        </div>
        <div class="menu-item flex items-center px-4 py-2.5 rounded-xl cursor-pointer text-slate-600 font-medium text-sm">
          <el-icon class="mr-3 text-lg"><Setting /></el-icon> 系统设置
        </div>
      </div>
    </aside>

    <!-- 右侧主内容区 -->
    <main class="flex-1 flex flex-col min-w-0">
      
      <!-- 顶部导航栏 -->
      <header class="glass-header h-16 flex items-center justify-between px-8 z-10 sticky top-0">
        <!-- 面包屑 -->
        <div class="flex items-center text-sm font-medium text-slate-500">
          <span class="hover:text-indigo-600 cursor-pointer transition-colors" @click="$router.push('/dashboard')">首页</span>
          <el-icon class="mx-2 text-slate-400"><ArrowRight /></el-icon>
          <span class="text-slate-800 font-semibold">租户管理</span>
        </div>

        <!-- 右侧工具栏 -->
        <div class="flex items-center space-x-5">
          <el-button text>
            <el-icon><Search /></el-icon>
          </el-button>
          <el-button text>
            <el-icon><Bell /></el-icon>
          </el-button>
          <div class="h-5 w-px bg-slate-200"></div>
          <div class="flex items-center cursor-pointer group">
            <img :src="userInfo.avatar" alt="Avatar" class="w-8 h-8 rounded-full border-2 border-slate-100 group-hover:border-indigo-200 transition-all">
            <span class="ml-2 text-sm font-semibold text-slate-700 group-hover:text-indigo-600 transition-colors">{{ userInfo.nickname }}</span>
            <el-icon class="ml-1 text-slate-400 group-hover:text-indigo-600"><ArrowDown /></el-icon>
          </div>
        </div>
      </header>

      <!-- 核心页面内容 -->
      <div class="flex-1 overflow-y-auto p-8">
        
        <!-- 页面标题 -->
        <div class="mb-6 flex justify-between items-end">
          <div>
            <h2 class="text-2xl font-bold text-slate-800 tracking-tight">租户列表</h2>
            <p class="text-sm text-slate-500 mt-1">管理系统内所有租户账号及权限状态</p>
          </div>
        </div>

        <!-- 主体卡片 -->
        <div class="glass-card rounded-2xl p-6">
          
          <!-- 操作与搜索栏 -->
          <div class="flex flex-wrap justify-between items-center mb-6 gap-4">
            <!-- 左侧：危险/批量操作 -->
            <div class="flex gap-3">
              <el-button type="danger" plain @click="batchDisable" :disabled="selectedTenants.length === 0">
                <el-icon><Remove /></el-icon> 批量禁用
              </el-button>
              <el-button @click="batchEnable" :disabled="selectedTenants.length === 0">
                <el-icon><Check /></el-icon> 批量启用
              </el-button>
            </div>
            
            <!-- 右侧：核心正向操作 -->
            <div class="flex gap-3 items-center">
              <el-input 
                v-model="searchKeyword" 
                placeholder="搜索租户名称" 
                clearable 
                class="w-64"
              >
                <template #prefix>
                  <el-icon><Search /></el-icon>
                </template>
              </el-input>
              <el-button type="primary" @click="showAddDialog">
                <el-icon><Plus /></el-icon> 新增租户
              </el-button>
            </div>
          </div>

          <!-- 高级数据表格 -->
          <el-table
            v-model:selection="selectedTenants"
            :data="filteredTenants"
            style="width: 100%"
            @selection-change="handleSelectionChange"
          >
            <el-table-column type="selection" width="55" align="center"></el-table-column>
            
            <el-table-column prop="id" label="租户编号" width="130">
              <template #default="{ row }">
                <span class="font-mono text-xs text-slate-500 bg-slate-100 px-2 py-1 rounded">T-{{ row.id.toString().padStart(6, '0') }}</span>
              </template>
            </el-table-column>

            <el-table-column prop="name" label="租户信息" min-width="200">
              <template #default="{ row }">
                <div class="flex items-center">
                  <div class="w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center text-white font-bold text-sm mr-3">
                    {{ row.name.charAt(0).toUpperCase() }}
                  </div>
                  <div class="flex flex-col">
                    <span class="font-semibold text-slate-700">{{ row.name }}</span>
                    <span class="text-xs text-slate-400 mt-0.5">{{ row.code }}</span>
                  </div>
                </div>
              </template>
            </el-table-column>

            <el-table-column label="联系人信息" min-width="180">
              <template #default="{ row }">
                <div class="flex flex-col">
                  <span class="text-sm font-medium text-slate-700">{{ row.contactName }}</span>
                  <span class="text-xs text-slate-400 mt-0.5">
                    <el-icon class="mr-1"><Phone /></el-icon>{{ row.contactPhone }}
                  </span>
                </div>
              </template>
            </el-table-column>

            <el-table-column prop="userCount" label="用户数量" width="120" align="center">
              <template #default="{ row }">
                <span class="text-sm font-semibold text-slate-700">{{ row.userCount }}</span>
              </template>
            </el-table-column>

            <el-table-column prop="status" label="租户状态" width="120" align="center">
              <template #default="{ row }">
                <div :class="['status-badge', row.status === 1 ? 'enabled' : 'disabled']">
                  <span class="status-dot"></span>
                  {{ row.status === 1 ? '启用' : '禁用' }}
                </div>
              </template>
            </el-table-column>

            <el-table-column prop="createTime" label="创建时间" width="160">
              <template #default="{ row }">
                <span class="text-sm text-slate-500">{{ row.createTime }}</span>
              </template>
            </el-table-column>

            <el-table-column label="操作" width="200" align="center" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="editTenant(row)">
                  <el-icon><Edit /></el-icon>编辑
                </el-button>
                <el-button link type="default" @click="viewUsers(row)">
                  <el-icon><User /></el-icon>用户
                </el-button>
                <el-button 
                  link 
                  :type="row.status === 1 ? 'danger' : 'success'" 
                  @click="toggleTenantStatus(row)"
                >
                  <el-icon>{{ row.status === 1 ? 'Remove' : 'Check' }}</el-icon>
                  {{ row.status === 1 ? '禁用' : '启用' }}
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <!-- 底部精美分页 -->
          <div class="mt-8 flex justify-between items-center">
            <span class="text-sm text-slate-500">
              共发现 <span class="font-semibold text-slate-700">{{ total }}</span> 个租户记录
            </span>
            <el-pagination
              v-model:current-page="currentPage"
              v-model:page-size="pageSize"
              :page-sizes="[10, 20, 50, 100]"
              background
              layout="sizes, prev, pager, next"
              :total="total"
            ></el-pagination>
          </div>

        </div>
      </div>
    </main>

    <!-- 新增/编辑租户弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      :before-close="handleClose"
    >
      <el-form :model="tenantForm" :rules="tenantRules" ref="tenantFormRef" label-width="100px">
        <el-form-item label="租户名称" prop="name">
          <el-input v-model="tenantForm.name" placeholder="请输入租户名称" />
        </el-form-item>
        <el-form-item label="租户编码" prop="code">
          <el-input v-model="tenantForm.code" placeholder="请输入租户编码" />
        </el-form-item>
        <el-form-item label="联系人" prop="contactName">
          <el-input v-model="tenantForm.contactName" placeholder="请输入联系人姓名" />
        </el-form-item>
        <el-form-item label="联系电话" prop="contactPhone">
          <el-input v-model="tenantForm.contactPhone" placeholder="请输入联系电话" />
        </el-form-item>
        <el-form-item label="邮箱" prop="contactEmail">
          <el-input v-model="tenantForm.contactEmail" placeholder="请输入邮箱地址" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="tenantForm.status">
            <el-radio :label="1">启用</el-radio>
            <el-radio :label="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitTenantForm" :loading="loading">确认</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'

const router = useRouter()

// 用户信息
const userInfo = ref({
  nickname: '超级管理员',
  avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=admin&backgroundColor=e2e8f0'
})

// 搜索和分页
const searchKeyword = ref('')
const selectedTenants = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 弹窗控制
const dialogVisible = ref(false)
const dialogTitle = ref('')
const loading = ref(false)
const tenantFormRef = ref<FormInstance>()

// 租户表单数据
const tenantForm = reactive({
  id: 0,
  name: '',
  code: '',
  contactName: '',
  contactPhone: '',
  contactEmail: '',
  status: 1
})

// 表单验证规则
const tenantRules = {
  name: [{ required: true, message: '请输入租户名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入租户编码', trigger: 'blur' }],
  contactName: [{ required: true, message: '请输入联系人姓名', trigger: 'blur' }]
}

// 模拟租户数据
const tenants = ref([
  {
    id: 1,
    name: '星辉科技',
    code: 'XH-TECH',
    contactName: '张三',
    contactPhone: '138 0013 8001',
    contactEmail: 'zhangsan@xh-tech.com',
    userCount: 45,
    status: 1,
    createTime: '2024-04-20 10:30'
  },
  {
    id: 2,
    name: '云创数据',
    code: 'YC-DATA',
    contactName: '李四',
    contactPhone: '159 8821 3304',
    contactEmail: 'lisi@yc-data.com',
    userCount: 28,
    status: 0,
    createTime: '2024-04-18 14:15'
  },
  {
    id: 3,
    name: '智联网络',
    code: 'ZL-NET',
    contactName: '王五',
    contactPhone: '136 2288 9911',
    contactEmail: 'wangwu@zl-net.com',
    userCount: 67,
    status: 1,
    createTime: '2024-04-15 09:20'
  }
])

// 过滤后的租户列表
const filteredTenants = computed(() => {
  if (!searchKeyword.value) return tenants.value
  return tenants.value.filter(tenant => 
    tenant.name.includes(searchKeyword.value) || 
    tenant.code.includes(searchKeyword.value)
  )
})

// 方法定义
const handleSelectionChange = (val: any) => {
  selectedTenants.value = val
}

const showAddDialog = () => {
  dialogTitle.value = '新增租户'
  Object.assign(tenantForm, {
    id: 0,
    name: '',
    code: '',
    contactName: '',
    contactPhone: '',
    contactEmail: '',
    status: 1
  })
  dialogVisible.value = true
}

const editTenant = (row: any) => {
  dialogTitle.value = '编辑租户'
  Object.assign(tenantForm, row)
  dialogVisible.value = true
}

const viewUsers = (row: any) => {
  ElMessage.info(`查看租户 ${row.name} 的用户列表`)
}

const toggleTenantStatus = async (row: any) => {
  const action = row.status === 1 ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(`确定要${action}租户 "${row.name}" 吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    row.status = row.status === 1 ? 0 : 1
    ElMessage.success(`已成功${action}租户 ${row.name}`)
  } catch {
    // 用户取消操作
  }
}

const batchEnable = () => {
  if (selectedTenants.value.length === 0) return
  ElMessage.success(`已批量启用 ${selectedTenants.value.length} 个租户`)
}

const batchDisable = () => {
  if (selectedTenants.value.length === 0) return
  ElMessage.warning(`已批量禁用 ${selectedTenants.value.length} 个租户`)
}

const handleClose = (done: () => void) => {
  ElMessageBox.confirm('确定要关闭吗？未保存的更改将会丢失。', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    done()
  }).catch(() => {
    // 取消关闭
  })
}

const submitTenantForm = async () => {
  if (!tenantFormRef.value) return
  
  const valid = await tenantFormRef.value.validate()
  if (!valid) return
  
  loading.value = true
  
  // 模拟 API 调用
  setTimeout(() => {
    loading.value = false
    
    if (tenantForm.id === 0) {
      // 新增
      const newTenant = {
        ...tenantForm,
        id: Math.max(...tenants.value.map(t => t.id)) + 1,
        userCount: 0,
        createTime: new Date().toLocaleString()
      }
      tenants.value.unshift(newTenant)
      ElMessage.success('租户创建成功！')
    } else {
      // 编辑
      const index = tenants.value.findIndex(t => t.id === tenantForm.id)
      if (index !== -1) {
        tenants.value[index] = { ...tenants.value[index], ...tenantForm }
        ElMessage.success('租户信息更新成功！')
      }
    }
    
    dialogVisible.value = false
  }, 1000)
}

onMounted(() => {
  total.value = tenants.value.length
})
</script>

<style scoped>
.glass-sidebar {
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-right: 1px solid rgba(255, 255, 255, 0.6);
  box-shadow: 
    2px 0 8px rgba(0, 0, 0, 0.02),
    inset 0 0 0 1px rgba(255, 255, 255, 0.5);
}

.glass-header {
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.6);
  box-shadow: 
    0 2px 8px rgba(0, 0, 0, 0.03),
    inset 0 0 0 1px rgba(255, 255, 255, 0.5);
}

.menu-item {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border-radius: 10px;
  margin: 2px 8px;
}

.menu-item:hover:not(.active) {
  background-color: rgba(241, 245, 249, 0.8);
  color: #334155;
  transform: translateX(4px);
}

.menu-item.active {
  background: linear-gradient(135deg, #4f46e5, #3b82f6);
  color: white;
  box-shadow: 0 4px 12px rgba(79, 70, 229, 0.25);
}
</style>