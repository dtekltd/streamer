import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

import LoginPage from '../pages/LoginPage.vue'
import ProfilesPage from '../pages/ProfilesPage.vue'
import StreamingPage from '../pages/StreamingPage.vue'

const routes = [
  {
    path: '/',
    redirect: () => (useAuthStore().isAuthenticated ? '/profiles' : '/login')
  },
  {
    path: '/login',
    component: LoginPage,
    meta: { public: true }
  },
  {
    path: '/profiles',
    component: ProfilesPage
  },
  {
    path: '/streaming',
    component: StreamingPage
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.public) {
    if (to.path === '/login' && auth.isAuthenticated) {
      return '/profiles'
    }
    return true
  }

  if (!auth.isAuthenticated) {
    return '/login'
  }

  return true
})

export default router
