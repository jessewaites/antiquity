package intro

import (
	"image/color"
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	tagline   = "AI-assisted historical investigation"
	frameRate = 40 * time.Millisecond

	// Timeline (milliseconds since start).
	tLensStart = 300
	tLensEnd   = 2600
	tRuleEnd   = 3000
	tTagStart  = 3000
	tTagEnd    = 4000
	tCaseStart = 4200
	tCaseEnd   = 5000
	tHint      = 5300
	tDone      = tHint

	padRows = 3   // blank rows above and below the letters so the lens rim shows
	lensRX  = 9.0 // lens radii in cells (terminal cells are ~2:1, so rx ≈ 2·ry)
	lensRY  = 4.6
)

// Palette: parchment, gold, umber; steel for the lens.
var (
	faceDark  = rgb{0xB8, 0x92, 0x52}
	faceLight = rgb{0xEA, 0xDA, 0xB4}
	faceGleam = rgb{0xFF, 0xF4, 0xDC}
	shadowCol = rgb{0x5E, 0x4A, 0x2E}
	blurCol   = rgb{0x3E, 0x36, 0x2C}
	blurShad  = rgb{0x2E, 0x29, 0x22}
	rimCol    = rgb{0xA9, 0xB4, 0xBC}
	handleCol = rgb{0x7A, 0x5A, 0x3A}
	glassBG   = rgb{0x1C, 0x22, 0x27}

	ruleStyle  = fg(rgb{0x8C, 0x77, 0x4F})
	gemStyle   = fg(rgb{0xC6, 0xA3, 0x61})
	dimStyle   = fg(rgb{0x8E, 0x92, 0x7E})
	labelStyle = fg(rgb{0x6F, 0x73, 0x66})
	nameStyle  = fg(rgb{0xE2, 0xD1, 0xAB}).Bold(true)
)

type rgb struct{ r, g, b uint8 }

func (c rgb) RGBA() (uint32, uint32, uint32, uint32) {
	return color.RGBA{c.r, c.g, c.b, 0xFF}.RGBA()
}

func (c rgb) mix(o rgb, t float64) rgb {
	t = min(1, max(0, t))
	f := func(a, b uint8) uint8 { return uint8(float64(a) + (float64(b)-float64(a))*t) }
	return rgb{f(c.r, o.r), f(c.g, o.g), f(c.b, o.b)}
}

func fg(c rgb) lipgloss.Style { return lipgloss.NewStyle().Foreground(c) }

type frame time.Time

type Model struct {
	width, height int
	caseName      string
	paused        bool
	ticks         int
	// Proceed is true when the user pressed enter on a finished screen.
	Proceed bool

	title [][]rune // letters only, no padding
}

// New builds the intro. caseName may be empty (plain title screen).
func New(caseName string, still bool) Model {
	m := Model{width: 80, height: 24, caseName: caseName, paused: still, title: titleRows("ANTIQUITY")}
	if still {
		m.ticks = doneTicks()
	}
	return m
}

func doneTicks() int { return tDone/int(frameRate/time.Millisecond) + 1 }

func tick() tea.Cmd {
	return tea.Tick(frameRate, func(t time.Time) tea.Msg { return frame(t) })
}

func (m Model) Init() tea.Cmd { return tick() }

func (m Model) elapsed() int { return m.ticks * int(frameRate/time.Millisecond) }

func (m Model) finished() bool { return m.elapsed() >= tDone }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if !m.finished() {
				m.ticks = doneTicks()
				return m, nil
			}
			m.Proceed = true
			return m, tea.Quit
		case "r":
			m.ticks = 0
			m.paused = false
		case "space", " ":
			m.paused = !m.paused
		}
	case frame:
		if !m.paused {
			m.ticks++
		}
		return m, tick()
	}
	return m, nil
}

func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

// Snapshot renders the frame at the given size and time offset (ms).
func (m Model) Snapshot(width, height, ms int) string {
	m.width, m.height = width, height
	m.ticks = (ms + int(frameRate/time.Millisecond) - 1) / int(frameRate/time.Millisecond)
	return m.render()
}

// hash is a small deterministic noise source.
func hash(a, b, c int) int {
	h := uint32(a)*0x9E3779B1 ^ uint32(b)*0x85EBCA77 ^ uint32(c)*0xC2B2AE3D
	h ^= h >> 15
	h *= 0x2C1B3C6D
	h ^= h >> 12
	return int(h & 0x7fffffff)
}

func lensDist(dx, dy float64) float64 {
	return math.Sqrt((dx/lensRX)*(dx/lensRX) + (dy/lensRY)*(dy/lensRY))
}

// rimRune picks a box-drawing rune that follows the lens rim at the given angle.
func rimRune(dx, dy float64) rune {
	a := math.Atan2(dy, dx) // y grows downward
	if a < 0 {
		a += 2 * math.Pi
	}
	oct := int(math.Round(a/(math.Pi/4))) % 8
	switch oct {
	case 0, 4:
		return '│'
	case 2, 6:
		return '─'
	case 1, 5: // bottom-right, top-left
		return '╱'
	default: // bottom-left, top-right
		return '╲'
	}
}

func (m Model) renderTitle() string {
	ms := m.elapsed()
	w := len(m.title[0])
	letterRows := len(m.title)
	rows := letterRows + 2*padRows

	// Lens centre travels left→right across the inscription.
	span := float64(w) + 2*lensRX + 8
	p := float64(ms-tLensStart) / float64(tLensEnd-tLensStart)
	cx := p*span - lensRX - 4
	cy := float64(rows-1) / 2
	lensActive := ms >= tLensStart && ms < tLensEnd+300

	// Gleam: a soft highlight band sweeps the inscription every ~5 s once focused.
	gleamAt := -100.0
	if ms >= tRuleEnd {
		cycle := (ms - tRuleEnd) % 5000
		if cycle < 1400 {
			gleamAt = float64(cycle)/1400*float64(w+16) - 8
		}
	}

	var sb strings.Builder
	for r := 0; r < rows; r++ {
		lr := r - padRows
		for c := 0; c < w; c++ {
			ch := ' '
			if lr >= 0 && lr < letterRows {
				ch = m.title[lr][c]
			}
			dx, dy := float64(c)-cx, float64(r)-cy
			d := lensDist(dx, dy)
			// Rim: a one-cell outline where an inside cell touches the outside.
			onRim := d <= 1 && (lensDist(dx-1, dy) > 1 || lensDist(dx+1, dy) > 1 || lensDist(dx, dy-1) > 1 || lensDist(dx, dy+1) > 1)
			inside := d <= 1 && !onRim

			// Handle: two cells of ╲ off the lower-right of the rim.
			if lensActive {
				hx := int(math.Round(cx + lensRX*0.72))
				hy := int(math.Round(cy + lensRY*0.72))
				if (c == hx+1 && r == hy+1) || (c == hx+2 && r == hy+2) {
					sb.WriteString(fg(handleCol).Render("╲"))
					continue
				}
			}

			switch {
			case lensActive && onRim:
				sb.WriteString(fg(rimCol).Background(glassBG).Render(string(rimRune(dx, dy))))
			case ch == ' ':
				if lensActive && inside {
					sb.WriteString(lipgloss.NewStyle().Background(glassBG).Render(" "))
				} else {
					sb.WriteByte(' ')
				}
			case ms < tLensStart || (lensActive && d > 1 && dx > 0) || (!lensActive && ms < tLensEnd):
				// Out of focus: smeared, dim, flickering.
				if isFace(ch) {
					g := []string{"░", "▒", "░", "▒"}[hash(c, r, m.ticks/3)%4]
					sb.WriteString(fg(blurCol).Render(g))
				} else if hash(c, r, 2)%3 == 0 {
					sb.WriteString(fg(blurShad).Render("░"))
				} else {
					sb.WriteByte(' ')
				}
			default:
				// In focus.
				st := lipgloss.NewStyle()
				if lensActive && inside {
					st = st.Background(glassBG)
				}
				if isFace(ch) {
					t := float64(lr) / float64(letterRows-1)
					col := faceLight.mix(faceDark, t*0.85)
					if lensActive && inside {
						col = col.mix(faceGleam, 0.35)
					}
					if g := float64(c) - gleamAt; g > -5 && g < 5 {
						col = col.mix(faceGleam, 1-(g*g)/25)
					}
					sb.WriteString(st.Foreground(col).Render(string(ch)))
				} else {
					sb.WriteString(st.Foreground(shadowCol).Render(string(ch)))
				}
			}
		}
		if r < rows-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

func (m Model) renderRule(width int) string {
	ms := m.elapsed()
	if ms < tLensEnd {
		return ""
	}
	half := (width - 3) / 2
	p := float64(ms-tLensEnd) / float64(tRuleEnd-tLensEnd)
	n := int(min(1, p) * float64(half))
	right := strings.Repeat("─", n) + strings.Repeat(" ", half-n)
	left := strings.Repeat(" ", half-n) + strings.Repeat("─", n)
	return ruleStyle.Render(left) + " " + gemStyle.Render("◆") + " " + ruleStyle.Render(right)
}

func typed(text string, ms, start, end int) string {
	if ms < start {
		return ""
	}
	p := float64(ms-start) / float64(end-start)
	n := int(min(1, p) * float64(len([]rune(text))))
	return string([]rune(text)[:n])
}

func (m Model) render() string {
	w, h := max(1, m.width), max(1, m.height)
	ms := m.elapsed()
	center := func(s string) string { return lipgloss.PlaceHorizontal(w, lipgloss.Center, s) }

	titleW := len(m.title[0])
	var title string
	if w < titleW+4 || h < 22 {
		title = nameStyle.Render("A N T I Q U I T Y")
	} else {
		title = m.renderTitle()
	}

	rule := m.renderRule(min(titleW-10, w-6))
	tag := dimStyle.Render(typed(tagline, ms, tTagStart, tTagEnd))

	cursor := " "
	if ms >= tCaseStart && (m.ticks/6)%2 == 0 {
		cursor = gemStyle.Render("▌")
	}
	var caseLine string
	if m.caseName != "" && ms >= tCaseStart {
		name := typed(m.caseName, ms, tCaseStart, tCaseEnd)
		caseLine = labelStyle.Render("case ") + " " + nameStyle.Render(name) + cursor
	}

	hint := ""
	if ms >= tHint {
		switch {
		case m.paused:
			hint = "space resume   r replay   q quit"
		case m.caseName != "":
			hint = "enter begin   r replay   q quit"
		default:
			hint = "r replay   q quit"
		}
	}

	parts := []string{
		center(title),
		center(rule),
		"",
		center(tag),
		"",
		center(caseLine),
		"",
		center(labelStyle.Render(hint)),
	}
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, strings.Join(parts, "\n"))
}
