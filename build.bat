@echo off
setlocal enabledelapsedexpansion
echo 正在构建前端...
cd frontend
call npm install
call npm run build
cd ..
rmdir /s /q backend\web
mkdir backend\web
xcopy /s /e /y frontend\dist\* backend\web\

:: 定义目标文件路径模板
set "htmlFile=backend\web\index.html"
set "jsPattern=backend\web\assets\index-*.js"

:: 方案一：调用PowerShell实现批量替换（推荐）
powershell -Command "$files = @('%htmlFile%') + (gci '%jsPattern%').FullName; foreach($f in $files){(gc $f) -replace 'assets/','fileserver/assets/' | sc $f}"

echo 文件内容已成功替换！


echo 正在构建后端...
cd backend
go build -o ../fileShare.exe
cd ..

echo 构建完成！
echo 可以直接运行 fileShare.exe 启动应用
fileShare.exe

