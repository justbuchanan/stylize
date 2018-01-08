package main

// https://github.com/prettier/prettier

import (
	"io"
	"os/exec"
)

type PrettierFormatter struct{}

func init() {
	RegisterFormatter(&PrettierFormatter{})
}

func (F *PrettierFormatter) Name() string {
	return "prettier"
}

func (F *PrettierFormatter) FileExtensions() []string {
	return []string{".md", ".json", ".css", ".scss", ".less", ".ts"}
}

func (F *PrettierFormatter) IsInstalled() bool {
	_, err := exec.LookPath("prettier")
	return err == nil
}

func (F *PrettierFormatter) FormatToBuffer(args []string, file string, in io.Reader, out io.Writer) error {
	return runIOCommand(append([]string{"prettier", "--stdin-filepath", file}, args...), in, out)
}

func (F *PrettierFormatter) FormatInPlace(args []string, file string) error {
	return runIOCommand(append([]string{"prettier", "--write", file}, args...), nil, nil)
}
