<template>
  <div class="bg-white rounded-lg shadow-md overflow-hidden">
    <div class="p-6">
      <div class="flex justify-between items-center mb-6">
        <h2 class="text-2xl font-bold text-gray-800">配置文件管理</h2>
        
        <div class="flex space-x-2">
          <button @click="$emit('show-url-dialog')" class="px-4 py-2 rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 cursor-pointer transition-colors">
            从URL添加
          </button>
          <label for="file-upload" class="px-4 py-2 rounded-md text-white bg-success-600 hover:bg-success-700 focus:outline-none focus:ring-2 focus:ring-success-500 focus:ring-offset-2 cursor-pointer transition-colors">
            上传配置文件
          </label>
          <input @change="uploadFile" id="file-upload" type="file" class="sr-only" accept=".yaml,.yml" />
        </div>
      </div>
      
      <div class="overflow-hidden border border-gray-200 sm:rounded-lg">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                文件名
              </th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                操作
              </th>
              <th scope="col" class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                使用
              </th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="config in configFiles" :key="config.path" 
                :class="{'bg-primary-50': currentConfig === config.path}">
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                <div v-if="config.editing">
                  <input 
                    v-model="config.editValue" 
                    class="border border-gray-300 px-2 py-1 rounded focus:outline-none focus:ring-2 focus:ring-primary-500"
                    @keyup.enter="saveConfigName(config)" 
                  />
                  <div class="mt-1 space-x-2">
                    <button 
                      @click="saveConfigName(config)"
                      class="text-xs text-success-600 hover:text-success-800"
                    >
                      保存
                    </button>
                    <button 
                      @click="cancelEditName(config)"
                      class="text-xs text-gray-500 hover:text-gray-700"
                    >
                      取消
                    </button>
                  </div>
                </div>
                <div v-else class="flex items-center">
                  <span>{{ config.display_name }}</span>
                  <button 
                    @click="editConfigName(config)"
                    class="ml-2 text-xs text-primary-600 hover:text-primary-800"
                  >
                    编辑
                  </button>
                  <span v-if="config.config_src" class="ml-2 text-xs px-1.5 py-0.5 bg-blue-100 text-blue-800 rounded">
                    订阅
                  </span>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm">
                <div class="flex space-x-2">
                  <button @click="viewConfigContent(config)"
                    class="px-3 py-1 rounded text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-1 transition-colors">
                    编辑
                  </button>
                  <button v-if="config.config_src" @click="updateFromUrl(config)"
                    class="px-3 py-1 rounded text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-amber-500 focus:ring-offset-1 transition-colors">
                    后台更新
                  </button>
                  <button v-if="config.config_src" @click="webUpdateFromUrl(config)"
                    class="px-3 py-1 rounded text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-1 transition-colors">
                    网页更新
                  </button>
                  <button @click="deleteConfig(config)"
                    class="px-3 py-1 rounded text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-1 transition-colors">
                    删除
                  </button>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
                <div class="flex justify-end">
                  <input 
                    type="radio" 
                    :id="`config-${config.path}`" 
                    name="config-selection" 
                    :value="config.path" 
                    :checked="currentConfig === config.path"
                    @change="switchConfig(config)"
                    :disabled="currentConfig === config.path && clashStatus"
                    class="form-radio h-5 w-5 text-primary-600 focus:ring-primary-500 cursor-pointer"
                  />
                </div>
              </td>
            </tr>
            <tr v-if="configFiles.length === 0">
              <td colspan="3" class="px-6 py-4 text-sm text-center text-gray-500">
                没有配置文件
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue'
import { clashApi } from '@/api/clashApi'
import { notificationUtil } from '@/utils/notificationUtil'
import axios from 'axios'

export default defineComponent({
  name: 'ConfigFileList',
  props: {
    configFiles: {
      type: Array,
      required: true
    },
    currentConfig: {
      type: String,
      default: ''
    },
    clashStatus: {
      type: Boolean,
      required: true
    }
  },
  emits: [
    'update:configFiles',
    'update:currentConfig', 
    'update:clashStatus',
    'show-url-dialog',
    'edit-config',
    'refresh-data'
  ],
  setup(props, { emit }) {
    const isLoading = ref(false)
    
    // 上传配置文件
    const uploadFile = async (event) => {
      const file = event.target.files[0]
      if (!file) return
      
      const isYAML = file.name.endsWith('.yml') || file.name.endsWith('.yaml')
      if (!isYAML) {
        notificationUtil.showWarningMessage('文件格式错误', '只能上传YAML文件!')
        return
      }
      
      try {
        const success = await clashApi.uploadConfigFile(file)
        if (success) {
          notificationUtil.showSuccessMessage('操作成功', '配置文件上传成功')
          emit('refresh-data')
        }
        event.target.value = '' // 清空文件选择
      } catch (error) {
        console.error('上传失败:', error)
        notificationUtil.showErrorMessage('上传失败', error.message)
      }
    }
    
    // 编辑配置名称
    const editConfigName = (config) => {
      config.editing = true
      config.editValue = config.display_name
    }
    
    // 保存配置名称
    const saveConfigName = async (config) => {
      if (!config.editValue) return
      try {
        const success = await clashApi.updateConfigName(config.path, config.editValue)
        if (success) {
          notificationUtil.showSuccessMessage('操作成功', '配置文件名称修改成功')
          config.display_name = config.editValue
          config.editing = false
          emit('refresh-data')
        }
      } catch (error) {
        console.error('修改配置文件名称失败:', error)
        notificationUtil.showErrorMessage('修改配置文件名称失败', error.message)
      }
    }
    
    // 取消编辑名称
    const cancelEditName = (config) => {
      config.editing = false
    }
    
    // 查看配置内容
    const viewConfigContent = (config) => {
      emit('edit-config', config)
    }
    
    // 从URL更新配置（后台获取）
    const updateFromUrl = async (config) => {
      try {
        const result = await clashApi.updateFromUrl(config.path)
        if (result.success) {
          notificationUtil.showSuccessMessage('操作成功', '配置文件已更新')
          
          // 如果需要重启Clash，弹出提示
          if (result.data?.needRestart) {
            confirmRestart()
          }
          
          emit('refresh-data')
        }
      } catch (error) {
        console.error('更新配置失败:', error)
        notificationUtil.showErrorMessage('更新配置失败', error.response?.data?.error || error.message)
      }
    }
    
    // 从URL更新配置（网页获取）
    const webUpdateFromUrl = async (config) => {
      if (!config.config_src) {
        notificationUtil.showWarningMessage('无法更新', '该配置文件没有订阅URL源')
        return
      }
      
      isLoading.value = true
      try {
        // 从前端获取配置内容
        const response = await axios.get(config.config_src, {
          responseType: 'text'
        })
        
        // 通过API提交更新
        const result = await clashApi.updateFromUrl(config.path, response.data)
        
        if (result.success) {
          notificationUtil.showSuccessMessage('操作成功', '配置文件已通过网页获取并更新')
          
          // 如果需要重启Clash，弹出提示
          if (result.data?.needRestart) {
            confirmRestart()
          }
          
          emit('refresh-data')
        }
      } catch (error) {
        console.error('网页更新配置失败:', error)
        notificationUtil.showErrorMessage('网页更新失败', error.response?.data?.error || error.message)
      } finally {
        isLoading.value = false
      }
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
          emit('update:clashStatus', true)
        }
      } catch (error) {
        console.error('重启Clash失败:', error)
        notificationUtil.showErrorMessage('重启Clash失败', error.message)
      }
    }
    
    // 删除配置文件
    const deleteConfig = (config) => {
      notificationUtil.confirmAction(
        '确认删除',
        `确定要删除配置文件 "${config.display_name}" 吗？此操作无法撤销。`,
        '是，删除'
      ).then(async (result) => {
        if (result.isConfirmed) {
          try {
            const success = await clashApi.deleteConfig(config.path)
            if (success) {
              notificationUtil.showSuccessMessage('操作成功', '配置文件已删除')
              emit('refresh-data')
            }
          } catch (error) {
            console.error('删除配置文件失败:', error)
            const errorMessage = error.response?.data?.error || error.message
            notificationUtil.showErrorMessage('删除配置文件失败', errorMessage)
          }
        }
      })
    }
    
    // 切换配置
    const switchConfig = async (config) => {
      try {
        const success = await clashApi.switchConfig(config.path)
        if (success) {
          notificationUtil.showSuccessMessage('操作成功', '切换配置文件成功')
          emit('update:currentConfig', config.path)
          emit('update:clashStatus', true)
          emit('refresh-data')
        }
      } catch (error) {
        console.error('切换配置文件失败:', error)
        notificationUtil.showErrorMessage('切换配置文件失败', error.message)
      }
    }
    
    return {
      isLoading,
      uploadFile,
      editConfigName,
      saveConfigName,
      cancelEditName,
      viewConfigContent,
      updateFromUrl,
      webUpdateFromUrl,
      switchConfig,
      deleteConfig
    }
  }
})
</script> 