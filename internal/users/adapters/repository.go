package adapters

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"go-micro/internal/users/domain"
	apperrors "go-micro/pkg/errors"
)

// UserModel is the GORM model for users (persistence layer)
type UserModel struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:100;not null"`
	Email     string    `gorm:"size:255;uniqueIndex;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (UserModel) TableName() string {
	return "users"
}

// PostgresUserRepository implements UserRepository using PostgreSQL
type PostgresUserRepository struct {
	db *gorm.DB
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *gorm.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// Migrate runs auto-migration for the user model
func (r *PostgresUserRepository) Migrate() error {
	return r.db.AutoMigrate(&UserModel{})
}

// Create creates a new user
func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	model := toModel(user)

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return result.Error
	}

	// Update domain entity with generated ID
	user.ID = model.ID
	user.CreatedAt = model.CreatedAt
	user.UpdatedAt = model.UpdatedAt

	return nil
}

// GetByID retrieves a user by ID
func (r *PostgresUserRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var model UserModel

	result := r.db.WithContext(ctx).First(&model, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.NewUserNotFound(id)
		}
		return nil, apperrors.NewInternal("failed to get user", result.Error)
	}

	return toDomain(&model), nil
}

// GetByEmail retrieves a user by email
func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model UserModel

	result := r.db.WithContext(ctx).Where("email = ?", email).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFound("user", email)
		}
		return nil, apperrors.NewInternal("failed to get user by email", result.Error)
	}

	return toDomain(&model), nil
}

// Update updates an existing user
func (r *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	model := toModel(user)

	result := r.db.WithContext(ctx).Save(model)
	if result.Error != nil {
		return apperrors.NewInternal("failed to update user", result.Error)
	}

	user.UpdatedAt = model.UpdatedAt
	return nil
}

// Delete deletes a user by ID
func (r *PostgresUserRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&UserModel{}, id)
	if result.Error != nil {
		return apperrors.NewInternal("failed to delete user", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.NewUserNotFound(id)
	}
	return nil
}

// toModel converts a domain entity to a GORM model
func toModel(user *domain.User) *UserModel {
	return &UserModel{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// toDomain converts a GORM model to a domain entity
func toDomain(model *UserModel) *domain.User {
	return &domain.User{
		ID:        model.ID,
		Name:      model.Name,
		Email:     model.Email,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}
