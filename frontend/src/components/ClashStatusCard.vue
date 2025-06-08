<template>
  <div class="bg-white rounded-lg shadow-md mb-6 overflow-hidden">
    <div class="p-6">
      <h2 class="text-2xl font-bold text-gray-800 mb-4">Clash 状态</h2>
      
      <div class="flex items-center mb-6">
        <span class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium" 
          :class="clashStatus ? 'bg-success-100 text-success-800' : 'bg-gray-100 text-gray-800'">
          {{ clashStatus ? '运行中' : '已停止' }}
        </span>
        
        <div class="ml-4 space-x-2">
          <button @click="startClash" 
            :disabled="clashStatus || !currentConfig"
            class="px-4 py-2 rounded-md text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
            启动
          </button>
          <button @click="stopClash" 
            :disabled="!clashStatus"
            class="px-4 py-2 rounded-md text-white bg-danger-600 hover:bg-danger-700 focus:outline-none focus:ring-2 focus:ring-danger-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
            停止
          </button>
          <button @click="restartClash" 
            :disabled="!clashStatus"
            class="px-4 py-2 rounded-md text-white bg-warning-600 hover:bg-warning-700 focus:outline-none focus:ring-2 focus:ring-warning-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
            重启
          </button>
        </div>
      </div>
      
      <div v-if="currentConfig" class="mb-4 text-gray-600">
        <p>当前配置文件: {{ configName }}</p>
      </div>
      
      <div class="flex items-center mb-4">
        <label class="relative inline-flex items-center cursor-pointer">
          <input type="checkbox" v-model="autoStartModel" @change="toggleAutoStart" class="sr-only peer">
          <div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"></div>
          <span class="ml-3 text-sm font-medium text-gray-700">自动启动</span>
        </label>
      </div>
      
      <div v-if="clashStatus" class="border-t border-gray-200 pt-4 mt-4">
        <h3 class="text-lg font-medium text-gray-800 mb-3">管理面板</h3>
        <div class="space-x-2">
          <button @click="openDashboard" 
            class="px-4 py-2 rounded-md text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 transition-colors">
            打开控制面板
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, ref, onMounted } from 'vue'
import { clashApi } from '@/api/clashApi'
import { notificationUtil } from '@/utils/notificationUtil'

export default defineComponent({
  name: 'ClashStatusCard',
  props: {
    clashStatus: {
      type: Boolean,
      required: true
    },
    currentConfig: {
      type: String,
      default: ''
    },
    configFiles: {
      type: Array,
      required: true
    }
  },
  emits: ['status-changed', 'control-info-updated'],
  setup(props, { emit }) {
    const autoStartModel = ref(true)
    const controlPort = ref('9090')
    const secret = ref('')
    
    // 获取当前配置名称
    const configName = computed(() => {
      if (!props.currentConfig) return ''
      const configFile = props.configFiles.find(file => file.path === props.currentConfig)
      return configFile ? configFile.display_name : props.currentConfig
    })
    
    // 获取自动启动设置
    const getAutoStartSetting = async () => {
      try {
        autoStartModel.value = await clashApi.getAutoStartSetting()
      } catch (error) {
        console.error('获取自动启动设置失败:', error)
      }
    }
    
    // 切换自动启动
    const toggleAutoStart = async (event) => {
      const value = event.target.checked
      try {
        const success = await clashApi.setAutoStart(value)
        if (success) {
          notificationUtil.showSuccessMessage('自动启动设置', `自动启动已${value ? '启用' : '禁用'}`)
        }
      } catch (error) {
        console.error('设置自动启动失败:', error)
        notificationUtil.showErrorMessage('设置自动启动失败', error.message)
        autoStartModel.value = !value // 恢复之前的值
      }
    }
    
    // 获取控制信息
    const getControlInfo = async () => {
      try {
        const info = await clashApi.getControlInfo()
        controlPort.value = info.port
        secret.value = info.secret
        emit('control-info-updated', info)
      } catch (error) {
        console.error('获取控制信息失败:', error)
      }
    }
    
    // 启动 Clash
    const startClash = async () => {
      if (!props.currentConfig) {
        notificationUtil.showWarningMessage('提示', '请先选择一个配置文件')
        return
      }
      
      try {
        const success = await clashApi.startClash()
        if (success) {
          notificationUtil.showSuccessMessage('操作成功', 'Clash 启动成功')
          emit('status-changed', true)
          getControlInfo()
        }
      } catch (error) {
        console.error('启动Clash失败:', error)
        notificationUtil.showErrorMessage('启动Clash失败', error.message)
      }
    }
    
    // 停止 Clash
    const stopClash = async () => {
      try {
        const success = await clashApi.stopClash()
        if (success) {
          notificationUtil.showSuccessMessage('操作成功', 'Clash 已停止')
          emit('status-changed', false)
        }
      } catch (error) {
        console.error('停止Clash失败:', error)
        notificationUtil.showErrorMessage('停止Clash失败', error.message)
      }
    }
    
    // 重启 Clash
    const restartClash = async () => {
      try {
        const success = await clashApi.restartClash()
        if (success) {
          notificationUtil.showSuccessMessage('操作成功', 'Clash 已重启')
          emit('status-changed', true)
          getControlInfo()
        }
      } catch (error) {
        console.error('重启Clash失败:', error)
        notificationUtil.showErrorMessage('重启Clash失败', error.message)
      }
    }
    
    // 打开控制面板
    const openDashboard = () => {
      const host = window.location.hostname
      const port = controlPort.value
      const sec = secret.value
      
      let url = `http://${host}:${port}/ui/#/setup?http=true&hostname=${host}&port=${port}&secret=${sec}`
      
      window.open(url, '_blank')
    }
    
    onMounted(() => {
      getAutoStartSetting()
      if (props.clashStatus) {
        getControlInfo()
      }
    })
    
    return {
      autoStartModel,
      configName,
      toggleAutoStart,
      startClash,
      stopClash,
      restartClash,
      openDashboard
    }
  }
})
</script> 