package common

import (
	"fileshare/models"
)

// 查找目录
func FindDirectory(dirs []*models.Directory, id string, result **models.Directory) {
	for _, dir := range dirs {
		if dir.ID == id {
			*result = dir
			return
		}

		if len(dir.Children) > 0 {
			FindDirectory(dir.Children, id, result)
			if *result != nil {
				return
			}
		}
	}
}

// 删除目录下的所有文件
func DeleteFilesInDirectory(dirID string) {
	// 删除直接属于该目录的文件
	for i := 0; i < len(models.Files); {
		if models.Files[i].DirectoryID == dirID {
			models.Files = append(models.Files[:i], models.Files[i+1:]...)
		} else {
			i++
		}
	}

	// 递归查找并删除子目录中的文件
	var findChildDirs func(dirs []*models.Directory, parentID string) []string
	findChildDirs = func(dirs []*models.Directory, parentID string) []string {
		childIDs := []string{}

		for _, dir := range dirs {
			if dir.ParentID == parentID {
				childIDs = append(childIDs, dir.ID)
				childIDs = append(childIDs, findChildDirs(dirs, dir.ID)...)
			}

			if len(dir.Children) > 0 {
				for _, child := range dir.Children {
					if child.ParentID == parentID {
						childIDs = append(childIDs, child.ID)
						childIDs = append(childIDs, findChildDirs(dir.Children, child.ID)...)
					}
				}
			}
		}

		return childIDs
	}

	// 获取所有子目录ID
	childDirIDs := findChildDirs(models.Directories, dirID)

	// 删除子目录中的文件
	for _, childID := range childDirIDs {
		for i := 0; i < len(models.Files); {
			if models.Files[i].DirectoryID == childID {
				models.Files = append(models.Files[:i], models.Files[i+1:]...)
			} else {
				i++
			}
		}
	}
}
