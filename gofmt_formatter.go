package main

import (
	"io"
)

type GolangFormatter struct{}

func init() {
	RegisterFormatter(&GolangFormatter{})
}

func (F *GolangFormatter) Name() string {
	return "gofmt"
}

func (F *GolangFormatter) FileExtensions() []string {
	return []string{".go"}
}

func (F *GolangFormatter) IsInstalled() bool {
	// if we're running go code, then gofmt has to be here...
	return true
}

func (F *GolangFormatter) FormatToBuffer(args []string, file string, in io.Reader, out io.Writer) error {
	return runIOCommand([]string{"gofmt"}, in, out)
}

func (F *GolangFormatter) FormatInPlace(args []string, absPath string) error {
	return runIOCommand([]string{"gofmt", "-l", "-w", absPath}, nil, nil)
}
