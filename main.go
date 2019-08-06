package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Remove date/time from logs
	log.SetFlags(0)

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Stylize - code formatting tool")
		flag.PrintDefaults()
	}
	inPlaceFlag := flag.Bool("i", false, "[WARNING] There's no undo button, make a commit first. If enabled, formats files in place. Default behavior is just to check which files need formatting.")
	var patchFile string
	flag.StringVar(&patchFile, "patch_output", "", "Path to output patch to. If '-', writes to stdout.")
	flag.StringVar(&patchFile, "o", "", "Alias for --patch_output")
	var configFile string
	flag.StringVar(&configFile, "config", ".stylize.yml", "Optional config file (defaults to .stylize.yml).")
	flag.StringVar(&configFile, "c", ".stylize.yml", "Alias for --config")
	dirFlag := flag.String("dir", ".", "Directory to recursively format.")
	excludeFlag := flag.String("exclude", "", "A list of exclude patterns (comma-separated).")
	var diffbase string
	flag.StringVar(&diffbase, "git_diffbase", "", "If provided, stylize only looks at files that differ from the given commit/branch.")
	flag.StringVar(&diffbase, "g", "", "Alias for git_diffbase")
	linesFlag := flag.Bool("lines", true, "When used with --git_diffbase/-g, stylize only formats the *lines* in the file that have changed. Formatters that don't support this option will format the whole file.")
	parallelismFlag := flag.Int("j", 8, "Number of files to process in parallel.")
	printFormattersFlag := flag.Bool("print_formatters", false, "Print map of file extension to formatter, then exit.")
	flag.Parse()

	// Read config file
	cfg, err := LoadConfig(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// log.Print("No config file")
		} else {
			log.Fatal(err)
		}
	} else {
		log.Printf("Loaded config from file %s", configFile)
	}

	ctx := StylizeContext{
		GitDiffbase: diffbase,
		Lines:       *linesFlag,
		InPlace:     *inPlaceFlag,
		Parallelism: *parallelismFlag,
	}

	if ctx.RootDir, err = filepath.Abs(*dirFlag); err != nil {
		log.Fatal(err)
	}

	// Exclude common vcs directories
	ctx.Exclude = append(ctx.Exclude, ".git", ".hg")

	if cfg != nil {
		ctx.Exclude = append(ctx.Exclude, cfg.ExcludePatterns...)
		ctx.FormatterArgs = cfg.FormatterArgs
	}

	// exclude dirs from flag
	if len(*excludeFlag) > 0 {
		ctx.Exclude = append(ctx.Exclude, strings.Split(*excludeFlag, ",")...)
	}

	// setup formatters
	if cfg != nil && cfg.FormattersByExt != nil {
		ctx.Formatters = LoadFormattersFromMapping(cfg.FormattersByExt)
	} else {
		ctx.Formatters = LoadDefaultFormatters()
	}

	if *printFormattersFlag {
		log.Println("Formatters:")
		for ext, formatter := range ctx.Formatters {
			log.Printf("%s: %s\n", ext, formatter.Name())
		}
		os.Exit(0)
	}

	if !*inPlaceFlag && len(patchFile) > 0 {
		// Setup patch output writer
		if patchFile == "-" {
			ctx.PatchOut = os.Stdout
			log.Print("Writing patch to stdout")
		} else {
			var patchFileOut *os.File
			if patchFileOut, err = os.Create(patchFile); err != nil {
				log.Fatal(err)
			}
			ctx.PatchOut = patchFileOut
			defer patchFileOut.Close()
			log.Printf("Writing patch to file %s", patchFile)
		}
	}

	stats := ctx.Run()

	if stats.Error > 0 {
		os.Exit(1)
	}

	// Signal that files need formatting
	if !*inPlaceFlag && stats.Change > 0 {
		os.Exit(2)
	}
}
