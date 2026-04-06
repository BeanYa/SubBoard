package service

import (
	"strings"
	"testing"
)

func TestConvertToClashYAML(t *testing.T) {
	nodes := []ParsedNode{
		{
			Name: "HK-VLESS", Protocol: "vless", Address: "1.2.3.4", Port: 443,
			Extra:   map[string]any{"uuid": "test-uuid", "transport": "ws", "path": "/ws", "tls": true, "sni": "example.com"},
			RawLink: "vless://test-uuid@1.2.3.4:443#HK-VLESS",
		},
		{
			Name: "US-VMess", Protocol: "vmess", Address: "5.6.7.8", Port: 443,
			Extra:   map[string]any{"uuid": "vmess-uuid", "security": "auto", "alterId": 0, "tls": true},
			RawLink: "vmess://vmess-uuid@5.6.7.8:443#US-VMess",
		},
		{
			Name: "JP-SS", Protocol: "shadowsocks", Address: "9.10.11.12", Port: 8388,
			Extra:   map[string]any{"method": "aes-256-gcm", "password": "pass123"},
			RawLink: "ss://pass@9.10.11.12:8388#JP-SS",
		},
		{
			Name: "SG-Trojan", Protocol: "trojan", Address: "13.14.15.16", Port: 443,
			Extra:   map[string]any{"password": "trojan-pass", "sni": "example.com"},
			RawLink: "trojan://trojan-pass@13.14.15.16:443#SG-Trojan",
		},
	}

	result := ConvertToClashYAML(nodes, "SubManager")

	if !strings.Contains(result, "proxies:") {
		t.Error("expected proxies section")
	}
	if !strings.Contains(result, "proxy-groups:") {
		t.Error("expected proxy-groups section")
	}
	if !strings.Contains(result, "HK-VLESS") {
		t.Error("expected HK-VLESS node")
	}
	if !strings.Contains(result, "US-VMess") {
		t.Error("expected US-VMess node")
	}
	if !strings.Contains(result, "JP-SS") {
		t.Error("expected JP-SS node")
	}
	if !strings.Contains(result, "SG-Trojan") {
		t.Error("expected SG-Trojan node")
	}
	if !strings.Contains(result, "type: vless") {
		t.Error("expected vless type")
	}
	if !strings.Contains(result, "type: vmess") {
		t.Error("expected vmess type")
	}
	if !strings.Contains(result, "type: ss") {
		t.Error("expected ss type")
	}
	if !strings.Contains(result, "type: trojan") {
		t.Error("expected trojan type")
	}
}

func TestConvertToClashYAML_Empty(t *testing.T) {
	result := ConvertToClashYAML(nil, "SubManager")
	if !strings.Contains(result, "proxies:") {
		t.Error("expected proxies section even with no nodes")
	}
}

func TestConvertToBase64(t *testing.T) {
	nodes := []ParsedNode{
		{RawLink: "vless://a@1.2.3.4:443#A"},
		{RawLink: "trojan://b@5.6.7.8:443#B"},
	}
	result := ConvertToBase64(nodes)
	if result == "" {
		t.Error("expected non-empty base64 result")
	}
	// Should be valid base64 containing both links
	if len(result) < 10 {
		t.Error("expected base64 output")
	}
}

func TestConvertToRaw(t *testing.T) {
	nodes := []ParsedNode{
		{RawLink: "vless://a@1.2.3.4:443#A"},
		{RawLink: "trojan://b@5.6.7.8:443#B"},
	}
	result := ConvertToRaw(nodes)
	if !strings.Contains(result, "vless://a@1.2.3.4:443#A") {
		t.Error("expected first link in raw output")
	}
	if !strings.Contains(result, "trojan://b@5.6.7.8:443#B") {
		t.Error("expected second link in raw output")
	}
	if strings.Count(result, "\n") != 1 {
		t.Error("expected exactly 1 newline (2 links)")
	}
}
