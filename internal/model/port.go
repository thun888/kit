package model

type ProcessInfo struct {
	Port       int      `json:"port"`
	Protocol   string   `json:"protocol"`
	State      string   `json:"state,omitempty"`
	PID        int      `json:"pid"`
	Name       string   `json:"name"`
	Command    string   `json:"command"`
	User       string   `json:"user"`
	LocalAddr  string   `json:"localAddress"`
	LocalIP    string   `json:"localIP,omitempty"`
	LocalPort  int      `json:"localPort,omitempty"`
	Inode      string   `json:"inode"`
	SocketRefs []string `json:"socketRefs,omitempty"`
}
