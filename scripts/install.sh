#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 定义变量
PROJECT_NAME="clash-center"
VERSION="v1.0.0"
INSTALL_DIR="/opt/clash-center"
GITHUB_RELEASE_URL="https://github.com/eventlOwOp/clash-center/releases/download/${VERSION}"
GITHUB_RAW_URL="https://github.com/eventlOwOp/clash-center/raw/refs/heads/master"
SERVICE_FILE="/etc/systemd/system/${PROJECT_NAME}.service"
BINARY="${INSTALL_DIR}/${PROJECT_NAME}"
MIHOMO_VERSION="v1.19.10"
MIHOMO_URL="https://github.com/MetaCubeX/mihomo/releases/download/${MIHOMO_VERSION}"
FRONTEND_DIST_URL="${GITHUB_RELEASE_URL}/dist.tar.gz"
DEFAULT_YAML_URL="${GITHUB_RAW_URL}/default.yaml"

# 打印带颜色的消息
print_msg() {
    echo -e "${2}${1}${NC}"
}

# 检测是否具有root权限
check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_msg "❌ 此脚本需要以root权限运行" "$RED"
        print_msg "🔑 请使用 sudo 重新运行此脚本" "$YELLOW"
        exit 1
    fi
}

# 显示菜单并获取用户选择
show_menu() {
    clear
    print_msg "✨✨✨ Clash Center 安装管理脚本 ✨✨✨" "$BLUE"
    echo ""
    print_msg "1) 🚀 安装 Clash Center" "$GREEN"
    print_msg "2) 🗑️ 卸载 Clash Center" "$RED"
    print_msg "q) 🚪 退出" "$YELLOW"
    echo ""
    read -p "请选择操作 [1/2/q]: " choice
    
    case "$choice" in
        1)
            install_clash_center
            ;;
        2)
            uninstall_clash_center
            ;;
        q|Q)
            print_msg "👋 再见!" "$YELLOW"
            exit 0
            ;;
        *)
            print_msg "❓ 无效的选择，请重新选择" "$RED"
            sleep 1
            show_menu
            ;;
    esac
}

# 检测系统架构
detect_arch() {
    local arch=$(uname -m)
    case "$arch" in
        x86_64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        armv7*)
            echo "armv7"
            ;;
        *)
            print_msg "❌ 不支持的架构: $arch" "$RED"
            exit 1
            ;;
    esac
}

# 创建systemd服务文件
create_service_file() {
    print_msg "📝 创建系统服务文件: ${SERVICE_FILE}" "$YELLOW"
    
    cat > "$SERVICE_FILE" << EOF
[Unit]
Description=Clash Center Service
After=network.target

[Service]
Type=simple
User=root
ExecStart=${BINARY} -H 0.0.0.0 -p 7788
Restart=on-failure
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

    if [[ ! -f "$SERVICE_FILE" ]]; then
        print_msg "❌ 创建服务文件失败" "$RED"
        exit 1
    fi
}

# 启用并启动服务
enable_service() {
    print_msg "🔄 正在启用 ${PROJECT_NAME} 服务..." "$YELLOW"
    systemctl daemon-reload
    systemctl enable "${PROJECT_NAME}.service"
    
    print_msg "▶️ 正在启动 ${PROJECT_NAME} 服务..." "$YELLOW"
    systemctl start "${PROJECT_NAME}.service"
    
    # 检查服务状态
    if systemctl is-active --quiet "${PROJECT_NAME}.service"; then
        print_msg "✅ ${PROJECT_NAME} 服务已成功启动!" "$GREEN"
    else
        print_msg "❌ ${PROJECT_NAME} 服务启动失败，请检查日志: journalctl -u ${PROJECT_NAME}.service" "$RED"
    fi
}

# 下载并安装mihomo
install_mihomo() {
    local arch=$1
    local mihomo_file="mihomo-linux-${arch}-${MIHOMO_VERSION}.gz"
    local mihomo_url="${MIHOMO_URL}/${mihomo_file}"
    local clash_dir="${INSTALL_DIR}/clash"
    local clash_meta="${clash_dir}/clash.meta"
    
    print_msg "📁 创建Clash目录: ${clash_dir}" "$YELLOW"
    mkdir -p "${clash_dir}"
    
    print_msg "📥 正在下载Mihomo(Clash.Meta) ${MIHOMO_VERSION} 版本..." "$YELLOW"
    
    # 下载压缩文件到临时文件
    if command -v wget > /dev/null; then
        wget -q "${mihomo_url}" -O "/tmp/${mihomo_file}"
    elif command -v curl > /dev/null; then
        curl -s -L "${mihomo_url}" -o "/tmp/${mihomo_file}"
    else
        print_msg "❌ 错误: 需要安装 wget 或 curl" "$RED"
        exit 1
    fi
    
    # 检查下载是否成功
    if [[ ! -f "/tmp/${mihomo_file}" ]]; then
        print_msg "❌ 下载失败: ${mihomo_url}" "$RED"
        exit 1
    fi
    
    # 解压缩文件
    print_msg "📦 正在解压Mihomo..." "$YELLOW"
    gzip -d -c "/tmp/${mihomo_file}" > "${clash_meta}"
    
    # 删除临时文件
    rm -f "/tmp/${mihomo_file}"
    
    # 添加执行权限
    chmod +x "${clash_meta}"
    
    # 验证安装
    if [[ -x "${clash_meta}" ]]; then
        print_msg "✅ Mihomo(Clash.Meta)安装成功!" "$GREEN"
    else
        print_msg "❌ Mihomo(Clash.Meta)安装失败" "$RED"
        exit 1
    fi
}

# 下载并安装前端资源
install_frontend() {
    local frontend_dir="${INSTALL_DIR}/frontend/dist"
    local dist_file="/tmp/dist.tar.gz"
    
    print_msg "📁 创建前端目录: ${frontend_dir}" "$YELLOW"
    mkdir -p "${frontend_dir}"
    
    print_msg "📥 正在下载前端资源..." "$YELLOW"
    
    # 下载前端资源压缩包
    if command -v wget > /dev/null; then
        wget -q "${FRONTEND_DIST_URL}" -O "${dist_file}"
    elif command -v curl > /dev/null; then
        curl -s -L "${FRONTEND_DIST_URL}" -o "${dist_file}"
    else
        print_msg "❌ 错误: 需要安装 wget 或 curl" "$RED"
        exit 1
    fi
    
    # 检查下载是否成功
    if [[ ! -f "${dist_file}" ]]; then
        print_msg "❌ 下载失败: ${FRONTEND_DIST_URL}" "$RED"
        exit 1
    fi
    
    # 解压缩前端资源
    print_msg "📦 正在解压前端资源..." "$YELLOW"
    tar -xzf "${dist_file}" -C "${frontend_dir}"
    
    # 删除压缩文件
    rm -f "${dist_file}"
    
    # 验证安装
    if [[ -d "${frontend_dir}" ]] && [[ "$(ls -A ${frontend_dir})" ]]; then
        print_msg "✅ 前端资源安装成功!" "$GREEN"
    else
        print_msg "❌ 前端资源安装失败" "$RED"
        exit 1
    fi
}

# 下载default.yaml配置文件
download_default_yaml() {
    print_msg "📥 正在下载默认配置文件..." "$YELLOW"
    
    # 下载配置文件
    if command -v wget > /dev/null; then
        wget -q "${DEFAULT_YAML_URL}" -O "${INSTALL_DIR}/default.yaml"
    elif command -v curl > /dev/null; then
        curl -s -L "${DEFAULT_YAML_URL}" -o "${INSTALL_DIR}/default.yaml"
    else
        print_msg "❌ 错误: 需要安装 wget 或 curl" "$RED"
        exit 1
    fi
    
    # 检查下载是否成功
    if [[ ! -f "${INSTALL_DIR}/default.yaml" ]]; then
        print_msg "❌ 下载配置文件失败: ${DEFAULT_YAML_URL}" "$RED"
        exit 1
    fi
    
    print_msg "✅ 默认配置文件下载成功!" "$GREEN"
}

# 安装 Clash Center
install_clash_center() {
    print_msg "🚀 开始安装 ${PROJECT_NAME}..." "$BLUE"
    
    # 检测架构
    ARCH=$(detect_arch)
    print_msg "🖥️ 检测到系统架构: $ARCH" "$YELLOW"
    
    # 构建文件名和下载URL
    BINARY_NAME="${PROJECT_NAME}-linux-${ARCH}"
    DOWNLOAD_LINK="${GITHUB_RELEASE_URL}/${BINARY_NAME}"
    
    # 创建安装目录
    print_msg "📁 创建安装目录: ${INSTALL_DIR}" "$YELLOW"
    if [[ ! -d "$INSTALL_DIR" ]]; then
        mkdir -p "$INSTALL_DIR"
    fi
    
    # 下载文件
    print_msg "📥 正在下载 ${BINARY_NAME}..." "$YELLOW"
    if command -v wget > /dev/null; then
        wget -q "$DOWNLOAD_LINK" -O "${BINARY}"
    elif command -v curl > /dev/null; then
        curl -s -L "$DOWNLOAD_LINK" -o "${BINARY}"
    else
        print_msg "❌ 错误: 需要安装 wget 或 curl" "$RED"
        exit 1
    fi
    
    # 检查下载是否成功
    if [[ ! -f "${BINARY}" ]]; then
        print_msg "❌ 下载失败: ${DOWNLOAD_LINK}" "$RED"
        exit 1
    fi
    
    # 添加执行权限
    print_msg "🔒 添加执行权限..." "$YELLOW"
    chmod +x "${BINARY}"
    
    # 验证安装
    if [[ -x "${BINARY}" ]]; then
        print_msg "✅ ${PROJECT_NAME} 安装成功!" "$GREEN"
    else
        print_msg "❌ 安装失败" "$RED"
        exit 1
    fi
    
    # 创建文件夹
    mkdir -p "$INSTALL_DIR/configs"
    
    # 下载默认配置文件
    download_default_yaml
    
    # 下载并安装mihomo
    install_mihomo "$ARCH"
    
    # 下载并安装前端资源
    install_frontend
    
    # 创建并启动服务
    create_service_file
    enable_service
    
    print_msg "🎉 === 安装完成! === 🎉" "$GREEN"
    print_msg "🔧 可以通过以下命令管理服务:" "$BLUE"
    print_msg "  ▶️ 启动: systemctl start ${PROJECT_NAME}" "$YELLOW"
    print_msg "  ⏹️ 停止: systemctl stop ${PROJECT_NAME}" "$YELLOW"
    print_msg "  🔄 重启: systemctl restart ${PROJECT_NAME}" "$YELLOW"
    print_msg "  📊 状态: systemctl status ${PROJECT_NAME}" "$YELLOW"
    print_msg "  📜 查看日志: journalctl -u ${PROJECT_NAME}" "$YELLOW"
    print_msg "🌐 您可以通过访问 http://服务器IP:7788 来使用 Clash Center" "$GREEN"
}

# 卸载 Clash Center
uninstall_clash_center() {
    print_msg "🗑️ 开始卸载 ${PROJECT_NAME}..." "$BLUE"
    
    # 检查服务是否存在
    if systemctl list-unit-files | grep -q "${PROJECT_NAME}.service"; then
        print_msg "⏹️ 停止服务..." "$YELLOW"
        systemctl stop "${PROJECT_NAME}.service"
        
        print_msg "🔄 禁用服务..." "$YELLOW"
        systemctl disable "${PROJECT_NAME}.service"
        
        print_msg "🗑️ 删除服务文件..." "$YELLOW"
        rm -f "${SERVICE_FILE}"
        
        systemctl daemon-reload
    else
        print_msg "ℹ️ 服务不存在，跳过服务卸载步骤" "$YELLOW"
    fi
    
    # 删除安装目录
    if [[ -d "${INSTALL_DIR}" ]]; then
        print_msg "🗑️ 删除安装目录: ${INSTALL_DIR}" "$YELLOW"
        rm -rf "${INSTALL_DIR}"
    else
        print_msg "ℹ️ 安装目录不存在，跳过删除步骤" "$YELLOW"
    fi
    
    print_msg "🎉 === 卸载完成! === 🎉" "$GREEN"
}

# 主函数
main() {
    # 检查root权限
    check_root
    
    # 显示菜单
    show_menu
}

# 执行主函数
main 