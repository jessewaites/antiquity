package wizard

import (
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/jessewaites/antiquity/internal/scaffold"
)

// Every placeholder line of the description must carry the faint colour.
func TestPlaceholderLinesShareColour(t *testing.T) {
	m := New("~", scaffold.Case{})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(Model)
	m.step = stepDescription
	m.focus()
	body := m.descView()
	colour := regexp.MustCompile(`38;2;111;115;102`) // #6F7366
	for i, line := range strings.Split(body, "\n") {
		if strings.TrimSpace(ansi.ReplaceAllString(line, "")) == "" {
			continue
		}
		if !colour.MatchString(line) {
			t.Errorf("line %d lacks the faint colour: %q", i, line)
		}
	}
}
