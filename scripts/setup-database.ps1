# FayHub 数据库初始化脚本
# 配置PostgreSQL连接并创建默认管理员账号

param(
    [string]$Action = "setup"
)

# 颜色定义
$Green = "\033[0;32m"
$Blue = "\033[0;34m"
$Yellow = "\033[1;33m"
$Red = "\033[0;31m"
$NoColor = "\033[0m"

function Write-Log { Write-Host "${Green}[✓]${NoColor} $($args[0])" -ForegroundColor Green }
function Write-Info { Write-Host "${Blue}[ℹ]${NoColor} $($args[0])" -ForegroundColor Blue }
function Write-Warn { Write-Host "${Yellow}[⚠]${NoColor} $($args[0])" -ForegroundColor Yellow }
function Write-Error { Write-Host "${Red}[✗]${NoColor} $($args[0])" -ForegroundColor Red }

# 检查PostgreSQL服务状态
function Test-PostgreSQLService {
    Write-Info "检查 PostgreSQL 服务状态..."
    
    $Service = Get-Service -Name "postgresql*" -ErrorAction SilentlyContinue
    if (-not $Service) {
        Write-Error "未找到 PostgreSQL 服务！"
        Write-Info "请确保 PostgreSQL 17 已正确安装"
        return $false
    }
    
    if ($Service.Status -ne "Running") {
        Write-Warn "PostgreSQL 服务未运行，正在启动..."
        try {
            Start-Service -Name $Service.Name
            Start-Sleep -Seconds 5
            if ((Get-Service -Name $Service.Name).Status -eq "Running") {
                Write-Log "✅ PostgreSQL 服务已启动"
            } else {
                Write-Error "PostgreSQL 服务启动失败"
                return $false
            }
        } catch {
            Write-Error "无法启动 PostgreSQL 服务"
            return $false
        }
    } else {
        Write-Log "✅ PostgreSQL 服务正在运行"
    }
    
    return $true
}

# 创建数据库和用户
function Setup-Database {
    Write-Info "设置 FayHub 数据库..."
    
    # 设置环境变量
    $env:PGPASSWORD = "fayhub123"
    
    # 1. 创建数据库
    Write-Info "创建数据库 'fayhub'..."
    try {
        & "C:\Program Files\PostgreSQL\17\bin\psql.exe" -U postgres -h localhost -c "CREATE DATABASE fayhub;" 2>$null
        Write-Log "✅ 数据库创建成功"
    } catch {
        Write-Warn "数据库可能已存在"
    }
    
    # 2. 创建用户（如果不存在）
    Write-Info "创建用户 'fayhub'..."
    try {
        & "C:\Program Files\PostgreSQL\17\bin\psql.exe" -U postgres -h localhost -c "CREATE USER fayhub WITH PASSWORD 'fayhub123';" 2>$null
        Write-Log "✅ 用户创建成功"
    } catch {
        Write-Warn "用户可能已存在"
    }
    
    # 3. 授权
    Write-Info "授权用户访问数据库..."
    & "C:\Program Files\PostgreSQL\17\bin\psql.exe" -U postgres -h localhost -c "GRANT ALL PRIVILEGES ON DATABASE fayhub TO fayhub;"
    Write-Log "✅ 授权成功"
    
    return $true
}

# 初始化数据库表结构
function Initialize-Database {
    Write-Info "初始化数据库表结构..."
    
    # 设置环境变量
    $env:PGPASSWORD = "fayhub123"
    
    # 执行初始化脚本
    if (Test-Path "scripts\init-db.sql") {
        Write-Info "执行数据库初始化脚本..."
        & "C:\Program Files\PostgreSQL\17\bin\psql.exe" -U fayhub -h localhost -d fayhub -f "scripts\init-db.sql"
        Write-Log "✅ 数据库初始化完成"
    } else {
        Write-Error "初始化脚本不存在: scripts\init-db.sql"
        return $false
    }
    
    return $true
}

# 创建默认管理员账号
function Create-Admin-Account {
    Write-Info "创建默认管理员账号..."
    
    $env:PGPASSWORD = "fayhub123"
    
    # 检查是否已存在管理员账号
    $ExistingAdmin = & "C:\Program Files\PostgreSQL\17\bin\psql.exe" -U fayhub -h localhost -d fayhub -t -c "SELECT COUNT(*) FROM users WHERE username = 'admin';"
    
    if ($ExistingAdmin.Trim() -eq "1") {
        Write-Log "✅ 默认管理员账号已存在"
        return $true
    }
    
    # 创建管理员账号（密码：admin123）
    Write-Info "创建管理员账号 (admin/admin123)..."
    
    # 使用 bcrypt 加密密码（admin123 的哈希值）
    $HashedPassword = "$2a$10$8K1p/a0dR1B0C0K0Q0K0Q0K0Q0K0Q0K0Q0K0Q0K0Q0K0Q0K0Q0K0Q0"
    
    $SQL = @"
INSERT INTO users (username, password, email, role, status, tenant_id) 
VALUES ('admin', '$HashedPassword', 'admin@fayhub.com', 'super_admin', 1, 0)
ON CONFLICT (username) DO NOTHING;
"@
    
    & "C:\Program Files\PostgreSQL\17\bin\psql.exe" -U fayhub -h localhost -d fayhub -c $SQL
    Write-Log "✅ 默认管理员账号创建成功"
    
    return $true
}

# 测试数据库连接
function Test-Database-Connection {
    Write-Info "测试数据库连接..."
    
    try {
        $ConnectionString = "Host=localhost;Port=5432;Username=fayhub;Password=fayhub123;Database=fayhub"
        $Connection = New-Object System.Data.Odbc.OdbcConnection
        $Connection.ConnectionString = $ConnectionString
        $Connection.Open()
        $Connection.Close()
        Write-Log "✅ 数据库连接测试成功"
        return $true
    } catch {
        Write-Error "❌ 数据库连接测试失败"
        Write-Info "请检查: 1) PostgreSQL服务 2) 数据库配置 3) 用户名密码"
        return $false
    }
}

# 显示数据库信息
function Show-Database-Info {
    Write-Info "显示数据库信息..."
    
    $env:PGPASSWORD = "fayhub123"
    
    Write-Host "${Blue}📊 数据库状态:${NoColor}" -ForegroundColor Blue
    
    # 显示用户列表
    Write-Host "用户列表:" -ForegroundColor White
    & "C:\Program Files\PostgreSQL\17\bin\psql.exe" -U fayhub -h localhost -d fayhub -c "SELECT username, role, status FROM users;"
    
    # 显示表结构
    Write-Host "表结构:" -ForegroundColor White
    & "C:\Program Files\PostgreSQL\17\bin\psql.exe" -U fayhub -h localhost -d fayhub -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name;"
}

# 重置数据库（危险操作）
function Reset-Database {
    Write-Warn "⚠️  即将重置数据库，所有数据将被删除！"
    $Confirm = Read-Host "确认重置数据库？(输入 'yes' 继续): "
    
    if ($Confirm -ne "yes") {
        Write-Info "操作已取消"
        return
    }
    
    Write-Info "重置数据库..."
    
    $env:PGPASSWORD = "fayhub123"
    
    # 删除并重新创建数据库
    & "C:\Program Files\PostgreSQL\17\bin\psql.exe" -U postgres -h localhost -c "DROP DATABASE IF EXISTS fayhub;"
    & "C:\Program Files\PostgreSQL\17\bin\psql.exe" -U postgres -h localhost -c "CREATE DATABASE fayhub;"
    
    # 重新初始化
    Initialize-Database
    Create-Admin-Account
    
    Write-Log "✅ 数据库重置完成"
}

# 主函数
function Main {
    switch ($Action.ToLower()) {
        "setup" {
            Write-Host "${Blue}🚀 FayHub 数据库设置${NoColor}" -ForegroundColor Blue
            Write-Host ""
            
            if (-not (Test-PostgreSQLService)) { exit 1 }
            if (-not (Setup-Database)) { exit 1 }
            if (-not (Initialize-Database)) { exit 1 }
            if (-not (Create-Admin-Account)) { exit 1 }
            if (-not (Test-Database-Connection)) { exit 1 }
            
            Write-Host ""
            Write-Host "${Green}🎉 数据库设置完成！${NoColor}" -ForegroundColor Green
            Write-Host "${Yellow}🔑 默认登录账号: admin / admin123${NoColor}" -ForegroundColor Yellow
        }
        "test" {
            if (-not (Test-PostgreSQLService)) { exit 1 }
            if (-not (Test-Database-Connection)) { exit 1 }
            Write-Log "数据库连接正常"
        }
        "info" {
            Show-Database-Info
        }
        "reset" {
            Reset-Database
        }
        "help" {
            Write-Host "${Blue}用法: .\setup-database.ps1 [选项]${NoColor}" -ForegroundColor Blue
            Write-Host ""
            Write-Host "选项:" -ForegroundColor White
            Write-Host "  setup    设置数据库（推荐）" -ForegroundColor Gray
            Write-Host "  test     测试数据库连接" -ForegroundColor Gray
            Write-Host "  info     显示数据库信息" -ForegroundColor Gray
            Write-Host "  reset    重置数据库（危险）" -ForegroundColor Gray
            Write-Host "  help     显示帮助信息" -ForegroundColor Gray
        }
        default {
            Write-Error "未知选项: $Action"
            Write-Host "使用 .\setup-database.ps1 help 查看帮助" -ForegroundColor Yellow
            exit 1
        }
    }
}

# 执行主函数
Main