This tool consists of a single package, so if `gr` is in-tree then verifying
that it needs to be rebuilt can be done by checking mtime of go.mod, go.sum
and *.go in this directory.

Caching key is derived from the source code:
- parse `go.mod` to find what's located where,
  taking into account both `import` and `replace` directives
- read all the local source code, follow the imports
- create a checksum using contents of source files,
  `go.mod` and `go.sum` files and compilation options.

Parsing is done lazily to make this tool usable in monorepos.

Instead of using insanesly slow `go list` to locate source code roll our own:
- match `*.go`, `*.S` and CGo files,
- ignore non-regular files,
- ignore `*_test.go`, `.*` and `_*`.
