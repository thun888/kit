package system

import (
	"os"
	"testing"
)

func TestFindFileUsersFindsOpenTempFile(t *testing.T) {
	file, err := os.CreateTemp("", "kit-file-usage-*")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	usages, err := FindFileUsers(file.Name())
	if err != nil {
		t.Fatalf("FindFileUsers: %v", err)
	}
	if len(usages) == 0 {
		t.Fatalf("expected at least one usage")
	}

	currentPID := os.Getpid()

	found := false
	for _, usage := range usages {
		if usage.PID == currentPID && usage.FD != "" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected current process %d to be reported, got %#v", currentPID, usages)
	}
}
