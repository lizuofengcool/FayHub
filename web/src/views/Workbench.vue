<template>
  <div class="workbench-page">
    <div class="content-grid">
      <div class="card">
        <div class="card-header">
          <h3>待办事项</h3>
          <button class="btn-text">+ 新增</button>
        </div>
        <div class="todo-list">
          <div class="todo-item" v-for="todo in todos" :key="todo.id" :class="{ done: todo.done }">
            <button class="todo-check" :class="{ checked: todo.done }" @click="todo.done = !todo.done">
              <span v-if="todo.done">✓</span>
            </button>
            <div class="todo-content">
              <div class="todo-title">{{ todo.title }}</div>
              <div class="todo-meta">
                <span class="todo-tag" :class="todo.priority">{{ todo.priorityText }}</span>
                <span class="todo-date">{{ todo.date }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <h3>快捷入口</h3>
        </div>
        <div class="shortcut-grid">
          <button class="shortcut-item" v-for="sc in shortcuts" :key="sc.label" @click="$router.push(sc.path)">
            <div class="shortcut-icon" :style="{ background: sc.bg }">
              <span v-html="sc.icon"></span>
            </div>
            <span class="shortcut-label">{{ sc.label }}</span>
          </button>
        </div>
      </div>

      <div class="card chart-card">
        <div class="card-header">
          <h3>项目进度</h3>
        </div>
        <div class="project-list">
          <div class="project-item" v-for="proj in projects" :key="proj.name">
            <div class="project-info">
              <div class="project-name">{{ proj.name }}</div>
              <div class="project-meta">{{ proj.members }}人 · {{ proj.tasks }}个任务</div>
            </div>
            <div class="project-progress">
              <div class="progress-bar">
                <div class="progress-fill" :style="{ width: proj.progress + '%', background: proj.color }"></div>
              </div>
              <span class="progress-text">{{ proj.progress }}%</span>
            </div>
          </div>
        </div>
      </div>

      <div class="card chart-card">
        <div class="card-header">
          <h3>动态</h3>
        </div>
        <div class="timeline">
          <div class="timeline-item" v-for="item in timeline" :key="item.id">
            <div class="timeline-dot" :style="{ background: item.color }"></div>
            <div class="timeline-content">
              <div class="timeline-title">
                <strong>{{ item.user }}</strong> {{ item.action }}
              </div>
              <div class="timeline-time">{{ item.time }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const todos = ref([
  { id: 1, title: '审核新注册租户"星辰科技"', done: false, priority: 'high', priorityText: '高优', date: '2026-05-07' },
  { id: 2, title: '处理用户反馈关于支付异常的问题', done: false, priority: 'high', priorityText: '高优', date: '2026-05-07' },
  { id: 3, title: '更新系统安全补丁到最新版本', done: false, priority: 'medium', priorityText: '中优', date: '2026-05-08' },
  { id: 4, title: '编写本月运营数据分析报告', done: false, priority: 'low', priorityText: '低优', date: '2026-05-10' },
  { id: 5, title: '整理插件市场上架审核规范文档', done: true, priority: 'low', priorityText: '低优', date: '2026-05-05' },
])

const shortcuts = [
  { label: '租户管理', icon: '🏢', bg: 'rgba(45,140,240,0.1)', path: '/system/tenant' },
  { label: '用户管理', icon: '👥', bg: 'rgba(24,160,88,0.1)', path: '/system/user' },
  { label: '订单管理', icon: '📋', bg: 'rgba(240,160,32,0.1)', path: '/payment/transactions' },
  { label: '插件市场', icon: '🧩', bg: 'rgba(208,48,80,0.1)', path: '/plugins/installed' },
  { label: '系统设置', icon: '⚙', bg: 'rgba(102,126,234,0.1)', path: '/system/settings' },
  { label: '审计日志', icon: '📝', bg: 'rgba(17,153,142,0.1)', path: '/system/audit' },
]

const projects = [
  { name: 'FayHub v2.0 升级', members: 8, tasks: 24, progress: 78, color: '#2d8cf0' },
  { name: 'Market 插件审核系统', members: 5, tasks: 16, progress: 45, color: '#18a058' },
  { name: 'SSO 单点登录优化', members: 3, tasks: 8, progress: 92, color: '#f0a020' },
  { name: '移动端适配', members: 6, tasks: 20, progress: 30, color: '#d03050' },
]

const timeline = [
  { id: 1, user: '张三', action: '创建了新租户"云端数据"', time: '10分钟前', color: '#2d8cf0' },
  { id: 2, user: '李四', action: '提交了插件"AI助手"的审核申请', time: '30分钟前', color: '#18a058' },
  { id: 3, user: '王五', action: '完成了订单 FH20260507001 的支付', time: '1小时前', color: '#f0a020' },
  { id: 4, user: '系统', action: '自动备份数据库完成', time: '2小时前', color: '#d03050' },
  { id: 5, user: '赵六', action: '更新了用户权限配置', time: '3小时前', color: '#2080f0' },
]
</script>

<style scoped>
.workbench-page { padding: 0; }

.content-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
@media (max-width: 1024px) { .content-grid { grid-template-columns: 1fr; } }

.card {
  background: var(--card-bg);
  border-radius: var(--radius);
  box-shadow: var(--card-shadow);
  overflow: hidden;
  padding: 20px;
}

.card-header {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 16px;
}
.card-header h3 { font-size: 16px; font-weight: 600; color: var(--text-title); margin: 0; }

.btn-text {
  padding: 4px 12px; border: none; background: none; cursor: pointer;
  font-size: 13px; color: var(--primary); border-radius: 4px; transition: all 0.15s;
}
.btn-text:hover { background: var(--primary-suppl); }

.todo-item {
  display: flex; align-items: flex-start; gap: 12px;
  padding: 10px 0; border-bottom: 0.8px solid var(--border-color);
}
.todo-item:last-child { border-bottom: none; }
.todo-item.done { opacity: 0.5; }
.todo-item.done .todo-title { text-decoration: line-through; }

.todo-check {
  width: 20px; height: 20px; border-radius: 50%;
  border: 2px solid var(--border-color); cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  font-size: 12px; color: #fff; flex-shrink: 0; margin-top: 2px;
  background: none; padding: 0; transition: all 0.15s;
}
.todo-check.checked { background: var(--success); border-color: var(--success); }
.todo-check:hover { border-color: var(--primary); }

.todo-content { flex: 1; min-width: 0; }
.todo-title { font-size: 14px; color: var(--text-primary); margin-bottom: 4px; }
.todo-meta { display: flex; gap: 8px; align-items: center; }
.todo-tag {
  padding: 1px 8px; border-radius: 8px; font-size: 11px; font-weight: 500;
}
.todo-tag.high { background: rgba(208,48,80,0.1); color: var(--danger); }
.todo-tag.medium { background: rgba(240,160,32,0.1); color: var(--warning); }
.todo-tag.low { background: rgba(45,140,240,0.1); color: var(--primary); }
.todo-date { font-size: 12px; color: var(--text-muted); }

.shortcut-grid {
  display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px;
}
.shortcut-item {
  display: flex; flex-direction: column; align-items: center; gap: 8px;
  padding: 16px 8px; border-radius: var(--radius); cursor: pointer;
  border: 0.8px solid var(--border-color); background: none;
  transition: all 0.15s;
}
.shortcut-item:hover { border-color: var(--primary); background: var(--primary-suppl); }
.shortcut-icon {
  width: 44px; height: 44px; border-radius: 12px;
  display: flex; align-items: center; justify-content: center; font-size: 20px;
}
.shortcut-label { font-size: 13px; color: var(--text-primary); }

.project-item {
  padding: 12px 0; border-bottom: 0.8px solid var(--border-color);
}
.project-item:last-child { border-bottom: none; }
.project-info { margin-bottom: 8px; }
.project-name { font-size: 14px; color: var(--text-primary); font-weight: 500; }
.project-meta { font-size: 12px; color: var(--text-muted); margin-top: 2px; }
.project-progress { display: flex; align-items: center; gap: 8px; }

.progress-bar {
  flex: 1; height: 6px; background: rgba(0,0,0,0.06);
  border-radius: 3px; overflow: hidden;
}
.progress-fill { height: 100%; border-radius: 3px; transition: width 0.5s; }
.progress-text { font-size: 12px; color: var(--text-muted); width: 36px; text-align: right; }

.timeline { padding: 0; }
.timeline-item {
  display: flex; gap: 12px; padding: 10px 0;
  border-bottom: 0.8px solid var(--border-color);
}
.timeline-item:last-child { border-bottom: none; }
.timeline-dot {
  width: 10px; height: 10px; border-radius: 50%; margin-top: 4px; flex-shrink: 0;
}
.timeline-content { flex: 1; min-width: 0; }
.timeline-title { font-size: 14px; color: var(--text-primary); }
.timeline-title strong { font-weight: 600; }
.timeline-time { font-size: 12px; color: var(--text-muted); margin-top: 2px; }
</style>
