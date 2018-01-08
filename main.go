package main

import (
	"flag"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	inPlaceFlag := flag.Bool("i", false, "[WARNING] There's no undo button, make a commit first. If enabled, formats files in place. Default behavior is just to check which files need formatting.")
	patchFileFlag := flag.String("patch_output", "", "Path to output patch to. If '-', writes to stdout.")
	configFileFlag := flag.String("config", ".stylize.yml", "Optional config file (defaults to .stylize.yml).")
	dirFlag := flag.String("dir", ".", "Directory to recursively format.")
	excludeDirFlag := flag.String("exclude_dirs", "", "Directories to exclude (comma-separated).")
	diffbaseFlag := flag.String("git_diffbase", "", "If provided, stylize only looks at files that differ from the given commit/branch.")
	parallelismFlag := flag.Int("j", 8, "Number of files to process in parallel.")
	flag.Parse()

	// Read config file
	cfg, err := LoadConfig(*configFileFlag)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal(err)
		} else {
			// log.Print("No config file")
		}
	} else {
		log.Printf("Loaded config from file %s", *configFileFlag)
	}

	// set style configs from config file
	// TODO: do better
	if cfg != nil {
		if len(cfg.ClangStyle) > 0 && len(*clangStyleFlag) == 0 {
			*clangStyleFlag = cfg.ClangStyle
		}
		if len(cfg.YapfStyle) > 0 && len(*yapfStyleFlag) == 0 {
			*yapfStyleFlag = cfg.YapfStyle
		}
	}

	var excludeDirs []string
	// exclude dirs from config
	if cfg != nil {
		excludeDirs = append(excludeDirs, cfg.ExcludeDirs...)
	}
	// exclude dirs from flag
	if len(*excludeDirFlag) > 0 {
		excludeDirs = append(excludeDirs, strings.Split(*excludeDirFlag, ",")...)
	}
	// make exclude dirs absolute
	for i := range excludeDirs {
		excludeDirs[i] = absPathOrFail(excludeDirs[i])
	}

	// setup formatters
	var formatters map[string]Formatter
	if cfg != nil && cfg.FormattersByExt != nil {
		formatters = LoadFormattersFromMapping(cfg.FormattersByExt)
	} else {
		formatters = LoadDefaultFormatters()
	}

	rootDir := absPathOrFail(*dirFlag)

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
		uglyCount, _, errCount = StylizeMain(formatters, rootDir, excludeDirs, *diffbaseFlag, patchOut, *inPlaceFlag, *parallelismFlag)
	} else {
		uglyCount, _, errCount = StylizeMain(formatters, rootDir, excludeDirs, *diffbaseFlag, nil, *inPlaceFlag, *parallelismFlag)
	}

	if errCount != 0 {
		os.Exit(1)
	}

	// Signal that files need formatting
	if !*inPlaceFlag && uglyCount > 0 {
		os.Exit(2)
	}
}
