package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Bird Sightings":        "bird-sightings",
		"  Dodo!! photos 1660 ": "dodo-photos-1660",
		"test-mystery":          "test-mystery",
		"--x--":                 "x",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWrite(t *testing.T) {
	parent := t.TempDir()
	c := Case{
		Slug:        "bird-sightings",
		Question:    "Pre-1800 bird sightings in Dutch colonial records",
		Description: "Searching ship logbooks for dodo sightings after 1660.",
		Keys:        map[string]string{"anthropic": "sk-ant-test", "openai": ""},
		Git:         true,
	}
	dir, err := Write(parent, c)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"AGENTS.md", "CASE.md", "HANDOFF.md", "NARRATIVE.md", "lessons.md", ".gitignore", "keys.example.yml", "keys.yml", "templates/FINDING.md", "templates/source.md", "templates/judge.md", "sources/README.md", "evidence/README.md"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("missing %s", f)
		}
	}
	agents, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	for _, want := range []string{c.Question, c.Description, "Keys were provided at setup for: anthropic."} {
		if !strings.Contains(string(agents), want) {
			t.Errorf("AGENTS.md missing %q", want)
		}
	}
	keys, _ := os.ReadFile(filepath.Join(dir, "keys.yml"))
	if !strings.Contains(string(keys), `anthropic: {api_key: "sk-ant-test"}`) || strings.Contains(string(keys), "openai") {
		t.Errorf("keys.yml wrong:\n%s", keys)
	}
	if info, _ := os.Stat(filepath.Join(dir, "keys.yml")); info.Mode().Perm() != 0o600 {
		t.Errorf("keys.yml mode = %o", info.Mode().Perm())
	}
	if info, err := os.Stat(filepath.Join(dir, ".git", "hooks", "pre-commit")); err != nil || info.Mode().Perm()&0o100 == 0 {
		t.Errorf("pre-commit hook missing or not executable")
	}
	if _, err := Write(parent, c); err == nil {
		t.Error("second Write should fail on non-empty dir")
	}
}

func TestWriteNoKeys(t *testing.T) {
	dir, err := Write(t.TempDir(), Case{Slug: "x", Question: "q"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "keys.yml")); err == nil {
		t.Error("keys.yml should not exist without keys")
	}
}
