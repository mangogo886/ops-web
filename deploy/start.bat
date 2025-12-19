@echo off
REM Ops-Web 启动脚本 (Windows)
REM 使用方法: start.bat

REM 获取脚本所在目录
set SCRIPT_DIR=%~dp0
set APP_DIR=%SCRIPT_DIR%\..

REM 进入应用目录
cd /d "%APP_DIR%"

REM 检查可执行文件是否存在
if not exist "ops-web.exe" (
    echo 错误: 找不到 ops-web.exe 可执行文件
    echo 请先编译程序: go build -o ops-web.exe .
    pause
    exit /b 1
)

REM 检查配置文件是否存在
if not exist "config\config.json" (
    echo 错误: 找不到配置文件 config\config.json
    echo 请从 deploy\config\config.json 复制并修改
    pause
    exit /b 1
)

REM 创建logs目录（如果不存在）
if not exist "logs" mkdir logs

REM 启动服务
echo 启动 Ops-Web 服务...
echo 应用目录: %APP_DIR%
echo 日志目录: %APP_DIR%\logs
echo.

ops-web.exe

pause


















