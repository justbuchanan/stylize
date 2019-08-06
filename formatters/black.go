package formatters

import (
	"github.com/justbuchanan/stylize/util"
	"io"
	"os/exec"
)

// https://github.com/ambv/black
type BlackFormatter struct{}

func (F *BlackFormatter) Name() string {
	return "black"
}

func (F *BlackFormatter) FileExtensions() []string {
	return []string{".py"}
}

func (F *BlackFormatter) IsInstalled() bool {
	_, err := exec.LookPath("black")
	return err == nil
}

func (F *BlackFormatter) FormatToBuffer(args []string, file util.FileInfo, in io.Reader, out io.Writer) error {
	return runIOCommand(append(append([]string{"black"}, args...), "-"), in, out)
}

func (F *BlackFormatter) FormatInPlace(args []string, file util.FileInfo) error {
	return runIOCommand(append([]string{"black", file.Path}, args...), nil, nil)
}
