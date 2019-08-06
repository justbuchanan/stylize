package formatters

// TODO: lines

import (
	"github.com/justbuchanan/stylize/util"
	"io"
	"os/exec"
)

type ClangFormatter struct{}

func (F *ClangFormatter) Name() string {
	return "clang"
}

func (F *ClangFormatter) FileExtensions() []string {
	return []string{".h", ".hpp", ".c", ".cc", ".cpp", ".cxx", ".hxx", ".proto", ".java"}
}

func (F *ClangFormatter) IsInstalled() bool {
	_, err := exec.LookPath("clang-format")
	return err == nil
}

func (F *ClangFormatter) FormatToBuffer(args []string, file util.FileInfo, in io.Reader, out io.Writer) error {
	if len(file.Lines) > 0 {
		args = append([]string{"-lines"})
		// TODO
	}
	return runIOCommand(append([]string{"clang-format"}, args...), in, out)
}

func (F *ClangFormatter) FormatInPlace(args []string, file util.FileInfo) error {
	return runIOCommand(append([]string{"clang-format", "-i", file.Path}, args...), nil, nil)
}
