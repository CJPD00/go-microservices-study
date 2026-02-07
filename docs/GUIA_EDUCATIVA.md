# 📚 Guía Educativa: Microservicios en Go

Esta guía explica **cada parte del proyecto** para principiantes en Go y microservicios.

---

## 📖 Tabla de Contenidos

1. [¿Qué es un Microservicio?](#1-qué-es-un-microservicio)
2. [Arquitectura del Proyecto](#2-arquitectura-del-proyecto)
3. [Estructura de Carpetas](#3-estructura-de-carpetas)
4. [Clean Architecture](#4-clean-architecture)
5. [Los 3 Servicios](#5-los-3-servicios)
6. [gRPC: Comunicación entre Servicios](#6-grpc-comunicación-entre-servicios)
7. [RabbitMQ: Eventos Asíncronos](#7-rabbitmq-eventos-asíncronos)
8. [Base de Datos con GORM](#8-base-de-datos-con-gorm)
9. [Manejo de Errores](#9-manejo-de-errores)
10. [Middleware y HTTP](#10-middleware-y-http)
11. [Configuración](#11-configuración)
12. [Flujo Completo de una Petición](#12-flujo-completo-de-una-petición)

---

## 1. ¿Qué es un Microservicio?

### Monolito vs Microservicios

**Monolito:** Una sola aplicación grande que hace todo.
```
┌─────────────────────────────┐
│      Aplicación Grande      │
│  Users + Orders + Pagos     │
│  + Inventario + Reportes    │
└─────────────────────────────┘
```

**Microservicios:** Varias aplicaciones pequeñas, cada una hace UNA cosa.
```
┌─────────┐  ┌─────────┐  ┌─────────┐
│  Users  │  │ Orders  │  │ Gateway │
└─────────┘  └─────────┘  └─────────┘
     │            │            │
     └────────────┴────────────┘
           Se comunican
```

### Ventajas
- **Escalabilidad:** Puedes escalar solo lo que necesitas
- **Independencia:** Equipos diferentes pueden trabajar en servicios diferentes
- **Tecnología:** Cada servicio puede usar diferentes tecnologías
- **Resiliencia:** Si uno falla, los demás siguen

### Desventajas
- **Complejidad:** Más piezas = más complejidad
- **Comunicación:** Los servicios deben "hablar" entre sí
- **Datos distribuidos:** Cada servicio tiene su propia base de datos

---

## 2. Arquitectura del Proyecto

```
                    INTERNET
                       │
                       ▼
         ┌─────────────────────────┐
         │      GATEWAY :8080      │  ← Único punto de entrada
         │   (REST API + Swagger)  │
         └───────────┬─────────────┘
                     │
            ┌────────┴────────┐
            │ gRPC (interno)  │
            ▼                 ▼
   ┌─────────────┐    ┌─────────────┐
   │   USERS     │    │   ORDERS    │
   │   :50051    │◄───│   :50052    │  ← Orders llama a Users
   └──────┬──────┘    └──────┬──────┘
          │                  │
          ▼                  ▼
   ┌─────────────┐    ┌─────────────┐
   │  users_db   │    │  orders_db  │  ← Bases de datos separadas
   └─────────────┘    └─────────────┘
          │                  │
          └────────┬─────────┘
                   ▼
          ┌─────────────────┐
          │    RabbitMQ     │  ← Eventos asíncronos
          │  (Mensajería)   │
          └─────────────────┘
```

### ¿Por qué un Gateway?

El **Gateway** es la "puerta de entrada" a todos los servicios:
- Los clientes (frontend, apps) solo conocen al Gateway
- El Gateway traduce peticiones REST a llamadas gRPC
- Centraliza autenticación, logging, rate limiting
- Expone Swagger para documentar la API

---

## 3. Estructura de Carpetas

```
go-micro/
├── api/                    # Contratos de comunicación
│   ├── gen/                # Código generado (gRPC)
│   └── proto/              # Definiciones .proto
│
├── cmd/                    # Punto de entrada de cada servicio
│   ├── gateway/main.go     # Inicia el Gateway
│   ├── users/main.go       # Inicia Users
│   └── orders/main.go      # Inicia Orders
│
├── internal/               # Lógica de negocio (privada)
│   ├── gateway/
│   ├── users/
│   └── orders/
│
├── pkg/                    # Código compartido (público)
│   ├── config/             # Configuración
│   ├── db/                 # Conexión a base de datos
│   ├── errors/             # Manejo de errores
│   ├── grpc/               # Interceptores gRPC
│   ├── logger/             # Logging
│   ├── middleware/         # Middleware HTTP
│   ├── rabbitmq/           # Mensajería
│   └── tls/                # Seguridad TLS
│
├── deploy/                 # Docker Compose
├── docs/                   # Swagger
└── scripts/                # Scripts útiles
```

### Convención `cmd/` vs `internal/` vs `pkg/`

| Carpeta | Propósito | ¿Quién puede usarlo? |
|---------|-----------|---------------------|
| `cmd/` | Punto de entrada (`main.go`) | Solo ese binario |
| `internal/` | Lógica de negocio | Solo este proyecto |
| `pkg/` | Utilidades compartidas | Cualquiera (incluso otros proyectos) |

---

## 4. Clean Architecture

Cada servicio sigue **Clean Architecture** (también llamada Hexagonal o Ports & Adapters):

```
internal/users/
├── domain/           # 🎯 Entidades y reglas de negocio
├── application/      # 📋 Casos de uso
├── ports/            # 🔌 Interfaces (contratos)
├── adapters/         # 🔧 Implementaciones
└── infrastructure/   # 🌐 Servidores HTTP/gRPC
```

### 4.1 Domain (Dominio)

**¿Qué es?** El corazón del negocio. Define QUÉ es un usuario.

```go
// internal/users/domain/entity.go
type User struct {
    ID        uint
    Name      string
    Email     string
    CreatedAt time.Time
}

// Validación: regla de negocio
func (u *User) Validate() error {
    if !strings.Contains(u.Email, "@") {
        return ErrInvalidEmail
    }
    return nil
}
```

**Regla importante:** El dominio NO sabe nada de bases de datos, HTTP, o gRPC.

### 4.2 Ports (Puertos/Interfaces)

**¿Qué es?** Contratos que definen CÓMO interactuar con el exterior.

```go
// internal/users/ports/ports.go
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id uint) (*domain.User, error)
}

type EventPublisher interface {
    PublishUserCreated(ctx context.Context, user *domain.User) error
}
```

**¿Por qué interfaces?** 
- Para testing: puedes crear mocks
- Para flexibilidad: cambiar PostgreSQL por MySQL sin tocar la lógica

### 4.3 Application (Casos de Uso)

**¿Qué es?** La lógica de aplicación. Orquesta el flujo.

```go
// internal/users/application/usecase.go
type UserUseCase struct {
    repo      ports.UserRepository   // ← Usa la interface, no la implementación
    publisher ports.EventPublisher
    log       *logger.Logger
}

func (uc *UserUseCase) CreateUser(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error) {
    // 1. Crear entidad de dominio
    user, err := domain.NewUser(input.Name, input.Email)
    if err != nil {
        return nil, err  // Error de validación
    }
    
    // 2. Guardar en repositorio
    if err := uc.repo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    // 3. Publicar evento
    uc.publisher.PublishUserCreated(ctx, user)
    
    return &CreateUserOutput{User: user}, nil
}
```

### 4.4 Adapters (Adaptadores)

**¿Qué es?** Implementaciones concretas de las interfaces.

```go
// internal/users/adapters/repository.go
type PostgresUserRepository struct {
    db *gorm.DB
}

// Implementa ports.UserRepository
func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
    model := toModel(user)  // Convertir dominio → modelo de DB
    return r.db.WithContext(ctx).Create(model).Error
}
```

### 4.5 Infrastructure (Infraestructura)

**¿Qué es?** Servidores HTTP y gRPC que exponen los casos de uso.

```go
// internal/users/infrastructure/http.go
type HTTPHandler struct {
    useCase *application.UserUseCase
}

func (h *HTTPHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    c.ShouldBindJSON(&req)
    
    output, err := h.useCase.CreateUser(c.Request.Context(), ...)
    if err != nil {
        c.Error(err)
        return
    }
    
    c.JSON(http.StatusCreated, output)
}
```

### Flujo de Dependencias

```
HTTP Request
     │
     ▼
┌─────────────────┐
│ Infrastructure  │  ← Recibe peticiones
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Application   │  ← Ejecuta lógica
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌───────┐ ┌───────────┐
│Ports  │ │  Domain   │  ← Usa interfaces y entidades
└───┬───┘ └───────────┘
    │
    ▼
┌───────────┐
│ Adapters  │  ← Implementa las interfaces
└───────────┘
```

---

## 5. Los 3 Servicios

### 5.1 Users Service

**Responsabilidad:** Gestionar usuarios

| Capa | Archivo | Qué hace |
|------|---------|----------|
| Domain | `entity.go` | Define User |
| Domain | `errors.go` | Errores de negocio |
| Ports | `ports.go` | Interfaces |
| Application | `usecase.go` | CreateUser, GetUser |
| Adapters | `repository.go` | Guarda en PostgreSQL |
| Adapters | `publisher.go` | Publica eventos |
| Infrastructure | `http.go` | REST API |
| Infrastructure | `grpc.go` | Servidor gRPC |

**main.go:**
```go
func main() {
    cfg := config.Load()                    // 1. Cargar config
    db := db.NewConnection(...)             // 2. Conectar DB
    repo := adapters.NewPostgresUserRepository(db)  // 3. Crear repositorio
    publisher := adapters.NewRabbitMQPublisher(...) // 4. Crear publisher
    useCase := application.NewUserUseCase(repo, publisher) // 5. Crear caso de uso
    
    // 6. Iniciar servidores
    go startHTTPServer(useCase)
    go startGRPCServer(useCase)
    
    waitForShutdown()
}
```

### 5.2 Orders Service

**Responsabilidad:** Gestionar órdenes

Similar a Users, pero con una diferencia clave:
- **Llama a Users por gRPC** para validar que el usuario existe antes de crear una orden

```go
// internal/orders/adapters/user_client.go
type GRPCUserClient struct {
    client userspb.UserServiceClient
}

func (c *GRPCUserClient) GetUser(ctx context.Context, userID uint) (*UserInfo, error) {
    // Llama al servicio Users por gRPC
    resp, err := c.client.GetUser(ctx, &userspb.GetUserRequest{Id: uint64(userID)})
    if err != nil {
        return nil, err  // Usuario no existe
    }
    return &UserInfo{ID: uint(resp.Id), Name: resp.Name, ...}, nil
}
```

### 5.3 Gateway

**Responsabilidad:** Punto de entrada único

```go
// El gateway NO tiene lógica de negocio
// Solo reenvía peticiones a los servicios internos

func (h *Handler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    c.ShouldBindJSON(&req)
    
    // Llama al servicio Users por gRPC
    resp, err := h.usersClient.CreateUser(ctx, &userspb.CreateUserRequest{
        Name:  req.Name,
        Email: req.Email,
    })
    
    c.JSON(http.StatusCreated, resp)
}
```

---

## 6. gRPC: Comunicación entre Servicios

### ¿Qué es gRPC?

**gRPC** es un protocolo de comunicación más eficiente que REST:
- Usa **Protocol Buffers** (binario, más compacto que JSON)
- Soporta **streaming**
- Tiene **tipado fuerte** (contratos definidos)

### Flujo

```
1. Defines el contrato (.proto)
2. Generas código Go
3. Implementas el servidor
4. Creas el cliente
```

### Archivo .proto

```protobuf
// api/proto/users/v1/users.proto
syntax = "proto3";
package users.v1;

service UserService {
    rpc GetUser(GetUserRequest) returns (UserResponse);
    rpc CreateUser(CreateUserRequest) returns (UserResponse);
}

message GetUserRequest {
    uint64 id = 1;
}

message UserResponse {
    uint64 id = 1;
    string name = 2;
    string email = 3;
}
```

### Servidor gRPC

```go
// internal/users/infrastructure/grpc.go
type GRPCServer struct {
    userspb.UnimplementedUserServiceServer  // Embedding para compatibilidad
    useCase *application.UserUseCase
}

func (s *GRPCServer) GetUser(ctx context.Context, req *userspb.GetUserRequest) (*userspb.UserResponse, error) {
    output, err := s.useCase.GetUser(ctx, application.GetUserInput{ID: uint(req.Id)})
    if err != nil {
        return nil, err
    }
    return &userspb.UserResponse{
        Id:    uint64(output.User.ID),
        Name:  output.User.Name,
        Email: output.User.Email,
    }, nil
}
```

### Cliente gRPC

```go
// En el Gateway o Orders
conn, _ := grpc.Dial("localhost:50051")
client := userspb.NewUserServiceClient(conn)

resp, err := client.GetUser(ctx, &userspb.GetUserRequest{Id: 1})
```

---

## 7. RabbitMQ: Eventos Asíncronos

### ¿Qué es?

**RabbitMQ** es un sistema de mensajería:
- Un servicio **publica** un mensaje
- Otro servicio **consume** ese mensaje
- Son **independientes** (no se bloquean mutuamente)

### Flujo de Eventos

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Users     │ ──────► │   RabbitMQ   │ ──────► │   Orders    │
│  (publica)  │         │   (cola)     │         │  (consume)  │
└─────────────┘         └──────────────┘         └─────────────┘
      │                                                │
      │  UserCreated                                   │
      │  {id: 1, name: "Juan", email: "..."}           │
      │                                                ▼
                                                 "Nuevo usuario
                                                  registrado!"
```

### Publisher (Publicador)

```go
// internal/users/adapters/publisher.go
func (p *RabbitMQPublisher) PublishUserCreated(ctx context.Context, user *domain.User) error {
    event := events.UserCreatedEvent{
        Type: "user.created",
        Payload: events.UserPayload{
            ID:    user.ID,
            Name:  user.Name,
            Email: user.Email,
        },
    }
    return p.publisher.Publish(ctx, "user.created", event)
}
```

### Consumer (Consumidor)

```go
// internal/orders/adapters/consumer.go
func (c *UserCreatedConsumer) handleMessage(ctx context.Context, body []byte) error {
    var event events.UserCreatedEvent
    json.Unmarshal(body, &event)
    
    // Hacer algo con el evento
    log.Info("Nuevo usuario registrado", zap.Uint("user_id", event.Payload.ID))
    
    return nil
}
```

---

## 8. Base de Datos con GORM

### ¿Qué es GORM?

**GORM** es un ORM (Object-Relational Mapping) para Go:
- Traduce estructuras Go a tablas SQL
- Ejecuta queries automáticamente
- Maneja migraciones

### Modelo de Persistencia

```go
// internal/users/adapters/repository.go

// Modelo para la base de datos (con tags de GORM)
type UserModel struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"size:100;not null"`
    Email     string    `gorm:"size:255;uniqueIndex;not null"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (UserModel) TableName() string {
    return "users"
}
```

### Separación Dominio vs Persistencia

```go
// Dominio: SIN tags de GORM (puro)
type User struct {
    ID        uint
    Name      string
    Email     string
    CreatedAt time.Time
}

// Conversiones
func toModel(user *domain.User) *UserModel {
    return &UserModel{
        ID:        user.ID,
        Name:      user.Name,
        Email:     user.Email,
        CreatedAt: user.CreatedAt,
    }
}

func toDomain(model *UserModel) *domain.User {
    return &domain.User{
        ID:        model.ID,
        Name:      model.Name,
        Email:     model.Email,
        CreatedAt: model.CreatedAt,
    }
}
```

### CRUD con GORM

```go
// CREATE
func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
    model := toModel(user)
    result := r.db.WithContext(ctx).Create(model)
    user.ID = model.ID  // Actualizar el ID generado
    return result.Error
}

// READ
func (r *PostgresUserRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
    var model UserModel
    result := r.db.WithContext(ctx).First(&model, id)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        return nil, domain.ErrUserNotFound
    }
    return toDomain(&model), nil
}
```

---

## 9. Manejo de Errores

### Tipos de Error

```go
// pkg/errors/errors.go
type AppError struct {
    Code    string      // "VALIDATION_ERROR", "NOT_FOUND", etc.
    Message string      // Mensaje legible
    Details interface{} // Información adicional
}

// Constructores
func NewValidation(message string, details interface{}) *AppError {
    return &AppError{Code: "VALIDATION_ERROR", Message: message, Details: details}
}

func NewNotFound(resource string, id interface{}) *AppError {
    return &AppError{
        Code:    "NOT_FOUND",
        Message: fmt.Sprintf("%s with id '%v' not found", resource, id),
    }
}
```

### Middleware de Errores

```go
// pkg/middleware/error.go
func ErrorHandler(log *logger.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()  // Procesar la petición
        
        // Si hubo errores
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            
            // Determinar código HTTP
            status := http.StatusInternalServerError
            if appErr, ok := err.(*errors.AppError); ok {
                switch appErr.Code {
                case "VALIDATION_ERROR":
                    status = http.StatusBadRequest
                case "NOT_FOUND":
                    status = http.StatusNotFound
                case "CONFLICT":
                    status = http.StatusConflict
                }
            }
            
            // Respuesta JSON consistente
            c.JSON(status, gin.H{
                "error": gin.H{
                    "code":    appErr.Code,
                    "message": appErr.Message,
                },
                "trace_id": c.GetString("trace_id"),
            })
        }
    }
}
```

### Uso en Handlers

```go
func (h *HTTPHandler) GetUser(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.Error(errors.NewValidation("invalid user id", nil))
        return  // El middleware maneja la respuesta
    }
    
    output, err := h.useCase.GetUser(ctx, application.GetUserInput{ID: uint(id)})
    if err != nil {
        c.Error(err)  // Puede ser NotFound, Internal, etc.
        return
    }
    
    c.JSON(http.StatusOK, output)
}
```

---

## 10. Middleware y HTTP

### ¿Qué es un Middleware?

Un **middleware** es código que se ejecuta ANTES o DESPUÉS de cada petición:

```
Request → [TraceID] → [Logger] → [ErrorHandler] → Handler → Response
            │           │           │
            │           │           └── Transforma errores en JSON
            │           └── Registra la petición en logs
            └── Genera ID único para rastreo
```

### TraceID Middleware

```go
// pkg/middleware/trace.go
func TraceID() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Generar ID único
        traceID := uuid.New().String()
        
        // Guardarlo en el contexto
        c.Set("trace_id", traceID)
        c.Request = c.Request.WithContext(
            context.WithValue(c.Request.Context(), "trace_id", traceID),
        )
        
        // Agregar al header de respuesta
        c.Header("X-Trace-ID", traceID)
        
        c.Next()
    }
}
```

### Logger Middleware

```go
// pkg/middleware/logger.go
func RequestLogger(log *logger.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()  // Procesar petición
        
        log.WithContext(c.Request.Context()).Info("request completed",
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
            zap.Int("status", c.Writer.Status()),
            zap.Duration("duration", time.Since(start)),
        )
    }
}
```

---

## 11. Configuración

### Variables de Entorno

```go
// pkg/config/config.go
type Config struct {
    // HTTP
    HTTPPort    string
    HTTPTimeout time.Duration
    
    // Database
    DBHost     string
    DBPort     string
    DBUser     string
    DBPassword string
    
    // RabbitMQ
    RabbitMQURL string
    
    // gRPC
    GRPCPort        string
    GRPCMTLSEnabled bool
}

func Load() *Config {
    godotenv.Load()  // Cargar .env
    
    return &Config{
        HTTPPort:    getEnvOrDefault("HTTP_PORT", "8080"),
        DBHost:      getEnvOrDefault("DB_HOST", "localhost"),
        RabbitMQURL: getEnvOrDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
        // ...
    }
}
```

### Archivo .env

```env
# .env
HTTP_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
LOG_LEVEL=debug
```

---

## 12. Flujo Completo de una Petición

### Ejemplo: Crear Orden

```
1. Cliente envía POST /api/v1/orders
   {"user_id": 1, "total": 99.99}
          │
          ▼
2. Gateway recibe la petición
   - Middleware TraceID genera: "abc-123"
   - Middleware Logger registra: "POST /api/v1/orders"
          │
          ▼
3. Gateway.CreateOrder() se ejecuta
   - Parsea JSON → CreateOrderRequest
   - Llama a Orders por gRPC
          │
          ▼
4. Orders.GRPCServer.CreateOrder() recibe
   - Delega a OrderUseCase.CreateOrder()
          │
          ▼
5. OrderUseCase valida el usuario
   - Llama a Users por gRPC: client.GetUser(1)
          │
          ▼
6. Users.GRPCServer.GetUser() responde
   - {id: 1, name: "Juan", email: "juan@..."}
          │
          ▼
7. OrderUseCase crea la orden
   - domain.NewOrder(userID=1, total=99.99)
   - repo.Create(order) → INSERT INTO orders...
   - publisher.PublishOrderCreated(order)
          │
          ▼
8. RabbitMQ recibe el evento
   - Otros servicios pueden consumirlo
          │
          ▼
9. Respuesta viaja de vuelta
   - Orders → Gateway → Cliente
   - {"data": {"id": 1, "status": "pending"}, "trace_id": "abc-123"}
```

---

## 🎯 Resumen de Conceptos Clave

| Concepto | Qué es | Dónde está |
|----------|--------|------------|
| **Microservicio** | Aplicación pequeña y enfocada | `cmd/users`, `cmd/orders` |
| **Gateway** | Punto de entrada único | `cmd/gateway` |
| **Clean Architecture** | Separación por capas | `internal/<service>/` |
| **Domain** | Entidades de negocio | `domain/entity.go` |
| **Ports** | Interfaces (contratos) | `ports/ports.go` |
| **Adapters** | Implementaciones | `adapters/*.go` |
| **UseCase** | Lógica de aplicación | `application/usecase.go` |
| **gRPC** | Comunicación entre servicios | `infrastructure/grpc.go` |
| **RabbitMQ** | Eventos asíncronos | `adapters/publisher.go` |
| **GORM** | ORM para base de datos | `adapters/repository.go` |
| **Middleware** | Pre/post procesamiento HTTP | `pkg/middleware/` |
| **Config** | Variables de entorno | `pkg/config/` |

---

## 📝 Próximos Pasos para Practicar

1. **Agregar un nuevo campo** al User (ej: `phone`)
2. **Crear un nuevo endpoint** (ej: ListUsers)
3. **Agregar un nuevo servicio** (ej: Products)
4. **Escribir más tests** unitarios
5. **Implementar autenticación** (JWT)
