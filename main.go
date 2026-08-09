package main

import "fmt"

func main() {
	err := addCommand(".env")
	if err != nil {
		fmt.Println(err)
	}
}
