#!/bin/bash

# 默认服务名称和描述
SERVICE_NAME="clash-center"
DESCRIPTION="Clash 配置管理中心服务"

# 获取脚本所在目录的绝对路径
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
APP_PATH="$SCRIPT_DIR/../clash-center"

# 检查是否以root用户运行
if [ "$EUID" -ne 0 ]; then
  echo "🔴 请使用root权限运行此脚本"
  echo "    例如: sudo $0"
  exit 1
fi

# 检查应用程序是否存在
if [ ! -f "$APP_PATH" ]; then
  echo "🔴 错误: 应用程序 '$APP_PATH' 不存在"
  exit 1
fi

# 确保应用程序有执行权限
chmod +x "$APP_PATH"

SERVICE_FILE="/etc/systemd/system/$SERVICE_NAME.service"

# 创建 systemd 服务文件
echo "🔧 正在创建 systemd 服务文件..."
cat > "$SERVICE_FILE" << EOF
[Unit]
Description=$DESCRIPTION
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=1
User=$(whoami)
ExecStart=$APP_PATH
WorkingDirectory=$SCRIPT_DIR

[Install]
WantedBy=multi-user.target
EOF

echo "✅ 服务文件已创建: $SERVICE_FILE"

# 设置服务文件权限
chmod 644 "$SERVICE_FILE"

# 重新加载 systemd 配置
echo "🔄 正在重新加载 systemd 配置..."
systemctl daemon-reload

# 启用服务
echo "🚀 正在启用服务..."
systemctl enable "$SERVICE_NAME"

# 启动服务
echo "▶️ 正在启动服务..."
systemctl start "$SERVICE_NAME"

# 检查服务状态
echo "📊 服务状态:"
systemctl status "$SERVICE_NAME" --no-pager

echo "✨ 安装完成!"
echo "📝 您可以使用以下命令管理服务:"
echo "   ▶️  启动服务: systemctl start $SERVICE_NAME"
echo "   ⏹️  停止服务: systemctl stop $SERVICE_NAME"
echo "   🔄  重启服务: systemctl restart $SERVICE_NAME"
echo "   📊  查看状态: systemctl status $SERVICE_NAME"
echo "   📜  查看日志: journalctl -u $SERVICE_NAME" 