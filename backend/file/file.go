package file

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"fileshare/common"
	"fileshare/config"
	"fileshare/models"
)

// 配置文件路径
const (
	FileConfigPath = "./config/config-file.json"
)

// 加载文件配置
func LoadFiles() {
	data, err := os.ReadFile(FileConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			models.Files = []*models.File{}
			return
		}
		log.Printf("Failed to read file config: %v", err)
		return
	}

	if err := json.Unmarshal(data, &models.Files); err != nil {
		log.Printf("Failed to parse file config: %v", err)
		models.Files = []*models.File{}
	}
}

// 保存文件配置
func SaveFiles() error {
	data, err := json.MarshalIndent(models.Files, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(FileConfigPath, data, 0644)
}

// 获取文件列表
func GetFiles(c *gin.Context) {
	directoryID := c.Query("directoryId")

	if directoryID == "" {
		c.JSON(http.StatusOK, models.Files)
		return
	}

	// 过滤指定目录的文件
	dirFiles := []*models.File{}
	for _, file := range models.Files {
		if file.DirectoryID == directoryID {
			dirFiles = append(dirFiles, file)
		}
	}

	c.JSON(http.StatusOK, dirFiles)
}

// 获取共享文件列表
func GetSharedFiles(c *gin.Context) {
	directoryID := c.Query("directoryId")

	// 过滤共享文件
	sharedFiles := []*models.File{}
	for _, file := range models.Files {
		if file.IsShared && (directoryID == "" || file.DirectoryID == directoryID) {
			sharedFiles = append(sharedFiles, file)
		}
	}

	c.JSON(http.StatusOK, sharedFiles)
}

// 上传文件
func UploadFiles(c *gin.Context) {
	directoryID := c.PostForm("directoryId")
	if directoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Directory ID is required"})
		return
	}

	// 检查目录是否存在
	var targetDir *models.Directory
	common.FindDirectory(models.Directories, directoryID, &targetDir)
	if targetDir == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Directory not found"})
		return
	}

	// 获取目录类型
	dirType := c.PostForm("dirType")
	if dirType == "" {
		// 如果未指定，使用目录的类型
		dirType = targetDir.DirType
		if dirType == "" {
			// 如果目录也没有指定类型，默认为存储型
			dirType = "storage"
		}
	}

	// 根据目录类型处理文件上传
	newFiles := []*models.File{}

	if dirType == "link" {
		// 检查是否允许添加链接型目录
		serverConfig := config.GetServerConfig()
		if !serverConfig.Server.LinkDirAdd {
			c.JSON(http.StatusForbidden, gin.H{"error": "Adding files to link directories is not allowed by server configuration"})
			return
		}

		// 链接型目录：只保存文件路径
		filePathsJSON := c.PostForm("filePaths")
		if filePathsJSON == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File paths are required for link directory"})
			return
		}

		var filePaths []string
		if err := json.Unmarshal([]byte(filePathsJSON), &filePaths); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file paths format"})
			return
		}

		for _, filePath := range filePaths {
			// 获取文件信息
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				// 如果文件不存在，跳过
				continue
			}

			// 创建文件记录
			fileID := uuid.New().String()
			fileExt := filepath.Ext(filePath)
			newFile := &models.File{
				ID:          fileID,
				Name:        filepath.Base(filePath),
				Path:        filePath, // 保存原始文件路径
				Size:        fileInfo.Size(),
				Type:        fileExt[1:],                              // 去掉点号
				AddTime:     time.Now().Format("2006-01-02 15:04:05"), // 格式化时间
				IsShared:    false,
				DirectoryID: directoryID,
			}

			models.Files = append(models.Files, newFile)
			newFiles = append(newFiles, newFile)
		}
	} else {
		// 存储型目录：上传实际文件
		// 获取上传的文件
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		uploadedFiles := form.File["files"]
		if len(uploadedFiles) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
			return
		}

		// 确保文件存储目录存在
		serverConfig := config.GetServerConfig()
		staticDir := serverConfig.Server.FilestorePath
		if _, err := os.Stat(staticDir); os.IsNotExist(err) {
			if err := os.MkdirAll(staticDir, 0755); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file storage directory"})
				return
			}
		}

		// 处理上传的文件
		for _, fileHeader := range uploadedFiles {
			// 生成唯一文件名
			fileID := uuid.New().String()
			fileExt := filepath.Ext(fileHeader.Filename)
			filePath := filepath.Join(staticDir, fileID+fileExt)

			// 保存文件到磁盘
			if err := c.SaveUploadedFile(fileHeader, filePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
				return
			}

			// 创建文件记录
			newFile := &models.File{
				ID:          fileID,
				Name:        fileHeader.Filename,
				Path:        filePath,
				Size:        fileHeader.Size,
				Type:        fileExt[1:],                              // 去掉点号
				AddTime:     time.Now().Format("2006-01-02 15:04:05"), // 格式化时间
				IsShared:    false,
				DirectoryID: directoryID,
			}

			models.Files = append(models.Files, newFile)
			newFiles = append(newFiles, newFile)
		}
	}

	// 保存配置
	if err := SaveFiles(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file records"})
		return
	}

	c.JSON(http.StatusCreated, newFiles)
}

// 删除文件
func DeleteFile(c *gin.Context) {
	id := c.Param("id")

	// 查找文件
	var fileToDelete *models.File
	var fileIndex int
	for i, file := range models.Files {
		if file.ID == id {
			fileToDelete = file
			fileIndex = i
			break
		}
	}

	if fileToDelete == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// 查找文件所在的目录，确定目录类型
	var targetDir *models.Directory
	common.FindDirectory(models.Directories, fileToDelete.DirectoryID, &targetDir)

	// 只有存储型目录才删除物理文件
	if targetDir == nil || targetDir.DirType != "link" {
		// 删除物理文件（非链接型目录）
		if err := os.Remove(fileToDelete.Path); err != nil && !os.IsNotExist(err) {
			log.Printf("Failed to delete file %s: %v", fileToDelete.Path, err)
			// 继续执行，即使物理文件删除失败
		}
	}

	// 从记录中删除
	models.Files = append(models.Files[:fileIndex], models.Files[fileIndex+1:]...)

	// 保存配置
	if err := SaveFiles(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file records"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

// 更新文件
func UpdateFile(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找并更新文件
	fileFound := false
	for _, file := range models.Files {
		if file.ID == id {
			file.Name = req.Name
			fileFound = true
			break
		}
	}

	if !fileFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// 保存配置
	if err := SaveFiles(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file records"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File updated successfully"})
}

// 切换文件共享状态
func ToggleFileShare(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		IsShared bool `json:"isShared"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找并更新文件共享状态
	fileFound := false
	for _, file := range models.Files {
		if file.ID == id {
			file.IsShared = req.IsShared
			fileFound = true
			break
		}
	}

	if !fileFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// 保存配置
	if err := SaveFiles(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file records"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File share status updated successfully"})
}

// 下载文件
func DownloadFile(c *gin.Context) {
	id := c.Param("id")

	// 查找文件
	var fileToDownload *models.File
	for _, file := range models.Files {
		if file.ID == id {
			fileToDownload = file
			break
		}
	}

	if fileToDownload == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// 检查文件是否共享
	if !fileToDownload.IsShared {
		c.JSON(http.StatusForbidden, gin.H{"error": "File is not shared"})
		return
	}

	// 打开文件
	file, err := os.Open(fileToDownload.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer file.Close()

	// 设置响应头
	c.Header("Content-Disposition", "attachment; filename="+fileToDownload.Name)
	c.Header("Content-Type", "application/octet-stream")

	// 发送文件内容
	_, err = io.Copy(c.Writer, file)
	if err != nil {
		log.Printf("Failed to send file: %v", err)
	}
}

// 管理员下载文件（不检查共享状态）
func AdminDownloadFile(c *gin.Context) {
	id := c.Param("id")

	// 查找文件
	var fileToDownload *models.File
	for _, file := range models.Files {
		if file.ID == id {
			fileToDownload = file
			break
		}
	}

	if fileToDownload == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// 打开文件
	file, err := os.Open(fileToDownload.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer file.Close()

	// 设置响应头
	c.Header("Content-Disposition", "attachment; filename="+fileToDownload.Name)
	c.Header("Content-Type", "application/octet-stream")

	// 发送文件内容
	_, err = io.Copy(c.Writer, file)
	if err != nil {
		log.Printf("Failed to send file: %v", err)
	}
}
