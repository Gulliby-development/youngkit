package base

import (
	"bytes"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

// IO provides access to the standard input and output streams.
type IO interface {
	ReadStdin() []byte
	Out(text string)
	Err(text string)
}

// StdIO connects IO to the process's standard streams.
type StdIO struct{}

// ReadStdin returns the bytes read from standard input.
func (StdIO) ReadStdin() []byte {
	data, _ := io.ReadAll(os.Stdin)
	return data
}

// Out writes text unchanged to standard output.
func (StdIO) Out(text string) {
	_, _ = os.Stdout.Write([]byte(text))
}

// Err writes text unchanged to standard error.
func (StdIO) Err(text string) {
	_, _ = os.Stderr.Write([]byte(text))
}

// Words returns args unchanged unless stdin is requested with a single "-".
func Words(input IO, args []string) ([]string, error) {
	if len(args) != 1 || args[0] != "-" {
		return args, nil
	}
	return decodeStdin(input.ReadStdin())
}

// decodeStdin validates stdin as UTF-8, removes one leading BOM, and splits it
// into words using only the four supported ASCII separators.
func decodeStdin(data []byte) ([]string, error) {
	if !utf8.Valid(data) {
		return nil, Fail("<stdin>: input is not valid UTF-8")
	}

	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	words := strings.FieldsFunc(string(data), func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\r' || r == '\n'
	})
	if len(words) == 0 {
		return nil, Fail("<stdin>: no input")
	}
	return words, nil
}
