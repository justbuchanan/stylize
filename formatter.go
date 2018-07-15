package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/pmezard/go-difflib/difflib"
)

// Common interface for all formatters.
//
// A new formatter can be added by implementing this interface and adding it to
// the global registry.
type Formatter interface {
	Name() string
	// Reads the input stream and writes a prettified version to the output.
	FormatToBuffer(args []string, file string, in io.Reader, out io.Writer) error
	// Reformats the given file in-place.
	FormatInPlace(args []string, file string) error
	// Check if the required binary is installed.
	IsInstalled() bool
	// A list of file extensions (including the '.') that this formatter applies to.
	FileExtensions() []string
}

func FormatInPlaceAndCheckModified(F Formatter, args []string, absPath string) (bool, error) {
	// record modification time before running formatter
	fi, err := os.Stat(absPath)
	if err != nil {
		return false, err
	}
	mtimeBefore := fi.ModTime()

	err = F.FormatInPlace(args, absPath)
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

func CreatePatchWithFormatter(F Formatter, args []string, wdir, file string) (string, error) {
	fileContent, err := ioutil.ReadFile(filepath.Join(wdir, file))
	if err != nil {
		return "", err
	}

	var formattedOutput bytes.Buffer
	err = F.FormatToBuffer(args, file, bytes.NewReader(fileContent), &formattedOutput)
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

func LookupFormatter(name string) Formatter {
	for _, f := range FormatterRegistry {
		if f.Name() == name {
			return f
		}
	}
	return nil
}

// Returns a map of file extension to formatter for the ones specied in the
// input mapping.
func LoadFormattersFromMapping(extToName map[string]string) map[string]Formatter {
	byExt := make(map[string]Formatter)
	for ext, name := range extToName {
		formatter := LookupFormatter(name)
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
	for _, f := range FormatterRegistry {
		if !f.IsInstalled() {
			log.Printf("Skipping formatter %s b/c it's not installed", f.Name())
			continue
		}

		for _, ext := range f.FileExtensions() {
			if byExt[ext] != nil {
				// log.Printf("Multiple formatters for extension '%s'", ext)
				continue
			}

			byExt[ext] = f
		}
	}

	return byExt
}

var (
	// Global list of all formatters.
	// If multiple formatters apply to the same file type, their order here
	// determines precedence. Lower index = higher priority.
	FormatterRegistry = []Formatter{
		&ClangFormatter{},
		&UncrustifyFormatter{},
		&PrettierFormatter{},
		&YapfFormatter{},
		&GofmtFormatter{},
		&BuildifierFormatter{},
		&RustfmtFormatter{},
	}
)
