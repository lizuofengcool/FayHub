<template>
  <div class="console-page">
    <div class="stat-cards">
      <div class="stat-card" v-for="card in statCards" :key="card.label">
        <div class="stat-main">
          <div class="stat-info">
            <div class="stat-label">{{ card.label }}</div>
            <div class="stat-value">{{ card.value }}</div>
          </div>
          <div class="stat-chart">
            <svg viewBox="0 0 80 30" class="mini-chart">
              <polyline
                fill="none"
                :stroke="card.trend > 0 ? '#18a058' : '#d03050'"
                stroke-width="2"
                :points="card.chartPoints"
              />
            </svg>
          </div>
        </div>
        <div class="stat-footer">
          <span class="stat-trend" :class="card.trend > 0 ? 'up' : 'down'">
            {{ card.trend > 0 ? '+' : '' }}{{ card.trend }}%
          </span>
          <span class="stat-compare">较昨日</span>
        </div>
      </div>
    </div>

    <div class="content-grid">
      <div class="card">
        <div class="card-header">
          <h3>订单来源</h3>
        </div>
        <div class="chart-body">
          <v-chart :option="orderSourceOption" autoresize style="height:360px" />
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <h3>用户分析</h3>
        </div>
        <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>名称</th>
                <th>头像</th>
                <th>性别</th>
                <th>地区</th>
                <th>资料完善</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in userAnalysis" :key="user.name">
                <td>{{ user.name }}</td>
                <td>
                  <img :src="user.avatar" class="avatar-sm" :alt="user.name" />
                </td>
                <td>
                  <span class="gender-tag" :class="user.gender">{{ user.gender }}</span>
                </td>
                <td>{{ user.region }}</td>
                <td>
                  <div class="progress-wrap">
                    <div class="progress-bar" :style="{ width: user.progress + '%' }"></div>
                  </div>
                  <span class="progress-text">{{ user.progress }}</span>
                </td>
                <td>
                  <button class="btn-link">详情</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div class="content-grid">
      <div class="card">
        <div class="card-header">
          <h3>最新订单</h3>
        </div>
        <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>序号</th>
                <th>下单用户</th>
                <th>商品名称</th>
                <th>商品库存</th>
                <th>订单金额</th>
                <th>商品图片</th>
                <th>付款状态</th>
                <th>客户标签</th>
                <th>购买日期</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(order, idx) in latestOrders" :key="idx">
                <td>{{ idx + 1 }}</td>
                <td>{{ order.user }}</td>
                <td>{{ order.product }}</td>
                <td>{{ order.stock }}</td>
                <td class="amount">&yen;{{ order.amount.toFixed(2) }}</td>
                <td>
                  <img :src="order.image" class="product-thumb" :alt="order.product" />
                </td>
                <td>
                  <span class="pay-status" :class="order.payStatus">{{ order.payStatusText }}</span>
                </td>
                <td>
                  <span class="customer-tag" v-for="tag in order.tags" :key="tag">{{ tag }}</span>
                </td>
                <td>{{ order.date }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <h3>消费排行</h3>
        </div>
        <div class="rank-list">
          <div class="rank-item" v-for="(item, idx) in consumptionRank" :key="idx">
            <span class="rank-num" :class="{ top: idx < 3 }">{{ idx + 1 }}</span>
            <img :src="item.avatar" class="avatar-sm" :alt="item.name" />
            <div class="rank-info">
              <div class="rank-name">{{ item.name }}</div>
              <div class="rank-ip">IP归属地：{{ item.ip }}</div>
            </div>
            <span class="rank-amount">&yen;{{ item.amount.toFixed(2) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, BarChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, LegendComponent, GridComponent } from 'echarts/components'

use([CanvasRenderer, LineChart, BarChart, TitleComponent, TooltipComponent, LegendComponent, GridComponent])

const statCards = [
  {
    label: '周销售额', value: '5.4万', trend: 2.6,
    chartPoints: '0,22 10,20 20,18 30,16 40,14 50,12 60,10 70,8 80,5'
  },
  {
    label: '新用户', value: '12.3万', trend: -0.5,
    chartPoints: '0,18 10,20 20,15 30,22 40,16 50,12 60,14 70,10 80,8'
  },
  {
    label: '采购订单', value: '32.5万', trend: 0.5,
    chartPoints: '0,20 10,18 20,22 30,16 40,14 50,18 60,12 70,10 80,6'
  },
  {
    label: '售后订单', value: '3.8万', trend: -0.8,
    chartPoints: '0,24 10,22 20,20 30,18 40,16 50,14 60,12 70,10 80,8'
  },
]

const orderSourceOption = computed(() => ({
  tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
  legend: {
    data: ['成交订单(个)', '总订单(个)', '转化率(%)'],
    bottom: 0,
    textStyle: { color: 'var(--text-secondary)', fontSize: 12 }
  },
  grid: { left: '3%', right: '4%', bottom: '12%', top: '10px', containLabel: true },
  xAxis: {
    type: 'category',
    data: ['西城区', '顺义区', '朝阳区', '大兴区', '海淀区', '昌平区', '西城区', '东城区', '丰台区'],
    axisLabel: { color: 'var(--text-secondary)', fontSize: 11 },
    axisLine: { lineStyle: { color: 'var(--border-color)' } },
  },
  yAxis: [
    {
      type: 'value',
      name: '个',
      splitLine: { lineStyle: { color: 'var(--border-color)' } },
      axisLabel: { color: 'var(--text-secondary)' },
    },
    {
      type: 'value',
      name: '%',
      splitLine: { show: false },
      axisLabel: { color: 'var(--text-secondary)', formatter: '{value} %' },
    },
  ],
  series: [
    {
      name: '成交订单(个)',
      type: 'bar',
      data: [45, 62, 38, 55, 72, 48, 41, 35, 29],
      itemStyle: { color: '#2d8cf0', borderRadius: [4, 4, 0, 0] },
      barWidth: 12,
    },
    {
      name: '总订单(个)',
      type: 'bar',
      data: [68, 85, 52, 78, 95, 65, 58, 50, 42],
      itemStyle: { color: '#18a058', borderRadius: [4, 4, 0, 0] },
      barWidth: 12,
    },
    {
      name: '转化率(%)',
      type: 'line',
      yAxisIndex: 1,
      data: [66, 73, 73, 71, 76, 74, 71, 70, 69],
      lineStyle: { color: '#f0a020', width: 2 },
      itemStyle: { color: '#f0a020' },
      symbol: 'circle',
      symbolSize: 6,
    },
  ],
}))

const userAnalysis = [
  { name: '小猪妹', avatar: 'https://assets.naiveadmin.com/assets/avatar/avatar-1.jpg', gender: '男', region: '深圳', progress: 35 },
  { name: '妖姬妹', avatar: 'https://assets.naiveadmin.com/assets/avatar/avatar-2.jpg', gender: '女', region: '广州', progress: 45 },
  { name: '爱斯', avatar: 'https://assets.naiveadmin.com/assets/avatar/avatar-3.jpg', gender: '男', region: '北京', progress: 55 },
  { name: '罗威娜', avatar: 'https://assets.naiveadmin.com/assets/avatar/avatar-4.jpg', gender: '女', region: '上海', progress: 65 },
  { name: '小娜', avatar: 'https://assets.naiveadmin.com/assets/avatar/avatar-5.jpg', gender: '男', region: '江苏', progress: 75 },
]

const latestOrders = [
  { user: 'Naive Admin', product: 'Naive Admin Pro', stock: 128, amount: 298.00, image: 'https://assets.naiveadmin.com/assets/product-1.png', payStatus: 'paid', payStatusText: '已付款', tags: ['老客户', '新客户'], date: '2022-09-19' },
  { user: 'Naive Admin', product: 'Naive Admin Pro', stock: 256, amount: 298.00, image: 'https://assets.naiveadmin.com/assets/product-2.png', payStatus: 'paid', payStatusText: '已付款', tags: ['老客户', '新客户'], date: '2022-09-19' },
  { user: 'Naive Admin', product: 'Naive Admin Pro', stock: 64, amount: 298.00, image: 'https://assets.naiveadmin.com/assets/product-3.png', payStatus: 'unpaid', payStatusText: '未付款', tags: ['老客户', '新客户'], date: '2022-09-19' },
  { user: 'Naive Admin', product: 'Naive Admin Pro', stock: 32, amount: 298.00, image: 'https://assets.naiveadmin.com/assets/product-4.png', payStatus: 'cancelled', payStatusText: '已取消', tags: ['老客户', '新客户'], date: '2022-09-19' },
  { user: 'Naive Admin', product: 'Naive Admin Pro', stock: 16, amount: 298.00, image: 'https://assets.naiveadmin.com/assets/product-5.png', payStatus: 'cancelled', payStatusText: '已取消', tags: ['老客户', '新客户'], date: '2022-09-19' },
]

const consumptionRank = [
  { name: 'Naive Admin', ip: '北京市/东城区', amount: 69800.00, avatar: 'https://assets.naiveadmin.com/assets/avatar/avatar-1.jpg' },
  { name: 'Naive Admin', ip: '上海市/黄浦区', amount: 59800.00, avatar: 'https://assets.naiveadmin.com/assets/avatar/avatar-2.jpg' },
  { name: 'Naive Admin', ip: '浙江省/杭州市', amount: 49800.00, avatar: 'https://assets.naiveadmin.com/assets/avatar/avatar-3.jpg' },
  { name: 'Naive Admin', ip: '江西省/萍乡市', amount: 39800.00, avatar: 'https://assets.naiveadmin.com/assets/avatar/avatar-4.jpg' },
  { name: 'Naive Admin', ip: '广东省/深圳市', amount: 29800.00, avatar: 'https://assets.naiveadmin.com/assets/avatar/avatar-5.jpg' },
  { name: 'Naive Admin', ip: '湖南省/衡阳市', amount: 19800.00, avatar: 'https://assets.naiveadmin.com/assets/avatar/avatar-1.jpg' },
]
</script>

<style scoped>
.console-page {
  padding: 0;
}

.stat-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}
@media (max-width: 1200px) {
  .stat-cards { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 640px) {
  .stat-cards { grid-template-columns: 1fr; }
}

.stat-card {
  background: var(--card-bg);
  border-radius: var(--radius);
  box-shadow: var(--card-shadow);
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.stat-main {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.stat-label {
  font-size: 14px;
  color: var(--text-secondary);
  margin-bottom: 6px;
}
.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-title);
}

.stat-chart {
  width: 80px;
  height: 30px;
  flex-shrink: 0;
}
.mini-chart {
  width: 100%;
  height: 100%;
}

.stat-footer {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}
.stat-trend { font-weight: 500; }
.stat-trend.up { color: var(--success); }
.stat-trend.down { color: var(--danger); }
.stat-compare { color: var(--text-muted); }

.content-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 16px;
}
@media (max-width: 1024px) {
  .content-grid { grid-template-columns: 1fr; }
}

.card {
  background: var(--card-bg);
  border-radius: var(--radius);
  box-shadow: var(--card-shadow);
  overflow: hidden;
  padding: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.card-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-title);
  margin: 0;
}

.chart-body {
  width: 100%;
}

.table-wrap {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
  min-width: 600px;
}
th {
  text-align: left;
  padding: 10px 12px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  background: rgba(0,0,0,0.02);
  border-bottom: 0.8px solid var(--border-color);
  white-space: nowrap;
}
td {
  padding: 10px 12px;
  font-size: 14px;
  color: var(--text-primary);
  border-bottom: 0.8px solid var(--border-color);
  white-space: nowrap;
}
tr:hover td {
  background: rgba(0,0,0,0.01);
}

.avatar-sm {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
}

.gender-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}
.gender-tag.男 { background: rgba(45,140,240,0.1); color: #2d8cf0; }
.gender-tag.女 { background: rgba(240,87,108,0.1); color: #f5576c; }

.progress-wrap {
  display: inline-block;
  width: 60px;
  height: 6px;
  background: rgba(0,0,0,0.06);
  border-radius: 3px;
  overflow: hidden;
  vertical-align: middle;
  margin-right: 6px;
}
.progress-bar {
  height: 100%;
  background: var(--primary);
  border-radius: 3px;
  transition: width 0.3s;
}
.progress-text {
  font-size: 12px;
  color: var(--text-muted);
}

.btn-link {
  padding: 4px 12px;
  border: 0.8px solid var(--border-color);
  background: none;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-secondary);
  border-radius: 4px;
  transition: all 0.15s;
}
.btn-link:hover { color: var(--primary); border-color: var(--primary); }

.amount {
  font-weight: 500;
  color: var(--text-title);
}

.product-thumb {
  width: 36px;
  height: 36px;
  border-radius: 4px;
  object-fit: cover;
}

.pay-status {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 500;
}
.pay-status.paid { background: rgba(24,160,88,0.1); color: var(--success); }
.pay-status.unpaid { background: rgba(240,160,32,0.1); color: var(--warning); }
.pay-status.cancelled { background: rgba(208,48,80,0.1); color: var(--danger); }

.customer-tag {
  display: inline-block;
  padding: 1px 8px;
  border-radius: 4px;
  font-size: 12px;
  background: rgba(0,0,0,0.04);
  color: var(--text-secondary);
  margin-right: 4px;
}

.rank-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.rank-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
}
.rank-num {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  font-size: 12px;
  font-weight: 600;
  background: rgba(0,0,0,0.06);
  color: var(--text-secondary);
  flex-shrink: 0;
}
.rank-num.top {
  background: var(--primary);
  color: #fff;
}
.rank-info {
  flex: 1;
  min-width: 0;
}
.rank-name {
  font-size: 14px;
  color: var(--text-primary);
  font-weight: 500;
}
.rank-ip {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 2px;
}
.rank-amount {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-title);
  flex-shrink: 0;
}
</style>
