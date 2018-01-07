package main

import (
	"github.com/pborman/getopt/v2"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	inPlaceFlag := getopt.BoolLong("in_place", 'i', "If enabled, formats files in place. Default behavior is just to check which files need formatting.")
	patchFileFlag := getopt.StringLong("output_patch_file", 'o', "", "Path to output patch to. If '-', writes to stdout.")
	dirFlag := getopt.StringLong("dir", 'd', ".", "Directory to recursively format.")
	excludeDirFlag := getopt.StringLong("exclude_dirs", 'e', "", "Directories to exclude")
	diffbaseFlag := getopt.StringLong("git_diffbase", 'g', "", "If provided, stylize only looks at files that differ from the given commit/branch.")
	parallelismFlag := getopt.IntLong("jobs", 'j', 8, "Number of files to process in parallel")
	getopt.Parse()

	var excludeDirs []string
	if len(*excludeDirFlag) > 0 {
		excludeDirs = strings.Split(*excludeDirFlag, ",")

		// absolutize exclude dirs
		for i := range excludeDirs {
			excludeDirs[i] = absPathOrDie(excludeDirs[i])
		}
	}

	rootDir := absPathOrDie(*dirFlag)

	var uglyCount, errCount int
	if !*inPlaceFlag && len(*patchFileFlag) > 0 {
		// Setup patch output writer
		var err error
		var patchOut io.Writer
		if *patchFileFlag == "-" {
			patchOut = os.Stdout
			log.Print("Writing patch to stdout")
		} else {
			var patchFileOut *os.File
			patchFileOut, err = os.Create(*patchFileFlag)
			patchOut = patchFileOut
			if err != nil {
				log.Fatal(err)
			}
			defer patchFileOut.Close()
			log.Printf("Writing patch to file %s", *patchFileFlag)
		}
		uglyCount, _, errCount = StylizeMain(rootDir, excludeDirs, *diffbaseFlag, patchOut, *inPlaceFlag, *parallelismFlag)
	} else {
		uglyCount, _, errCount = StylizeMain(rootDir, excludeDirs, *diffbaseFlag, nil, *inPlaceFlag, *parallelismFlag)
	}

	if errCount != 0 {
		os.Exit(1)
	}

	// Signal that files need formatting
	if !*inPlaceFlag && uglyCount > 0 {
		os.Exit(2)
	}
}
