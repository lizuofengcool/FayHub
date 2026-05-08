<template>
  <div class="three-level-menu">
    <!-- 一级菜单 -->
    <div
      v-for="level1 in menuTree"
      :key="level1.id"
      class="menu-level-1"
    >
      <!-- 一级菜单标题 -->
      <div
        class="level-1-header"
        :class="{ active: activeLevel1 === level1.id }"
        @click="toggleLevel1(level1)"
      >
        <el-icon class="menu-icon">
          <component :is="iconMap[level1.icon]" v-if="level1.icon && iconMap[level1.icon]" />
          <Folder v-else />
        </el-icon>
        <span class="menu-title">{{ level1.title }}</span>
        <el-icon
          v-if="level1.children && level1.children.length > 0"
          class="arrow-icon"
          :class="{ rotated: activeLevel1 === level1.id }"
        >
          <ArrowRight />
        </el-icon>
      </div>

      <!-- 二级菜单 -->
      <transition name="slide">
        <div
          v-if="level1.children && level1.children.length > 0 && activeLevel1 === level1.id"
          class="level-2-container"
        >
          <div
            v-for="level2 in level1.children"
            :key="level2.id"
            class="menu-level-2"
          >
            <!-- 二级菜单标题 -->
            <div
              class="level-2-header"
              :class="{ active: activeLevel2 === level2.id }"
              @click="toggleLevel2(level2)"
            >
              <el-icon class="menu-icon">
                <component :is="iconMap[level2.icon]" v-if="level2.icon && iconMap[level2.icon]" />
                <Document v-else />
              </el-icon>
              <span class="menu-title">{{ level2.title }}</span>
              <el-icon
                v-if="level2.children && level2.children.length > 0"
                class="arrow-icon"
                :class="{ rotated: activeLevel2 === level2.id }"
              >
                <ArrowRight />
              </el-icon>
            </div>

            <!-- 三级菜单 -->
            <transition name="slide">
              <div
                v-if="level2.children && level2.children.length > 0 && activeLevel2 === level2.id"
                class="level-3-container"
              >
                <div
                  v-for="level3 in level2.children"
                  :key="level3.id"
                  class="level-3-item"
                  :class="{ active: isMenuActive(level3) }"
                  @click="navigateTo(level3)"
                >
                  <span class="dot"></span>
                  <span class="menu-title">{{ level3.title }}</span>
                </div>
              </div>
            </transition>
          </div>
        </div>
      </transition>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  ArrowRight, Monitor, OfficeBuilding, User, Lock, Setting,
  Box, Menu, Connection, Shop, DataAnalysis, Grid, Key, List, Management,
  Tickets, CreditCard, Wallet, Folder, Upload, Document, Link, Tools, Promotion,
  FullScreen
} from '@element-plus/icons-vue'
import type { Menu as MenuType } from '@/api/menu'

interface Props {
  menus: MenuType[]
}

const props = defineProps<Props>()
const router = useRouter()
const route = useRoute()

const activeLevel1 = ref<number | null>(null)
const activeLevel2 = ref<number | null>(null)

const iconMap: Record<string, any> = {
  Monitor, OfficeBuilding, User, Lock, Setting,
  Box, Menu, Connection, Shop, DataAnalysis, Grid, Key, List, Management,
  Tickets, CreditCard, Wallet, Folder, Upload, Document, Link, Tools, Promotion,
  FullScreen
}

// 构建三级菜单树（API已返回树形结构，直接使用）
const menuTree = computed(() => {
  return props.menus
    .filter(menu => menu.status === 1)
    .sort((a, b) => a.sort - b.sort)
    .map(menu => ({
      ...menu,
      children: (menu.children || [])
        .filter((child: MenuType) => child.status === 1)
        .sort((a: MenuType, b: MenuType) => a.sort - b.sort)
        .map((child: MenuType) => ({
          ...child,
          children: (child.children || [])
            .filter((grandchild: MenuType) => grandchild.status === 1)
            .sort((a: MenuType, b: MenuType) => a.sort - b.sort)
        }))
    }))
})

const toggleLevel1 = (menu: MenuType) => {
  if (menu.children && menu.children.length > 0) {
    if (activeLevel1.value === menu.id) {
      activeLevel1.value = null
      activeLevel2.value = null
    } else {
      activeLevel1.value = menu.id
      activeLevel2.value = null
    }
  } else {
    navigateTo(menu)
  }
}

const toggleLevel2 = (menu: MenuType) => {
  if (menu.children && menu.children.length > 0) {
    if (activeLevel2.value === menu.id) {
      activeLevel2.value = null
    } else {
      activeLevel2.value = menu.id
    }
  } else {
    navigateTo(menu)
  }
}

const navigateTo = (menu: MenuType) => {
  if (menu.layout === 'fullscreen' && menu.path) {
    const pluginPath = menu.path.replace(/^\/plugin-apps\//, '')
    router.push(`/plugin-fullscreen/${pluginPath}`)
  } else {
    router.push(menu.path)
  }
}

const isMenuActive = (menu: MenuType): boolean => {
  if (menu.layout === 'fullscreen') {
    const pluginPath = menu.path.replace(/^\/plugin-apps\//, '')
    return route.path === `/plugin-fullscreen/${pluginPath}`
  }
  return route.path === menu.path
}

// 自动展开当前路由对应的菜单
import { watch } from 'vue'
const findActiveMenu = () => {
  for (const l1 of menuTree.value) {
    if (l1.children) {
      for (const l2 of l1.children) {
        if (l2.children) {
          for (const l3 of l2.children) {
            if (isMenuActive(l3)) {
              activeLevel1.value = l1.id
              activeLevel2.value = l2.id
              return
            }
          }
        }
        if (isMenuActive(l2)) {
          activeLevel1.value = l1.id
          activeLevel2.value = null
          return
        }
      }
    }
    if (isMenuActive(l1)) {
      activeLevel1.value = null
      activeLevel2.value = null
      return
    }
  }
}

watch(() => route.path, findActiveMenu, { immediate: true })
</script>

<style scoped>
.three-level-menu {
  padding: 4px 0;
}

/* 一级菜单 */
.level-1-header {
  display: flex;
  align-items: center;
  padding: 0 16px;
  height: 44px;
  cursor: pointer;
  transition: all 0.2s;
  color: var(--sidebar-text, #666);
  position: relative;
  border-radius: 0;
  margin: 2px 8px;
  border-radius: 4px;
}

.level-1-header:hover {
  color: var(--primary, #18a058);
  background: var(--primary-suppl, rgba(24, 160, 88, 0.08));
}

.level-1-header.active {
  color: var(--primary, #18a058);
  background: var(--sidebar-active-bg, rgba(24, 160, 88, 0.12));
  font-weight: 500;
}

.level-1-header.active::before {
  content: '';
  position: absolute;
  left: -8px;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 20px;
  background: var(--primary, #18a058);
  border-radius: 0 2px 2px 0;
}

/* 二级菜单 */
.level-2-container {
  overflow: hidden;
}

.level-2-header {
  display: flex;
  align-items: center;
  padding: 0 16px 0 40px;
  height: 40px;
  cursor: pointer;
  transition: all 0.2s;
  color: var(--sidebar-text, #666);
  margin: 1px 8px;
  border-radius: 4px;
}

.level-2-header:hover {
  color: var(--primary, #18a058);
  background: var(--sidebar-hover, rgba(24, 160, 88, 0.06));
}

.level-2-header.active {
  color: var(--primary, #18a058);
  font-weight: 500;
}

/* 三级菜单 */
.level-3-container {
  overflow: hidden;
}

.level-3-item {
  display: flex;
  align-items: center;
  padding: 0 16px 0 48px;
  height: 36px;
  cursor: pointer;
  transition: all 0.2s;
  color: var(--sidebar-text, #666);
  position: relative;
  margin: 1px 8px;
  border-radius: 4px;
}

.level-3-item:hover {
  color: var(--primary, #18a058);
  background: var(--sidebar-hover, rgba(24, 160, 88, 0.06));
}

.level-3-item.active {
  color: var(--primary, #18a058);
  background: var(--sidebar-active-bg, rgba(24, 160, 88, 0.1));
  font-weight: 500;
}

.level-3-item.active::before {
  content: '';
  position: absolute;
  left: -8px;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 16px;
  background: var(--primary, #18a058);
  border-radius: 0 2px 2px 0;
}

/* 菜单图标 */
.menu-icon {
  margin-right: 10px;
  font-size: 16px;
  flex-shrink: 0;
}

/* 菜单标题 */
.menu-title {
  flex: 1;
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 箭头图标 */
.arrow-icon {
  font-size: 12px;
  transition: transform 0.2s;
  flex-shrink: 0;
  color: var(--text-muted, #999);
}

.arrow-icon.rotated {
  transform: rotate(90deg);
}

/* 小圆点 */
.dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--border-color, #ccc);
  margin-right: 10px;
  flex-shrink: 0;
}

.level-3-item.active .dot {
  background: var(--primary, #18a058);
}

/* 动画 */
.slide-enter-active,
.slide-leave-active {
  transition: all 0.2s ease;
}

.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  max-height: 0;
}

.slide-enter-to,
.slide-leave-from {
  opacity: 1;
  max-height: 500px;
}
</style>