<template>
  <div v-if="visible" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white rounded-lg w-3/4 h-3/4 flex flex-col overflow-hidden">
      <div class="flex justify-between items-center p-4 border-b">
        <h3 class="text-lg font-medium text-gray-800">编辑配置文件: {{ config?.display_name }}</h3>
        <div class="space-x-2">
          <button @click="saveContent" 
            class="px-4 py-2 rounded-md text-white bg-success-600 hover:bg-success-700 focus:outline-none focus:ring-2 focus:ring-success-500 transition-colors">
            保存
          </button>
          <button @click="close" 
            class="px-4 py-2 rounded-md text-white bg-gray-600 hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-gray-500 transition-colors">
            关闭
          </button>
        </div>
      </div>
      <div class="flex-1 h-full">
        <vue-monaco-editor
          v-model:value="content"
          language="yaml"
          theme="vs"
          :options="editorOptions"
          @mount="handleEditorMount"
          class="h-full"
        >
          <template #default>
            加载编辑器中...
          </template>
          <template #failure>
            加载失败，请刷新重试
          </template>
        </vue-monaco-editor>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, watch } from 'vue'
import { VueMonacoEditor } from '@guolao/vue-monaco-editor'
import { clashApi } from '@/api/clashApi'
import { notificationUtil } from '@/utils/notificationUtil'

export default defineComponent({
  name: 'ConfigEditor',
  components: {
    VueMonacoEditor
  },
  props: {
    visible: {
      type: Boolean,
      default: false
    },
    config: {
      type: Object,
      default: null
    },
    currentConfig: {
      type: String,
      default: ''
    }
  },
  emits: ['update:visible', 'saved', 'confirm-restart'],
  setup(props, { emit }) {
    const content = ref('')
    const editorInstance = ref(null)
    const editorOptions = {
      automaticLayout: true,
      scrollBeyondLastLine: false,
      minimap: { enabled: true },
      formatOnType: true,
      formatOnPaste: true
    }
    
    // 监听配置变化，加载配置内容
    watch(() => props.config, async (newConfig) => {
      if (newConfig && props.visible) {
        await loadConfigContent()
      }
    })
    
    // 监听可见性变化，加载配置内容
    watch(() => props.visible, async (isVisible) => {
      if (isVisible && props.config) {
        await loadConfigContent()
      }
    })
    
    // 加载配置内容
    const loadConfigContent = async () => {
      if (!props.config) return
      
      try {
        content.value = await clashApi.getConfigContent(props.config.path)
      } catch (error) {
        console.error('获取配置文件内容失败:', error)
        notificationUtil.showErrorMessage('获取配置文件内容失败', error.message)
        close()
      }
    }
    
    // 编辑器挂载
    const handleEditorMount = (editor) => {
      editorInstance.value = editor
    }
    
    // 保存配置内容
    const saveContent = async () => {
      if (!editorInstance.value || !props.config) return
      
      try {
        const editingPath = props.config.path
        const success = await clashApi.saveConfigContent(editingPath, content.value)
        
        if (success) {
          notificationUtil.showSuccessMessage('操作成功', '配置文件保存成功')
          emit('saved')
          close()
          
          // 如果当前编辑的就是正在使用的配置，询问是否重启
          if (editingPath === props.currentConfig) {
            emit('confirm-restart')
          }
        }
      } catch (error) {
        console.error('保存配置文件失败:', error)
        notificationUtil.showErrorMessage('保存配置文件失败', error.message)
      }
    }
    
    // 关闭编辑器
    const close = () => {
      emit('update:visible', false)
    }
    
    return {
      content,
      editorOptions,
      handleEditorMount,
      saveContent,
      close
    }
  }
})
</script>

<style>
/* 编辑器容器样式 */
.monaco-editor {
  width: 100%;
  height: 100%;
}
</style> 