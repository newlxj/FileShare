@echo off
setlocal enabledelapsedexpansion
echo ���ڹ���ǰ��...
cd frontend
call npm install
call npm run build
cd ..
rmdir /s /q backend\web
mkdir backend\web
xcopy /s /e /y frontend\dist\* backend\web\

:: ����Ŀ���ļ�·��ģ��
set "htmlFile=backend\web\index.html"
set "jsPattern=backend\web\assets\index-*.js"

:: ����һ������PowerShellʵ�������滻���Ƽ���
powershell -Command "$files = @('%htmlFile%') + (gci '%jsPattern%').FullName; foreach($f in $files){(gc $f) -replace 'assets/','fileserver/assets/' | sc $f}"

echo �ļ������ѳɹ��滻��


echo ���ڹ������...
cd backend
go build -o ../fileShare.exe
cd ..

echo ������ɣ�
echo ����ֱ������ fileShare.exe ����Ӧ��
fileShare.exe

