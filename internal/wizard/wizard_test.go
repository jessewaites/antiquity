package wizard

import (
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/jessewaites/antiquity/internal/scaffold"
)

var ansi = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func plain(v tea.View) string { return ansi.ReplaceAllString(v.Content, "") }

func press(m Model, keys ...string) Model {
	for _, k := range keys {
		var msg tea.Msg
		switch k {
		case "enter":
			msg = tea.KeyPressMsg{Code: tea.KeyEnter}
		case "tab":
			msg = tea.KeyPressMsg{Code: tea.KeyTab}
		case "esc":
			msg = tea.KeyPressMsg{Code: tea.KeyEscape}
		case "ctrl+s":
			msg = tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}
		default:
			for _, r := range k {
				next, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
				m = next.(Model)
			}
			continue
		}
		next, _ := m.Update(msg)
		m = next.(Model)
	}
	return m
}

func TestWalkthrough(t *testing.T) {
	m := New("~/Code", scaffold.Case{Git: true})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(Model)

	m = press(m, "enter")
	if m.step != stepName || m.errMsg == "" {
		t.Fatalf("empty name should be refused; step=%d err=%q", m.step, m.errMsg)
	}
	m = press(m, "Bird Sightings", "enter")
	if m.step != stepQuestion {
		t.Fatalf("step=%d", m.step)
	}
	m = press(m, "Pre-1800 bird sightings", "enter")
	m = press(m, "Searching logbooks.", "tab")
	if m.step != stepKeys {
		t.Fatalf("step=%d", m.step)
	}
	m = press(m, "sk-ant-abc", "ctrl+s")
	if m.step != stepConfirm {
		t.Fatalf("step=%d", m.step)
	}
	view := plain(m.View())
	for _, want := range []string{"bird-sightings/", "Pre-1800 bird sightings", "Anthropic", "enter create"} {
		if !strings.Contains(view, want) {
			t.Errorf("confirm screen missing %q:\n%s", want, view)
		}
	}
	m = press(m, "enter")
	if !m.Confirmed || m.Case.Slug != "bird-sightings" || m.Case.Keys["anthropic"] != "sk-ant-abc" || m.Case.Description != "Searching logbooks." {
		t.Errorf("collected case wrong: %+v", m.Case)
	}
	t.Log("\n" + view)
}

func TestEveryScreenRenders(t *testing.T) {
	m := New("~/Code", scaffold.Case{})
	for s := step(0); s < stepCount; s++ {
		m.step = s
		m.focus()
		v := plain(m.View())
		if !strings.Contains(v, "ANTIQUITY") {
			t.Errorf("step %d missing header", s)
		}
		t.Logf("step %d:\n%s", s, v)
	}
}
