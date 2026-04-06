package service

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestConvertToSingboxJSON(t *testing.T) {
	nodes := []ParsedNode{
		{
			Name: "HK-VLESS", Protocol: "vless", Address: "1.2.3.4", Port: 443,
			Extra:   map[string]any{"uuid": "test-uuid", "transport": "ws", "path": "/ws", "host": "example.com", "tls": true, "sni": "example.com"},
			RawLink: "vless://test-uuid@1.2.3.4:443#HK-VLESS",
		},
		{
			Name: "US-SS", Protocol: "shadowsocks", Address: "5.6.7.8", Port: 8388,
			Extra:   map[string]any{"method": "aes-256-gcm", "password": "pass123"},
			RawLink: "ss://pass@5.6.7.8:8388#US-SS",
		},
		{
			Name: "JP-Trojan", Protocol: "trojan", Address: "9.10.11.12", Port: 443,
			Extra:   map[string]any{"password": "trojan-pass", "sni": "example.com"},
			RawLink: "trojan://trojan-pass@9.10.11.12:443#JP-Trojan",
		},
	}

	result := ConvertToSingboxJSON(nodes, "SubManager")

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("Result is not valid JSON: %v", err)
	}

	// Check outbounds array
	outbounds, ok := parsed["outbounds"].([]interface{})
	if !ok {
		t.Fatal("Expected outbounds array")
	}

	// Should have 3 nodes + 1 selector = 4
	if len(outbounds) != 4 {
		t.Fatalf("Expected 4 outbounds (3 nodes + 1 selector), got %d", len(outbounds))
	}

	// Check first outbound is vless
	first := outbounds[0].(map[string]interface{})
	if first["type"] != "vless" {
		t.Errorf("Expected first outbound type vless, got %v", first["type"])
	}
	if first["server"] != "1.2.3.4" {
		t.Errorf("Expected server 1.2.3.4, got %v", first["server"])
	}

	// Check selector outbound (last)
	selector := outbounds[3].(map[string]interface{})
	if selector["type"] != "selector" {
		t.Errorf("Expected selector type, got %v", selector["type"])
	}
	if selector["tag"] != "SubManager" {
		t.Errorf("Expected selector tag SubManager, got %v", selector["tag"])
	}
}

func TestConvertToSingboxJSON_VMessWithTLS(t *testing.T) {
	nodes := []ParsedNode{
		{
			Name: "VMess-TLS", Protocol: "vmess", Address: "1.2.3.4", Port: 443,
			Extra:   map[string]any{"uuid": "vmess-uuid", "security": "auto", "alterId": 0, "tls": true, "sni": "example.com", "transport": "ws", "path": "/ws"},
			RawLink: "vmess://vmess-uuid@1.2.3.4:443#VMess-TLS",
		},
	}

	result := ConvertToSingboxJSON(nodes, "Test")

	var parsed map[string]interface{}
	json.Unmarshal([]byte(result), &parsed)

	outbounds := parsed["outbounds"].([]interface{})
	first := outbounds[0].(map[string]interface{})

	if first["type"] != "vmess" {
		t.Errorf("Expected vmess type, got %v", first["type"])
	}
	if first["uuid"] != "vmess-uuid" {
		t.Errorf("Expected uuid, got %v", first["uuid"])
	}

	// Check TLS
	tls, ok := first["tls"].(map[string]interface{})
	if !ok {
		t.Error("Expected TLS config")
	} else {
		if tls["enabled"] != true {
			t.Error("Expected TLS enabled")
		}
		if tls["server_name"] != "example.com" {
			t.Errorf("Expected server_name example.com, got %v", tls["server_name"])
		}
	}

	// Check transport
	transport, ok := first["transport"].(map[string]interface{})
	if !ok {
		t.Error("Expected transport config")
	} else {
		if transport["type"] != "ws" {
			t.Errorf("Expected transport ws, got %v", transport["type"])
		}
	}
}

func TestConvertToSingboxJSON_Empty(t *testing.T) {
	result := ConvertToSingboxJSON(nil, "Test")
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("Empty result should be valid JSON: %v", err)
	}
}

func TestConvertToSingboxJSON_AllProtocols(t *testing.T) {
	nodes := []ParsedNode{
		{Name: "V", Protocol: "vless", Address: "1.1.1.1", Port: 443, Extra: map[string]any{"uuid": "u1"}},
		{Name: "M", Protocol: "vmess", Address: "2.2.2.2", Port: 443, Extra: map[string]any{"uuid": "u2"}},
		{Name: "S", Protocol: "shadowsocks", Address: "3.3.3.3", Port: 8388, Extra: map[string]any{"method": "aes-256-gcm", "password": "p"}},
		{Name: "T", Protocol: "trojan", Address: "4.4.4.4", Port: 443, Extra: map[string]any{"password": "p"}},
	}

	result := ConvertToSingboxJSON(nodes, "All")

	if !strings.Contains(result, `"type": "vless"`) {
		t.Error("Expected vless outbound")
	}
	if !strings.Contains(result, `"type": "vmess"`) {
		t.Error("Expected vmess outbound")
	}
	if !strings.Contains(result, `"type": "shadowsocks"`) {
		t.Error("Expected shadowsocks outbound")
	}
	if !strings.Contains(result, `"type": "trojan"`) {
		t.Error("Expected trojan outbound")
	}
	if !strings.Contains(result, `"type": "selector"`) {
		t.Error("Expected selector outbound")
	}
}
