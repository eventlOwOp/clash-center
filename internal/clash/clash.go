package clash

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"clash-center/internal/config"
)

var (
	// Clash 进程
	ClashCmd *exec.Cmd
	// Clash 运行状态
	IsRunning bool
	// Clash 可执行文件
	ClashPath = "./clash/clash.meta"
	// Clash 主目录
	ClashHome = "./clash"
)

// 启动 Clash 服务
func StartClash() error {
	if IsRunning {
		StopClash()
	}

	log.Printf("启动Clash\n")

	path, err := exec.LookPath(ClashPath)
	if err != nil {
		log.Fatal("Clash.Meta not found")
	}

	// 构建启动命令
	ClashCmd = exec.Command(path, "-d", ClashHome)

	// 设置输出
	ClashCmd.Stdout = os.Stdout
	ClashCmd.Stderr = os.Stderr

	// 启动进程
	err = ClashCmd.Start()
	if err != nil {
		return fmt.Errorf("启动Clash失败: %v", err)
	}

	IsRunning = true

	// 异步等待进程结束
	go func() {
		err := ClashCmd.Wait()
		if err != nil {
			log.Printf("Clash进程结束，错误: %v", err)
		}
		IsRunning = false
	}()

	return nil
}

// 使用当前配置启动Clash
func StartClashWithCurrentConfig() error {
	// 如果没有当前配置但有保存的上次配置，则使用上次配置
	if config.OriginalConfigName == "" {
		appConfig := config.LoadAppConfig()
		if appConfig.LastConfig != "" {
			config.OriginalConfigName = appConfig.LastConfig
		}
	}

	if config.OriginalConfigName == "" {
		return fmt.Errorf("没有选择配置文件")
	}

	// 合并配置文件
	err := config.MergeConfig(config.OriginalConfigName)
	if err != nil {
		return fmt.Errorf("合并配置文件失败: %v", err)
	}

	// 启动Clash
	return StartClash()
}

// 停止 Clash 服务
func StopClash() error {
	if !IsRunning || ClashCmd == nil {
		return nil
	}

	log.Println("停止Clash服务...")

	// 终止进程
	if ClashCmd.Process != nil {
		err := ClashCmd.Process.Kill()
		if err != nil {
			return fmt.Errorf("无法终止Clash进程: %v", err)
		}
	}

	IsRunning = false
	return nil
}
