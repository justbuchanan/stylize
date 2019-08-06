package formatters

// TODO: lines?

import (
	"github.com/justbuchanan/stylize/util"
	"io"
	"os/exec"
)

type RustfmtFormatter struct{}

func (F *RustfmtFormatter) Name() string {
	return "rustfmt"
}

func (F *RustfmtFormatter) FileExtensions() []string {
	return []string{".rs"}
}

func (F *RustfmtFormatter) IsInstalled() bool {
	_, err := exec.LookPath("rustfmt")
	return err == nil
}

func (F *RustfmtFormatter) FormatToBuffer(args []string, file util.FileInfo, in io.Reader, out io.Writer) error {
	return runIOCommand(append([]string{"rustfmt"}, args...), in, out)
}

func (F *RustfmtFormatter) FormatInPlace(args []string, file util.FileInfo) error {
	return runIOCommand(append([]string{"rustfmt", file.Path}, args...), nil, nil)
}
