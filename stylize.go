package main

// This file contains the main logic of the program. In general this is setup as
// a "pipeline" system, meaning that functions consume input channels and send
// their results to a returned output channel.
//
// The first stage is a file input. We either recursively search a given source
// directory or we use git to get a list of files that have changed since a
// given diffbase. Each of these inputs modes corresponds to a function that
// returns a channel of strings, on which will be sent a series of file paths to
// format/check.
//
// The next stage runs a pool of formatters asynchronously on the input file
// paths. The formatting/checking results are then forwarded to an output
// channel.
//
// From there, further operations collect stats, diffs, and log actions.

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/bradfitz/slice"
	"github.com/justbuchanan/stylize/util"
	"github.com/pkg/errors"
)

type FormattingResult struct {
	FilePath     string
	FormatNeeded bool
	Patch        string
	Error        error
}

// All parameters are required!
type StylizeContext struct {
	// The set of formatters to apply, keyed by file extension.
	Formatters map[string]Formatter
	// Command-line args to pass to each formatter, keyed by formatter name.
	FormatterArgs map[string][]string
	// Root directory to search for files under.
	RootDir string
	// File exclude patterns
	Exclude []string
	// If provided, only looks at files that differ from the diffbase. Otherwise looks at all files.
	GitDiffbase string
	// If Lines=true and GitDiffbase is provided, only lines that have changed
	// will be formatted. Some formatters don't support this, in which case the
	// whole file will be formatted.
	Lines bool
	// If given, a patch is written to the output showing changes that the formatters would make.
	PatchOut io.Writer
	// If true, formats all files in-place rather than performing a compliance check.
	InPlace bool
	// How many files to format simultaneously.
	Parallelism int
}

// Walks the given directory and sends all non-excluded files to the returned channel.
// @param rootDir absolute path to root directory
// @return file paths relative to rootDir
func IterateAllFiles(rootDir string, exclude []string) <-chan util.FileInfo {
	files := make(chan util.FileInfo)

	go func() {
		defer close(files)
		filepath.Walk(rootDir, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			relPath, _ := filepath.Rel(rootDir, path)
			isExcluded := util.FileIsExcluded(relPath, exclude)

			// Skip the entire directory
			if fi.IsDir() && isExcluded {
				return filepath.SkipDir
			}

			if !isExcluded && !fi.IsDir() {
				files <- util.FileInfo{Path: relPath}
			}

			return nil
		})
	}()

	return files
}

// Returns the toplevel directory of the git repo given a subdirectory contained
// within it (or the toplevel directory itself). This is the directory that
// contains the ".git" subdirectory.
func findGitRoot(dir string) (string, error) {
	var gitRootOut, stderr bytes.Buffer
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Stdout = &gitRootOut
	cmd.Stderr = &stderr
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return "", errors.Wrap(err, stderr.String())
	}
	return strings.Trim(gitRootOut.String(), "\n"), nil

}

// Finds files that have been modified since the common ancestor of HEAD and
// diffbase and sends them onto the returned channel.
// @return file paths relative to rootDir
func IterateGitChangedFiles(rootDir string, exclude []string, diffbase string) (<-chan util.FileInfo, error) {
	gitRoot, err := findGitRoot(rootDir)
	if err != nil {
		return nil, err
	}
	changedFiles, err := util.GitChangedFiles(rootDir, diffbase)
	if err != nil {
		return nil, err
	}

	files := make(chan util.FileInfo)
	go func() {
		defer close(files)

		for _, file := range changedFiles {
			absPath := filepath.Join(gitRoot, file)

			// get file path relative to root directory
			relPath, err := filepath.Rel(rootDir, absPath)
			if err != nil {
				log.Fatal(err)
			}

			if util.FileIsExcluded(relPath, exclude) {
				continue
			}

			// git diff will show files that have been deleted - we don't want
			// to try to format these since they don't exist anymore.
			// TODO: use os.IsNotExist(err) instead. this doesn't work for directories, though
			if _, err := os.Stat(absPath); err != nil {
				continue
			}

			files <- util.FileInfo{Path: relPath}
			// TODO: lines
		}
	}()

	return files, nil
}

// @param file FileInfo with Path relative to rootDir
func runFormatter(rootDir string, file util.FileInfo, formatter Formatter, formatterArgs []string, inPlace bool) FormattingResult {
	result := FormattingResult{
		FilePath: file.Path,
	}

	// TODO: figure out absolute vs relative file paths. need to know wdir in
	// order to put rel paths in patch. regular formatter is happier with just
	// an abs path.

	if inPlace {
		result.FormatNeeded, result.Error = FormatInPlaceAndCheckModified(formatter, formatterArgs, util.FileInfo{Path: filepath.Join(rootDir, file.Path), Lines: file.Lines})
	} else {
		result.Patch, result.Error = CreatePatchWithFormatter(formatter, formatterArgs, rootDir, file)
		result.FormatNeeded = len(result.Patch) > 0
	}

	return result
}

// Reads all incoming results and forwards them to the output channel. When all
// results have been read, writes the patch to the output writer.
func CollectPatch(results <-chan FormattingResult, patchOut io.Writer) <-chan FormattingResult {
	resultsOut := make(chan FormattingResult)

	go func() {
		defer close(resultsOut)

		// collect relevant results from the input channel and forward them to the output
		var formatResults []FormattingResult
		for r := range results {
			if r.Error == nil && r.FormatNeeded {
				formatResults = append(formatResults, r)
			}
			resultsOut <- r
		}

		// sort to ensure patches are consistent
		slice.Sort(formatResults, func(i, j int) bool {
			return formatResults[i].FilePath < formatResults[j].FilePath
		})

		// write patch output
		for _, r := range formatResults {
			patchOut.Write([]byte(r.Patch + "\n"))
		}
	}()

	return resultsOut
}

func (ctx *StylizeContext) RunFormattersOnFiles(fileChan <-chan util.FileInfo) <-chan FormattingResult {
	// use semaphore to limit how many formatting operations we run in parallel
	semaphore := make(chan int, ctx.Parallelism)
	var wg sync.WaitGroup

	resulstOut := make(chan FormattingResult)
	go func() {
		for file := range fileChan {
			ext := filepath.Ext(file.Path)
			if len(ext) == 0 {
				// if file doesn't have an extension, use the file name
				ext = filepath.Base(file.Path)
			}
			formatter := ctx.Formatters[ext]
			if formatter == nil {
				continue
			}

			wg.Add(1)
			semaphore <- 0 // acquire
			go func(file util.FileInfo, formatter Formatter, inPlace bool) {
				resulstOut <- runFormatter(ctx.RootDir, file, formatter, ctx.FormatterArgs[formatter.Name()], ctx.InPlace)
				wg.Done()
				<-semaphore // release
			}(file, formatter, ctx.InPlace)
		}

		wg.Wait()
		close(resulstOut)
	}()

	return resulstOut
}

type RunStats struct {
	Change, Total, Error int
}

// Consumes the input channel, logging all actions made and collecting stats.
// If the output is a terminal, prints files that are checked, but don't need formatting.
func LogActionsAndCollectStats(results <-chan FormattingResult, inPlace bool) RunStats {
	// Calculate terminal width so text can be padded appropriately for line-
	// overwriting (done only when output is a terminal).
	var termWidth int
	isTerm := isTerminal(os.Stderr)
	if isTerm {
		termWidth = int(getTermWidth(uintptr(syscall.Stderr)))
	} else {
		termWidth = 0
	}

	printf := func(tmp bool, format string, a ...interface{}) {
		fmt.Fprintf(os.Stderr, padToWidth(fmt.Sprintf(format, a...), termWidth))
		if tmp {
			// Print a \r at the end so that the next line printed overwrites
			// this one. Printing-in-place shows that the program is working,
			// but doesn't fill up the screen with unnecessary info
			fmt.Fprintf(os.Stderr, "\r")
		} else {
			fmt.Fprintf(os.Stderr, "\n")
		}
	}

	// iterate through all results, collecting basic stats and logging actions.
	var stats RunStats
	for r := range results {
		stats.Total++

		if r.Error != nil {
			if inPlace {
				printf(false, "Error formatting file '%s': %q", r.FilePath, r.Error)
			} else {
				printf(false, "Error checking file '%s': %q", r.FilePath, r.Error)
			}
			stats.Error++
			continue
		}

		if r.FormatNeeded {
			stats.Change++

			if inPlace {
				printf(false, "Formatted: '%s'", r.FilePath)
			} else {
				printf(false, "Needs formatting: '%s'", r.FilePath)
			}
		} else if isTerm {
			printf(true, "Checked '%s'", r.FilePath)
		}
	}

	if inPlace {
		printf(false, "%d / %d formatted", stats.Change, stats.Total)
	} else {
		printf(false, "%d / %d need formatting", stats.Change, stats.Total)
	}

	return stats
}

func (ctx *StylizeContext) Run() RunStats {
	if ctx.InPlace && ctx.PatchOut != nil {
		log.Fatal("Patch output writer should only be provided in non-inplace runs")
	}
	if !filepath.IsAbs(ctx.RootDir) {
		log.Fatalf("root directory should be an absolute path: '%s'", ctx.RootDir)
	}

	for _, excl := range ctx.Exclude {
		if filepath.IsAbs(excl) {
			log.Fatal("exclude directories should not be absolute")
		}
	}

	// setup file source
	var err error
	var fileChan <-chan util.FileInfo
	if len(ctx.GitDiffbase) > 0 {
		log.Printf("Examining files that have changed in git since %s", ctx.GitDiffbase)
		fileChan, err = IterateGitChangedFiles(ctx.RootDir, ctx.Exclude, ctx.GitDiffbase)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Print("Examining all files")
		fileChan = IterateAllFiles(ctx.RootDir, ctx.Exclude)
	}

	// run formatter on all files
	results := ctx.RunFormattersOnFiles(fileChan)

	// write patch to output if requested
	if ctx.PatchOut != nil {
		results = CollectPatch(results, ctx.PatchOut)
	}

	return LogActionsAndCollectStats(results, ctx.InPlace)
}
