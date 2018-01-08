package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Stylize - code formatting tool")
		flag.PrintDefaults()
	}
	inPlaceFlag := flag.Bool("i", false, "[WARNING] There's no undo button, make a commit first. If enabled, formats files in place. Default behavior is just to check which files need formatting.")
	patchFileFlag := flag.String("patch_output", "", "Path to output patch to. If '-', writes to stdout.")
	configFileFlag := flag.String("config", ".stylize.yml", "Optional config file (defaults to .stylize.yml).")
	dirFlag := flag.String("dir", ".", "Directory to recursively format.")
	excludeFlag := flag.String("exclude", "", "A list of exclude patterns (comma-separated).")
	diffbaseFlag := flag.String("git_diffbase", "", "If provided, stylize only looks at files that differ from the given commit/branch.")
	parallelismFlag := flag.Int("j", 8, "Number of files to process in parallel.")
	printFormattersFlag := flag.Bool("print_formatters", false, "Print map of file extension to formatter, then exit.")
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

	rootDir, err := filepath.Abs(*dirFlag)
	if err != nil {
		log.Fatal(err)
	}

	var excludePatterns []string
	// exclude dirs from config
	if cfg != nil {
		excludePatterns = append(excludePatterns, cfg.ExcludePatterns...)
	}
	// exclude dirs from flag
	if len(*excludeFlag) > 0 {
		excludePatterns = append(excludePatterns, strings.Split(*excludeFlag, ",")...)
	}

	// setup formatters
	var formatters map[string]Formatter
	if cfg != nil && cfg.FormattersByExt != nil {
		formatters = LoadFormattersFromMapping(cfg.FormattersByExt)
	} else {
		formatters = LoadDefaultFormatters()
	}

	var formatterArgs map[string][]string
	if cfg != nil {
		formatterArgs = cfg.FormatterArgs
	}

	if *printFormattersFlag {
		fmt.Fprintln(os.Stderr, "Formatters:")
		for ext, formatter := range formatters {
			fmt.Fprintf(os.Stderr, "%s: %s\n", ext, formatter.Name())
		}
		os.Exit(1)
	}

	var stats RunStats
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
		stats = StylizeMain(formatters, formatterArgs, rootDir, excludePatterns, *diffbaseFlag, patchOut, *inPlaceFlag, *parallelismFlag)
	} else {
		stats = StylizeMain(formatters, formatterArgs, rootDir, excludePatterns, *diffbaseFlag, nil, *inPlaceFlag, *parallelismFlag)
	}

	if stats.Error != 0 {
		os.Exit(1)
	}

	// Signal that files need formatting
	if !*inPlaceFlag && stats.Change > 0 {
		os.Exit(2)
	}
}
