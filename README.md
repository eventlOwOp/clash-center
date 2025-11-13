# 🚀 Clash Center

<div align="center">
  <h3>A friendly Clash configuration management center</h3>
  <p>Easily manage and switch your Clash configurations via web interface</p>
  
  <p>
    <a href="https://github.com/eventlOwOp/clash-center/blob/master/README.md">English</a> | 
    <a href="https://github.com/eventlOwOp/clash-center/blob/master/README_CN.md">简体中文</a>
  </p>
</div>

<p align="center">
  <img src="https://img.shields.io/github/v/release/eventlOwOp/clash-center" alt="GitHub release" />
  <img src="https://img.shields.io/github/license/eventlOwOp/clash-center" alt="License" />
</p>

<img width="2596" height="1602" alt="image" src="https://github.com/user-attachments/assets/a6f5e1ac-59d8-4d3f-8605-cf3f723a3696" />

## ✨ Features

- 🌐 **Web Management Interface**: Manage Clash configurations through a clean web UI
- 🔄 **Configuration Switching**: Quickly switch between different proxy configurations with one click
- 📈 **Subscription Updates**: Update your proxy subscription links directly through the web interface
- 📊 **Traffic Monitoring**: Real-time proxy traffic statistics and visualization
- 🌍 **Multi-platform Support**: Linux support for multiple architectures (amd64/arm64/armv7)
- 🧰 **Easy Integration**: Run as a system service with auto-start capability

## 📥 One-Click Installation

### Linux (Supports x86_64/ARM64/ARMv7)

Simply copy and paste this command in your terminal:

```bash
curl -fsSL https://github.com/eventlOwOp/clash-center/raw/refs/heads/master/scripts/install.sh | sudo bash
```

Or alternatively:

```bash
wget -O install.sh https://github.com/eventlOwOp/clash-center/raw/refs/heads/master/scripts/install.sh
chmod +x install.sh
sudo ./install.sh
```

The installation script provides an interactive menu to guide you through the installation process.

## 🖥️ System Requirements

- Operating System: Linux (x86_64, ARM64, or ARMv7 architecture)

## 📝 Usage Guide

1. 📌 **Access Web Interface**:
   - Open `http://server-ip:7788` in your browser to access the Clash Center web interface
   
2. 🔄 **Manage Configurations**:
   - Configuration files are stored in the `/opt/clash-center/configs` directory
   - Upload configurations via the web interface or manually place them in this directory
   - Supports various Clash configuration formats

3. 🔄 **Update Subscriptions**:
   - Update your proxy subscription links directly from the web UI
   - Keep your proxy configurations up-to-date with one click

4. 🚦 **Switch Proxy Settings**:
   - Select and apply different configuration files from the web interface
   - Currently used configuration is saved automatically and will be used on next start

5. 📊 **View Traffic Statistics**:
   - Real-time traffic monitoring charts on the home page
   - View total upload/download traffic and current bandwidth usage

## 🛠️ Common Commands

```bash
# Start the service
sudo systemctl start clash-center

# Stop the service
sudo systemctl stop clash-center

# Restart the service
sudo systemctl restart clash-center

# Check service status
sudo systemctl status clash-center

# View service logs
sudo journalctl -u clash-center
```

## 📂 File Structure

```
/opt/clash-center/
├── clash-center       # Main executable file
├── default.yaml       # Default configuration
├── clash/
│   └── clash.meta     # Mihomo(Clash.Meta) core
├── configs/           # User configurations directory
└── frontend/
    └── dist/          # Web frontend files
```

## 📄 Default Configuration File

The `default.yaml` file serves as a base configuration for Clash. It contains essential settings that override the corresponding parameters in your proxy configuration files:

- 🌐 **DNS Settings**: Configure DNS servers and routing
- 🔌 **TUN Mode**: Enable/disable and configure TUN mode
- 🚪 **Ports Configuration**: HTTP/SOCKS5/mixed ports
- 🎮 **API Configuration**: External controller address and port
- 🔓 **LAN Access**: Allow connections from LAN devices
- 🧩 **Other Core Settings**: Mode, log level, etc.

This configuration ensures that critical system settings remain consistent regardless of which proxy configuration you're using. When you switch between different proxy configurations, these base settings will always be applied, while specific proxy settings (servers, rules) will be loaded from your selected configuration file.

You can modify this file to customize how Clash operates on your system without altering your proxy configurations.

## ⚙️ Command Line Arguments

Clash Center supports the following command line arguments:

- `-H, --host`: Set the listen address (default: 0.0.0.0)
- `-p, --port`: Set the listen port (default: 7788)
- `-h, --clash-home`: Set the Clash home directory (default: clash directory)
- `-c, --config-dir`: Set the configuration directory (default: configs directory)
- `-v, --verbose`: Enable verbose logging

## 🔄 Uninstallation

Uninstall using the installation script:

```bash
wget -O install.sh https://github.com/eventlOwOp/clash-center/raw/refs/heads/master/scripts/install.sh
chmod +x install.sh
sudo ./install.sh
# Select the uninstall option
```

Or manually:

```bash
# Stop and disable the service
sudo systemctl stop clash-center
sudo systemctl disable clash-center

# Remove service file and installation directory
sudo rm -f /etc/systemd/system/clash-center.service
sudo rm -rf /opt/clash-center

# Reload systemd configuration
sudo systemctl daemon-reload
```

## 🙋 FAQ

### 🔍 How to change the port number?

Edit the service file to modify the port:

```bash
sudo nano /etc/systemd/system/clash-center.service
# Modify the -p parameter in the ExecStart line
# Then save and restart the service
sudo systemctl daemon-reload
sudo systemctl restart clash-center
```

### 🧭 Why can't I access the web interface?

1. Check if the service is running: `sudo systemctl status clash-center`
2. Make sure the port isn't blocked by a firewall: `sudo ufw allow 7788/tcp` (Ubuntu/Debian)
3. Verify the server IP is correct, try accessing `http://localhost:7788` locally on the server

## 🔄 Update Method

Currently, reinstallation is the recommended way to update:

```bash
# Download and run the installation script
wget -O install.sh https://github.com/eventlOwOp/clash-center/raw/refs/heads/master/scripts/install.sh
chmod +x install.sh
sudo ./install.sh
# Select the install option
```

## 🛡️ License

This project is open source under the [MIT License](LICENSE).

## 🙏 Acknowledgements

- 💖 Thanks to [Mihomo(Clash.Meta)](https://github.com/MetaCubeX/mihomo) for providing an excellent core
- 🌟 Thanks to all developers contributing to the open source community

## 🔗 Related Links

- [Installation Guide](scripts/README.md)
- [Build Guide](BUILD.md)
- [Changelog](CHANGELOG.md)

---

<div align="center">
  <p>⭐ If you like this project, please give it a star! ⭐</p>
  <p>Made with ❤️ by eventlOwOp</p>
</div> 
