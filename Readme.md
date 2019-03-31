# GoRestructure

Gonna change the name.

Check the releases for a win x86_64 build, something something run this binary.

Or just install the deps and `go build main.go`

### Some bugs I already know about:

* Global consts and vars are renamed only in the files they are declared in.

If you define a global, it is only in scope of the file for now. Working on a fix.

* Functions are not renamed.

Working on that as well.

* String obfuscation import, "encoding/hex", not imported in the main file by the module.

Unless your main file already imports "encoding/hex" you will need to add it. That fix is currently my priority.

### Requirements

```$xslt
github.com/davecgh/go-spew/spew
golang.org/x/tools/go/ast/astutil
```