@echo off
echo 🧪 开始数据库连接测试...
echo.

REM 检查配置文件是否存在
if not exist "..\config.yaml" (
    echo ❌ 配置文件 config.yaml 不存在
    echo 请从 config.example.yaml 复制并修改配置
    pause
    exit /b 1
)

REM 构建测试工具
echo 📦 构建数据库测试工具...
go build -o test_db.exe test_db.go
if %errorlevel% neq 0 (
    echo ❌ 构建失败
    pause
    exit /b 1
)

REM 执行测试
echo 🔗 执行数据库连接测试...
.\test_db.exe
if %errorlevel% neq 0 (
    echo ❌ 测试执行失败
    pause
    exit /b 1
)

echo.
echo 🎉 数据库连接测试通过！
echo.
echo 📋 测试结果：
echo    - 数据库连接：✅ 正常
echo    - 基本操作：✅ 正常
echo    - 多租户隔离：✅ 正常
echo    - 性能测试：✅ 正常
echo.
pause