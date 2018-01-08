package main

import (
	"bytes"
	"fmt"
	"github.com/bradfitz/slice"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

type FormattingResult struct {
	FilePath     string
	FormatNeeded bool
	Patch        string
	Error        error
}

func fileIsExcluded(file string, excludeDirs []string) bool {
	for _, eDir := range excludeDirs {
		if filepath.HasPrefix(file, eDir) {
			return true
		}
	}
	return false
}

// Walks the given directory and sends all non-excluded files to the returned channel.
// @param rootDir absolute path to root directory
// @return file paths relative to rootDir
func IterateAllFiles(rootDir string, excludeDirs []string) <-chan string {
	files := make(chan string)

	go func() {
		defer close(files)
		filepath.Walk(rootDir, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			excludeDirs = append(excludeDirs, absPathOrFail(".git"), absPathOrFail(".hg"))
			if fi.IsDir() && fileIsExcluded(path, excludeDirs) {
				return filepath.SkipDir
			}

			if fi.IsDir() {
				return nil
			}

			relPath, _ := filepath.Rel(rootDir, path)
			files <- relPath

			return nil
		})
	}()

	return files
}

// Finds files that have been modified since the common ancestor of HEAD and
// diffbase and sends them onto the returned channel.
// @return file paths relative to rootDir
func IterateGitChangedFiles(rootDir string, excludeDirs []string, diffbase string) (<-chan string, error) {
	changedFiles, err := gitChangedFiles(rootDir, diffbase)
	if err != nil {
		return nil, err
	}

	// find ancestor directory of rootDir that has the .git directory
	var gitRootOut, stderr bytes.Buffer
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Stdout = &gitRootOut
	cmd.Stderr = &stderr
	cmd.Dir = rootDir
	err = cmd.Run()
	if err != nil {
		return nil, errors.Wrap(err, stderr.String())
	}
	gitRoot := strings.Trim(gitRootOut.String(), "\n")

	files := make(chan string)
	go func() {
		defer close(files)

		for _, file := range changedFiles {
			absPath := filepath.Join(gitRoot, file)
			if fileIsExcluded(absPath, excludeDirs) {
				// log.Printf("Excluding file: %s", absPath)
				continue
			}

			// git diff will show files that have been deleted - we don't want
			// to try to format these since they don't exist anymore.
			// TODO: use os.IsNotExist(err) instead. this doesn't work for directories, though
			if _, err := os.Stat(absPath); err != nil {
				continue
			}

			// get file path relative to root directory
			relPath, err := filepath.Rel(rootDir, absPath)
			if err != nil {
				log.Fatal(err)
			}

			files <- relPath
		}
	}()

	return files, nil
}

func runFormatter(rootDir, file string, formatter Formatter, inPlace bool) FormattingResult {
	result := FormattingResult{
		FilePath: file,
	}

	if inPlace {
		result.FormatNeeded, result.Error = FormatInPlaceAndCheckModified(formatter, filepath.Join(rootDir, file))
	} else {
		result.Patch, result.Error = CreatePatchWithFormatter(formatter, rootDir, file)
		if len(result.Patch) > 0 {
			result.FormatNeeded = true
		}
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
		var resultList []FormattingResult
		for r := range results {
			if r.Error == nil && r.FormatNeeded {
				resultList = append(resultList, r)
			}
			resultsOut <- r
		}

		// sort to ensure patches are consistent
		slice.Sort(resultList, func(i, j int) bool {
			return resultList[i].FilePath < resultList[j].FilePath
		})

		// write patch output
		for _, r := range resultList {
			patchOut.Write([]byte(r.Patch + "\n"))
		}
	}()

	return resultsOut
}

func RunFormattersOnFiles(formatters map[string]Formatter, fileChan <-chan string, rootDir string, inPlace bool, parallelism int) <-chan FormattingResult {
	// use semaphore to limit how many formatting operations we run in parallel
	semaphore := make(chan int, parallelism)
	var wg sync.WaitGroup

	resultOut := make(chan FormattingResult)
	go func() {
		for file := range fileChan {
			ext := filepath.Ext(file)
			if len(ext) == 0 {
				// if file doesn't have an extension, use the file name
				ext = filepath.Base(file)
			}
			formatter := formatters[ext]
			if formatter == nil {
				continue
			}

			wg.Add(1)
			semaphore <- 0 // acquire
			go func(file string, formatter Formatter, inPlace bool) {
				resultOut <- runFormatter(rootDir, file, formatter, inPlace)
				wg.Done()
				<-semaphore // release
			}(file, formatter, inPlace)
		}

		wg.Wait()
		close(resultOut)
	}()

	return resultOut
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
			printf(false, "Error formatting file '%s': %q", r.FilePath, r.Error)
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

// @param gitDiffbase If provided, only looks at files that differ from the
//     diffbase. Otherwise looks at all files.
// @param formatters A map of file extension -> formatter
// @return (changeCount, totalCount, errCount)
func StylizeMain(formatters map[string]Formatter, rootDir string, excludeDirs []string, gitDiffbase string, patchOut io.Writer, inPlace bool, parallelism int) RunStats {
	if inPlace && patchOut != nil {
		log.Fatal("Patch output writer should only be provided in non-inplace runs")
	}
	if !filepath.IsAbs(rootDir) {
		log.Fatalf("root directory should be an absolute path: '%s'", rootDir)
	}

	for _, excl := range excludeDirs {
		if !filepath.IsAbs(excl) {
			log.Fatal("exclude directories should be absolute")
		}
	}

	// setup file source
	var err error
	var fileChan <-chan string
	if len(gitDiffbase) > 0 {
		fileChan, err = IterateGitChangedFiles(rootDir, excludeDirs, gitDiffbase)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fileChan = IterateAllFiles(rootDir, excludeDirs)
	}

	// run formatter on all files
	results := RunFormattersOnFiles(formatters, fileChan, rootDir, inPlace, parallelism)

	// write patch to output if requested
	if patchOut != nil {
		results = CollectPatch(results, patchOut)
	}

	return LogActionsAndCollectStats(results, inPlace)
}
