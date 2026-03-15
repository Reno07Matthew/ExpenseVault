package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Developer struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	APIHash   string `json:"api_hash"`
	CreatedAt string `json:"created_at"`
}

const fileName = "apikeys.json"

func loadDevelopers() []Developer {

	// If file does not exist create it
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		os.WriteFile(fileName, []byte("[]"), 0644)
	}

	data, err := os.ReadFile(fileName)

	if err != nil {
		fmt.Println("Error reading file.")
		return []Developer{}
	}

	var devs []Developer

	err = json.Unmarshal(data, &devs)

	if err != nil {
		fmt.Println("Warning: apikeys.json is corrupted. Starting with empty registry.")
		return []Developer{}
	}

	return devs
}

func saveDevelopers(devs []Developer) {

	data, err := json.MarshalIndent(devs, "", "  ")

	if err != nil {
		fmt.Println("Error saving data.")
		return
	}

	os.WriteFile(fileName, data, 0644)
}

func registerDeveloper(devs []Developer) []Developer {

	var name, email, apiKey string

	fmt.Print("Enter developer name: ")
	fmt.Scanln(&name)

	fmt.Print("Enter email: ")
	fmt.Scanln(&email)

	fmt.Print("Enter API key: ")
	fmt.Scanln(&apiKey)

	// duplicate email check
	for _, d := range devs {
		if d.Email == email {
			fmt.Println("Email already registered.")
			return devs
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)

	if err != nil {
		fmt.Println("Error hashing API key")
		return devs
	}

	newDev := Developer{
		Name:      name,
		Email:     email,
		APIHash:   string(hash),
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	devs = append(devs, newDev)

	fmt.Println("Developer registered successfully.")

	return devs
}

func verifyAPIKey(devs []Developer) {

	var email, apiKey string

	fmt.Print("Enter email: ")
	fmt.Scanln(&email)

	fmt.Print("Enter API key: ")
	fmt.Scanln(&apiKey)

	for _, d := range devs {

		if d.Email == email {

			err := bcrypt.CompareHashAndPassword(
				[]byte(d.APIHash),
				[]byte(apiKey),
			)

			if err == nil {
				fmt.Println("API key verified. Access granted.")
			} else {
				fmt.Println("Invalid API key.")
			}

			return
		}
	}

	fmt.Println("Developer not found.")
}

func listDevelopers(devs []Developer) {

	fmt.Println("\nRegistered Developers\n")

	for _, d := range devs {

		fmt.Println("Name:", d.Name)
		fmt.Println("Email:", d.Email)
		fmt.Println("Created:", d.CreatedAt)
		fmt.Println()
	}
}

func main() {

	devs := loadDevelopers()

	for {

		fmt.Println("\n1. Register Developer")
		fmt.Println("2. Verify API Key")
		fmt.Println("3. List Developers")
		fmt.Println("4. Exit")

		var choice int

		fmt.Print("Choose option: ")
		fmt.Scanln(&choice)

		switch choice {

		case 1:
			devs = registerDeveloper(devs)
			saveDevelopers(devs)

		case 2:
			verifyAPIKey(devs)

		case 3:
			listDevelopers(devs)

		case 4:
			fmt.Println("Exiting...")
			return

		default:
			fmt.Println("Invalid choice")
		}
	}
}
