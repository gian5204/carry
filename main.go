package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gian5204/carry/cmd"
	"github.com/gian5204/carry/internal/ui"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printHelp(os.Stdout)
		return
	}

	command := os.Args[1]

	switch command {
	case "help", "--help", "-h":
		printHelp(os.Stdout)

	case "version", "--version", "-v":
		fmt.Println(versionText())

	case "add":
		if len(os.Args) < 3 {
			fmt.Println(ui.Usage("add", "<path...>"))
			return
		}

		err := cmd.Add(os.Args[2:])
		if err != nil {
			fmt.Println(err)
			return
		}

	case "list":
		err := cmd.List()
		if err != nil {
			fmt.Println(err)
			return
		}

	case "discover":
		err := cmd.Discover()
		if err != nil {
			fmt.Println(err)
			return
		}

	case "remove":
		if len(os.Args) < 3 {
			fmt.Println(ui.Usage("remove", "<path...>"))
			return
		}

		err := cmd.Remove(os.Args[2:])
		if err != nil {
			fmt.Println(err)
			return
		}

	case "copy":
		if len(os.Args) < 3 {
			fmt.Println(ui.Usage("copy", "<destination>"))
			return
		}

		err := cmd.Copy(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Run 'carry --help' for usage.")
	}
}

func printHelp(output io.Writer) {
	commands := []struct {
		name        string
		arguments   string
		description string
	}{
		{name: "add", arguments: "<path...>", description: "Add files to Carry"},
		{name: "remove", arguments: "<path...>", description: "Stop managing files"},
		{name: "list", description: "List managed files"},
		{name: "discover", description: "Discover ignored local files"},
		{name: "copy", arguments: "<destination>", description: "Copy managed files to another clone"},
		{name: "version", description: "Show Carry version"},
		{name: "help", description: "Show this help"},
	}

	fmt.Fprintf(output, "%s — move local config between Git clones\n\n", ui.Bold("Carry"))
	fmt.Fprintf(output, "%s\n", ui.Bold("Usage:"))
	fmt.Fprintln(output, "  carry <command> [arguments]")
	fmt.Fprintln(output)
	fmt.Fprintf(output, "%s\n", ui.Bold("Commands:"))

	for _, command := range commands {
		labelWidth := len(command.name)
		fmt.Fprintf(output, "  %s", ui.Cyan(command.name))
		if command.arguments != "" {
			labelWidth += 1 + len(command.arguments)
			fmt.Fprintf(output, " %s", ui.Dim(command.arguments))
		}

		fmt.Fprintf(
			output,
			"%s%s\n",
			strings.Repeat(" ", 22-labelWidth),
			command.description,
		)
	}
}

func versionText() string {
	return fmt.Sprintf("Carry %s", version)
}
