package models

// 目录结构
type Directory struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	ParentID string       `json:"parentId,omitempty"`
	IsShared bool         `json:"isShared"`
	Password string       `json:"password,omitempty"`
	DirType  string       `json:"dirType,omitempty"` // 目录类型：link(链接型) 或 storage(存储型)
	Children []*Directory `json:"children,omitempty"`
}

// 文件结构
type File struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	Type        string `json:"type"`
	AddTime     string `json:"addTime"` // 格式化的时间字符串：2006-03-12 22:12:33
	IsShared    bool   `json:"isShared"`
	DirectoryID string `json:"directoryId"`
}

// 全局变量
var (
	// 目录存储
	Directories []*Directory

	// 文件存储
	Files []*File
)
