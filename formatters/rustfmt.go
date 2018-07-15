package formatters

import (
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

func (F *RustfmtFormatter) FormatToBuffer(args []string, file string, in io.Reader, out io.Writer) error {
	return runIOCommand(append([]string{"rustfmt"}, args...), in, out)
}

func (F *RustfmtFormatter) FormatInPlace(args []string, absPath string) error {
	return runIOCommand(append([]string{"rustfmt", absPath}, args...), nil, nil)
}
