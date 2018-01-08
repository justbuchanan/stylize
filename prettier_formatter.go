package main

// https://github.com/prettier/prettier

// TODO: add configuration options

import (
	"io"
	"os/exec"
)

type PrettierFormatter struct{}

func init() {
	RegisterFormatter("prettier", &PrettierFormatter{})
}

func (F *PrettierFormatter) FileExtensions() []string {
	return []string{".md", ".json", ".css", ".scss", ".less", ".ts"}
}

func (F *PrettierFormatter) IsInstalled() bool {
	_, err := exec.LookPath("prettier")
	return err == nil
}

func (F *PrettierFormatter) FormatToBuffer(file string, in io.Reader, out io.Writer) error {
	return runIOCommand([]string{"prettier", "--stdin-filepath", file}, in, out)
}

func (F *PrettierFormatter) FormatInPlace(file string) error {
	return runIOCommand([]string{"prettier", "--write", file}, nil, nil)
}
