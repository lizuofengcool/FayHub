-- FayHub 数据库初始化脚本
-- 创建必要的表结构和初始数据

-- 创建租户表
CREATE TABLE IF NOT EXISTS tenants (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    domain VARCHAR(200) UNIQUE,
    description VARCHAR(500),
    status INTEGER DEFAULT 1,
    expired_at BIGINT DEFAULT 0,
    max_users INTEGER DEFAULT 10,
    contact_name VARCHAR(100),
    contact_phone VARCHAR(20),
    contact_email VARCHAR(200),
    logo VARCHAR(500),
    address VARCHAR(500),
    industry VARCHAR(100),
    scale VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(200) NOT NULL,
    email VARCHAR(200) UNIQUE,
    phone VARCHAR(20) UNIQUE,
    status INTEGER DEFAULT 1,
    role VARCHAR(50) DEFAULT 'user',
    tenant_id INTEGER DEFAULT 0,
    last_login_at BIGINT DEFAULT 0,
    login_ip VARCHAR(50),
    avatar VARCHAR(500),
    real_name VARCHAR(100),
    department VARCHAR(100),
    position VARCHAR(100),
    gender INTEGER DEFAULT 0,
    birthday BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- 创建默认超级管理员用户
INSERT INTO users (username, password, email, role, status) VALUES 
('admin', '$2a$10$8K1p/a0dR1B0C0K0Q0K0Q0K0Q0K0Q0K0Q0K0Q0K0Q0K0Q0K0Q0K0Q0', 'admin@fayhub.com', 'super_admin', 1)
ON CONFLICT (username) DO NOTHING;

-- 创建默认平台租户
INSERT INTO tenants (name, code, domain, description, status, max_users) VALUES 
('FayHub平台', 'fayhub', 'platform.fayhub.com', 'FayHub多租户SaaS平台', 1, 1000)
ON CONFLICT (code) DO NOTHING;

-- 创建必要的索引
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- 创建其他必要的表结构（根据实际业务需求添加）
-- 这里可以添加更多的表创建语句

-- 输出初始化完成信息
SELECT 'FayHub数据库初始化完成' AS message;