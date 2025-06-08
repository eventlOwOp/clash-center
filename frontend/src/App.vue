<template>
  <header class="bg-primary-700 text-white p-4 shadow-md">
    <div class="max-w-7xl mx-auto">
      <h1 class="text-2xl font-bold">Clash Center</h1>
    </div>
  </header>
  
  <main class="py-6">
    <router-view />
  </main>
</template>

<script>
export default {
  name: 'App',
  components: {}
}
</script>

<style>
body {
  margin: 0;
  padding: 0;
  font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif;
}

.main-container {
  min-height: 100vh;
}

.el-header {
  background-color: #409EFF;
  color: white;
  line-height: 60px;
  text-align: center;
}

.el-footer {
  background-color: #f5f5f5;
  color: #666;
  text-align: center;
  line-height: 60px;
}

.el-main {
  padding: 20px;
  background-color: #f0f2f5;
}
</style> 