<template>
  <div class="max-w-4xl mx-auto px-4 py-8">
    <!-- Clash 状态卡片 -->
    <ClashStatusCard 
      :clash-status="clashStatus" 
      :current-config="currentConfig" 
      :config-files="configFiles"
      @status-changed="handleStatusChanged"
      @control-info-updated="handleControlInfoUpdated"
    />

    <!-- 配置文件管理卡片 -->
    <ConfigFileList 
      :config-files="configFiles" 
      :current-config="currentConfig" 
      :clash-status="clashStatus"
      @update:current-config="currentConfig = $event"
      @update:clash-status="clashStatus = $event"
      @show-url-dialog="showUrlDialog = true"
      @edit-config="handleEditConfig"
      @refresh-data="fetchData"
    />

    <!-- 配置文件编辑器对话框 -->
    <ConfigEditor 
      v-model:visible="showEditor" 
      :config="currentEditingConfig"
      :current-config="currentConfig"
      @saved="handleConfigSaved"
      @confirm-restart="confirmRestart"
    />

    <!-- 从URL添加配置对话框 -->
    <AddFromUrlDialog 
      v-model:visible="showUrlDialog"
      @added="fetchData"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted } from 'vue'
import { clashApi } from '@/api/clashApi'
import { notificationUtil } from '@/utils/notificationUtil'
import ClashStatusCard from '@/components/ClashStatusCard.vue'
import ConfigFileList from '@/components/ConfigFileList.vue'
import ConfigEditor from '@/components/ConfigEditor.vue'
import AddFromUrlDialog from '@/components/AddFromUrlDialog.vue'

export default defineComponent({
  name: 'Home',
  components: {
    ClashStatusCard,
    ConfigFileList,
    ConfigEditor,
    AddFromUrlDialog
  },
  setup() {
    const clashStatus = ref(false)
    const configFiles = ref([])
    const currentConfig = ref('')
    const loading = ref(false)
    const controlInfo = ref({ port: '9090', secret: '' })
    const showEditor = ref(false)
    const currentEditingConfig = ref(null)
    const showUrlDialog = ref(false)
    
    // 获取数据
    const fetchData = async () => {
      loading.value = true
      try {
        const data = await clashApi.fetchData()
        configFiles.value = data.configFiles
        currentConfig.value = data.currentConfig
        clashStatus.value = data.clashStatus
        
        console.log('获取数据成功:', {
          configs: configFiles.value,
          current: currentConfig.value,
          status: clashStatus.value
        })
        
        // 如果Clash正在运行，获取控制信息
        if (clashStatus.value) {
          await getControlInfo()
        }
      } catch (error) {
        console.error('获取数据失败:', error)
        notificationUtil.showErrorMessage('获取数据失败', error.message)
      } finally {
        loading.value = false
      }
    }
    
    // 获取控制信息
    const getControlInfo = async () => {
      try {
        const info = await clashApi.getControlInfo()
        controlInfo.value = info
      } catch (error) {
        console.error('获取控制信息失败:', error)
      }
    }
    
    // 处理状态变化
    const handleStatusChanged = (status) => {
      clashStatus.value = status
    }
    
    // 处理控制信息更新
    const handleControlInfoUpdated = (info) => {
      controlInfo.value = info
    }
    
    // 处理编辑配置
    const handleEditConfig = (config) => {
      currentEditingConfig.value = config
      showEditor.value = true
    }
    
    // 处理配置保存
    const handleConfigSaved = () => {
      fetchData()
    }
    
    // 确认重启
    const confirmRestart = () => {
      notificationUtil.confirmAction(
        '配置已更改',
        '配置文件已更改，是否要重启Clash以应用新配置？',
        '是，立即重启'
      ).then((result) => {
        if (result.isConfirmed) {
          restartClash()
        }
      })
    }
    
    // 重启 Clash
    const restartClash = async () => {
      try {
        const success = await clashApi.restartClash()
        if (success) {
          notificationUtil.showSuccessMessage('操作成功', 'Clash 已重启')
          clashStatus.value = true
          await getControlInfo()
        }
      } catch (error) {
        console.error('重启Clash失败:', error)
        notificationUtil.showErrorMessage('重启Clash失败', error.message)
      }
    }
    
    onMounted(() => {
      fetchData()
    })
    
    return {
      clashStatus,
      configFiles,
      currentConfig,
      loading,
      showEditor,
      currentEditingConfig,
      showUrlDialog,
      fetchData,
      handleStatusChanged,
      handleControlInfoUpdated,
      handleEditConfig,
      handleConfigSaved,
      confirmRestart
    }
  }
})
</script>

<style>
.monaco-editor {
  width: 100%;
  height: 100%;
}
</style> 