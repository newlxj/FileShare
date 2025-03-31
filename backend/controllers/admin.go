package controllers

import (
	"fmt"
	"net/http"

	"fileshare/config"
	"fileshare/utils"

	"github.com/gin-gonic/gin"
)

// AdminLoginRequest 管理员登录请求结构
type AdminLoginRequest struct {
	Password string `json:"password" binding:"required"`
}

// AdminLoginResponse 管理员登录响应结构
type AdminLoginResponse struct {
	Token string `json:"token"`
}

// AdminLogin 处理管理员登录
func AdminLogin(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 验证密码
	serverConfig := config.GetServerConfig()
	if req.Password != fmt.Sprintf("%d", serverConfig.Server.ManagePassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		return
	}

	// 生成token
	token := utils.GenerateToken()

	// 返回token
	c.JSON(http.StatusOK, AdminLoginResponse{Token: token})
}
