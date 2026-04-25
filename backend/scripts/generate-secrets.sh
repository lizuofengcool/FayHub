#!/bin/bash

# FayHub 安全密钥生成脚本
# 用于生成生产环境的安全密钥和密码

echo "🚀 开始生成FayHub生产环境安全密钥..."

# 生成JWT密钥（32位随机字符串）
JWT_SECRET=$(openssl rand -base64 32)
echo "✅ JWT密钥已生成"

# 生成数据库密码（16位随机字符串）
DB_PASSWORD=$(openssl rand -base64 12)
echo "✅ 数据库密码已生成"

# 生成Redis密码（16位随机字符串）
REDIS_PASSWORD=$(openssl rand -base64 12)
echo "✅ Redis密码已生成"

# 创建安全配置文件
cat > .env.secrets << EOF
# FayHub 安全密钥配置文件
# ⚠️ 警告：此文件包含敏感信息，请妥善保管！

# JWT密钥
JWT_SECRET=${JWT_SECRET}

# 数据库密码
DB_PASSWORD=${DB_PASSWORD}

# Redis密码
REDIS_PASSWORD=${REDIS_PASSWORD}
EOF

echo ""
echo "🔐 安全密钥生成完成！"
echo "📁 密钥文件: .env.secrets"
echo ""
echo "⚠️  重要安全提示："
echo "1. 请立即将此文件备份到安全位置"
echo "2. 不要将此文件提交到版本控制系统"
echo "3. 在生产服务器上设置适当的文件权限"
echo "4. 定期轮换这些密钥"

# 显示生成的密钥（仅用于测试环境）
if [ "$1" = "--show" ]; then
    echo ""
    echo "📋 生成的密钥（仅用于测试）："
    echo "JWT_SECRET: ${JWT_SECRET}"
    echo "DB_PASSWORD: ${DB_PASSWORD}"
    echo "REDIS_PASSWORD: ${REDIS_PASSWORD}"
fi