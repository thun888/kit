package system

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"kit/internal/model"
)

// FindFileUsers returns processes that currently hold an open handle to the file.
func FindFileUsers(path string) ([]model.FileUsageInfo, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	resolvedPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return nil, fmt.Errorf("resolve file path %q: %w", path, err)
	}
	targetInfo, err := os.Stat(resolvedPath)
	if err != nil {
		return nil, err
	}

	users, err := loadUsers()
	if err != nil {
		return nil, err
	}

	dirs, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	results := make([]model.FileUsageInfo, 0)
	seen := make(map[string]struct{})
	for _, dir := range dirs {
		pid, err := strconv.Atoi(dir.Name())
		if err != nil {
			continue
		}

		fdDir := filepath.Join("/proc", dir.Name(), "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}

		processInfo, err := processInfoForPID(pid, users)
		if err != nil {
			continue
		}

		for _, fd := range fds {
			fdPath := filepath.Join(fdDir, fd.Name())
			fdInfo, err := os.Stat(fdPath)
			if err != nil {
				continue
			}
			if !os.SameFile(targetInfo, fdInfo) {
				continue
			}

			target, _ := os.Readlink(fdPath)
			access, err := readFDInfoAccess(pid, fd.Name())
			if err != nil {
				access = ""
			}

			key := fmt.Sprintf("%d:%s", pid, fd.Name())
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}

			results = append(results, model.FileUsageInfo{
				Path:         absPath,
				ResolvedPath: resolvedPath,
				PID:          processInfo.PID,
				Name:         processInfo.Name,
				Command:      processInfo.Command,
				User:         processInfo.User,
				FD:           fd.Name(),
				FDTarget:     target,
				Access:       access,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].PID == results[j].PID {
			return results[i].FD < results[j].FD
		}
		return results[i].PID < results[j].PID
	})

	return results, nil
}

func readFDInfoAccess(pid int, fd string) (string, error) {
	infoPath := filepath.Join("/proc", strconv.Itoa(pid), "fdinfo", fd)
	data, err := os.ReadFile(infoPath)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "flags:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "flags:")), nil
		}
	}
	return "", nil
}
