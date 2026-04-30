<template>
  <div class="menu-page">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold text-slate-800">菜单管理</h2>
        <p class="text-slate-500 mt-1 text-sm">管理系统菜单，支持树形结构</p>
      </div>
      <el-button type="primary" @click="openCreateDialog(0)">
        <el-icon class="mr-1"><Plus /></el-icon> 新建根菜单
      </el-button>
    </div>

    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm p-6">
      <el-table v-loading="loading" :data="menuTree" row-key="id" :tree-props="{ children: 'children' }" default-expand-all>
        <el-table-column prop="title" label="菜单名称" min-width="180" />
        <el-table-column prop="path" label="路由路径" min-width="160" />
        <el-table-column prop="component" label="组件路径" min-width="160" />
        <el-table-column prop="icon" label="图标" width="100" align="center">
          <template #default="{ row }">
            <el-icon v-if="iconComponentMap[row.icon]"><component :is="iconComponentMap[row.icon]" /></el-icon>
            <span v-else class="text-slate-300">{{ row.icon || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="layout" label="布局模式" width="110" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.layout === 'fullscreen'" type="warning" size="small">全屏模式</el-tag>
            <el-tag v-else type="info" size="small">内嵌模式</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openCreateDialog(row.id)">添加子菜单</el-button>
            <el-button type="warning" link size="small" @click="openEditDialog(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑菜单' : '新建菜单'" width="560px" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-form-item label="上级菜单">
          <el-input :value="parentName" disabled />
        </el-form-item>
        <el-form-item label="菜单名称" prop="title">
          <el-input v-model="form.title" placeholder="请输入菜单名称" />
        </el-form-item>
        <el-form-item label="菜单类型" prop="type">
          <el-radio-group v-model="form.type">
            <el-radio :value="1">目录</el-radio>
            <el-radio :value="2">菜单</el-radio>
            <el-radio :value="3">按钮</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="路由路径" prop="path">
          <el-input v-model="form.path" placeholder="如 /users" />
        </el-form-item>
        <el-form-item label="组件路径" prop="component">
          <el-input v-model="form.component" placeholder="如 views/UserManagement" />
        </el-form-item>
        <el-form-item label="图标">
          <el-input v-model="form.icon" placeholder="如 User, Setting" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :max="9999" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="布局模式" v-if="form.type === 2">
          <el-radio-group v-model="form.layout">
            <el-radio value="embedded">内嵌模式（保留侧边栏）</el-radio>
            <el-radio value="fullscreen">全屏模式（独占界面）</el-radio>
          </el-radio-group>
          <div class="text-xs text-slate-400 mt-1.5 leading-relaxed">
            <template v-if="form.layout === 'fullscreen'">
              全屏模式：插件独占整个界面，适合大型独立应用（如商城、CRM）
            </template>
            <template v-else>
              内嵌模式：在后台框架内显示，保留导航栏和侧边栏
            </template>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  Monitor, OfficeBuilding, User, Setting, Box, Tickets,
  Connection, List, CreditCard, Wallet, Management
} from '@element-plus/icons-vue'
import menuApi, { type Menu, type CreateMenuParams, type UpdateMenuParams } from '@/api/menu'

const iconComponentMap: Record<string, any> = {
  'Monitor': Monitor,
  'OfficeBuilding': OfficeBuilding,
  'User': User,
  'Setting': Setting,
  'Box': Box,
  'Tickets': Tickets,
  'Connection': Connection,
  'List': List,
  'CreditCard': CreditCard,
  'Wallet': Wallet,
  'Management': Management
}

const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref<number>(0)
const parentId = ref<number>(0)
const formRef = ref<FormInstance>()
const menuTree = ref<Menu[]>([])

const form = reactive<CreateMenuParams>({
  parent_id: 0,
  title: '',
  path: '',
  component: '',
  icon: '',
  sort: 0,
  type: 1,
  status: 1,
  permission: '',
  layout: 'embedded'
})

const formRules = {
  title: [{ required: true, message: '请输入菜单名称', trigger: 'blur' }],
  path: [{ required: true, message: '请输入路由路径', trigger: 'blur' }]
}

const parentName = computed(() => {
  if (parentId.value === 0) return '根菜单'
  return findMenuName(menuTree.value, parentId.value) || '未知'
})

function findMenuName(menus: Menu[], id: number): string | null {
  for (const m of menus) {
    if (m.id === id) return m.title
    if (m.children) {
      const found = findMenuName(m.children, id)
      if (found) return found
    }
  }
  return null
}

onMounted(() => { fetchTree() })

async function fetchTree() {
  loading.value = true
  try {
    const res = await menuApi.getMenuTree()
    menuTree.value = res.data || []
  } catch (err: any) {
    ElMessage.error(err.message || '获取菜单树失败')
  } finally {
    loading.value = false
  }
}

function openCreateDialog(pid: number) {
  isEdit.value = false
  editId.value = 0
  parentId.value = pid
  Object.assign(form, {
    parent_id: pid,
    title: '',
    path: '',
    component: '',
    icon: '',
    sort: 0,
    type: 1,
    status: 1,
    permission: '',
    layout: 'embedded'
  })
  dialogVisible.value = true
}

function openEditDialog(row: Menu) {
  isEdit.value = true
  editId.value = row.id
  parentId.value = row.parent_id
  Object.assign(form, {
    parent_id: row.parent_id,
    title: row.title,
    path: row.path,
    component: row.component,
    icon: row.icon,
    sort: row.sort,
    type: row.type,
    status: row.status,
    permission: row.permission,
    layout: (row.layout as 'embedded' | 'fullscreen') || 'embedded'
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (isEdit.value) {
      const params: UpdateMenuParams = {
        title: form.title,
        path: form.path,
        component: form.component,
        icon: form.icon,
        sort: form.sort,
        type: form.type,
        status: form.status,
        permission: form.permission,
        layout: form.layout
      }
      await menuApi.updateMenu(editId.value, params)
      ElMessage.success('菜单更新成功')
    } else {
      await menuApi.createMenu(form)
      ElMessage.success('菜单创建成功')
    }
    dialogVisible.value = false
    fetchTree()
  } catch (err: any) {
    ElMessage.error(err.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: Menu) {
  if (row.children && row.children.length > 0) {
    ElMessage.warning('请先删除子菜单')
    return
  }
  try {
    await ElMessageBox.confirm(`确定要删除菜单「${row.title}」吗？`, '警告', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'error'
    })
    await menuApi.deleteMenu(row.id)
    ElMessage.success('删除成功')
    fetchTree()
  } catch {}
}
</script>
