package app

import (
	"fmt"

	filetool "kit/internal/tools/file"
	"kit/internal/tools/port"
)

// Run dispatches the first CLI argument as the tool name.
func Run(args []string) error {
	if len(args) == 0 {
		printRootUsage()
		return nil
	}

	switch args[0] {
	case "help", "-h", "--help":
		printRootUsage()
		return nil
	case "port":
		return port.Run(args[1:])
	case "file":
		return filetool.Run(args[1:])
	default:
		return fmt.Errorf("unknown tool %q", args[0])
	}
}

func printRootUsage() {
	fmt.Println("kit <tool> [args]")
	fmt.Println()
	fmt.Println("Tools:")
	fmt.Println("  file   Scan which processes are using a file")
	fmt.Println("  port   Port reverse lookup, listing, and kill helpers")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  kit file scan /var/log/syslog")
	fmt.Println("  kit file scan --json /var/log/syslog")
	fmt.Println("  kit port kill 8080")
	fmt.Println("  kit port kill --force 8080")
	fmt.Println("  kit port list 8000-9000")
	fmt.Println("  kit port list --json 8000 9000")
	fmt.Println()
	fmt.Println("Tip: use -h or --help after a tool or subcommand for more usage details.")
}
