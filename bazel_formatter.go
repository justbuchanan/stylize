package main

import (
	"io"
	"os/exec"
)

type BazelFormatter struct{}

func init() {
	RegisterFormatter("bazel", &BazelFormatter{})
}

func (F *BazelFormatter) FileExtensions() []string {
	return []string{".BUILD", "WORKSPACE", "BUILD"}
}

func (F *BazelFormatter) IsInstalled() bool {
	_, err := exec.LookPath("buildifier")
	return err == nil
}

func (F *BazelFormatter) FormatToBuffer(in io.Reader, out io.Writer) error {
	return runIOCommand([]string{"buildifier"}, in, out)
}

func (F *BazelFormatter) FormatInPlace(file string) error {
	return runIOCommand([]string{"buildifier", file}, nil, nil)
}
