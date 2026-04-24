# 🚀 FayHub 全栈项目开发宪法 (AI 研发行为准则 V3.0 终极版)

## 📌 1. 项目愿景与全局定位
**FayHub** 是下一代 AI 原生的多租户 SaaS 生态底座平台。
- **架构核心：** 极严的数据隔离（多租户）、GVA 风格的分层组管理模式、原生支持 MCP 协议。

## 📂 2. 严格的目录规范 (AI 必读)
**【AI 警告】绝不允许随意起名或随意存放文件！**
- `docs/`：存放所有的说明文档，必须使用**中文或带编号的中文文件夹**（如 `docs/01_产品规划/`）。
- `internal/` 与 `pkg/`：后端代码，必须使用**英文蛇形命名**（如 `user_service.go`）。

---

## � 3. 后端铁律：绝对的多租户安全 (Multi-Tenant) - 【核心保命条款】
**【AI 警告】违反此部分的任何代码将被视为 P0 级严重安全事故！**
1. **绝对禁止全局 DB 直连：** Service/Model 层严禁直接使用 `global.DB`。
2. **强制 Context 链路透传：** HTTP 请求进入 Gin 后，由 `TenantMiddleware` 提取 `X-Tenant-ID` 写入 `gin.Context`。Controller 层必须将 `gin.Context` 转换为 `context.Context`，作为**第一个参数**透传给 Service 层。
3. **强制使用 DB 包装器：** 数据库 CRUD 必须调用 `utils.GetDB(ctx)`。该工具会自动解析 ctx 中的 `tenant_id`，利用 GORM 自动拼装 `WHERE tenant_id = ?`。
4. **Model 继承规范：**
   - 商家业务表（如订单、商品），**必须强制继承 `TenantModel`**（底层带 `tenant_id`）。
   - 全局平台表（如系统设置），继承 `BaseModel`。

---

## 🏗️ 4. 后端 GVA 分层架构与组管理规范
必须严格遵循 `Router -> Controller -> Service -> Model` 单向依赖，**严禁跨层调用**。

1. **`enter.go` 组管理模式（强制）：**
   - 所有 `controller`, `service`, `router` 层必须使用 `enter.go` 暴露各自的 `GroupApp`（如 `service.ServiceGroupApp`），以此作为模块间通信的唯一入口，防止循环引用。
2. **Controller 层 (`internal/controller/`)：**
   - 负责参数校验，调用 Service，返回统一 JSON `response.OkWithData` 或 `response.FailWithMessage`。
   - **必须**为每一个对外 API 编写详尽的 Swagger 注释。
3. **Service 层 (`internal/service/`)：**
   - 函数签名首个参数必须是 `ctx context.Context`。只返回 `(结果, error)`，不处理 HTTP 上下文。

---

## � 5. 前端开发规范 (Vue 3 + Vite)
1. **核心技术栈：** Vue 3.4+ (`<script setup>`) / Pinia / Element Plus / UnoCSS。
2. **API 调用与工具类（强制复用）：**
   - 所有 HTTP 请求**必须**封装在 `src/api/` 下，通过 `@/utils/request.js` 发送，严禁组件内直写 axios。
   - 日期处理优先使用 `@/utils/date.js`。
   - 字典转换、布尔值格式化优先使用 `@/utils/format.js`。
3. **状态与页面：** 优先使用抽屉 (`el-drawer`) 和弹窗 (`el-dialog`) 进行新增/编辑操作，并强制加上 `destroy-on-close` 属性防止内存泄漏。

---

## 🧩 6. 插件化引擎规范 (Plugin Engine)
1. 主干极度精简：底座只保留用户、租户、权限、存储基建。
2. 结构预留：业务模块开发统一以"插件"思想对待，预留 `plugin/` 目录挂载机制。
3. 插件注册必须提供 `RouterPath` 和 `Initialize` 方法。

---

## 🤖 7. 对 Trae / Codebuddy 的最高执行指令
作为辅助开发的 AI 助手，在执行任务时，你必须严格遵守：
1. **优先确认规矩：** 每次被 `@project_rules.md` 唤醒时，优先检查【多租户隔离】和【enter.go 组模式】。
2. **思考优于编码：** 收到需求后，先用中文输出【实现思路与文件存放规划】。
3. **自我审查：** 输出 Go 代码前默念："我传 context 了吗？我绕过租户隔离了吗？我写 Swagger 注释了吗？"
4. **最小破坏原则：** 修改现有文件时，明确指出要修改的行，不要随意删除正确的业务逻辑。