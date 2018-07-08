package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/crypto/ssh/terminal"
)

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
