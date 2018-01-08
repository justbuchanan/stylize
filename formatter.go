package main

import (
	"bytes"
	"github.com/pmezard/go-difflib/difflib"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// Common interface for all formatters.
//
// New formatters can be added by implementing this interface and registering an
// instance with RegisterFormatter().
type Formatter interface {
	// Reads the input stream and writes a prettified version to the output.
	FormatToBuffer(file string, in io.Reader, out io.Writer) error
	// Reformats the given file in-place.
	FormatInPlace(file string) error
	// Check if the required binary is installed.
	IsInstalled() bool
	// A list of file extensions (including the '.') that this formatter applies to.
	FileExtensions() []string
}

func FormatInPlaceAndCheckModified(F Formatter, absPath string) (bool, error) {
	// record modification time before running formatter
	fi, err := os.Stat(absPath)
	if err != nil {
		return false, err
	}
	mtimeBefore := fi.ModTime()

	err = F.FormatInPlace(absPath)
	if err != nil {
		return false, err
	}

	// See if file was modified
	fi, err = os.Stat(absPath)
	if err != nil {
		return false, err
	}
	mtimeAfter := fi.ModTime()
	modified := mtimeAfter.After(mtimeBefore)

	return modified, nil
}

func CreatePatchWithFormatter(F Formatter, wdir, file string) (string, error) {
	fileContent, err := ioutil.ReadFile(filepath.Join(wdir, file))
	if err != nil {
		return "", err
	}

	var formattedOutput bytes.Buffer
	err = F.FormatToBuffer(file, bytes.NewReader(fileContent), &formattedOutput)
	if err != nil {
		return "", err
	}

	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(fileContent)),
		B:        difflib.SplitLines(formattedOutput.String()),
		FromFile: "a/" + file,
		ToFile:   "b/" + file,
		Context:  3,
	})

	return diff, nil
}

// Returns a map of file extension to formatter for the ones specied in the
// input mapping.
func LoadFormattersFromMapping(extToName map[string]string) map[string]Formatter {
	byExt := make(map[string]Formatter)
	for ext, name := range extToName {
		formatter := FormatterRegistry[name]
		if formatter == nil {
			log.Fatalf("Unknown formatter: %s", name)
		}
		if !formatter.IsInstalled() {
			log.Fatalf("Formatter %s not installed", name)
		}
		if byExt[ext] != nil {
			log.Fatalf("Multiple formatters for extension '%s'", ext)
		}
		byExt[ext] = formatter
	}

	return byExt
}

// Returns a map of file extension to formatter for all installed formatters in
// the registry.
func LoadDefaultFormatters() map[string]Formatter {
	byExt := make(map[string]Formatter)
	for name, f := range FormatterRegistry {
		if !f.IsInstalled() {
			log.Printf("Skipping formatter %s b/c it's not installed", name)
			continue
		}

		for _, ext := range f.FileExtensions() {
			if byExt[ext] != nil {
				log.Fatalf("Multiple formatters for extension '%s'", ext)
			}

			byExt[ext] = f
		}
	}

	return byExt
}

func RegisterFormatter(name string, f Formatter) {
	if FormatterRegistry[name] != nil {
		log.Fatalf("Attempt to double-register formatter '%s'\n", name)
	}
	FormatterRegistry[name] = f
}

var (
	// Formatters by name.
	FormatterRegistry = make(map[string]Formatter)
)
