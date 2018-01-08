package main

import (
	"bytes"
	"flag"
	"github.com/pmezard/go-difflib/difflib"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

var (
	generateGoldens = flag.Bool("generate_goldens", false, "Generate golden files")
	PARALLELISM     = 5
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func expectMatch(t *testing.T, match bool, pattern, file string) {
	m := filePatternMatch(pattern, file)
	if m != match {
		t.Logf("'%s' '%s'", pattern, file)
		if match {
			t.Error("Expected match, but didn't get one")
		} else {
			t.Error("It matched, but wasnt supposed to")
		}
	}
}

func TestMatch(t *testing.T) {
	expectMatch(t, true, "exclude/", "exclude/file.cpp")
	expectMatch(t, true, "exclude", "exclude/file.cpp")
	expectMatch(t, false, "exclud", "exclude/file.cpp")
	expectMatch(t, false, "bbb", "bbb.cpp")
	expectMatch(t, true, "*", "bad.cpp")
	expectMatch(t, true, "*", "files/bad.cpp")
	// TODO: gitignore handles this case, should we?
	expectMatch(t, false, "/files", "files/bad.cpp")
}

func TestCreatePatch(t *testing.T) {
	goldenFile := "testdata/patch.golden"

	var patchOut io.Writer

	// only used when not generating goldens
	var patchBuffer bytes.Buffer

	if *generateGoldens {
		patchFile, err := os.Create(goldenFile)
		tCheckErr(t, err)
		patchOut = patchFile
		defer patchFile.Close()
		t.Logf("Writing golden file to %s", goldenFile)
	} else {
		patchOut = &patchBuffer
	}

	absDirPath, _ := filepath.Abs("testdata")
	StylizeMain(LoadDefaultFormatters(), nil, absDirPath, []string{"exclude"}, "", patchOut, false, PARALLELISM)

	if !*generateGoldens {
		assertGoldenMatch(t, goldenFile, patchBuffer.String())
	}

	// TODO: test that git can apply patch
}

func isDirectoryFormatted(t *testing.T, dir string, exclude []string) bool {
	stats := StylizeMain(LoadDefaultFormatters(), nil, dir, exclude, "", nil, false, PARALLELISM)
	return stats.Change == 0 && stats.Error == 0
}

func TestInPlace(t *testing.T) {
	tmp := mktmp(t)
	dir := copyTestData(t, tmp)

	exclude := []string{"exclude"}
	t.Log("exclude: " + strings.Join(exclude, ","))

	// run in-place formatting
	stats := StylizeMain(LoadDefaultFormatters(), nil, dir, exclude, "", nil, true, PARALLELISM)
	if stats.Error > 0 {
		t.Fatal("Formatting failed")
	}

	if !isDirectoryFormatted(t, dir, exclude) {
		t.Fatal("Second run of formatter showed a diff. Everything should have been fixed in-place the first time.")
	}

	// TODO: test exclude not modified

	os.RemoveAll(tmp)
}

func TestInPlaceWithConfig(t *testing.T) {
	tmp := mktmp(t)
	dir := copyTestData(t, tmp)

	cfgPath := path.Join(dir, ".stylize.yml")
	err := ioutil.WriteFile(cfgPath, []byte("---\nformatters:\n  .py: yapf\nexclude:\n  - exclude"), 0644)
	tCheckErr(t, err)
	t.Logf("Wrote config file: %s", cfgPath)

	cfg, err := LoadConfig(cfgPath)
	tCheckErr(t, err)
	t.Log("Read config file")
	t.Log("exclude: " + strings.Join(cfg.ExcludePatterns, ","))

	formatters := LoadFormattersFromMapping(cfg.FormattersByExt)

	// run in-place formatting
	stats := StylizeMain(formatters, nil, dir, cfg.ExcludePatterns, "", nil, true, PARALLELISM)
	t.Logf("Stylize results: %d, %d, %d", stats.Change, stats.Total, stats.Error)

	if stats.Change != 1 {
		t.Fatal("One file should have changed")
	}

	stats = StylizeMain(formatters, nil, dir, cfg.ExcludePatterns, "", nil, true, PARALLELISM)
	t.Logf("Stylize results: %d, %d, %d", stats.Change, stats.Total, stats.Error)

	if stats.Change != 0 {
		t.Fatal("No files should have changed")
	}

	os.RemoveAll(tmp)
}

func TestGitDiffbase(t *testing.T) {
	tmp := mktmp(t)
	dir := copyTestData(t, tmp)

	// initial commit
	runCmd(t, dir, "git", "init")
	runCmd(t, dir, "git", "add", ".")
	runCmd(t, dir, "git", "commit", "-m", "first commit")

	// new branch
	runCmd(t, dir, "git", "checkout", "-b", "other")

	// add new files, delete a couple
	runCmd(t, dir, "cp", "bad.cpp", "exclude/bad2.cpp")
	runCmd(t, dir, "cp", "bad.cpp", "bad2.cpp")
	runCmd(t, dir, "rm", "bad.py", "bad.cpp")
	runCmd(t, dir, "git", "add", ".")
	runCmd(t, dir, "git", "commit", "-m", "added files")

	exclude := []string{"exclude"}
	t.Log("exclude: " + strings.Join(exclude, ","))

	// run stylize with diffbase provided
	stats := StylizeMain(LoadDefaultFormatters(), nil, dir, exclude, "master", nil, true, PARALLELISM)
	if stats.Change != 1 {
		t.Fatalf("Stylize should have formatted one and only one file. Instead it was %d", stats.Change)
	}
	if stats.Error > 0 {
		t.Fatalf("Error formatting files")
	}

	// git diff
	files, err := gitChangedFiles(dir, "HEAD")
	tCheckErr(t, err)

	t.Logf("Files changed by stylize: %s", strings.Join(files, ", "))

	if len(files) != 1 || files[0] != "bad2.cpp" {
		t.Fatal("Stylize should only have formatted bad2.cpp.")
	}

	// Remove directory if test passed
	os.RemoveAll(tmp)
}

func TestCollectPatch(t *testing.T) {
	// Send fake results to a new channel.
	results := make(chan FormattingResult)
	go func() {
		results <- FormattingResult{
			FilePath:     "a",
			Patch:        "diff1\ndiff2\n",
			FormatNeeded: true,
		}
		results <- FormattingResult{
			FilePath:     "b",
			Patch:        "diff3\ndiff4\n",
			FormatNeeded: true,
		}
		close(results)
	}()

	var patchOut bytes.Buffer
	resultsAfterPatch := CollectPatch(results, &patchOut)
	// consume results to run pipeline
	for _ = range resultsAfterPatch {
	}

	expected := "diff1\ndiff2\n\ndiff3\ndiff4\n\n"
	if expected != patchOut.String() {
		t.Logf("Expected: %s", expected)
		t.Fatalf("Got: %s", patchOut.String())
	}
}

func tCheckErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func mktmp(t *testing.T) string {
	tmp, err := ioutil.TempDir("", "stylize")
	tCheckErr(t, err)

	t.Log("tmp dir: " + tmp)
	return tmp
}

func copyTestData(t *testing.T, dir string) string {
	cpCmd := exec.Command("cp", "-r", "testdata/", dir)
	err := cpCmd.Run()
	tCheckErr(t, err)
	return path.Join(dir, "testdata")
}

func runCmd(t *testing.T, dir string, bin string, args ...string) {
	t.Logf("cmd: %s %s", bin, strings.Join(args, " "))
	cmd := exec.Command(bin, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Dir = dir
	err := cmd.Run()
	if err != nil {
		t.Fatal(err, stderr.String())
	}
}

func assertGoldenMatch(t *testing.T, goldenFile string, genfileContent string) {
	goldenContent, err := ioutil.ReadFile(goldenFile)
	tCheckErr(t, err)

	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(goldenContent)),
		B:        difflib.SplitLines(genfileContent),
		FromFile: "golden",
		ToFile:   "generated",
		Context:  3,
	})

	if len(diff) > 0 {
		t.Log("Diff:\n" + diff)
		t.Fatal("Generated patch doesn't match golden")
	}
}
