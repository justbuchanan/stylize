package main

import (
	"flag"
	"io"
	"os/exec"
)

var (
	clangStyleFlag = flag.String("clang_style", "", "Style to pass to clang-format. See `clang-format --help` for more info.")
)

type ClangFormatter struct{}

func init() {
	RegisterFormatter("clang", &ClangFormatter{})
}

func (F *ClangFormatter) FileExtensions() []string {
	return []string{".h", ".hpp", ".c", ".cc", ".cpp", ".proto", ".java", ".js"}
}

func (F *ClangFormatter) IsInstalled() bool {
	_, err := exec.LookPath("clang-format")
	return err == nil
}

func maybeAppendClangStyleArgs(args []string) []string {
	if len(*clangStyleFlag) > 0 {
		return append(args, "-style", *clangStyleFlag)
	}
	return args
}

func (F *ClangFormatter) FormatToBuffer(in io.Reader, out io.Writer) error {
	args := maybeAppendClangStyleArgs([]string{"clang-format"})
	return runIOCommand(args, in, out)
}

func (F *ClangFormatter) FormatInPlace(absPath string) error {
	args := maybeAppendClangStyleArgs([]string{"clang-format", "-i", absPath})
	return runIOCommand(args, nil, nil)
}
