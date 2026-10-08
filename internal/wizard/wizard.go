// Package wizard collects the details of a new case interactively.
package wizard

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jessewaites/antiquity/internal/scaffold"
)

type step int

const (
	stepName step = iota
	stepQuestion
	stepDescription
	stepKeys
	stepConfirm
	stepCount
)

var (
	gold      = lipgloss.Color("#C6A361")
	parchment = lipgloss.Color("#E2D1AB")
	dim       = lipgloss.Color("#8E927E")
	faint     = lipgloss.Color("#6F7366")
	rust      = lipgloss.Color("#D38C6F")
	rule      = lipgloss.Color("#8C774F")

	brandStyle = lipgloss.NewStyle().Foreground(gold).Bold(true)
	stepStyle  = lipgloss.NewStyle().Foreground(faint)
	labelStyle = lipgloss.NewStyle().Foreground(parchment).Bold(true)
	helpStyle  = lipgloss.NewStyle().Foreground(dim)
	hintStyle  = lipgloss.NewStyle().Foreground(faint)
	errStyle   = lipgloss.NewStyle().Foreground(rust)
	keyLabel   = lipgloss.NewStyle().Foreground(dim).Width(16)
	treeStyle  = lipgloss.NewStyle().Foreground(dim)
	valueStyle = lipgloss.NewStyle().Foreground(parchment)
)

// Model is the wizard. After Run, read Case and Confirmed.
type Model struct {
	width, height int
	step          step
	parent        string

	name     textinput.Model
	question textinput.Model
	desc     textarea.Model
	keys     []textinput.Model
	keyIdx   int
	errMsg   string

	git bool

	// Case holds the collected answers. Confirmed is true if the user chose
	// to create the case.
	Case      scaffold.Case
	Confirmed bool
}

// New builds the wizard with any values already known.
func New(parent string, prefill scaffold.Case) Model {
	m := Model{width: 80, height: 24, parent: parent, git: prefill.Git}

	m.name = newInput("bird-sightings")
	m.name.SetValue(prefill.Slug)
	m.name.CharLimit = 48

	m.question = newInput("Pre-1800 bird sightings in Dutch colonial records")
	m.question.SetValue(prefill.Question)
	m.question.CharLimit = 160

	m.desc = textarea.New()
	m.desc.Placeholder = "Why it matters, which sources you suspect hold the answer, what would count as a find, what is already known.\n\ne.g. Searching ship logbooks and Dutch newspapers for dodo sightings after 1660, to see whether any post-date the accepted extinction."
	m.desc.ShowLineNumbers = false
	m.desc.Prompt = ""
	m.desc.CharLimit = 2000
	m.desc.SetHeight(6)
	m.desc.SetVirtualCursor(true)
	m.desc.EndOfBufferCharacter = ' '
	ts := textarea.DefaultDarkStyles()
	ts.Focused.Base = lipgloss.NewStyle()
	ts.Focused.CursorLine = lipgloss.NewStyle()
	ts.Focused.Text = lipgloss.NewStyle().Foreground(parchment)
	ts.Focused.Placeholder = lipgloss.NewStyle().Foreground(faint)
	ts.Blurred = ts.Focused
	m.desc.SetStyles(ts)
	m.desc.SetValue(prefill.Description)

	for _, p := range scaffold.Providers {
		in := newInput("")
		in.EchoMode = textinput.EchoPassword
		in.EchoCharacter = '•'
		in.CharLimit = 256
		if v, ok := prefill.Keys[p.Name]; ok {
			in.SetValue(v)
		}
		m.keys = append(m.keys, in)
	}

	m.focus()
	return m
}

func newInput(placeholder string) textinput.Model {
	in := textinput.New()
	in.Prompt = "› "
	in.Placeholder = placeholder
	in.SetVirtualCursor(true)
	st := textinput.DefaultDarkStyles()
	st.Focused.Prompt = lipgloss.NewStyle().Foreground(gold)
	st.Focused.Text = lipgloss.NewStyle().Foreground(parchment)
	st.Focused.Placeholder = lipgloss.NewStyle().Foreground(faint)
	st.Blurred = st.Focused
	st.Blurred.Prompt = lipgloss.NewStyle().Foreground(faint)
	in.SetStyles(st)
	return in
}

func (m *Model) focus() tea.Cmd {
	m.name.Blur()
	m.question.Blur()
	m.desc.Blur()
	for i := range m.keys {
		m.keys[i].Blur()
	}
	switch m.step {
	case stepName:
		return m.name.Focus()
	case stepQuestion:
		return m.question.Focus()
	case stepDescription:
		return m.desc.Focus()
	case stepKeys:
		return m.keys[m.keyIdx].Focus()
	}
	return nil
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) inputWidth() int { return min(72, max(20, m.width-10)) }

func (m Model) collect() scaffold.Case {
	c := scaffold.Case{
		Slug:        scaffold.Slugify(m.name.Value()),
		Question:    strings.TrimSpace(m.question.Value()),
		Description: strings.TrimSpace(m.desc.Value()),
		Keys:        map[string]string{},
		Git:         m.git,
	}
	for i, p := range scaffold.Providers {
		c.Keys[p.Name] = m.keys[i].Value()
	}
	return c
}

func (m *Model) next() tea.Cmd {
	m.errMsg = ""
	switch m.step {
	case stepName:
		if scaffold.Slugify(m.name.Value()) == "" {
			m.errMsg = "A case needs a name."
			return nil
		}
	case stepQuestion:
		if strings.TrimSpace(m.question.Value()) == "" {
			m.errMsg = "One line is enough, but the agent needs to know what it is looking for."
			return nil
		}
	case stepKeys:
		if m.keyIdx < len(m.keys)-1 {
			m.keyIdx++
			return m.focus()
		}
	}
	m.step++
	m.keyIdx = 0
	return m.focus()
}

func (m *Model) back() tea.Cmd {
	m.errMsg = ""
	if m.step == stepKeys && m.keyIdx > 0 {
		m.keyIdx--
		return m.focus()
	}
	if m.step == 0 {
		return nil
	}
	m.step--
	if m.step == stepKeys {
		m.keyIdx = len(m.keys) - 1
	}
	return m.focus()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		w := m.inputWidth()
		m.name.SetWidth(w)
		m.question.SetWidth(w)
		m.desc.SetWidth(w)
		for i := range m.keys {
			m.keys[i].SetWidth(w)
		}
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.step == 0 {
				return m, tea.Quit
			}
			return m, m.back()
		case "enter":
			if m.step == stepConfirm {
				m.Case = m.collect()
				m.Confirmed = true
				return m, tea.Quit
			}
			if m.step != stepDescription {
				return m, m.next()
			}
		case "tab":
			if m.step == stepDescription || m.step == stepKeys {
				return m, m.next()
			}
			if m.step != stepConfirm {
				return m, m.next()
			}
		case "shift+tab":
			return m, m.back()
		case "ctrl+s":
			if m.step == stepKeys {
				m.step = stepConfirm
				return m, m.focus()
			}
		}
	}

	var cmd tea.Cmd
	switch m.step {
	case stepName:
		m.name, cmd = m.name.Update(msg)
	case stepQuestion:
		m.question, cmd = m.question.Update(msg)
	case stepDescription:
		m.desc, cmd = m.desc.Update(msg)
	case stepKeys:
		m.keys[m.keyIdx], cmd = m.keys[m.keyIdx].Update(msg)
	}
	return m, cmd
}

func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m Model) header() string {
	dots := make([]string, stepCount)
	for i := range dots {
		if step(i) <= m.step {
			dots[i] = lipgloss.NewStyle().Foreground(gold).Render("●")
		} else {
			dots[i] = lipgloss.NewStyle().Foreground(faint).Render("○")
		}
	}
	return brandStyle.Render("ANTIQUITY") + stepStyle.Render("  ·  new case   ") + strings.Join(dots, " ")
}

func (m Model) render() string {
	w, h := max(1, m.width), max(1, m.height)
	var label, help, body, hint string

	switch m.step {
	case stepName:
		label = "What shall we call the case?"
		help = "Short, lowercase, hyphens. It becomes the folder name."
		body = m.name.View()
		if slug := scaffold.Slugify(m.name.Value()); slug != "" {
			body += "\n" + hintStyle.Render("  → "+m.parent+"/"+slug+"/")
		}
		hint = "enter continue   esc quit"
	case stepQuestion:
		label = "What are you investigating?"
		help = "One line. This is the question the agent will carry through every run."
		body = m.question.View()
		hint = "enter continue   esc back"
	case stepDescription:
		label = "Tell the agent more."
		help = "Optional. Goes into AGENTS.md and CASE.md, so the agent starts with your context."
		body = m.desc.View()
		hint = "tab continue   enter newline   esc back"
	case stepKeys:
		label = "API keys, if you have them handy."
		help = "Optional. Leave blank to skip. Written to keys.yml (chmod 600, gitignored); you can fill it in later instead."
		var rows []string
		for i, p := range scaffold.Providers {
			rows = append(rows, keyLabel.Render(p.Label)+m.keys[i].View())
		}
		body = strings.Join(rows, "\n")
		hint = "enter next field   ctrl+s skip all   esc back"
	case stepConfirm:
		c := m.collect()
		label = "Create the case?"
		help = m.parent + "/" + c.Slug + "/"
		var sb strings.Builder
		sb.WriteString(keyLabel.Render("question") + valueStyle.Render(truncate(c.Question, m.inputWidth()-16)) + "\n")
		if c.Description != "" {
			sb.WriteString(keyLabel.Render("description") + valueStyle.Render(truncate(strings.ReplaceAll(c.Description, "\n", " "), m.inputWidth()-16)) + "\n")
		}
		var provided []string
		for _, p := range scaffold.Providers {
			if strings.TrimSpace(c.Keys[p.Name]) != "" {
				provided = append(provided, p.Label)
			}
		}
		keys := "none yet (keys.example.yml will be written)"
		if len(provided) > 0 {
			keys = strings.Join(provided, ", ")
		}
		sb.WriteString(keyLabel.Render("keys") + valueStyle.Render(keys) + "\n")
		git := "yes, with a pre-commit secret check"
		if !c.Git {
			git = "no"
		}
		sb.WriteString(keyLabel.Render("git init") + valueStyle.Render(git) + "\n")
		sb.WriteString(treeStyle.Render(scaffold.Tree(c)))
		body = sb.String()
		hint = "enter create   esc back"
	}

	var sb strings.Builder
	sb.WriteString(m.header() + "\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(rule).Render(strings.Repeat("─", m.inputWidth())) + "\n\n")
	sb.WriteString(labelStyle.Render(label) + "\n")
	sb.WriteString(helpStyle.Width(m.inputWidth()).Render(help) + "\n\n")
	sb.WriteString(body + "\n")
	if m.errMsg != "" {
		sb.WriteString("\n" + errStyle.Render(m.errMsg) + "\n")
	}
	sb.WriteString("\n" + hintStyle.Render(hint))

	content := lipgloss.NewStyle().Width(m.inputWidth()).Render(sb.String())
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content)
}

func truncate(s string, n int) string {
	r := []rune(s)
	if n < 4 || len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

// Describe prints the collected case in plain text (for logs and --dry-run).
func Describe(c scaffold.Case) string {
	return fmt.Sprintf("case %s\nquestion: %s\n", c.Slug, c.Question)
}
