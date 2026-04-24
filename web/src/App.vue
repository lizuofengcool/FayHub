<template>
  <div id="app">
    <!-- 动态弥散光晕背景 -->
    <div class="bg-mesh">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
      <div class="blob blob-3"></div>
    </div>
    
    <!-- 路由视图 -->
    <router-view />
  </div>
</template>

<script setup lang="ts">
// 这里可以添加全局逻辑
</script>

<style scoped>
/* 背景弥散光晕动画 */
.bg-mesh {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  z-index: 0;
  overflow: hidden;
  background: #f1f5f9;
}

.blob {
  position: absolute;
  filter: blur(80px);
  z-index: -1;
  opacity: 0.3;
  animation: float 20s infinite ease-in-out alternate;
}

.blob-1 {
  top: -20%; left: -10%;
  width: 40vw; height: 40vw;
  background: radial-gradient(circle, rgba(99,102,241,0.2) 0%, rgba(99,102,241,0) 70%);
  animation-delay: 0s;
}

.blob-2 {
  bottom: -30%; right: -10%;
  width: 50vw; height: 50vw;
  background: radial-gradient(circle, rgba(14,165,233,0.15) 0%, rgba(14,165,233,0) 70%);
  animation-delay: -5s;
}

.blob-3 {
  top: 20%; left: 70%;
  width: 35vw; height: 35vw;
  background: radial-gradient(circle, rgba(168,85,247,0.15) 0%, rgba(168,85,247,0) 70%);
  animation-delay: -10s;
}

@keyframes float {
  0% { transform: translate(0, 0) scale(1); }
  50% { transform: translate(5%, 10%) scale(1.1); }
  100% { transform: translate(-5%, 5%) scale(0.9); }
}
</style>