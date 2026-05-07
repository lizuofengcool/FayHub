<template>
  <div class="monitor-page">
    <div class="stat-cards">
      <div class="stat-card" v-for="card in statCards" :key="card.label">
        <div class="stat-info">
          <div class="stat-label">{{ card.label }}</div>
          <div class="stat-value">
            <span class="stat-num">{{ card.value }}</span>
            <span class="stat-unit" v-if="card.unit">{{ card.unit }}</span>
          </div>
        </div>
        <div class="stat-icon" :style="{ background: card.bg }">
          <span v-html="card.icon"></span>
        </div>
      </div>
    </div>

    <div class="content-grid">
      <div class="card chart-card">
        <div class="card-header">
          <h3>CPU 使用率</h3>
          <span class="card-badge" :class="cpuStatus">{{ cpuStatusText }}</span>
        </div>
        <div class="chart-body">
          <v-chart :option="cpuChartOption" autoresize style="height:280px" />
        </div>
      </div>

      <div class="card chart-card">
        <div class="card-header">
          <h3>内存使用</h3>
          <span class="card-badge normal">正常</span>
        </div>
        <div class="chart-body">
          <v-chart :option="memoryChartOption" autoresize style="height:280px" />
        </div>
      </div>

      <div class="card chart-card">
        <div class="card-header">
          <h3>磁盘 IO</h3>
        </div>
        <div class="chart-body">
          <v-chart :option="diskIOOption" autoresize style="height:280px" />
        </div>
      </div>

      <div class="card chart-card">
        <div class="card-header">
          <h3>网络流量</h3>
        </div>
        <div class="chart-body">
          <v-chart :option="networkOption" autoresize style="height:280px" />
        </div>
      </div>
    </div>

    <div class="card table-card">
      <div class="card-header">
        <h3>服务健康状态</h3>
      </div>
      <table>
        <thead>
          <tr>
            <th>服务名称</th>
            <th>状态</th>
            <th>响应时间</th>
            <th>可用率</th>
            <th>最后检查</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="svc in services" :key="svc.name">
            <td>{{ svc.name }}</td>
            <td>
              <span class="status-dot" :class="svc.status"></span>
              {{ svc.statusText }}
            </td>
            <td>{{ svc.responseTime }}ms</td>
            <td>
              <div class="progress-bar">
                <div class="progress-fill" :style="{ width: svc.uptime + '%', background: svc.uptime > 99 ? 'var(--success)' : 'var(--warning)' }"></div>
              </div>
              <span class="progress-text">{{ svc.uptime }}%</span>
            </td>
            <td>{{ svc.lastCheck }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, BarChart, GaugeChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, LegendComponent, GridComponent } from 'echarts/components'

use([CanvasRenderer, LineChart, BarChart, GaugeChart, TitleComponent, TooltipComponent, LegendComponent, GridComponent])

const statCards = [
  { label: 'CPU 使用率', value: '23', unit: '%', icon: '🖥', bg: 'rgba(45,140,240,0.1)' },
  { label: '内存使用', value: '6.2', unit: 'GB / 16GB', icon: '💾', bg: 'rgba(24,160,88,0.1)' },
  { label: '磁盘使用', value: '128', unit: 'GB / 500GB', icon: '💿', bg: 'rgba(240,160,32,0.1)' },
  { label: '网络带宽', value: '3.8', unit: 'Mbps', icon: '🌐', bg: 'rgba(208,48,80,0.1)' },
]

const cpuStatus = computed(() => 23 > 80 ? 'danger' : 23 > 60 ? 'warning' : 'normal')
const cpuStatusText = computed(() => 23 > 80 ? '高负载' : 23 > 60 ? '中等' : '正常')

const timeLabels = ['14:00', '14:05', '14:10', '14:15', '14:20', '14:25', '14:30', '14:35', '14:40', '14:45', '14:50', '14:55']

const cpuChartOption = {
  tooltip: { trigger: 'axis' },
  grid: { left: '3%', right: '4%', bottom: '3%', top: '10px', containLabel: true },
  xAxis: { type: 'category', data: timeLabels, axisLabel: { color: 'var(--text-secondary)', fontSize: 11 }, axisLine: { lineStyle: { color: 'var(--border-color)' } } },
  yAxis: { type: 'value', max: 100, splitLine: { lineStyle: { color: 'var(--border-color)' } }, axisLabel: { color: 'var(--text-secondary)' } },
  series: [{
    data: [18, 22, 25, 20, 28, 23, 30, 26, 22, 19, 24, 23],
    type: 'line',
    smooth: true,
    symbol: 'none',
    lineStyle: { color: '#2d8cf0', width: 2 },
    areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(45,140,240,0.25)' }, { offset: 1, color: 'rgba(45,140,240,0.02)' }] } },
  }],
}

const memoryChartOption = {
  tooltip: { trigger: 'axis' },
  grid: { left: '3%', right: '4%', bottom: '3%', top: '10px', containLabel: true },
  xAxis: { type: 'category', data: timeLabels, axisLabel: { color: 'var(--text-secondary)', fontSize: 11 }, axisLine: { lineStyle: { color: 'var(--border-color)' } } },
  yAxis: { type: 'value', splitLine: { lineStyle: { color: 'var(--border-color)' } }, axisLabel: { color: 'var(--text-secondary)' } },
  series: [
    { name: '已用', data: [5.8, 5.9, 6.0, 6.1, 6.0, 6.2, 6.3, 6.1, 6.2, 6.0, 6.1, 6.2], type: 'line', smooth: true, symbol: 'none', lineStyle: { color: '#18a058', width: 2 }, areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(24,160,88,0.25)' }, { offset: 1, color: 'rgba(24,160,88,0.02)' }] } } },
    { name: '缓存', data: [2.1, 2.2, 2.0, 2.3, 2.1, 2.0, 2.2, 2.1, 2.0, 2.2, 2.1, 2.0], type: 'line', smooth: true, symbol: 'none', lineStyle: { color: '#f0a020', width: 2, type: 'dashed' }, areaStyle: { color: 'transparent' } },
  ],
}

const diskIOOption = {
  tooltip: { trigger: 'axis' },
  grid: { left: '3%', right: '4%', bottom: '3%', top: '10px', containLabel: true },
  xAxis: { type: 'category', data: timeLabels, axisLabel: { color: 'var(--text-secondary)', fontSize: 11 }, axisLine: { lineStyle: { color: 'var(--border-color)' } } },
  yAxis: { type: 'value', splitLine: { lineStyle: { color: 'var(--border-color)' } }, axisLabel: { color: 'var(--text-secondary)' } },
  series: [
    { name: '读取', data: [12, 15, 18, 14, 20, 16, 22, 18, 15, 13, 17, 14], type: 'bar', barWidth: 8, itemStyle: { color: '#2d8cf0', borderRadius: [4, 4, 0, 0] } },
    { name: '写入', data: [8, 10, 12, 9, 14, 11, 15, 12, 10, 8, 11, 9], type: 'bar', barWidth: 8, itemStyle: { color: '#18a058', borderRadius: [4, 4, 0, 0] } },
  ],
}

const networkOption = {
  tooltip: { trigger: 'axis' },
  grid: { left: '3%', right: '4%', bottom: '3%', top: '10px', containLabel: true },
  xAxis: { type: 'category', data: timeLabels, axisLabel: { color: 'var(--text-secondary)', fontSize: 11 }, axisLine: { lineStyle: { color: 'var(--border-color)' } } },
  yAxis: { type: 'value', splitLine: { lineStyle: { color: 'var(--border-color)' } }, axisLabel: { color: 'var(--text-secondary)' } },
  series: [
    { name: '入站', data: [3.2, 3.5, 4.0, 3.8, 4.2, 3.9, 4.5, 4.1, 3.7, 3.4, 3.8, 3.6], type: 'line', smooth: true, symbol: 'none', lineStyle: { color: '#2d8cf0', width: 2 } },
    { name: '出站', data: [1.8, 2.0, 2.2, 2.1, 2.5, 2.3, 2.8, 2.4, 2.1, 1.9, 2.2, 2.0], type: 'line', smooth: true, symbol: 'none', lineStyle: { color: '#d03050', width: 2 } },
  ],
}

const services = [
  { name: 'FayHub API', status: 'up', statusText: '正常', responseTime: 12, uptime: 99.99, lastCheck: '14:55:01' },
  { name: 'Market API', status: 'up', statusText: '正常', responseTime: 18, uptime: 99.95, lastCheck: '14:55:02' },
  { name: 'SSO 认证服务', status: 'up', statusText: '正常', responseTime: 8, uptime: 99.99, lastCheck: '14:55:03' },
  { name: '数据库主库', status: 'up', statusText: '正常', responseTime: 3, uptime: 99.99, lastCheck: '14:55:04' },
  { name: 'Redis 缓存', status: 'up', statusText: '正常', responseTime: 1, uptime: 99.99, lastCheck: '14:55:05' },
  { name: '消息队列', status: 'up', statusText: '正常', responseTime: 5, uptime: 99.98, lastCheck: '14:55:06' },
]
</script>

<style scoped>
.monitor-page { padding: 0; }

.stat-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}
@media (max-width: 1200px) { .stat-cards { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 640px) { .stat-cards { grid-template-columns: 1fr; } }

.stat-card {
  background: var(--card-bg);
  border-radius: var(--radius);
  box-shadow: var(--card-shadow);
  padding: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.stat-label { font-size: 14px; color: var(--text-secondary); margin-bottom: 8px; }
.stat-num { font-size: 28px; font-weight: 700; color: var(--text-title); }
.stat-unit { font-size: 14px; color: var(--text-secondary); margin-left: 4px; }
.stat-icon {
  width: 48px; height: 48px; border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
  font-size: 22px;
}

.content-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 16px;
}
@media (max-width: 1024px) { .content-grid { grid-template-columns: 1fr; } }

.card {
  background: var(--card-bg);
  border-radius: var(--radius);
  box-shadow: var(--card-shadow);
  overflow: hidden;
}
.chart-card { padding: 20px; }
.table-card { padding: 20px; margin-bottom: 16px; }

.card-header {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 16px;
}
.card-header h3 { font-size: 16px; font-weight: 600; color: var(--text-title); margin: 0; }

.card-badge {
  padding: 2px 10px; border-radius: 10px; font-size: 12px; font-weight: 500;
}
.card-badge.normal { background: rgba(24,160,88,0.1); color: var(--success); }
.card-badge.warning { background: rgba(240,160,32,0.1); color: var(--warning); }
.card-badge.danger { background: rgba(208,48,80,0.1); color: var(--danger); }

.chart-body { width: 100%; }

table { width: 100%; border-collapse: collapse; }
th {
  text-align: left; padding: 10px 12px; font-size: 13px; font-weight: 500;
  color: var(--text-secondary); background: rgba(0,0,0,0.02);
  border-bottom: 0.8px solid var(--border-color);
}
td {
  padding: 12px; font-size: 14px; color: var(--text-primary);
  border-bottom: 0.8px solid var(--border-color);
}
tr:hover td { background: rgba(0,0,0,0.01); }

.status-dot {
  display: inline-block; width: 8px; height: 8px; border-radius: 50%; margin-right: 6px;
}
.status-dot.up { background: var(--success); }
.status-dot.down { background: var(--danger); }

.progress-bar {
  display: inline-block; width: 80px; height: 6px;
  background: rgba(0,0,0,0.06); border-radius: 3px; overflow: hidden;
  vertical-align: middle; margin-right: 8px;
}
.progress-fill { height: 100%; border-radius: 3px; transition: width 0.5s; }
.progress-text { font-size: 12px; color: var(--text-muted); }
</style>
