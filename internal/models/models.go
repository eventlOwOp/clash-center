package models

// ConfigFile 配置文件信息
type ConfigFile struct {
	Path        string `json:"path"`
	DisplayName string `json:"display_name"`
	ConfigSrc   string `json:"config_src"`
}

// AppConfig 应用程序配置
type AppConfig struct {
	LastConfig string `json:"last_config"`
	AutoStart  bool   `json:"auto_start"`
}

// APIResponse API响应通用结构
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
}
