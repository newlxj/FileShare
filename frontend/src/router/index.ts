import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/fileserver',
      name: 'share',
      component: () => import('../views/ShareView.vue'),
    },
    {
      path: '/manage',
      name: 'manage',
      component: () => import('../views/ManageView.vue'),
    },
    {
      path: '/about',
      name: 'about',
      // 声明 Vue 组件类型
      component: () => import('../views/AboutView.vue') ,
    },
  ],
})

export default router
