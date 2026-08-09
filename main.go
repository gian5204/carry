package main

import "fmt"

func main() {
	repo, err := detectRepo()
	if err != nil {
		fmt.Println("Error detecting git repository:", err)
		return
	}
	fmt.Printf("Repo detected:\n%s\n", repo.Root)
}
