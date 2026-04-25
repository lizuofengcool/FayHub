# FayHub 安全密钥生成脚本 (PowerShell版本)
# 用于生成生产环境的安全密钥和密码

Write-Host "🚀 开始生成FayHub生产环境安全密钥..." -ForegroundColor Green

# 生成JWT密钥（32位随机字符串）
$JWT_SECRET = [Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Maximum 256 }))
Write-Host "✅ JWT密钥已生成" -ForegroundColor Green

# 生成数据库密码（16位随机字符串）
$DB_PASSWORD = [Convert]::ToBase64String((1..12 | ForEach-Object { Get-Random -Maximum 256 }))
Write-Host "✅ 数据库密码已生成" -ForegroundColor Green

# 生成Redis密码（16位随机字符串）
$REDIS_PASSWORD = [Convert]::ToBase64String((1..12 | ForEach-Object { Get-Random -Maximum 256 }))
Write-Host "✅ Redis密码已生成" -ForegroundColor Green

# 创建安全配置文件
$SECRETS_CONTENT = @"
# FayHub 安全密钥配置文件
# ⚠️ 警告：此文件包含敏感信息，请妥善保管！

# JWT密钥
JWT_SECRET=$JWT_SECRET

# 数据库密码
DB_PASSWORD=$DB_PASSWORD

# Redis密码
REDIS_PASSWORD=$REDIS_PASSWORD
"@

$SECRETS_CONTENT | Out-File -FilePath ".env.secrets" -Encoding UTF8

Write-Host ""
Write-Host "🔐 安全密钥生成完成！" -ForegroundColor Green
Write-Host "📁 密钥文件: .env.secrets" -ForegroundColor Yellow
Write-Host ""
Write-Host "⚠️  重要安全提示：" -ForegroundColor Red
Write-Host "1. 请立即将此文件备份到安全位置" -ForegroundColor Yellow
Write-Host "2. 不要将此文件提交到版本控制系统" -ForegroundColor Yellow
Write-Host "3. 在生产服务器上设置适当的文件权限" -ForegroundColor Yellow
Write-Host "4. 定期轮换这些密钥" -ForegroundColor Yellow

# 显示生成的密钥（仅用于测试环境）
if ($args[0] -eq "--show") {
    Write-Host ""
    Write-Host "📋 生成的密钥（仅用于测试）：" -ForegroundColor Cyan
    Write-Host "JWT_SECRET: $JWT_SECRET" -ForegroundColor White
    Write-Host "DB_PASSWORD: $DB_PASSWORD" -ForegroundColor White
    Write-Host "REDIS_PASSWORD: $REDIS_PASSWORD" -ForegroundColor White
}