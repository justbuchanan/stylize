package formatters

import (
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

func (F *BlackFormatter) FormatToBuffer(args []string, file string, in io.Reader, out io.Writer) error {
	return runIOCommand(append(append([]string{"black"}, args...), "-"), in, out)
}

func (F *BlackFormatter) FormatInPlace(args []string, file string) error {
	return runIOCommand(append([]string{"black", file}, args...), nil, nil)
}
