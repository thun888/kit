package model

type FileUsageInfo struct {
	Path         string `json:"path"`
	ResolvedPath string `json:"resolvedPath,omitempty"`
	PID          int    `json:"pid"`
	Name         string `json:"name"`
	Command      string `json:"command"`
	User         string `json:"user"`
	FD           string `json:"fd"`
	FDTarget     string `json:"fdTarget,omitempty"`
	Access       string `json:"access,omitempty"`
}
