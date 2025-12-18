#!/bin/bash

# ============================================
# Linux 数据库初始化脚本
# 说明: 此脚本用于初始化数据库表结构和初始管理员账户
# ============================================

set -e

DB_NAME=${1:-ops}
DB_USER=${2:-root}
DB_PASS=${3}

if [ -z "$DB_PASS" ]; then
    read -sp "请输入MySQL密码: " DB_PASS
    echo
fi

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "=========================================="
echo "档案审核管理系统 - 数据库初始化"
echo "=========================================="
echo ""
echo "数据库名: $DB_NAME"
echo "MySQL用户: $DB_USER"
echo ""

# 创建数据库（如果不存在）
echo "[1/3] 创建数据库..."
mysql -u "$DB_USER" -p"$DB_PASS" -e "CREATE DATABASE IF NOT EXISTS \`$DB_NAME\` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>/dev/null || {
    echo "错误: 无法创建数据库，请检查数据库连接信息"
    exit 1
}
echo "✓ 数据库创建成功"
echo ""

# 执行建表SQL
echo "[2/3] 创建数据库表..."
mysql -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" < "$SCRIPT_DIR/create-all-tables.sql"
if [ $? -eq 0 ]; then
    echo "✓ 数据库表创建成功"
else
    echo "✗ 数据库表创建失败"
    exit 1
fi
echo ""

# 初始化管理员账户
echo "[3/3] 初始化管理员账户..."
mysql -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" < "$SCRIPT_DIR/init-admin-user.sql"
if [ $? -eq 0 ]; then
    echo "✓ 管理员账户初始化完成"
else
    echo "⚠️  警告: 初始化管理员账户失败，可能账户已存在"
fi
echo ""

echo "=========================================="
echo "数据库初始化完成"
echo "=========================================="
echo ""
echo "默认管理员账号："
echo "  用户名: admin"
echo "  密码: admin123"
echo ""
echo "⚠️  首次登录后请立即修改密码！"
echo ""

