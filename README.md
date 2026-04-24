# FayHub

**AI 原生多租户 SaaS 插件化底座（Go）**
一款面向企业级 B2B 场景的开源 SaaS 底座，内置强隔离多租户、热插拔插件体系，支持免费商用，禁止竞品套娃。

---

## 🔍 项目定位

FayHub 是**企业级多租户 SaaS 开发底座**，专注解决：
- 多租户数据安全隔离（杜绝串库）
- 插件化模块化开发（热插拔、不重启）
- 快速交付私有化/ SaaS 商用系统
- 内置权限、租户、用户核心基建

不搞虚概念，只做**能落地、能商用、能赚钱**的 SaaS 底层框架。

---

## ✅ 核心能力

1. **强隔离多租户引擎**
   - 基于 Context + GORM Hook 自动注入 tenant_id
   - 业务层无感知，强制租户数据隔离
   - 支持共享库/独立库灵活切换

2. **微内核插件化架构**
   - 核心仅保留：租户、用户、权限、基础 API
   - 业务模块全插件化，安装/卸载/升级不重启
   - 主程序与插件进程隔离，稳定安全

3. **生产级标准架构**
   - Go + Gin + GORM 企业级技术栈
   - 支持 MySQL / PostgreSQL / SQLite
   - 规范分层：Controller → Service → Repository
   - 适配私有化部署、SaaS 云托管、外包交付

---

## 🧱 技术栈

- 后端：Go 1.22+ / Gin / GORM
- 数据库：MySQL 8.0+ / PostgreSQL 16+ / SQLite
- 插件通信：gRPC + Protobuf
- 前端：Vue3 + Element Plus（规划中）

---

## 💰 商用说明（开发者最关心｜一眼看懂）

### ✅ 完全免费允许（无任何限制）

- 企业内部系统开发使用
- 基于 FayHub 开发**自己的业务 SaaS**（商城/CRM/ERP/管理系统）并售卖
- 为客户做**定制开发+私有化部署交付**
- 二次修改、内部迭代、公司商用

### ❌ 严格禁止（核心保护条款）

- 禁止将 FayHub 改名为其他框架**二次售卖/开源**
- 禁止基于本项目搭建**插件交易市场/应用商店**并抽成
- 禁止将本底座作为 PaaS 平台向第三方开发者收费

> 简单说：**你做自己的生意完全免费，禁止拿它跟我抢饭碗**。

---

## 🚀 快速启动

```bash
# 1. 克隆代码
git clone https://github.com/lizuofengcool/FayHub.git
cd FayHub/FayHub

# 2. 安装依赖
go mod tidy

# 3. 启动服务
go run cmd/main.go

# 4. 访问测试
curl -H "X-Tenant-ID: 1001" http://localhost:8080/api/health
```

---

## 📁 项目结构

```
FayHub/
├── cmd/                    # 程序入口
├── docs/                   # 产品/架构/开发文档
├── internal/               # 核心业务代码（私有）
│   ├── controller          # 控制器
│   ├── service             # 业务逻辑
│   ├── repository          # 数据访问
│   ├── middleware          # 中间件（租户/认证）
│   ├── model               # 数据模型
│   └── router              # 路由
├── pkg/                    # 公共工具
├── LICENSE                 # 许可协议
└── go.mod                  # 依赖管理
```

---

## � 许可协议

本项目采用 **FayHub License（Fair-Code 模式）**
- 基于 MIT 开源，免费商用
- 禁止竞品套娃、禁止插件市场抽成、禁止二次售卖底座
- 完整协议：[LICENSE](./LICENSE)

---

## 🤝 贡献与反馈

- GitHub：https://github.com/lizuofengcool/FayHub
- 问题反馈：Issues
- 商业授权：商务合作请私信
