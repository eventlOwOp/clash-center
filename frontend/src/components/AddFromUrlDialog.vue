<template>
  <div v-if="visible" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white rounded-lg w-1/2 flex flex-col overflow-hidden">
      <div class="flex justify-between items-center p-4 border-b">
        <h3 class="text-lg font-medium text-gray-800">从URL添加配置</h3>
        <button @click="close" class="text-gray-500 hover:text-gray-700">
          <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg">
            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"></path>
          </svg>
        </button>
      </div>
      <div class="p-6">
        <div class="mb-4">
          <label for="config-url" class="block text-sm font-medium text-gray-700 mb-1">订阅URL</label>
          <input
            id="config-url"
            v-model="form.url"
            type="text"
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500"
            placeholder="输入返回base64编码yaml的URL"
          />
        </div>
        <div class="mb-4">
          <label for="config-name" class="block text-sm font-medium text-gray-700 mb-1">配置名称</label>
          <input
            id="config-name"
            v-model="form.configName"
            type="text"
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500"
            placeholder="输入配置显示名称"
          />
        </div>
        <div class="mb-4">
          <label for="file-name" class="block text-sm font-medium text-gray-700 mb-1">文件名称 (可选)</label>
          <input
            id="file-name"
            v-model="form.fileName"
            type="text"
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500"
            placeholder="输入文件名称，不含扩展名"
          />
        </div>
        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-1">获取方式</label>
          <div class="flex space-x-4">
            <label class="inline-flex items-center">
              <input type="radio" v-model="form.fetchMethod" value="backend" class="form-radio text-primary-600">
              <span class="ml-2">后台获取</span>
            </label>
            <label class="inline-flex items-center">
              <input type="radio" v-model="form.fetchMethod" value="frontend" class="form-radio text-primary-600">
              <span class="ml-2">网页获取</span>
            </label>
          </div>
        </div>
        <div v-if="isLoading" class="flex justify-center items-center mb-4">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
          <span class="ml-2 text-primary-600">正在获取配置...</span>
        </div>
      </div>
      <div class="flex justify-end space-x-2 p-4 border-t">
        <button 
          @click="close"
          class="px-4 py-2 border border-gray-300 rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 transition-colors"
        >
          取消
        </button>
        <button 
          @click="submit"
          :disabled="!form.url || !form.configName || isLoading"
          class="px-4 py-2 rounded-md text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          添加
        </button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, reactive, ref, watch } from 'vue'
import { clashApi } from '@/api/clashApi'
import { notificationUtil } from '@/utils/notificationUtil'
import axios from 'axios'

export default defineComponent({
  name: 'AddFromUrlDialog',
  props: {
    visible: {
      type: Boolean,
      default: false
    }
  },
  emits: ['update:visible', 'added'],
  setup(props, { emit }) {
    const form = reactive({
      url: '',
      configName: '',
      fileName: '',
      fetchMethod: 'backend', // 默认后台获取
      rawConfig: null
    })
    
    const isLoading = ref(false)
    
    // 重置表单
    const resetForm = () => {
      form.url = ''
      form.configName = ''
      form.fileName = ''
      form.fetchMethod = 'backend'
      form.rawConfig = null
    }
    
    // 监听可见性变化
    watch(() => props.visible, (isVisible) => {
      if (isVisible) {
        resetForm()
      }
    })
    
    // 前端获取配置内容
    const fetchConfigFromFrontend = async () => {
      if (!form.url) return null
      
      isLoading.value = true
      try {
        const response = await axios.get(form.url, {
          responseType: 'text'
        })
        return response.data
      } catch (error) {
        console.error('获取配置失败:', error)
        notificationUtil.showErrorMessage('获取配置失败', error.message)
        return null
      } finally {
        isLoading.value = false
      }
    }
    
    // 提交表单
    const submit = async () => {
      if (!form.url || !form.configName) return
      
      try {
        let requestData = {
          url: form.url,
          configName: form.configName,
          fileName: form.fileName
        }
        
        // 如果选择前端获取，先获取配置内容
        if (form.fetchMethod === 'frontend') {
          isLoading.value = true
          const rawConfig = await fetchConfigFromFrontend()
          if (!rawConfig) {
            return
          }
          requestData['rawConfig'] = rawConfig
        }
        
        const result = await clashApi.addFromUrl(requestData)
        
        if (result.success) {
          notificationUtil.showSuccessMessage('操作成功', '配置文件添加成功')
          emit('added')
          close()
        }
      } catch (error) {
        console.error('添加配置失败:', error)
        notificationUtil.showErrorMessage('添加配置失败', error.response?.data?.error || error.message)
      } finally {
        isLoading.value = false
      }
    }
    
    // 关闭对话框
    const close = () => {
      emit('update:visible', false)
    }
    
    return {
      form,
      isLoading,
      submit,
      close
    }
  }
})
</script> 