package main

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type sut struct {
	dir         string
	exe         string
	coverageDir string
}

func must(t *testing.T, err error) {
	if err != nil {
		t.Errorf("%v", err)
	}
}

func mustBuildSUT(t *testing.T) sut {
	dir := t.TempDir()
	sut := sut{
		dir:         dir,
		exe:         filepath.Join(dir, "exe"),
		coverageDir: filepath.Join(dir, "coverage"),
	}

	compileCmd := exec.Command("go", "build", "-trimpath", "-buildvcs=false", "-o", sut.exe)
	if testing.CoverMode() != "" {
		compileCmd.Args = append(compileCmd.Args, "-cover")
		os.MkdirAll(sut.coverageDir, 0o755)
	}
	compileCmd.Args = append(compileCmd.Args, ".")
	compileCmd.Stdout = os.Stdout
	compileCmd.Stderr = os.Stderr
	must(t, compileCmd.Run())
	return sut
}

func (sut sut) done(t *testing.T) {
	if testing.CoverMode() != "" {
		must(t, os.MkdirAll(sut.coverageDir+"-all", 0o755))
		mergeCoverageCmd := exec.Command("go", "tool", "covdata", "merge", "-i="+sut.coverageDir, "-o="+sut.coverageDir+"-all")
		mergeCoverageCmd.Stdout = os.Stdout
		mergeCoverageCmd.Stderr = os.Stderr
		must(t, mergeCoverageCmd.Run())

		convertCoverageCmd := exec.Command("go", "tool", "covdata", "textfmt", "-i="+sut.coverageDir+"-all", "-o=coverage.txt")
		convertCoverageCmd.Stdout = os.Stdout
		convertCoverageCmd.Stderr = os.Stderr
		must(t, convertCoverageCmd.Run())
	}
}

func (sut sut) run(t *testing.T, args []string, env []string) (retStdout string, retStderr string, retExitCode int, _ error) {
	exe := filepath.Join(sut.dir, "exe")

	runCmd := exec.Command(exe, args...)
	runCmd.Env = append(os.Environ(), "HOME="+sut.dir) // Make sure every test case gets a separate cache
	runCmd.Env = append(runCmd.Env, env...)
	if testing.CoverMode() != "" {
		coverageDir := filepath.Join(sut.dir, "coverage")
		must(t, os.MkdirAll(coverageDir, 0o755))
		runCmd.Env = append(runCmd.Env, "GOCOVERDIR="+coverageDir)
	}

	// Make sure module cache does not affect the output
	modCacheDir := t.TempDir()
	runCmd.Env = append(runCmd.Env, "GOMODCACHE="+modCacheDir)
	// Cache directories are created with mode 555, so their permissions need to be adjusted before cleanup
	defer func() {
		must(t, filepath.WalkDir(modCacheDir, func(path string, de fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if de.IsDir() {
				must(t, os.Chmod(path, 0o755))
			}
			return nil
		}))
	}()

	stdoutPipe, err := runCmd.StdoutPipe()
	must(t, err)
	stderrPipe, err := runCmd.StderrPipe()
	must(t, err)

	must(t, runCmd.Start())

	stdout, err := io.ReadAll(stdoutPipe)
	must(t, err)
	stderr, err := io.ReadAll(stderrPipe)
	must(t, err)

	if err := runCmd.Wait(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return string(stdout), string(stderr), exitErr.ExitCode(), nil
		}
		t.Fatalf("Failed to run command: %v", err)
	}

	return string(stdout), string(stderr), 0, nil
}

var anything = regexp.MustCompile(``)

type cliTestCase struct {
	args []string
	env  []string

	exitCode int

	stdout   string
	stdoutRx *regexp.Regexp // has priority over stdout

	stderr   string
	stderrRx *regexp.Regexp // has priority over stderr
}

func cliTestCaseName(tc cliTestCase) string {
	s := "gr " + strings.Join(tc.args, " ")
	if len(tc.env) > 0 {
		s = strings.Join(tc.env, " ") + " " + s
	}
	return s
}

func TestCLI(t *testing.T) {
	sut := mustBuildSUT(t)
	defer sut.done(t)

	for _, tc := range []cliTestCase{
		{exitCode: 2, stderrRx: anything}, // no args -> usage
		{args: []string{"./testdata/basic"}, stdout: "Hello world!\n"},
		{args: []string{"./testdata/basic"}, stdout: "Hello world!\n"}, // run twice
		{args: []string{"./testdata/ext"}, stdout: "Hello world!\n", stderr: "go: downloading github.com/dottedmag/must v1.0.0\n"},
		{args: []string{"./testdata/exit3"}, exitCode: 3},

		// Run even if required module is erroneously marked as indirect
		{args: []string{"./testdata/wrong-module-indirect"}, stdout: "Hello world!\n", stderr: "go: downloading golang.org/x/crypto v0.27.0\n"},

		// Compilation failures
		{args: []string{"./testdata/syntax-error"}, exitCode: 255, stderrRx: regexp.MustCompile(`undefined: fmt\.Printz`)},
		// Weird things
		{args: []string{"./testdata/basic"}, env: []string{"HOME="}, exitCode: 255, stderrRx: regexp.MustCompile(`gr: can't run:`)},
	} {
		t.Run(cliTestCaseName(tc), func(t *testing.T) {
			stdout, stderr, exitCode, err := sut.run(t, tc.args, tc.env)
			must(t, err)

			if tc.exitCode != exitCode {
				t.Errorf("Expected exit code %d, got %d", tc.exitCode, exitCode)
			}

			if tc.stdoutRx != nil {
				if !tc.stdoutRx.MatchString(stdout) {
					t.Fatalf("failed to match %#q regexp against %q", tc.stdoutRx, stdout)
				}
			} else {
				if tc.stdout != stdout {
					t.Errorf("%s != %s", tc.stdout, stdout)
				}
			}

			if tc.stderrRx != nil {
				if !tc.stderrRx.MatchString(stderr) {
					t.Fatalf("failed to match %#q regexp against %q", tc.stderrRx, stderr)
				}
			} else {
				if tc.stdout != stdout {
					t.Errorf("%s != %s", tc.stdout, stdout)
				}
			}
		})
	}
}
