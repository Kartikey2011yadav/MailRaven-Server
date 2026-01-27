//go:build integration
// +build integration

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestE2E_Admin_UserManagement(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.cleanup()

	ctx := context.Background()

	// 1. Setup Admin User directly in DB
	adminPassHash, _ := env.hashPassword("admin123")
	adminUser := &domain.User{
		Email:        "admin@example.com",
		PasswordHash: adminPassHash,
		Role:         domain.RoleAdmin,
		CreatedAt:    time.Now(),
		LastLoginAt:  time.Now(),
	}
	require.NoError(t, env.userRepo.Create(ctx, adminUser))

	// 2. Login as Admin to get Token
	authBody := map[string]string{
		"email":    "admin@example.com",
		"password": "admin123",
	}
	body, _ := json.Marshal(authBody)
	resp, err := http.Post(env.server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResp struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResp)
	token := loginResp.Token
	require.NotEmpty(t, token)

	// 3. Create User via API
	createBody := map[string]string{
		"email":    "newuser@example.com",
		"password": "password123",
		"role":     "user",
	}
	body, _ = json.Marshal(createBody)
	req, _ := http.NewRequest("POST", env.server.URL+"/api/v1/admin/users", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// 4. Verify User Created in DB
	u, err := env.userRepo.FindByEmail(ctx, "newuser@example.com")
	require.NoError(t, err)
	assert.Equal(t, "newuser@example.com", u.Email)
	assert.Equal(t, domain.RoleUser, u.Role)

	// 5. List Users
	req, _ = http.NewRequest("GET", env.server.URL+"/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var users []domain.User
	json.NewDecoder(resp.Body).Decode(&users)
	assert.GreaterOrEqual(t, len(users), 2) // Admin + NewUser

	// 6. Delete User
	req, _ = http.NewRequest("DELETE", env.server.URL+"/api/v1/admin/users/newuser@example.com", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 7. Verify Deleted
	_, err = env.userRepo.FindByEmail(ctx, "newuser@example.com")
	assert.Error(t, err)
}

func (e *testEnvironment) hashPassword(p string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(bytes), err
}
