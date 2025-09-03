package main

import (
	"maps"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func testChecksums(t *testing.T, moduleDir string, expectedFilenames []string) {
	cwd, _ := os.Getwd()
	prefix := cwd + "/testdata/"

	checksums, err := packageSourceChecksums("testdata/" + moduleDir)
	if err != nil {
		t.Errorf("failed to calculate source checksums: %v\n", err)
	}

	var actualFilenames []string
	for name := range maps.Keys(checksums) {
		actualFilenames = append(actualFilenames, strings.TrimPrefix(name, prefix))
	}

	sort.Strings(actualFilenames)
	sort.Strings(expectedFilenames)
	if !reflect.DeepEqual(expectedFilenames, actualFilenames) {
		t.Errorf("%v != %v", expectedFilenames, actualFilenames)
	}
}

func TestSourceAndGoMod(t *testing.T) {
	testChecksums(t, "basic", []string{
		"basic/basic.go",
		"basic/go.mod",
	})
}

func TestSourceAndGoModSum(t *testing.T) {
	testChecksums(t, "basic-gosum", []string{
		"basic-gosum/basic.go",
		"basic-gosum/go.mod",
		"basic-gosum/go.sum",
	})
}

func TestExt(t *testing.T) {
	testChecksums(t, "ext", []string{
		"ext/go.mod",
		"ext/go.sum",
		"ext/ext.go",
	})
}

func TestChecksumsSubmodules(t *testing.T) {
	testChecksums(t, "submodule/main", []string{
		"submodule/main/go.mod",
		"submodule/main/intra/go.mod",
		"submodule/main/intra/intra.go",
		"submodule/main/main.go",
		"submodule/neighbour/go.mod",
		"submodule/neighbour/go.sum",
		"submodule/neighbour/neighbour.go",
	})
}

func TestChecksumsInModule(t *testing.T) {
	testChecksums(t, "in-module", []string{
		"in-module/go.mod",
		"in-module/main.go",
		"in-module/inside/go.mod",
		"in-module/inside/inside.go",
	})
}
