package formatters

import (
	"github.com/justbuchanan/stylize/util"
	"io"
)

type GofmtFormatter struct{}

func (F *GofmtFormatter) Name() string {
	return "gofmt"
}

func (F *GofmtFormatter) FileExtensions() []string {
	return []string{".go"}
}

func (F *GofmtFormatter) IsInstalled() bool {
	// if we're running go code, then gofmt has to be here...
	return true
}

func (F *GofmtFormatter) FormatToBuffer(args []string, file util.FileInfo, in io.Reader, out io.Writer) error {
	return runIOCommand([]string{"gofmt"}, in, out)
}

func (F *GofmtFormatter) FormatInPlace(args []string, file util.FileInfo) error {
	return runIOCommand([]string{"gofmt", "-l", "-w", file.Path}, nil, nil)
}
