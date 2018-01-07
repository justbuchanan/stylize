package main

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

func absPathOrDie(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}
	return absPath
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
