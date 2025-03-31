package config

import (
	"encoding/json"
	"os"
	"sync"
)

// ServerConfig 服务器配置结构体
type ServerConfig struct {
	Server struct {
		Port              int    `json:"port"`
		ContextPath       string `json:"contextPath"`
		ContextManagePath string `json:"contextManagePath"` // 管理API的上下文路径，不是web页面路径
		ContextSharePath  string `json:"contextSharePath"`  // 共享API的上下文路径，不是web页面路径
		ManagePassword    int    `json:"managePassword"`
		LogPath           string `json:"logPath"`
		LinkDirAdd        bool   `json:"linkDirAdd"`
		FilestorePath     string `json:"filestorePath"` // 文件存储路径
	} `json:"server"`
}

var (
	serverConfig     *ServerConfig
	serverConfigOnce sync.Once
)

// GetServerConfig 获取服务器配置
func GetServerConfig() *ServerConfig {
	serverConfigOnce.Do(func() {
		serverConfig = &ServerConfig{}
		serverConfig.Server.Port = 8080 // 默认值
		serverConfig.Server.ContextPath = "/fileshare"
		serverConfig.Server.ContextManagePath = "/fileshare" // 管理API的上下文路径，不是web页面路径
		serverConfig.Server.ContextSharePath = "/fileshare"  // 共享API的上下文路径，不是web页面路径
		serverConfig.Server.ManagePassword = 123456
		serverConfig.Server.LogPath = "./recode.log"
		serverConfig.Server.LinkDirAdd = true          // 默认允许添加链接型目录
		serverConfig.Server.FilestorePath = "./static" // 默认文件存储路径

		// 尝试从配置文件加载
		data, err := os.ReadFile("./config/server.json")
		if err == nil {
			_ = json.Unmarshal(data, serverConfig)
		}
	})

	return serverConfig
}
