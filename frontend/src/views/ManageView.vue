<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Lock, Download } from '@element-plus/icons-vue'
import type Node from 'element-plus/es/components/tree/src/model/node'

// 目录树数据结构
interface TreeNode {
  id: string
  label: string
  children?: TreeNode[]
  isShared?: boolean
  parentId?: string
  dirType?: string // 目录类型：'link'(链接型) 或 'storage'(存储型)
  password?: string // 目录密码
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

// 认证状态
const isAuthenticated = ref(false)

// 目录树数据
const treeData = ref<TreeNode[]>([])

// 当前选中的目录节点
const currentNode = ref<TreeNode | null>(null)

// 文件列表数据
const fileList = ref<FileItem[]>([])

// 登录密码
const password = ref('')

// 获取token
const getToken = (): string | null => {
  const cookies = document.cookie.split(';')
  for (const cookie of cookies) {
    const [name, value] = cookie.trim().split('=')
    if (name === 'admin_token') {
      return value
    }
  }
  return null
}

// 检查是否已认证
const checkAuthentication = () => {
  const token = getToken()
  isAuthenticated.value = !!token
}

// 右键菜单相关
const contextMenuVisible = ref(false)
const contextMenuPosition = reactive({
  top: '0px',
  left: '0px'
})

// 加载目录树数据
const loadTreeData = async () => {
  try {
    const token = getToken()
    if (!token) {
      ElMessage.error('未授权，请先登录')
      return
    }
    
    const response = await fetch('/fileshare/api/directories', {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })
    
    if (response.status === 401) {
      ElMessage.error('授权已过期，请重新登录')
      logout()
      return
    }
    
    const data = await response.json()
    // 将后端返回的name字段映射为前端TreeNode接口所需的label字段
    treeData.value = mapNameToLabel(data)
  } catch (error) {
    ElMessage.error('加载目录数据失败')
    console.error(error)
  }
}

// 加载文件列表
const loadFileList = async (directoryId: string) => {
  try {
    const token = getToken()
    if (!token) {
      ElMessage.error('未授权，请先登录')
      return
    }
    
    const response = await fetch(`/fileshare/api/files?directoryId=${directoryId}`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })
    
    if (response.status === 401) {
      ElMessage.error('授权已过期，请重新登录')
      logout()
      return
    }
    
    fileList.value = await response.json()
  } catch (error) {
    ElMessage.error('加载文件列表失败')
    console.error(error)
  }
}

// 选择目录节点
const handleNodeClick = (data: TreeNode) => {
  currentNode.value = data
  loadFileList(data.id)
}

// 显示右键菜单
const showContextMenu = (event: MouseEvent, data: TreeNode) => {
  event.preventDefault()
  currentNode.value = data
  contextMenuPosition.top = `${event.clientY}px`
  contextMenuPosition.left = `${event.clientX}px`
  contextMenuVisible.value = true
  
  // 点击其他地方关闭菜单
  document.addEventListener('click', closeContextMenu, { once: true })
}

// 关闭右键菜单
const closeContextMenu = () => {
  contextMenuVisible.value = false
}

// 添加目录
const addDirectory = async () => {
  try {
    if (!currentNode.value) {
      ElMessage.warning('请先选择一个目录')
      return
    }
    
    
    // 使用ElMessageBox.prompt的高级用法，自定义内容
    const { value: formValues } = await ElMessageBox.prompt(
      `<div>
        
        <div>
          <label style="display: block; margin-bottom: 5px;">目录类型</label>
          <div style="display: flex; gap: 15px;">
            <label style="display: flex; align-items: center;">
              <input type="radio" name="dirType" value="storage" checked /> 存储型
            </label>
            <label style="display: flex; align-items: center;">
              <input type="radio" name="dirType" value="link" /> 链接型
            </label>
          </div>
        </div>
        <div style="margin-bottom: 15px;">
          <label style="display: block; margin-bottom: 5px;">目录名称</label>
          <input 
            class="el-input__inner" 
            value="" 
            placeholder="输入目录名称，在这里输入下面那个不是" 
            id="dirName" 
            style="
              width: 80vh;
              border: 1px solid #dcdfe6;
              border-radius: 4px;
              padding: 0 15px;
              height: 32px;
              line-height: 32px;
              background-color: #fff;
              color: #606266;
            "
          />
        </div>
      </div>`,
      '添加目录',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        dangerouslyUseHTMLString: true,
        inputValidator: () => {
          const dirName = document.getElementById('dirName') as HTMLInputElement
          if (!dirName || !dirName.value.trim()) {
            return '目录名称不能为空'
          }
          return true
        },
        beforeClose: (action, instance, done) => {
          if (action === 'confirm') {
            const dirName = document.getElementById('dirName') as HTMLInputElement
            const dirTypeRadios = document.getElementsByName('dirType') as NodeListOf<HTMLInputElement>
            let dirType = 'storage' // 默认为存储型
            
            for (const radio of dirTypeRadios) {
              if (radio.checked) {
                dirType = radio.value
                break
              }
            }
            
            // 定义表单值对象
            const formValue = {
              name: dirName.value.trim(),
              dirType: dirType
            }
            // 将值赋给instance.inputValue
            instance.inputValue = JSON.stringify(formValue)
          }
          done()
        }
      }
    )
    
    if (!formValues) return
    
    // 解析表单值
    const { name: directoryName, dirType } = JSON.parse(formValues)
    
    // 调用后端API创建目录
    const token = getToken()
    if (!token) {
      ElMessage.error('未授权，请先登录')
      return
    }
    
    const response = await fetch('/fileshare/api/directories', {
      method: 'POST',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ 
        name: directoryName, 
        parentId: currentNode.value.id,
        dirType: dirType // 添加目录类型
      })
    })
    
    // 检查响应状态
    if (!response.ok) {
      const errorData = await response.json()
      // 特别处理链接型目录未开启的情况
      if (errorData.error && errorData.error.includes('Adding link directories is not allowed by server configuration')) {
        ElMessage.error('服务器配置不允许创建链接型目录，请联系管理员')
        return
      }
      throw new Error(errorData.error || '添加目录失败')
    }
    
    // 重新加载目录树
    await loadTreeData()
    const newId = Date.now().toString()
    const newNode: TreeNode = {
      id: newId,
      label: directoryName,
      isShared: false,
      parentId: currentNode.value.id,
      dirType: dirType, // 添加目录类型
      children: []
    }
    
    if (!currentNode.value.children) {
      currentNode.value.children = []
    }
    
    currentNode.value.children.push(newNode)
    ElMessage.success('添加目录成功')
  } catch (error) {
    ElMessage.error('添加目录失败')
    console.error(error)
  }
}

// 删除目录
const deleteDirectory = async () => {
  try {
    if (!currentNode.value) {
      ElMessage.warning('请先选择一个目录')
      return
    }
    
    await ElMessageBox.confirm(
      `确定要删除目录 "${currentNode.value.label}" 吗？删除后将无法恢复，且会删除该目录下的所有文件。`,
      '删除目录',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 调用后端API删除目录
    const token = getToken()
    if (!token) {
      ElMessage.error('未授权，请先登录')
      return
    }
    
    await fetch(`/fileshare/api/directories/${currentNode.value.id}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })
    
    // 重新加载目录树
    await loadTreeData()
    if (currentNode.value.parentId) {
      const parentNode = findNodeById(treeData.value, currentNode.value.parentId)
      if (parentNode && parentNode.children) {
        const index = parentNode.children.findIndex(child => child.id === currentNode.value?.id)
        if (index !== -1) {
          parentNode.children.splice(index, 1)
          ElMessage.success('删除目录成功')
          currentNode.value = null
          fileList.value = []
        }
      }
    } else {
      const index = treeData.value.findIndex(node => node.id === currentNode.value?.id)
      if (index !== -1) {
        treeData.value.splice(index, 1)
        ElMessage.success('删除目录成功')
        currentNode.value = null
        fileList.value = []
      }
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除目录失败')
      console.error(error)
    }
  }
}

// 重命名目录
const renameDirectory = async () => {
  try {
    if (!currentNode.value) {
      ElMessage.warning('请先选择一个目录')
      return
    }
    
    const { value: newName } = await ElMessageBox.prompt(
      '请输入新的目录名称',
      '重命名目录',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        inputValue: currentNode.value.label,
        inputValidator: (value) => {
          if (!value) {
            return '目录名称不能为空'
          }
          return true
        }
      }
    )
    
    if (!newName || newName === currentNode.value.label) return
    
    // 调用后端API重命名目录
    const token = getToken()
    if (!token) {
      ElMessage.error('未授权，请先登录')
      return
    }
    
    await fetch(`/fileshare/api/directories/${currentNode.value.id}`, {
      method: 'PUT',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ name: newName })
    })
    
    // 重新加载目录树
    await loadTreeData()
    currentNode.value.label = newName
    ElMessage.success('重命名目录成功')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('重命名目录失败')
      console.error(error)
    }
  }
}

// 切换目录共享状态
const toggleDirectoryShare = async () => {
  try {
    if (!currentNode.value) {
      ElMessage.warning('请先选择一个目录')
      return
    }
    
    const newShareStatus = !currentNode.value.isShared
    
    // 调用后端API更新目录共享状态
    const token = getToken()
    if (!token) {
      ElMessage.error('未授权，请先登录')
      return
    }
    
    await fetch(`/fileshare/api/directories/${currentNode.value.id}/share`, {
      method: 'PATCH',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ isShared: newShareStatus })
    })
    
    // 更新当前节点的共享状态
    currentNode.value.isShared = newShareStatus
    
    if (newShareStatus) {
      // 如果是设置为共享，则递归向上查找并设置所有父级目录为共享状态
      setParentDirectoriesShared(currentNode.value.parentId, true)
    } else {
      // 如果是取消共享，则递归向下设置所有子目录为非共享状态
      setChildDirectoriesShared(currentNode.value, false)
    }
    
    ElMessage.success(`目录已${newShareStatus ? '共享' : '取消共享'}`)
  } catch (error) {
    ElMessage.error('更新目录共享状态失败')
    console.error(error)
  }
}

// 递归设置父级目录的共享状态
const setParentDirectoriesShared = async (parentId: string | undefined, isShared: boolean) => {
  if (!parentId) return
  
  // 查找父级目录节点
  const parentNode = findNodeById(treeData.value, parentId)
  if (!parentNode) return
  
  // 如果父级目录已经是共享状态，则不需要再设置
  if (parentNode.isShared === isShared) return
  
  // 获取token
  const token = getToken()
  if (!token) return
  
  // 调用API更新父级目录共享状态
  try {
    await fetch(`/fileshare/api/directories/${parentNode.id}/share`, {
      method: 'PATCH',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ isShared })
    })
    
    // 更新父级目录的共享状态
    parentNode.isShared = isShared
    
    // 继续向上递归设置
    setParentDirectoriesShared(parentNode.parentId, isShared)
  } catch (error) {
    console.error('更新父级目录共享状态失败:', error)
  }
}

// 递归设置子目录的共享状态
const setChildDirectoriesShared = async (node: TreeNode, isShared: boolean) => {
  if (!node.children || node.children.length === 0) return
  
  const token = getToken()
  if (!token) return
  
  // 遍历所有子目录
  for (const child of node.children) {
    // 如果子目录的共享状态与要设置的状态不同，则更新
    if (child.isShared !== isShared) {
      try {
        await fetch(`/fileshare/api/directories/${child.id}/share`, {
          method: 'PATCH',
          headers: { 
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
          },
          body: JSON.stringify({ isShared })
        })
        
        // 更新子目录的共享状态
        child.isShared = isShared
      } catch (error) {
        console.error('更新子目录共享状态失败:', error)
        continue
      }
    }
    
    // 递归处理子目录的子目录
    setChildDirectoriesShared(child, isShared)
  }
}

// 设置目录密码
const setDirectoryPassword = async () => {
  try {
    if (!currentNode.value) {
      ElMessage.warning('请先选择一个目录')
      return
    }
    
    const { value: password } = await ElMessageBox.prompt(
      '请输入目录密码（留空表示不设置密码）',
      '设置密码',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        inputType: 'password',
        inputValue: '',
      }
    )
    
    // 调用后端API设置密码
    const token = getToken()
    if (!token) {
      ElMessage.error('未授权，请先登录')
      return
    }
    
    await fetch(`/fileshare/api/directories/${currentNode.value.id}/password`, {
      method: 'PATCH',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ password })
    })
    
    ElMessage.success(password ? '密码设置成功' : '密码已清除')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('设置密码失败')
      console.error(error)
    }
  }
}

// 添加文件
const addFile = () => {
  if (!currentNode.value) {
    ElMessage.warning('请先选择一个目录')
    return
  }
  
  // 触发文件选择框
  const fileInput = document.createElement('input')
  fileInput.type = 'file'
  fileInput.multiple = true
  fileInput.onchange = handleFileSelect
  fileInput.click()
}

// 处理文件选择
const handleFileSelect = async (event: Event) => {
  const input = event.target as HTMLInputElement
  if (!input.files || !currentNode.value) return
  
  try {
    const files = Array.from(input.files)
    
    // 获取当前目录类型
    const dirType = currentNode.value.dirType || 'storage' // 默认为存储型
    
    // 调用后端API上传文件
    const token = getToken()
    if (!token) {
      ElMessage.error('未授权，请先登录')
      return
    }
    
    const formData = new FormData()
    
    if (dirType === 'link') {
      // 链接型目录：需要获取文件的完整路径
      // 注意：由于浏览器安全限制，无法直接获取文件的完整路径
      // 这里需要用户手动输入文件路径
      
      // 获取所选文件的文件名，用于在输入框中显示
      const fileNames = files.map(file => file.name).join('\n')
      
      const { value: filePaths } = await ElMessageBox.prompt(
        '请在文件名前补充完整路径，多个文件路径请用换行分隔，例如:c:\\temp\\showInfo.png  或linux /opt/dameng/info.log',
        '链接型目录文件路径',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          inputType: 'textarea',
          inputValue: fileNames, // 预填充文件名
          inputValidator: (value) => {
            if (!value) {
              return '文件路径不能为空'
            }
            return true
          }
        }
      )
      
      if (!filePaths) return
      
      // 分割多个文件路径
      const paths = filePaths.split('\n').filter(path => path.trim() !== '')
      
      formData.append('filePaths', JSON.stringify(paths))
      formData.append('directoryId', currentNode.value.id)
      formData.append('dirType', 'link')
      
      // 发送请求
      const response = await fetch('/fileshare/api/files', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`
        },
        body: formData
      })
      
      // 检查响应状态
      if (!response.ok) {
        const errorData = await response.json()
        // 特别处理链接型目录未开启的情况
        if (errorData.error && errorData.error.includes('Adding files to link directories is not allowed by server configuration')) {
          ElMessage.error('服务器配置不允许向链接型目录添加文件，请联系管理员')
          return
        }
        throw new Error(errorData.error || '添加文件失败')
      }
      
      // 获取响应数据
      const newFiles = await response.json()
      
      // 更新文件列表
      fileList.value = [...fileList.value, ...newFiles]
      ElMessage.success(`成功添加 ${newFiles.length} 个文件`)
    } else {
      // 存储型目录：上传实际文件
      files.forEach(file => formData.append('files', file))
      formData.append('directoryId', currentNode.value.id)
      formData.append('dirType', 'storage')
      
      // 发送请求
      const response = await fetch('/fileshare/api/files', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`
        },
        body: formData
      })
      // 检查响应状态
      if (!response.ok) {
        const errorData = await response.json()
        
        // 特别处理链接型目录未开启的情况
        if (errorData.error && errorData.error.includes('Adding files to link directories is not allowed by server configuration')) {
          ElMessage.error('服务器配置不允许向链接型目录添加文件，请联系管理员')
          return
        }
        throw new Error(errorData.error || '添加文件失败')
      }
      
      // 获取响应数据
      const newFiles = await response.json()
      
      // 更新文件列表
      fileList.value = [...fileList.value, ...newFiles]
      ElMessage.success(`成功添加 ${newFiles.length} 个文件`)
    }
    
    // 重新加载文件列表
    await loadFileList(currentNode.value.id)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('添加文件失败')
      console.error(error)
    }
  }
}

// 下载文件
const downloadFile = async (file: FileItem) => {
  try {
    // 获取token
    const token = getToken()
    if (!token) {
      ElMessage.error('未授权，请先登录')
      return
    }
    
    // 调用后端API下载文件
    const response = await fetch(`/fileshare/api/files/${file.id}/download`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })
    
    // 检查响应状态
    if (response.status === 401) {
      ElMessage.error('授权已过期，请重新登录')
      logout()
      return
    }
    
    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.error || '下载文件失败')
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

// 删除文件
const deleteFile = async (file: FileItem) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除文件 "${file.name}" 吗？删除后将无法恢复。`,
      '删除文件',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 调用后端API删除文件
    const token = getToken()
    if (!token) {
      ElMessage.error('未授权，请先登录')
      return
    }
    
    await fetch(`/fileshare/api/files/${file.id}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })
    
    // 模拟删除文件
    const index = fileList.value.findIndex(item => item.id === file.id)
    if (index !== -1) {
      fileList.value.splice(index, 1)
      ElMessage.success('删除文件成功')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除文件失败')
      console.error(error)
    }
  }
}

// 重命名文件
const renameFile = async (file: FileItem) => {
  try {
    const { value: newName } = await ElMessageBox.prompt(
      '请输入新的文件名称',
      '重命名文件',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        inputValue: file.name,
        inputValidator: (value) => {
          if (!value) {
            return '文件名称不能为空'
          }
          return true
        }
      }
    )
    
    if (!newName || newName === file.name) return
    
    // 调用后端API重命名文件
    const token = getToken()
    if (!token) {
      ElMessage.error('未授权，请先登录')
      return
    }
    
    await fetch(`/fileshare/api/files/${file.id}`, {
      method: 'PATCH',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ name: newName })
    })
    
    // 模拟重命名文件
    file.name = newName
    ElMessage.success('重命名文件成功')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('重命名文件失败')
      console.error(error)
    }
  }
}

// 切换文件共享状态
const toggleFileShare = async (file: FileItem) => {
  try {
    const newShareStatus = !file.isShared
    
    // 调用后端API更新文件共享状态
    const token = getToken()
    if (!token) {
      ElMessage.error('未授权，请先登录')
      return
    }
    
    await fetch(`/fileshare/api/files/${file.id}/share`, {
      method: 'PATCH',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ isShared: newShareStatus })
    })
    
    // 更新共享状态
    file.isShared = newShareStatus
    ElMessage.success(`文件已${newShareStatus ? '共享' : '取消共享'}`)
  } catch (error) {
    ElMessage.error('更新文件共享状态失败')
    console.error(error)
  }
}

// 处理文件拖放
const handleFileDrop = async (event: DragEvent) => {
  event.preventDefault()
  event.stopPropagation()
  
  if (!currentNode.value) {
    ElMessage.warning('请先选择一个目录')
    return
  }
  
  if (!event.dataTransfer?.files?.length) {
    ElMessage.warning('没有有效的文件')
    return
  }
  
  try {
    const files = Array.from(event.dataTransfer.files)
    
    // 获取当前目录类型
    const dirType = currentNode.value.dirType || 'storage' // 默认为存储型
    
    // 调用后端API上传文件
    const token = getToken()
    if (!token) {
      ElMessage.error('未授权，请先登录')
      return
    }
    
    const formData = new FormData()
    
    if (dirType === 'link') {
      // 链接型目录：需要获取文件的完整路径
      // 由于浏览器安全限制，无法直接获取文件的完整路径
      // 这里需要用户手动输入文件路径
      
      // 获取所选文件的文件名，用于在输入框中显示
      const fileNames = files.map(file => file.name).join('\n')
      
      const { value: filePaths } = await ElMessageBox.prompt(
        '请在文件名前补充完整路径，多个文件路径请用换行分隔，例如:c:\\temp\\showInfo.png  或linux /opt/dameng/info.log',
        '链接型目录文件路径',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          inputType: 'textarea',
          inputValue: fileNames, // 预填充文件名
          inputValidator: (value) => {
            if (!value) {
              return '文件路径不能为空'
            }
            return true
          }
        }
      )
      
      if (!filePaths) return
      
      // 分割多个文件路径
      const paths = filePaths.split('\n').filter(path => path.trim() !== '')
      
      formData.append('filePaths', JSON.stringify(paths))
      formData.append('directoryId', currentNode.value.id)
      formData.append('dirType', 'link')
    } else {
      // 存储型目录：上传实际文件
      files.forEach(file => formData.append('files', file))
      formData.append('directoryId', currentNode.value.id)
      formData.append('dirType', 'storage')
    }
    
    const response = await fetch('/fileshare/api/files', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`
      },
      body: formData
    })
    
    // 获取响应数据
    const newFiles = await response.json()
    
    // 更新文件列表
    if (Array.isArray(newFiles) && newFiles.length > 0) {
      fileList.value = [...fileList.value, ...newFiles]
      ElMessage.success(`成功添加 ${newFiles.length} 个文件`)
    } else {
      // 如果返回空数组或非数组，可能是添加失败
      ElMessage.warning('未能添加文件，请检查文件路径是否正确')
    }
    
    // 重新加载文件列表以确保数据同步
    await loadFileList(currentNode.value.id)
  } catch (error) {
    ElMessage.error('添加文件失败')
    console.error(error)
  }
}

// 允许拖放文件
const allowDrop = (event: DragEvent) => {
  event.preventDefault()
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

// 将目录数据中的name字段映射为label字段
const mapNameToLabel = (directories: any[]): TreeNode[] => {
  return directories.map(dir => ({
    ...dir,
    label: dir.name, // 将name映射为label
    children: dir.children ? mapNameToLabel(dir.children) : undefined
  }))
}

// 根据ID查找节点
const findNodeById = (nodes: TreeNode[], id: string): TreeNode | null => {
  for (const node of nodes) {
    if (node.id === id) {
      return node
    }
    if (node.children && node.children.length > 0) {
      const found = findNodeById(node.children, id)
      if (found) {
        return found
      }
    }
  }
  return null
}

// 登录函数
const login = async () => {
  try {
    if (!password.value) {
      ElMessage.warning('请输入管理密码')
      return
    }
    
    const response = await fetch('/fileshare/api/admin/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password: password.value })
    })
    
    if (!response.ok) {
      const errorData = await response.json()
      ElMessage.error(errorData.error || '登录失败，密码错误')
      return
    }
    
    const data = await response.json()
    // 将token存储到cookie中
    document.cookie = `admin_token=${data.token}; path=/; max-age=86400`
    isAuthenticated.value = true
    password.value = ''
    ElMessage.success('登录成功')
    loadTreeData()
  } catch (error) {
    ElMessage.error('登录失败，请稍后重试')
    console.error(error)
  }
}

// 退出登录
const logout = () => {
  document.cookie = 'admin_token=; path=/; expires=Thu, 01 Jan 1970 00:00:01 GMT;'
  isAuthenticated.value = false
  treeData.value = []
  fileList.value = []
  currentNode.value = null
  ElMessage.success('已退出登录')
}

// 页面加载时获取数据
onMounted(() => {
  checkAuthentication()
  if (isAuthenticated.value) {
    loadTreeData()
  }
})
</script>

<template>
  <div class="manage-container">
    <!-- 登录表单 -->
    <div v-if="!isAuthenticated" class="login-container">
      <div class="login-form">
        <h2>管理员登录</h2>
        <el-form @submit.prevent="login">
          <el-form-item label="管理密码">
            <el-input v-model="password" type="password" placeholder="请输入管理密码" @keyup.enter.prevent="login" autofocus />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="login">登录</el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>
    
    <!-- 管理界面 -->
    <template v-else>
      
      
      <!-- 左侧目录树 -->
      <div class="directory-tree">
        <div class="header-actions">
        <h2>目录管理</h2>
        <el-button type="danger" size="small" @click="logout">退出登录</el-button>
      </div>
        
        <el-tree
          :data="treeData"
          node-key="id"
          default-expand-all
          :expand-on-click-node="false"
          highlight-current
          @node-click="handleNodeClick"
          @node-contextmenu="showContextMenu"
        >
          <template #default="{ node, data }">
            <span class="custom-tree-node">
              <el-icon><Folder /></el-icon>
              <span>{{ node.label }}</span>
              <el-tag v-if="data.isShared" size="small" type="success" effect="plain">已共享</el-tag>
              <el-tag v-if="data.dirType === 'link'" size="small" type="info" effect="plain">链接型</el-tag>
              <el-tag v-else-if="data.dirType === 'storage'" size="small" type="primary" effect="plain">存储型</el-tag>
              <el-icon v-if="data.password" color="#E6A23C"><Lock /></el-icon>
            </span>
          </template>
        </el-tree>
        
        <!-- 右键菜单 -->
        <div v-show="contextMenuVisible" class="context-menu" :style="contextMenuPosition">
          <ul>
            <li @click="addDirectory">添加子目录</li>
            <li @click="renameDirectory">重命名</li>
            <li @click="toggleDirectoryShare">
              {{ currentNode?.isShared ? '取消共享' : '设为共享' }}
            </li>
            <li @click="setDirectoryPassword">设置密码</li>
            <li @click="deleteDirectory" class="danger">删除</li>
          </ul>
        </div>
      </div>
      
      <!-- 右侧文件列表 -->
      <div 
        class="file-list" 
        @dragover="allowDrop" 
        @drop="handleFileDrop"
      >
        <div class="file-list-header">
          <h2>文件列表 - {{ currentNode?.label || '请选择目录' }}</h2>
          <el-button 
            type="primary" 
            :disabled="!currentNode" 
            @click="addFile"
          >
            <el-icon><Plus /></el-icon> 添加文件
          </el-button>
        </div>
        
        <div v-if="!currentNode" class="empty-tip">
          请先从左侧选择一个目录
        </div>
        
        <div v-else-if="fileList.length === 0" class="empty-tip">
          <el-empty description="暂无文件，请添加文件或拖拽文件到此处" />
        </div>
        
        <el-table v-else :data="fileList" style="width: 100%">
          <el-table-column label="文件名" min-width="200">
            <template #default="{ row }">
              <div class="file-name">
                <el-icon><Document /></el-icon>
                <span>{{ row.name }}</span>
                <el-tag v-if="row.isShared" size="small" type="success" effect="plain">已共享</el-tag>
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
          <el-table-column label="操作" width="220">
            <template #default="{ row }">
              <el-button-group>
                <el-button size="small" @click="renameFile(row)">
                  <el-icon><Edit /></el-icon>
                </el-button>
                <el-button 
                  size="small" 
                  :type="row.isShared ? 'success' : 'info'" 
                  @click="toggleFileShare(row)"
                >
                  <el-icon><Share /></el-icon>
                </el-button>
                <el-button size="small" type="primary" @click="downloadFile(row)">
                  <el-icon><Download /></el-icon>
                </el-button>
                <el-button size="small" type="danger" @click="deleteFile(row)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </template>
  </div>
</template>

<style scoped>
.manage-container {
  display: flex;
  height: 100%;
  gap: 20px;
  width: 100%;
  min-height: 100vh;
  padding: 20px;
  box-sizing: border-box;
}

/* 登录表单样式 */
.login-container {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}

.login-form {
  width: 400px;
  padding: 30px;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  background-color: white;
}

.login-form h2 {
  text-align: center;
  margin-bottom: 20px;
  color: #409eff;
}

/* 管理界面样式 */
.header-actions {
  display: flex;
  justify-content:flex-start;
  margin-bottom: 20px;
}

.directory-tree {
  width: 300px;
  border-right: 1px solid #e4e7ed;
  padding-right: 20px;
  overflow: auto;
}

.file-list {
  flex: 1;
  overflow: auto;
}

.file-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.empty-tip {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 200px;
  color: #909399;
}

.file-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.context-menu {
  position: fixed;
  z-index: 1000;
  background: white;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  padding: 5px 0;
}

.context-menu ul {
  list-style: none;
  margin: 0;
  padding: 0;
}

.context-menu li {
  padding: 8px 16px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.context-menu li:hover {
  background-color: #f5f7fa;
}

.context-menu li.danger {
  color: #f56c6c;
}

.context-menu li.danger:hover {
  background-color: #fef0f0;
}

.custom-tree-node {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>