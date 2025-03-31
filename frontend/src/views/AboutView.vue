<template>
  <div class="about">
    <div class="about-content">
      <h1 class="about-title" v-animate-on-scroll>使用说明</h1>
      <div class="about-description">
        <p class="welcome-text" v-animate-on-scroll>本系统是对外进行文件共享的工具，在server.json可配置服务端口及管理密码(默认：123456)，管理维护页面左侧是分类树，可以在上级节点上右键添加子节点、修改名称、目录共享、设置目录密码等。右侧文件列表可点添加或拖拽文件进来添加。删除也是虚拟删除。</p>
        <p class="solution-text" v-animate-on-scroll>本系统提供文件存储和本机文件引用共享两个功能，分别是存储型目录和链接型目录，存储型目录下上传的文件都会存储到服务器上，链接型类似引用功能(快捷方式)，只共享链接指定的文件，不会再次进行存储。</p>
        <div class="feature-container" v-animate-on-scroll>
          <div class="feature-card" v-for="(feature, index) in features" :key="index">
            <i :class="feature.icon"></i>
            <span>{{ feature.text }}</span>
          </div>
        </div>
      </div>
      <div class="contact-info" v-animate-on-scroll>
        <p>Create By 刘秀君</p>
        <p class="email"><i class="fas fa-envelope"></i>文件共享系统</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const features = [
  { text: '快速传输', icon: 'fas fa-bolt' },
  { text: '安全可靠', icon: 'fas fa-shield-alt' },
  { text: '简单易用', icon: 'fas fa-magic' },
  { text: '多平台支持', icon: 'fas fa-desktop' }
]
</script>

<style scoped>
.about {
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  padding: 2rem;
}

.about-content {
  max-width: 1200px;
  margin: 0 auto;
  text-align: center;
}

.about-title {
  font-size: 3rem;
  color: #2c3e50;
  margin-bottom: 2rem;
  position: relative;
  display: inline-block;
}

.about-title::after {
  content: '';
  position: absolute;
  bottom: -10px;
  left: 50%;
  transform: translateX(-50%);
  width: 60px;
  height: 3px;
  background: #3498db;
  transition: width 0.3s ease;
}

.about-title:hover::after {
  width: 100px;
}

.welcome-text, .solution-text {
  font-size: 1.5rem;
  color: #34495e;
  margin: 1rem 0;
}

.feature-container {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 2rem;
  margin: 3rem 0;
}

.feature-card {
  background: white;
  padding: 2rem;
  border-radius: 10px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.feature-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 6px 12px rgba(0, 0, 0, 0.15);
}

.feature-card i {
  font-size: 2rem;
  color: #3498db;
  margin-bottom: 1rem;
}

.contact-info {
  margin-top: 3rem;
  padding: 2rem;
  background: rgba(255, 255, 255, 0.9);
  border-radius: 10px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.email {
  color: #3498db;
  font-size: 1.2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

/* 添加动画效果 */
[v-animate-on-scroll] {
  opacity: 0;
  transform: translateY(30px);
  animation: fadeInUp 0.6s ease forwards;
}

@keyframes fadeInUp {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 768px) {
  .about-title {
    font-size: 2rem;
  }
  
  .welcome-text, .solution-text {
    font-size: 1.2rem;
  }
  
  .feature-container {
    grid-template-columns: 1fr;
  }
}
</style>

<style>
@media (min-width: 1024px) {
  .about {
    min-height: 20vh;
    display: flex;
    align-items: center;
  }
}
</style>
