package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {

	AuthVaultixApp := Authvaultix{}
	AuthVaultixApp.Api(
		"",
		"",
		"",
		"",
	)
	AuthVaultixApp.Init()

	fmt.Println("\n[1] Login\n[2] Register\n[3] License Login\n[4] Exit")
	fmt.Print("Choose option: ")

	var choice string
	fmt.Scan(&choice)

	reader := bufio.NewReader(os.Stdin)
	// 🧹 FIX: Clear leftover newline from fmt.Scan()
	reader.ReadString('\n')

	switch choice {
	case "1":
		fmt.Print("Username: ")
		username, _ := reader.ReadString('\n')
		fmt.Print("Password: ")
		password, _ := reader.ReadString('\n')
		AuthVaultixApp.Login(strings.TrimSpace(username), strings.TrimSpace(password))

	case "2":
		fmt.Print("Username: ")
		username, _ := reader.ReadString('\n')
		fmt.Print("Password: ")
		password, _ := reader.ReadString('\n')
		fmt.Print("License: ")
		license, _ := reader.ReadString('\n')
		AuthVaultixApp.Register(strings.TrimSpace(username), strings.TrimSpace(password), strings.TrimSpace(license))

	case "3":
		fmt.Print("License: ")
		license, _ := reader.ReadString('\n')
		AuthVaultixApp.License(strings.TrimSpace(license))

	default:
		fmt.Println("Goodbye!")
	}
}
