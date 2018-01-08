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

func TestCreatePatch(t *testing.T) {
	goldenFile := "testdata/patch.golden"

	var patchOut io.Writer

	// only used when not generating goldens
	var patchBuffer bytes.Buffer

	if *generateGoldens {
		patchFile, err := os.Create(goldenFile)
		if err != nil {
			t.Fatal(err)
		}
		patchOut = patchFile
		defer patchFile.Close()
		t.Log("Writing golden file to %s", goldenFile)
	} else {
		patchOut = &patchBuffer
	}

	StylizeMain(LoadDefaultFormatters(), absPathOrFail("testdata"), []string{absPathOrFail("testdata/exclude")}, "", patchOut, false, PARALLELISM)

	if !*generateGoldens {
		assertGoldenMatch(t, goldenFile, patchBuffer.String())
	}

	// TODO: test that git can apply patch
}

func isDirectoryFormatted(t *testing.T, dir string, exclude []string) bool {
	numFormatted, _, numError := StylizeMain(LoadDefaultFormatters(), dir, exclude, "", nil, false, PARALLELISM)
	return numFormatted == 0 && numError == 0
}

func TestInPlace(t *testing.T) {
	tmp, err := ioutil.TempDir("", "stylize")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("tmp dir: " + tmp)

	dir := copyTestData(t, tmp)

	exclude := []string{path.Join(dir, "exclude")}
	t.Log("exclude: " + strings.Join(exclude, ","))

	// run in-place formatting
	StylizeMain(LoadDefaultFormatters(), dir, exclude, "", nil, true, PARALLELISM)

	if !isDirectoryFormatted(t, dir, exclude) {
		t.Fatal("Second run of formatter showed a diff. Everything should have been fixed in-place the first time.")
	}

	// TODO: test exclude not modified

	os.RemoveAll(tmp)
}

func TestInPlaceWithConfig(t *testing.T) {
	tmp, err := ioutil.TempDir("", "stylize")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("tmp dir: " + tmp)

	dir := copyTestData(t, tmp)

	cfgPath := path.Join(dir, ".stylize.yml")
	err = ioutil.WriteFile(cfgPath, []byte("---\nformatters:\n  .py: yapf\nexclude_dirs:\n  - exclude"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Wrote config file: %s", cfgPath)

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Read config file")

	for i, edir := range cfg.ExcludeDirs {
		cfg.ExcludeDirs[i] = filepath.Join(dir, edir)
	}
	t.Log("exclude: " + strings.Join(cfg.ExcludeDirs, ","))

	if err != nil {
		t.Fatal(err)
	}

	formatters := LoadFormattersFromMapping(cfg.FormattersByExt)

	// run in-place formatting
	numChanged, numTotal, numErr := StylizeMain(formatters, dir, cfg.ExcludeDirs, "", nil, true, PARALLELISM)
	t.Log("Ran stylize")
	t.Logf("%d, %d, %d", numChanged, numTotal, numErr)

	if numChanged != 1 {
		t.Fatal("One file should have changed")
	}

	numChanged, numTotal, numErr = StylizeMain(formatters, dir, cfg.ExcludeDirs, "", nil, true, PARALLELISM)
	t.Log("Ran stylize")
	t.Logf("%d, %d, %d", numChanged, numTotal, numErr)

	if numChanged != 0 {
		t.Fatal("No files should have changed")
	}

	os.RemoveAll(tmp)
}

func TestGitDiffbase(t *testing.T) {
	tmp, err := ioutil.TempDir("", "stylize")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("tmp dir: " + tmp)

	// copy testdata to output
	dir := copyTestData(t, tmp)

	runCmd(t, dir, "git", "init")
	runCmd(t, dir, "git", "add", ".")
	runCmd(t, dir, "git", "commit", "-m", "first commit")
	runCmd(t, dir, "git", "checkout", "-b", "other")

	// add new files
	runCmd(t, dir, "cp", "bad.cpp", "exclude/bad2.cpp")
	runCmd(t, dir, "git", "mv", "bad.cpp", "bad2.cpp")
	runCmd(t, dir, "rm", "bad.py")
	runCmd(t, dir, "git", "add", ".")
	runCmd(t, dir, "git", "commit", "-m", "added files")

	exclude := []string{path.Join(dir, "exclude")}
	t.Log("exclude: " + strings.Join(exclude, ","))

	// run stylize with diffbase provided
	numFormatted, _, numErr := StylizeMain(LoadDefaultFormatters(), dir, exclude, "master", nil, true, PARALLELISM)
	if numFormatted != 1 {
		t.Fatalf("Stylize should have formatted one and only one file. Instead it was %d", numFormatted)
	}
	if numErr > 0 {
		t.Fatalf("Error formatting files")
	}

	// git diff
	files, err := gitChangedFiles(dir, "HEAD")
	if err != nil {
		t.Fatal(err)
	}

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

func copyTestData(t *testing.T, dir string) string {
	cpCmd := exec.Command("cp", "-r", "testdata/", dir)
	err := cpCmd.Run()
	if err != nil {
		t.Fatal(err)
	}
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
	if err != nil {
		t.Fatal(err)
	}

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
