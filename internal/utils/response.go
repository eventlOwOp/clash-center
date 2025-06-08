package utils

import (
	"encoding/json"
	"net/http"
	"time"

	"clash-center/internal/models"
)

// 发送JSON响应
func SendJSONResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// 发送成功响应
func SendSuccessResponse(w http.ResponseWriter, message string, data ...map[string]any) {
	// 创建一个响应对象
	response := map[string]any{
		"success": true,
	}

	// 添加消息（如果有）
	if message != "" {
		response["message"] = message
	}

	// 如果提供了额外数据，添加到响应中
	if len(data) > 0 {
		for key, value := range data[0] {
			response[key] = value
		}
	}

	SendJSONResponse(w, http.StatusOK, response)
}

// 发送错误响应
func SendErrorResponse(w http.ResponseWriter, statusCode int, errMsg string) {
	response := models.APIResponse{
		Success: false,
		Error:   errMsg,
	}
	SendJSONResponse(w, statusCode, response)
}

// GetTimestamp 获取当前的Unix时间戳
func GetTimestamp() int64 {
	return time.Now().Unix()
}
