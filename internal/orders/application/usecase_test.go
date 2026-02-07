package application

import (
	"context"
	"testing"

	"go-micro/internal/orders/domain"
	"go-micro/internal/orders/ports"
	"go-micro/pkg/errors"
	"go-micro/pkg/logger"
)

// MockOrderRepository is a mock implementation of OrderRepository
type MockOrderRepository struct {
	orders map[uint]*domain.Order
	nextID uint
}

func NewMockOrderRepository() *MockOrderRepository {
	return &MockOrderRepository{
		orders: make(map[uint]*domain.Order),
		nextID: 1,
	}
}

func (m *MockOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	order.ID = m.nextID
	m.nextID++
	m.orders[order.ID] = order
	return nil
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id uint) (*domain.Order, error) {
	order, ok := m.orders[id]
	if !ok {
		return nil, domain.NewOrderNotFound(id)
	}
	return order, nil
}

func (m *MockOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	m.orders[order.ID] = order
	return nil
}

func (m *MockOrderRepository) Delete(ctx context.Context, id uint) error {
	delete(m.orders, id)
	return nil
}

func (m *MockOrderRepository) GetByUserID(ctx context.Context, userID uint) ([]*domain.Order, error) {
	var result []*domain.Order
	for _, order := range m.orders {
		if order.UserID == userID {
			result = append(result, order)
		}
	}
	return result, nil
}

// MockEventPublisher is a mock implementation of EventPublisher
type MockEventPublisher struct {
	events []interface{}
}

func (m *MockEventPublisher) PublishOrderCreated(ctx context.Context, order *domain.Order) error {
	m.events = append(m.events, order)
	return nil
}

// MockUserClient is a mock implementation of UserClient
type MockUserClient struct {
	users map[uint]*ports.UserInfo
}

func NewMockUserClient() *MockUserClient {
	return &MockUserClient{
		users: map[uint]*ports.UserInfo{
			1: {ID: 1, Name: "John Doe", Email: "john@example.com"},
		},
	}
}

func (m *MockUserClient) GetUser(ctx context.Context, userID uint) (*ports.UserInfo, error) {
	user, ok := m.users[userID]
	if !ok {
		return nil, errors.NewNotFound("user", userID)
	}
	return user, nil
}

func TestCreateOrder_Success(t *testing.T) {
	// Arrange
	repo := NewMockOrderRepository()
	publisher := &MockEventPublisher{}
	userClient := NewMockUserClient()
	log := logger.New("test", "debug")
	useCase := NewOrderUseCase(repo, publisher, userClient, log)

	input := CreateOrderInput{
		UserID: 1,
		Total:  99.99,
	}

	// Act
	output, err := useCase.CreateOrder(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if output.Order.ID != 1 {
		t.Errorf("expected ID 1, got %d", output.Order.ID)
	}

	if output.Order.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", output.Order.UserID)
	}

	if output.Order.Total != 99.99 {
		t.Errorf("expected Total 99.99, got %f", output.Order.Total)
	}

	if output.Order.Status != domain.OrderStatusPending {
		t.Errorf("expected status pending, got %s", output.Order.Status)
	}

	if len(publisher.events) != 1 {
		t.Errorf("expected 1 event published, got %d", len(publisher.events))
	}
}

func TestCreateOrder_InvalidTotal(t *testing.T) {
	// Arrange
	repo := NewMockOrderRepository()
	publisher := &MockEventPublisher{}
	userClient := NewMockUserClient()
	log := logger.New("test", "debug")
	useCase := NewOrderUseCase(repo, publisher, userClient, log)

	input := CreateOrderInput{
		UserID: 1,
		Total:  -10.00, // Invalid negative total
	}

	// Act
	_, err := useCase.CreateOrder(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, errors.CodeValidation) {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestCreateOrder_UserNotFound(t *testing.T) {
	// Arrange
	repo := NewMockOrderRepository()
	publisher := &MockEventPublisher{}
	userClient := NewMockUserClient()
	log := logger.New("test", "debug")
	useCase := NewOrderUseCase(repo, publisher, userClient, log)

	input := CreateOrderInput{
		UserID: 999, // Non-existent user
		Total:  99.99,
	}

	// Act
	_, err := useCase.CreateOrder(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, errors.CodeValidation) {
		t.Errorf("expected validation error (user not found), got %v", err)
	}
}

func TestGetOrder_Success(t *testing.T) {
	// Arrange
	repo := NewMockOrderRepository()
	publisher := &MockEventPublisher{}
	userClient := NewMockUserClient()
	log := logger.New("test", "debug")
	useCase := NewOrderUseCase(repo, publisher, userClient, log)

	// Create order first
	createInput := CreateOrderInput{
		UserID: 1,
		Total:  99.99,
	}
	createOutput, _ := useCase.CreateOrder(context.Background(), createInput)

	// Act
	getInput := GetOrderInput{ID: createOutput.Order.ID}
	output, err := useCase.GetOrder(context.Background(), getInput)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if output.Order.ID != createOutput.Order.ID {
		t.Errorf("expected ID %d, got %d", createOutput.Order.ID, output.Order.ID)
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	// Arrange
	repo := NewMockOrderRepository()
	publisher := &MockEventPublisher{}
	userClient := NewMockUserClient()
	log := logger.New("test", "debug")
	useCase := NewOrderUseCase(repo, publisher, userClient, log)

	// Act
	input := GetOrderInput{ID: 999}
	_, err := useCase.GetOrder(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, errors.CodeNotFound) {
		t.Errorf("expected not found error, got %v", err)
	}
}
