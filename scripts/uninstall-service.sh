#!/bin/bash

# 默认服务名称
SERVICE_NAME="clash-center"
SERVICE_FILE="/etc/systemd/system/$SERVICE_NAME.service"

# 检查是否以root用户运行
if [ "$EUID" -ne 0 ]; then
  echo "🔴 请使用root权限运行此脚本"
  echo "    例如: sudo $0"
  exit 1
fi

# 检查服务文件是否存在
if [ ! -f "$SERVICE_FILE" ]; then
  echo "🔴 错误: 服务 '$SERVICE_NAME' 不存在"
  exit 1
fi

# 停止服务
echo "⏹️ 正在停止服务..."
systemctl stop "$SERVICE_NAME"

# 禁用服务
echo "🚫 正在禁用服务..."
systemctl disable "$SERVICE_NAME"

# 删除服务文件
echo "🗑️ 正在删除服务文件..."
rm -f "$SERVICE_FILE"

# 重新加载 systemd 配置
echo "🔄 正在重新加载 systemd 配置..."
systemctl daemon-reload

echo "✅ 卸载完成!"
echo "🎉 服务 '$SERVICE_NAME' 已成功卸载" 