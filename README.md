# flowmark

Hard-wrapped plain text - the kind produced by an editor that breaks lines
at 72 or 80 columns - throws away information. Once the text is on disk,
there's no way to tell whether a newline is a real paragraph break or just
where the wrapping happened to land. That means you can't safely reflow it
to a different width: join everything back into logical lines first and
you might glue two separate paragraphs together, or leave a paragraph
broken across several.

[RFC 3676](https://www.rfc-editor.org/rfc/rfc3676) ("format=flowed") solves
this for plain text email by marking soft line breaks explicitly: a line
that ends with a trailing space is a continuation of the same paragraph,
and a line that doesn't is the real end of it. Once the text is in that
form, reflowing to any width is lossless and mechanical.

flowmark converts between the two:

- `flow` reads ordinary hard-wrapped text (paragraphs separated by blank
  lines) and writes format=flowed text.
- `unflow` reads format=flowed text and writes it back out as hard-wrapped
  text at a chosen width.

## Usage

```
$ go build -o flowmark .

$ cat notes.txt
This is a paragraph that
was hard-wrapped by hand
at whatever width the
original editor used.

Second paragraph here.

$ ./flowmark flow -w 72 < notes.txt > notes.flowed
```

`notes.flowed` now holds one paragraph per logical run of lines, with every
line but the last ending in a single trailing space (not visible in a
terminal, but present in the file):

```
This is a paragraph that was hard-wrapped by hand at whatever width␣
the original editor used.
Second paragraph here.
```

Reflowing to a new width is now just a decode/re-encode, with no guessing
about where paragraphs begin and end:

```
$ ./flowmark unflow -w 40 < notes.flowed
This is a paragraph that was
hard-wrapped by hand at whatever width
the original editor used.

Second paragraph here.
```

## Format notes

- A word wider than the wrap width is placed on its own line rather than
  split.
- If a line's content would otherwise start with a space or `>`, it is
  "space-stuffed" with an extra leading space so it can't be confused with
  a soft-break artifact or an email quote marker. This is undone on decode.
- Wrap width is counted in Unicode runes, not bytes and not terminal
  display columns, so combining characters and wide (e.g. CJK) characters
  aren't handled with full visual accuracy yet.
- `unflow`/`flow` use `-w 0` to mean "don't wrap" - one line per paragraph.
- Quote depth (email reply nesting) is understood: a leading run of `>`
  characters on a line marks that paragraph's depth, the wrap width for a
  quoted line is reduced to leave room for its `>` prefix, and a change in
  quote depth always ends a paragraph, even without a blank line between.

## delsp

By default the trailing space on a soft-broken line is the word separator
(`delsp=no`). With `-delsp` it is only a marker that is deleted when lines
are joined (`delsp=yes`). `flow -delsp` writes two trailing spaces on soft
lines so that one space survives the join, and `unflow -delsp` deletes one
trailing space per soft line. Use the same setting in both directions, and
match whatever the other end of the pipe declares in its `DelSp` header
parameter.

## Files

With no file argument flowmark reads stdin and writes stdout. A single
file argument is read instead of stdin (`-` still means stdin), and `-o`
writes to a file instead of stdout. Flags must come before the file name.

```
$ ./flowmark flow -o notes.flowed notes.txt
$ ./flowmark unflow -w 40 -o notes.txt notes.txt
```

The input is read completely before the output file is opened, so `-o` can
name the input file to convert it in place.
