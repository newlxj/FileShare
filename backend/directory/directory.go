package directory

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"fileshare/common"
	"fileshare/config"
	"fileshare/models"
)

// 配置文件路径
const (
	GroupConfigPath = "./config/config-group.json"
)

// 加载目录配置
func LoadDirectories() {
	data, err := os.ReadFile(GroupConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			models.Directories = []*models.Directory{}
			return
		}
		log.Printf("Failed to read directory config: %v", err)
		return
	}

	if err := json.Unmarshal(data, &models.Directories); err != nil {
		log.Printf("Failed to parse directory config: %v", err)
		models.Directories = []*models.Directory{}
	}
}

// 保存目录配置
func SaveDirectories() error {
	data, err := json.MarshalIndent(models.Directories, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(GroupConfigPath, data, 0644)
}

// 获取所有目录
func GetDirectories(c *gin.Context) {
	c.JSON(http.StatusOK, models.Directories)
}

// 获取共享目录
func GetSharedDirectories(c *gin.Context) {
	sharedDirs := FilterSharedDirectories(models.Directories)
	c.JSON(http.StatusOK, sharedDirs)
}

// 过滤共享目录
func FilterSharedDirectories(dirs []*models.Directory) []*models.Directory {
	result := []*models.Directory{}

	for _, dir := range dirs {
		if dir.IsShared {
			// 创建一个新的目录对象，避免修改原始数据
			sharedDir := &models.Directory{
				ID:       dir.ID,
				Name:     dir.Name,
				ParentID: dir.ParentID,
				IsShared: dir.IsShared,
				Password: dir.Password,
			}

			// 递归处理子目录
			if len(dir.Children) > 0 {
				sharedDir.Children = FilterSharedDirectories(dir.Children)
			}

			result = append(result, sharedDir)
		} else if len(dir.Children) > 0 {
			// 检查子目录是否有共享的
			sharedChildren := FilterSharedDirectories(dir.Children)
			if len(sharedChildren) > 0 {
				// 创建一个新的目录对象，只包含共享的子目录
				sharedDir := &models.Directory{
					ID:       dir.ID,
					Name:     dir.Name,
					ParentID: dir.ParentID,
					IsShared: dir.IsShared,
					Password: dir.Password,
					Children: sharedChildren,
				}
				result = append(result, sharedDir)
			}
		}
	}

	return result
}

// 创建目录
func CreateDirectory(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		ParentID string `json:"parentId"`
		DirType  string `json:"dirType"` // 目录类型：link(链接型) 或 storage(存储型)
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 如果未指定目录类型，默认为存储型
	if req.DirType == "" {
		req.DirType = "storage"
	}

	// 检查是否允许添加链接型目录
	serverConfig := config.GetServerConfig()
	if req.DirType == "link" && !serverConfig.Server.LinkDirAdd {
		c.JSON(http.StatusForbidden, gin.H{"error": "Adding link directories is not allowed by server configuration"})
		return
	}

	// 创建新目录
	newDir := &models.Directory{
		ID:       uuid.New().String(),
		Name:     req.Name,
		ParentID: req.ParentID,
		IsShared: false,
		DirType:  req.DirType,
		Children: []*models.Directory{},
	}

	// 如果有父目录，添加到父目录的子目录中
	if req.ParentID != "" {
		parentFound := false
		AddToParent(models.Directories, req.ParentID, newDir, &parentFound)

		if !parentFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Parent directory not found"})
			return
		}
	} else {
		// 添加到根目录
		models.Directories = append(models.Directories, newDir)
	}

	// 保存配置
	if err := SaveDirectories(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save directory"})
		return
	}

	c.JSON(http.StatusCreated, newDir)
}

// 递归添加到父目录
func AddToParent(dirs []*models.Directory, parentID string, newDir *models.Directory, found *bool) {
	for _, dir := range dirs {
		if dir.ID == parentID {
			dir.Children = append(dir.Children, newDir)
			*found = true
			return
		}

		if len(dir.Children) > 0 {
			AddToParent(dir.Children, parentID, newDir, found)
			if *found {
				return
			}
		}
	}
}

// 更新目录
func UpdateDirectory(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找并更新目录
	dirFound := false
	UpdateDirectoryName(models.Directories, id, req.Name, &dirFound)

	if !dirFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Directory not found"})
		return
	}

	// 保存配置
	if err := SaveDirectories(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save directory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Directory updated successfully"})
}

// 递归更新目录名称
func UpdateDirectoryName(dirs []*models.Directory, id string, newName string, found *bool) {
	for _, dir := range dirs {
		if dir.ID == id {
			dir.Name = newName
			*found = true
			return
		}

		if len(dir.Children) > 0 {
			UpdateDirectoryName(dir.Children, id, newName, found)
			if *found {
				return
			}
		}
	}
}

// 删除目录
func DeleteDirectory(c *gin.Context) {
	id := c.Param("id")

	// 从根目录中删除
	for i, dir := range models.Directories {
		if dir.ID == id {
			// 删除该目录下的所有文件
			common.DeleteFilesInDirectory(id)

			// 删除目录
			models.Directories = append(models.Directories[:i], models.Directories[i+1:]...)

			// 保存配置
			if err := SaveDirectories(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save directory"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Directory deleted successfully"})
			return
		}
	}

	// 从子目录中删除
	dirFound := false
	DeleteFromParent(models.Directories, id, &dirFound)

	if !dirFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Directory not found"})
		return
	}

	// 删除该目录下的所有文件
	common.DeleteFilesInDirectory(id)

	// 保存配置
	if err := SaveDirectories(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save directory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Directory deleted successfully"})
}

// 递归从父目录中删除子目录
func DeleteFromParent(dirs []*models.Directory, id string, found *bool) {
	for _, dir := range dirs {
		if len(dir.Children) > 0 {
			for i, child := range dir.Children {
				if child.ID == id {
					// 删除子目录
					dir.Children = append(dir.Children[:i], dir.Children[i+1:]...)
					*found = true
					return
				}
			}

			DeleteFromParent(dir.Children, id, found)
			if *found {
				return
			}
		}
	}
}

// 切换目录共享状态
func ToggleDirectoryShare(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		IsShared bool `json:"isShared"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找并更新目录共享状态
	dirFound := false
	UpdateDirectoryShare(models.Directories, id, req.IsShared, &dirFound)

	if !dirFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Directory not found"})
		return
	}

	// 保存配置
	if err := SaveDirectories(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save directory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Directory share status updated successfully"})
}

// 递归更新目录共享状态
func UpdateDirectoryShare(dirs []*models.Directory, id string, isShared bool, found *bool) {
	for _, dir := range dirs {
		if dir.ID == id {
			dir.IsShared = isShared
			*found = true
			return
		}

		if len(dir.Children) > 0 {
			UpdateDirectoryShare(dir.Children, id, isShared, found)
			if *found {
				return
			}
		}
	}
}

// 设置目录密码
func SetDirectoryPassword(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找并更新目录密码
	dirFound := false
	UpdateDirectoryPassword(models.Directories, id, req.Password, &dirFound)

	if !dirFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Directory not found"})
		return
	}

	// 保存配置
	if err := SaveDirectories(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save directory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Directory password updated successfully"})
}

// 递归更新目录密码
func UpdateDirectoryPassword(dirs []*models.Directory, id string, password string, found *bool) {
	for _, dir := range dirs {
		if dir.ID == id {
			dir.Password = password
			*found = true
			return
		}

		if len(dir.Children) > 0 {
			UpdateDirectoryPassword(dir.Children, id, password, found)
			if *found {
				return
			}
		}
	}
}

// 验证目录密码
func VerifyDirectoryPassword(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找目录
	var targetDir *models.Directory
	common.FindDirectory(models.Directories, id, &targetDir)

	if targetDir == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Directory not found"})
		return
	}

	// 验证密码
	if targetDir.Password != "" && targetDir.Password != req.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password verified successfully"})
}

// 检查目录是否存在
func CheckDirectoryExists(dirs []*models.Directory, id string, exists *bool) {
	for _, dir := range dirs {
		if dir.ID == id {
			*exists = true
			return
		}

		if len(dir.Children) > 0 {
			CheckDirectoryExists(dir.Children, id, exists)
			if *exists {
				return
			}
		}
	}
}
