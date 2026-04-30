# ============================================================
# FayHub 数据库一键初始化脚本（Windows PowerShell）
# 用途: 创建 PostgreSQL 数据库、用户
# 使用: .\init-db.ps1
# ============================================================

$ErrorActionPreference = "Stop"

$PSQL_PATH = "C:\Program Files\PostgreSQL\17\bin\psql.exe"
$PG_HOST = "localhost"
$PG_PORT = "5432"
$DB_NAME = "fayhub"
$DB_USER = "fayhub"
$DB_PASS = "fayhub123"
$JWT_SECRET = "fayhub_jwt_secret_dev_2026"

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "  FayHub 数据库初始化向导" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

# 检查 psql 是否存在
if (-not (Test-Path $PSQL_PATH)) {
    $pgDirs = Get-ChildItem "C:\Program Files\PostgreSQL" -Directory -ErrorAction SilentlyContinue
    if ($pgDirs) {
        $latest = $pgDirs | Sort-Object Name -Descending | Select-Object -First 1
        $PSQL_PATH = Join-Path $latest.FullName "bin\psql.exe"
    }
}

if (-not (Test-Path $PSQL_PATH)) {
    Write-Host "[错误] 未找到 PostgreSQL 的 psql.exe" -ForegroundColor Red
    Write-Host "请确认 PostgreSQL 已安装" -ForegroundColor Yellow
    exit 1
}

Write-Host "[OK] 找到 psql: $PSQL_PATH" -ForegroundColor Green

# postgres 超级用户密码
$PG_SUPER_PASS = "123456"

# 测试连接
Write-Host ""
Write-Host "[1/3] 测试 PostgreSQL 连接..." -ForegroundColor Cyan
$env:PGPASSWORD = $PG_SUPER_PASS
$result = & $PSQL_PATH -U postgres -h $PG_HOST -p $PG_PORT -c "SELECT version();" 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "[错误] 无法连接 PostgreSQL: $result" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] PostgreSQL 连接成功" -ForegroundColor Green

# 创建用户和数据库
Write-Host ""
Write-Host "[2/3] 创建/更新数据库用户和数据库..." -ForegroundColor Cyan

$createUserSQL = @"
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '$DB_USER') THEN
    CREATE ROLE $DB_USER WITH LOGIN PASSWORD '$DB_PASS';
  ELSE
    ALTER USER $DB_USER WITH PASSWORD '$DB_PASS';
  END IF;
END
\$\$;
"@

& $PSQL_PATH -U postgres -h $PG_HOST -p $PG_PORT -c $createUserSQL 2>$null

$createDbSQL = "SELECT 'CREATE DATABASE $DB_NAME OWNER $DB_USER ENCODING ''UTF8''' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$DB_NAME')\gexec"
& $PSQL_PATH -U postgres -h $PG_HOST -p $PG_PORT -c $createDbSQL 2>$null

$grantSQL = "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"
& $PSQL_PATH -U postgres -h $PG_HOST -p $PG_PORT -c $grantSQL 2>$null

$grantSchemaSQL = "GRANT ALL ON SCHEMA public TO $DB_USER;"
& $PSQL_PATH -U postgres -h $PG_HOST -p $PG_PORT -d $DB_NAME -c $grantSchemaSQL 2>$null

Write-Host "[OK] 数据库和用户就绪" -ForegroundColor Green

# 验证连接
Write-Host ""
Write-Host "[3/3] 验证 fayhub 用户连接..." -ForegroundColor Cyan
$env:PGPASSWORD = $DB_PASS
$result = & $PSQL_PATH -U $DB_USER -h $PG_HOST -p $PG_PORT -d $DB_NAME -c "SELECT current_database(), current_user;" 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "[错误] fayhub 用户连接失败: $result" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] fayhub 用户连接成功" -ForegroundColor Green

$env:PGPASSWORD = ""

# 设置当前会话环境变量
$env:FAYHUB_DB_PASSWORD = $DB_PASS
$env:FAYHUB_JWT_SECRET = $JWT_SECRET

Write-Host ""
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "  初始化完成！" -ForegroundColor Green
Write-Host "  数据库: $DB_NAME" -ForegroundColor White
Write-Host "  用户:   $DB_USER" -ForegroundColor White
Write-Host "  密码:   $DB_PASS" -ForegroundColor White
Write-Host ""
Write-Host "  启动后端: 运行 start-dev.bat" -ForegroundColor Yellow
Write-Host "=========================================" -ForegroundColor Cyan
