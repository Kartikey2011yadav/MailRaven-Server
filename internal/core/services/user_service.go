package services

import (
	"context"
	"errors"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
)

// UserService handles business logic for user management
type UserService struct {
	userRepo ports.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo ports.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// UpdateQuota sets the storage quota for a user
func (s *UserService) UpdateQuota(ctx context.Context, email string, bytes int64) error {
	if bytes < 0 {
		return errors.New("quota cannot be negative")
	}
	return s.userRepo.UpdateQuota(ctx, email, bytes)
}
