package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {

	AuthVaultixApp := NewAuthVaultix(
		"test_app",
		"5d36476ca4",
		"7b9729387300a04a9a128f2dbe8a9b24659047ab7933ab312dfdca3d5397fb59",
		"1.0",
	)

	fmt.Println("Connecting...")
	AuthVaultixApp.Init()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n[1] Login\n[2] Register\n[3] License Login\n[4] Upgrade\n[5] Forgot Password\n[6] Exit")
		fmt.Print("Choose option: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

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
			AuthVaultixApp.Register(strings.TrimSpace(username), strings.TrimSpace(password), strings.TrimSpace(license), "")

		case "3":
			fmt.Print("License: ")
			license, _ := reader.ReadString('\n')
			AuthVaultixApp.LicenseLogin(strings.TrimSpace(license))

		case "4":
			fmt.Print("Username: ")
			username, _ := reader.ReadString('\n')
			fmt.Print("License: ")
			license, _ := reader.ReadString('\n')
			AuthVaultixApp.Upgrade(strings.TrimSpace(username), strings.TrimSpace(license))

		case "5":
			fmt.Print("Username: ")
			username, _ := reader.ReadString('\n')
			fmt.Print("Email: ")
			email, _ := reader.ReadString('\n')
			AuthVaultixApp.ForgotPassword(strings.TrimSpace(username), strings.TrimSpace(email))

		case "6":
			fmt.Println("Goodbye!")
			os.Exit(0)

		default:
			fmt.Println("Invalid option!")
		}
	}
}
