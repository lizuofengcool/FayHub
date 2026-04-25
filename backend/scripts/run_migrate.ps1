# FayHub 数据库迁移脚本
# PowerShell 版本

Write-Host "🚀 开始执行数据库迁移脚本..." -ForegroundColor Green
Write-Host ""

# 检查配置文件是否存在
if (-not (Test-Path "../config.yaml")) {
    Write-Host "❌ 配置文件 config.yaml 不存在" -ForegroundColor Red
    Write-Host "请从 config.example.yaml 复制并修改配置" -ForegroundColor Yellow
    Read-Host "按任意键退出"
    exit 1
}

# 构建迁移工具
Write-Host "📦 构建数据库迁移工具..." -ForegroundColor Cyan
go build -o migrate.exe migrate.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 构建失败" -ForegroundColor Red
    Read-Host "按任意键退出"
    exit 1
}

# 执行迁移
Write-Host "📊 执行数据库迁移..." -ForegroundColor Cyan
.\migrate.exe
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 迁移执行失败" -ForegroundColor Red
    Read-Host "按任意键退出"
    exit 1
}

Write-Host ""
Write-Host "🎉 数据库迁移完成！" -ForegroundColor Green
Write-Host ""
Write-Host "📋 默认账户信息：" -ForegroundColor Yellow
Write-Host "    用户名：admin" -ForegroundColor White
Write-Host "    密码：admin123456" -ForegroundColor White
Write-Host "    邮箱：admin@fayhub.com" -ForegroundColor White
Write-Host ""
Read-Host "按任意键退出"