# gr - instant `go run`

`gr` is a drop-in replacement for `go run` with the following improvements:
- preserves the exit code of the executed tool,
- caches built binaries, so that repeated runs are instantaneous.

Limitations (not yet resolved):
- works only on packages, not on individual files,
- supports only Linux and macOS.

## Usage

### From this repository

Place the [gr](scripts/gr) script into your `$PATH`, and make it executable. On first run,
it will download, and cache a fixed version of `gr`, and then execute it.

Ajustable settings can be found in the script.

### Monorepo

There is no ready-made recipe. Copy `gr` source code into the monorepo and create
a script that checks whether it has been updated by comparison file mtimes.

## Performance

Although Go started caching link results for `go run` and `go tool` in version 1.24,
it still does a remarkably poor job. Benchmarks on a Mac M1 Max (macOS 15.2):

```
% hyperfine --warmup=100 -i 'go tool drozd' 'gr ./tools/drozd'
Benchmark 1: go tool drozd
  Time (mean ± σ):     187.0 ms ±   4.8 ms    [User: 261.7 ms, System: 952.7 ms]
  Range (min … max):   176.6 ms … 195.4 ms    16 runs

Benchmark 2: gr ./tools/drozd
  Time (mean ± σ):      14.4 ms ±   1.6 ms    [User: 7.9 ms, System: 5.2 ms]
  Range (min … max):    13.5 ms …  32.5 ms    168 runs

Summary
  gr ./tools/drozd ran
   12.97 ± 1.47 times faster than go tool drozd
```

## Legal

Copyright Mikhail Gusarov <dottedmag@dottedmag.net>.

Licensed under the terms of the Apache 2.0 license (see `LICENSE-2.0.txt`).
