package config

import (
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"clash-center/internal/models"

	"gopkg.in/yaml.v3"
)

var (
	// 配置文件目录
	ConfigDir = "./configs"
	// 合并后的配置文件路径
	MergedConfigPath = "./clash/config.yaml"
	// 默认配置文件路径
	DefaultConfigPath = "./default.yaml"
	// 应用程序配置文件路径
	AppConfigPath = "./app_config.json"
	// 当前使用的原始配置文件路径（用于显示）
	OriginalConfigName string
)

// 加载应用程序配置
func LoadAppConfig() models.AppConfig {
	var config models.AppConfig
	config.AutoStart = true // 默认启用自动启动

	// 检查配置文件是否存在
	_, err := os.Stat(AppConfigPath)
	if os.IsNotExist(err) {
		// 配置文件不存在，保存默认配置
		SaveAppConfig(config)
		return config
	}

	// 读取配置文件
	file, err := os.Open(AppConfigPath)
	if err != nil {
		log.Printf("无法打开应用配置文件: %v", err)
		return config
	}
	defer file.Close()

	// 解析JSON
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Printf("解析应用配置文件失败: %v", err)
		return models.AppConfig{AutoStart: true}
	}

	return config
}

// 保存应用程序配置
func SaveAppConfig(config models.AppConfig) error {
	// 创建或打开配置文件
	file, err := os.Create(AppConfigPath)
	if err != nil {
		return fmt.Errorf("创建应用配置文件失败: %v", err)
	}
	defer file.Close()

	// 编码为JSON并保存
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("写入应用配置失败: %v", err)
	}

	return nil
}

// 更新上次使用的配置文件
func UpdateLastConfig(configPath string) {
	config := LoadAppConfig()
	config.LastConfig = configPath
	SaveAppConfig(config)
}

// 获取配置文件列表
func GetConfigFiles() ([]models.ConfigFile, error) {
	files, err := os.ReadDir(ConfigDir)
	if err != nil {
		return nil, err
	}

	var configs []models.ConfigFile
	for _, file := range files {
		if !file.IsDir() && (filepath.Ext(file.Name()) == ".yaml" || filepath.Ext(file.Name()) == ".yml") {
			filePath := file.Name()
			displayName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			configSrc := ""

			// 尝试从YAML文件中读取config_name和config_src字段
			yamlFile, err := os.Open(filepath.Join(ConfigDir, filePath))
			if err == nil {
				defer yamlFile.Close()

				var yamlConfig map[string]any
				decoder := yaml.NewDecoder(yamlFile)
				if err := decoder.Decode(&yamlConfig); err == nil {
					// 检查是否存在config_name字段
					if configName, ok := yamlConfig["config_name"].(string); ok && configName != "" {
						displayName = configName
					}

					// 获取config_src字段的值
					if src, ok := yamlConfig["config_src"].(string); ok {
						configSrc = src
					}
				}
			}

			configs = append(configs, models.ConfigFile{
				Path:        filePath,
				DisplayName: displayName,
				ConfigSrc:   configSrc,
			})
		}
	}

	return configs, nil
}

// 合并配置文件
func MergeConfig(targetConfigPath string) error {
	// 读取默认配置文件
	defaultConfig := make(map[string]any)
	defaultExists := false

	defaultFile, err := os.Open(DefaultConfigPath)
	if err == nil {
		defer defaultFile.Close()
		defaultExists = true

		decoder := yaml.NewDecoder(defaultFile)
		if err := decoder.Decode(&defaultConfig); err != nil {
			return fmt.Errorf("解析默认配置文件失败: %v", err)
		}
	} else {
		log.Printf("默认配置文件不存在，跳过合并: %v", err)
	}

	// 读取目标配置文件
	targetConfig := make(map[string]any)
	targetFile, err := os.Open(filepath.Join(ConfigDir, targetConfigPath))
	if err != nil {
		return fmt.Errorf("打开目标配置文件失败: %v", err)
	}
	defer targetFile.Close()

	decoder := yaml.NewDecoder(targetFile)
	if err := decoder.Decode(&targetConfig); err != nil {
		return fmt.Errorf("解析目标配置文件失败: %v", err)
	}

	// 将目标配置中的设置合并到默认配置中
	finalConfig := make(map[string]any)
	// 先复制目标配置
	maps.Copy(finalConfig, targetConfig)

	// 合并配置（如果默认配置存在）
	if defaultExists {
		// 然后用默认配置覆盖
		maps.Copy(finalConfig, defaultConfig)
		log.Printf("已将默认配置覆盖到目标配置")
	}

	// 写入合并后的配置到文件
	mergedFile, err := os.Create(MergedConfigPath)
	if err != nil {
		return fmt.Errorf("创建配置文件失败: %v", err)
	}
	defer mergedFile.Close()

	encoder := yaml.NewEncoder(mergedFile)
	encoder.SetIndent(2)
	if err := encoder.Encode(finalConfig); err != nil {
		return fmt.Errorf("写入配置失败: %v", err)
	}

	log.Printf("配置已成功合并并写入到: %s", MergedConfigPath)
	return nil
}

// 获取配置信息
func GetConfigInfo(configPath string) (map[string]any, error) {
	// 读取目标配置文件
	configData := make(map[string]any)

	file, err := os.Open(filepath.Join(ConfigDir, configPath))
	if err != nil {
		return nil, fmt.Errorf("打开配置文件失败: %v", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&configData); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return configData, nil
}

// 更新配置文件名称
func UpdateConfigName(configPathName, configName string) error {
	// 读取原YAML文件
	yamlConfig, err := GetConfigInfo(configPathName)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 更新config_name字段
	yamlConfig["config_name"] = configName

	// 重写YAML文件
	file, err := os.Create(filepath.Join(ConfigDir, configPathName))
	if err != nil {
		return fmt.Errorf("无法打开配置文件进行写入: %v", err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	if err := encoder.Encode(yamlConfig); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}
