package service

import (
	"encoding/base64"
	"strings"
)

// ConvertToBase64 returns base64-encoded raw proxy links (one per line).
func ConvertToBase64(nodes []ParsedNode) string {
	var lines []string
	for _, n := range nodes {
		lines = append(lines, n.RawLink)
	}
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(lines, "\n")))
}

// ConvertToRaw returns raw proxy links (one per line).
func ConvertToRaw(nodes []ParsedNode) string {
	var lines []string
	for _, n := range nodes {
		lines = append(lines, n.RawLink)
	}
	return strings.Join(lines, "\n")
}
