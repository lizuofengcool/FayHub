# WASM示例插件

## 概述

这是一个完整的 FayHub 插件示例，展示如何开发一个功能齐全的 WASM 插件开发方法。

## 功能特性

- ✅ 日志写入
- ✅ HTTP请求
- ✅ 数据库操作
- ✅ 缓存操作
- ✅ 自定义路由
- ✅ API端点
- ✅ 菜单注册

## 文件结构

```
wasm_plugin/
├── manifest.json       # 插件清单
├── README.md         # 插件说明
├── main.go          # 源代码（Go语言示例）
└── build.sh        # 构建脚本
```

## 快速开始

### 1. 配置清单

manifest.json 包含了插件的所有配置信息。

### 2. 开发WASM模块

使用你喜欢的语言编写代码：

- Go (推荐)
- Rust
- AssemblyScript
- C/C++

### 3. 编译为WASM

```bash
# Go示例
GOOS=wasip1 GOARCH=wasm go build -o main.wasm
```

### 4. 安装插件

在 FayHub 系统中安装此插件。

## manifest.json 详解

### 基本信息

```json
{
  "name": "插件名称",
  "version": "版本号",
  "description": "插件描述"
}
```

### 权限声明

```json
{
  "permissions": [
    "log:write",
    "db:read",
    "http:get"
  ]
}
```

### 路由注册

```json
{
  "routes": [
    {
      "method": "GET",
      "path": "/hello",
      "handler": "handle_hello"
    }
  ]
}
```

### API注册

```json
{
  "apis": [
    {
      "method": "GET",
      "path": "/api/info",
      "group": "plugin"
    }
  ]
}
```

## Host Functions 使用

### host_log

```c
#include <stdio.h>

void host_log(const char* msg, int len) {
    // 实现日志写入
}

// 使用示例
host_log("Hello World", 11);
```

### host_http_request

发起HTTP请求

```c
int host_http_request(
    const char* method, int method_len,
    const char* url, int url_len,
    const char* body, int body_len,
    char* result, int result_max_len
);
```

### 数据库操作

```c
int host_db_query(const char* query, int query_len, char* result, int max_len);
int host_db_exec(const char* query, int query_len);
```

### 缓存操作

```c
int host_cache_get(const char* key, int key_len, char* result, int max_len);
void host_cache_set(const char* key, int key_len, const char* value, int value_len);
```

## 使用SDK

使用 FayHub 提供的 Go SDK 简化开发：

```go
import "fayhub/pkg/plugin"

sdk := plugin.NewSDK()
manifest := sdk.CreateManifest("我的插件", "1.0.0", "插件描述")
sdk.AddPermission(manifest, plugin.PermLogWrite)
```

## 测试插件

1. 在本地开发环境测试
2. 使用 FayHub 插件系统测试
3. 发布到插件市场

## 最佳实践

1. **权限最小化** - 只申请必要权限
2. **错误处理** - 完善异常情况处理
3. **资源管理** - 注意沙箱限制
4. **性能优化** - 优化WASM模块大小
5. **安全验证** - 严格验证输入数据

## 更多信息

请参考 `docs/PLUGIN_SDK_GUIDE.md` 获取详细文档。
