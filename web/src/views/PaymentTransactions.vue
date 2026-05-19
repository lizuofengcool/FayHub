﻿<template>
  <div class="payment-transactions-page">
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm">
      <div class="p-4 pb-3 flex items-center justify-between">
        <div>
          <h2 class="text-lg font-bold text-slate-800">交易记录</h2>
          <p class="text-slate-400 text-xs mt-0.5">查看平台所有交易流水与结算状态</p>
        </div>
        <div class="flex gap-3">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            @change="loadTransactions"
          />
          <el-button @click="loadTransactions" :loading="loading">
            <el-icon class="mr-1"><Refresh /></el-icon> 刷新
          </el-button>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-4 gap-4 px-4 mb-4">
      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-6">
        <div class="flex items-center justify-between mb-4">
          <span class="text-sm font-medium text-slate-500">总交易额</span>
          <el-icon class="text-2xl text-green-500"><TrendCharts /></el-icon>
        </div>
        <p class="text-3xl font-bold text-slate-800">¥{{ stats.totalAmount }}</p>
        <p class="text-sm text-slate-500 mt-1">累计交易金额</p>
      </div>

      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-6">
        <div class="flex items-center justify-between mb-4">
          <span class="text-sm font-medium text-slate-500">交易笔数</span>
          <el-icon class="text-2xl text-blue-500"><Document /></el-icon>
        </div>
        <p class="text-3xl font-bold text-slate-800">{{ stats.totalCount }}</p>
        <p class="text-sm text-slate-500 mt-1">累计交易笔数</p>
      </div>

      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-6">
        <div class="flex items-center justify-between mb-4">
          <span class="text-sm font-medium text-slate-500">平台收入</span>
          <el-icon class="text-2xl text-amber-500"><Coin /></el-icon>
        </div>
        <p class="text-3xl font-bold text-slate-800">¥{{ stats.platformIncome }}</p>
        <p class="text-sm text-slate-500 mt-1">抽佣收入</p>
      </div>

      <div class="bg-white rounded-2xl shadow-sm border border-slate-100 p-6">
        <div class="flex items-center justify-between mb-4">
          <span class="text-sm font-medium text-slate-500">待结算</span>
          <el-icon class="text-2xl text-orange-500"><Timer /></el-icon>
        </div>
        <p class="text-3xl font-bold text-slate-800">¥{{ stats.pendingSettlement }}</p>
        <p class="text-sm text-slate-500 mt-1">待结算金额</p>
      </div>
    </div>

    <div class="bg-white rounded-2xl shadow-sm border border-slate-100">
      <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
        <h3 class="text-lg font-semibold text-slate-800">交易明细</h3>
        <el-select v-model="statusFilter" placeholder="交易状态" style="width: 140px" @change="loadTransactions" clearable>
          <el-option label="全部" value="" />
          <el-option label="成功" value="success" />
          <el-option label="待支付" value="pending" />
          <el-option label="已退款" value="refunded" />
          <el-option label="已关闭" value="closed" />
        </el-select>
      </div>

      <el-table :data="transactions" stripe class="w-full" empty-text="暂无交易记录">
        <el-table-column prop="order_no" label="订单号" min-width="180" />
        <el-table-column prop="plugin_name" label="插件名称" min-width="120" />
        <el-table-column prop="buyer" label="购买方" min-width="120" />
        <el-table-column label="金额" width="120">
          <template #default="{ row }">
            <span class="font-semibold" :class="row.status === 'refunded' ? 'text-red-500' : 'text-green-600'">
              {{ row.status === 'refunded' ? '-' : '+' }}¥{{ row.amount }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <n-tag :type="statusTagType(row.status)" size="small">{{ statusLabel(row.status) }}</n-tag>
          </template>
        </el-table-column>
        <el-table-column prop="pay_method" label="支付方式" width="100" />
        <el-table-column prop="created_at" label="交易时间" width="180" />
      </el-table>

      <div class="px-6 py-4 flex justify-end">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @size-change="loadTransactions"
          @current-change="loadTransactions"
        />
      </div>
    </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import { Refresh, TrendCharts, Document, Coin, Timer } from '@element-plus/icons-vue'
import request from '@/api/request'

const message = useMessage()

interface Transaction {
  order_no: string
  plugin_name: string
  buyer: string
  amount: string
  status: string
  pay_method: string
  created_at: string
}

interface TransactionStats {
  totalAmount: string
  totalCount: number
  platformIncome: string
  pendingSettlement: string
}

const loading = ref(false)
const transactions = ref<Transaction[]>([])
const stats = ref<TransactionStats>({
  totalAmount: '0.00',
  totalCount: 0,
  platformIncome: '0.00',
  pendingSettlement: '0.00'
})
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const statusFilter = ref('')
const dateRange = ref<string[]>([])

function statusTagType(status: string): 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, 'success' | 'warning' | 'danger' | 'info'> = {
    success: 'success',
    pending: 'warning',
    refunded: 'danger',
    closed: 'info'
  }
  return map[status] || 'info'
}

function statusLabel(status: string): string {
  const map: Record<string, string> = {
    success: '成功',
    pending: '待支付',
    refunded: '已退款',
    closed: '已关闭'
  }
  return map[status] || status
}

async function loadTransactions() {
  loading.value = true
  try {
    const params: Record<string, string> = {
      page: String(page.value),
      page_size: String(pageSize.value)
    }
    if (statusFilter.value) params.status = statusFilter.value
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }

    const data = await request.get('/payment/transactions', { params })
    if (data.data) {
      transactions.value = data.data.list || []
      total.value = data.data.total || 0
      if (data.data.stats) {
        stats.value = { ...stats.value, ...data.data.stats }
      }
    }
  } catch (e: any) {
    message.error('加载交易记录失败: ' + (e.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadTransactions()
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

:deep(.el-date-editor.el-input__wrapper) {
  height: 32px;
}
</style>
