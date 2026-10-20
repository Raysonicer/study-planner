<script setup lang="ts">
import { useRoute } from 'vue-router'
import { computed } from 'vue'

const route = useRoute()
const active = computed(() => route.path)

const menus = [
  { path: '/today', title: '今日', icon: 'Sunny' },
  { path: '/week', title: '周视图', icon: 'Calendar' },
  { path: '/goals', title: '目标', icon: 'Aim' },
  { path: '/tasks', title: '任务', icon: 'List' },
  { path: '/schedule', title: '日程', icon: 'Clock' },
  { path: '/stats', title: '统计', icon: 'TrendCharts' },
  { path: '/review', title: '复盘', icon: 'EditPen' },
  { path: '/settings', title: '设置', icon: 'Setting' },
]
</script>

<template>
  <div class="app-shell">
    <div class="bg-deco bg-deco-1" />
    <div class="bg-deco bg-deco-2" />
    <div class="bg-deco bg-deco-3" />
    <el-container class="layout">
      <el-aside width="212px" class="aside">
        <div class="logo"><span class="logo-dot" />StudyPilot</div>
        <el-menu :default-active="active" router>
          <el-menu-item v-for="m in menus" :key="m.path" :index="m.path">
            <el-icon><component :is="m.icon" /></el-icon>
            <span>{{ m.title }}</span>
          </el-menu-item>
        </el-menu>
        <div class="aside-foot">v0.2.2 · 数据存于浏览器</div>
      </el-aside>
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </div>
</template>

<style>
/* ---- 全局主题（覆盖 Element Plus 变量） ---- */
:root {
  --el-color-primary: #4c6ef5;
  --el-color-primary-light-3: #7c95f8;
  --el-color-primary-light-5: #a5b6fb;
  --el-color-primary-light-7: #cdd7fd;
  --el-color-primary-light-8: #e0e7fe;
  --el-color-primary-light-9: #f0f3ff;
  --el-color-primary-dark-2: #3b5bdb;
}

html,
body,
#app {
  height: 100%;
  margin: 0;
}

body {
  font-family: 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', -apple-system, sans-serif;
  -webkit-font-smoothing: antialiased;
}

.app-shell {
  position: relative;
  height: 100%;
  overflow: auto;
  background:
    linear-gradient(140deg, #e9effd 0%, #f4f7fc 28%, #fdf6ee 62%, #eef7f2 100%);
  background-attachment: fixed;
}

/* 背景装饰光斑：让底色有层次、不单调 */
.bg-deco {
  position: fixed;
  border-radius: 50%;
  filter: blur(90px);
  opacity: 0.55;
  pointer-events: none;
  z-index: 0;
}
.bg-deco-1 {
  width: 480px;
  height: 480px;
  background: radial-gradient(circle, #cdd9ff 0%, transparent 70%);
  top: -140px;
  right: -120px;
}
.bg-deco-2 {
  width: 420px;
  height: 420px;
  background: radial-gradient(circle, #d9f2e5 0%, transparent 70%);
  bottom: -160px;
  left: 200px;
}
.bg-deco-3 {
  width: 320px;
  height: 320px;
  background: radial-gradient(circle, #ffe3d1 0%, transparent 70%);
  top: 40%;
  left: -120px;
}

.layout {
  position: relative;
  z-index: 1;
  height: 100%;
  background: transparent;
}

.aside {
  display: flex;
  flex-direction: column;
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(14px);
  border-right: 1px solid rgba(148, 163, 184, 0.18);
}

.logo {
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 22px 18px 14px;
  font-size: 19px;
  font-weight: 800;
  background: linear-gradient(90deg, #3b5bdb, #6741d9);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}
.logo-dot {
  width: 11px;
  height: 11px;
  border-radius: 3px;
  background: linear-gradient(135deg, #4c6ef5, #6741d9);
}

.aside .el-menu {
  background: transparent;
  border-right: none;
  padding: 4px 10px;
}
.aside .el-menu-item {
  border-radius: 9px;
  margin: 3px 0;
  color: #475569;
}
.aside .el-menu-item.is-active {
  background: linear-gradient(90deg, rgba(76, 110, 245, 0.14), rgba(76, 110, 245, 0.05));
  color: #3b5bdb;
  font-weight: 700;
}
.aside .el-menu-item:hover {
  background: rgba(148, 163, 184, 0.12);
}

.aside-foot {
  margin-top: auto;
  padding: 14px 18px;
  font-size: 12px;
  color: #94a3b8;
}

.main {
  padding: 20px 24px;
}

.el-card {
  border-radius: 12px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: rgba(255, 255, 255, 0.86);
  backdrop-filter: blur(8px);
  box-shadow: 0 4px 20px rgba(30, 41, 82, 0.06);
}
.el-card__header {
  border-bottom: 1px solid rgba(148, 163, 184, 0.14);
}

/* 暗色主题下的渐变与毛玻璃 */
html.dark .app-shell {
  background: linear-gradient(140deg, #101426 0%, #1a2038 30%, #241d33 65%, #12231d 100%);
}
html.dark .aside {
  background: rgba(24, 28, 45, 0.78);
  border-right-color: rgba(255, 255, 255, 0.08);
}
html.dark .el-card {
  background: rgba(28, 32, 50, 0.82);
}
html.dark .aside .el-menu-item {
  color: #a5b0c8;
}
html.dark .aside .el-menu-item.is-active {
  background: linear-gradient(90deg, rgba(124, 149, 248, 0.2), rgba(124, 149, 248, 0.06));
  color: #a5b6fb;
}

/* 窄屏：侧栏折叠为纯图标 */
@media (max-width: 900px) {
  .aside {
    width: 60px !important;
  }
  .logo {
    justify-content: center;
    padding: 18px 0 12px;
    font-size: 0;
  }
  .aside .el-menu-item {
    justify-content: center;
    padding: 0 !important;
  }
  .aside .el-menu-item span {
    display: none;
  }
  .aside-foot {
    display: none;
  }
  .main {
    padding: 12px;
  }
}
</style>


