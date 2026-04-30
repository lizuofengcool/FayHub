-- ============================================================
-- FayHub 数据库初始化脚本
-- 用途: 创建 PostgreSQL 数据库和用户
-- 使用: psql -U postgres -f init-db.sql
-- 注意: 请设置环境变量 FAYHUB_DB_PASSWORD
-- ============================================================

-- 检查环境变量
\set db_password `echo $FAYHUB_DB_PASSWORD`

-- 创建 fayhub 用户（如果不存在）
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'fayhub') THEN
    EXECUTE format('CREATE ROLE fayhub WITH LOGIN PASSWORD %L', :'db_password');
    RAISE NOTICE '已创建用户 fayhub';
  ELSE
    EXECUTE format('ALTER USER fayhub WITH PASSWORD %L', :'db_password');
    RAISE NOTICE '用户 fayhub 已存在，已更新密码';
  END IF;
END
$$;

-- 创建 fayhub 数据库（如果不存在）
SELECT 'CREATE DATABASE fayhub OWNER fayhub ENCODING ''UTF8'''
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'fayhub')\gexec

-- 授权
GRANT ALL PRIVILEGES ON DATABASE fayhub TO fayhub;

-- 输出信息
\echo '========================================='
\echo '  FayHub 数据库初始化完成！'
\echo '  数据库名: fayhub'
\echo '  用户名:   fayhub'
\echo '  主机:     localhost:5432'
\echo '========================================='
\echo ''
\echo '启动后端请设置环境变量:'
\echo '  set FAYHUB_DB_PASSWORD=your_secure_password'
\echo '  set FAYHUB_JWT_SECRET=your_secure_jwt_secret'
\echo '然后运行: start-dev.bat'
