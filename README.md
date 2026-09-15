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

## Known limitations

This is a first pass. It does not yet understand RFC 3676's quote-depth
convention (`>` prefixes marking nested email replies) - a `>` at the start
of a line is currently just stuffed like any other awkward character, not
treated as a quote marker that changes paragraph boundaries. See the
roadmap in the project history for what's planned next.
