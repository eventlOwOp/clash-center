package converter

import (
	"clash-center/internal/config"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseAndEnrichConfig 解析配置内容并添加元数据
func ParseAndEnrichConfig(content []byte, url string, configName string) ([]byte, error) {
	// 尝试Base64解码（大多数订阅都是Base64编码的）
	contentStr := string(content)
	decoded, err := base64.StdEncoding.DecodeString(contentStr)
	// 如果解码失败，使用原始内容
	if err != nil {
		log.Printf("Base64解码失败，尝试直接解析内容")
		decoded = content
	}

	// 首先尝试解析为YAML
	var yamlConfig map[string]any
	err = yaml.Unmarshal(decoded, &yamlConfig)
	if err != nil {
		// 不是YAML格式，可能是节点URL列表，尝试解析为订阅内容
		log.Printf("解析为YAML失败，尝试解析为节点URL列表")
		yamlConfig, err = ParseSubscriptionContent(decoded)
		if err != nil {
			return nil, fmt.Errorf("解析订阅内容失败: %v", err)
		}
	}

	// 添加配置来源和名称
	yamlConfig["config_src"] = url
	if configName != "" {
		yamlConfig["config_name"] = configName
	}

	// 将修改后的配置编码回YAML
	modifiedYAML, err := yaml.Marshal(yamlConfig)
	if err != nil {
		return nil, fmt.Errorf("编码YAML失败: %v", err)
	}

	return modifiedYAML, nil
}

// SaveConfigToFile 将处理后的配置内容保存到文件
func SaveConfigToFile(configContent []byte, filePathName string) error {
	// 确保目录存在
	os.MkdirAll(filepath.Dir(filePathName), 0755)

	// 写入文件
	err := os.WriteFile(filepath.Join(config.ConfigDir, filePathName), configContent, 0644)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// SaveRawConfig 处理并保存原始配置内容
func SaveRawConfig(rawConfig []byte, configSrc string, configName string, filePathName string) error {
	// 解析和丰富配置内容
	modifiedYAML, err := ParseAndEnrichConfig(rawConfig, configSrc, configName)
	if err != nil {
		return fmt.Errorf("处理配置内容失败: %v", err)
	}

	// 保存到文件
	return SaveConfigToFile(modifiedYAML, filePathName)
}

// FetchAndSaveConfig 从URL获取配置并保存到文件
func FetchAndSaveConfig(url string, filePathName string, configName string) error {
	// 发送HTTP请求获取配置
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("请求URL失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求URL返回错误状态码: %d", resp.StatusCode)
	}

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应内容失败: %v", err)
	}

	// 解析和丰富配置内容
	modifiedYAML, err := ParseAndEnrichConfig(body, url, configName)
	if err != nil {
		return err
	}

	// 保存到文件
	return SaveConfigToFile(modifiedYAML, filePathName)
}

// ParseSubscriptionContent 解析订阅内容为Clash配置
func ParseSubscriptionContent(content []byte) (map[string]any, error) {
	// 按行分割
	lines := strings.Split(string(content), "\n")

	// 存储解析出的代理
	var proxies []map[string]any
	// 用于确保名称唯一性的映射
	names := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var proxy map[string]any

		// 根据协议类型解析
		if strings.HasPrefix(line, "vmess://") {
			proxy = ParseVmessURL(line)
		} else if strings.HasPrefix(line, "ss://") {
			proxy = ParseSSURL(line)
		} else if strings.HasPrefix(line, "trojan://") {
			proxy = ParseTrojanURL(line)
		} else if strings.HasPrefix(line, "vless://") {
			proxy = ParseVlessURL(line)
		} else if strings.HasPrefix(line, "hysteria2://") || strings.HasPrefix(line, "hy2://") {
			proxy = ParseHysteria2URL(line)
		} else if strings.HasPrefix(line, "hysteria://") {
			proxy = ParseHysteriaURL(line)
		} else if strings.HasPrefix(line, "tuic://") {
			proxy = ParseTuicURL(line)
		} else if strings.HasPrefix(line, "ssr://") {
			proxy = ParseSSRURL(line)
		}

		if len(proxy) > 0 {
			// 确保名称唯一
			if name, ok := proxy["name"].(string); ok {
				proxy["name"] = UniqueName(names, name)
			}
			proxies = append(proxies, proxy)
		}
	}

	// 生成Clash配置
	if len(proxies) > 0 {
		return GenerateClashConfig(proxies), nil
	}

	return nil, fmt.Errorf("未能解析任何有效的代理节点")
}

// ParseVmessURL 解析VMess URL
func ParseVmessURL(vmessURL string) map[string]any {
	// 移除前缀
	encoded := vmessURL[8:]

	// 尝试解码Base64
	decoded, err := Base64RawStdDecode(encoded)
	if err != nil {
		// 可能是Xray VMessAEAD分享链接格式
		log.Printf("VMess标准格式解码失败，尝试解析为Xray VMessAEAD格式")

		u, err := url.Parse(vmessURL)
		if err != nil {
			log.Printf("VMess URL解析失败: %v", err)
			return nil
		}

		if u.Scheme != "vmess" {
			return nil
		}

		// 解析Xray VMessAEAD格式
		uuid := u.User.String()
		server := u.Hostname()
		port := u.Port()

		if server == "" || port == "" || uuid == "" {
			log.Printf("VMess URL缺少必要参数")
			return nil
		}

		portInt, err := strconv.Atoi(port)
		if err != nil {
			log.Printf("VMess端口号格式错误: %v", err)
			return nil
		}

		name := u.Fragment
		if name == "" {
			name = "VMess节点"
		}

		// 解析查询参数
		query := u.Query()

		vmess := map[string]any{
			"name":    name,
			"type":    "vmess",
			"server":  server,
			"port":    portInt,
			"uuid":    uuid,
			"alterId": 0,
			"cipher":  "auto",
			"udp":     true,
			"xudp":    true,
		}

		// 加密方式
		encryption := query.Get("encryption")
		if encryption != "" {
			vmess["cipher"] = encryption
		}

		// 处理网络设置
		network := query.Get("type")
		if network == "" {
			network = "tcp"
		}
		vmess["network"] = network

		// TLS设置
		security := query.Get("security")
		if security == "tls" || security == "xtls" {
			vmess["tls"] = true

			// SNI设置
			sni := query.Get("sni")
			if sni != "" {
				vmess["servername"] = sni
			}

			// ALPN设置
			alpn := query.Get("alpn")
			if alpn != "" {
				vmess["alpn"] = strings.Split(alpn, ",")
			}
		}

		// 处理各种网络特定设置
		if network == "ws" {
			wsOpts := map[string]any{}

			// 路径设置
			path := query.Get("path")
			if path != "" {
				wsOpts["path"] = path
			}

			// 主机头设置
			host := query.Get("host")
			if host != "" {
				wsOpts["headers"] = map[string]any{
					"Host": host,
				}
			}

			vmess["ws-opts"] = wsOpts
		} else if network == "h2" || network == "http" {
			h2Opts := map[string]any{}

			// 路径设置
			path := query.Get("path")
			if path != "" {
				h2Opts["path"] = path
			}

			// 主机头设置
			host := query.Get("host")
			if host != "" {
				h2Opts["host"] = []string{host}
			}

			if network == "h2" {
				vmess["h2-opts"] = h2Opts
			} else {
				vmess["http-opts"] = h2Opts
			}
		} else if network == "grpc" {
			grpcOpts := map[string]any{}

			// 服务名称设置
			serviceName := query.Get("serviceName")
			if serviceName != "" {
				grpcOpts["grpc-service-name"] = serviceName
			}

			vmess["grpc-opts"] = grpcOpts
		}

		return vmess
	}

	// 标准VMess格式，解析JSON
	var config map[string]any
	err = json.Unmarshal([]byte(decoded), &config)
	if err != nil {
		log.Printf("VMess配置解析失败: %v", err)
		return nil
	}

	// 转换为Clash格式
	proxy := map[string]any{
		"name":    config["ps"],
		"type":    "vmess",
		"server":  config["add"],
		"port":    config["port"],
		"uuid":    config["id"],
		"alterId": config["aid"],
		"udp":     true,
		"xudp":    true,
	}

	// 加密方式
	cipher := GetStringOrDefault(config["scy"], "auto")
	proxy["cipher"] = cipher

	// 处理网络设置
	if network, ok := config["net"].(string); ok {
		proxy["network"] = network

		// WebSocket设置
		if network == "ws" {
			wsOpts := map[string]any{}
			if path, ok := config["path"].(string); ok {
				wsOpts["path"] = path
			}
			if host, ok := config["host"].(string); ok {
				wsOpts["headers"] = map[string]any{
					"Host": host,
				}
			}
			proxy["ws-opts"] = wsOpts
		} else if network == "h2" {
			h2Opts := map[string]any{}
			if path, ok := config["path"].(string); ok {
				h2Opts["path"] = path
			}
			if host, ok := config["host"].(string); ok {
				h2Opts["host"] = []string{host}
			}
			proxy["h2-opts"] = h2Opts
		} else if network == "http" {
			httpOpts := map[string]any{}
			if path, ok := config["path"].(string); ok {
				httpOpts["path"] = path
			}
			if host, ok := config["host"].(string); ok {
				httpOpts["headers"] = map[string]any{
					"Host": host,
				}
			}
			proxy["http-opts"] = httpOpts
		} else if network == "grpc" {
			grpcOpts := map[string]any{}
			if path, ok := config["path"].(string); ok {
				grpcOpts["grpc-service-name"] = path
			}
			proxy["grpc-opts"] = grpcOpts
		}
	}

	// TLS设置
	if tls, ok := config["tls"].(string); ok && tls == "tls" {
		proxy["tls"] = true

		// SNI设置
		if sni, ok := config["sni"].(string); ok && sni != "" {
			proxy["servername"] = sni
		}

		// ALPN设置
		if alpn, ok := config["alpn"].(string); ok && alpn != "" {
			proxy["alpn"] = strings.Split(alpn, ",")
		}
	}

	return proxy
}

// ParseSSURL 解析Shadowsocks URL
func ParseSSURL(ssURL string) map[string]any {
	// 移除前缀
	content := ssURL[5:]

	// 分离名称部分
	var name string
	if idx := strings.LastIndex(content, "#"); idx > 0 {
		name = content[idx+1:]
		name, _ = url.QueryUnescape(name)
		content = content[:idx]
	} else {
		name = "SS节点"
	}

	// 处理Base64编码的内容
	var server, port, method, password string

	if strings.Contains(content, "@") {
		// 新格式：method:password@server:port
		parts := strings.SplitN(content, "@", 2)
		authPart := parts[0]
		serverPart := parts[1]

		// 解码认证部分
		decodedAuth, err := base64.StdEncoding.DecodeString(authPart)
		if err == nil {
			authStr := string(decodedAuth)
			if idx := strings.Index(authStr, ":"); idx > 0 {
				method = authStr[:idx]
				password = authStr[idx+1:]
			}
		} else if idx := strings.Index(authPart, ":"); idx > 0 {
			method = authPart[:idx]
			password = authPart[idx+1:]
		}

		// 解析服务器部分
		if idx := strings.LastIndex(serverPart, ":"); idx > 0 {
			server = serverPart[:idx]
			port = serverPart[idx+1:]
		}
	} else {
		// 旧格式：整个内容是Base64编码
		decoded, err := Base64RawStdDecode(content)
		if err != nil {
			log.Printf("SS URL解码失败: %v", err)
			return nil
		}

		decodedStr := decoded
		if idx := strings.LastIndex(decodedStr, "@"); idx > 0 {
			methodPassPart := decodedStr[:idx]
			serverPortPart := decodedStr[idx+1:]

			if idx := strings.Index(methodPassPart, ":"); idx > 0 {
				method = methodPassPart[:idx]
				password = methodPassPart[idx+1:]
			}

			if idx := strings.LastIndex(serverPortPart, ":"); idx > 0 {
				server = serverPortPart[:idx]
				port = serverPortPart[idx+1:]
			}
		}
	}

	// 验证所有必要字段
	if server == "" || port == "" || method == "" || password == "" {
		log.Printf("SS URL格式无效或不完整")
		return nil
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Printf("SS端口号格式错误: %v", err)
		return nil
	}

	ss := map[string]any{
		"name":     name,
		"type":     "ss",
		"server":   server,
		"port":     portInt,
		"cipher":   method,
		"password": password,
		"udp":      true,
	}

	// 解析查询参数
	if idx := strings.Index(content, "?"); idx > 0 {
		queryStr := content[idx+1:]
		query, err := url.ParseQuery(queryStr)
		if err == nil {
			// 处理插件
			plugin := query.Get("plugin")
			if strings.Contains(plugin, "obfs") {
				pluginOpts := query.Get("plugin-opts")
				if pluginOpts == "" {
					// 尝试解析老格式的插件参数
					obfsParams := strings.Split(plugin, ";")
					if len(obfsParams) >= 3 {
						ss["plugin"] = "obfs"

						var mode, host string
						for _, param := range obfsParams {
							if strings.HasPrefix(param, "obfs=") {
								mode = param[5:]
							} else if strings.HasPrefix(param, "obfs-host=") {
								host = param[10:]
							}
						}

						ss["plugin-opts"] = map[string]any{
							"mode": mode,
							"host": host,
						}
					}
				} else {
					// 解析新格式的插件参数
					obfsParams := strings.Split(pluginOpts, ";")
					ss["plugin"] = "obfs"

					pluginOptsMap := map[string]any{}
					for _, param := range obfsParams {
						if strings.HasPrefix(param, "mode=") {
							pluginOptsMap["mode"] = param[5:]
						} else if strings.HasPrefix(param, "host=") {
							pluginOptsMap["host"] = param[5:]
						}
					}

					ss["plugin-opts"] = pluginOptsMap
				}
			}

			// 处理UDP over TCP
			if query.Get("udp-over-tcp") == "true" || query.Get("uot") == "1" {
				ss["udp"] = true
			}
		}
	}

	return ss
}

// ParseTrojanURL 解析Trojan URL
func ParseTrojanURL(trojanURL string) map[string]any {
	// trojan://password@server:port?params#name
	u, err := url.Parse(trojanURL)
	if err != nil {
		log.Printf("Trojan URL解析失败: %v", err)
		return nil
	}

	if u.Scheme != "trojan" {
		return nil
	}

	password := u.User.String()
	server := u.Hostname()
	port := u.Port()

	if server == "" || port == "" || password == "" {
		log.Printf("Trojan URL缺少必要参数")
		return nil
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Printf("Trojan端口号格式错误: %v", err)
		return nil
	}

	name := u.Fragment
	if name == "" {
		name = "Trojan节点"
	}

	// 解析查询参数
	query := u.Query()
	skipCertVerify := query.Get("allowInsecure") == "1"
	sni := query.Get("sni")
	if sni == "" {
		sni = server
	}

	return map[string]any{
		"name":             name,
		"type":             "trojan",
		"server":           server,
		"port":             portInt,
		"password":         password,
		"skip-cert-verify": skipCertVerify,
		"sni":              sni,
	}
}

// ParseVlessURL 解析VLESS URL
func ParseVlessURL(vlessURL string) map[string]any {
	// vless://uuid@server:port?params#name
	u, err := url.Parse(vlessURL)
	if err != nil {
		log.Printf("VLESS URL解析失败: %v", err)
		return nil
	}

	if u.Scheme != "vless" {
		return nil
	}

	uuid := u.User.String()
	server := u.Hostname()
	port := u.Port()

	if server == "" || port == "" || uuid == "" {
		log.Printf("VLESS URL缺少必要参数")
		return nil
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Printf("VLESS端口号格式错误: %v", err)
		return nil
	}

	name := u.Fragment
	if name == "" {
		name = "VLESS节点"
	}

	// 解析查询参数
	query := u.Query()
	network := query.Get("type")
	if network == "" {
		network = "tcp"
	}

	security := query.Get("security")
	tls := security == "tls" || security == "reality"

	proxy := map[string]any{
		"name":    name,
		"type":    "vless",
		"server":  server,
		"port":    portInt,
		"uuid":    uuid,
		"network": network,
		"tls":     tls,
		"udp":     true,
	}

	// Reality 设置
	if security == "reality" {
		realityOpts := map[string]any{
			"public-key": query.Get("pbk"),
			"short-id":   query.Get("sid"),
		}
		proxy["reality-opts"] = realityOpts
	}

	// 流控设置
	flow := query.Get("flow")
	if flow != "" {
		proxy["flow"] = flow
	}

	fp := query.Get("fp")
	if fp != "" {
		proxy["client-fingerprint"] = fp
	}

	sni := query.Get("sni")
	if sni != "" {
		proxy["servername"] = sni
	}

	return proxy
}

// ParseHysteria2URL 解析Hysteria2 URL
func ParseHysteria2URL(hysteria2URL string) map[string]any {
	// hysteria2://password@server:port/?params#name
	u, err := url.Parse(hysteria2URL)
	if err != nil {
		log.Printf("Hysteria2 URL解析失败: %v", err)
		return nil
	}

	if u.Scheme != "hysteria2" {
		return nil
	}

	password := u.User.String()
	server := u.Hostname()
	port := u.Port()

	if server == "" || port == "" || password == "" {
		log.Printf("Hysteria2 URL缺少必要参数")
		return nil
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Printf("Hysteria2端口号格式错误: %v", err)
		return nil
	}

	name := u.Fragment
	if name == "" {
		name = "Hysteria2节点"
	}

	// 解析查询参数
	query := u.Query()
	skipCertVerify := query.Get("insecure") == "1"
	sni := query.Get("sni")
	if sni == "" {
		sni = server
	}

	return map[string]any{
		"name":             name,
		"type":             "hysteria2",
		"server":           server,
		"port":             portInt,
		"password":         password,
		"skip-cert-verify": skipCertVerify,
		"sni":              sni,
	}
}

// ParseHysteriaURL 解析Hysteria URL
func ParseHysteriaURL(hysteriaURL string) map[string]any {
	// hysteria://password@server:port/?params#name
	u, err := url.Parse(hysteriaURL)
	if err != nil {
		log.Printf("Hysteria URL解析失败: %v", err)
		return nil
	}

	if u.Scheme != "hysteria" {
		return nil
	}

	server := u.Hostname()
	port := u.Port()
	password := u.User.String()

	if server == "" || port == "" {
		log.Printf("Hysteria URL缺少必要参数")
		return nil
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Printf("Hysteria端口号格式错误: %v", err)
		return nil
	}

	name := u.Fragment
	if name == "" {
		name = "Hysteria节点"
	}

	// 解析查询参数
	query := u.Query()

	hysteria := map[string]any{
		"name":   name,
		"type":   "hysteria",
		"server": server,
		"port":   portInt,
	}

	// 添加认证信息
	if password != "" {
		hysteria["auth_str"] = password
	}

	// 添加SNI
	sni := query.Get("peer")
	if sni != "" {
		hysteria["sni"] = sni
	}

	// 添加混淆
	obfs := query.Get("obfs")
	if obfs != "" {
		hysteria["obfs"] = obfs
	}

	// 添加ALPN
	alpn := query.Get("alpn")
	if alpn != "" {
		hysteria["alpn"] = strings.Split(alpn, ",")
	}

	// 添加协议
	protocol := query.Get("protocol")
	if protocol != "" {
		hysteria["protocol"] = protocol
	}

	// 添加上下行速率
	up := query.Get("up")
	if up == "" {
		up = query.Get("upmbps")
	}
	if up != "" {
		hysteria["up"] = up
	}

	down := query.Get("down")
	if down == "" {
		down = query.Get("downmbps")
	}
	if down != "" {
		hysteria["down"] = down
	}

	// 添加证书验证设置
	insecure := query.Get("insecure")
	if insecure == "1" {
		hysteria["skip-cert-verify"] = true
	}

	return hysteria
}

// ParseTuicURL 解析TUIC URL
func ParseTuicURL(tuicURL string) map[string]any {
	// tuic://token@server:port/?params#name
	u, err := url.Parse(tuicURL)
	if err != nil {
		log.Printf("TUIC URL解析失败: %v", err)
		return nil
	}

	if u.Scheme != "tuic" {
		return nil
	}

	server := u.Hostname()
	port := u.Port()

	if server == "" || port == "" {
		log.Printf("TUIC URL缺少必要参数")
		return nil
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Printf("TUIC端口号格式错误: %v", err)
		return nil
	}

	name := u.Fragment
	if name == "" {
		name = "TUIC节点"
	}

	// 解析查询参数
	query := u.Query()

	tuic := map[string]any{
		"name":   name,
		"type":   "tuic",
		"server": server,
		"port":   portInt,
		"udp":    true,
	}

	// 处理认证信息
	password, passwordSet := u.User.Password()
	if passwordSet {
		// TUICv5 格式: uuid:password
		tuic["uuid"] = u.User.Username()
		tuic["password"] = password
	} else {
		// TUICv4 格式: token
		tuic["token"] = u.User.Username()
	}

	// 拥塞控制
	cc := query.Get("congestion_control")
	if cc != "" {
		tuic["congestion-control"] = cc
	}

	// ALPN
	alpn := query.Get("alpn")
	if alpn != "" {
		tuic["alpn"] = strings.Split(alpn, ",")
	}

	// SNI
	sni := query.Get("sni")
	if sni != "" {
		tuic["sni"] = sni
	}

	// 禁用SNI
	if query.Get("disable_sni") == "1" {
		tuic["disable-sni"] = true
	}

	// UDP中继模式
	udpRelayMode := query.Get("udp_relay_mode")
	if udpRelayMode != "" {
		tuic["udp-relay-mode"] = udpRelayMode
	}

	return tuic
}

// ParseSSRURL 解析ShadowsocksR URL
func ParseSSRURL(ssrURL string) map[string]any {
	// ssr://base64编码的内容
	if !strings.HasPrefix(ssrURL, "ssr://") {
		return nil
	}

	// 移除前缀并解码
	encoded := ssrURL[6:]
	decoded, err := Base64RawStdDecode(encoded)
	if err != nil {
		log.Printf("SSR URL解码失败: %v", err)
		return nil
	}

	// 分离参数部分
	var beforePart, afterPart string
	parts := strings.SplitN(decoded, "/?", 2)
	if len(parts) == 2 {
		beforePart = parts[0]
		afterPart = parts[1]
	} else {
		beforePart = parts[0]
		afterPart = ""
	}

	// 解析服务器信息部分
	beforeArr := strings.Split(beforePart, ":")
	if len(beforeArr) < 6 {
		log.Printf("SSR URL格式无效")
		return nil
	}

	host := beforeArr[0]
	port := beforeArr[1]
	protocol := beforeArr[2]
	method := beforeArr[3]
	obfs := beforeArr[4]

	// 解码密码
	passwordEncoded := URLSafe(beforeArr[5])
	password, err := Base64RawURLDecode(passwordEncoded)
	if err != nil {
		log.Printf("SSR密码解码失败: %v", err)
		return nil
	}

	// 解析查询参数
	var obfsParam, protocolParam, remarks string
	if afterPart != "" {
		query, err := url.ParseQuery(URLSafe(afterPart))
		if err != nil {
			log.Printf("SSR参数解析失败: %v", err)
		} else {
			if query.Get("obfsparam") != "" {
				obfsParamEncoded := query.Get("obfsparam")
				obfsParam, _ = Base64RawURLDecode(obfsParamEncoded)
			}

			if query.Get("protoparam") != "" {
				protocolParamEncoded := query.Get("protoparam")
				protocolParam, _ = Base64RawURLDecode(protocolParamEncoded)
			}

			if query.Get("remarks") != "" {
				remarksEncoded := query.Get("remarks")
				remarks, _ = Base64RawURLDecode(remarksEncoded)
			}
		}
	}

	if remarks == "" {
		remarks = "SSR节点"
	}

	// 转换为整数的端口
	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Printf("SSR端口号格式错误: %v", err)
		return nil
	}

	ssr := map[string]any{
		"name":     remarks,
		"type":     "ssr",
		"server":   host,
		"port":     portInt,
		"cipher":   method,
		"password": password,
		"protocol": protocol,
		"obfs":     obfs,
		"udp":      true,
	}

	if obfsParam != "" {
		ssr["obfs-param"] = obfsParam
	}

	if protocolParam != "" {
		ssr["protocol-param"] = protocolParam
	}

	return ssr
}

// GenerateClashConfig 生成Clash配置
func GenerateClashConfig(proxies []map[string]any) map[string]any {
	// 基本配置
	config := map[string]any{
		"proxies": proxies,
	}

	// 代理组配置
	proxyNames := make([]any, len(proxies))
	for i, proxy := range proxies {
		proxyNames[i] = proxy["name"]
	}

	proxyGroups := []map[string]any{
		{
			"name":    "🚀 节点选择",
			"type":    "select",
			"proxies": append([]any{"♻️ 自动选择", "DIRECT"}, proxyNames...),
		},
		{
			"name":     "♻️ 自动选择",
			"type":     "url-test",
			"proxies":  proxyNames,
			"url":      "http://www.gstatic.com/generate_204",
			"interval": 300,
		},
	}

	config["proxy-groups"] = proxyGroups

	// 规则配置
	rules := []string{
		"DOMAIN-SUFFIX,local,DIRECT",
		"IP-CIDR,127.0.0.0/8,DIRECT",
		"IP-CIDR,172.16.0.0/12,DIRECT",
		"IP-CIDR,192.168.0.0/16,DIRECT",
		"IP-CIDR,10.0.0.0/8,DIRECT",
		"GEOIP,CN,DIRECT",
		"MATCH,🚀 节点选择",
	}

	config["rules"] = rules

	return config
}

// GetStringOrDefault 获取字符串值或默认值
func GetStringOrDefault(value any, defaultValue string) string {
	if str, ok := value.(string); ok {
		return str
	}
	return defaultValue
}

// UniqueName 确保名称在映射中唯一
func UniqueName(names map[string]bool, name string) string {
	if name == "" {
		name = "未命名节点"
	}

	originalName := name
	counter := 1

	for {
		if _, exists := names[name]; !exists {
			names[name] = true
			return name
		}
		name = fmt.Sprintf("%s-%d", originalName, counter)
		counter++
	}
}

// Base64RawURLDecode 解码URL安全的Base64字符串
func Base64RawURLDecode(s string) (string, error) {
	if len(s)%4 != 0 {
		s = s + strings.Repeat("=", 4-len(s)%4)
	}
	bytes, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Base64RawStdDecode 解码标准Base64字符串
func Base64RawStdDecode(s string) (string, error) {
	if len(s)%4 != 0 {
		s = s + strings.Repeat("=", 4-len(s)%4)
	}
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// URLSafe 使字符串URL安全
func URLSafe(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "+", "-"), "/", "_")
}
