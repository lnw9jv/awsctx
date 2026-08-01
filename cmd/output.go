package cmd

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

func writeOutput(w io.Writer, output string) error {
	n, err := io.WriteString(w, output)
	if err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	if n != len(output) {
		return fmt.Errorf("write output: %w", io.ErrShortWrite)
	}
	return nil
}

func validateExportValue(name, value string) error {
	if strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return fmt.Errorf("%s contains control characters", name)
	}
	return nil
}
