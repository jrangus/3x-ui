package main

import (
	"reflect"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

func TestFirewallProtocols(t *testing.T) {
	tests := []struct {
		name string
		in   model.Inbound
		want []string
	}{
		{"hysteria2", model.Inbound{Protocol: model.Hysteria}, []string{"udp"}},
		{"tuic", model.Inbound{Protocol: model.TUIC}, []string{"udp"}},
		{"wireguard", model.Inbound{Protocol: model.WireGuard}, []string{"udp"}},
		{"shadowsocks", model.Inbound{Protocol: model.Shadowsocks}, []string{"tcp", "udp"}},
		{"vless kcp", model.Inbound{Protocol: model.VLESS, StreamSettings: `{"network":"kcp"}`}, []string{"udp"}},
		{"vless websocket", model.Inbound{Protocol: model.VLESS, StreamSettings: `{"network":"ws"}`}, []string{"tcp"}},
		{"legacy method", model.Inbound{Protocol: model.VMESS, StreamSettings: `{"method":"quic"}`}, []string{"udp"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firewallProtocols(&tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("firewallProtocols() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsLoopbackListen(t *testing.T) {
	for _, addr := range []string{"127.0.0.1", "::1", "[::1]", "localhost"} {
		if !isLoopbackListen(addr) {
			t.Fatalf("%q should be loopback", addr)
		}
	}
	for _, addr := range []string{"", "0.0.0.0", "::", "192.0.2.1"} {
		if isLoopbackListen(addr) {
			t.Fatalf("%q should not be loopback", addr)
		}
	}
}
