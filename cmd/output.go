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

func statusMessage(profile, region string, cleared bool) string {
	switch {
	case cleared:
		return "Cleared AWS profile and region\n"
	case profile != "" && region != "":
		return fmt.Sprintf("Switched to profile %s and region %s\n", profile, region)
	case profile != "":
		return fmt.Sprintf("Switched to profile %s\n", profile)
	default:
		return fmt.Sprintf("Switched to region %s\n", region)
	}
}

func validateExportValue(name, value string) error {
	if strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return fmt.Errorf("%s contains control characters", name)
	}
	return nil
}
