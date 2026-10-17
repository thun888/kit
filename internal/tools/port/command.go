package port

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"kit/internal/model"
	"kit/internal/system"
	"kit/internal/ui"
)

// Run executes the port tool subcommands.
func Run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "help", "-h", "--help":
		printUsage()
		return nil
	case "kill":
		return runKill(args[1:])
	case "list":
		return runList(args[1:])
	default:
		return fmt.Errorf("unknown port action %q", args[0])
	}
}

func runKill(args []string) error {
	fs := flag.NewFlagSet("port kill", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	force := fs.Bool("force", false, "kill without confirmation")
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: kit port kill [--force] [--json] <port>")
	}
	port, err := strconv.Atoi(fs.Arg(0))
	if err != nil {
		return fmt.Errorf("invalid port %q: %w", fs.Arg(0), err)
	}

	processes, err := system.FindByPort(port)
	if err != nil {
		return err
	}
	if len(processes) == 0 {
		fmt.Printf("no process found on port %d\n", port)
		return nil
	}

	if !*jsonOutput {
		printProcesses(processes)
	}
	if !*force {
		confirmed, err := ui.Confirm(fmt.Sprintf("kill %d process(es) on port %d?", len(processes), port))
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("aborted")
			return nil
		}
	}

	failed := 0
	for _, process := range processes {
		if err := system.KillPID(process.PID); err != nil {
			failed++
			fmt.Fprintf(os.Stderr, "failed to kill pid %d: %v\n", process.PID, err)
		}
	}
	if failed > 0 {
		return fmt.Errorf("killed %d process(es), %d failed", len(processes)-failed, failed)
	}
	if *jsonOutput {
		return printJSON(map[string]any{
			"port":      port,
			"killed":    len(processes),
			"processes": processes,
		})
	}
	fmt.Printf("killed %d process(es) on port %d\n", len(processes), port)
	return nil
}

func runList(args []string) error {
	fs := flag.NewFlagSet("port list", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}

	start, end, err := parsePortRange(fs.Args())
	if err != nil {
		return err
	}

	processes, err := system.FindByPortRange(start, end)
	if err != nil {
		return err
	}
	if *jsonOutput {
		return printJSON(map[string]any{
			"start":     start,
			"end":       end,
			"processes": processes,
		})
	}
	if len(processes) == 0 {
		fmt.Printf("no process found in range %d-%d\n", start, end)
		return nil
	}
	printProcesses(processes)
	return nil
}

func parsePortRange(args []string) (int, int, error) {
	if len(args) == 0 {
		return 0, 0, errors.New("usage: kit port list [--json] <start-end> or <start> <end>")
	}
	if len(args) == 1 {
		parts := strings.Split(args[0], "-")
		if len(parts) != 2 {
			return 0, 0, fmt.Errorf("invalid port range %q", args[0])
		}
		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid range start %q: %w", parts[0], err)
		}
		end, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid range end %q: %w", parts[1], err)
		}
		return start, end, nil
	}
	if len(args) == 2 {
		start, err := strconv.Atoi(args[0])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid range start %q: %w", args[0], err)
		}
		end, err := strconv.Atoi(args[1])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid range end %q: %w", args[1], err)
		}
		return start, end, nil
	}
	return 0, 0, fmt.Errorf("invalid port range args: %v", args)
}

func printUsage() {
	fmt.Println("kit port <action> [args]")
	fmt.Println()
	fmt.Println("Actions:")
	fmt.Println("  kill   Look up the process on a port and kill it after confirmation")
	fmt.Println("  list   List processes occupying ports in a range")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  kit port kill 8080")
	fmt.Println("  kit port kill --force 8080")
	fmt.Println("  kit port list 8000-9000")
	fmt.Println("  kit port list --json 8000 9000")
}

func printProcesses(processes []model.ProcessInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PORT\tPROTOCOL\tPID\tUSER\tNAME\tCOMMAND\tADDR\tSTATE")
	for _, process := range processes {
		fmt.Fprintf(w, "%d\t%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
			process.Port,
			process.Protocol,
			process.PID,
			process.User,
			process.Name,
			process.Command,
			process.LocalAddr,
			process.State,
		)
	}
	_ = w.Flush()
}

func printJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
