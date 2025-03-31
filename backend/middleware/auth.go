package middleware

import (
	"net/http"
	"strings"

	"fileshare/utils"

	"github.com/gin-gonic/gin"
)

// AdminAuth 中间件用于验证管理员权限
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取token
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权，请先登录"})
			c.Abort()
			return
		}

		// 处理Bearer token格式
		token := auth
		if strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimPrefix(auth, "Bearer ")
		}

		// 验证token
		if !utils.ValidateToken(token) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token，请重新登录"})
			c.Abort()
			return
		}

		// 继续处理请求
		c.Next()
	}
}
