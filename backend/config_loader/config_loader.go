package config_loader

import (
	"os"

	"fileshare/directory"
	"fileshare/file"
)

// 配置文件路径
const (
	ConfigDirPath = "./config"
)

// 确保配置目录存在
func EnsureConfigDir() error {
	if _, err := os.Stat(ConfigDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(ConfigDirPath, 0755); err != nil {
			return err
		}

		// 创建空的配置文件
		if err := directory.SaveDirectories(); err != nil {
			return err
		}
		if err := file.SaveFiles(); err != nil {
			return err
		}
	}
	return nil
}

// 加载所有配置
func LoadAllConfigs() {
	// 加载目录配置
	directory.LoadDirectories()

	// 加载文件配置
	file.LoadFiles()
}
