package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"submanager/model"
)

// ParsedNode represents a unified proxy node.
type ParsedNode struct {
	Name     string         `json:"name"`
	Protocol string         `json:"protocol"`
	Address  string         `json:"address"`
	Port     int            `json:"port"`
	Extra    map[string]any `json:"extra"`
	RawLink  string         `json:"raw_link"`
}

// ParseSubscriptionContent decodes and parses subscription content into nodes.
func ParseSubscriptionContent(content string) ([]ParsedNode, error) {
	// Try base64 decode first
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(content))
	if err != nil {
		// Try URL-safe base64
		decoded, err = base64.URLEncoding.DecodeString(strings.TrimSpace(content))
		if err != nil {
			// Assume raw content
			decoded = []byte(content)
		}
	}

	lines := strings.Split(strings.TrimSpace(string(decoded)), "\n")
	var nodes []ParsedNode

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		node, err := ParseProxyLink(line)
		if err != nil {
			continue
		}
		nodes = append(nodes, *node)
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no valid proxy nodes found")
	}

	return nodes, nil
}

// ParseProxyLink parses a single proxy link (vless://, vmess://, ss://, trojan://).
func ParseProxyLink(link string) (*ParsedNode, error) {
	link = strings.TrimSpace(link)

	if strings.HasPrefix(link, "vless://") {
		return parseVLESS(link)
	}
	if strings.HasPrefix(link, "vmess://") {
		return parseVMess(link)
	}
	if strings.HasPrefix(link, "ss://") {
		return parseShadowsocks(link)
	}
	if strings.HasPrefix(link, "trojan://") {
		return parseTrojan(link)
	}

	return nil, fmt.Errorf("unsupported protocol: %s", link[:min(20, len(link))])
}

func parseVLESS(link string) (*ParsedNode, error) {
	// vless://uuid@address:port?params#name
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	port, _ := strconv.Atoi(u.Port())
	node := &ParsedNode{
		Protocol: "vless",
		Address:  u.Hostname(),
		Port:     port,
		RawLink:  link,
		Extra: map[string]any{
			"uuid": u.User.Username(),
		},
	}

	query := u.Query()
	if transport := query.Get("type"); transport != "" {
		node.Extra["transport"] = transport
	}
	if path := query.Get("path"); path != "" {
		node.Extra["path"] = path
	}
	if host := query.Get("host"); host != "" {
		node.Extra["host"] = host
	}
	if sni := query.Get("sni"); sni != "" {
		node.Extra["sni"] = sni
	}
	if query.Get("security") == "tls" {
		node.Extra["tls"] = true
	}
	if flow := query.Get("flow"); flow != "" {
		node.Extra["flow"] = flow
	}

	fragment, _ := url.PathUnescape(u.Fragment)
	node.Name = fragment
	if node.Name == "" {
		node.Name = fmt.Sprintf("VLESS-%s:%d", node.Address, node.Port)
	}

	return node, nil
}

func parseVMess(link string) (*ParsedNode, error) {
	// vmess://base64encodedJSON
	data := strings.TrimPrefix(link, "vmess://")
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decode vmess: %w", err)
		}
	}

	var vmess struct {
		V    string `json:"v"`
		Ps   string `json:"ps"`
		Add  string `json:"add"`
		Port any    `json:"port"`
		ID   string `json:"id"`
		Aid  any    `json:"aid"`
		Scy  string `json:"scy"`
		Net  string `json:"net"`
		Type string `json:"type"`
		Host string `json:"host"`
		Path string `json:"path"`
		TLS  string `json:"tls"`
		Sni  string `json:"sni"`
	}

	if err := parseJSON(decoded, &vmess); err != nil {
		return nil, fmt.Errorf("failed to parse vmess json: %w", err)
	}

	var port int
	switch v := vmess.Port.(type) {
	case float64:
		port = int(v)
	case string:
		port, _ = strconv.Atoi(v)
	}

	extra := map[string]any{
		"uuid":      vmess.ID,
		"transport": vmess.Net,
	}
	if vmess.Scy != "" {
		extra["security"] = vmess.Scy
	} else {
		extra["security"] = "auto"
	}
	if vmess.Host != "" {
		extra["host"] = vmess.Host
	}
	if vmess.Path != "" {
		extra["path"] = vmess.Path
	}
	if vmess.Sni != "" {
		extra["sni"] = vmess.Sni
	}
	if vmess.TLS == "tls" {
		extra["tls"] = true
	}

	alterID := 0
	switch v := vmess.Aid.(type) {
	case float64:
		alterID = int(v)
	case string:
		alterID, _ = strconv.Atoi(v)
	}
	extra["alterId"] = alterID

	name := vmess.Ps
	if name == "" {
		name = fmt.Sprintf("VMess-%s:%d", vmess.Add, port)
	}

	return &ParsedNode{
		Name:     name,
		Protocol: "vmess",
		Address:  vmess.Add,
		Port:     port,
		Extra:    extra,
		RawLink:  link,
	}, nil
}

func parseShadowsocks(link string) (*ParsedNode, error) {
	// ss://base64(method:password)@address:port#name
	// or ss://base64(method:password@address:port)#name (SIP002 legacy)
	data := strings.TrimPrefix(link, "ss://")
	parts := strings.SplitN(data, "#", 2)

	var name string
	if len(parts) == 2 {
		name, _ = url.PathUnescape(parts[1])
	}

	credHost := parts[0]

	// Try SIP002 format: base64(method:password)@host:port
	if atIdx := strings.LastIndex(credHost, "@"); atIdx != -1 {
		methodPass := credHost[:atIdx]
		hostPort := credHost[atIdx+1:]

		decoded, err := base64.StdEncoding.DecodeString(methodPass)
		if err != nil {
			decoded, err = base64.URLEncoding.DecodeString(methodPass)
			if err != nil {
				return nil, fmt.Errorf("failed to decode ss credentials: %w", err)
			}
		}

		colonIdx := strings.Index(string(decoded), ":")
		if colonIdx == -1 {
			return nil, fmt.Errorf("invalid ss method:password format")
		}
		method := string(decoded[:colonIdx])
		password := string(decoded[colonIdx+1:])

		host, portStr, err := net.SplitHostPort(hostPort)
		if err != nil {
			return nil, fmt.Errorf("invalid ss host:port: %w", err)
		}
		port, _ := strconv.Atoi(portStr)

		if name == "" {
			name = fmt.Sprintf("SS-%s:%d", host, port)
		}

		return &ParsedNode{
			Name:     name,
			Protocol: "shadowsocks",
			Address:  host,
			Port:     port,
			RawLink:  link,
			Extra: map[string]any{
				"method":   method,
				"password": password,
			},
		}, nil
	}

	// Legacy format: base64(method:password@host:port)
	decoded, err := base64.StdEncoding.DecodeString(credHost)
	if err != nil {
		return nil, err
	}

	hostStart := strings.LastIndex(string(decoded), "@")
	if hostStart == -1 {
		return nil, fmt.Errorf("invalid ss link format")
	}

	methodPass := string(decoded[:hostStart])
	hostPort := string(decoded[hostStart+1:])

	colonIdx := strings.Index(methodPass, ":")
	if colonIdx == -1 {
		return nil, fmt.Errorf("invalid ss method:password")
	}

	host, portStr, err := net.SplitHostPort(hostPort)
	if err != nil {
		return nil, err
	}
	port, _ := strconv.Atoi(portStr)

	if name == "" {
		name = fmt.Sprintf("SS-%s:%d", host, port)
	}

	return &ParsedNode{
		Name:     name,
		Protocol: "shadowsocks",
		Address:  host,
		Port:     port,
		RawLink:  link,
		Extra: map[string]any{
			"method":   methodPass[:colonIdx],
			"password": methodPass[colonIdx+1:],
		},
	}, nil
}

func parseTrojan(link string) (*ParsedNode, error) {
	// trojan://password@address:port?params#name
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	port, _ := strconv.Atoi(u.Port())
	node := &ParsedNode{
		Protocol: "trojan",
		Address:  u.Hostname(),
		Port:     port,
		RawLink:  link,
		Extra: map[string]any{
			"password": u.User.Username(),
		},
	}

	query := u.Query()
	if sni := query.Get("sni"); sni != "" {
		node.Extra["sni"] = sni
	}
	if transport := query.Get("type"); transport != "" {
		node.Extra["transport"] = transport
	}
	if path := query.Get("path"); path != "" {
		node.Extra["path"] = path
	}
	if host := query.Get("host"); host != "" {
		node.Extra["host"] = host
	}

	fragment, _ := url.PathUnescape(u.Fragment)
	node.Name = fragment
	if node.Name == "" {
		node.Name = fmt.Sprintf("Trojan-%s:%d", node.Address, node.Port)
	}

	return node, nil
}

// IsPrivateIP checks if an IP address is in private/reserved ranges.
func IsPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	privateRanges := []struct {
		network *net.IPNet
	}{
		{mustParseCIDR("10.0.0.0/8")},
		{mustParseCIDR("172.16.0.0/12")},
		{mustParseCIDR("192.168.0.0/16")},
		{mustParseCIDR("127.0.0.0/8")},
		{mustParseCIDR("169.254.0.0/16")},
		{mustParseCIDR("::1/128")},
		{mustParseCIDR("fc00::/7")},
		{mustParseCIDR("fe80::/10")},
	}

	for _, r := range privateRanges {
		if r.network.Contains(ip) {
			return true
		}
	}
	return false
}

func mustParseCIDR(s string) *net.IPNet {
	_, network, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return network
}

// FetchSubscription fetches subscription content from a URL with SSRF protection.
func FetchSubscription(rawURL string, headers map[string]any) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Resolve hostname and check for SSRF
	host := parsedURL.Hostname()
	ips, err := net.LookupIP(host)
	if err != nil {
		return "", fmt.Errorf("failed to resolve host: %w", err)
	}

	for _, ip := range ips {
		if IsPrivateIP(ip.String()) {
			return "", fmt.Errorf("SSRF blocked: %s resolves to private IP %s", host, ip)
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "SubManager/1.0")
	for k, v := range headers {
		req.Header.Set(k, fmt.Sprintf("%v", v))
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch returned status %d", resp.StatusCode)
	}

	buf := new(strings.Builder)
	n, err := io.Copy(buf, io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return "", fmt.Errorf("read body failed: %w", err)
	}
	_ = n
	return buf.String(), nil
}

// ToNodeCache converts a ParsedNode to a NodeCache model.
func ToNodeCache(node ParsedNode, sourceType string, sourceID uint) model.NodeCache {
	return model.NodeCache{
		SourceType: sourceType,
		SourceID:   sourceID,
		Name:       node.Name,
		RawLink:    node.RawLink,
		Protocol:   node.Protocol,
		Address:    node.Address,
		Port:       node.Port,
		Extra:      model.JSONMap(node.Extra),
	}
}

func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
