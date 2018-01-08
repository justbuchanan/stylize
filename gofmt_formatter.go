package main

import (
	"io"
)

type GolangFormatter struct{}

func init() {
	RegisterFormatter("gofmt", &GolangFormatter{})
}

func (F *GolangFormatter) FileExtensions() []string {
	return []string{".go"}
}

func (F *GolangFormatter) IsInstalled() bool {
	// if we're running go code, then gofmt has to be here...
	return true
}

func (F *GolangFormatter) FormatToBuffer(file string, in io.Reader, out io.Writer) error {
	return runIOCommand([]string{"gofmt"}, in, out)
}

func (F *GolangFormatter) FormatInPlace(absPath string) error {
	return runIOCommand([]string{"gofmt", "-l", "-w", absPath}, nil, nil)
}
