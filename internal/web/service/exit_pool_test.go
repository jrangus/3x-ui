package service

import (
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
