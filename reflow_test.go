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
		want  []Para
	}{
		{"empty input", "", nil},
		{"blank input", "\n\n\n", nil},
		{"single line", "hello world", []Para{{0, "hello world"}}},
		{
			"wrapped lines join with a single space",
			"line one\nline two\n\nsecond para",
			[]Para{{0, "line one line two"}, {0, "second para"}},
		},
		{"multiple blank lines still just one break", "a\n\n\n\nb", []Para{{0, "a"}, {0, "b"}}},
		{"crlf line endings", "a\r\nb\r\n\r\nc", []Para{{0, "a b"}, {0, "c"}}},
		{"tabs and runs of spaces collapse", "a\tb   c", []Para{{0, "a b c"}}},
		{"leading and trailing blank lines are dropped", "\n\nhello\n\n", []Para{{0, "hello"}}},
		{
			"quoted lines join at the same depth",
			"> line one\n> line two",
			[]Para{{1, "line one line two"}},
		},
		{
			"a depth change splits the paragraph without a blank line",
			"> quoted\nunquoted",
			[]Para{{1, "quoted"}, {0, "unquoted"}},
		},
		{
			"nested quote depth",
			">> deeper reply",
			[]Para{{2, "deeper reply"}},
		},
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
		paras []Para
		width int
		want  string
	}{
		{"no paragraphs", nil, 72, ""},
		{"one short word per line", []Para{{0, "hello world"}}, 5, "hello\nworld\n"},
		{"blank line between paragraphs", []Para{{0, "a b"}, {0, "c d"}}, 10, "a b\n\nc d\n"},
		{"unsplittable long word", []Para{{0, "supercalifragilistic"}}, 5, "supercalifragilistic\n"},
		{
			"quoted paragraph gets a prefix on every line",
			[]Para{{1, "a longer reply that wraps"}},
			10,
			"> a longer\n> reply\n> that\n> wraps\n",
		},
		{
			"nested quote prefix",
			[]Para{{2, "short"}},
			10,
			">> short\n",
		},
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
	got := FormatFlowed([]Para{{0, ">tag rest"}}, 4)
	want := " >tag \nrest\n"
	if got != want {
		t.Errorf("FormatFlowed stuffing case = %q, want %q", got, want)
	}
}

func TestFormatFlowedQuoted(t *testing.T) {
	got := FormatFlowed([]Para{{1, "a longer reply that wraps"}}, 10)
	want := "> a longer \n> reply \n> that \n> wraps\n"
	if got != want {
		t.Errorf("FormatFlowed quoted case = %q, want %q", got, want)
	}
}

func TestParseFlowed(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []Para
	}{
		{"empty input", "", nil},
		{"soft break joins with the trailing space", "hello \nworld\n", []Para{{0, "hello world"}}},
		{"stuffed quote marker is destuffed", " >tag \nrest\n", []Para{{0, ">tag rest"}}},
		{"two fixed lines are two paragraphs", "a\nb\n", []Para{{0, "a"}, {0, "b"}}},
		{"stray blank line between paragraphs is harmless", "a\n\nb\n", []Para{{0, "a"}, {0, "b"}}},
		{
			"quoted soft break joins at the same depth",
			"> a longer \n> reply\n",
			[]Para{{1, "a longer reply"}},
		},
		{
			"a depth change ends the paragraph even mid soft-break",
			"> quoted \nunquoted\n",
			[]Para{{1, "quoted"}, {0, "unquoted"}},
		},
		{
			"nested quote depth",
			">> deeper reply\n",
			[]Para{{2, "deeper reply"}},
		},
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
		paras []Para
		width int
	}{
		{"plain paragraph", []Para{{0, "the quick brown fox jumps over the lazy dog"}}, 10},
		{"quote-like content", []Para{{0, ">quoted line of text"}}, 6},
		{"single very long word", []Para{{0, "supercalifragilisticexpialidocious"}}, 8},
		{"unlimited width", []Para{{0, "a b c d e"}}, 0},
		{"several paragraphs", []Para{{0, "first paragraph"}, {0, "second one"}, {0, "third and last"}}, 12},
		{"unicode content", []Para{{0, "héllo wörld"}, {0, "你好 世界"}}, 4},
		{"quoted reply", []Para{{1, "this is a quoted reply that wraps"}}, 10},
		{"nested quoted reply", []Para{{2, "double quoted text"}}, 8},
		{"mixed quote depths", []Para{{0, "top level text"}, {1, "a reply"}, {2, "a nested reply"}}, 8},
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
