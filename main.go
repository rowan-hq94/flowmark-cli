// Command flowmark converts between hard-wrapped plain text and RFC 3676
// format=flowed text.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]
	if cmd != "flow" && cmd != "unflow" {
		usage()
		os.Exit(2)
	}

	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	width := fs.Int("w", 72, "wrap width in columns")
	delsp := fs.Bool("delsp", false, "use RFC 3676 delsp=yes soft-break semantics")
	fs.Parse(os.Args[2:])

	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "flowmark: reading stdin:", err)
		os.Exit(1)
	}

	var out string
	switch cmd {
	case "flow":
		paras := ParseWrapped(string(input))
		if *delsp {
			out = FormatFlowedDelSp(paras, *width)
		} else {
			out = FormatFlowed(paras, *width)
		}
	case "unflow":
		var paras []Para
		if *delsp {
			paras = ParseFlowedDelSp(string(input))
		} else {
			paras = ParseFlowed(string(input))
		}
		out = FormatWrapped(paras, *width)
	}

	if _, err := io.WriteString(os.Stdout, out); err != nil {
		fmt.Fprintln(os.Stderr, "flowmark: writing stdout:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage: flowmark <flow|unflow> [-w width] [-delsp]

  flow    convert hard-wrapped plain text (paragraphs separated by
          blank lines) into RFC 3676 format=flowed text
  unflow  convert format=flowed text back into hard-wrapped plain
          text at the given width

  -w width   wrap width in columns (default 72)
  -delsp     treat the trailing space of a soft break as a marker that
             is deleted on join (delsp=yes) instead of a word separator`)
}
