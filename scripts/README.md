# Clash Center 构建与安装指南

## 编译脚本 (Windows)

在 Windows 系统上，你可以使用 `build.ps1` PowerShell 脚本为 Linux 的不同架构编译 Clash Center：

```powershell
# 在项目根目录下执行
.\scripts\build.ps1
```

该脚本会自动编译以下版本并输出到 `build` 目录：
- `clash-center_linux_amd64`: 适用于 Linux x86_64 架构
- `clash-center_linux_arm64`: 适用于 Linux ARM64 架构 (如 Raspberry Pi 4 64位系统)
- `clash-center_linux_armv7`: 适用于 Linux ARMv7 架构 (如 Raspberry Pi 3/4 32位系统)

## 安装与卸载脚本 (Linux)

在 Linux 系统上，你可以使用 `install.sh` 脚本进行 Clash Center 的安装与卸载：

```bash
# 下载安装脚本
wget https://your-download-server.com/scripts/install.sh
chmod +x install.sh

# 以root权限运行
sudo ./install.sh
```

安装脚本会提供一个交互式菜单，让你选择：
1. 安装 Clash Center
2. 卸载 Clash Center
3. 退出

### 安装功能

选择安装选项后，脚本会自动：
1. 检测你的系统架构
2. 从GitHub下载对应架构的clash-center二进制文件
3. 安装到 `/opt/clash-center/clash-center`
4. 下载默认配置文件 default.yaml
5. 创建配置目录 `/opt/clash-center/configs`
6. 下载并安装对应架构的 Mihomo(Clash.Meta) 核心
7. 下载并安装前端资源文件
8. 创建并启用系统服务
9. 启动服务

### Clash Center 二进制文件

安装过程会自动从GitHub下载适合当前系统架构的Clash Center二进制文件，版本为v1.0.0：
- 对于 amd64 架构：clash-center-linux-amd64
- 对于 arm64 架构：clash-center-linux-arm64
- 对于 armv7 架构：clash-center-linux-armv7

下载地址：`https://github.com/eventlOwOp/clash-center/releases/download/v1.0.0/`

### 默认配置文件

安装脚本会下载默认的Clash配置文件：
- 默认配置文件：`default.yaml`
- 安装位置：`/opt/clash-center/default.yaml`
- 下载地址：`https://github.com/eventlOwOp/clash-center/raw/refs/heads/master/default.yaml`

### Mihomo(Clash.Meta) 核心

安装过程会自动下载适合当前系统架构的 Mihomo(Clash.Meta) 核心，版本为 v1.19.10：
- 对于 amd64 架构：mihomo-linux-amd64-v1.19.10.gz
- 对于 arm64 架构：mihomo-linux-arm64-v1.19.10.gz
- 对于 armv7 架构：mihomo-linux-armv7-v1.19.10.gz

下载后的核心文件会解压并安装到 `/opt/clash-center/clash/clash.meta` 路径。

### 前端资源

安装过程会自动下载并解压前端资源文件：
- 下载 `dist.tar.gz` 前端资源压缩包
- 解压到 `/opt/clash-center/frontend/dist` 目录
- 自动删除下载的压缩文件

下载地址：`https://github.com/eventlOwOp/clash-center/releases/download/v1.0.0/dist.tar.gz`

### 卸载功能

选择卸载选项后，脚本会自动：
1. 停止并禁用系统服务
2. 删除服务文件
3. 删除安装目录及所有文件

## 服务管理

安装完成后，你可以使用以下命令管理服务：
```bash
# 启动服务
sudo systemctl start clash-center

# 停止服务
sudo systemctl stop clash-center

# 重启服务
sudo systemctl restart clash-center

# 查看服务状态
sudo systemctl status clash-center

# 查看服务日志
sudo journalctl -u clash-center
```

## 手动安装

如果你想手动安装，可以按照以下步骤操作：

```bash
# 创建所需目录
sudo mkdir -p /opt/clash-center/clash
sudo mkdir -p /opt/clash-center/configs
sudo mkdir -p /opt/clash-center/frontend/dist

# 下载clash-center二进制文件
ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/' | sed 's/armv7.*/armv7/')
sudo wget -q "https://github.com/eventlOwOp/clash-center/releases/download/v1.0.0/clash-center-linux-${ARCH}" -O /opt/clash-center/clash-center
sudo chmod +x /opt/clash-center/clash-center

# 下载默认配置文件
sudo wget -q "https://github.com/eventlOwOp/clash-center/raw/refs/heads/master/default.yaml" -O /opt/clash-center/default.yaml

# 下载并安装 Mihomo(Clash.Meta)
sudo wget -q "https://github.com/MetaCubeX/mihomo/releases/download/v1.19.10/mihomo-linux-${ARCH}-v1.19.10.gz" -O /tmp/mihomo.gz
sudo gzip -d -c /tmp/mihomo.gz > /opt/clash-center/clash/clash.meta
sudo chmod +x /opt/clash-center/clash/clash.meta
sudo rm -f /tmp/mihomo.gz

# 下载并安装前端资源
sudo wget -q "https://github.com/eventlOwOp/clash-center/releases/download/v1.0.0/dist.tar.gz" -O /tmp/dist.tar.gz
sudo tar -xzf /tmp/dist.tar.gz -C /opt/clash-center/frontend/dist
sudo rm -f /tmp/dist.tar.gz

# 创建系统服务
sudo bash -c 'cat > /etc/systemd/system/clash-center.service << EOF
[Unit]
Description=Clash Center Service
After=network.target

[Service]
Type=simple
User=root
ExecStart=/opt/clash-center/clash-center -H 0.0.0.0 -p 7788
Restart=on-failure
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF'

# 启用并启动服务
sudo systemctl daemon-reload
sudo systemctl enable clash-center.service
sudo systemctl start clash-center.service
```

## 支持的架构

- `amd64`: 64位 x86 架构 (Intel/AMD 处理器)
- `arm64`: 64位 ARM 架构 (如 Raspberry Pi 4 64位系统，部分服务器ARM处理器)
- `armv7`: 32位 ARM v7 架构 (如 Raspberry Pi 3/4 32位系统)

## 注意事项

- 安装脚本需要 root 权限
- 安装后，Clash Center 将作为系统服务运行在 `0.0.0.0:7788`
- 安装完成后可通过浏览器访问 `http://服务器IP:7788` 使用 Clash Center 