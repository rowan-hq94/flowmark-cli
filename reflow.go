package main

import (
	"strings"
	"unicode/utf8"
)

// Para is one logical paragraph: its text with all internal whitespace
// collapsed, and its RFC 3676 quote depth (the number of '>' markers on
// the lines it came from, 0 for unquoted text).
type Para struct {
	Depth int
	Text  string
}

// wrapWords packs words into lines no wider than width columns, using a
// greedy fit. A word wider than width is never split - it gets a line to
// itself. width <= 0 means unlimited (everything on one line).
func wrapWords(words []string, width int) []string {
	if len(words) == 0 {
		return nil
	}
	var lines []string
	var cur strings.Builder
	curLen := 0
	for _, w := range words {
		wLen := utf8.RuneCountInString(w)
		if curLen == 0 {
			cur.WriteString(w)
			curLen = wLen
			continue
		}
		if width > 0 && curLen+1+wLen > width {
			lines = append(lines, cur.String())
			cur.Reset()
			cur.WriteString(w)
			curLen = wLen
		} else {
			cur.WriteByte(' ')
			cur.WriteString(w)
			curLen += 1 + wLen
		}
	}
	lines = append(lines, cur.String())
	return lines
}

// splitBlank groups consecutive non-blank lines into blocks, treating one
// or more blank lines as a separator. Leading and trailing blank lines are
// dropped.
func splitBlank(input string) []string {
	lines := strings.Split(input, "\n")
	var blocks []string
	var cur []string
	flush := func() {
		if len(cur) > 0 {
			blocks = append(blocks, strings.Join(cur, "\n"))
			cur = nil
		}
	}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			flush()
			continue
		}
		cur = append(cur, line)
	}
	flush()
	return blocks
}

// splitQuoteDepth strips a leading run of '>' characters from line and
// reports how many there were (the RFC 3676 quote depth). If any markers
// were found, a single space directly after them is also consumed - it's
// the mandatory separator a generator always inserts before the quoted
// content, not part of the content itself.
func splitQuoteDepth(line string) (int, string) {
	depth := 0
	for len(line) > 0 && line[0] == '>' {
		depth++
		line = line[1:]
	}
	if depth > 0 && strings.HasPrefix(line, " ") {
		line = line[1:]
	}
	return depth, line
}

// quotePrefix renders the leading marker for a hard-wrapped line at the
// given quote depth: the '>' run plus its mandatory separator space.
func quotePrefix(depth int) string {
	if depth == 0 {
		return ""
	}
	return strings.Repeat(">", depth) + " "
}

// ParseWrapped reads hard-wrapped plain text: paragraphs separated by one
// or more blank lines, with newlines inside a paragraph treated as
// incidental wrap points rather than meaningful breaks. Runs of whitespace
// (including tabs and the newlines themselves) collapse to a single space.
// A leading run of '>' on a line marks its quote depth; a change in depth
// starts a new paragraph even without an intervening blank line, since a
// reply quoted at one level is never the same paragraph as text quoted at
// another.
func ParseWrapped(input string) []Para {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	var paras []Para
	for _, block := range splitBlank(input) {
		curDepth := 0
		var words []string
		flush := func() {
			if len(words) > 0 {
				paras = append(paras, Para{Depth: curDepth, Text: strings.Join(words, " ")})
				words = nil
			}
		}
		for _, line := range strings.Split(block, "\n") {
			depth, rest := splitQuoteDepth(line)
			if len(words) > 0 && depth != curDepth {
				flush()
			}
			curDepth = depth
			words = append(words, strings.Fields(rest)...)
		}
		flush()
	}
	return paras
}

// FormatWrapped renders paragraphs as hard-wrapped plain text at width,
// separating paragraphs with a single blank line. Each line of a quoted
// paragraph (Depth > 0) is given its quote prefix before wrapping is
// measured, so the wrapped text plus prefix never exceeds width.
func FormatWrapped(paras []Para, width int) string {
	var b strings.Builder
	for i, p := range paras {
		prefix := quotePrefix(p.Depth)
		inner := width
		if width > 0 {
			inner = width - utf8.RuneCountInString(prefix)
			if inner < 1 {
				inner = 1
			}
		}
		lines := wrapWords(strings.Fields(p.Text), inner)
		for _, line := range lines {
			b.WriteString(prefix)
			b.WriteString(line)
			b.WriteString("\n")
		}
		if i != len(paras)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// FormatFlowed renders paragraphs as RFC 3676 format=flowed text (delsp=no):
// every line except the last one of a paragraph ends with a single trailing
// space, marking it as a soft break that should be joined with the next
// line rather than treated as the end of a paragraph. A quoted paragraph
// (Depth > 0) gets its '>' run and mandatory separator space before the
// content on every line. A line whose content would otherwise start with
// a space or '>' is space-stuffed so it can't be mistaken for a soft-break
// artifact or a deeper quote marker.
func FormatFlowed(paras []Para, width int) string {
	return formatFlowed(paras, width, false)
}

// FormatFlowedDelSp is FormatFlowed for delsp=yes streams. A reader honouring
// delsp deletes the one trailing space on each soft-broken line when it
// joins lines, so the space that separates two words has to be written
// twice for a single space to survive the round trip.
func FormatFlowedDelSp(paras []Para, width int) string {
	return formatFlowed(paras, width, true)
}

func formatFlowed(paras []Para, width int, delsp bool) string {
	softEnd := " "
	if delsp {
		softEnd = "  "
	}
	var b strings.Builder
	for _, p := range paras {
		words := strings.Fields(p.Text)
		if len(words) == 0 {
			continue
		}
		prefix := strings.Repeat(">", p.Depth)
		sep := ""
		if p.Depth > 0 {
			sep = " "
		}
		inner := width
		if width > 0 {
			inner = width - utf8.RuneCountInString(prefix) - utf8.RuneCountInString(sep)
			if inner < 1 {
				inner = 1
			}
		}
		lines := wrapWords(words, inner)
		for i, line := range lines {
			if strings.HasPrefix(line, " ") || strings.HasPrefix(line, ">") {
				line = " " + line
			}
			full := prefix + sep + line
			if i != len(lines)-1 {
				full += softEnd
			}
			b.WriteString(full)
			b.WriteString("\n")
		}
	}
	return b.String()
}

// ParseFlowed reverses FormatFlowed: a line ending in a space is joined
// directly onto the next line (the trailing space itself is the word
// separator), and a line that doesn't end in a space closes the paragraph.
// Quote depth is read off each line's leading '>' run; a change in depth
// always closes the paragraph in progress, even mid soft-break, since RFC
// 3676 treats a quote depth change as an unconditional boundary.
func ParseFlowed(input string) []Para {
	return parseFlowed(input, false)
}

// ParseFlowedDelSp is ParseFlowed for delsp=yes streams: the trailing space
// of a soft-broken line is a marker only and is dropped when the line is
// joined to the next, instead of serving as the word separator.
func ParseFlowedDelSp(input string) []Para {
	return parseFlowed(input, true)
}

func parseFlowed(input string, delsp bool) []Para {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	input = strings.TrimSuffix(input, "\n")
	if input == "" {
		return nil
	}
	var paras []Para
	var cur strings.Builder
	curDepth := 0
	flush := func() {
		// A paragraph can end while the last line written was a soft
		// break (a quote-depth change cuts it short before the line
		// it would have joined with arrives). Drop that dangling
		// join-separator space rather than leaving it in the text.
		s := strings.TrimSuffix(cur.String(), " ")
		if s != "" {
			paras = append(paras, Para{Depth: curDepth, Text: s})
		}
		cur.Reset()
	}
	for _, line := range strings.Split(input, "\n") {
		soft := strings.HasSuffix(line, " ")
		if soft && delsp {
			// Drop the marker before looking at quote and stuffing
			// spaces, so a line that is only a marker can't be
			// consumed twice.
			line = line[:len(line)-1]
		}
		depth, rest := splitQuoteDepth(line)
		if strings.HasPrefix(rest, " ") {
			rest = rest[1:]
		}
		if cur.Len() > 0 && depth != curDepth {
			flush()
		}
		curDepth = depth
		cur.WriteString(rest)
		if !soft {
			flush()
		}
	}
	flush()
	return paras
}
