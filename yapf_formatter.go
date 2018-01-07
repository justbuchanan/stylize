package main

import (
	"flag"
	"io"
	"os/exec"
)

var (
	yapfStyleFlag = flag.String("yapf_style", "", "Style to pass to yapf. See `yapf --help` for info.")
)

type YapfFormatter struct{}

func init() {
	RegisterFormatter("yapf", &YapfFormatter{})
}

func (F *YapfFormatter) FileExtensions() []string {
	return []string{".py"}
}

func (F *YapfFormatter) IsInstalled() bool {
	_, err := exec.LookPath("yapf")
	return err == nil
}

func maybeAppendYapfStyleArgs(args []string) []string {
	if len(*yapfStyleFlag) > 0 {
		return append(args, "--style", *yapfStyleFlag)
	}
	return args
}

func (F *YapfFormatter) FormatToBuffer(in io.Reader, out io.Writer) error {
	args := maybeAppendYapfStyleArgs([]string{"yapf"})
	return runIOCommand(args, in, out)
}

func (F *YapfFormatter) FormatInPlace(absPath string) error {
	args := maybeAppendYapfStyleArgs([]string{"yapf", "-i", absPath})
	return runIOCommand(args, nil, nil)
}
