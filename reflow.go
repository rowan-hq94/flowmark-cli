package main

import (
	"strings"
	"unicode/utf8"
)

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

// ParseWrapped reads hard-wrapped plain text: paragraphs separated by one
// or more blank lines, with newlines inside a paragraph treated as
// incidental wrap points rather than meaningful breaks. Runs of whitespace
// (including tabs and the newlines themselves) collapse to a single space.
func ParseWrapped(input string) []string {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	var paras []string
	for _, block := range splitBlank(input) {
		words := strings.Fields(block)
		if len(words) == 0 {
			continue
		}
		paras = append(paras, strings.Join(words, " "))
	}
	return paras
}

// FormatWrapped renders paragraphs as hard-wrapped plain text at width,
// separating paragraphs with a single blank line.
func FormatWrapped(paras []string, width int) string {
	var b strings.Builder
	for i, p := range paras {
		lines := wrapWords(strings.Fields(p), width)
		b.WriteString(strings.Join(lines, "\n"))
		b.WriteString("\n")
		if i != len(paras)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// FormatFlowed renders paragraphs as RFC 3676 format=flowed text (delsp=no):
// every line except the last one of a paragraph ends with a single trailing
// space, marking it as a soft break that should be joined with the next
// line rather than treated as the end of a paragraph. A line whose content
// would otherwise start with a space or '>' is space-stuffed so it can't be
// mistaken for a quote marker or an accidental leading space.
func FormatFlowed(paras []string, width int) string {
	var b strings.Builder
	for _, p := range paras {
		words := strings.Fields(p)
		if len(words) == 0 {
			continue
		}
		lines := wrapWords(words, width)
		for i, line := range lines {
			if strings.HasPrefix(line, " ") || strings.HasPrefix(line, ">") {
				line = " " + line
			}
			if i != len(lines)-1 {
				line += " "
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	return b.String()
}

// ParseFlowed reverses FormatFlowed: a line ending in a space is joined
// directly onto the next line (the trailing space itself is the word
// separator), and a line that doesn't end in a space closes the paragraph.
// One leading stuffed space is stripped from every line before joining.
func ParseFlowed(input string) []string {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	input = strings.TrimSuffix(input, "\n")
	if input == "" {
		return nil
	}
	var paras []string
	var cur strings.Builder
	for _, line := range strings.Split(input, "\n") {
		soft := strings.HasSuffix(line, " ")
		content := line
		if strings.HasPrefix(content, " ") {
			content = content[1:]
		}
		cur.WriteString(content)
		if !soft {
			if s := cur.String(); s != "" {
				paras = append(paras, s)
			}
			cur.Reset()
		}
	}
	if s := cur.String(); s != "" {
		paras = append(paras, s)
	}
	return paras
}
