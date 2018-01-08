package main

import (
	"io"
	"os/exec"
)

type YapfFormatter struct{}

func init() {
	RegisterFormatter(&YapfFormatter{})
}

func (F *YapfFormatter) Name() string {
	return "yapf"
}

func (F *YapfFormatter) FileExtensions() []string {
	return []string{".py"}
}

func (F *YapfFormatter) IsInstalled() bool {
	_, err := exec.LookPath("yapf")
	return err == nil
}

func (F *YapfFormatter) FormatToBuffer(args []string, file string, in io.Reader, out io.Writer) error {
	args2 := append([]string{"yapf"}, args...)
	return runIOCommand(args2, in, out)
}

func (F *YapfFormatter) FormatInPlace(args []string, absPath string) error {
	return runIOCommand(append([]string{"yapf", "-i", absPath}, args...), nil, nil)
}
