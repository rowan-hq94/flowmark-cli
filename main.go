// Command flowmark converts between hard-wrapped plain text and RFC 3676
// format=flowed text.
package main

import (
	"errors"
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
	err := run(os.Args[1], os.Args[2:], os.Stdin, os.Stdout)
	switch {
	case err == nil:
	case errors.Is(err, flag.ErrHelp):
	case errors.Is(err, errUsage):
		usage()
		os.Exit(2)
	default:
		fmt.Fprintln(os.Stderr, "flowmark:", err)
		os.Exit(1)
	}
}

var errUsage = errors.New("usage")

// run executes one subcommand. Input comes from the single file argument
// if given (or "-"), otherwise from stdin; output goes to the -o file if
// given, otherwise to stdout.
func run(cmd string, args []string, stdin io.Reader, stdout io.Writer) error {
	if cmd != "flow" && cmd != "unflow" {
		return errUsage
	}

	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	width := fs.Int("w", 72, "wrap width in columns")
	delsp := fs.Bool("delsp", false, "use RFC 3676 delsp=yes soft-break semantics")
	outPath := fs.String("o", "", "write output to this file instead of stdout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 1 {
		return errUsage
	}

	var input []byte
	var err error
	if fs.NArg() == 1 && fs.Arg(0) != "-" {
		input, err = os.ReadFile(fs.Arg(0))
	} else {
		input, err = io.ReadAll(stdin)
	}
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
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

	// The input is fully read before the output file is opened, so -o may
	// name the same file as the input to convert in place.
	if *outPath != "" && *outPath != "-" {
		if err := os.WriteFile(*outPath, []byte(out), 0o644); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		return nil
	}
	if _, err := io.WriteString(stdout, out); err != nil {
		return fmt.Errorf("writing stdout: %w", err)
	}
	return nil
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage: flowmark <flow|unflow> [-w width] [-delsp] [-o file] [file]

  flow    convert hard-wrapped plain text (paragraphs separated by
          blank lines) into RFC 3676 format=flowed text
  unflow  convert format=flowed text back into hard-wrapped plain
          text at the given width

  -w width   wrap width in columns (default 72)
  -delsp     treat the trailing space of a soft break as a marker that
             is deleted on join (delsp=yes) instead of a word separator
  -o file    write output to file instead of stdout; may be the same
             file as the input
  file       read input from file instead of stdin ("-" means stdin)

Flags must come before the file name.`)
}
