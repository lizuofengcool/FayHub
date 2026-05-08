<template>
  <div class="backup-page">
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 pb-3 flex items-center justify-between">
        <div>
          <h2 class="text-lg font-bold text-slate-800">数据维护</h2>
          <p class="text-slate-400 text-xs mt-0.5">数据库备份、恢复、SQL执行及数据管理</p>
        </div>
      </div>

      <el-tabs v-model="activeTab" class="data-maintenance-tabs">
      <el-tab-pane label="数据备份" name="backup">
        <div class="space-y-4">
          <div class="flex items-center gap-3 flex-wrap">
            <el-input v-model="backupSearch" placeholder="搜索表名或注释" clearable style="width: 280px" @input="filterTables">
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
            <el-select v-model="backupSort" placeholder="排序方式" style="width: 160px" @change="sortTables">
              <el-option label="表名升序" value="name_asc" />
              <el-option label="表名降序" value="name_desc" />
              <el-option label="大小升序" value="size_asc" />
              <el-option label="大小降序" value="size_desc" />
              <el-option label="记录升序" value="rows_asc" />
              <el-option label="记录降序" value="rows_desc" />
              <el-option label="时间升序" value="time_asc" />
              <el-option label="时间降序" value="time_desc" />
            </el-select>
            <el-button type="primary" @click="backupSelectedTables" :loading="backingUp" :disabled="selectedTables.length === 0">
              <el-icon class="mr-1"><FolderAdd /></el-icon> 备份选中表
            </el-button>
            <el-button @click="toggleSelectAll">
              {{ isAllSelected ? '取消全选' : '全选' }}
            </el-button>
            <span class="text-sm text-slate-500">已选 {{ selectedTables.length }} / {{ filteredTableList.length }} 张表</span>
          </div>

          <div class="bg-white rounded-xl border border-slate-100 overflow-hidden">
            <el-table :data="filteredTableList" v-loading="tablesLoading" stripe max-height="560" @selection-change="handleTableSelectionChange" ref="backupTableRef">
              <el-table-column type="selection" width="45" />
              <el-table-column prop="name" label="表名" min-width="220">
                <template #default="{ row }">
                  <span class="font-mono text-sm text-blue-600 cursor-pointer" @click="showFieldDict(row)">{{ row.name }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="comment" label="注释" min-width="140">
                <template #default="{ row }">
                  <span class="text-sm">{{ row.comment || '-' }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="total_size" label="大小" width="120" align="center">
                <template #default="{ row }">
                  <span class="text-sm">{{ row.total_size || '-' }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="row_count" label="记录数" width="100" align="center">
                <template #default="{ row }">
                  <span class="text-sm cursor-pointer text-blue-600" @click="previewTable(row)">{{ row.row_count }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="update_time" label="更新时间" width="170" align="center">
                <template #default="{ row }">
                  <span class="text-sm text-slate-500">{{ row.update_time || 'N/A' }}</span>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="100" align="center" fixed="right">
                <template #default="{ row }">
                  <el-button text type="primary" size="small" @click="backupSingleTable(row)">备份</el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="数据恢复" name="recover">
        <div class="space-y-4">
          <div class="flex items-center gap-3">
            <el-button type="primary" @click="fetchBackups" :loading="backupsLoading">
              <el-icon class="mr-1"><Refresh /></el-icon> 刷新列表
            </el-button>
            <el-button @click="showUploadDialog = true">
              <el-icon class="mr-1"><Upload /></el-icon> 上传恢复
            </el-button>
          </div>

          <div class="bg-white rounded-xl border border-slate-100 overflow-hidden">
            <el-table :data="backups" v-loading="backupsLoading" stripe empty-text="暂无备份记录">
              <el-table-column type="selection" width="45" />
              <el-table-column prop="filename" label="备份系列" min-width="280">
                <template #default="{ row }">
                  <div class="flex items-center gap-2">
                    <el-icon class="text-amber-500"><Folder /></el-icon>
                    <span class="font-mono text-sm">{{ row.filename }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="notes" label="备注" width="220">
                <template #default="{ row }">
                  <el-input v-model="row.notes" size="small" placeholder="添加备注" @blur="updateNotes(row)" />
                </template>
              </el-table-column>
              <el-table-column prop="file_size" label="文件大小" width="110" align="center">
                <template #default="{ row }">
                  <span class="text-sm">{{ formatFileSize(row.file_size) }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="created_at" label="备份时间" width="170" align="center">
                <template #default="{ row }">
                  <span class="text-sm text-slate-500">{{ formatTime(row.created_at) }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="volumes" label="分卷" width="70" align="center">
                <template #default="{ row }">
                  <span class="text-sm">{{ row.volumes || '-' }}</span>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="160" align="center" fixed="right">
                <template #default="{ row }">
                  <div class="flex items-center justify-center gap-1">
                    <el-popconfirm title="确定恢复此备份？现有数据将被覆盖，此操作不可恢复" @confirm="restoreBackupByID(row)">
                      <template #reference>
                        <el-button text type="warning" size="small">导入</el-button>
                      </template>
                    </el-popconfirm>
                    <el-button text type="primary" size="small" @click="downloadBackup(row)">下载</el-button>
                    <el-popconfirm title="确定删除此备份？删除后不可恢复" @confirm="deleteBackup(row)">
                      <template #reference>
                        <el-button text type="danger" size="small">删除</el-button>
                      </template>
                    </el-popconfirm>
                  </div>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="执行语句" name="execute">
        <div class="space-y-4">
          <el-alert title="注意：执行SQL语句将直接操作数据库，请谨慎操作" type="warning" :closable="false" show-icon />
          <div class="bg-white rounded-xl border border-slate-100 p-4">
            <el-input v-model="sqlInput" type="textarea" :rows="8" placeholder="请输入SQL语句，例如：SELECT * FROM users LIMIT 10" font="monospace" />
            <div class="flex items-center gap-4 mt-4">
              <el-button type="danger" @click="executeSQL" :loading="executingSQL">
                <el-icon class="mr-1"><VideoPlay /></el-icon> 执行语句
              </el-button>
              <el-checkbox v-model="sqlShowErrors">显示报错</el-checkbox>
              <el-button @click="sqlInput = ''">清空</el-button>
            </div>
          </div>

          <div v-if="sqlResult" class="bg-white rounded-xl border border-slate-100 overflow-hidden">
            <div class="px-4 py-3 border-b border-slate-100 flex items-center justify-between">
              <span class="text-sm font-medium text-slate-700">执行结果</span>
              <span class="text-xs text-slate-400">{{ sqlResultTime }}</span>
            </div>
            <div v-if="sqlResultColumns.length > 0" class="overflow-x-auto">
              <el-table :data="sqlResultRows" stripe size="small" max-height="400">
                <el-table-column v-for="col in sqlResultColumns" :key="col" :prop="col" :label="col" min-width="120">
                  <template #default="{ row }">
                    <span class="text-xs font-mono">{{ row[col] }}</span>
                  </template>
                </el-table-column>
              </el-table>
              <div class="px-4 py-2 text-sm text-slate-500">
                共 {{ sqlResultRows.length }} 条记录
              </div>
            </div>
            <div v-else class="p-4">
              <el-alert :title="sqlResultMessage" :type="sqlResultSuccess ? 'success' : 'error'" :closable="false" show-icon />
            </div>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="显示进程" name="process">
        <div class="space-y-4">
          <div class="flex items-center gap-3">
            <el-button type="primary" @click="fetchProcesses" :loading="processesLoading">
              <el-icon class="mr-1"><Refresh /></el-icon> 刷新进程
            </el-button>
            <el-button type="danger" @click="killSelectedProcesses" :disabled="selectedProcesses.length === 0">
              <el-icon class="mr-1"><SwitchButton /></el-icon> 结束选中进程
            </el-button>
            <span class="text-sm text-slate-500">当前 {{ processList.length }} 个进程</span>
          </div>

          <div class="bg-white rounded-xl border border-slate-100 overflow-hidden">
            <el-table :data="processList" v-loading="processesLoading" stripe empty-text="暂无进程">
              <el-table-column type="selection" width="45" @selection-change="handleProcessSelectionChange" />
              <el-table-column prop="pid" label="PID" width="80" align="center" />
              <el-table-column prop="user" label="用户" width="140" />
              <el-table-column prop="host" label="主机" width="140" />
              <el-table-column prop="database" label="数据库" width="140" />
              <el-table-column prop="command" label="命令" width="100" />
              <el-table-column prop="time" label="时间" width="80" align="center" />
              <el-table-column prop="state" label="状态" width="120" />
              <el-table-column prop="query" label="SQL查询" min-width="260">
                <template #default="{ row }">
                  <el-tooltip :content="row.query" placement="top">
                    <span class="text-xs font-mono text-slate-600 truncate block max-w-xs">{{ row.query }}</span>
                  </el-tooltip>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="80" align="center" fixed="right">
                <template #default="{ row }">
                  <el-popconfirm title="确定结束此进程？此操作不可撤销" @confirm="killProcess(row)">
                    <template #reference>
                      <el-button text type="danger" size="small">结束</el-button>
                    </template>
                  </el-popconfirm>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="字段校验" name="verify">
        <div class="space-y-4">
          <div class="flex items-center gap-3">
            <el-button type="primary" @click="runVerify" :loading="verifying">
              <el-icon class="mr-1"><CircleCheck /></el-icon> 开始校验
            </el-button>
            <el-input v-model="verifySearch" placeholder="搜索表名或字段" clearable style="width: 260px" @input="filterVerifyResults">
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
            <el-select v-model="verifyStatusFilter" placeholder="状态筛选" style="width: 140px" clearable @change="filterVerifyResults">
              <el-option label="全部" value="" />
              <el-option label="通过" value="pass" />
              <el-option label="异常" value="error" />
              <el-option label="未知" value="unknown" />
            </el-select>
          </div>

          <div class="bg-white rounded-xl border border-slate-100 overflow-hidden">
            <el-table :data="filteredVerifyList" v-loading="verifying" stripe empty-text="点击开始校验">
              <el-table-column prop="table_name" label="表名" min-width="220">
                <template #default="{ row }">
                  <span class="font-mono text-sm text-blue-600">{{ row.table_name }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="field_count" label="字段数" width="100" align="center" />
              <el-table-column prop="row_count" label="记录数" width="100" align="center" />
              <el-table-column prop="status" label="校验结果" width="120" align="center">
                <template #default="{ row }">
                  <el-tag v-if="row.status === 'pass'" type="success" size="small">通过</el-tag>
                  <el-tag v-else-if="row.status === 'error'" type="danger" size="small">异常</el-tag>
                  <el-tag v-else type="info" size="small">未知</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="详情" width="80" align="center">
                <template #default="{ row }">
                  <el-button v-if="row.issues && row.issues.length > 0" text type="primary" size="small" @click="showVerifyDetail(row)">查看</el-button>
                  <span v-else class="text-slate-400 text-sm">-</span>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="字符替换" name="replace">
        <div class="space-y-6">
          <div class="bg-white rounded-xl border border-slate-100 p-5">
            <h3 class="text-base font-semibold text-slate-700 mb-4">备份内容替换</h3>
            <el-form label-width="100px" label-position="right">
              <el-form-item label="备份系列">
                <el-select v-model="replaceFileForm.backupSeries" placeholder="选择备份文件系列" style="width: 100%">
                  <el-option v-for="b in backups" :key="b.id" :label="b.filename" :value="b.filename" />
                </el-select>
              </el-form-item>
              <el-form-item label="查找">
                <el-input v-model="replaceFileForm.find" placeholder="输入要查找的内容" />
              </el-form-item>
              <el-form-item label="替换为">
                <el-input v-model="replaceFileForm.replace" placeholder="输入替换后的内容" />
              </el-form-item>
              <el-form-item>
                <el-button type="danger" @click="executeFileReplace" :loading="replacingFile">执行</el-button>
              </el-form-item>
            </el-form>
          </div>

          <div class="bg-white rounded-xl border border-slate-100 p-5">
            <h3 class="text-base font-semibold text-slate-700 mb-4">数据内容替换</h3>
            <el-form label-width="100px" label-position="right">
              <el-form-item label="替换目标">
                <el-select v-model="replaceDataForm.table" placeholder="全部数据表" clearable style="width: 100%" @change="onReplaceTableChange">
                  <el-option label="全部数据表" value="" />
                  <el-option v-for="t in tableList" :key="t.name" :label="`${t.name} (${t.comment})`" :value="t.name" />
                </el-select>
              </el-form-item>
              <el-form-item v-if="replaceDataForm.table" label="指定字段">
                <el-select v-model="replaceDataForm.field" placeholder="全部字段" clearable style="width: 100%">
                  <el-option label="全部字段" value="" />
                  <el-option v-for="f in replaceTableFields" :key="f.name" :label="`${f.name} (${f.comment || f.type})`" :value="f.name" />
                </el-select>
              </el-form-item>
              <el-form-item label="替换类型">
                <el-radio-group v-model="replaceDataForm.replaceType">
                  <el-radio :value="1">直接替换</el-radio>
                  <el-radio :value="2">头部追加</el-radio>
                  <el-radio :value="3">尾部追加</el-radio>
                </el-radio-group>
              </el-form-item>
              <el-form-item v-if="replaceDataForm.replaceType === 1" label="查找">
                <el-input v-model="replaceDataForm.find" type="textarea" :rows="2" placeholder="输入要查找的内容" />
              </el-form-item>
              <el-form-item v-if="replaceDataForm.replaceType === 1" label="替换为">
                <el-input v-model="replaceDataForm.replace" type="textarea" :rows="2" placeholder="输入替换后的内容" />
              </el-form-item>
              <el-form-item v-if="replaceDataForm.replaceType !== 1" label="追加内容">
                <el-input v-model="replaceDataForm.addContent" type="textarea" :rows="2" placeholder="输入追加内容" />
              </el-form-item>
              <el-form-item label="替换条件">
                <el-input v-model="replaceDataForm.condition" placeholder="AND开头的MySQL条件语句，例如 AND status=3" />
              </el-form-item>
              <el-form-item label="每轮查询">
                <el-input-number v-model="replaceDataForm.batchSize" :min="100" :max="10000" :step="500" />
                <span class="text-sm text-slate-400 ml-2">条</span>
              </el-form-item>
              <el-form-item>
                <el-button type="danger" @click="executeDataReplace" :loading="replacingData">执行</el-button>
              </el-form-item>
            </el-form>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="数据互转" name="transfer">
        <div class="bg-white rounded-xl border border-slate-100 p-5">
          <el-form label-width="100px" label-position="right">
            <el-form-item label="来源表">
              <el-select v-model="transferForm.sourceTable" placeholder="选择来源表" filterable style="width: 100%">
                <el-option v-for="t in tableList" :key="t.name" :label="`${t.name} (${t.comment})`" :value="t.name" />
              </el-select>
            </el-form-item>
            <el-form-item label="目标表">
              <el-select v-model="transferForm.targetTable" placeholder="选择目标表" filterable style="width: 100%">
                <el-option v-for="t in tableList" :key="t.name" :label="`${t.name} (${t.comment})`" :value="t.name" />
              </el-select>
            </el-form-item>
            <el-form-item label="转移条件">
              <el-input v-model="transferForm.condition" placeholder="AND开头的MySQL条件语句，例如 AND status=3" />
              <div class="text-xs text-slate-400 mt-1 leading-relaxed">
                可直接写SQL调用条件，必须以and开头。例如 and catid=123 表示调用分类ID为123的信息
              </div>
            </el-form-item>
            <el-form-item label="删除源数据">
              <el-radio-group v-model="transferForm.deleteSource">
                <el-radio :value="true">是</el-radio>
                <el-radio :value="false">否</el-radio>
              </el-radio-group>
              <div class="text-xs text-slate-400 mt-1">如果选是，源数据会移入回收站，不会直接删除</div>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="executeTransfer" :loading="transferring">执行</el-button>
            </el-form-item>
          </el-form>
        </div>
      </el-tab-pane>

      <el-tab-pane label="数据导入" name="import">
        <div class="bg-white rounded-xl border border-slate-100 p-5">
          <el-alert class="mb-4" title="导入说明" type="info" :closable="false" show-icon>
            <template #default>
              <div class="text-xs leading-relaxed">
                第一行为字段中文名，仅为方便录入，对导入数据无影响，可留空；第二行为数据表对应字段名，必须和数据表内字段一致；第三行及以后的行需要录入待导入的数据。支持 .sql、.csv、.xls、.xlsx 格式。
              </div>
            </template>
          </el-alert>
          <el-form label-width="100px" label-position="right">
            <el-form-item label="导入目标">
              <el-select v-model="importForm.table" placeholder="请选择目标表" filterable style="width: 100%">
                <el-option v-for="t in tableList" :key="t.name" :label="`${t.name} (${t.comment})`" :value="t.name" />
              </el-select>
            </el-form-item>
            <el-form-item label="数据文件">
              <el-upload ref="importUploadRef" drag :auto-upload="false" :limit="1" accept=".sql,.csv,.xls,.xlsx" :on-change="handleImportFileChange" :on-remove="handleImportFileRemove">
                <el-icon class="text-4xl text-slate-400 mb-3"><UploadFilled /></el-icon>
                <div class="text-sm text-slate-600">将文件拖到此处，或点击上传</div>
                <template #tip>
                  <div class="text-xs text-slate-400 mt-2">支持 .sql、.csv、.xls、.xlsx 格式</div>
                </template>
              </el-upload>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="executeImport" :loading="importing" :disabled="!importForm.table || !importFile">开始导入</el-button>
            </el-form-item>
          </el-form>
        </div>
      </el-tab-pane>

      <el-tab-pane label="数据导出" name="export">
        <div class="bg-white rounded-xl border border-slate-100 p-5">
          <el-form label-width="100px" label-position="right">
            <el-form-item label="数据来源">
              <el-select v-model="exportForm.table" placeholder="选择表" filterable style="width: 100%" @change="onExportTableChange">
                <el-option v-for="t in tableList" :key="t.name" :label="`${t.name} (${t.comment})`" :value="t.name" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="exportTableFields.length > 0" label="导出字段">
              <el-select v-model="exportForm.fields" multiple collapse-tags collapse-tags-tooltip placeholder="全部字段" style="width: 100%">
                <el-option v-for="f in exportTableFields" :key="f.name" :label="`${f.name} (${f.comment || f.type})`" :value="f.name" />
              </el-select>
            </el-form-item>
            <el-form-item label="导出条件">
              <el-input v-model="exportForm.condition" placeholder="AND开头的MySQL条件语句，例如 AND status=3" />
            </el-form-item>
            <el-form-item label="时间字段">
              <el-select v-model="exportForm.timeField" placeholder="选择时间字段" clearable style="width: 100%">
                <el-option v-for="f in exportTimeFields" :key="f.name" :label="f.name" :value="f.name" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="exportForm.timeField" label="时间范围">
              <el-date-picker v-model="exportForm.dateRange" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" style="width: 100%" />
            </el-form-item>
            <el-form-item label="排序方式">
              <el-input v-model="exportForm.order" placeholder="例如 id DESC" />
            </el-form-item>
            <el-form-item label="导出格式">
              <el-select v-model="exportForm.format" style="width: 200px">
                <el-option label="SQL" value="sql" />
                <el-option label="CSV" value="csv" />
                <el-option label="XML" value="xml" />
                <el-option label="JSON" value="json" />
              </el-select>
            </el-form-item>
            <el-form-item label="每轮查询">
              <el-input-number v-model="exportForm.pageSize" :min="100" :max="50000" :step="1000" />
              <span class="text-sm text-slate-400 ml-2">条</span>
            </el-form-item>
            <el-form-item label="页码">
              <el-input-number v-model="exportForm.page" :min="1" />
              <span class="text-sm text-slate-400 ml-2">共 {{ exportTotalPages }} 页 / {{ exportTotalCount }} 条</span>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="executeExport" :loading="exporting">导出</el-button>
            </el-form-item>
          </el-form>
        </div>
      </el-tab-pane>
    </el-tabs>
    </div>

    <el-dialog v-model="showUploadDialog" title="上传备份恢复" width="480px" :close-on-click-modal="false">
      <div class="space-y-4">
        <el-alert title="警告：恢复操作将覆盖当前数据库所有数据，请谨慎操作！" type="warning" :closable="false" show-icon />
        <el-upload ref="uploadRef" drag :auto-upload="false" :limit="1" accept=".sql" :on-change="handleFileChange" :on-remove="handleFileRemove">
          <el-icon class="text-4xl text-slate-400 mb-3"><UploadFilled /></el-icon>
          <div class="text-sm text-slate-600">将 .sql 备份文件拖到此处，或点击上传</div>
          <template #tip>
            <div class="text-xs text-slate-400 mt-2">仅支持 .sql 格式的备份文件</div>
          </template>
        </el-upload>
      </div>
      <template #footer>
        <el-button @click="showUploadDialog = false">取消</el-button>
        <el-button type="danger" @click="restoreBackup" :loading="restoring" :disabled="!uploadFile">确认恢复</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showPreviewDialog" :title="`预览表 - ${previewTableName}`" width="80%" top="5vh">
      <div class="overflow-x-auto">
        <el-table :data="previewRows" stripe size="small" max-height="500" v-loading="previewLoading">
          <el-table-column v-for="col in previewColumns" :key="col" :prop="col" :label="col" min-width="120">
            <template #default="{ row }">
              <span class="text-xs font-mono">{{ row[col] }}</span>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-dialog>

    <el-dialog v-model="showFieldDictDialog" :title="`数据字典 - ${fieldDictTableName}`" width="70%" top="5vh">
      <el-table :data="fieldDictList" stripe size="small" v-loading="fieldDictLoading">
        <el-table-column prop="name" label="字段名" min-width="150">
          <template #default="{ row }">
            <span class="font-mono text-sm">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" min-width="140">
          <template #default="{ row }">
            <span class="text-sm">{{ row.type }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="nullable" label="允许空" width="90" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.nullable === 'YES'" type="warning" size="small">YES</el-tag>
            <el-tag v-else type="success" size="small">NO</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="default" label="默认值" min-width="120">
          <template #default="{ row }">
            <span class="text-sm font-mono">{{ row.default || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="comment" label="注释" min-width="160">
          <template #default="{ row }">
            <span class="text-sm">{{ row.comment || '-' }}</span>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <el-dialog v-model="showVerifyDetailDialog" :title="`校验详情 - ${verifyDetailTable}`" width="600px">
      <div class="space-y-2">
        <div v-for="(issue, idx) in verifyDetailIssues" :key="idx" class="text-sm text-red-600 flex items-start gap-2">
          <el-icon class="mt-0.5"><WarningFilled /></el-icon>
          <span>{{ issue }}</span>
        </div>
        <div v-if="verifyDetailIssues.length === 0" class="text-sm text-green-600">无异常</div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import type { UploadFile } from 'element-plus'
import {
  Search, Refresh, FolderAdd, Upload, UploadFilled, Folder,
  VideoPlay, SwitchButton, CircleCheck, WarningFilled
} from '@element-plus/icons-vue'
import backupApi from '@/api/backup'

function getErrMsg(err: unknown, fallback: string): string {
  if (err && typeof err === 'object' && 'response' in err) {
    const resp = (err as any).response
    if (resp?.data?.message) return resp.data.message
  }
  if (err instanceof Error) return err.message || fallback
  return fallback
}

const activeTab = ref('backup')

const tableList = ref<any[]>([])
const tablesLoading = ref(false)
const backupSearch = ref('')
const backupSort = ref('')
const selectedTables = ref<any[]>([])
const backingUp = ref(false)
const backupTableRef = ref()

const filteredTableList = computed(() => {
  let list = [...tableList.value]
  if (backupSearch.value) {
    const kw = backupSearch.value.toLowerCase()
    list = list.filter(t => t.name?.toLowerCase().includes(kw) || t.comment?.toLowerCase().includes(kw))
  }
  if (backupSort.value) {
    const sortMap: Record<string, (a: any, b: any) => number> = {
      name_asc: (a, b) => (a.name || '').localeCompare(b.name || ''),
      name_desc: (a, b) => (b.name || '').localeCompare(a.name || ''),
      size_asc: (a, b) => (a.row_count || 0) - (b.row_count || 0),
      size_desc: (a, b) => (b.row_count || 0) - (a.row_count || 0),
      rows_asc: (a, b) => (a.row_count || 0) - (b.row_count || 0),
      rows_desc: (a, b) => (b.row_count || 0) - (a.row_count || 0),
      time_asc: (a, b) => (a.update_time || '').localeCompare(b.update_time || ''),
      time_desc: (a, b) => (b.update_time || '').localeCompare(a.update_time || ''),
    }
    if (sortMap[backupSort.value]) list.sort(sortMap[backupSort.value])
  }
  return list
})

const isAllSelected = computed(() => filteredTableList.value.length > 0 && selectedTables.value.length === filteredTableList.value.length)

const backups = ref<any[]>([])
const backupsLoading = ref(false)
const restoring = ref(false)
const showUploadDialog = ref(false)
const uploadFile = ref<File | null>(null)
const uploadRef = ref()

const sqlInput = ref('')
const sqlShowErrors = ref(false)
const executingSQL = ref(false)
const sqlResult = ref(false)
const sqlResultColumns = ref<string[]>([])
const sqlResultRows = ref<any[]>([])
const sqlResultMessage = ref('')
const sqlResultSuccess = ref(false)
const sqlResultTime = ref('')

const processList = ref<any[]>([])
const processesLoading = ref(false)
const selectedProcesses = ref<any[]>([])

const verifyResults = ref<any[]>([])
const verifying = ref(false)
const verifySearch = ref('')
const verifyStatusFilter = ref('')
const showVerifyDetailDialog = ref(false)
const verifyDetailTable = ref('')
const verifyDetailIssues = ref<string[]>([])

const filteredVerifyList = computed(() => {
  let list = [...verifyResults.value]
  if (verifySearch.value) {
    const kw = verifySearch.value.toLowerCase()
    list = list.filter(t => t.table_name?.toLowerCase().includes(kw))
  }
  if (verifyStatusFilter.value) {
    list = list.filter(t => t.status === verifyStatusFilter.value)
  }
  return list
})

const replaceFileForm = ref({ backupSeries: '', find: '', replace: '' })
const replacingFile = ref(false)
const replaceDataForm = ref({
  table: '', field: '', replaceType: 1, find: '', replace: '',
  addContent: '', condition: '', batchSize: 1000
})
const replacingData = ref(false)
const replaceTableFields = ref<any[]>([])

const transferForm = ref({ sourceTable: '', targetTable: '', condition: '', deleteSource: false })
const transferring = ref(false)

const importForm = ref({ table: '' })
const importFile = ref<File | null>(null)
const importUploadRef = ref()
const importing = ref(false)

const exportForm = ref({
  table: '', fields: [] as string[], condition: '', timeField: '',
  dateRange: null as string[] | null, order: '', format: 'csv',
  pageSize: 5000, page: 1
})
const exporting = ref(false)
const exportTableFields = ref<any[]>([])
const exportTotalPages = ref(0)
const exportTotalCount = ref(0)
const exportTimeFields = computed(() => exportTableFields.value.filter(f => f.type?.includes('time') || f.type?.includes('date')))

const showPreviewDialog = ref(false)
const previewTableName = ref('')
const previewRows = ref<any[]>([])
const previewColumns = ref<string[]>([])
const previewLoading = ref(false)

const showFieldDictDialog = ref(false)
const fieldDictTableName = ref('')
const fieldDictList = ref<any[]>([])
const fieldDictLoading = ref(false)

onMounted(() => {
  fetchTables()
  fetchBackups()
})

async function fetchTables() {
  tablesLoading.value = true
  try {
    const res = await backupApi.listTables()
    if (res?.data) {
      tableList.value = res.data.list || res.data || []
    }
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '获取表列表失败'))
  } finally {
    tablesLoading.value = false
  }
}

async function fetchBackups() {
  backupsLoading.value = true
  try {
    const res = await backupApi.listBackups()
    if (res?.data) {
      backups.value = res.data.list || []
    }
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '获取备份列表失败'))
  } finally {
    backupsLoading.value = false
  }
}

function handleTableSelectionChange(selection: any[]) {
  selectedTables.value = selection
}

function toggleSelectAll() {
  if (isAllSelected.value) {
    backupTableRef.value?.clearSelection()
  } else {
    filteredTableList.value.forEach(row => {
      backupTableRef.value?.toggleRowSelection(row, true)
    })
  }
}

function filterTables() {}
function sortTables() {}

async function backupSelectedTables() {
  if (selectedTables.value.length === 0) {
    ElMessage.warning('请选择要备份的表')
    return
  }
  backingUp.value = true
  try {
    const tables = selectedTables.value.map(t => t.name)
    await backupApi.createBackupForTables(tables)
    ElMessage.success('备份创建成功')
    fetchBackups()
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '创建备份失败'))
  } finally {
    backingUp.value = false
  }
}

async function backupSingleTable(row: any) {
  backingUp.value = true
  try {
    await backupApi.createBackupForTables([row.name])
    ElMessage.success(`表 ${row.name} 备份成功`)
    fetchBackups()
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '创建备份失败'))
  } finally {
    backingUp.value = false
  }
}

async function showFieldDict(row: any) {
  fieldDictTableName.value = row.name
  fieldDictLoading.value = true
  showFieldDictDialog.value = true
  fieldDictList.value = []
  try {
    const res = await backupApi.getTableFields(row.name)
    if (res?.data) {
      fieldDictList.value = Array.isArray(res.data) ? res.data : []
    }
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '获取字段信息失败'))
  } finally {
    fieldDictLoading.value = false
  }
}

async function previewTable(row: any) {
  previewTableName.value = row.name
  previewLoading.value = true
  showPreviewDialog.value = true
  previewColumns.value = []
  previewRows.value = []
  try {
    const res = await backupApi.previewTable(row.name, 50)
    if (res?.data) {
      previewColumns.value = res.data.columns || []
      previewRows.value = res.data.rows || []
    }
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '预览表失败'))
  } finally {
    previewLoading.value = false
  }
}

function downloadBackup(row: any) {
  const token = localStorage.getItem('fayhub_token') || ''
  const link = document.createElement('a')
  link.href = `/api/backups/${row.id}/download?token=${token}`
  link.download = row.filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

async function deleteBackup(row: any) {
  try {
    await backupApi.deleteBackup(row.id)
    ElMessage.success('备份删除成功')
    fetchBackups()
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '删除备份失败'))
  }
}

async function restoreBackupByID(row: any) {
  restoring.value = true
  try {
    await backupApi.restoreBackupByID(row.id)
    ElMessage.success('数据库恢复成功')
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '恢复数据库失败'))
  } finally {
    restoring.value = false
  }
}

function handleFileChange(file: UploadFile) {
  uploadFile.value = file.raw || null
}

function handleFileRemove() {
  uploadFile.value = null
}

async function restoreBackup() {
  if (!uploadFile.value) {
    ElMessage.warning('请先选择备份文件')
    return
  }
  restoring.value = true
  try {
    await backupApi.restoreBackup(uploadFile.value)
    ElMessage.success('数据库恢复成功')
    showUploadDialog.value = false
    uploadFile.value = null
    uploadRef.value?.clearFiles()
    fetchBackups()
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '恢复数据库失败'))
  } finally {
    restoring.value = false
  }
}

async function updateNotes(row: any) {
  try {
    await backupApi.updateBackupNotes(row.id, row.notes || '')
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '更新备注失败'))
  }
}

async function executeSQL() {
  if (!sqlInput.value.trim()) {
    ElMessage.warning('请输入SQL语句')
    return
  }
  const isWrite = !sqlInput.value.trim().toLowerCase().startsWith('select')
  if (isWrite) {
    try {
      await backupApi.executeWriteSQL(sqlInput.value.trim())
      sqlResult.value = true
      sqlResultColumns.value = []
      sqlResultRows.value = []
      sqlResultMessage.value = '语句执行成功'
      sqlResultSuccess.value = true
      sqlResultTime.value = new Date().toLocaleString()
      ElMessage.success('语句执行成功')
    } catch (err: unknown) {
      sqlResult.value = true
      sqlResultMessage.value = getErrMsg(err, '语句执行失败')
      sqlResultSuccess.value = false
      sqlResultTime.value = new Date().toLocaleString()
      if (sqlShowErrors.value) {
        ElMessage.error(getErrMsg(err, '语句执行失败'))
      }
    }
    return
  }
  executingSQL.value = true
  try {
    const res = await backupApi.executeSQL(sqlInput.value.trim(), sqlShowErrors.value)
    sqlResult.value = true
    sqlResultTime.value = new Date().toLocaleString()
    if (res?.data) {
      const data = res.data
      if (data.columns && data.rows) {
        sqlResultColumns.value = data.columns
        sqlResultRows.value = data.rows
        sqlResultMessage.value = ''
      } else if (Array.isArray(data) && data.length > 0) {
        sqlResultColumns.value = Object.keys(data[0])
        sqlResultRows.value = data
        sqlResultMessage.value = ''
      } else {
        sqlResultColumns.value = []
        sqlResultRows.value = []
        sqlResultMessage.value = '查询结果为空'
        sqlResultSuccess.value = true
      }
    }
  } catch (err: unknown) {
    sqlResult.value = true
    sqlResultMessage.value = getErrMsg(err, '语句执行失败')
    sqlResultSuccess.value = false
    sqlResultTime.value = new Date().toLocaleString()
    if (sqlShowErrors.value) {
      ElMessage.error(getErrMsg(err, '语句执行失败'))
    }
  } finally {
    executingSQL.value = false
  }
}

async function fetchProcesses() {
  processesLoading.value = true
  try {
    const res = await backupApi.listProcesses()
    if (res?.data) {
      processList.value = Array.isArray(res.data) ? res.data : res.data.list || []
    }
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '获取进程列表失败'))
  } finally {
    processesLoading.value = false
  }
}

function handleProcessSelectionChange(selection: any[]) {
  selectedProcesses.value = selection
}

async function killSelectedProcesses() {
  if (selectedProcesses.value.length === 0) return
  for (const proc of selectedProcesses.value) {
    try {
      await backupApi.killProcess(proc.pid)
    } catch (err: unknown) {
      ElMessage.error(`结束进程 ${proc.pid} 失败: ${getErrMsg(err, '')}`)
    }
  }
  ElMessage.success('选中进程已结束')
  fetchProcesses()
}

async function killProcess(row: any) {
  try {
    await backupApi.killProcess(row.pid)
    ElMessage.success(`进程 ${row.pid} 已结束`)
    fetchProcesses()
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '结束进程失败'))
  }
}

async function runVerify() {
  verifying.value = true
  try {
    const res = await backupApi.verifyFields()
    if (res?.data) {
      verifyResults.value = Array.isArray(res.data) ? res.data : res.data.list || []
    }
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '字段校验失败'))
  } finally {
    verifying.value = false
  }
}

function filterVerifyResults() {}

function showVerifyDetail(row: any) {
  verifyDetailTable.value = row.table_name
  verifyDetailIssues.value = row.issues || []
  showVerifyDetailDialog.value = true
}

async function onReplaceTableChange(tableName: string) {
  if (!tableName) {
    replaceTableFields.value = []
    return
  }
  try {
    const res = await backupApi.getTableFields(tableName)
    if (res?.data) {
      replaceTableFields.value = Array.isArray(res.data) ? res.data : res.data.list || []
    }
  } catch (err: unknown) {
    replaceTableFields.value = []
  }
}

async function executeFileReplace() {
  if (!replaceFileForm.value.backupSeries) {
    ElMessage.warning('请选择备份系列')
    return
  }
  if (!replaceFileForm.value.find) {
    ElMessage.warning('请输入查找内容')
    return
  }
  replacingFile.value = true
  try {
    await backupApi.advancedReplace({
      table: replaceFileForm.value.backupSeries,
      find: replaceFileForm.value.find,
      replace: replaceFileForm.value.replace,
      replace_type: 1
    })
    ElMessage.success('备份内容替换成功')
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '替换失败'))
  } finally {
    replacingFile.value = false
  }
}

async function executeDataReplace() {
  if (replaceDataForm.value.replaceType === 1 && !replaceDataForm.value.find) {
    ElMessage.warning('请输入查找内容')
    return
  }
  if (replaceDataForm.value.replaceType !== 1 && !replaceDataForm.value.addContent) {
    ElMessage.warning('请输入追加内容')
    return
  }
  replacingData.value = true
  try {
    await backupApi.advancedReplace({
      table: replaceDataForm.value.table || '',
      field: replaceDataForm.value.field || undefined,
      find: replaceDataForm.value.replaceType === 1 ? replaceDataForm.value.find : replaceDataForm.value.addContent,
      replace: replaceDataForm.value.replaceType === 1 ? replaceDataForm.value.replace : replaceDataForm.value.addContent,
      replace_type: replaceDataForm.value.replaceType,
      condition: replaceDataForm.value.condition || undefined,
      batch_size: replaceDataForm.value.batchSize
    })
    ElMessage.success('数据内容替换成功')
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '替换失败'))
  } finally {
    replacingData.value = false
  }
}

async function executeTransfer() {
  if (!transferForm.value.sourceTable) {
    ElMessage.warning('请选择来源表')
    return
  }
  if (!transferForm.value.targetTable) {
    ElMessage.warning('请选择目标表')
    return
  }
  if (transferForm.value.sourceTable === transferForm.value.targetTable) {
    ElMessage.warning('来源表和目标表不能相同')
    return
  }
  transferring.value = true
  try {
    await backupApi.dataTransfer({
      source_table: transferForm.value.sourceTable,
      target_table: transferForm.value.targetTable,
      condition: transferForm.value.condition || undefined,
      delete_source: transferForm.value.deleteSource
    })
    ElMessage.success('数据互转成功')
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '数据互转失败'))
  } finally {
    transferring.value = false
  }
}

function handleImportFileChange(file: UploadFile) {
  importFile.value = file.raw || null
}

function handleImportFileRemove() {
  importFile.value = null
}

async function executeImport() {
  if (!importForm.value.table) {
    ElMessage.warning('请选择目标表')
    return
  }
  if (!importFile.value) {
    ElMessage.warning('请选择数据文件')
    return
  }
  importing.value = true
  try {
    await backupApi.importData(importForm.value.table, importFile.value)
    ElMessage.success('数据导入成功')
    importFile.value = null
    importUploadRef.value?.clearFiles()
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '数据导入失败'))
  } finally {
    importing.value = false
  }
}

async function onExportTableChange(tableName: string) {
  if (!tableName) {
    exportTableFields.value = []
    exportTotalPages.value = 0
    exportTotalCount.value = 0
    return
  }
  try {
    const [fieldsRes, countRes] = await Promise.all([
      backupApi.getTableFields(tableName),
      backupApi.getTableCount(tableName, exportForm.value.condition || undefined)
    ])
    if (fieldsRes?.data) {
      exportTableFields.value = Array.isArray(fieldsRes.data) ? fieldsRes.data : fieldsRes.data.list || []
    }
    if (countRes?.data) {
      exportTotalCount.value = countRes.data.total || 0
      exportTotalPages.value = Math.ceil(exportTotalCount.value / exportForm.value.pageSize) || 0
    }
  } catch (err: unknown) {
    exportTableFields.value = []
  }
}

async function executeExport() {
  if (!exportForm.value.table) {
    ElMessage.warning('请选择数据表')
    return
  }
  exporting.value = true
  try {
    const params: any = {
      table: exportForm.value.table,
      format: exportForm.value.format,
      page_size: exportForm.value.pageSize,
      page: exportForm.value.page
    }
    if (exportForm.value.fields.length > 0) params.fields = exportForm.value.fields.join(',')
    if (exportForm.value.condition) params.condition = exportForm.value.condition
    if (exportForm.value.timeField && exportForm.value.dateRange) {
      params.time_field = exportForm.value.timeField
      params.from_date = exportForm.value.dateRange[0]
      params.to_date = exportForm.value.dateRange[1]
    }
    if (exportForm.value.order) params.order = exportForm.value.order
    const res = await backupApi.advancedExport(params)
    if (res) {
      const blob = res instanceof Blob ? res : new Blob([res as any], { type: 'application/octet-stream' })
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `${exportForm.value.table}.${exportForm.value.format}`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
      ElMessage.success('数据导出成功')
    }
  } catch (err: unknown) {
    ElMessage.error(getErrMsg(err, '数据导出失败'))
  } finally {
    exporting.value = false
  }
}

function formatFileSize(bytes: number): string {
  if (!bytes || bytes === 0) return '-'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let size = bytes
  while (size >= 1024 && i < units.length - 1) {
    size /= 1024
    i++
  }
  return `${size.toFixed(1)} ${units[i]}`
}

function formatTime(dateStr: string): string {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}
</script>

<style scoped>
.data-maintenance-tabs :deep(.el-tabs__header) {
  padding: 0 20px;
  margin-bottom: 0;
}
.data-maintenance-tabs :deep(.el-tabs__content) {
  padding: 16px 20px 20px;
}
.data-maintenance-tabs :deep(.el-tab-pane) {
  min-height: 300px;
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

:deep(.el-button) {
  height: 32px;
  padding: 8px 12px;
}
</style>
