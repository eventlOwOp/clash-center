package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// 设置API路由
func SetupRoutes(verbose bool) http.Handler {
	r := chi.NewRouter()

	// 中间件
	if verbose {
		r.Use(middleware.Logger)
	}
	r.Use(middleware.Recoverer)
	r.Use(middleware.AllowContentType("application/json", "multipart/form-data"))

	// CORS配置
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(corsMiddleware.Handler)

	// API路由
	r.Route("/api", func(r chi.Router) {
		// 配置文件相关
		r.Get("/configs", HandleGetConfigs)
		r.Post("/switch", HandleSwitchConfig)
		r.Post("/updateconfigname", HandleUpdateConfigName)
		r.Post("/upload", HandleUploadConfig)
		r.Post("/save-config", HandleEditConfigFile)
		r.Get("/config-content", HandleGetConfigContent)
		r.Post("/add-from-url", HandleAddConfigFromURL)
		r.Post("/update-from-url", HandleUpdateConfigFromURL)
		r.Post("/delete-config", HandleDeleteConfig)

		// Clash控制相关
		r.Get("/status", HandleGetStatus)
		r.Post("/start", HandleStartClash)
		r.Post("/stop", HandleStopClash)
		r.Post("/restart", HandleRestartClash)
		r.Get("/controlinfo", HandleGetControlInfo)

		// 应用设置相关
		r.Post("/autostart", HandleToggleAutoStart)
		r.Get("/getautostart", HandleGetAutoStart)
	})

	// 静态文件服务
	fileServer := http.FileServer(http.Dir("./frontend/dist"))
	r.Handle("/*", fileServer)

	return r
}
