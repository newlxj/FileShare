<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

// 目录树数据结构
interface TreeNode {
  id: string
  label: string
  children?: TreeNode[]
  isShared: boolean
  password?: string // 添加密码字段，用于判断是否显示锁图标
}

// 文件数据结构
interface FileItem {
  id: string
  name: string
  path: string
  size: number
  type: string
  addTime: string
  isShared: boolean
  directoryId: string
}

// 目录树数据
const treeData = ref<TreeNode[]>([])

// 当前选中的目录节点
const currentNode = ref<TreeNode | null>(null)

// 文件列表数据
const fileList = ref<FileItem[]>([])

// 密码验证相关
// 存储目录ID和对应的密码，以及验证状态
const passwordVerified = reactive(new Map<string, boolean>())
const directoryPasswords = reactive(new Map<string, string>())

// 加载共享目录树数据
const loadSharedTreeData = async () => {
  try {
    const response = await fetch('/filesharePreview/api/directories/shared')
    const data = await response.json()
    // 将后端返回的name字段映射为前端TreeNode接口所需的label字段
    treeData.value = mapNameToLabel(data)
  } catch (error) {
    ElMessage.error('加载共享目录数据失败')
    console.error(error)
  }
}

// 将目录数据中的name字段映射为label字段
const mapNameToLabel = (directories: any[]): TreeNode[] => {
  return directories.map(dir => ({
    ...dir,
    label: dir.name, // 将name映射为label
    children: dir.children ? mapNameToLabel(dir.children) : undefined,
    // 确保密码信息被保留，用于显示锁图标
    password: dir.password
  }))
}

// 加载共享文件列表
const loadSharedFileList = async (directoryId: string) => {
  try {
    // 检查是否需要密码验证
    if (!passwordVerified.get(directoryId)) {
      const verifyResult = await verifyDirectoryPassword(directoryId)
      if (!verifyResult) {
        return
      }
    }
    
    const response = await fetch(`/filesharePreview/api/files/shared?directoryId=${directoryId}`)
    fileList.value = await response.json()
  } catch (error) {
    ElMessage.error('加载共享文件列表失败')
    console.error(error)
  }
}

// 验证目录密码
const verifyDirectoryPassword = async (directoryId: string): Promise<boolean> => {
  try {
    // 先尝试不带密码访问，看是否需要密码
    const testResponse = await fetch(`/filesharePreview/api/directories/${directoryId}/verify`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password: '' })
    })
    
    const testResult = await testResponse.json()
    
    // 如果不需要密码或密码为空，直接返回成功
    // 检查后端返回的消息格式，可能是{valid:true}或{message:"Password verified successfully"}
    if (testResult.valid || testResult.message === "Password verified successfully") {
      passwordVerified.set(directoryId, true)
      directoryPasswords.set(directoryId, '') // 存储空密码
      return true
    }
    
    // 需要密码，弹出密码输入框
    const { value: password } = await ElMessageBox.prompt(
      '此目录受密码保护，请输入密码',
      '密码验证',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        inputType: 'password',
        inputValidator: (value) => {
          if (!value) {
            return '密码不能为空'
          }
          return true
        }
      }
    )
    
    if (!password) return false
    
    // 验证密码
    const response = await fetch(`/filesharePreview/api/directories/${directoryId}/verify`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password })
    })
    
    const result = await response.json()
    
    // 检查后端返回的消息格式，可能是{valid:true}或{message:"Password verified successfully"}
    if (result.valid || result.message === "Password verified successfully") {
      passwordVerified.set(directoryId, true)
      directoryPasswords.set(directoryId, password) // 存储验证成功的密码
      return true
    } else {
      ElMessage.error('密码错误')
      return false
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('验证密码失败')
      console.error(error)
    }
    return false
  }
}

// 选择目录节点
const handleNodeClick = (data: TreeNode) => {
  currentNode.value = data
  loadSharedFileList(data.id)
}

// 下载文件
const downloadFile = async (file: FileItem) => {
  try {
    // 检查是否需要密码验证
    if (!passwordVerified.get(file.directoryId)) {
      const verifyResult = await verifyDirectoryPassword(file.directoryId)
      if (!verifyResult) {
        return
      }
    }
    
    // 实际调用后端API下载文件，带上已验证的密码参数
    const password = directoryPasswords.get(file.directoryId) || ''
    const response = await fetch(`/filesharePreview/api/files/${file.id}/download?password=${password}`)
    
    // 检查响应状态
    if (response.status === 403) {
      const errorData = await response.json()
      if (errorData.requirePassword) {
        // 需要密码验证
        const verifyResult = await verifyDirectoryPassword(errorData.directoryId)
        if (verifyResult) {
          // 验证成功，重新下载
          return downloadFile(file)
        }
        return
      }
    }
    
    const blob = await response.blob()
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = file.name
    document.body.appendChild(a)
    a.click()
    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)
    
    ElMessage.success(`开始下载文件: ${file.name}`)
  } catch (error) {
    ElMessage.error('下载文件失败')
    console.error(error)
  }
}

// 格式化文件大小
const formatFileSize = (size: number): string => {
  if (size < 1024) {
    return size + ' B'
  } else if (size < 1024 * 1024) {
    return (size / 1024).toFixed(2) + ' KB'
  } else if (size < 1024 * 1024 * 1024) {
    return (size / (1024 * 1024)).toFixed(2) + ' MB'
  } else {
    return (size / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
  }
}

// 过滤只显示共享的目录和文件
const filterSharedNodes = (nodes: TreeNode[]): TreeNode[] => {
  return nodes
    .filter(node => node.isShared || (node.children && node.children.some(child => child.isShared)))
    .map(node => {
      if (node.children) {
        return {
          ...node,
          children: filterSharedNodes(node.children)
        }
      }
      return node
    })
}

// 页面加载时获取数据
onMounted(() => {
  loadSharedTreeData()
})
</script>

<template>
  <div class="share-container">
    <!-- 左侧共享目录树 -->
    <div class="directory-tree">
      <h2>共享目录</h2>
      <el-tree
        :data="filterSharedNodes(treeData)"
        node-key="id"
        default-expand-all
        :expand-on-click-node="false"
        highlight-current
        @node-click="handleNodeClick"
      >
        <template #default="{ node, data }">
          <span class="custom-tree-node">
            <el-icon><Folder /></el-icon>
            <span>{{ node.label }}</span>
            <el-icon v-if="data.password" class="lock-icon"><Lock /></el-icon>
          </span>
        </template>
      </el-tree>
    </div>
    
    <!-- 右侧共享文件列表 -->
    <div class="file-list">
      <h2>共享文件 - {{ currentNode?.label || '请选择目录' }}</h2>
      
      <div v-if="!currentNode" class="empty-tip">
        请先从左侧选择一个共享目录
      </div>
      
      <div v-else-if="fileList.length === 0" class="empty-tip">
        <el-empty description="该目录下暂无共享文件" />
      </div>
      
      <el-table v-else :data="fileList" style="width: 100%">
        <el-table-column label="文件名" min-width="200">
          <template #default="{ row }">
            <div class="file-name">
              <el-icon><Document /></el-icon>
              <el-link type="primary" @click="downloadFile(row)">{{ row.name }}</el-link>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="100" />
        <el-table-column label="大小" width="120">
          <template #default="{ row }">
            {{ formatFileSize(row.size) }}
          </template>
        </el-table-column>
        <el-table-column prop="addTime" label="添加时间" width="180" />
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="downloadFile(row)">
              <el-icon><Download /></el-icon> 下载
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<style scoped>
.share-container {
  display: flex;
  width: 100%;
  min-height: calc(100vh - 120px);
  gap: 20px;
  padding: 20px;
  box-sizing: border-box;
}

.directory-tree {
  width: 300px;
  border-right: 1px solid #e4e7ed;
  padding: 20px;
}

.file-list {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

.file-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.custom-tree-node {
  display: flex;
  align-items: center;
  gap: 5px;
}

.file-name {
  display: flex;
  align-items: center;
  gap: 5px;
}

.empty-tip {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 200px;
  color: #909399;
}

.lock-icon {
  color: #E6A23C;
  margin-left: 5px;
}
</style>