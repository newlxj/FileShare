<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { ElConfigProvider } from 'element-plus'
import 'element-plus/dist/index.css'
</script>

<template>
  <el-config-provider>
    <div class="app-container">
      <header>
        <div class="wrapper">
          <h1 class="app-title">文件共享系统</h1>
          <nav>
            <RouterLink to="/fileserver">共享</RouterLink>
            <RouterLink to="/manage">管理维护</RouterLink>
            <RouterLink to="/about">关于</RouterLink>
          </nav>
        </div>
      </header>

      <main>
        <RouterView />
      </main>
    </div>
  </el-config-provider>
</template>

<style scoped>
.app-container {
  width: 100%;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  margin: 0 auto;
}

header {
  line-height: 1.5;
  padding: 1rem 2rem;
  background-color: #f0f2f5;
  border-bottom: 1px solid #e4e7ed;
}

.app-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: #409eff;
  margin: 0;
}

nav {
  margin-top: 1rem;
}

nav a {
  display: inline-block;
  padding: 0.5rem 1rem;
  margin-right: 1rem;
  border-radius: 4px;
  color: #606266;
  text-decoration: none;
  transition: all 0.3s;
}

nav a.router-link-active {
  color: #409eff;
  background-color: #ecf5ff;
}

main {
  flex: 1;
  overflow: auto;
}
</style>
