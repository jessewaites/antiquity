package intro

// Block letters for the title (figlet "ANSI Shadow" shapes). Face runes are
// █ ▄ ▀; the box-drawing runes form the engraved shadow.
var letters = map[rune][]string{
	'A': {
		" █████╗ ",
		"██╔══██╗",
		"███████║",
		"██╔══██║",
		"██║  ██║",
		"╚═╝  ╚═╝",
	},
	'N': {
		"███╗   ██╗",
		"████╗  ██║",
		"██╔██╗ ██║",
		"██║╚██╗██║",
		"██║ ╚████║",
		"╚═╝  ╚═══╝",
	},
	'T': {
		"████████╗",
		"╚══██╔══╝",
		"   ██║   ",
		"   ██║   ",
		"   ██║   ",
		"   ╚═╝   ",
	},
	'I': {
		"██╗",
		"██║",
		"██║",
		"██║",
		"██║",
		"╚═╝",
	},
	'Q': {
		" ██████╗ ",
		"██╔═══██╗",
		"██║   ██║",
		"██║▄▄ ██║",
		"╚██████╔╝",
		" ╚══▀▀═╝ ",
	},
	'U': {
		"██╗   ██╗",
		"██║   ██║",
		"██║   ██║",
		"██║   ██║",
		"╚██████╔╝",
		" ╚═════╝ ",
	},
	'Y': {
		"██╗   ██╗",
		"╚██╗ ██╔╝",
		" ╚████╔╝ ",
		"  ╚██╔╝  ",
		"   ██║   ",
		"   ╚═╝   ",
	},
}

// titleRows returns the word as rune rows of equal width.
func titleRows(word string) [][]rune {
	rows := make([][]rune, 6)
	for _, ch := range word {
		glyph := letters[ch]
		for r := range rows {
			rows[r] = append(rows[r], []rune(glyph[r])...)
		}
	}
	return rows
}

func isFace(r rune) bool { return r == '█' || r == '▄' || r == '▀' }
