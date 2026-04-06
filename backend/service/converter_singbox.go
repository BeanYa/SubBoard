package service

import (
	"encoding/json"
)

// ConvertToSingboxJSON converts nodes to sing-box JSON outbound configuration.
func ConvertToSingboxJSON(nodes []ParsedNode, subName string) string {
	outbounds := make([]interface{}, 0, len(nodes)+1)
	tags := make([]string, 0, len(nodes))

	for _, n := range nodes {
		var outbound map[string]interface{}
		switch n.Protocol {
		case "vless":
			outbound = singboxVLESS(n)
		case "vmess":
			outbound = singboxVMess(n)
		case "shadowsocks":
			outbound = singboxSS(n)
		case "trojan":
			outbound = singboxTrojan(n)
		default:
			continue
		}
		outbounds = append(outbounds, outbound)
		tags = append(tags, n.Name)
	}

	// Add selector outbound
	if len(tags) > 0 {
		selector := map[string]interface{}{
			"type":      "selector",
			"tag":       subName,
			"outbounds": tags,
		}
		outbounds = append(outbounds, selector)
	}

	result := map[string]interface{}{
		"outbounds": outbounds,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data)
}

func singboxVLESS(n ParsedNode) map[string]interface{} {
	outbound := map[string]interface{}{
		"type":        "vless",
		"tag":         n.Name,
		"server":      n.Address,
		"server_port": n.Port,
		"uuid":        getStr(n.Extra, "uuid"),
	}

	if getBool(n.Extra, "tls") {
		tls := map[string]interface{}{"enabled": true}
		if sni := getStr(n.Extra, "sni"); sni != "" {
			tls["server_name"] = sni
		}
		outbound["tls"] = tls
	}

	if flow := getStr(n.Extra, "flow"); flow != "" {
		outbound["flow"] = flow
	}

	addTransport(n, outbound)

	return outbound
}

func singboxVMess(n ParsedNode) map[string]interface{} {
	outbound := map[string]interface{}{
		"type":        "vmess",
		"tag":         n.Name,
		"server":      n.Address,
		"server_port": n.Port,
		"uuid":        getStr(n.Extra, "uuid"),
		"alter_id":    getInt(n.Extra, "alterId"),
		"security":    getStr(n.Extra, "security"),
	}

	if getBool(n.Extra, "tls") {
		tls := map[string]interface{}{"enabled": true}
		if sni := getStr(n.Extra, "sni"); sni != "" {
			tls["server_name"] = sni
		}
		outbound["tls"] = tls
	}

	addTransport(n, outbound)

	return outbound
}

func singboxSS(n ParsedNode) map[string]interface{} {
	return map[string]interface{}{
		"type":        "shadowsocks",
		"tag":         n.Name,
		"server":      n.Address,
		"server_port": n.Port,
		"method":      getStr(n.Extra, "method"),
		"password":    getStr(n.Extra, "password"),
	}
}

func singboxTrojan(n ParsedNode) map[string]interface{} {
	outbound := map[string]interface{}{
		"type":        "trojan",
		"tag":         n.Name,
		"server":      n.Address,
		"server_port": n.Port,
		"password":    getStr(n.Extra, "password"),
	}

	// Trojan always uses TLS
	tls := map[string]interface{}{"enabled": true}
	if sni := getStr(n.Extra, "sni"); sni != "" {
		tls["server_name"] = sni
	}
	outbound["tls"] = tls

	addTransport(n, outbound)

	return outbound
}

// addTransport adds transport configuration if present.
func addTransport(n ParsedNode, outbound map[string]interface{}) {
	transport := getStr(n.Extra, "transport")
	switch transport {
	case "ws":
		ws := map[string]interface{}{"type": "ws"}
		if path := getStr(n.Extra, "path"); path != "" {
			ws["path"] = path
		}
		if host := getStr(n.Extra, "host"); host != "" {
			ws["headers"] = map[string]string{"Host": host}
		}
		outbound["transport"] = ws
	case "grpc":
		outbound["transport"] = map[string]interface{}{"type": "grpc"}
	}
}
