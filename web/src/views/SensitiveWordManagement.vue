<template>
  <div class="sensitive-word-page">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold text-slate-800">敏感词管理</h2>
        <p class="text-slate-500 mt-1 text-sm">管理敏感词库，支持 DFA 高效匹配过滤，用于内容审核、违规词检测等场景</p>
      </div>
      <div class="flex gap-2">
        <el-tooltip content="当添加/修改大量敏感词后，重建匹配器可提高过滤效率" placement="top">
          <el-button @click="handleRebuild">
            <el-icon class="mr-1"><Refresh /></el-icon>
            重建匹配器
          </el-button>
        </el-tooltip>
        <el-button type="info" @click="downloadTemplate">
          <el-icon class="mr-1"><Download /></el-icon>
          下载模板
        </el-button>
        <el-button type="success" @click="batchDialogVisible = true">
          <el-icon class="mr-1"><DocumentAdd /></el-icon>
          批量导入
        </el-button>
        <el-button type="primary" @click="openDialog()">
          <el-icon class="mr-1"><Plus /></el-icon>
          新增敏感词
        </el-button>
      </div>
    </div>

    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 border-b border-slate-100 flex gap-3 flex-wrap">
        <el-input v-model="filters.keyword" placeholder="搜索敏感词" clearable style="width: 200px" />
        <el-input v-model="filters.category" placeholder="分类标签" clearable style="width: 150px" />
        <el-select v-model="filters.level" placeholder="等级" clearable style="width: 100px">
          <el-option label="一级" :value="1" />
          <el-option label="二级" :value="2" />
          <el-option label="三级" :value="3" />
        </el-select>
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>

      <el-table v-loading="loading" :data="words" stripe class="w-full">
        <el-table-column prop="word" label="敏感词" min-width="180" show-overflow-tooltip />
        <el-table-column prop="category" label="分类" width="120" show-overflow-tooltip />
        <el-table-column prop="level" label="等级" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="levelTagType(row.level)" size="small">{{ 'L' + row.level }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170" />
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDialog(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="p-4 flex justify-end">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @current-change="fetchWords"
          @size-change="fetchWords"
        />
      </div>
    </div>

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑敏感词' : '新增敏感词'" width="500px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="敏感词" required>
          <el-input v-model="form.word" placeholder="请输入敏感词" />
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="form.category" placeholder="如：政治、色情、暴力" />
        </el-form-item>
        <el-form-item label="等级">
          <el-select v-model="form.level" style="width: 100%">
            <el-option label="一级（直接替换）" :value="1" />
            <el-option label="二级（审核拦截）" :value="2" />
            <el-option label="三级（记录告警）" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="form.id" label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSave">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="batchDialogVisible" title="批量导入敏感词" width="500px">
      <el-form label-width="80px">
        <el-form-item label="敏感词">
          <el-input v-model="batchWords" type="textarea" :rows="8" placeholder="每行一个敏感词" />
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="batchCategory" placeholder="统一分类标签" />
        </el-form-item>
        <el-form-item label="等级">
          <el-select v-model="batchLevel" style="width: 100%">
            <el-option label="一级（直接替换）" :value="1" />
            <el-option label="二级（审核拦截）" :value="2" />
            <el-option label="三级（记录告警）" :value="3" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="batchDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="batchLoading" @click="handleBatchCreate">导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, DocumentAdd, Download } from '@element-plus/icons-vue'
import sensitiveWordApi, { type SensitiveWord } from '@/api/sensitiveWord'

const loading = ref(false)
const words = ref<SensitiveWord[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const filters = reactive({
  keyword: '',
  category: '',
  level: undefined as number | undefined
})

const dialogVisible = ref(false)
const submitLoading = ref(false)
const form = reactive<Partial<SensitiveWord>>({
  id: undefined,
  word: '',
  category: '',
  level: 1,
  status: 1
})

const batchDialogVisible = ref(false)
const batchLoading = ref(false)
const batchWords = ref('')
const batchCategory = ref('')
const batchLevel = ref(1)

function levelTagType(level: number) {
  if (level === 1) return 'danger'
  if (level === 2) return 'warning'
  return 'info'
}

function downloadTemplate() {
  // 添加 BOM 前缀，解决 Excel 打开中文乱码问题
  const BOM = '\uFEFF'
  const template = `${BOM}敏感词,分类,等级
示例敏感词1,政治,1
示例敏感词2,色情,2
示例敏感词3,广告,3`
  const blob = new Blob([template], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = '敏感词导入模板.csv'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
  ElMessage.success('模板下载成功，请按格式填写后导入')
}

async function fetchWords() {
  loading.value = true
  try {
    const params: any = {
      page: page.value,
      page_size: pageSize.value
    }
    if (filters.keyword) params.keyword = filters.keyword
    if (filters.category) params.category = filters.category
    if (filters.level) params.level = filters.level

    const res = await sensitiveWordApi.list(params)
    words.value = res.data.list
    total.value = res.data.total
  } catch (e: any) {
    ElMessage.error(e?.message || '查询失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  page.value = 1
  fetchWords()
}

function resetFilters() {
  filters.keyword = ''
  filters.category = ''
  filters.level = undefined
  handleSearch()
}

function openDialog(row?: SensitiveWord) {
  if (row) {
    form.id = row.id
    form.word = row.word
    form.category = row.category
    form.level = row.level
    form.status = row.status
  } else {
    form.id = undefined
    form.word = ''
    form.category = ''
    form.level = 1
    form.status = 1
  }
  dialogVisible.value = true
}

async function handleSave() {
  if (!form.word) {
    ElMessage.warning('请输入敏感词')
    return
  }

  submitLoading.value = true
  try {
    if (form.id) {
      await sensitiveWordApi.update(form.id, {
        word: form.word,
        category: form.category,
        level: form.level,
        status: form.status
      })
      ElMessage.success('更新成功')
    } else {
      await sensitiveWordApi.create({
        word: form.word!,
        category: form.category,
        level: form.level
      })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchWords()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: SensitiveWord) {
  try {
    await ElMessageBox.confirm(`确定删除敏感词「${row.word}」吗？`, '确认删除', {
      type: 'warning'
    })
    await sensitiveWordApi.delete(row.id)
    ElMessage.success('删除成功')
    fetchWords()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
  }
}

async function handleBatchCreate() {
  const lines = batchWords.value
    .split('\n')
    .map(s => s.trim())
    .filter(s => s.length > 0)

  if (lines.length === 0) {
    ElMessage.warning('请输入敏感词')
    return
  }

  batchLoading.value = true
  try {
    const res = await sensitiveWordApi.batchCreate({
      words: lines,
      category: batchCategory.value,
      level: batchLevel.value
    })
    ElMessage.success(`成功导入 ${res.data.count} 个敏感词`)
    batchDialogVisible.value = false
    batchWords.value = ''
    fetchWords()
  } catch (e: any) {
    ElMessage.error(e?.message || '导入失败')
  } finally {
    batchLoading.value = false
  }
}

async function handleRebuild() {
  try {
    await sensitiveWordApi.rebuild()
    ElMessage.success('匹配器重建成功')
  } catch (e: any) {
    ElMessage.error(e?.message || '重建失败')
  }
}

onMounted(() => {
  fetchWords()
})
</script>
