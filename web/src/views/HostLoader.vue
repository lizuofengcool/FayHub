<template>
  <div
    ref="hostRef"
    :data-plugin-container="pluginId"
    class="host-loader"
  >
    <div v-if="status === 'loading'" class="flex items-center justify-center py-20">
      <el-icon class="text-3xl text-slate-300 animate-spin"><Loading /></el-icon>
      <p class="text-slate-400 ml-3">加载插件中...</p>
    </div>

    <div v-else-if="status === 'error'" class="flex flex-col items-center justify-center py-20">
      <el-icon class="text-5xl text-red-300 mb-4"><WarningFilled /></el-icon>
      <p class="text-red-500 text-lg mb-2">插件加载失败</p>
      <p class="text-slate-400 text-sm mb-4">{{ errorMessage }}</p>
      <el-button type="primary" size="small" @click="retryLoad">重新加载</el-button>
    </div>

    <template v-else-if="status === 'loaded'">
      <template v-if="renderMode === 'custom' && dynamicComponent">
        <component
          :is="dynamicComponent"
          v-bind="componentProps"
          @plugin-event="handlePluginEvent"
        />
      </template>

      <template v-else-if="renderMode === 'schema'">
        <div class="schema-renderer">
          <div v-if="schemaConfig?.stats" class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
            <div
              v-for="stat in schemaConfig.stats"
              :key="stat.key"
              class="glass-card rounded-xl p-5 flex items-center gap-4"
            >
              <div
                :class="['w-12 h-12 rounded-xl flex items-center justify-center', statBgClass(stat.color)]"
              >
                <span :class="['text-2xl', statTextClass(stat.color)]">
                  {{ stat.icon || '📊' }}
                </span>
              </div>
              <div>
                <p class="text-2xl font-bold text-slate-800">{{ statValue(stat.key) }}</p>
                <p class="text-sm text-slate-500">{{ stat.label }}</p>
              </div>
            </div>
          </div>

          <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
            <div
              v-if="schemaConfig?.columns"
              class="p-4 border-b border-slate-100 flex items-center gap-3"
            >
              <el-input
                v-model="schemaSearch"
                placeholder="搜索..."
                clearable
                class="w-64"
              />
              <el-select
                v-if="schemaStatusField"
                v-model="schemaStatusFilter"
                placeholder="状态筛选"
                clearable
                class="w-36"
              >
                <el-option
                  v-for="opt in schemaStatusOptions"
                  :key="opt.value"
                  :label="opt.label"
                  :value="opt.value"
                />
              </el-select>
              <div class="flex-1" />
              <el-button
                v-if="schemaHasCreate"
                type="primary"
                @click="openSchemaDialog()"
              >
                {{ schemaCreateLabel }}
              </el-button>
            </div>

            <el-table
              v-if="schemaConfig?.type === 'table' && schemaFilteredItems.length > 0"
              :data="schemaFilteredItems"
              stripe
              class="w-full"
            >
              <el-table-column
                v-for="col in schemaConfig.columns"
                :key="col.key"
                :prop="col.key"
                :label="col.label"
                :width="col.width"
                :min-width="col.width ? undefined : 120"
              >
                <template #default="{ row }">
                  <template v-if="col.type === 'tag' && col.options">
                    <el-tag :type="schemaTagType(row[col.key], col)" size="small">
                      {{ schemaTagLabel(row[col.key], col) }}
                    </el-tag>
                  </template>
                  <template v-else>
                    {{ row[col.key] ?? '-' }}
                  </template>
                </template>
              </el-table-column>
              <el-table-column
                v-if="schemaHasRowActions"
                label="操作"
                width="200"
                fixed="right"
                align="center"
              >
                <template #default="{ row }">
                  <el-button
                    v-if="schemaHasEdit"
                    type="primary"
                    link
                    size="small"
                    @click="openSchemaDialog(row)"
                  >
                    编辑
                  </el-button>
                  <el-button
                    v-if="schemaHasDelete"
                    type="danger"
                    link
                    size="small"
                    @click="handleSchemaDelete(row)"
                  >
                    删除
                  </el-button>
                </template>
              </el-table-column>
            </el-table>

            <div
              v-else-if="schemaConfig?.type === 'html'"
              class="p-8"
              v-html="schemaData.html || ''"
            />

            <div v-else class="text-center py-16 text-slate-400">
              <p class="text-lg">暂无数据</p>
            </div>
          </div>

          <el-dialog
            v-model="schemaDialogVisible"
            :title="schemaEditingItem ? '编辑' : schemaCreateLabel"
            width="640px"
            destroy-on-close
          >
            <el-form
              ref="schemaFormRef"
              :model="schemaFormData"
              :rules="schemaFormRules"
              label-width="90px"
            >
              <template v-for="field in schemaConfig?.form" :key="field.key">
                <el-form-item :label="field.label" :prop="field.key">
                  <el-input
                    v-if="field.type === 'input'"
                    v-model="schemaFormData[field.key]"
                    :placeholder="'请输入' + field.label"
                    :maxlength="field.maxlength"
                    :show-word-limit="!!field.maxlength"
                  />
                  <el-input
                    v-else-if="field.type === 'textarea'"
                    v-model="schemaFormData[field.key]"
                    type="textarea"
                    :rows="field.rows || 4"
                    :placeholder="'请输入' + field.label"
                    :maxlength="field.maxlength"
                    :show-word-limit="!!field.maxlength"
                  />
                  <el-switch
                    v-else-if="field.type === 'switch'"
                    v-model="schemaFormData[field.key]"
                    active-text="是"
                    inactive-text="否"
                  />
                  <el-radio-group
                    v-else-if="field.type === 'radio'"
                    v-model="schemaFormData[field.key]"
                  >
                    <el-radio
                      v-for="opt in field.options"
                      :key="opt.value"
                      :value="opt.value"
                    >{{ opt.label }}</el-radio>
                  </el-radio-group>
                </el-form-item>
              </template>
            </el-form>
            <template #footer>
              <el-button @click="schemaDialogVisible = false">取消</el-button>
              <el-button
                type="primary"
                :loading="schemaSaveLoading"
                @click="handleSchemaSubmit"
              >
                保存
              </el-button>
            </template>
          </el-dialog>
        </div>
      </template>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch, reactive, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Loading, WarningFilled } from '@element-plus/icons-vue'
import { loadPlugin, unloadPlugin, getLoadedPlugin, type PluginManifest, type LoadedPlugin } from '@/plugin/loader'
import { createShadowContainer, injectStyleToShadow, injectElementPlusStylesToShadow, removePluginStyle } from '@/plugin/style-isolation'
import { useUserStore } from '@/stores/user'
import request from '@/api/request'

export interface HostLoaderProps {
  pluginId: string
  renderMode: 'custom' | 'schema'
  useShadowDom?: boolean
  manifest?: Partial<PluginManifest>
  componentProps?: Record<string, any>
}

const props = withDefaults(defineProps<HostLoaderProps>(), {
  useShadowDom: false,
  manifest: () => ({}),
  componentProps: () => ({}),
})

const emit = defineEmits<{
  loaded: [pluginId: string]
  error: [pluginId: string, error: string]
  unloaded: [pluginId: string]
}>()

const router = useRouter()
const userStore = useUserStore()

const hostRef = ref<HTMLElement | null>(null)
const status = ref<'idle' | 'loading' | 'loaded' | 'error'>('idle')
const errorMessage = ref('')
const dynamicComponent = ref<any>(null)
const shadowRoot = ref<ShadowRoot | null>(null)

const schemaConfig = ref<any>(null)
const schemaData = ref<any>({})
const schemaSearch = ref('')
const schemaStatusFilter = ref('')
const schemaDialogVisible = ref(false)
const schemaEditingItem = ref<any>(null)
const schemaSaveLoading = ref(false)
const schemaFormRef = ref<FormInstance>()
const schemaFormData = reactive<Record<string, any>>({})
const schemaFormRules = reactive<FormRules>({})

const schemaStatusField = computed(() => {
  if (!schemaConfig.value?.columns) return null
  return schemaConfig.value.columns.find((c: any) => c.type === 'tag')
})

const schemaStatusOptions = computed(() => {
  return schemaStatusField.value?.options || []
})

const schemaHasCreate = computed(() =>
  schemaConfig.value?.actions?.some((a: any) => a.key === 'create')
)
const schemaHasEdit = computed(() =>
  schemaConfig.value?.actions?.some((a: any) => a.key === 'edit')
)
const schemaHasDelete = computed(() =>
  schemaConfig.value?.actions?.some((a: any) => a.key === 'delete')
)
const schemaHasRowActions = computed(() => schemaHasEdit.value || schemaHasDelete.value)
const schemaCreateLabel = computed(() => {
  const action = schemaConfig.value?.actions?.find((a: any) => a.key === 'create')
  return action?.label || '新建'
})

const schemaFilteredItems = computed(() => {
  let items = schemaData.value?.items || []
  if (schemaSearch.value) {
    const keyword = schemaSearch.value.toLowerCase()
    items = items.filter((item: any) =>
      Object.values(item).some(v => String(v || '').toLowerCase().includes(keyword))
    )
  }
  if (schemaStatusFilter.value && schemaStatusField.value) {
    items = items.filter(
      (item: any) => item[schemaStatusField.value.key] === schemaStatusFilter.value
    )
  }
  return items
})

async function initPlugin() {
  if (!props.pluginId) return

  status.value = 'loading'
  errorMessage.value = ''

  try {
    if (props.renderMode === 'custom') {
      await initCustomMode()
    } else {
      await initSchemaMode()
    }
    status.value = 'loaded'
    emit('loaded', props.pluginId)
  } catch (err: any) {
    status.value = 'error'
    errorMessage.value = err.message || '未知错误'
    emit('error', props.pluginId, errorMessage.value)
  }
}

async function initCustomMode() {
  const tenantId = userStore.tenantId || 0

  const fullManifest: PluginManifest = {
    id: props.pluginId,
    name: props.manifest.name || props.pluginId,
    version: props.manifest.version || '1.0.0',
    entry: props.manifest.entry || 'index.js',
    style: props.manifest.style,
    renderMode: 'custom',
    permissions: props.manifest.permissions,
    allowedApiPrefixes: props.manifest.allowedApiPrefixes,
    compatibleBaseVersion: props.manifest.compatibleBaseVersion,
    signature: props.manifest.signature,
  }

  const loaded: LoadedPlugin = await loadPlugin(fullManifest, tenantId)

  if (loaded.status === 'error') {
    throw new Error(loaded.error || '插件加载失败')
  }

  if (!loaded.component) {
    throw new Error('插件未导出有效组件')
  }

  await nextTick()

  if (props.useShadowDom && hostRef.value) {
    initShadowDom(loaded)
  }

  setupBridgeEvents(loaded)

  dynamicComponent.value = loaded.component
}

async function initSchemaMode() {
  const res = await request.get(`/plugin-engine/plugins/${props.pluginId}/page`)
  const data = res.data || {}

  schemaData.value = data
  schemaConfig.value = {
    type: data.type,
    columns: data.columns,
    actions: data.actions,
    form: data.form,
    stats: data.stats,
    api: data.api,
  }
  buildSchemaFormConfig()
}

function initShadowDom(loaded: LoadedPlugin) {
  if (!hostRef.value) return

  const existingShadow = hostRef.value.shadowRoot
  if (existingShadow) {
    shadowRoot.value = existingShadow
    return
  }

  shadowRoot.value = createShadowContainer(hostRef.value, props.pluginId)

  injectElementPlusStylesToShadow(shadowRoot.value)

  if (loaded.manifest.style) {
    fetch(`/plugin-assets/${props.pluginId}/${loaded.manifest.style}`)
      .then(r => r.text())
      .then(css => {
        if (shadowRoot.value) {
          injectStyleToShadow(shadowRoot.value, css, props.pluginId)
        }
      })
      .catch(() => {})
  }
}

function setupBridgeEvents(loaded: LoadedPlugin) {
  const bridge = loaded.bridge

  bridge.on('toast', (data: any) => {
    const opts = data as { message: string; type?: string; duration?: number }
    ElMessage({
      message: opts.message,
      type: (opts.type as any) || 'info',
      duration: opts.duration || 3000,
    })
  })

  bridge.on('confirm', (data: any) => {
    const opts = data as {
      title?: string
      message: string
      confirmText?: string
      cancelText?: string
      type?: string
      _resolve: (v: boolean) => void
    }
    ElMessageBox.confirm(opts.message, opts.title || '确认', {
      confirmButtonText: opts.confirmText || '确定',
      cancelButtonText: opts.cancelText || '取消',
      type: (opts.type as any) || 'info',
    })
      .then(() => opts._resolve(true))
      .catch(() => opts._resolve(false))
  })

  bridge.on('navigate', (data: any) => {
    const opts = data as { path: string; query?: Record<string, string> }
    if (opts.path === '__BACK__') {
      router.back()
    } else {
      router.push({ path: opts.path, query: opts.query })
    }
  })
}

function handlePluginEvent(event: string, payload: any) {
  emit(event as any, payload)
}

function retryLoad() {
  const existing = getLoadedPlugin(props.pluginId)
  if (existing) {
    unloadPlugin(props.pluginId)
  }
  initPlugin()
}

function buildSchemaFormConfig() {
  if (!schemaConfig.value?.form) return
  for (const field of schemaConfig.value.form) {
    if (field.type === 'switch') {
      schemaFormData[field.key] = false
    } else if (field.type === 'radio') {
      schemaFormData[field.key] = field.default || (field.options?.[0]?.value ?? '')
    } else {
      schemaFormData[field.key] = ''
    }
    if (field.required) {
      schemaFormRules[field.key] = [
        { required: true, message: `请输入${field.label}`, trigger: field.type === 'switch' ? 'change' : 'blur' },
      ]
    }
  }
}

function resetSchemaForm() {
  if (!schemaConfig.value?.form) return
  for (const field of schemaConfig.value.form) {
    if (field.type === 'switch') {
      schemaFormData[field.key] = false
    } else if (field.type === 'radio') {
      schemaFormData[field.key] = field.default || (field.options?.[0]?.value ?? '')
    } else {
      schemaFormData[field.key] = ''
    }
  }
  schemaEditingItem.value = null
}

function openSchemaDialog(item?: any) {
  resetSchemaForm()
  if (item && schemaConfig.value?.form) {
    schemaEditingItem.value = item
    for (const field of schemaConfig.value.form) {
      schemaFormData[field.key] = item[field.key] ?? (field.type === 'switch' ? false : '')
    }
  }
  schemaDialogVisible.value = true
}

async function handleSchemaSubmit() {
  if (!schemaFormRef.value) return
  await schemaFormRef.value.validate()
  schemaSaveLoading.value = true
  try {
    const api = schemaConfig.value?.api
    if (schemaEditingItem.value) {
      if (api) await request.put(`${api}/${schemaEditingItem.value.id}`, schemaFormData)
      ElMessage.success('更新成功')
    } else {
      if (api) await request.post(api, schemaFormData)
      ElMessage.success('创建成功')
    }
    schemaDialogVisible.value = false
    await refreshSchemaData()
  } catch (err: any) {
    ElMessage.error(err.message || '操作失败')
  } finally {
    schemaSaveLoading.value = false
  }
}

async function handleSchemaDelete(row: any) {
  try {
    await ElMessageBox.confirm('确定要删除吗？删除后不可恢复。', '确认删除', {
      type: 'error',
      confirmButtonText: '确定删除',
      confirmButtonClass: 'el-button--danger',
    })
    const api = schemaConfig.value?.api
    if (api) await request.delete(`${api}/${row.id}`)
    ElMessage.success('删除成功')
    await refreshSchemaData()
  } catch {}
}

async function refreshSchemaData() {
  try {
    const res = await request.get(`/plugin-engine/plugins/${props.pluginId}/page`)
    schemaData.value = res.data || {}
  } catch {}
}

function statValue(key: string): number {
  const items = schemaData.value?.items || []
  if (key === 'total') return items.length
  const stat = schemaConfig.value?.stats?.find((s: any) => s.key === key)
  if (stat?.filter) {
    const [field, value] = stat.filter.split('=')
    return items.filter((i: any) => String(i[field]) === value).length
  }
  return 0
}

function statBgClass(color: string) {
  const map: Record<string, string> = {
    blue: 'bg-blue-100', green: 'bg-green-100',
    orange: 'bg-orange-100', red: 'bg-red-100',
  }
  return map[color] || 'bg-slate-100'
}

function statTextClass(color: string) {
  const map: Record<string, string> = {
    blue: 'text-blue-600', green: 'text-green-600',
    orange: 'text-orange-600', red: 'text-red-600',
  }
  return map[color] || 'text-slate-600'
}

function schemaTagType(value: any, col: any) {
  const opt = (col.options || []).find((o: any) => o.value === value)
  return opt?.type || 'info'
}

function schemaTagLabel(value: any, col: any) {
  const opt = (col.options || []).find((o: any) => o.value === value)
  return opt?.label || value
}

onMounted(() => {
  initPlugin()
})

onBeforeUnmount(() => {
  if (props.renderMode === 'custom') {
    unloadPlugin(props.pluginId)
  }
  removePluginStyle(props.pluginId)
  emit('unloaded', props.pluginId)
})

watch(() => props.pluginId, (newId, oldId) => {
  if (newId && newId !== oldId) {
    if (oldId && props.renderMode === 'custom') {
      unloadPlugin(oldId)
    }
    initPlugin()
  }
})
</script>

<style scoped>
.host-loader {
  min-height: 100%;
}
.glass-card {
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(226, 232, 240, 0.8);
}
</style>
