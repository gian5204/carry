package main

import (
	"fmt"
	"os"

	"github.com/gian5204/carry/cmd"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println(`Usage: carry <command> [arguments]
Commands:
	add <path> [path...]     Add files to Carry
	discover                 Discover unmanaged local files
	list                     List files managed by Carry
	remove <path> [path...]  Remove files from Carry`)
		return
	}

	command := os.Args[1]

	switch command {
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Usage: carry add <path> [path...]")
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
			fmt.Println("Usage: carry remove <path> [path...]")
			return
		}

		err := cmd.Remove(os.Args[2:])
		if err != nil {
			fmt.Println(err)
			return
		}

	case "copy":
		if len(os.Args) < 3 {
			fmt.Println("Usage: carry copy <path>")
			return
		}

		err := cmd.Copy(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}
