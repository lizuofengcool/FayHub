@echo off
echo 🚀 开始执行数据库迁移脚本...
echo.

REM 检查配置文件是否存在
if not exist "..\config.yaml" (
    echo ❌ 配置文件 config.yaml 不存在
    echo 请从 config.example.yaml 复制并修改配置
    pause
    exit /b 1
)

REM 构建迁移工具
echo 📦 构建数据库迁移工具...
go build -o migrate.exe migrate.go
if %errorlevel% neq 0 (
    echo ❌ 构建失败
    pause
    exit /b 1
)

REM 执行迁移
echo 📊 执行数据库迁移...
.\migrate.exe
if %errorlevel% neq 0 (
    echo ❌ 迁移执行失败
    pause
    exit /b 1
)

echo.
echo 🎉 数据库迁移完成！
echo.
echo 📋 默认账户信息：
echo    用户名：admin
echo    密码：admin123456
echo    邮箱：admin@fayhub.com
echo.
pause