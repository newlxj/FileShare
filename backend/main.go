package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"fileshare/config"
	"fileshare/config_loader"
	"fileshare/controllers"
	"fileshare/directory"
	"fileshare/file"
	"fileshare/middleware"
)

//go:embed web/*
var webFS embed.FS

func main() {
	// 确保配置目录存在
	if err := config_loader.EnsureConfigDir(); err != nil {
		log.Fatalf("Failed to create config directory: %v", err)
	}

	// 加载配置
	config_loader.LoadAllConfigs()

	// 获取服务器配置
	serverConfig := config.GetServerConfig()

	// 创建Gin路由
	r := gin.New()

	// 使用日志中间件
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 静态文件服务
	r.Static("/static", "./static")

	// 设置嵌入式web目录的静态文件服务
	// 获取web子文件系统
	subFS, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("Failed to get web subdirectory: %v", err)
	}

	// 设置上下文路径,管理端API的上下文路径
	contextManagePath := serverConfig.Server.ContextManagePath

	// 管理员API路由组
	admin := r.Group(contextManagePath + "/api/admin")
	{
		// 管理员登录
		admin.POST("/login", controllers.AdminLogin)
	}

	// 管理API路由组（需要认证）
	api := r.Group(serverConfig.Server.ContextManagePath + "/api")
	// 添加认证中间件
	api.Use(middleware.AdminAuth())
	{
		// 目录相关API
		api.GET("/directories", directory.GetDirectories)
		api.POST("/directories", directory.CreateDirectory)
		api.PUT("/directories/:id", directory.UpdateDirectory)
		api.DELETE("/directories/:id", directory.DeleteDirectory)
		api.PATCH("/directories/:id/share", directory.ToggleDirectoryShare)
		api.PATCH("/directories/:id/password", directory.SetDirectoryPassword)

		// 文件相关API
		api.GET("/files", file.GetFiles)
		api.POST("/files", file.UploadFiles)
		api.DELETE("/files/:id", file.DeleteFile)
		api.PATCH("/files/:id", file.UpdateFile)
		api.PATCH("/files/:id/share", file.ToggleFileShare)
		api.GET("/files/:id/download", file.AdminDownloadFile)
	}

	// 共享预览API路由组（不需要认证）
	shareApi := r.Group(serverConfig.Server.ContextSharePath + "/api")
	{
		// 共享目录和文件API
		shareApi.GET("/directories/shared", directory.GetSharedDirectories)
		shareApi.POST("/directories/:id/verify", directory.VerifyDirectoryPassword)
		shareApi.GET("/files/shared", file.GetSharedFiles)
		shareApi.GET("/files/:id/download", file.DownloadFile)
	}

	// 提供嵌入式web目录
	r.StaticFS("/fileserver", http.FS(subFS))
	// r.StaticFS("/index.html", http.FS(subFS))

	// 提供favicon.ico
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS("favicon.ico", http.FS(subFS))
	})
	// 提供根路径和index.html的处理
	// r.GET("/", func(c *gin.Context) {
	// 	c.FileFromFS("assets", http.FS(subFS))
	// })
	// r.GET("/fileserver", func(c *gin.Context) {
	// 	c.FileFromFS("assets", http.FS(subFS))
	// })

	// r.GET("/index.html", func(c *gin.Context) {
	// 	c.FileFromFS("index.html", http.FS(subFS))
	// })

	// 处理SPA路由 - 对于前端路由的所有请求返回index.html
	r.NoRoute(func(c *gin.Context) {
		// 如果请求的是API路径，则返回404
		if strings.HasPrefix(c.Request.URL.Path, "/api/") ||
			strings.HasPrefix(c.Request.URL.Path, serverConfig.Server.ContextManagePath+"/api/") ||
			strings.HasPrefix(c.Request.URL.Path, serverConfig.Server.ContextSharePath+"/api/") {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// 否则返回index.html以支持前端路由
		c.FileFromFS("index.html", http.FS(subFS))
	})

	// 启动服务器
	port := serverConfig.Server.Port
	log.Printf("\n\n 服务器已运行，访问地址： http://localhost:%d/fileserver/ \n\n", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
