package utils

import (
	"crypto/md5"
	"encoding/hex"
	"time"

	"fileshare/config"
)

// 存储有效的token
var validTokens = make(map[string]time.Time)

// GenerateToken 生成管理员token
func GenerateToken() string {
	// 获取当前时间戳
	timestamp := time.Now().Unix()

	// 获取管理密码
	serverConfig := config.GetServerConfig()
	password := serverConfig.Server.ManagePassword

	// 生成token (密码 + 时间戳的MD5)
	data := []byte(string(password) + string(timestamp))
	hash := md5.Sum(data)
	token := hex.EncodeToString(hash[:])

	// 存储token，有效期24小时
	validTokens[token] = time.Now().Add(24 * time.Hour)

	return token
}

// ValidateToken 验证token是否有效
func ValidateToken(token string) bool {
	expiry, exists := validTokens[token]
	if !exists {
		return false
	}

	// 检查token是否过期
	if time.Now().After(expiry) {
		delete(validTokens, token)
		return false
	}

	return true
}

// InvalidateToken 使token失效
func InvalidateToken(token string) {
	delete(validTokens, token)
}
