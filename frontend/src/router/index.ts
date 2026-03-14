import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { pinia } from '@/plugins/pinia'
import { useAuthStore } from '@/stores'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
    meta: { guest: true },
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('@/views/RegisterView.vue'),
    meta: { guest: true },
  },
  {
    path: '/',
    name: 'app',
    component: () => import('@/views/AppView.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'home',
        redirect: '/channels',
      },
      {
        path: 'channels/:channelId?',
        name: 'channels',
        component: () => import('@/views/ChannelView.vue'),
      },
      {
        path: 'projects/:projectId',
        name: 'project',
        component: () => import('@/views/ProjectView.vue'),
        children: [
          {
            path: 'channels/:channelId?',
            name: 'project-channel',
            component: () => import('@/views/ChannelView.vue'),
          },
          {
            path: 'tasks',
            name: 'project-tasks',
            component: () => import('@/views/TasksView.vue'),
          },
        ],
      },
      {
        path: 'tasks',
        name: 'tasks',
        component: () => import('@/views/TasksView.vue'),
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to) => {
  const authStore = useAuthStore(pinia)

  if (!authStore.initialized) {
    await authStore.initialize()
  }

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return '/login'
  }
  if (to.meta.guest && authStore.isAuthenticated) {
    return '/'
  }

  return true
})

export default router
