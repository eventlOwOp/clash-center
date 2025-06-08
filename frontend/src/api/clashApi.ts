import axios from 'axios'

interface ConfigFile {
  path: string
  display_name: string
  has_config_src: boolean
  editing?: boolean
  editValue?: string
}

interface UrlFormData {
  url: string
  configName: string
  fileName?: string
  fetchMethod?: string
  rawConfig?: string
}

export const clashApi = {
  // 获取配置列表和状态
  async fetchData() {
    const response = await axios.get('/api/configs')
    return {
      configFiles: (response.data.data || []).map((config: ConfigFile) => ({
        ...config,
        editing: false,
        editValue: ''
      })),
      currentConfig: response.data.current || '',
      clashStatus: response.data.status || false
    }
  },

  // 获取控制信息
  async getControlInfo() {
    const response = await axios.get('/api/controlinfo')
    if (response.data.success) {
      return {
        port: response.data.port,
        secret: response.data.secret
      }
    }
    throw new Error('获取控制信息失败')
  },

  // 获取自动启动设置
  async getAutoStartSetting() {
    const response = await axios.get('/api/getautostart')
    if (response.data.success) {
      return response.data.autoStart
    }
    return false
  },

  // 设置自动启动
  async setAutoStart(value: boolean) {
    const response = await axios.post('/api/autostart', { autoStart: value })
    return response.data.success
  },

  // 启动 Clash
  async startClash() {
    const response = await axios.post('/api/start')
    return response.data.success
  },

  // 停止 Clash
  async stopClash() {
    const response = await axios.post('/api/stop')
    return response.data.success
  },

  // 重启 Clash
  async restartClash() {
    const response = await axios.post('/api/restart')
    return response.data.success
  },

  // 切换配置
  async switchConfig(configPath: string) {
    const response = await axios.post('/api/switch', { configPath })
    return response.data.success
  },

  // 上传配置文件
  async uploadConfigFile(file: File) {
    const formData = new FormData()
    formData.append('file', file)
    
    const response = await axios.post('/api/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
    
    return response.data.success
  },

  // 更新配置名称
  async updateConfigName(configPath: string, configName: string) {
    const response = await axios.post('/api/updateconfigname', {
      configPath,
      configName
    })
    
    return response.data.success
  },

  // 获取配置文件内容
  async getConfigContent(configPath: string) {
    const response = await axios.get(`/api/config-content?path=${encodeURIComponent(configPath)}`)
    if (response.data.success) {
      return response.data.content
    }
    throw new Error('获取配置文件内容失败')
  },

  // 保存配置文件内容
  async saveConfigContent(path: string, content: string) {
    const response = await axios.post('/api/save-config', { path, content })
    return response.data.success
  },

  // 从URL添加配置
  async addFromUrl(formData: UrlFormData) {
    const response = await axios.post('/api/add-from-url', formData)
    return response.data
  },

  // 从URL更新配置
  async updateFromUrl(configPath: string, rawConfig = null) {
    const requestData = rawConfig ? 
      { configPath, rawConfig } : 
      { configPath }
    const response = await axios.post('/api/update-from-url', requestData)
    return response.data
  },

  // 删除配置文件
  async deleteConfig(configPath: string) {
    const response = await axios.post('/api/delete-config', { configPath })
    return response.data.success
  }
} 