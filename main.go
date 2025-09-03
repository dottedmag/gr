package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func execProgram(path string, argv0 string, args []string) error {
	return syscall.Exec(path, append([]string{argv0}, args...), os.Environ())
}

func realMain() {
	cli := parseCLI()

	cacheDir := os.Getenv("GR_CACHE_DIR")
	if cacheDir == "" {
		fmt.Fprintf(os.Stderr, "gr: $GR_CACHE_DIR is not set\n")
		return
	}
	if !strings.HasPrefix(cacheDir, "/") {
		fmt.Fprintf(os.Stderr, "gr: cache dir %q is not absolute\n", cacheDir)
		return
	}
	cacheDir = filepath.Join(cacheDir, "pkg")

	absPackagePath, err := filepath.Abs(cli.packagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gr: can't find absolute path for package %q: %v\n", cli.packagePath, err)
		return
	}

	sum, err := checksum(cli.packagePath, cli.compilerFlags, cli.compilerEnv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gr: internal error: can't calculate checksum for package %q: %v\n", cli.packagePath, err)
		return
	}

	p := packageCacheFile(cacheDir, absPackagePath, sum)

	err = execProgram(p, filepath.Base(absPackagePath), cli.runArgs)
	if !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "gr: failed to run program: %v\n", err)
		return
	}

	// The executable didn't exist. Let's build it and try to run again.

	updated, err := updateCache(cacheDir, absPackagePath, sum, cli.compilerFlags, cli.compilerEnv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gr: failed to build program: %v\n", err)
		return
	}

	if !updated {
		return
	}

	err = execProgram(p, filepath.Base(absPackagePath), cli.runArgs)
	fmt.Fprintf(os.Stderr, "gr: failed to run program: %v\n", err)
}

func main() {
	realMain()
	os.Exit(255)
}
