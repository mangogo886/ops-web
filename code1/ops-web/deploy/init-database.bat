@echo off
REM ============================================
REM Windows 数据库初始化脚本
REM 说明: 此脚本用于初始化数据库表结构和初始管理员账户
REM ============================================

setlocal enabledelayedexpansion

echo ===========================================
echo 档案审核管理系统 - 数据库初始化
echo ===========================================
echo.

REM 设置默认值
set DB_NAME=ops
set DB_USER=root
set DB_PASS=

REM 获取参数
if not "%1"=="" set DB_NAME=%1
if not "%2"=="" set DB_USER=%2
if not "%3"=="" set DB_PASS=%3

REM 如果没有提供密码，提示输入
if "%DB_PASS%"=="" (
    set /p DB_PASS="请输入MySQL密码: "
)

echo 数据库名: %DB_NAME%
echo MySQL用户: %DB_USER%
echo.

REM 检查mysql命令是否存在
where mysql >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未找到mysql命令，请确保MySQL已安装并添加到PATH环境变量
    pause
    exit /b 1
)

REM 创建数据库（如果不存在）
echo [1/3] 创建数据库...
mysql -u %DB_USER% -p%DB_PASS% -e "CREATE DATABASE IF NOT EXISTS `%DB_NAME%` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>nul
if %errorlevel% neq 0 (
    echo 错误: 无法创建数据库，请检查数据库连接信息
    pause
    exit /b 1
)
echo ✓ 数据库创建成功
echo.

REM 执行建表SQL
echo [2/3] 创建数据库表...
mysql -u %DB_USER% -p%DB_PASS% %DB_NAME% < "%~dp0create-all-tables.sql"
if %errorlevel% neq 0 (
    echo 错误: 创建表失败，请检查SQL文件是否存在
    pause
    exit /b 1
)
echo ✓ 数据库表创建成功
echo.

REM 初始化管理员账户
echo [3/3] 初始化管理员账户...
mysql -u %DB_USER% -p%DB_PASS% %DB_NAME% < "%~dp0init-admin-user.sql"
if %errorlevel% neq 0 (
    echo 警告: 初始化管理员账户失败，可能账户已存在
)
echo ✓ 管理员账户初始化完成
echo.

echo ===========================================
echo 数据库初始化完成
echo ===========================================
echo.
echo 默认管理员账号：
echo   用户名: admin
echo   密码: admin123
echo.
echo ⚠️  首次登录后请立即修改密码！
echo.
pause


