package formatters

import (
	"github.com/justbuchanan/stylize/util"
	"io"
	"os/exec"
)

type UncrustifyFormatter struct{}

func (F *UncrustifyFormatter) Name() string {
	return "uncrustify"
}

func (F *UncrustifyFormatter) FileExtensions() []string {
	return []string{".h", ".hpp", ".c", ".cc", ".cpp"}
}

func (F *UncrustifyFormatter) IsInstalled() bool {
	_, err := exec.LookPath("uncrustify")
	return err == nil
}

func (F *UncrustifyFormatter) FormatToBuffer(args []string, file util.FileInfo, in io.Reader, out io.Writer) error {
	return runIOCommand(append([]string{"uncrustify", "-q"}, args...), in, out)
}

func (F *UncrustifyFormatter) FormatInPlace(args []string, file util.FileInfo) error {
	return runIOCommand(append([]string{"uncrustify", file.Path, "--no-backup"}, args...), nil, nil)
}
