import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/today' },
    { path: '/today', name: 'today', component: () => import('@/views/Today.vue'), meta: { title: '今日' } },
    { path: '/week', name: 'week', component: () => import('@/views/Week.vue'), meta: { title: '周视图' } },
    { path: '/goals', name: 'goals', component: () => import('@/views/Goals.vue'), meta: { title: '目标' } },
    { path: '/tasks', name: 'tasks', component: () => import('@/views/Tasks.vue'), meta: { title: '任务' } },
    { path: '/schedule', name: 'schedule', component: () => import('@/views/Schedule.vue'), meta: { title: '日程' } },
    { path: '/stats', name: 'stats', component: () => import('@/views/Stats.vue'), meta: { title: '统计' } },
    { path: '/review', name: 'review', component: () => import('@/views/Review.vue'), meta: { title: '复盘' } },
    { path: '/settings', name: 'settings', component: () => import('@/views/Settings.vue'), meta: { title: '设置' } },
  ],
})

router.afterEach((to) => {
  document.title = to.meta.title ? `${String(to.meta.title)} · StudyPilot` : 'StudyPilot'
})

export default router