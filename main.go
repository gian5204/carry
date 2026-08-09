package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
        fmt.Println("Usage: carry <command> [arguments]")
        return
    }

	command := os.Args[1]

    switch command {
    case "add":
		if len(os.Args) < 3 {
			fmt.Println("Usage: carry add <path>")
			return
		}

		err := addCommand(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}

    default:
        fmt.Printf("Unknown command: %s\n", command)
    }
}
