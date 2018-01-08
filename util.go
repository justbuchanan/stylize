package main

import (
	"bytes"
	"fmt"
	"github.com/danwakefield/fnmatch"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

// Helper method that wraps exec.Command
func runIOCommand(args []string, in io.Reader, out io.Writer) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = in
	cmd.Stdout = out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, stderr.String())
	}

	return nil
}

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func isTerminal(fd *os.File) bool {
	return terminal.IsTerminal(int(fd.Fd()))
}

func getTermWidth(scall uintptr) uint {
	ws := &winsize{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(scall),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		panic(errno)
	}
	return uint(ws.Col)
}

func padToWidth(text string, w int) string {
	spCount := w - len(text)
	if spCount < 0 {
		spCount = 0
	}
	sp := strings.Repeat(" ", spCount)
	return fmt.Sprintf("%s%s", text, sp)
}

// Returns a list of files that have changed since the given git diffbase. These
// file paths are relative to the root of the git repo, not necessarily the
// given rootDir.
func gitChangedFiles(rootDir, diffbase string) ([]string, error) {
	cmd := exec.Command("git", "--no-pager", "diff", "--name-only", diffbase)
	cmd.Dir = rootDir
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return nil, errors.Wrap(err, stderr.String())
	}

	// note: these paths are all relative to the git root directory
	changedFiles := strings.Split(strings.Trim(out.String(), "\n"), "\n")

	return changedFiles, nil
}

// TODO: this implementation has a lot of flaws
// It would be nice to do something similar to gitignore
func filePatternMatch(pattern, file string) bool {
	if fnmatch.Match(pattern, file, fnmatch.FNM_PATHNAME|fnmatch.FNM_LEADING_DIR) {
		return true
	}

	// TODO: this is a hack
	hasGlobChars := strings.ContainsAny(pattern, "*?")
	if hasGlobChars {
		return false
	}

	// If pattern ends in '/', ignore all files in that directory recursively
	if strings.HasSuffix(pattern, "/") {
		if strings.HasPrefix(file, pattern) {
			return true
		}
	}

	return false
}

func fileIsExcluded(file string, exclude []string) bool {
	for _, e := range exclude {
		if filePatternMatch(e, file) {
			return true
		}
	}
	return false
}
