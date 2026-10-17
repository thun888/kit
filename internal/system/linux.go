package system

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"kit/internal/model"
)

type socketEntry struct {
	protocol string
	state    string
	localIP  string
	localPort int
	uid      string
	inode    string
	localRaw string
}

// FindByPort returns all processes bound to the given port.
func FindByPort(port int) ([]model.ProcessInfo, error) {
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("port out of range: %d", port)
	}

	entries, err := readSocketEntries()
	if err != nil {
		return nil, err
	}
	matches := filterEntriesByPort(entries, port)
	if len(matches) == 0 {
		return nil, nil
	}

	inodeToPIDs, err := mapInodesToPIDs()
	if err != nil {
		return nil, err
	}

	users, err := loadUsers()
	if err != nil {
		return nil, err
	}

	results := make([]model.ProcessInfo, 0, len(matches))
	seen := make(map[string]struct{})
	for _, entry := range matches {
		pids := inodeToPIDs[entry.inode]
		for _, pid := range pids {
			key := entry.inode + ":" + strconv.Itoa(pid)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}

			info, err := processInfoForPID(pid, users)
			if err != nil {
				continue
			}
			info.Port = port
			info.Protocol = entry.protocol
			info.State = entry.state
			info.LocalAddr = entry.localRaw
			info.LocalIP = entry.localIP
			info.LocalPort = entry.localPort
			info.Inode = entry.inode
			results = append(results, info)
		}
	}

	return results, nil
}

// FindByPortRange returns all processes whose port is inside the inclusive range.
func FindByPortRange(start, end int) ([]model.ProcessInfo, error) {
	if start < 1 || end < 1 || start > 65535 || end > 65535 {
		return nil, fmt.Errorf("port range out of range: %d-%d", start, end)
	}
	if start > end {
		return nil, fmt.Errorf("invalid port range: %d-%d", start, end)
	}

	entries, err := readSocketEntries()
	if err != nil {
		return nil, err
	}
	inodeToPIDs, err := mapInodesToPIDs()
	if err != nil {
		return nil, err
	}
	users, err := loadUsers()
	if err != nil {
		return nil, err
	}

	results := make([]model.ProcessInfo, 0)
	seen := make(map[string]struct{})
	for _, entry := range entries {
		if entry.localPort < start || entry.localPort > end {
			continue
		}
		for _, pid := range inodeToPIDs[entry.inode] {
			key := entry.inode + ":" + strconv.Itoa(pid)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}

			info, err := processInfoForPID(pid, users)
			if err != nil {
				continue
			}
			info.Port = entry.localPort
			info.Protocol = entry.protocol
			info.State = entry.state
			info.LocalAddr = entry.localRaw
			info.LocalIP = entry.localIP
			info.LocalPort = entry.localPort
			info.Inode = entry.inode
			results = append(results, info)
		}
	}

	return results, nil
}

func KillPID(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(syscall.SIGTERM)
}

func processInfoForPID(pid int, users map[string]string) (model.ProcessInfo, error) {
	name, cmdline, uid, err := readProcessMetadata(pid)
	if err != nil {
		return model.ProcessInfo{}, err
	}
	user := users[uid]
	if user == "" {
		user = uid
	}
	return model.ProcessInfo{
		PID:     pid,
		Name:    name,
		Command: cmdline,
		User:    user,
	}, nil
}

func readProcessMetadata(pid int) (name string, cmdline string, uid string, err error) {
	statusPath := filepath.Join("/proc", strconv.Itoa(pid), "status")
	statusFile, err := os.Open(statusPath)
	if err != nil {
		return "", "", "", err
	}
	defer statusFile.Close()

	scanner := bufio.NewScanner(statusFile)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "Name:\t"):
			name = strings.TrimSpace(strings.TrimPrefix(line, "Name:"))
		case strings.HasPrefix(line, "Uid:\t"):
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				uid = fields[1]
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", "", err
	}

	cmdPath := filepath.Join("/proc", strconv.Itoa(pid), "cmdline")
	cmdBytes, err := os.ReadFile(cmdPath)
	if err == nil {
		parts := strings.Split(strings.TrimRight(string(cmdBytes), "\x00"), "\x00")
		filtered := parts[:0]
		for _, part := range parts {
			if part != "" {
				filtered = append(filtered, part)
			}
		}
		cmdline = strings.Join(filtered, " ")
	}
	if cmdline == "" {
		cmdline = name
	}
	if name == "" {
		name = cmdline
	}
	return name, cmdline, uid, nil
}

func readSocketEntries() ([]socketEntry, error) {
	files := []struct {
		path     string
		protocol string
	}{
		{"/proc/net/tcp", "tcp"},
		{"/proc/net/tcp6", "tcp6"},
		{"/proc/net/udp", "udp"},
		{"/proc/net/udp6", "udp6"},
	}

	var entries []socketEntry
	for _, file := range files {
		current, err := parseProcNet(file.path, file.protocol)
		if err != nil {
			return nil, err
		}
		entries = append(entries, current...)
	}
	return entries, nil
}

func parseProcNet(path string, protocol string) ([]socketEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var entries []socketEntry
	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		if lineNo == 1 {
			continue
		}
		fields := strings.Fields(scanner.Text())
		if len(fields) < 10 {
			continue
		}
		localIP, localPort, localRaw, err := decodeAddress(fields[1])
		if err != nil {
			continue
		}
		entries = append(entries, socketEntry{
			protocol: protocol,
			state:    decodeState(fields[3]),
			localIP:  localIP,
			localPort: localPort,
			uid:      fields[7],
			inode:    fields[9],
			localRaw: localRaw,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func decodeAddress(raw string) (string, int, string, error) {
	parts := strings.Split(raw, ":")
	if len(parts) != 2 {
		return "", 0, "", fmt.Errorf("invalid address: %s", raw)
	}
	port, err := strconv.ParseInt(parts[1], 16, 32)
	if err != nil {
		return "", 0, "", err
	}
	if len(parts[0]) == 8 {
		b, err := hex.DecodeString(parts[0])
		if err != nil || len(b) != 4 {
			return "", 0, "", fmt.Errorf("invalid ipv4 hex: %s", parts[0])
		}
		ip := net.IPv4(b[3], b[2], b[1], b[0]).String()
		return ip, int(port), raw, nil
	}
	if len(parts[0]) == 32 {
		b, err := hex.DecodeString(parts[0])
		if err != nil || len(b) != 16 {
			return "", 0, "", fmt.Errorf("invalid ipv6 hex: %s", parts[0])
		}
		for i := 0; i < len(b); i += 4 {
			b[i], b[i+3] = b[i+3], b[i]
			b[i+1], b[i+2] = b[i+2], b[i+1]
		}
		ip := net.IP(b).String()
		return ip, int(port), raw, nil
	}
	return "", 0, "", fmt.Errorf("unexpected address width: %s", raw)
}

func decodeState(raw string) string {
	switch raw {
	case "01":
		return "ESTABLISHED"
	case "02":
		return "SYN_SENT"
	case "03":
		return "SYN_RECV"
	case "04":
		return "FIN_WAIT1"
	case "05":
		return "FIN_WAIT2"
	case "06":
		return "TIME_WAIT"
	case "07":
		return "CLOSE"
	case "08":
		return "CLOSE_WAIT"
	case "09":
		return "LAST_ACK"
	case "0A":
		return "LISTEN"
	case "0B":
		return "CLOSING"
	default:
		return raw
	}
}

func filterEntriesByPort(entries []socketEntry, port int) []socketEntry {
	filtered := make([]socketEntry, 0)
	for _, entry := range entries {
		if entry.localPort == port {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func mapInodesToPIDs() (map[string][]int, error) {
	dirs, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	result := make(map[string][]int)
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
		for _, fd := range fds {
			linkPath := filepath.Join(fdDir, fd.Name())
			target, err := os.Readlink(linkPath)
			if err != nil {
				continue
			}
			inode, ok := socketInode(target)
			if !ok {
				continue
			}
			result[inode] = append(result[inode], pid)
		}
	}
	return result, nil
}

func socketInode(target string) (string, bool) {
	if !strings.HasPrefix(target, "socket:[") || !strings.HasSuffix(target, "]") {
		return "", false
	}
	return strings.TrimSuffix(strings.TrimPrefix(target, "socket:["), "]"), true
}

func loadUsers() (map[string]string, error) {
	file, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	users := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ":")
		if len(fields) < 3 {
			continue
		}
		users[fields[2]] = fields[0]
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
