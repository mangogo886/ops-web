#!/bin/bash

# 数据库初始化脚本
# 使用方法: ./init-database.sh [数据库名] [MySQL用户] [MySQL密码]

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
echo "初始化数据库: $DB_NAME"
echo "=========================================="
echo ""

# 创建数据库（如果不存在）
echo "1. 创建数据库..."
mysql -u "$DB_USER" -p"$DB_PASS" -e "CREATE DATABASE IF NOT EXISTS \`$DB_NAME\` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>/dev/null || {
    echo "错误: 无法创建数据库，请检查数据库连接信息"
    exit 1
}
echo "✓ 数据库创建成功"
echo ""

# 执行SQL脚本
echo "2. 创建用户和角色表..."
mysql -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" < "$SCRIPT_DIR/create-user-tables.sql"
if [ $? -eq 0 ]; then
    echo "✓ 用户表创建成功"
else
    echo "✗ 用户表创建失败"
    exit 1
fi
echo ""

read -p "是否创建审核相关表？(y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "3. 创建审核相关表..."
    mysql -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" < "$SCRIPT_DIR/create-audit-tables.sql"
    if [ $? -eq 0 ]; then
        echo "✓ 审核表创建成功"
    else
        echo "✗ 审核表创建失败"
        exit 1
    fi
    echo ""
fi

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











