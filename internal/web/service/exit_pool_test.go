package service

import (
	"net/http"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

func TestExitPoolClientDisplayName(t *testing.T) {
	client := model.Client{
		Group:   exitPoolGroupPrefix + "World",
		Comment: "xnv3:7:JP:Japan:203.0.113.7",
	}
	got, ok := ExitPoolClientDisplayName(client)
	if !ok || got != "World-xnvgatebox-stable-JP-S7" {
		t.Fatalf("display name = %q, %v", got, ok)
	}
	client.Comment = "xnv3:8:US:United States:198.51.100.8:sstp_pool"
	if got, ok := ExitPoolClientDisplayName(client); !ok || got != "World-sstp-US-S8" {
		t.Fatalf("SSTP display name = %q, %v", got, ok)
	}
	if _, ok := ExitPoolClientDisplayName(model.Client{Group: "staff"}); ok {
		t.Fatal("ordinary client identified as exit-pool client")
	}
}

func TestExitPoolSourceCategoriesAreIsolatedAndBackwardCompatible(t *testing.T) {
	old, ok := parseExitPoolComment("xnv3:7:JP:Japan:203.0.113.7")
	if !ok || old.SourceCategory != exitPoolSourceStable {
		t.Fatalf("legacy member source = %q, %v", old.SourceCategory, ok)
	}

	home, err := normalizeExitPoolExit(ExitPoolExit{Slot: 100101, SourceCategory: exitPoolSourceSSTP, IPType: "unknown", ProxyAddress: "8.8.8.8", ProxyPort: 1080})
	if err != nil || home.Stable {
		t.Fatalf("SSTP-class supplemental exit rejected or marked stable: %+v, %v", home, err)
	}
	localHome, err := normalizeExitPoolExit(ExitPoolExit{Slot: 2, Port: 17929, SourceCategory: exitPoolSourceSSTP, IPType: "unknown", Transport: "sstp", SocksUser: "u", SocksPass: "p"})
	if err != nil || localHome.Stable || localHome.ProxyAddress != "" {
		t.Fatalf("promoted SSTP residential exit rejected: %+v, %v", localHome, err)
	}
	if _, err := normalizeExitPoolExit(ExitPoolExit{Slot: 2, Port: 17929, SourceCategory: exitPoolSourceSSTP, IPType: "unknown", Transport: "sstp", ProxyAddress: "127.0.0.1"}); err == nil {
		t.Fatal("ambiguous local promoted listener accepted")
	}
	if _, err := normalizeExitPoolExit(ExitPoolExit{Slot: 2, Port: 17929, SourceCategory: exitPoolSourceSSTP, IPType: "unknown"}); err == nil {
		t.Fatal("untyped optional loopback listener accepted")
	}
	legacy, err := normalizeExitPoolExit(ExitPoolExit{Slot: 100104, SourceCategory: exitPoolSourceLegacyNativeDC, IPType: "datacenter", Native: true, ProxyAddress: "8.8.8.8", ProxyPort: 1080})
	if err != nil || legacy.SourceCategory != exitPoolSourceSSTP {
		t.Fatalf("legacy category was not migrated to SSTP pool: %+v, %v", legacy, err)
	}
	if _, err := normalizeExitPoolExit(ExitPoolExit{Slot: 100105, SourceCategory: exitPoolSourceSSTP, IPType: "unknown", ProxyAddress: "127.0.0.1", ProxyPort: 1080}); err == nil {
		t.Fatal("private supplemental proxy was accepted")
	}
	if _, err := normalizeExitPoolExit(ExitPoolExit{Slot: 100106, SourceCategory: exitPoolSourceSSTP, IPType: "unknown", ProxyAddress: "2001:4860:4860::8888", ProxyPort: 1080}); err == nil {
		t.Fatal("IPv6 supplemental proxy was accepted before IPv6 egress verification is supported")
	}
	if _, err := normalizeExitPoolExit(ExitPoolExit{Slot: 1, Port: 1080, SourceCategory: exitPoolSourceStable, ProxyAddress: "8.8.8.8", ProxyPort: 1080}); err == nil {
		t.Fatal("stable category accepted an external proxy")
	}
}

func TestExitPoolGroupRejectsMixedSources(t *testing.T) {
	stable := ExitPoolExit{Slot: 1, SourceCategory: exitPoolSourceStable}
	home := ExitPoolExit{Slot: 2, SourceCategory: exitPoolSourceSSTP}
	if err := validateExitPoolGroupSources([]ExitPoolExit{stable, stable}); err != nil {
		t.Fatalf("same-source group rejected: %v", err)
	}
	if err := validateExitPoolGroupSources([]ExitPoolExit{stable, home}); err == nil {
		t.Fatal("mixed-source group accepted")
	}
}

func TestValidateExitPoolListenAddr(t *testing.T) {
	if err := validateExitPoolListenAddr("127.0.0.1"); err != nil {
		t.Fatalf("loopback listener rejected: %v", err)
	}
	for _, value := range []string{"", "0.0.0.0", "::", "127.0.0.2"} {
		if err := validateExitPoolListenAddr(value); err == nil {
			t.Fatalf("unsafe listener %q accepted", value)
		}
	}
}

func TestExitPoolRoutingUsesIsolatedProxyAndFailsClosed(t *testing.T) {
	stable := exitPoolSOCKSServer(ExitPoolExit{SourceCategory: exitPoolSourceStable, Port: 20001, SocksUser: "u", SocksPass: "p"})
	if stable["address"] != "127.0.0.1" || stable["port"] != 20001 || stable["users"] == nil {
		t.Fatalf("unexpected stable SOCKS server: %#v", stable)
	}
	extra := exitPoolSOCKSServer(ExitPoolExit{SourceCategory: exitPoolSourceSSTP, ProxyAddress: "8.8.8.8", ProxyPort: 1080})
	if extra["address"] != "8.8.8.8" || extra["port"] != 1080 || extra["users"] != nil {
		t.Fatalf("unexpected supplemental SOCKS server: %#v", extra)
	}
	promoted := exitPoolSOCKSServer(ExitPoolExit{SourceCategory: exitPoolSourceSSTP, Port: 17929, SocksUser: "u", SocksPass: "p"})
	if promoted["address"] != "127.0.0.1" || promoted["port"] != 17929 || promoted["users"] == nil {
		t.Fatalf("unexpected promoted SSTP SOCKS server: %#v", promoted)
	}
	rule := exitPoolRoutingRule("inbound-1", "pool@example", "xnv3-exit-100001")
	if rule["network"] != "tcp,udp" || rule["outboundTag"] != "xnv3-exit-100001" {
		t.Fatalf("routing can bypass selected exit: %#v", rule)
	}
}

func TestExitPoolConsoleOperationAllowed(t *testing.T) {
	tests := []struct {
		method string
		path   string
		want   bool
	}{
		{http.MethodGet, "/api/settings", true},
		{http.MethodPost, "/api/settings", false},
		{http.MethodPost, "/api/update/apply", false},
		{http.MethodPost, "/api/exits/refresh", true},
		{http.MethodPost, "/api/../api/settings", false},
	}
	for _, tt := range tests {
		if got := exitPoolConsoleOperationAllowed(tt.method, tt.path); got != tt.want {
			t.Errorf("%s %s allowed = %v, want %v", tt.method, tt.path, got, tt.want)
		}
	}
}
