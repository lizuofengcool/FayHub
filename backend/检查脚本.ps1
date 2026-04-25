# FayHub 项目 AI 开发预检脚本
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  FayHub AI 开发 5 分钟快速预检" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 1. 编译检查
Write-Host "[1/6] 编译检查..." -ForegroundColor Yellow
go build ./...
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 编译失败！请先修复编译错误" -ForegroundColor Red
    exit 1
}
Write-Host "✅ 编译通过" -ForegroundColor Green
Write-Host ""

# 2. 检查全局DB直连
Write-Host "[2/6] 检查多租户隔离（全局DB直连）..." -ForegroundColor Yellow
$globalDbCount = Get-ChildItem -Path . -Recurse -Filter *.go | Select-String "GetGlobalDB" | Measure-Object | Select-Object -ExpandProperty Count
if ($globalDbCount -gt 0) {
    Write-Host "❌ 发现 $globalDbCount 处使用了 GetGlobalDB，必须改为 GetDB(ctx)" -ForegroundColor Red
    Get-ChildItem -Path . -Recurse -Filter *.go | Select-String "GetGlobalDB"
    exit 1
}
Write-Host "✅ 未发现全局DB直连" -ForegroundColor Green
Write-Host ""

# 3. 检查响应工具统一
Write-Host "[3/6] 检查响应工具统一..." -ForegroundColor Yellow
$oldResponseCount = Get-ChildItem -Path internal/controller -Filter *.go | Select-String "utils\.(Ok|Fail)" | Measure-Object | Select-Object -ExpandProperty Count
if ($oldResponseCount -gt 0) {
    Write-Host "❌ 发现 $oldResponseCount 处使用了旧响应工具，必须改为 response.Gin*" -ForegroundColor Red
    Get-ChildItem -Path internal/controller -Filter *.go | Select-String "utils\.(Ok|Fail)"
    exit 1
}
Write-Host "✅ 响应工具统一" -ForegroundColor Green
Write-Host ""

# 4. 检查硬编码配置
Write-Host "[4/6] 检查硬编码配置..." -ForegroundColor Yellow
$hardCodeCount = Get-ChildItem -Path . -Recurse -Filter *.go | Select-String "127.0.0.1:3306|root:.*@tcp|JWT_SECRET" | Measure-Object | Select-Object -ExpandProperty Count
if ($hardCodeCount -gt 0) {
    Write-Host "⚠️  发现 $hardCodeCount 处疑似硬编码配置，请确认是否从配置文件读取" -ForegroundColor Yellow
    Get-ChildItem -Path . -Recurse -Filter *.go | Select-String "127.0.0.1:3306|root:.*@tcp|JWT_SECRET"
} else {
    Write-Host "✅ 未发现硬编码敏感配置" -ForegroundColor Green
}
Write-Host ""

# 5. Git状态检查
Write-Host "[5/6] Git 修改范围检查..." -ForegroundColor Yellow
git status
Write-Host ""

# 6. Git diff预览
Write-Host "[6/6] Git 代码修改预览..." -ForegroundColor Yellow
git diff --stat
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  ✅ 快速预检通过！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  下一步：请继续深入检查功能逻辑与安全项" -ForegroundColor Gray
Write-Host "========================================" -ForegroundColor Cyan
