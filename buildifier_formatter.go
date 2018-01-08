package main

import (
	"io"
	"os/exec"
)

type BuildifierFormatter struct{}

func init() {
	RegisterFormatter("buildifier", &BuildifierFormatter{})
}

func (F *BuildifierFormatter) FileExtensions() []string {
	return []string{".BUILD", "WORKSPACE", "BUILD"}
}

func (F *BuildifierFormatter) IsInstalled() bool {
	_, err := exec.LookPath("buildifier")
	return err == nil
}

func (F *BuildifierFormatter) FormatToBuffer(file string, in io.Reader, out io.Writer) error {
	return runIOCommand([]string{"buildifier"}, in, out)
}

func (F *BuildifierFormatter) FormatInPlace(file string) error {
	return runIOCommand([]string{"buildifier", file}, nil, nil)
}
