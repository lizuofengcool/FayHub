?<template>
  <div class="backup-page">
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 pb-3 flex items-center justify-between">
        <div>
          <h2 class="text-lg font-bold text-slate-800">����ά��</h2>
          <p class="text-slate-400 text-xs mt-0.5">���ݿⱸ�ݡ��ָ���SQLִ�м����ݹ���</p>
        </div>
      </div>

      <el-tabs v-model="activeTab" class="data-maintenance-tabs">
      <el-tab-pane label="���ݱ���" name="backup">
        <div class="space-y-4">
          <div class="flex items-center gap-3 flex-wrap">
            <el-input v-model="backupSearch" placeholder="����������ע��" clearable style="width: 280px" @input="filterTables">
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
            <el-select v-model="backupSort" placeholder="����ʽ" style="width: 160px" @change="sortTables">
              <el-option label="��������" value="name_asc" />
              <el-option label="��������" value="name_desc" />
              <el-option label="��С����" value="size_asc" />
              <el-option label="��С����" value="size_desc" />
              <el-option label="��¼����" value="rows_asc" />
              <el-option label="��¼����" value="rows_desc" />
              <el-option label="ʱ������" value="time_asc" />
              <el-option label="ʱ�併��" value="time_desc" />
            </el-select>
            <el-button type="default" @click="backupSelectedTables" :loading="backingUp" :disabled="selectedTables.length === 0">
              <el-icon class="mr-1"><FolderAdd /></el-icon> ����ѡ�б�
            </el-button>
            <el-button @click="toggleSelectAll">
              {{ isAllSelected ? 'ȡ��ȫѡ' : 'ȫѡ' }}
            </el-button>
            <span class="text-sm text-slate-500">��ѡ {{ selectedTables.length }} / {{ filteredTableList.length }} �ű�</span>
          </div>

          <div class="bg-white rounded-xl border border-slate-100 overflow-hidden">
            <el-table :data="filteredTableList" v-loading="tablesLoading" stripe max-height="560" @selection-change="handleTableSelectionChange" ref="backupTableRef">
              <el-table-column type="selection" width="45" />
              <el-table-column prop="name" label="����" min-width="220">
                <template #default="{ row }">
                  <span class="font-mono text-sm text-blue-600 cursor-pointer" @click="showFieldDict(row)">{{ row.name }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="comment" label="ע��" min-width="140">
                <template #default="{ row }">
                  <span class="text-sm">{{ row.comment || '-' }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="total_size" label="��С" width="120" align="center">
                <template #default="{ row }">
                  <span class="text-sm">{{ row.total_size || '-' }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="row_count" label="��¼��" width="100" align="center">
                <template #default="{ row }">
                  <span class="text-sm cursor-pointer text-blue-600" @click="previewTable(row)">{{ row.row_count }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="update_time" label="����ʱ��" width="170" align="center">
                <template #default="{ row }">
                  <span class="text-sm text-slate-500">{{ row.update_time || 'N/A' }}</span>
                </template>
              </el-table-column>
              <el-table-column label="����" width="100" align="center" fixed="right">
                <template #default="{ row }">
                  <el-button text type="default" size="small" @click="backupSingleTable(row)">����</el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="���ݻָ�" name="recover">
        <div class="space-y-4">
          <div class="flex items-center gap-3">
            <el-button type="default" @click="fetchBackups" :loading="backupsLoading">
              <el-icon class="mr-1"><Refresh /></el-icon> ˢ���б�
            </el-button>
            <el-button @click="showUploadDialog = true">
              <el-icon class="mr-1"><Upload /></el-icon> �ϴ��ָ�
            </el-button>
          </div>

          <div class="bg-white rounded-xl border border-slate-100 overflow-hidden">
            <el-table :data="backups" v-loading="backupsLoading" stripe empty-text="���ޱ��ݼ�¼">
              <el-table-column type="selection" width="45" />
              <el-table-column prop="filename" label="����ϵ��" min-width="280">
                <template #default="{ row }">
                  <div class="flex items-center gap-2">
                    <el-icon class="text-amber-500"><Folder /></el-icon>
                    <span class="font-mono text-sm">{{ row.filename }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="notes" label="��ע" width="220">
                <template #default="{ row }">
                  <el-input v-model="row.notes" size="small" placeholder="���ӱ�ע" @blur="updateNotes(row)" />
                </template>
              </el-table-column>
              <el-table-column prop="file_size" label="�ļ���С" width="110" align="center">
                <template #default="{ row }">
                  <span class="text-sm">{{ formatFileSize(row.file_size) }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="created_at" label="����ʱ��" width="170" align="center">
                <template #default="{ row }">
                  <span class="text-sm text-slate-500">{{ formatTime(row.created_at) }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="volumes" label="�־�" width="70" align="center">
                <template #default="{ row }">
                  <span class="text-sm">{{ row.volumes || '-' }}</span>
                </template>
              </el-table-column>
              <el-table-column label="����" width="160" align="center" fixed="right">
                <template #default="{ row }">
                  <div class="flex items-center justify-center gap-1">
                    <n-popconfirm title="ȷ���ָ��˱��ݣ��������ݽ������ǣ��˲������ɻָ�" @confirm="restoreBackupByID(row)">
                      <template #trigger>
                        <el-button text type="warning" size="small">����</el-button>
                      </template>
                    </n-popconfirm>
                    <el-button text type="default" size="small" @click="downloadBackup(row)">����</el-button>
                    <n-popconfirm title="ȷ��ɾ���˱��ݣ�ɾ���󲻿ɻָ�" @confirm="deleteBackup(row)">
                      <template #trigger>
                        <el-button text type="error" size="small">ɾ��</el-button>
                      </template>
                    </n-popconfirm>
                  </div>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="ִ�����" name="execute">
        <div class="space-y-4">
          <n-alert title="ע�⣺ִ��SQL��佫ֱ�Ӳ������ݿ⣬���������" type="warning" :closable="false" show-icon />
          <div class="bg-white rounded-xl border border-slate-100 p-4">
            <el-input v-model="sqlInput" type="textarea" :rows="8" placeholder="������SQL��䣬���磺SELECT * FROM users LIMIT 10" font="monospace" />
            <div class="flex items-center gap-4 mt-4">
              <el-button type="error" @click="executeSQL" :loading="executingSQL">
                <el-icon class="mr-1"><VideoPlay /></el-icon> ִ�����
              </el-button>
              <el-checkbox v-model="sqlShowErrors">��ʾ����</el-checkbox>
              <el-button @click="sqlInput = ''">���</el-button>
            </div>
          </div>

          <div v-if="sqlResult" class="bg-white rounded-xl border border-slate-100 overflow-hidden">
            <div class="px-4 py-3 border-b border-slate-100 flex items-center justify-between">
              <span class="text-sm font-medium text-slate-700">ִ�н��</span>
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
                �� {{ sqlResultRows.length }} ����¼
              </div>
            </div>
            <div v-else class="p-4">
              <n-alert :title="sqlResultMessage" :type="sqlResultSuccess ? 'success' : 'error'" :closable="false" show-icon />
            </div>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="��ʾ����" name="process">
        <div class="space-y-4">
          <div class="flex items-center gap-3">
            <el-button type="default" @click="fetchProcesses" :loading="processesLoading">
              <el-icon class="mr-1"><Refresh /></el-icon> ˢ�½���
            </el-button>
            <el-button type="error" @click="killSelectedProcesses" :disabled="selectedProcesses.length === 0">
              <el-icon class="mr-1"><SwitchButton /></el-icon> ����ѡ�н���
            </el-button>
            <span class="text-sm text-slate-500">��ǰ {{ processList.length }} ������</span>
          </div>

          <div class="bg-white rounded-xl border border-slate-100 overflow-hidden">
            <el-table :data="processList" v-loading="processesLoading" stripe empty-text="���޽���">
              <el-table-column type="selection" width="45" @selection-change="handleProcessSelectionChange" />
              <el-table-column prop="pid" label="PID" width="80" align="center" />
              <el-table-column prop="user" label="�û�" width="140" />
              <el-table-column prop="host" label="����" width="140" />
              <el-table-column prop="database" label="���ݿ�" width="140" />
              <el-table-column prop="command" label="����" width="100" />
              <el-table-column prop="time" label="ʱ��" width="80" align="center" />
              <el-table-column prop="state" label="״̬" width="120" />
              <el-table-column prop="query" label="SQL��ѯ" min-width="260">
                <template #default="{ row }">
                  <n-tooltip trigger="hover">
                    <template #trigger>
                      <span class="text-xs font-mono text-slate-600 truncate block max-w-xs">{{ row.query }}</span>
                    </template>
                    {{ row.query }}
                  </n-tooltip>
                </template>
              </el-table-column>
              <el-table-column label="����" width="80" align="center" fixed="right">
                <template #default="{ row }">
                  <n-popconfirm title="ȷ�������˽��̣��˲������ɳ���" @confirm="killProcess(row)">
                    <template #trigger>
                      <el-button text type="error" size="small">����</el-button>
                    </template>
                  </n-popconfirm>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="�ֶ�У��" name="verify">
        <div class="space-y-4">
          <div class="flex items-center gap-3">
            <el-button type="default" @click="runVerify" :loading="verifying">
              <el-icon class="mr-1"><CircleCheck /></el-icon> ��ʼУ��
            </el-button>
            <el-input v-model="verifySearch" placeholder="�����������ֶ�" clearable style="width: 260px" @input="filterVerifyResults">
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
            <el-select v-model="verifyStatusFilter" placeholder="״̬ɸѡ" style="width: 140px" clearable @change="filterVerifyResults">
              <el-option label="ȫ��" value="" />
              <el-option label="ͨ��" value="pass" />
              <el-option label="�쳣" value="error" />
              <el-option label="δ֪" value="unknown" />
            </el-select>
          </div>

          <div class="bg-white rounded-xl border border-slate-100 overflow-hidden">
            <el-table :data="filteredVerifyList" v-loading="verifying" stripe empty-text="�����ʼУ��">
              <el-table-column prop="table_name" label="����" min-width="220">
                <template #default="{ row }">
                  <span class="font-mono text-sm text-blue-600">{{ row.table_name }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="field_count" label="�ֶ���" width="100" align="center" />
              <el-table-column prop="row_count" label="��¼��" width="100" align="center" />
              <el-table-column prop="status" label="У����" width="120" align="center">
                <template #default="{ row }">
                  <n-tag v-if="row.status === 'pass'" type="success" size="small">ͨ��</n-tag>
                  <n-tag v-else-if="row.status === 'error'" type="error" size="small">�쳣</n-tag>
                  <n-tag v-else type="default" size="small">δ֪</n-tag>
                </template>
              </el-table-column>
              <el-table-column label="����" width="80" align="center">
                <template #default="{ row }">
                  <el-button v-if="row.issues && row.issues.length > 0" text type="default" size="small" @click="showVerifyDetail(row)">�鿴</el-button>
                  <span v-else class="text-slate-400 text-sm">-</span>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="�ַ��滻" name="replace">
        <div class="space-y-6">
          <div class="bg-white rounded-xl border border-slate-100 p-5">
            <h3 class="text-base font-semibold text-slate-700 mb-4">���������滻</h3>
            <el-form label-width="100px" label-position="right">
              <el-form-item label="����ϵ��">
                <el-select v-model="replaceFileForm.backupSeries" placeholder="ѡ�񱸷��ļ�ϵ��" style="width: 100%">
                  <el-option v-for="b in backups" :key="b.id" :label="b.filename" :value="b.filename" />
                </el-select>
              </el-form-item>
              <el-form-item label="����">
                <el-input v-model="replaceFileForm.find" placeholder="����Ҫ���ҵ�����" />
              </el-form-item>
              <el-form-item label="�滻Ϊ">
                <el-input v-model="replaceFileForm.replace" placeholder="�����滻�������" />
              </el-form-item>
              <el-form-item>
                <el-button type="error" @click="executeFileReplace" :loading="replacingFile">ִ��</el-button>
              </el-form-item>
            </el-form>
          </div>

          <div class="bg-white rounded-xl border border-slate-100 p-5">
            <h3 class="text-base font-semibold text-slate-700 mb-4">���������滻</h3>
            <el-form label-width="100px" label-position="right">
              <el-form-item label="�滻Ŀ��">
                <el-select v-model="replaceDataForm.table" placeholder="ȫ�����ݱ�" clearable style="width: 100%" @change="onReplaceTableChange">
                  <el-option label="ȫ�����ݱ�" value="" />
                  <el-option v-for="t in tableList" :key="t.name" :label="`${t.name} (${t.comment})`" :value="t.name" />
                </el-select>
              </el-form-item>
              <el-form-item v-if="replaceDataForm.table" label="ָ���ֶ�">
                <el-select v-model="replaceDataForm.field" placeholder="ȫ���ֶ�" clearable style="width: 100%">
                  <el-option label="ȫ���ֶ�" value="" />
                  <el-option v-for="f in replaceTableFields" :key="f.name" :label="`${f.name} (${f.comment || f.type})`" :value="f.name" />
                </el-select>
              </el-form-item>
              <el-form-item label="�滻����">
                <el-radio-group v-model="replaceDataForm.replaceType">
                  <el-radio :value="1">ֱ���滻</el-radio>
                  <el-radio :value="2">ͷ��׷��</el-radio>
                  <el-radio :value="3">β��׷��</el-radio>
                </el-radio-group>
              </el-form-item>
              <el-form-item v-if="replaceDataForm.replaceType === 1" label="����">
                <el-input v-model="replaceDataForm.find" type="textarea" :rows="2" placeholder="����Ҫ���ҵ�����" />
              </el-form-item>
              <el-form-item v-if="replaceDataForm.replaceType === 1" label="�滻Ϊ">
                <el-input v-model="replaceDataForm.replace" type="textarea" :rows="2" placeholder="�����滻�������" />
              </el-form-item>
              <el-form-item v-if="replaceDataForm.replaceType !== 1" label="׷������">
                <el-input v-model="replaceDataForm.addContent" type="textarea" :rows="2" placeholder="����׷������" />
              </el-form-item>
              <el-form-item label="�滻����">
                <el-input v-model="replaceDataForm.condition" placeholder="AND��ͷ��MySQL������䣬���� AND status=3" />
              </el-form-item>
              <el-form-item label="ÿ�ֲ�ѯ">
                <el-input-number v-model="replaceDataForm.batchSize" :min="100" :max="10000" :step="500" />
                <span class="text-sm text-slate-400 ml-2">��</span>
              </el-form-item>
              <el-form-item>
                <el-button type="error" @click="executeDataReplace" :loading="replacingData">ִ��</el-button>
              </el-form-item>
            </el-form>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="���ݻ�ת" name="transfer">
        <div class="bg-white rounded-xl border border-slate-100 p-5">
          <el-form label-width="100px" label-position="right">
            <el-form-item label="��Դ��">
              <el-select v-model="transferForm.sourceTable" placeholder="ѡ����Դ��" filterable style="width: 100%">
                <el-option v-for="t in tableList" :key="t.name" :label="`${t.name} (${t.comment})`" :value="t.name" />
              </el-select>
            </el-form-item>
            <el-form-item label="Ŀ���">
              <el-select v-model="transferForm.targetTable" placeholder="ѡ��Ŀ���" filterable style="width: 100%">
                <el-option v-for="t in tableList" :key="t.name" :label="`${t.name} (${t.comment})`" :value="t.name" />
              </el-select>
            </el-form-item>
            <el-form-item label="ת������">
              <el-input v-model="transferForm.condition" placeholder="AND��ͷ��MySQL������䣬���� AND status=3" />
              <div class="text-xs text-slate-400 mt-1 leading-relaxed">
                ��ֱ��дSQL����������������and��ͷ������ and catid=123 ��ʾ���÷���IDΪ123����Ϣ
              </div>
            </el-form-item>
            <el-form-item label="ɾ��Դ����">
              <el-radio-group v-model="transferForm.deleteSource">
                <el-radio :value="true">��</el-radio>
                <el-radio :value="false">��</el-radio>
              </el-radio-group>
              <div class="text-xs text-slate-400 mt-1">���ѡ�ǣ�Դ���ݻ��������վ������ֱ��ɾ��</div>
            </el-form-item>
            <el-form-item>
              <el-button type="default" @click="executeTransfer" :loading="transferring">ִ��</el-button>
            </el-form-item>
          </el-form>
        </div>
      </el-tab-pane>

      <el-tab-pane label="���ݵ���" name="import">
        <div class="bg-white rounded-xl border border-slate-100 p-5">
          <n-alert class="mb-4" title="����˵��" type="default" :closable="false" show-icon>
            <template #default>
              <div class="text-xs leading-relaxed">
                ��һ��Ϊ�ֶ�����������Ϊ����¼�룬�Ե���������Ӱ�죬�����գ��ڶ���Ϊ���ݱ���Ӧ�ֶ�������������ݱ����ֶ�һ�£������м��Ժ������Ҫ¼�����������ݡ�֧�� .sql��.csv��.xls��.xlsx ��ʽ��
              </div>
            </template>
          </n-alert>
          <el-form label-width="100px" label-position="right">
            <el-form-item label="����Ŀ��">
              <el-select v-model="importForm.table" placeholder="��ѡ��Ŀ���" filterable style="width: 100%">
                <el-option v-for="t in tableList" :key="t.name" :label="`${t.name} (${t.comment})`" :value="t.name" />
              </el-select>
            </el-form-item>
            <el-form-item label="�����ļ�">
              <el-upload ref="importUploadRef" drag :auto-upload="false" :limit="1" accept=".sql,.csv,.xls,.xlsx" :on-change="handleImportFileChange" :on-remove="handleImportFileRemove">
                <el-icon class="text-4xl text-slate-400 mb-3"><UploadFilled /></el-icon>
                <div class="text-sm text-slate-600">���ļ��ϵ��˴��������ϴ�</div>
                <template #tip>
                  <div class="text-xs text-slate-400 mt-2">֧�� .sql��.csv��.xls��.xlsx ��ʽ</div>
                </template>
              </el-upload>
            </el-form-item>
            <el-form-item>
              <el-button type="default" @click="executeImport" :loading="importing" :disabled="!importForm.table || !importFile">��ʼ����</el-button>
            </el-form-item>
          </el-form>
        </div>
      </el-tab-pane>

      <el-tab-pane label="���ݵ���" name="export">
        <div class="bg-white rounded-xl border border-slate-100 p-5">
          <el-form label-width="100px" label-position="right">
            <el-form-item label="������Դ">
              <el-select v-model="exportForm.table" placeholder="ѡ���" filterable style="width: 100%" @change="onExportTableChange">
                <el-option v-for="t in tableList" :key="t.name" :label="`${t.name} (${t.comment})`" :value="t.name" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="exportTableFields.length > 0" label="�����ֶ�">
              <el-select v-model="exportForm.fields" multiple collapse-tags collapse-tags-tooltip placeholder="ȫ���ֶ�" style="width: 100%">
                <el-option v-for="f in exportTableFields" :key="f.name" :label="`${f.name} (${f.comment || f.type})`" :value="f.name" />
              </el-select>
            </el-form-item>
            <el-form-item label="��������">
              <el-input v-model="exportForm.condition" placeholder="AND��ͷ��MySQL������䣬���� AND status=3" />
            </el-form-item>
            <el-form-item label="ʱ���ֶ�">
              <el-select v-model="exportForm.timeField" placeholder="ѡ��ʱ���ֶ�" clearable style="width: 100%">
                <el-option v-for="f in exportTimeFields" :key="f.name" :label="f.name" :value="f.name" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="exportForm.timeField" label="ʱ�䷶Χ">
              <el-date-picker v-model="exportForm.dateRange" type="daterange" range-separator="��" start-placeholder="��ʼ����" end-placeholder="��������" value-format="YYYY-MM-DD" style="width: 100%" />
            </el-form-item>
            <el-form-item label="����ʽ">
              <el-input v-model="exportForm.order" placeholder="���� id DESC" />
            </el-form-item>
            <el-form-item label="������ʽ">
              <el-select v-model="exportForm.format" style="width: 200px">
                <el-option label="SQL" value="sql" />
                <el-option label="CSV" value="csv" />
                <el-option label="XML" value="xml" />
                <el-option label="JSON" value="json" />
              </el-select>
            </el-form-item>
            <el-form-item label="ÿ�ֲ�ѯ">
              <el-input-number v-model="exportForm.pageSize" :min="100" :max="50000" :step="1000" />
              <span class="text-sm text-slate-400 ml-2">��</span>
            </el-form-item>
            <el-form-item label="ҳ��">
              <el-input-number v-model="exportForm.page" :min="1" />
              <span class="text-sm text-slate-400 ml-2">�� {{ exportTotalPages }} ҳ / {{ exportTotalCount }} ��</span>
            </el-form-item>
            <el-form-item>
              <el-button type="default" @click="executeExport" :loading="exporting">����</el-button>
            </el-form-item>
          </el-form>
        </div>
      </el-tab-pane>
    </el-tabs>
    </div>

    <el-dialog v-model="showUploadDialog" title="�ϴ����ݻָ�" width="480px" :close-on-click-modal="false">
      <div class="space-y-4">
        <n-alert title="���棺�ָ����������ǵ�ǰ���ݿ��������ݣ������������" type="warning" :closable="false" show-icon />
        <el-upload ref="uploadRef" drag :auto-upload="false" :limit="1" accept=".sql" :on-change="handleFileChange" :on-remove="handleFileRemove">
          <el-icon class="text-4xl text-slate-400 mb-3"><UploadFilled /></el-icon>
          <div class="text-sm text-slate-600">�� .sql �����ļ��ϵ��˴��������ϴ�</div>
          <template #tip>
            <div class="text-xs text-slate-400 mt-2">��֧�� .sql ��ʽ�ı����ļ�</div>
          </template>
        </el-upload>
      </div>
      <template #footer>
        <el-button @click="showUploadDialog = false">ȡ��</el-button>
        <el-button type="error" @click="restoreBackup" :loading="restoring" :disabled="!uploadFile">ȷ�ϻָ�</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showPreviewDialog" :title="`Ԥ���� - ${previewTableName}`" width="80%" top="5vh">
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

    <el-dialog v-model="showFieldDictDialog" :title="`�����ֵ� - ${fieldDictTableName}`" width="70%" top="5vh">
      <el-table :data="fieldDictList" stripe size="small" v-loading="fieldDictLoading">
        <el-table-column prop="name" label="�ֶ���" min-width="150">
          <template #default="{ row }">
            <span class="font-mono text-sm">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="����" min-width="140">
          <template #default="{ row }">
            <span class="text-sm">{{ row.type }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="nullable" label="������" width="90" align="center">
          <template #default="{ row }">
            <n-tag v-if="row.nullable === 'YES'" type="warning" size="small">YES</n-tag>
            <n-tag v-else type="success" size="small">NO</n-tag>
          </template>
        </el-table-column>
        <el-table-column prop="default" label="Ĭ��ֵ" min-width="120">
          <template #default="{ row }">
            <span class="text-sm font-mono">{{ row.default || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="comment" label="ע��" min-width="160">
          <template #default="{ row }">
            <span class="text-sm">{{ row.comment || '-' }}</span>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <el-dialog v-model="showVerifyDetailDialog" :title="`У������ - ${verifyDetailTable}`" width="600px">
      <div class="space-y-2">
        <div v-for="(issue, idx) in verifyDetailIssues" :key="idx" class="text-sm text-red-600 flex items-start gap-2">
          <el-icon class="mt-0.5"><WarningFilled /></el-icon>
          <span>{{ issue }}</span>
        </div>
        <div v-if="verifyDetailIssues.length === 0" class="text-sm text-green-600">���쳣</div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
const message = useMessage()
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
      tableList.value = res.data || []
    }
  } catch (err: unknown) {
    message.error(getErrMsg(err, '��ȡ���б�ʧ��'))
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
    message.error(getErrMsg(err, '��ȡ�����б�ʧ��'))
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
    message.warning('��ѡ��Ҫ���ݵı�')
    return
  }
  backingUp.value = true
  try {
    const tables = selectedTables.value.map(t => t.name)
    await backupApi.createBackupForTables(tables)
    message.success('���ݴ����ɹ�')
    fetchBackups()
  } catch (err: unknown) {
    message.error(getErrMsg(err, '��������ʧ��'))
  } finally {
    backingUp.value = false
  }
}

async function backupSingleTable(row: any) {
  backingUp.value = true
  try {
    await backupApi.createBackupForTables([row.name])
    message.success(`�� ${row.name} ���ݳɹ�`)
    fetchBackups()
  } catch (err: unknown) {
    message.error(getErrMsg(err, '��������ʧ��'))
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
    message.error(getErrMsg(err, '��ȡ�ֶ���Ϣʧ��'))
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
    message.error(getErrMsg(err, 'Ԥ����ʧ��'))
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
    message.success('����ɾ���ɹ�')
    fetchBackups()
  } catch (err: unknown) {
    message.error(getErrMsg(err, 'ɾ������ʧ��'))
  }
}

async function restoreBackupByID(row: any) {
  restoring.value = true
  try {
    await backupApi.restoreBackupByID(row.id)
    message.success('���ݿ�ָ��ɹ�')
  } catch (err: unknown) {
    message.error(getErrMsg(err, '�ָ����ݿ�ʧ��'))
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
    message.warning('����ѡ�񱸷��ļ�')
    return
  }
  restoring.value = true
  try {
    await backupApi.restoreBackup(uploadFile.value)
    message.success('���ݿ�ָ��ɹ�')
    showUploadDialog.value = false
    uploadFile.value = null
    uploadRef.value?.clearFiles()
    fetchBackups()
  } catch (err: unknown) {
    message.error(getErrMsg(err, '�ָ����ݿ�ʧ��'))
  } finally {
    restoring.value = false
  }
}

async function updateNotes(row: any) {
  try {
    await backupApi.updateBackupNotes(row.id, row.notes || '')
  } catch (err: unknown) {
    message.error(getErrMsg(err, '���±�עʧ��'))
  }
}

async function executeSQL() {
  if (!sqlInput.value.trim()) {
    message.warning('������SQL���')
    return
  }
  const isWrite = !sqlInput.value.trim().toLowerCase().startsWith('select')
  if (isWrite) {
    try {
      await backupApi.executeWriteSQL(sqlInput.value.trim())
      sqlResult.value = true
      sqlResultColumns.value = []
      sqlResultRows.value = []
      sqlResultMessage.value = '���ִ�гɹ�'
      sqlResultSuccess.value = true
      sqlResultTime.value = new Date().toLocaleString()
      message.success('���ִ�гɹ�')
    } catch (err: unknown) {
      sqlResult.value = true
      sqlResultMessage.value = getErrMsg(err, '���ִ��ʧ��')
      sqlResultSuccess.value = false
      sqlResultTime.value = new Date().toLocaleString()
      if (sqlShowErrors.value) {
        message.error(getErrMsg(err, '���ִ��ʧ��'))
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
        sqlResultMessage.value = '��ѯ���Ϊ��'
        sqlResultSuccess.value = true
      }
    }
  } catch (err: unknown) {
    sqlResult.value = true
    sqlResultMessage.value = getErrMsg(err, '���ִ��ʧ��')
    sqlResultSuccess.value = false
    sqlResultTime.value = new Date().toLocaleString()
    if (sqlShowErrors.value) {
      message.error(getErrMsg(err, '���ִ��ʧ��'))
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
    message.error(getErrMsg(err, '��ȡ�����б�ʧ��'))
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
      message.error(`�������� ${proc.pid} ʧ��: ${getErrMsg(err, '')}`)
    }
  }
  message.success('ѡ�н����ѽ���')
  fetchProcesses()
}

async function killProcess(row: any) {
  try {
    await backupApi.killProcess(row.pid)
    message.success(`���� ${row.pid} �ѽ���`)
    fetchProcesses()
  } catch (err: unknown) {
    message.error(getErrMsg(err, '��������ʧ��'))
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
    message.error(getErrMsg(err, '�ֶ�У��ʧ��'))
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
    message.warning('��ѡ�񱸷�ϵ��')
    return
  }
  if (!replaceFileForm.value.find) {
    message.warning('�������������')
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
    message.success('���������滻�ɹ�')
  } catch (err: unknown) {
    message.error(getErrMsg(err, '�滻ʧ��'))
  } finally {
    replacingFile.value = false
  }
}

async function executeDataReplace() {
  if (replaceDataForm.value.replaceType === 1 && !replaceDataForm.value.find) {
    message.warning('�������������')
    return
  }
  if (replaceDataForm.value.replaceType !== 1 && !replaceDataForm.value.addContent) {
    message.warning('������׷������')
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
    message.success('���������滻�ɹ�')
  } catch (err: unknown) {
    message.error(getErrMsg(err, '�滻ʧ��'))
  } finally {
    replacingData.value = false
  }
}

async function executeTransfer() {
  if (!transferForm.value.sourceTable) {
    message.warning('��ѡ����Դ��')
    return
  }
  if (!transferForm.value.targetTable) {
    message.warning('��ѡ��Ŀ���')
    return
  }
  if (transferForm.value.sourceTable === transferForm.value.targetTable) {
    message.warning('��Դ����Ŀ���������ͬ')
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
    message.success('���ݻ�ת�ɹ�')
  } catch (err: unknown) {
    message.error(getErrMsg(err, '���ݻ�תʧ��'))
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
    message.warning('��ѡ��Ŀ���')
    return
  }
  if (!importFile.value) {
    message.warning('��ѡ�������ļ�')
    return
  }
  importing.value = true
  try {
    await backupApi.importData(importForm.value.table, importFile.value)
    message.success('���ݵ���ɹ�')
    importFile.value = null
    importUploadRef.value?.clearFiles()
  } catch (err: unknown) {
    message.error(getErrMsg(err, '���ݵ���ʧ��'))
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
    message.warning('��ѡ�����ݱ�')
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
      message.success('���ݵ����ɹ�')
    }
  } catch (err: unknown) {
    message.error(getErrMsg(err, '���ݵ���ʧ��'))
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
