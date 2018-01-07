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

	StylizeMain(absPathOrDie("testdata"), []string{absPathOrDie("testdata/exclude")}, "", patchOut, false, PARALLELISM)

	if !*generateGoldens {
		assertGoldenMatch(t, goldenFile, patchBuffer.String())
	}

	// TODO: test that git can apply patch
}

func isDirectoryFormatted(t *testing.T, dir string, exclude []string) bool {
	numFormatted, _ := StylizeMain(dir, exclude, "", nil, false, PARALLELISM)
	return numFormatted == 0
}

func TestInPlace(t *testing.T) {
	tmp, err := ioutil.TempDir("", "stylize")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	t.Log("tmp dir: " + tmp)

	// copy testdata to output
	cpCmd := exec.Command("cp", "-r", "testdata/", tmp)
	err = cpCmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	dir := path.Join(tmp, "testdata")

	// TODO: test exclude not modified

	exclude := []string{path.Join(dir, "exclude")}
	t.Log("exclude: " + strings.Join(exclude, ","))
	// run in-place formatting
	StylizeMain(dir, exclude, "", nil, true, PARALLELISM)

	if !isDirectoryFormatted(t, dir, exclude) {
		t.Fatal("Second run of formatter showed a diff. Everything should have been fixed in-place the first time.")
	}
}

func TestGitDiffbase(t *testing.T) {
	tmp, err := ioutil.TempDir("", "stylize")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("tmp dir: " + tmp)

	// copy testdata to output
	cpCmd := exec.Command("cp", "-r", "testdata/", tmp)
	err = cpCmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	dir := path.Join(tmp, "testdata")

	runCmd(t, dir, "git", "init")
	runCmd(t, dir, "git", "add", ".")
	runCmd(t, dir, "git", "commit", "-m", "first commit")
	runCmd(t, dir, "git", "checkout", "-b", "other")

	// add new files
	runCmd(t, dir, "cp", "bad.cpp", "bad2.cpp")
	runCmd(t, dir, "cp", "bad.cpp", "exclude/bad2.cpp")
	runCmd(t, dir, "git", "add", ".")
	runCmd(t, dir, "git", "commit", "-m", "added files")

	exclude := []string{path.Join(dir, "exclude")}
	t.Log("exclude: " + strings.Join(exclude, ","))

	// run stylize with diffbase provided
	numFormatted, _ := StylizeMain(dir, exclude, "master", nil, true, PARALLELISM)
	if numFormatted != 1 {
		t.Fatalf("Stylize should have formatted one and only one file. Instead it was %d", numFormatted)
	}

	diffCmd := exec.Command("git", "--no-pager", "diff", "--name-only")
	diffCmd.Dir = dir
	var stdout bytes.Buffer
	diffCmd.Stdout = &stdout
	err = diffCmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	files := strings.Split(strings.Trim(stdout.String(), "\n"), "\n")

	t.Logf("Files changed by stylize: %s", strings.Join(files, ", "))

	if len(files) != 1 || files[0] != "bad2.cpp" {
		t.Fatal("Stylize should only have formatted bad2.cpp.")
	}

	// Remove directory if test passed
	defer os.RemoveAll(tmp)
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

	expected := "diff1\ndiff2\ndiff3\ndiff4\n"
	if expected != patchOut.String() {
		t.Logf("Expected: %s", expected)
		t.Fatalf("Got: %s", patchOut.String())
	}
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
