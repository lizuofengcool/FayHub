<template>
  <div class="flex min-h-screen items-center justify-center">
    <div class="text-center">
      <h1 class="text-8xl font-bold text-slate-200 mb-4">403</h1>
      <p class="text-xl text-slate-500 mb-8">抱歉，您没有权限访问此页面</p>
      <el-button type="primary" @click="$router.push('/dashboard')">返回首页</el-button>
    </div>
  </div>
</template>
