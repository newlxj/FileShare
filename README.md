# 文件分享系统  -一键运行的前后端文件共享服务器
# ⭐️支持这个项目
如果你觉得这个工具不错，请点击右上方给我一个'star'⭐️吧！
## 概述
   本系统实现文件共享功能，主要用于本地需要对外进行共享使用，也可以当作文件服务器使用。管理端可自己创建文件共享目录，目录分为存储型和链接型，存储型上传的文件会存储到server上，链接型目录只做文件完整路径的链接，服务端不会存储实际文件，会从运行端磁盘直接获取文件共享给他人下载，比如文件上G就不用本地再上传存储一次。

   server.json是配置文件，如果当服务器使用建议将配置中链接型linkDirAdd开关关闭，避免出现安全事故。

   如果你下载已经制作好的exe直接运行就可以，目前由于只有Windows环境，只打包了windows版本，支持linux、mac等多系统，需要自己去编译。

   [Windows打包版本下载 FileShare-V20250331 ](https://github.com/newlxj/FileShare/releases/download/v1.0.0/FileShare-V20250331.zip)
## 运行截图

<img src=image/1share.png width=70% />
<img src=image/2login.png width=70% />
<img src=image/3manage.png width=70% />

## 技术实现

- 前端使用vue3 后端使用golang

## 构建步骤

### 自动构建（推荐）

项目提供了自动构建脚本，可以一键完成前后端的构建：

1. 进入`backend`目录
2. 运行`build.bat`脚本
3. 脚本会自动构建前端和后端，并生成包含前端资源的可执行文件`fileShare.exe`

### 手动构建

如果需要手动构建，请按照以下步骤操作：

1. 构建前端：
   ```
   cd frontend
   npm install
   npm run build
   ```
 ```
   前端页面特别注意，由于golang将前端打包到exe中方便运行，如果重新打包需要手动将web/index.html 和web/assets/index***.js中 assets/ 替换成fileserver/assets/
  ```
2. 构建后端：
   ```
   cd backend
   go build -o fileShare.exe
   ```
3.全自动构建前后端
   ```
  build.bat
   ```
## 运行方式

构建完成后，直接运行生成的可执行文件即可：

```
./fileShare.exe
```

或在Windows环境下双击`fileShare.exe`。

访问地址：http://localhost:8080/fileserver/
管理端密码默认：123456
## 注意事项

- 首次运行时，系统会自动创建必要的配置文件
- 默认端口为8080，可以通过配置文件修改
- 嵌入的前端资源会通过配置的上下文路径提供服务
- 

## 配置说明

系统配置文件位于`config`目录下：

- `server.json`: 服务器配置，包括端口、上下文路径等
- `config-group.json`: 目录配置
- `config-file.json`: 文件配置

## 优势

- 简化部署流程，只需一个可执行文件
- 无需配置前端服务器
- 前后端版本一致，避免兼容性问题
- 便于分发和安装
- 文件服务双击启动
- 支持mac等多种设备