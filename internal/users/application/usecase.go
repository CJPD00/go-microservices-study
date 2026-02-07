package application

import (
	"context"

	"go-micro/internal/users/domain"
	"go-micro/internal/users/ports"
	"go-micro/pkg/errors"
	"go-micro/pkg/logger"

	"go.uber.org/zap"
)

// UserUseCase handles user business logic
type UserUseCase struct {
	repo      ports.UserRepository
	publisher ports.EventPublisher
	log       *logger.Logger
}

// NewUserUseCase creates a new user use case
func NewUserUseCase(repo ports.UserRepository, publisher ports.EventPublisher, log *logger.Logger) *UserUseCase {
	return &UserUseCase{
		repo:      repo,
		publisher: publisher,
		log:       log,
	}
}

// CreateUserInput represents the input for creating a user
type CreateUserInput struct {
	Name  string
	Email string
}

// CreateUserOutput represents the output of creating a user
type CreateUserOutput struct {
	User *domain.User
}

// CreateUser creates a new user
func (uc *UserUseCase) CreateUser(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error) {
	// Create domain entity with validation
	user, err := domain.NewUser(input.Name, input.Email)
	if err != nil {
		return nil, err
	}

	// Check if email already exists
	existing, err := uc.repo.GetByEmail(ctx, user.Email)
	if err != nil && !errors.Is(err, errors.CodeNotFound) {
		return nil, errors.NewInternal("failed to check email existence", err)
	}
	if existing != nil {
		return nil, domain.ErrEmailExists
	}

	// Create user in repository
	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, errors.NewInternal("failed to create user", err)
	}

	// Publish event (async, don't fail on error)
	if uc.publisher != nil {
		if err := uc.publisher.PublishUserCreated(ctx, user); err != nil {
			uc.log.WithContext(ctx).Error("failed to publish user created event",
				zap.Error(err),
				zap.Uint("user_id", user.ID),
			)
		}
	}

	uc.log.WithContext(ctx).Info("user created",
		zap.Uint("user_id", user.ID),
		zap.String("email", user.Email),
	)

	return &CreateUserOutput{User: user}, nil
}

// GetUserInput represents the input for getting a user
type GetUserInput struct {
	ID uint
}

// GetUserOutput represents the output of getting a user
type GetUserOutput struct {
	User *domain.User
}

// GetUser retrieves a user by ID
func (uc *UserUseCase) GetUser(ctx context.Context, input GetUserInput) (*GetUserOutput, error) {
	user, err := uc.repo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	return &GetUserOutput{User: user}, nil
}
