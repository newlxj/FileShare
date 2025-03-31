@echo off
echo ===== 开始构建文件分享系统 =====

REM 设置工作目录
set BACKEND_DIR=%~dp0
set FRONTEND_DIR=%BACKEND_DIR%..\frontend

echo 当前工作目录: %BACKEND_DIR%
echo 前端目录: %FRONTEND_DIR%

REM 检查前端目录是否存在
if not exist "%FRONTEND_DIR%" (
    echo 错误: 前端目录不存在
    exit /b 1
)

REM 构建前端
echo ===== 构建前端 =====
cd "%FRONTEND_DIR%"

REM 检查npm是否安装
npm -v >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo 错误: 未安装npm，请先安装Node.js
    exit /b 1
)

REM 安装依赖并构建
echo 安装前端依赖...
call npm install
if %ERRORLEVEL% neq 0 (
    echo 错误: 安装前端依赖失败
    exit /b 1
)

echo 构建前端...
call npm run build
if %ERRORLEVEL% neq 0 (
    echo 错误: 构建前端失败
    exit /b 1
)

REM 检查dist目录是否生成
if not exist "%FRONTEND_DIR%\dist" (
    echo 错误: 前端构建失败，dist目录不存在
    exit /b 1
)

REM 构建后端
echo ===== 构建后端 =====
cd "%BACKEND_DIR%"

REM 检查go是否安装
go version >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo 错误: 未安装Go，请先安装Go
    exit /b 1
)

REM 构建可执行文件
echo 构建后端可执行文件...
go build -o fileShare.exe
if %ERRORLEVEL% neq 0 (
    echo 错误: 构建后端失败
    exit /b 1
)

echo ===== 构建完成 =====
echo 可执行文件: %BACKEND_DIR%fileShare.exe
echo.
echo 运行方式: 直接双击fileShare.exe或在命令行中运行
echo.

pause