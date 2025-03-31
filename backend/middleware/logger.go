package middleware

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"fileshare/config"
)

// Logger 中间件用于记录请求日志
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取服务器配置
		serverConfig := config.GetServerConfig()
		logPath := serverConfig.Server.LogPath

		// 打开日志文件
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("无法打开日志文件: %v\n", err)
			c.Next()
			return
		}
		defer logFile.Close()

		// 记录请求开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 计算请求处理时间
		latency := time.Since(startTime)

		// 获取客户端IP
		clientIP := c.ClientIP()

		// 获取请求方法和路径
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		if query != "" {
			path = path + "?" + query
		}

		// 获取状态码
		statusCode := c.Writer.Status()

		// 获取错误信息
		errorMsg := ""
		if len(c.Errors) > 0 {
			errorMsg = c.Errors.String()
		}

		// 构建日志内容
		logContent := fmt.Sprintf("[%s] %s | %d | %v | %s | %s",
			time.Now().Format("2006-01-02 15:04:05"),
			clientIP,
			statusCode,
			latency,
			method+" "+path,
			errorMsg,
		)

		// 写入日志文件
		_, err = io.WriteString(logFile, logContent+"\n")
		if err != nil {
			fmt.Printf("写入日志失败: %v\n", err)
		}
	}
}
