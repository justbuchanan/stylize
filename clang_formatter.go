package main

import (
	"io"
	"os/exec"
)

type ClangFormatter struct{}

func (F *ClangFormatter) Name() string {
	return "clang"
}

func (F *ClangFormatter) FileExtensions() []string {
	return []string{".h", ".hpp", ".c", ".cc", ".cpp", ".proto", ".java"}
}

func (F *ClangFormatter) IsInstalled() bool {
	_, err := exec.LookPath("clang-format")
	return err == nil
}

func (F *ClangFormatter) FormatToBuffer(args []string, file string, in io.Reader, out io.Writer) error {
	return runIOCommand(append([]string{"clang-format"}, args...), in, out)
}

func (F *ClangFormatter) FormatInPlace(args []string, file string) error {
	return runIOCommand(append([]string{"clang-format", "-i", file}, args...), nil, nil)
}
