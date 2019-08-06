package formatters

// https://github.com/prettier/prettier

import (
	"github.com/justbuchanan/stylize/util"
	"io"
	"os/exec"
)

type PrettierFormatter struct{}

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

func (F *PrettierFormatter) FormatToBuffer(args []string, file util.FileInfo, in io.Reader, out io.Writer) error {
	return runIOCommand(append([]string{"prettier", "--stdin-filepath", file.Path}, args...), in, out)
}

func (F *PrettierFormatter) FormatInPlace(args []string, file util.FileInfo) error {
	return runIOCommand(append([]string{"prettier", "--write", file.Path}, args...), nil, nil)
}
