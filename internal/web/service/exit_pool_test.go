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
	if !ok || got != "World-JP-S7" {
		t.Fatalf("display name = %q, %v", got, ok)
	}
	if _, ok := ExitPoolClientDisplayName(model.Client{Group: "staff"}); ok {
		t.Fatal("ordinary client identified as exit-pool client")
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

