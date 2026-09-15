package main

import (
	"reflect"
	"testing"
)

func TestWrapWords(t *testing.T) {
	cases := []struct {
		name  string
		words []string
		width int
		want  []string
	}{
		{"empty", nil, 10, nil},
		{"single short word", []string{"hi"}, 10, []string{"hi"}},
		{
			"word longer than width is never split",
			[]string{"supercalifragilisticexpialidocious"},
			10,
			[]string{"supercalifragilisticexpialidocious"},
		},
		{
			"long word after a short one gets its own line",
			[]string{"short", "supercalifragilisticexpialidocious"},
			10,
			[]string{"short", "supercalifragilisticexpialidocious"},
		},
		{"exact fit stays on one line", []string{"abcde", "fg"}, 8, []string{"abcde fg"}},
		{"one over the width breaks", []string{"abcde", "fgh"}, 8, []string{"abcde", "fgh"}},
		{
			"greedy fill across several words",
			[]string{"the", "quick", "brown", "fox"},
			9,
			[]string{"the quick", "brown fox"},
		},
		{"width <= 0 means unlimited", []string{"a", "b", "c"}, 0, []string{"a b c"}},
		{
			"rune width, not byte width",
			[]string{"héllo", "wörld"},
			11,
			[]string{"héllo wörld"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := wrapWords(c.words, c.width)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("wrapWords(%q, %d) = %q, want %q", c.words, c.width, got, c.want)
			}
		})
	}
}

func TestParseWrapped(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty input", "", nil},
		{"blank input", "\n\n\n", nil},
		{"single line", "hello world", []string{"hello world"}},
		{
			"wrapped lines join with a single space",
			"line one\nline two\n\nsecond para",
			[]string{"line one line two", "second para"},
		},
		{"multiple blank lines still just one break", "a\n\n\n\nb", []string{"a", "b"}},
		{"crlf line endings", "a\r\nb\r\n\r\nc", []string{"a b", "c"}},
		{"tabs and runs of spaces collapse", "a\tb   c", []string{"a b c"}},
		{"leading and trailing blank lines are dropped", "\n\nhello\n\n", []string{"hello"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ParseWrapped(c.input)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("ParseWrapped(%q) = %#v, want %#v", c.input, got, c.want)
			}
		})
	}
}

func TestFormatWrapped(t *testing.T) {
	cases := []struct {
		name  string
		paras []string
		width int
		want  string
	}{
		{"no paragraphs", nil, 72, ""},
		{"one short word per line", []string{"hello world"}, 5, "hello\nworld\n"},
		{"blank line between paragraphs", []string{"a b", "c d"}, 10, "a b\n\nc d\n"},
		{"unsplittable long word", []string{"supercalifragilistic"}, 5, "supercalifragilistic\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := FormatWrapped(c.paras, c.width)
			if got != c.want {
				t.Errorf("FormatWrapped(%q, %d) = %q, want %q", c.paras, c.width, got, c.want)
			}
		})
	}
}

func TestFormatFlowedStuffing(t *testing.T) {
	// Force ">tag" onto its own line so we can check that it gets a
	// stuffed leading space and a soft-break trailing space, while the
	// final, unbroken line of the paragraph gets neither.
	got := FormatFlowed([]string{">tag rest"}, 4)
	want := " >tag \nrest\n"
	if got != want {
		t.Errorf("FormatFlowed stuffing case = %q, want %q", got, want)
	}
}

func TestParseFlowed(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty input", "", nil},
		{"soft break joins with the trailing space", "hello \nworld\n", []string{"hello world"}},
		{"stuffed quote marker is destuffed", " >tag \nrest\n", []string{">tag rest"}},
		{"two fixed lines are two paragraphs", "a\nb\n", []string{"a", "b"}},
		{"stray blank line between paragraphs is harmless", "a\n\nb\n", []string{"a", "b"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ParseFlowed(c.input)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("ParseFlowed(%q) = %#v, want %#v", c.input, got, c.want)
			}
		})
	}
}

// TestFlowedRoundTrip checks that encoding to format=flowed and decoding
// back never loses or mangles a paragraph, across widths and the same
// awkward content covered above.
func TestFlowedRoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		paras []string
		width int
	}{
		{"plain paragraph", []string{"the quick brown fox jumps over the lazy dog"}, 10},
		{"quote-like content", []string{">quoted line of text"}, 6},
		{"single very long word", []string{"supercalifragilisticexpialidocious"}, 8},
		{"unlimited width", []string{"a b c d e"}, 0},
		{"several paragraphs", []string{"first paragraph", "second one", "third and last"}, 12},
		{"unicode content", []string{"héllo wörld", "你好 世界"}, 4},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			flowed := FormatFlowed(c.paras, c.width)
			got := ParseFlowed(flowed)
			if !reflect.DeepEqual(got, c.paras) {
				t.Errorf("round trip at width %d: got %#v, want %#v (flowed text: %q)", c.width, got, c.paras, flowed)
			}
		})
	}
}
