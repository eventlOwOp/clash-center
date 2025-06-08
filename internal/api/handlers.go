package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"clash-center/internal/clash"
	"clash-center/internal/config"
	"clash-center/internal/converter"
	"clash-center/internal/utils"

	"gopkg.in/yaml.v3"
)

// 处理获取配置文件列表请求
func HandleGetConfigs(w http.ResponseWriter, r *http.Request) {
	configs, err := config.GetConfigFiles()
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("获取配置文件失败: %v", err))
		return
	}

	utils.SendSuccessResponse(w, "获取配置文件成功", map[string]any{
		"data":    configs,
		"current": config.OriginalConfigName,
		"status":  clash.IsRunning,
	})
}

// 处理切换配置文件请求
func HandleSwitchConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendErrorResponse(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	// 解析请求体
	var requestBody struct {
		ConfigPath string `json:"configPath"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("解析请求失败: %v", err))
		return
	}

	// 只获取文件名部分，避免任何路径遍历攻击
	fileName := filepath.Base(requestBody.ConfigPath)
	configPath := filepath.Join(config.ConfigDir, fileName)

	// 检查文件是否存在
	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		utils.SendErrorResponse(w, http.StatusNotFound, "配置文件不存在")
		return
	}

	log.Printf("切换到配置文件: %s\n", fileName)

	// 保存当前配置文件路径
	config.OriginalConfigName = fileName
	config.UpdateLastConfig(fileName)

	// 如果Clash正在运行，重启它
	if clash.IsRunning {
		err := clash.StopClash()
		if err != nil {
			utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("停止Clash失败: %v", err))
			return
		}
	}

	// 启动Clash
	err = clash.StartClashWithCurrentConfig()
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("启动Clash失败: %v", err))
		return
	}

	utils.SendSuccessResponse(w, "已切换配置文件并重启Clash")
}

// 处理启动Clash请求
func HandleStartClash(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendErrorResponse(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	if clash.IsRunning {
		utils.SendSuccessResponse(w, "Clash已经在运行")
		return
	}

	// 启动Clash
	err := clash.StartClashWithCurrentConfig()
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("启动Clash失败: %v", err))
		return
	}

	utils.SendSuccessResponse(w, "Clash已启动")
}

// 处理停止Clash请求
func HandleStopClash(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendErrorResponse(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	if !clash.IsRunning {
		utils.SendSuccessResponse(w, "Clash未运行")
		return
	}

	err := clash.StopClash()
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("停止Clash失败: %v", err))
		return
	}

	utils.SendSuccessResponse(w, "Clash已停止")
}

// 处理重启Clash请求
func HandleRestartClash(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendErrorResponse(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	// 如果Clash正在运行，先停止
	if clash.IsRunning {
		err := clash.StopClash()
		if err != nil {
			utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("停止Clash失败: %v", err))
			return
		}
	}

	// 启动Clash
	err := clash.StartClashWithCurrentConfig()
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("启动Clash失败: %v", err))
		return
	}

	utils.SendSuccessResponse(w, "Clash已重启")
}

// 处理获取Clash状态请求
func HandleGetStatus(w http.ResponseWriter, r *http.Request) {
	utils.SendSuccessResponse(w, "", map[string]any{
		"running": clash.IsRunning,
		"current": config.OriginalConfigName,
	})
}

// 处理上传配置文件请求
func HandleUploadConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendErrorResponse(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	// 解析表单
	err := r.ParseMultipartForm(10 << 20) // 限制上传大小为10MB
	if err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("解析表单失败: %v", err))
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("获取文件失败: %v", err))
		return
	}
	defer file.Close()

	// 检查文件扩展名
	ext := filepath.Ext(handler.Filename)
	if ext != ".yaml" && ext != ".yml" {
		utils.SendErrorResponse(w, http.StatusBadRequest, "仅支持yaml或yml格式的配置文件")
		return
	}

	// 创建目标文件
	dst, err := os.Create(filepath.Join(config.ConfigDir, handler.Filename))
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("创建文件失败: %v", err))
		return
	}
	defer dst.Close()

	// 复制上传的文件内容到目标文件
	_, err = io.Copy(dst, file)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("保存文件失败: %v", err))
		return
	}

	log.Printf("上传配置文件成功: %s\n", handler.Filename)

	utils.SendSuccessResponse(w, fmt.Sprintf("配置文件 %s 上传成功", handler.Filename))
}

// 处理修改自动启动设置请求
func HandleToggleAutoStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendErrorResponse(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	// 解析请求体
	var requestBody struct {
		AutoStart bool `json:"autoStart"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("解析请求失败: %v", err))
		return
	}

	// 更新配置
	appConfig := config.LoadAppConfig()
	appConfig.AutoStart = requestBody.AutoStart
	err = config.SaveAppConfig(appConfig)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("保存配置失败: %v", err))
		return
	}

	utils.SendSuccessResponse(w, fmt.Sprintf("自动启动设置已更新为: %v", requestBody.AutoStart), map[string]any{
		"autoStart": requestBody.AutoStart,
	})
}

// 获取自动启动设置
func HandleGetAutoStart(w http.ResponseWriter, r *http.Request) {
	appConfig := config.LoadAppConfig()

	utils.SendSuccessResponse(w, "", map[string]any{
		"autoStart": appConfig.AutoStart,
	})
}

// 处理获取控制信息请求
func HandleGetControlInfo(w http.ResponseWriter, r *http.Request) {
	var controlPort string = "9090" // 默认端口
	var secret string = ""          // 默认密钥

	// 如果有当前配置文件，尝试从中获取信息
	if config.OriginalConfigName != "" {
		configData, err := config.GetConfigInfo(config.OriginalConfigName)
		if err == nil {
			// 尝试获取external-controller
			if controller, ok := configData["external-controller"].(string); ok {
				// 解析端口
				parts := strings.Split(controller, ":")
				if len(parts) > 1 {
					controlPort = parts[len(parts)-1]
				}
			}

			// 尝试获取secret
			if s, ok := configData["secret"].(string); ok {
				secret = s
			}
		}
	}

	utils.SendSuccessResponse(w, "", map[string]any{
		"port":   controlPort,
		"secret": secret,
	})
}

// 处理修改配置文件名称请求
func HandleUpdateConfigName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendErrorResponse(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	// 解析请求体
	var requestBody struct {
		ConfigPath string `json:"configPath"` // 配置文件路径
		ConfigName string `json:"configName"` // 新的配置名称
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("解析请求失败: %v", err))
		return
	}

	// 只获取文件名部分，避免任何路径遍历攻击
	fileName := filepath.Base(requestBody.ConfigPath)
	configPath := filepath.Join(config.ConfigDir, fileName)

	// 检查文件是否存在
	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		utils.SendErrorResponse(w, http.StatusNotFound, "配置文件不存在")
		return
	}

	log.Printf("更新配置文件名称: %s -> %s\n", fileName, requestBody.ConfigName)

	// 更新配置文件名称
	err = config.UpdateConfigName(fileName, requestBody.ConfigName)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("更新配置文件名称失败: %v", err))
		return
	}

	utils.SendSuccessResponse(w, "配置文件名称已更新")
}

// 处理编辑配置文件请求
func HandleEditConfigFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendErrorResponse(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	// 解析请求体
	var requestBody struct {
		ConfigPath string `json:"path"`    // 配置文件路径
		Content    string `json:"content"` // 编辑后的配置内容(字符串格式)
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("解析请求失败: %v", err))
		return
	}

	// 检查参数
	if requestBody.ConfigPath == "" {
		utils.SendErrorResponse(w, http.StatusBadRequest, "缺少配置文件路径")
		return
	}

	// 只获取文件名部分，避免任何路径遍历攻击
	fileName := filepath.Base(requestBody.ConfigPath)
	configPath := filepath.Join(config.ConfigDir, fileName)

	// 检查文件是否存在
	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		utils.SendErrorResponse(w, http.StatusNotFound, "配置文件不存在")
		return
	}

	log.Printf("开始编辑配置文件: %s\n", fileName)

	// 读取原始配置文件，提取config_开头的字段
	originalContent, err := os.ReadFile(configPath)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("读取原配置文件失败: %v", err))
		return
	}

	// 解析原始YAML
	var originalYamlConfig map[string]any
	if err := yaml.Unmarshal(originalContent, &originalYamlConfig); err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("解析原配置文件失败: %v", err))
		return
	}

	// 提取所有config_开头的配置项
	configPrefixItems := make(map[string]any)
	for key, value := range originalYamlConfig {
		if strings.HasPrefix(key, "config_") {
			configPrefixItems[key] = value
		}
	}

	// 解析新的配置内容
	var newYamlConfig map[string]any
	if requestBody.Content == "" {
		// 如果内容为空，初始化一个空的map
		newYamlConfig = make(map[string]any)
	} else if err := yaml.Unmarshal([]byte(requestBody.Content), &newYamlConfig); err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("解析新配置内容失败: %v", err))
		return
	}

	// 添加原来config_开头的配置项
	maps.Copy(newYamlConfig, configPrefixItems)

	// 转换回YAML字符串
	mergedContent, err := yaml.Marshal(newYamlConfig)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("处理配置内容失败: %v", err))
		return
	}

	// 创建或覆盖配置文件
	file, err := os.Create(configPath)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("创建配置文件失败: %v", err))
		return
	}
	defer file.Close()

	// 写入合并后的内容
	_, err = file.Write(mergedContent)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("写入配置失败: %v", err))
		return
	}

	// 如果正在使用此配置，需要重新加载Clash
	isCurrentConfig := config.OriginalConfigName == fileName
	needRestart := false

	if isCurrentConfig && clash.IsRunning {
		needRestart = true
		// 停止Clash
		err = clash.StopClash()
		if err != nil {
			utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("停止Clash失败: %v", err))
			return
		}
	}

	// 如果需要重启，启动Clash
	if needRestart {
		err = clash.StartClashWithCurrentConfig()
		if err != nil {
			utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("重启Clash失败: %v", err))
			return
		}
		log.Printf("配置已更新并重启Clash")
	} else {
		log.Printf("配置已更新")
	}

	utils.SendSuccessResponse(w, "配置文件已成功更新", map[string]any{
		"restarted": needRestart,
	})
}

// 处理获取配置文件内容请求
func HandleGetConfigContent(w http.ResponseWriter, r *http.Request) {
	// 获取参数
	configPath := r.URL.Query().Get("path")
	if configPath == "" {
		utils.SendErrorResponse(w, http.StatusBadRequest, "缺少配置文件路径")
		return
	}

	// 只获取文件名部分，避免任何路径遍历攻击
	fileName := filepath.Base(configPath)
	configPath = filepath.Join(config.ConfigDir, fileName)

	// 检查文件是否存在
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		utils.SendErrorResponse(w, http.StatusNotFound, "配置文件不存在")
		return
	}

	log.Printf("获取配置文件内容: %s\n", fileName)

	// 读取配置文件内容
	content, err := os.ReadFile(configPath)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("读取配置文件失败: %v", err))
		return
	}

	// 解析YAML文件，删除以config_开头的配置项
	var yamlConfig map[string]any
	if err := yaml.Unmarshal(content, &yamlConfig); err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("解析配置文件失败: %v", err))
		return
	}

	// 删除config_开头的项
	for key := range yamlConfig {
		if strings.HasPrefix(key, "config_") {
			delete(yamlConfig, key)
		}
	}

	// 转换回YAML字符串
	filteredContent, err := yaml.Marshal(yamlConfig)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("处理配置文件失败: %v", err))
		return
	}

	utils.SendSuccessResponse(w, "获取配置文件内容成功", map[string]any{
		"content": string(filteredContent),
	})
}

// ProcessConfigUpdate 处理配置更新的通用逻辑
func ProcessConfigUpdate(fileName, rawConfig, configSrc, configName string) error {
	// 如果有原始配置内容
	if rawConfig != "" {
		// 使用converter直接处理并保存前端提供的配置
		if err := converter.SaveRawConfig([]byte(rawConfig), configSrc, configName, fileName); err != nil {
			return err
		}
	} else {
		// 从URL获取并更新配置
		if err := converter.FetchAndSaveConfig(configSrc, fileName, configName); err != nil {
			return fmt.Errorf("获取配置失败: %v", err)
		}
	}

	return nil
}

// 处理从URL添加配置文件请求
func HandleAddConfigFromURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendErrorResponse(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	// 解析请求体
	var requestBody struct {
		URL        string `json:"url"`
		ConfigName string `json:"configName"`
		FileName   string `json:"fileName"`
		RawConfig  string `json:"rawConfig"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("解析请求失败: %v", err))
		return
	}

	if requestBody.URL == "" {
		utils.SendErrorResponse(w, http.StatusBadRequest, "URL不能为空")
		return
	}

	if requestBody.FileName == "" {
		// 生成文件名
		requestBody.FileName = fmt.Sprintf("config_%d.yaml", utils.GetTimestamp())
	} else if !strings.HasSuffix(requestBody.FileName, ".yaml") && !strings.HasSuffix(requestBody.FileName, ".yml") {
		// 确保文件名有正确的扩展名
		requestBody.FileName = requestBody.FileName + ".yaml"
	}

	// 处理配置更新
	err = ProcessConfigUpdate(requestBody.FileName, requestBody.RawConfig, requestBody.URL, requestBody.ConfigName)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendSuccessResponse(w, "已成功添加配置", map[string]any{
		"path": requestBody.FileName,
		"name": requestBody.ConfigName,
	})
}

// 处理从URL更新配置文件请求
func HandleUpdateConfigFromURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendErrorResponse(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	// 解析请求体
	var requestBody struct {
		ConfigPath string `json:"configPath"`
		RawConfig  string `json:"rawConfig"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("解析请求失败: %v", err))
		return
	}

	if requestBody.ConfigPath == "" {
		utils.SendErrorResponse(w, http.StatusBadRequest, "配置文件路径不能为空")
		return
	}

	// 只获取文件名部分，避免任何路径遍历攻击
	fileName := filepath.Base(requestBody.ConfigPath)

	// 从文件中读取配置
	yamlConfig, err := config.GetConfigInfo(fileName)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("读取配置文件失败: %v", err))
		return
	}

	// 检查是否有config_src字段
	configSrc, ok := yamlConfig["config_src"].(string)
	if !ok || configSrc == "" {
		utils.SendErrorResponse(w, http.StatusBadRequest, "该配置文件没有订阅URL源")
		return
	}

	// 获取当前配置名称
	configName, _ := yamlConfig["config_name"].(string)

	// 处理配置更新
	err = ProcessConfigUpdate(fileName, requestBody.RawConfig, configSrc, configName)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 如果这是当前使用的配置，并且Clash正在运行，提示需要重启
	needRestart := fileName == config.OriginalConfigName && clash.IsRunning

	utils.SendSuccessResponse(w, "配置已更新", map[string]any{
		"needRestart": needRestart,
	})
}

// 处理删除配置文件请求
func HandleDeleteConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendErrorResponse(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	// 解析请求体
	var requestBody struct {
		ConfigPath string `json:"configPath"` // 配置文件路径
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("解析请求失败: %v", err))
		return
	}

	// 检查参数
	if requestBody.ConfigPath == "" {
		utils.SendErrorResponse(w, http.StatusBadRequest, "缺少配置文件路径")
		return
	}

	// 只获取文件名部分，避免任何路径遍历攻击
	fileName := filepath.Base(requestBody.ConfigPath)
	configPath := filepath.Join(config.ConfigDir, fileName)

	// 检查文件是否存在
	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		utils.SendErrorResponse(w, http.StatusNotFound, "配置文件不存在")
		return
	}

	// 检查是否为当前使用的配置文件
	if fileName == config.OriginalConfigName {
		utils.SendErrorResponse(w, http.StatusBadRequest, "无法删除正在使用的配置文件")
		return
	}

	// 删除文件
	err = os.Remove(configPath)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("删除配置文件失败: %v", err))
		return
	}

	log.Printf("配置文件已删除: %s\n", configPath)
	utils.SendSuccessResponse(w, "配置文件已删除")
}
