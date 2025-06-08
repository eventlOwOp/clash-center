# 🚀 Clash Center

<div align="center">
  <h3>一个友好的 Clash 配置管理中心</h3>
  <p>通过网页界面轻松管理和切换你的 Clash 配置文件</p>
  
  <p>
    <a href="https://github.com/eventlOwOp/clash-center/blob/master/README.md">English</a> | 
    <a href="https://github.com/eventlOwOp/clash-center/blob/master/README_CN.md">简体中文</a>
  </p>
</div>

<p align="center">
  <img src="https://img.shields.io/github/v/release/eventlOwOp/clash-center" alt="GitHub release" />
  <img src="https://img.shields.io/github/license/eventlOwOp/clash-center" alt="License" />
</p>

## ✨ 功能特点

- 🌐 **网页管理界面**：通过清爽的Web界面管理Clash配置文件
- 🔄 **配置文件切换**：一键快速切换不同的代理配置
- 📈 **订阅更新**：直接通过网页界面更新你的代理订阅链接
- 📊 **流量监控**：实时的代理流量统计和可视化
- 🌍 **多平台支持**：支持多种架构的Linux系统 (amd64/arm64/armv7)
- 🧰 **易于集成**：作为系统服务运行，支持开机自启

## 📥 一键安装

### Linux (支持 x86_64/ARM64/ARMv7)

只需复制以下命令到终端运行：

```bash
curl -fsSL https://github.com/eventlOwOp/clash-center/raw/refs/heads/master/scripts/install.sh | sudo bash
```

或者：

```bash
wget -O install.sh https://github.com/eventlOwOp/clash-center/raw/refs/heads/master/scripts/install.sh
chmod +x install.sh
sudo ./install.sh
```

安装脚本将提供交互式菜单，帮助你完成安装过程。

## 🖥️ 系统要求

- 操作系统：Linux（支持 x86_64, ARM64, ARMv7 架构）

## 📝 使用说明

1. 📌 **访问 Web 界面**：
   - 在浏览器中打开 `http://服务器IP:7788` 进入 Clash Center Web 界面
   
2. 🔄 **管理配置文件**：
   - 配置文件存放在 `/opt/clash-center/configs` 目录中
   - 可以通过 Web 界面上传或手动将配置文件放入该目录
   - 支持多种格式的 Clash 配置文件

3. 🔄 **更新订阅**：
   - 直接通过网页界面更新你的代理订阅链接
   - 一键保持你的代理配置最新

4. 🚦 **切换代理设置**：
   - 在 Web 界面中选择并应用不同的配置文件
   - 当前使用的配置会自动保存，下次启动时会继续使用

5. 📊 **查看流量统计**：
   - 首页提供实时流量监控图表
   - 可以查看总上传/下载流量和当前带宽使用情况

## 🛠️ 常用命令

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

## 📂 文件结构

```
/opt/clash-center/
├── clash-center       # 主程序可执行文件
├── default.yaml       # 默认配置文件
├── clash/
│   └── clash.meta     # Mihomo(Clash.Meta) 核心
├── configs/           # 用户配置文件目录
└── frontend/
    └── dist/          # Web 前端文件
```

## 📄 默认配置文件

`default.yaml` 文件作为 Clash 的基础配置，包含了重要的设置，这些设置会覆盖（override）你的代理配置文件中相应的参数：

- 🌐 **DNS 设置**：配置 DNS 服务器和路由
- 🔌 **TUN 模式**：启用/禁用和配置 TUN 模式
- 🚪 **端口配置**：HTTP/SOCKS5/混合端口
- 🎮 **API 配置**：External controller 地址和端口
- 🔓 **局域网访问**：允许来自局域网设备的连接
- 🧩 **其它核心设置**：模式、日志级别等

这个配置确保了无论你使用哪个代理配置，关键的系统设置都保持一致。当你在不同的代理配置之间切换时，这些基本设置将始终被应用，而特定的代理设置（服务器、规则）则从你选择的配置文件加载。

你可以修改此文件来自定义 Clash 在你系统上的运行方式，而无需修改你的代理配置。

## ⚙️ 命令行参数

Clash Center 支持以下命令行参数：

- `-H, --host`：设置监听地址（默认：0.0.0.0）
- `-p, --port`：设置监听端口（默认：7788）
- `-h, --clash-home`：设置 Clash 主目录（默认：clash目录）
- `-c, --config-dir`：设置配置文件目录（默认：configs目录）
- `-v, --verbose`：启用详细日志

## 🔄 卸载方法

使用安装脚本卸载：

```bash
wget -O install.sh https://github.com/eventlOwOp/clash-center/raw/refs/heads/master/scripts/install.sh
chmod +x install.sh
sudo ./install.sh
# 选择卸载选项
```

或手动卸载：

```bash
# 停止并禁用服务
sudo systemctl stop clash-center
sudo systemctl disable clash-center

# 删除服务文件和安装目录
sudo rm -f /etc/systemd/system/clash-center.service
sudo rm -rf /opt/clash-center

# 重新加载systemd配置
sudo systemctl daemon-reload
```

## 🙋 常见问题

### 🔍 如何更改端口号？

可以编辑服务文件修改端口：

```bash
sudo nano /etc/systemd/system/clash-center.service
# 修改 ExecStart 行的 -p 参数
# 然后保存并重启服务
sudo systemctl daemon-reload
sudo systemctl restart clash-center
```

### 🧭 为什么无法访问 Web 界面？

1. 检查服务是否正在运行：`sudo systemctl status clash-center`
2. 确保端口没有被防火墙阻止：`sudo ufw allow 7788/tcp` (Ubuntu/Debian)
3. 检查服务器IP是否正确，尝试使用 `http://localhost:7788` 在服务器本地访问

## 🔄 更新方法

目前推荐使用重新安装的方式进行更新：

```bash
# 下载并运行安装脚本
wget -O install.sh https://github.com/eventlOwOp/clash-center/raw/refs/heads/master/scripts/install.sh
chmod +x install.sh
sudo ./install.sh
# 选择安装选项
```

## 🛡️ 许可证

本项目基于 [MIT 许可证](LICENSE) 开源。

## 🙏 感谢

- 💖 感谢 [Mihomo(Clash.Meta)](https://github.com/MetaCubeX/mihomo) 项目提供的优秀内核
- 🌟 感谢所有为开源社区做出贡献的开发者

## 🔗 相关链接

- [安装指南](scripts/README.md)
- [构建指南](BUILD.md)
- [更新日志](CHANGELOG.md)

---

<div align="center">
  <p>⭐ 如果你喜欢这个项目，请给它一个星星！ ⭐</p>
  <p>Made with ❤️ by eventlOwOp</p>
</div> 