package main

import (
	"io"
	"os/exec"
)

type BuildifierFormatter struct{}

func (F *BuildifierFormatter) Name() string {
	return "buildifier"
}

func (F *BuildifierFormatter) FileExtensions() []string {
	return []string{".BUILD", ".bzl", "WORKSPACE", "BUILD"}
}

func (F *BuildifierFormatter) IsInstalled() bool {
	_, err := exec.LookPath("buildifier")
	return err == nil
}

func (F *BuildifierFormatter) FormatToBuffer(args []string, file string, in io.Reader, out io.Writer) error {
	return runIOCommand([]string{"buildifier"}, in, out)
}

func (F *BuildifierFormatter) FormatInPlace(args []string, file string) error {
	return runIOCommand([]string{"buildifier", file}, nil, nil)
}
