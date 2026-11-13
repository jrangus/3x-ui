package main

import (
	"os"
	"strings"
	"testing"
)

func TestForkUpdateChannelUsesImmutableDatedRelease(t *testing.T) {
	update, err := os.ReadFile("update.sh")
	if err != nil {
		t.Fatal(err)
	}
	text := string(update)
	for _, required := range []string{
		`download_tag="${expected_panel_version}"`,
		`script_ref="${download_tag}"`,
		`releases/download/${download_tag}/x-ui-linux-$(arch).tar.gz`,
		`Release identity mismatch: requested`,
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("update.sh is missing immutable-channel guard %q", required)
		}
	}
	for _, forbidden := range []string{
		"github.com/MHSanaei/3x-ui/releases/download/",
		`script_ref="main"`,
		`releases/download/${tag_version}/x-ui-linux-$(arch).tar.gz`,
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("update.sh still contains mutable or upstream path %q", forbidden)
		}
	}
}

func TestDatedPanelUsesItsOwnPublishedScripts(t *testing.T) {
	menu, err := os.ReadFile("x-ui.sh")
	if err != nil {
		t.Fatal(err)
	}
	text := string(menu)
	if !strings.Contains(text, `3x-ui-[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]) ref="${version}"`) {
		t.Fatal("dated release does not select its own immutable update script")
	}
	if !strings.Contains(text, "refusing to mix it with main") {
		t.Fatal("menu may still silently fall back to main")
	}
}

func TestReleaseWorkflowVerifiesBuiltIdentity(t *testing.T) {
	workflow, err := os.ReadFile(".github/workflows/release.yml")
	if err != nil {
		t.Fatal(err)
	}
	text := string(workflow)
	if strings.Count(text, `release identity mismatch: requested=$RELEASE_TAG built=$BUILT_VERSION`) != 2 {
		t.Fatal("Linux and Windows release builds must both verify the embedded release identity")
	}
}
