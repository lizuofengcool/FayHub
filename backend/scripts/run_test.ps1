# FayHub 数据库连接测试脚本
# PowerShell 版本

Write-Host "🧪 开始数据库连接测试..." -ForegroundColor Green
Write-Host ""

# 检查配置文件是否存在
if (-not (Test-Path "../config.yaml")) {
    Write-Host "❌ 配置文件 config.yaml 不存在" -ForegroundColor Red
    Write-Host "请从 config.example.yaml 复制并修改配置" -ForegroundColor Yellow
    Read-Host "按任意键退出"
    exit 1
}

# 构建测试工具
Write-Host "📦 构建数据库测试工具..." -ForegroundColor Cyan
go build -o test_db.exe test_db.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 构建失败" -ForegroundColor Red
    Read-Host "按任意键退出"
    exit 1
}

# 执行测试
Write-Host "🔗 执行数据库连接测试..." -ForegroundColor Cyan
.\test_db.exe
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 测试执行失败" -ForegroundColor Red
    Read-Host "按任意键退出"
    exit 1
}

Write-Host ""
Write-Host "🎉 数据库连接测试通过！" -ForegroundColor Green
Write-Host ""
Write-Host "📋 测试结果：" -ForegroundColor Yellow
Write-Host "    - 数据库连接：✅ 正常" -ForegroundColor White
Write-Host "    - 基本操作：✅ 正常" -ForegroundColor White
Write-Host "    - 多租户隔离：✅ 正常" -ForegroundColor White
Write-Host "    - 性能测试：✅ 正常" -ForegroundColor White
Write-Host ""
Read-Host "按任意键退出"