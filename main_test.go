package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunStdinStdout(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("one two\nthree\n")
	if err := run("flow", []string{"-w", "72"}, in, &out); err != nil {
		t.Fatal(err)
	}
	want := FormatFlowed(ParseWrapped("one two\nthree\n"), 72)
	if out.String() != want {
		t.Errorf("got %q, want %q", out.String(), want)
	}
}

func TestRunFileToFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "in.txt")
	dst := filepath.Join(dir, "out.txt")
	text := "alpha beta\ngamma\n\nsecond paragraph\n"
	if err := os.WriteFile(src, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	if err := run("flow", []string{"-o", dst, src}, strings.NewReader(""), &stdout); err != nil {
		t.Fatal(err)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout should be empty with -o, got %q", stdout.String())
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if want := FormatFlowed(ParseWrapped(text), 72); string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRunInPlace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "doc.txt")
	text := "alpha beta\ngamma\n"
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run("flow", []string{"-o", path, path}, strings.NewReader(""), &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := FormatFlowed(ParseWrapped(text), 72); string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRunDashIsStdin(t *testing.T) {
	var out bytes.Buffer
	if err := run("unflow", []string{"-"}, strings.NewReader("hello\n"), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "hello") {
		t.Errorf("got %q", out.String())
	}
}

func TestRunMissingInput(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope.txt")
	err := run("flow", []string{missing}, strings.NewReader(""), &bytes.Buffer{})
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Errorf("want not-exist error, got %v", err)
	}
}

func TestRunUsageErrors(t *testing.T) {
	cases := [][]string{
		{"bogus"},
		{"flow", "a.txt", "b.txt"},
	}
	for _, c := range cases {
		err := run(c[0], c[1:], strings.NewReader(""), &bytes.Buffer{})
		if !errors.Is(err, errUsage) {
			t.Errorf("run(%v) = %v, want usage error", c, err)
		}
	}
}
