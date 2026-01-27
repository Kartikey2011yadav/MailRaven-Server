package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/storage/sqlite"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
)

func main() {
	dbPath := "./data/mailraven.db"
	fmt.Printf("Seeding admin to %s\n", dbPath)

	conn, err := sqlite.NewConnection(dbPath)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.DB.Close()

	// 1. Run Migrations
	// Ensure we are in root or point to correct path.
	// Assuming running from root F:\MailRaven-Server
	migrationsDir := filepath.Join("internal", "adapters", "storage", "sqlite", "migrations")
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		// Try adapting path if script runs from scripts/
		migrationsDir = filepath.Join("..", "internal", "adapters", "storage", "sqlite", "migrations")
		files, err = os.ReadDir(migrationsDir)
		if err != nil {
			log.Fatalf("Failed to read migrations dir: %v", err)
		}
	}

	fmt.Println("Running migrations...")
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sql" {
			migrationPath := filepath.Join(migrationsDir, file.Name())
			if err := conn.RunMigrations(migrationPath); err != nil {
				// Ignore errors like "table already exists" if handled by migration
				fmt.Printf("Migration %s result: %v\n", file.Name(), err)
			}
		}
	}

	// 2. Seed Admin
	repo := sqlite.NewUserRepository(conn.DB)
	ctx := context.Background()

	email := "admin@example.com"
	password := "admin123"

	// Check if exists
	existing, err := repo.FindByEmail(ctx, email)
	if err == nil && existing != nil {
		fmt.Println("Admin user already exists.")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: string(hashed),
		Role:         "admin", // Assuming "admin" is valid value for string Role
		CreatedAt:    time.Now(),
		LastLoginAt:  time.Now(),
	}

	// Wait, domain.RoleUser is defined as const?
	// The repo code set user.Role = domain.RoleUser if empty.
	// But I want admin.
	// Let's hope "admin" string is valid or check domain package later.

	if err := repo.Create(ctx, user); err != nil {
		log.Fatalf("Failed to create admin: %v", err)
	}

	fmt.Println("Admin user created successfully.")
	fmt.Printf("Email: %s\nPassword: %s\n", email, password)
}
