package main

import (
	"flag"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	inPlaceFlag := flag.Bool("i", false, "If enabled, formats files in place. Default behavior is just to check which files need formatting.")
	patchFileFlag := flag.String("o", "", "Path to output patch to. If '-', writes to stdout.")
	dirFlag := flag.String("dir", ".", "Directory to recursively format.")
	excludeDirFlag := flag.String("exclude_dirs", "", "Directories to exclude")
	diffbaseFlag := flag.String("git_diffbase", "", "If provided, stylize only looks at files that differ from the given commit/branch.")
	parallelismFlag := flag.Int("j", 8, "Number of files to process in parallel")
	flag.Parse()

	var excludeDirs []string
	if len(*excludeDirFlag) > 0 {
		excludeDirs = strings.Split(*excludeDirFlag, ",")

		// absolutize exclude dirs
		for i := range excludeDirs {
			excludeDirs[i] = absPathOrDie(excludeDirs[i])
		}
	}

	rootDir := absPathOrDie(*dirFlag)

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
		StylizeMain(rootDir, excludeDirs, *diffbaseFlag, patchOut, *inPlaceFlag, *parallelismFlag)
	} else {
		StylizeMain(rootDir, excludeDirs, *diffbaseFlag, nil, *inPlaceFlag, *parallelismFlag)
	}
}
