// Copyright 2026 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/require"
)

func writeSource(t *testing.T, source, name, content string) {
	t.Helper()
	filename := filepath.Join(source, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0755))
	require.NoError(t, os.WriteFile(filename, []byte(content), 0644))
}

func TestErrdocIgnore(t *testing.T) {
	source := t.TempDir()
	writeSource(t, source, "errors.go", "package example; var ErrRoot = new(error)")
	writeSource(t, source, "skip/errors.go", "package example; var ErrSkipped = new(error)")
	writeSource(t, source, "skip/nested/errors.go", "package example; var ErrNested = new(error)")
	writeSource(t, source, autoDirectoryName+"/main.go", "invalid retained generator")
	previousIgnore := opt.ignore
	t.Cleanup(func() { opt.ignore = previousIgnore })
	for _, test := range []struct {
		ignore string
		count  int
	}{
		{"", 3},
		{", ,", 3},
		{"skip", 1},
		{"./skip", 1},
		{"./skip/", 1},
		{"skip/,skip/nested", 1},
		{" , skip, , ", 1},
	} {
		t.Run(test.ignore, func(t *testing.T) {
			opt.ignore = test.ignore
			decls, err := errdoc(source, "example.com/project")
			require.NoError(t, err)
			require.Len(t, decls, test.count)
		})
	}
}

func TestErrdocScriptArguments(t *testing.T) {
	previousIgnore := opt.ignore
	t.Cleanup(func() { opt.ignore = previousIgnore })
	for _, test := range []struct {
		module string
		ignore string
	}{
		{"github.com/pingcap/tidb", "pkg/parser,pkg/tidb-binlog"},
		{"github.com/pingcap/tiflow", "proto,dm,deployments,engine,cdc/scheduler/internal"},
		{"github.com/tikv/pd", "./client/errs,./tools"},
	} {
		t.Run(test.module, func(t *testing.T) {
			source := t.TempDir()
			writeSource(t, source, "errors.go", "package example; var ErrRoot = new(error)")
			for _, dir := range strings.Split(test.ignore, ",") {
				writeSource(t, source, filepath.Join(dir, "invalid.go"), "invalid Go")
			}
			opt.ignore = test.ignore
			decls, err := errdoc(source, test.module)
			require.NoError(t, err)
			require.Len(t, decls, 1)
			require.Equal(t, test.module, decls[0].ImportPath)
		})
	}
}

func TestErrdocInvalidIgnore(t *testing.T) {
	source := t.TempDir()
	writeSource(t, source, "errors.go", "package example; var ErrRoot = new(error)")
	previousIgnore := opt.ignore
	t.Cleanup(func() { opt.ignore = previousIgnore })
	for _, ignore := range []string{
		".", "./", "..", "../outside", "skip/../outside", "skip/./nested", "skip//nested",
		"/absolute", "C:\\absolute", "skip\\nested", "with space", "with\nnewline", "with\"quote",
		"with`backtick", "with;semicolon", "skip/..hidden",
	} {
		t.Run(ignore, func(t *testing.T) {
			opt.ignore = "valid," + ignore
			decls, err := errdoc(source, "example.com/project")
			require.ErrorContains(t, err, "invalid --ignore directory")
			require.Nil(t, decls)
		})
	}
}

func TestErrdocPackageAliases(t *testing.T) {
	source := t.TempDir()
	directories := []string{".", "a/b", "a_b", "a-b", "v1.2", "123", "fmt", "for"}
	for _, dir := range directories {
		writeSource(t, source, filepath.Join(dir, "errors.go"), "package example; var ErrExample = new(error)")
	}
	writeSource(t, source, "more.go", "package example; var ErrAnother = new(error)")
	decls, err := errdoc(source, "example.com/project")
	require.NoError(t, err)
	require.Len(t, decls, len(directories))
	aliases := make(map[string]bool)
	imports := make(map[string][]string)
	for _, decl := range decls {
		require.True(t, token.IsIdentifier(decl.PackageName), decl.PackageName)
		require.False(t, aliases[decl.PackageName], "duplicate alias %s", decl.PackageName)
		aliases[decl.PackageName] = true
		imports[decl.ImportPath] = decl.ErrNames
	}
	require.ElementsMatch(t, []string{"ErrExample", "ErrAnother"}, imports["example.com/project"])
	for _, dir := range directories[1:] {
		require.Equal(t, []string{"ErrExample"}, imports["example.com/project/"+dir])
	}
}

func TestErrdocInvalidModule(t *testing.T) {
	source := t.TempDir()
	writeSource(t, source, "errors.go", "package example; var ErrExample = new(error)")
	for _, module := range []string{
		"", "example.com/project\"\n)\nfunc init() {}\n//", "example.com/with space",
		"example.com/with`backtick", "example.com/with\\backslash", "/absolute", "../relative",
		"example.com/../escape", "example.com//empty", "example.com/trailing/", "-example.com/project",
		"example..com/project", "example.com/a..b", "example.com/.hidden",
		"example.com/a~b", "example.com/a+b",
	} {
		t.Run(module, func(t *testing.T) {
			decls, err := errdoc(source, module)
			require.ErrorContains(t, err, "invalid --module")
			require.Nil(t, decls)
		})
	}
}

func TestErrdocInvalidDirectory(t *testing.T) {
	for _, dir := range []string{"with space", "with\"quote", "with\nnewline", "with`backtick", "with;semicolon", "a..b", ".hidden", "a+b", "a~b"} {
		t.Run(dir, func(t *testing.T) {
			if runtime.GOOS == "windows" && strings.ContainsAny(dir, "\"\n") {
				t.Skip("directory name cannot be created on Windows")
			}
			source := t.TempDir()
			writeSource(t, source, filepath.Join(dir, "errors.go"), "package example; var ErrExample = new(error)")
			decls, err := errdoc(source, "example.com/project")
			require.ErrorContains(t, err, "invalid package directory")
			require.Nil(t, decls)
		})
	}
}

func TestErrdocCLIProcess(t *testing.T) {
	if os.Getenv("ERRDOC_GEN_TEST_PROCESS") != "1" {
		return
	}
	os.Args = append(os.Args[:1], flag.Args()...)
	main()
	os.Exit(0)
}

func TestErrdocCLIRejectsNonNormalizedErrors(t *testing.T) {
	repo, err := filepath.Abs("..")
	require.NoError(t, err)
	module, err := os.ReadFile(filepath.Join(repo, "go.mod"))
	require.NoError(t, err)
	sums, err := os.ReadFile(filepath.Join(repo, "go.sum"))
	require.NoError(t, err)
	fixtureModule := strings.Replace(string(module), "module github.com/pingcap/errors", "module example.com/project", 1) +
		"\nrequire github.com/pingcap/errors v0.0.0\nreplace github.com/pingcap/errors => " + strconv.Quote(repo) + "\n"
	for _, retain := range []bool{false, true} {
		t.Run("retain-code="+strconv.FormatBool(retain), func(t *testing.T) {
			source := t.TempDir()
			writeSource(t, source, "go.mod", fixtureModule)
			writeSource(t, source, "go.sum", string(sums))
			writeSource(t, source, "errors.go", "package example\nimport \"errors\"\nvar ErrExample = errors.New(\"ordinary error\")\n")
			const previous = "[\"example:existing\"]\nerror = 'existing'\ndescription = 'keep me'\n"
			writeSource(t, source, "errors.toml", previous)
			cmd := exec.Command(os.Args[0], "-test.run=^TestErrdocCLIProcess$", "--", "--source", source,
				"--module", "example.com/project", "--retain-code="+strconv.FormatBool(retain))
			cmd.Env = append(os.Environ(), "ERRDOC_GEN_TEST_PROCESS=1", "GOPROXY=off", "GOFLAGS=-mod=readonly")
			output, err := cmd.CombinedOutput()
			require.Error(t, err, "%s", output)
			require.Contains(t, string(output), "No normalized errors found")
			require.NotContains(t, string(output), "Generate successfully")
			content, err := os.ReadFile(filepath.Join(source, "errors.toml"))
			require.NoError(t, err)
			require.Equal(t, previous, string(content))
			var definitions map[string]map[string]string
			require.NoError(t, toml.Unmarshal(content, &definitions))
			require.Equal(t, "keep me", definitions["example:existing"]["description"])
			if retain {
				require.FileExists(t, filepath.Join(source, autoDirectoryName, entryFileName))
			} else {
				require.NoDirExists(t, filepath.Join(source, autoDirectoryName))
			}
		})
	}
}

func TestErrdocCLIPreservesOutputOnFailure(t *testing.T) {
	for _, test := range []struct {
		name    string
		files   map[string]string
		args    []string
		message string
	}{
		{
			name:    "empty scan",
			message: "no error declarations found",
		},
		{
			name:    "parse error after valid declaration",
			files:   map[string]string{"errors.go": "package example; var ErrExample = new(error)", "z.go": "invalid Go"},
			message: "parse ",
		},
		{
			name:    "all packages ignored",
			files:   map[string]string{"skip/errors.go": "package example; var ErrExample = new(error)"},
			args:    []string{"--ignore", "skip"},
			message: "no error declarations found",
		},
		{
			name:    "module injection",
			files:   map[string]string{"errors.go": "package example; var ErrExample = new(error)"},
			args:    []string{"--module", "example.com/project\"\n)\nfunc init() {}\n//"},
			message: "invalid --module",
		},
		{
			name:    "directory injection",
			files:   map[string]string{"bad;dir/errors.go": "package example; var ErrExample = new(error)"},
			message: "invalid package directory",
		},
		{
			name:    "ignore source root",
			args:    []string{"--ignore", "."},
			message: "invalid --ignore directory",
		},
		{
			name:    "ignore outside source",
			args:    []string{"--ignore", "../outside"},
			message: "invalid --ignore directory",
		},
		{
			name:    "ignore absolute path",
			args:    []string{"--ignore", filepath.Join(t.TempDir(), "outside")},
			message: "invalid --ignore directory",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := t.TempDir()
			writeSource(t, source, "go.mod", "module example.com/project\n")
			const previous = "# Existing documentation\n[\"example:existing\"]\nerror = 'existing'\ndescription = 'keep me'\n"
			writeSource(t, source, "errors.toml", previous)
			for name, content := range test.files {
				writeSource(t, source, name, content)
			}
			args := []string{"-test.run=^TestErrdocCLIProcess$", "--", "--source", source, "--module", "example.com/project"}
			cmd := exec.Command(os.Args[0], append(args, test.args...)...)
			cmd.Env = append(os.Environ(), "ERRDOC_GEN_TEST_PROCESS=1")
			output, err := cmd.CombinedOutput()
			require.Error(t, err, "%s", output)
			require.Contains(t, string(output), test.message)
			require.NotContains(t, string(output), "Generate successfully")
			content, err := os.ReadFile(filepath.Join(source, "errors.toml"))
			require.NoError(t, err)
			require.Equal(t, previous, string(content))
			require.NoDirExists(t, filepath.Join(source, autoDirectoryName))
		})
	}
}
