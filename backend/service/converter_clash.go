package service

import (
	"fmt"
	"strings"
)

// ConvertToClashYAML converts nodes to Clash YAML config format.
func ConvertToClashYAML(nodes []ParsedNode, subName string) string {
	var sb strings.Builder

	sb.WriteString("proxies:\n")
	for _, n := range nodes {
		switch n.Protocol {
		case "vless":
			sb.WriteString(clashVLESS(n))
		case "vmess":
			sb.WriteString(clashVMess(n))
		case "shadowsocks":
			sb.WriteString(clashSS(n))
		case "trojan":
			sb.WriteString(clashTrojan(n))
		}
	}

	sb.WriteString("\nproxy-groups:\n")
	sb.WriteString(fmt.Sprintf("  - name: \"%s\"\n", subName))
	sb.WriteString("    type: select\n")
	sb.WriteString("    proxies:\n")
	for _, n := range nodes {
		sb.WriteString(fmt.Sprintf("      - \"%s\"\n", clashEscape(n.Name)))
	}

	return sb.String()
}

func clashVLESS(n ParsedNode) string {
	tls := getBool(n.Extra, "tls")
	transport := getStr(n.Extra, "transport")

	yaml := fmt.Sprintf("  - name: \"%s\"\n    type: vless\n    server: %s\n    port: %d\n    uuid: %s\n",
		clashEscape(n.Name), n.Address, n.Port, getStr(n.Extra, "uuid"))

	if tls {
		yaml += "    tls: true\n"
		if sni := getStr(n.Extra, "sni"); sni != "" {
			yaml += fmt.Sprintf("    sni: %s\n", sni)
		}
		if flow := getStr(n.Extra, "flow"); flow != "" {
			yaml += fmt.Sprintf("    flow: %s\n", flow)
		}
	}

	if transport == "ws" {
		yaml += "    network: ws\n"
		if path := getStr(n.Extra, "path"); path != "" {
			yaml += fmt.Sprintf("    ws-opts:\n      path: \"%s\"\n", path)
		}
		if host := getStr(n.Extra, "host"); host != "" {
			yaml += fmt.Sprintf("      headers:\n        Host: \"%s\"\n", host)
		}
	} else if transport == "grpc" {
		yaml += "    network: grpc\n"
	}

	return yaml
}

func clashVMess(n ParsedNode) string {
	tls := getBool(n.Extra, "tls")
	transport := getStr(n.Extra, "transport")
	alterID := getInt(n.Extra, "alterId")

	yaml := fmt.Sprintf("  - name: \"%s\"\n    type: vmess\n    server: %s\n    port: %d\n    uuid: %s\n    alterId: %d\n    cipher: %s\n",
		clashEscape(n.Name), n.Address, n.Port, getStr(n.Extra, "uuid"), alterID, getStr(n.Extra, "security"))

	if tls {
		yaml += "    tls: true\n"
		if sni := getStr(n.Extra, "sni"); sni != "" {
			yaml += fmt.Sprintf("    sni: %s\n", sni)
		}
	}

	if transport == "ws" {
		yaml += "    network: ws\n"
		opts := ""
		if path := getStr(n.Extra, "path"); path != "" {
			opts += fmt.Sprintf("      path: \"%s\"\n", path)
		}
		if host := getStr(n.Extra, "host"); host != "" {
			opts += fmt.Sprintf("      headers:\n        Host: \"%s\"\n", host)
		}
		if opts != "" {
			yaml += "    ws-opts:\n" + opts
		}
	} else if transport == "grpc" {
		yaml += "    network: grpc\n"
	}

	return yaml
}

func clashSS(n ParsedNode) string {
	return fmt.Sprintf("  - name: \"%s\"\n    type: ss\n    server: %s\n    port: %d\n    cipher: %s\n    password: \"%s\"\n",
		clashEscape(n.Name), n.Address, n.Port, getStr(n.Extra, "method"), getStr(n.Extra, "password"))
}

func clashTrojan(n ParsedNode) string {
	yaml := fmt.Sprintf("  - name: \"%s\"\n    type: trojan\n    server: %s\n    port: %d\n    password: \"%s\"\n",
		clashEscape(n.Name), n.Address, n.Port, getStr(n.Extra, "password"))

	if sni := getStr(n.Extra, "sni"); sni != "" {
		yaml += fmt.Sprintf("    sni: %s\n", sni)
	}

	transport := getStr(n.Extra, "transport")
	if transport == "ws" {
		yaml += "    network: ws\n"
		if path := getStr(n.Extra, "path"); path != "" {
			yaml += fmt.Sprintf("    ws-opts:\n      path: \"%s\"\n", path)
		}
	} else if transport == "grpc" {
		yaml += "    network: grpc\n"
	}

	return yaml
}

func clashEscape(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
}

func getStr(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(m map[string]any, key string) int {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case int:
			return n
		case float64:
			return int(n)
		}
	}
	return 0
}

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
