package application

import (
	"context"
	"testing"

	"go-micro/internal/users/domain"
	"go-micro/pkg/errors"
	"go-micro/pkg/logger"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	users     map[uint]*domain.User
	byEmail   map[string]*domain.User
	nextID    uint
	createFn  func(ctx context.Context, user *domain.User) error
	getByIDFn func(ctx context.Context, id uint) (*domain.User, error)
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:   make(map[uint]*domain.User),
		byEmail: make(map[string]*domain.User),
		nextID:  1,
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	user.ID = m.nextID
	m.nextID++
	m.users[user.ID] = user
	m.byEmail[user.Email] = user
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	user, ok := m.users[id]
	if !ok {
		return nil, domain.NewUserNotFound(id)
	}
	return user, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, ok := m.byEmail[email]
	if !ok {
		return nil, errors.NewNotFound("user", email)
	}
	return user, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	delete(m.users, id)
	return nil
}

// MockEventPublisher is a mock implementation of EventPublisher
type MockEventPublisher struct {
	events []interface{}
}

func (m *MockEventPublisher) PublishUserCreated(ctx context.Context, user *domain.User) error {
	m.events = append(m.events, user)
	return nil
}

func TestCreateUser_Success(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	publisher := &MockEventPublisher{}
	log := logger.New("test", "debug")
	useCase := NewUserUseCase(repo, publisher, log)

	input := CreateUserInput{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	// Act
	output, err := useCase.CreateUser(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if output.User.ID != 1 {
		t.Errorf("expected ID 1, got %d", output.User.ID)
	}

	if output.User.Name != "John Doe" {
		t.Errorf("expected name 'John Doe', got '%s'", output.User.Name)
	}

	if output.User.Email != "john@example.com" {
		t.Errorf("expected email 'john@example.com', got '%s'", output.User.Email)
	}

	if len(publisher.events) != 1 {
		t.Errorf("expected 1 event published, got %d", len(publisher.events))
	}
}

func TestCreateUser_InvalidEmail(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	publisher := &MockEventPublisher{}
	log := logger.New("test", "debug")
	useCase := NewUserUseCase(repo, publisher, log)

	input := CreateUserInput{
		Name:  "John Doe",
		Email: "invalid-email",
	}

	// Act
	_, err := useCase.CreateUser(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, errors.CodeValidation) {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	publisher := &MockEventPublisher{}
	log := logger.New("test", "debug")
	useCase := NewUserUseCase(repo, publisher, log)

	// Create first user
	input1 := CreateUserInput{
		Name:  "John Doe",
		Email: "john@example.com",
	}
	_, _ = useCase.CreateUser(context.Background(), input1)

	// Try to create second user with same email
	input2 := CreateUserInput{
		Name:  "Jane Doe",
		Email: "john@example.com",
	}

	// Act
	_, err := useCase.CreateUser(context.Background(), input2)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, errors.CodeConflict) {
		t.Errorf("expected conflict error, got %v", err)
	}
}

func TestGetUser_Success(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	publisher := &MockEventPublisher{}
	log := logger.New("test", "debug")
	useCase := NewUserUseCase(repo, publisher, log)

	// Create user first
	createInput := CreateUserInput{
		Name:  "John Doe",
		Email: "john@example.com",
	}
	createOutput, _ := useCase.CreateUser(context.Background(), createInput)

	// Act
	getInput := GetUserInput{ID: createOutput.User.ID}
	output, err := useCase.GetUser(context.Background(), getInput)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if output.User.ID != createOutput.User.ID {
		t.Errorf("expected ID %d, got %d", createOutput.User.ID, output.User.ID)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	publisher := &MockEventPublisher{}
	log := logger.New("test", "debug")
	useCase := NewUserUseCase(repo, publisher, log)

	// Act
	input := GetUserInput{ID: 999}
	_, err := useCase.GetUser(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, errors.CodeNotFound) {
		t.Errorf("expected not found error, got %v", err)
	}
}
