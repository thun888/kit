package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Confirm asks a yes/no question on stdin.
func Confirm(message string) (bool, error) {
	fmt.Printf("%s [y/N]: ", message)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}
