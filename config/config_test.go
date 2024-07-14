package config

import (
	"testing"
)

func TestEmbedConfig(t *testing.T) {
	wantIP := "192.168.3.3"
	EmbedTestConfig(`KindleIP = "192.168.3.3"`)
	kindleIP := GetConfig().KindleIP

	if kindleIP != wantIP {
		t.Errorf("got %q, want %q", kindleIP, wantIP)
	}

	var overWriteTestConfig = `KindleIP = "192.168.4.4"`

	wantIP = "192.168.4.4"
	EmbedTestConfig(overWriteTestConfig)
	kindleIP = GetConfig().KindleIP

	if kindleIP != wantIP {
		t.Errorf("got %q, want %q", kindleIP, wantIP)
	}
}
