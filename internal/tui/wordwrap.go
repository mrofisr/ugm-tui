package tui

import "strings"

// wordWrap wraps text at the given width, breaking on spaces.
func wordWrap(s string, width int) string {
	if width <= 0 {
		return s
	}
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		col := 0
		for i, word := range strings.Fields(line) {
			wl := len(word)
			if i > 0 && col+1+wl > width {
				b.WriteByte('\n')
				col = 0
			} else if i > 0 {
				b.WriteByte(' ')
				col++
			}
			b.WriteString(word)
			col += wl
		}
	}
	return b.String()
}
