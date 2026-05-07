<template>
  <div class="redirect-page">
    <div class="flex flex-col items-center justify-center h-full">
      <el-icon class="text-4xl text-blue-500 mb-4 animate-spin">
        <Loading />
      </el-icon>
      <p class="text-slate-600">页面刷新中...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Loading } from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()

onMounted(() => {
  const targetPath = route.path.replace('/redirect', '')
  
  // 延迟跳转以显示加载效果
  setTimeout(() => {
    router.replace(targetPath)
  }, 300)
})
</script>

<style scoped>
.redirect-page {
  @apply h-full w-full flex items-center justify-center;
}
</style>