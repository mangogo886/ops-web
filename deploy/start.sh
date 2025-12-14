#!/bin/bash

# Ops-Web 启动脚本
# 使用方法: ./start.sh

# 获取脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
APP_DIR="$(dirname "$SCRIPT_DIR")"

# 进入应用目录
cd "$APP_DIR" || exit 1

# 检查可执行文件是否存在
if [ ! -f "ops-web" ]; then
    echo "错误: 找不到 ops-web 可执行文件"
    echo "请先编译程序: go build -o ops-web ."
    exit 1
fi

# 检查配置文件是否存在
if [ ! -f "config/config.json" ]; then
    echo "错误: 找不到配置文件 config/config.json"
    echo "请从 deploy/config/config.json 复制并修改"
    exit 1
fi

# 创建logs目录（如果不存在）
mkdir -p logs

# 启动服务
echo "启动 Ops-Web 服务..."
echo "应用目录: $APP_DIR"
echo "日志目录: $APP_DIR/logs"
echo ""

./ops-web





