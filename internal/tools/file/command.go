package filetool

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"kit/internal/model"
	"kit/internal/system"
)

// Run executes the file tool subcommands.
func Run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "help", "-h", "--help":
		printUsage()
		return nil
	case "scan":
		return runScan(args[1:])
	default:
		return fmt.Errorf("unknown file action %q", args[0])
	}
}

func runScan(args []string) error {
	fs := flag.NewFlagSet("file scan", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: kit file scan [--json] <path>")
	}

	path := fs.Arg(0)
	usages, err := system.FindFileUsers(path)
	if err != nil {
		return err
	}
	if len(usages) == 0 {
		fmt.Printf("no process is using %s\n", path)
		return nil
	}

	if *jsonOutput {
		return printJSON(map[string]any{
			"path":   path,
			"usages": usages,
		})
	}

	printUsages(usages)
	return nil
}

func printUsage() {
	fmt.Println("kit file <action> [args]")
	fmt.Println()
	fmt.Println("Actions:")
	fmt.Println("  scan   Scan which processes are using a file")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  kit file scan /var/log/syslog")
	fmt.Println("  kit file scan --json /var/log/syslog")
}

func printUsages(usages []model.FileUsageInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PID\tUSER\tNAME\tFD\tACCESS\tTARGET\tCOMMAND")
	for _, usage := range usages {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			usage.PID,
			usage.User,
			usage.Name,
			usage.FD,
			usage.Access,
			usage.FDTarget,
			usage.Command,
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
