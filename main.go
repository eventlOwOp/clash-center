package main

import (
	"log"
	"net/http"
	"strconv"

	"clash-center/internal/api"
	"clash-center/internal/clash"
	"clash-center/internal/config"

	"github.com/spf13/pflag"
)

func init() {
	// 在init函数中不做任何初始化工作，将其移至main函数
}

func main() {
	// 定义命令行参数
	host := pflag.StringP("host", "H", "0.0.0.0", "服务器监听地址")
	port := pflag.IntP("port", "p", 7788, "服务器监听端口")
	clashHome := pflag.StringP("clash-home", "h", clash.ClashHome, "Clash主目录路径")
	configDir := pflag.StringP("config-dir", "c", config.ConfigDir, "配置文件目录路径")
	verbose := pflag.BoolP("verbose", "v", false, "启用详细日志输出")

	// 解析命令行参数
	pflag.Parse()

	// 应用命令行参数
	clash.ClashHome = *clashHome
	config.ConfigDir = *configDir
	isVerbose := *verbose

	log.Printf("Clash主目录: %s\n", clash.ClashHome)
	log.Printf("配置文件目录: %s\n", config.ConfigDir)

	// 加载应用程序配置
	appConfig := config.LoadAppConfig()

	// 如果配置了自动启动并且有上次使用的配置文件，则启动Clash
	if appConfig.AutoStart && appConfig.LastConfig != "" {
		// 记录原始配置文件路径
		config.OriginalConfigName = appConfig.LastConfig

		// 启动Clash
		err := clash.StartClashWithCurrentConfig()
		if err != nil {
			log.Printf("自动启动失败: %v", err)
		} else {
			log.Printf("已自动启动Clash，使用配置文件: %s", appConfig.LastConfig)
		}
	}

	// 设置路由
	router := api.SetupRoutes(isVerbose)

	// 启动服务器
	serverAddr := *host + ":" + strconv.Itoa(*port)
	log.Printf("启动服务器，监听地址 %s...\n", serverAddr)
	err := http.ListenAndServe(serverAddr, router)
	if err != nil {
		log.Fatalf("启动服务器失败: %v\n", err)
	}
}
