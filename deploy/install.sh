#!/bin/bash

# Ops-Web 安装脚本
# 用于在生产服务器上快速部署

set -e

echo "=========================================="
echo "Ops-Web 安装脚本"
echo "=========================================="
echo ""

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "错误: 未找到 Go 编译器"
    echo "请先安装 Go: https://golang.org/dl/"
    exit 1
fi

# 检查MySQL客户端是否安装
if ! command -v mysql &> /dev/null; then
    echo "警告: 未找到 mysql 客户端"
    echo "将跳过数据库初始化步骤"
    SKIP_DB=1
fi

# 获取项目目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "项目目录: $PROJECT_DIR"
echo ""

# 1. 编译程序
echo "步骤 1/5: 编译程序..."
cd "$PROJECT_DIR"
go build -o ops-web .
if [ $? -eq 0 ]; then
    echo "✓ 编译成功"
else
    echo "✗ 编译失败"
    exit 1
fi
echo ""

# 2. 检查配置文件
echo "步骤 2/5: 检查配置文件..."
if [ ! -f "config/config.json" ]; then
    echo "创建配置文件..."
    mkdir -p config
    cp deploy/config/config.json config/config.json
    echo "✓ 配置文件已创建，请编辑 config/config.json 设置数据库连接信息"
else
    echo "✓ 配置文件已存在"
fi
echo ""

# 3. 创建logs目录
echo "步骤 3/5: 创建日志目录..."
mkdir -p logs
echo "✓ 日志目录已创建"
echo ""

# 4. 数据库初始化提示
echo "步骤 4/5: 数据库初始化..."
if [ -z "$SKIP_DB" ]; then
    read -p "是否现在初始化数据库？(y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        read -p "请输入数据库名称 [ops]: " DB_NAME
        DB_NAME=${DB_NAME:-ops}
        
        read -p "请输入MySQL root密码: " -s DB_PASS
        echo
        
        echo "执行SQL脚本..."
        mysql -u root -p"$DB_PASS" "$DB_NAME" < deploy/create-user-tables.sql
        if [ $? -eq 0 ]; then
            echo "✓ 用户表创建成功"
        else
            echo "✗ 用户表创建失败"
        fi
        
        read -p "是否创建审核相关表？(y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            mysql -u root -p"$DB_PASS" "$DB_NAME" < deploy/create-audit-tables.sql
            if [ $? -eq 0 ]; then
                echo "✓ 审核表创建成功"
            else
                echo "✗ 审核表创建失败"
            fi
        fi
    else
        echo "跳过数据库初始化"
        echo "请手动执行以下SQL脚本："
        echo "  - deploy/create-user-tables.sql"
        echo "  - deploy/create-audit-tables.sql (如需要)"
    fi
else
    echo "跳过数据库初始化（未找到mysql客户端）"
    echo "请手动执行SQL脚本"
fi
echo ""

# 5. 完成
echo "步骤 5/5: 安装完成！"
echo ""
echo "=========================================="
echo "安装完成"
echo "=========================================="
echo ""
echo "下一步："
echo "1. 编辑 config/config.json 设置数据库连接信息"
echo "2. 执行数据库初始化SQL脚本（如未执行）"
echo "3. 运行 ./ops-web 启动服务"
echo "4. 访问 http://127.0.0.1:8080/login"
echo ""
echo "默认管理员账号："
echo "  用户名: admin"
echo "  密码: admin123"
echo ""
echo "⚠️  首次登录后请立即修改密码！"
echo ""






















