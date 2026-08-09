package main

import "fmt"

func main() {
	repo, err := detectRepo()
	if err != nil {
		fmt.Println("Error detecting git repository:", err)
		return
	}
	fmt.Printf("Repo detected:\n%s\n", repo.Root)


	tempPath := ".env"
	ignored, err := repo.IsIgnored(tempPath)
	if err != nil {
		fmt.Println("Error checking whether file is ignored:", err)
		return
	}
	if ignored {
		fmt.Printf("%s is ignored\n", tempPath)
	} else {
		fmt.Printf("%s is not ignored\n", tempPath)
	}
}
