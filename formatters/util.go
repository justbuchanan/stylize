package formatters

import (
	"bytes"
	"io"
	"log"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
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
		log.Print("Error running command: ", strings.Join(args, " "))
		return errors.Wrap(err, stderr.String())
	}

	return nil
}
