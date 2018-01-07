package main

import (
	"bytes"
	"github.com/pkg/errors"
	"io"
	"log"
	"os/exec"
	"path/filepath"
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
