//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/storage/sqlite"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run scripts/create_user.go <email> <password> [role]")
		os.Exit(1)
	}

	email := os.Args[1]
	password := os.Args[2]
	role := "ADMIN"
	if len(os.Args) > 3 {
		role = os.Args[3]
	}

	dbPath := "./data/mailraven.db"

	fmt.Printf("Connecting to database at %s...\n", dbPath)
	conn, err := sqlite.NewConnection(dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	fmt.Printf("Hashing password...\n")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	repo := sqlite.NewUserRepository(conn.DB)

	// Check if user exists
	existing, err := repo.FindByEmail(context.Background(), email)
	if err == nil && existing != nil {
		log.Fatalf("User %s already exists", email)
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         domain.Role(role),
		CreatedAt:    time.Now(),
		LastLoginAt:  time.Now(),
	}

	fmt.Printf("Creating user %s with role %s...\n", email, role)
	if err := repo.Create(context.Background(), user); err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Println("User created successfully!")
}
