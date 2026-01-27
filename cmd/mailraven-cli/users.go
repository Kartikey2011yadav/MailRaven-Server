package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type User struct {
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	LastLoginAt time.Time `json:"last_login_at"`
}

func handleUsers(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: users <list|create|delete|role>")
		return
	}

	subcmd := args[0]
	switch subcmd {
	case "list":
		listUsers()
	case "create":
		if len(args) < 3 {
			fmt.Println("Usage: users create <email> <password> [role]")
			return
		}
		role := "user"
		if len(args) >= 4 {
			role = args[3]
		}
		createUser(args[1], args[2], role)
	case "delete":
		if len(args) < 2 {
			fmt.Println("Usage: users delete <email>")
			return
		}
		deleteUser(args[1])
	case "role":
		if len(args) < 3 {
			fmt.Println("Usage: users role <email> <role>")
			return
		}
		updateUserRole(args[1], args[2])
	default:
		fmt.Printf("Unknown user command: %s\n", subcmd)
	}
}

func listUsers() {
	resp, err := apiRequest("GET", "/users", nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("API Error: Status %s\n", resp.Status)
		return
	}

	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		fmt.Printf("Failed to parse response: %v\n", err)
		return
	}

	fmt.Printf("%-30s %-10s %-20s\n", "EMAIL", "ROLE", "CREATED")
	fmt.Println("--------------------------------------------------------------")
	for _, u := range users {
		fmt.Printf("%-30s %-10s %-20s\n", u.Email, u.Role, u.CreatedAt.Format("2006-01-02"))
	}
}

func createUser(email, password, role string) {
	req := map[string]string{
		"email":    email,
		"password": password,
		"role":     role,
	}

	resp, err := apiRequest("POST", "/users", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 {
		fmt.Printf("User %s created.\n", email)
	} else {
		fmt.Printf("Failed: Status %s\n", resp.Status)
	}
}

func deleteUser(email string) {
	resp, err := apiRequest("DELETE", "/users/"+email, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Printf("User %s deleted.\n", email)
	} else {
		fmt.Printf("Failed: Status %s\n", resp.Status)
	}
}

func updateUserRole(email, role string) {
	req := map[string]string{
		"role": role,
	}

	resp, err := apiRequest("PUT", "/users/"+email+"/role", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Printf("User %s updated to role %s.\n", email, role)
	} else {
		fmt.Printf("Failed: Status %s\n", resp.Status)
	}
}
