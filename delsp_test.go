package main

import (
	"reflect"
	"testing"
)

func TestFormatFlowedDelSp(t *testing.T) {
	got := FormatFlowedDelSp([]Para{{0, "aa bb cc dd"}}, 5)
	want := "aa bb  \ncc dd\n"
	if got != want {
		t.Errorf("FormatFlowedDelSp = %q, want %q", got, want)
	}
}

func TestParseFlowedDelSp(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []Para
	}{
		{"marker space deleted", "abc \ndef\n", []Para{{0, "abcdef"}}},
		{"doubled space keeps one", "abc  \ndef\n", []Para{{0, "abc def"}}},
		{"quoted", "> abc  \n> def\n", []Para{{1, "abc def"}}},
		{"marker-only line", " \nabc\n", []Para{{0, "abc"}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ParseFlowedDelSp(c.input)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("ParseFlowedDelSp(%q) = %#v, want %#v", c.input, got, c.want)
			}
		})
	}
}

func TestFlowedDelSpRoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		paras []Para
		width int
	}{
		{"plain paragraph", []Para{{0, "the quick brown fox jumps over the lazy dog"}}, 10},
		{"quote-like content", []Para{{0, ">quoted line of text"}}, 6},
		{"several paragraphs", []Para{{0, "first paragraph"}, {0, "second one"}, {0, "third and last"}}, 12},
		{"mixed quote depths", []Para{{0, "top level text"}, {1, "a reply"}, {2, "a nested reply"}}, 8},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			flowed := FormatFlowedDelSp(c.paras, c.width)
			got := ParseFlowedDelSp(flowed)
			if !reflect.DeepEqual(got, c.paras) {
				t.Errorf("round trip at width %d: got %#v, want %#v (flowed text: %q)", c.width, got, c.paras, flowed)
			}
		})
	}
}
