package service

import (
	"testing"
)

func TestParseProxyLink_VLESS(t *testing.T) {
	link := "vless://a3482e88-686a-4a58-8126-99c9df64b7bf@1.2.3.4:443?type=ws&path=%2Fvless-ws&host=example.com&security=tls&sni=example.com#HK-Node"
	node, err := ParseProxyLink(link)
	if err != nil {
		t.Fatalf("ParseProxyLink failed: %v", err)
	}
	if node.Protocol != "vless" {
		t.Errorf("expected protocol vless, got %s", node.Protocol)
	}
	if node.Address != "1.2.3.4" {
		t.Errorf("expected address 1.2.3.4, got %s", node.Address)
	}
	if node.Port != 443 {
		t.Errorf("expected port 443, got %d", node.Port)
	}
	if node.Name != "HK-Node" {
		t.Errorf("expected name HK-Node, got %s", node.Name)
	}
	if getStr(node.Extra, "uuid") != "a3482e88-686a-4a58-8126-99c9df64b7bf" {
		t.Errorf("expected uuid in extra, got %v", node.Extra["uuid"])
	}
	if getStr(node.Extra, "transport") != "ws" {
		t.Errorf("expected transport ws, got %v", node.Extra["transport"])
	}
	if !getBool(node.Extra, "tls") {
		t.Error("expected tls=true")
	}
}

func TestParseProxyLink_VMess(t *testing.T) {
	// vmess://base64({"v":"2","ps":"test","add":"1.2.3.4","port":443,"id":"uuid-1234","aid":0,"net":"ws","host":"example.com","path":"/ws","tls":"tls"})
	link := "vmess://eyJ2IjoiMiIsInBzIjoidGVzdCIsImFkZCI6IjEuMi4zLjQiLCJwb3J0Ijo0NDMsImlkIjoidXVpZC0xMjM0IiwiYWlkIjowLCJuZXQiOiJ3cyIsImhvc3QiOiJleGFtcGxlLmNvbSIsInBhdGgiOiIvd3MiLCJ0bHMiOiJ0bHMifQ=="
	node, err := ParseProxyLink(link)
	if err != nil {
		t.Fatalf("ParseProxyLink failed: %v", err)
	}
	if node.Protocol != "vmess" {
		t.Errorf("expected protocol vmess, got %s", node.Protocol)
	}
	if node.Address != "1.2.3.4" {
		t.Errorf("expected address 1.2.3.4, got %s", node.Address)
	}
	if node.Port != 443 {
		t.Errorf("expected port 443, got %d", node.Port)
	}
	if node.Name != "test" {
		t.Errorf("expected name test, got %s", node.Name)
	}
}

func TestParseProxyLink_Trojan(t *testing.T) {
	link := "trojan://password123@1.2.3.4:443?sni=example.com&security=tls#US-Node"
	node, err := ParseProxyLink(link)
	if err != nil {
		t.Fatalf("ParseProxyLink failed: %v", err)
	}
	if node.Protocol != "trojan" {
		t.Errorf("expected protocol trojan, got %s", node.Protocol)
	}
	if node.Address != "1.2.3.4" {
		t.Errorf("expected address 1.2.3.4, got %s", node.Address)
	}
	if getStr(node.Extra, "password") != "password123" {
		t.Errorf("expected password password123, got %v", node.Extra["password"])
	}
}

func TestParseProxyLink_Shadowsocks(t *testing.T) {
	// ss://base64(method:password)@host:port#name
	link := "ss://YWVzLTI1Ni1nY206YXNzd29yZA==@1.2.3.4:8388#SS-Node"
	node, err := ParseProxyLink(link)
	if err != nil {
		t.Fatalf("ParseProxyLink failed: %v", err)
	}
	if node.Protocol != "shadowsocks" {
		t.Errorf("expected protocol shadowsocks, got %s", node.Protocol)
	}
	if node.Address != "1.2.3.4" {
		t.Errorf("expected address 1.2.3.4, got %s", node.Address)
	}
}

func TestParseProxyLink_Unsupported(t *testing.T) {
	_, err := ParseProxyLink("http://example.com")
	if err == nil {
		t.Error("expected error for unsupported protocol")
	}
}

func TestParseSubscriptionContent(t *testing.T) {
	// Raw content (one node per line)
	raw := "vless://uuid@1.2.3.4:443?type=ws#Node1\ntrojan://pass@5.6.7.8:443?sni=example.com#Node2"
	nodes, err := ParseSubscriptionContent(raw)
	if err != nil {
		t.Fatalf("ParseSubscriptionContent failed: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].Protocol != "vless" {
		t.Errorf("expected first node vless, got %s", nodes[0].Protocol)
	}
	if nodes[1].Protocol != "trojan" {
		t.Errorf("expected second node trojan, got %s", nodes[1].Protocol)
	}
}

func TestParseSubscriptionContent_Empty(t *testing.T) {
	_, err := ParseSubscriptionContent("")
	if err == nil {
		t.Error("expected error for empty content")
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"127.0.0.1", true},
		{"8.8.8.8", false},
		{"1.2.3.4", false},
		{"", false},
	}
	for _, tt := range tests {
		result := IsPrivateIP(tt.ip)
		if result != tt.expected {
			t.Errorf("IsPrivateIP(%s) = %v, want %v", tt.ip, result, tt.expected)
		}
	}
}

func TestToNodeCache(t *testing.T) {
	node := ParsedNode{
		Name:     "Test-Node",
		Protocol: "vless",
		Address:  "1.2.3.4",
		Port:     443,
		Extra:    map[string]any{"uuid": "test-uuid"},
		RawLink:  "vless://test-uuid@1.2.3.4:443#Test-Node",
	}
	nc := ToNodeCache(node, "subscription", 1)
	if nc.SourceType != "subscription" {
		t.Errorf("expected source_type subscription, got %s", nc.SourceType)
	}
	if nc.SourceID != 1 {
		t.Errorf("expected source_id 1, got %d", nc.SourceID)
	}
	if nc.Name != "Test-Node" {
		t.Errorf("expected name Test-Node, got %s", nc.Name)
	}
	if nc.Protocol != "vless" {
		t.Errorf("expected protocol vless, got %s", nc.Protocol)
	}
}
