# Go-Micro: Microservices Study Project

Monorepo de microservicios en Go con estándares profesionales para estudio.

## 🏗️ Arquitectura

```
┌─────────────────────┐
│   Cliente (HTTPS)   │
└─────────┬───────────┘
          │
┌─────────▼───────────┐
│  Gateway :8443      │  ← Swagger UI, TLS
│  (REST + Swagger)   │
└─────────┬───────────┘
          │ gRPC (mTLS)
    ┌─────┴─────┐
    │           │
┌───▼───┐   ┌───▼───┐
│ Users │   │Orders │
│:50051 │◄──│:50052 │  ← gRPC mTLS entre servicios
└───┬───┘   └───┬───┘
    │           │
┌───▼───┐   ┌───▼───┐      ┌──────────┐
│users  │   │orders │      │ RabbitMQ │
│_db    │   │_db    │      │ Events   │
└───────┘   └───────┘      └──────────┘
```

## 📁 Estructura del Proyecto

```
go-micro/
├── api/
│   ├── gen/           # Código gRPC generado
│   └── proto/         # Definiciones .proto
├── cmd/
│   ├── gateway/       # Entrypoint gateway
│   ├── users/         # Entrypoint users
│   └── orders/        # Entrypoint orders
├── internal/
│   ├── gateway/       # Handlers, clients
│   ├── users/         # Domain, application, adapters
│   └── orders/        # Domain, application, adapters
├── pkg/               # Paquetes compartidos
│   ├── config/        # Configuración
│   ├── db/            # Conexión GORM
│   ├── errors/        # Errores estándar
│   ├── events/        # Contratos de eventos
│   ├── grpc/          # Interceptores gRPC
│   ├── logger/        # Logger zap
│   ├── middleware/    # Middleware HTTP
│   ├── rabbitmq/      # Publisher/Consumer
│   └── tls/           # Utilidades TLS/mTLS
├── certs/             # Certificados (generados)
├── deploy/            # docker-compose
├── docs/swagger/      # Swagger generado
└── scripts/certs/     # Generación de certs
```

## 🚀 Quick Start

### Requisitos

- Go 1.21+
- Docker & Docker Compose
- Make
- OpenSSL (para certificados)

### 1. Clonar y configurar

```bash
cd go-micro
cp .env.example .env
```

### 2. Ejecutar con Docker

```bash
# Iniciar todos los servicios
make up

# Ver logs
make logs

# Detener
make down
```

### 3. Ejecutar localmente (desarrollo)

```bash
# Terminal 1: Iniciar infraestructura
docker-compose -f deploy/docker-compose.yml up users-db orders-db rabbitmq -d

# Terminal 2: Users service
make run-users

# Terminal 3: Orders service
make run-orders

# Terminal 4: Gateway
make run-gateway
```

### 4. Acceder a Swagger

- **HTTP**: http://localhost:8080/swagger/index.html
- **HTTPS**: https://localhost:8443/swagger/index.html (con TLS)

## 📋 API Endpoints

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| POST | `/api/v1/users` | Crear usuario |
| GET | `/api/v1/users/:id` | Obtener usuario |
| POST | `/api/v1/orders` | Crear orden |
| GET | `/api/v1/orders/:id` | Obtener orden |

### Ejemplo de flujo completo

```bash
# 1. Crear usuario
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com"}'

# Respuesta:
# {"data":{"id":1,"name":"John Doe","email":"john@example.com",...},"trace_id":"..."}

# 2. Crear orden (valida usuario por gRPC)
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "total": 99.99}'

# 3. Obtener orden
curl http://localhost:8080/api/v1/orders/1
```

## 🔐 TLS y mTLS

### ¿Qué es TLS y mTLS?

- **TLS (Transport Layer Security)**: Encripta la comunicación entre cliente y servidor. El cliente verifica el certificado del servidor.
- **mTLS (mutual TLS)**: Ambas partes se autentican mutuamente. El servidor también verifica el certificado del cliente.

### Generar certificados

```bash
make certs
```

Esto genera en `/certs`:
- `ca.crt/key` - Autoridad Certificadora local
- `gateway.crt/key` - Certificado del gateway (HTTPS)
- `gateway-client.crt/key` - Certificado cliente del gateway
- `users.crt/key` - Certificado servidor users
- `orders.crt/key` - Certificado servidor orders
- `orders-client.crt/key` - Certificado cliente orders→users

### Ejecutar con mTLS

```bash
# En .env
TLS_ENABLED=true
GRPC_MTLS_ENABLED=true

# Generar certs primero
make certs

# Iniciar servicios
make run-users   # En terminal 1
make run-orders  # En terminal 2
make run-gateway # En terminal 3

# Acceder con HTTPS
curl -k https://localhost:8443/api/v1/users
```

## 📨 Eventos RabbitMQ

### Flujo de eventos

1. **UserCreated**: Users → RabbitMQ → Orders (consume para demo)
2. **OrderCreated**: Orders → RabbitMQ

### Ver eventos

Acceder a RabbitMQ Management: http://localhost:15672
- Usuario: `guest`
- Password: `guest`

## 🧪 Testing

```bash
# Todos los tests
make test

# Solo tests unitarios
make test-unit

# Con cobertura
go test -v -cover ./...
```

## 🛠️ Comandos Make

| Comando | Descripción |
|---------|-------------|
| `make build` | Compilar todos los servicios |
| `make test` | Ejecutar tests |
| `make proto` | Generar código gRPC |
| `make swagger` | Generar documentación Swagger |
| `make certs` | Generar certificados TLS |
| `make up` | Iniciar con Docker Compose |
| `make down` | Detener Docker Compose |
| `make run-gateway` | Ejecutar gateway localmente |
| `make run-users` | Ejecutar users localmente |
| `make run-orders` | Ejecutar orders localmente |
| `make tools` | Instalar herramientas de desarrollo |

## 🏛️ Decisiones de Arquitectura

### Clean Architecture / Hexagonal

```
internal/<service>/
├── domain/          # Entidades, reglas de negocio
├── application/     # Casos de uso (usecases)
├── ports/           # Interfaces (repository, publisher)
├── adapters/        # Implementaciones (GORM, RabbitMQ)
└── infrastructure/  # Servidores (HTTP, gRPC)
```

**Flujo**: Handler → UseCase → Repository(interface) → GORM Implementation

### Manejo de errores

- **HTTP**: Middleware captura errores y panics, responde JSON consistente
- **gRPC**: Interceptor traduce errores de dominio a status codes
- **Formato uniforme**:
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "user with id '5' not found"
  },
  "trace_id": "550e8400-e29b..."
}
```

### Comunicación

- **Gateway ↔ Servicios**: gRPC (con mTLS opcional)
- **Servicios ↔ Servicios**: gRPC (orders→users para validar)
- **Eventos**: RabbitMQ con exchanges topic y ack manual

### Persistencia

- Base de datos separada por servicio
- Modelo de dominio sin tags GORM
- Modelo de persistencia con mapeo explícito

## 📚 Tecnologías

- **Framework HTTP**: Gin
- **gRPC**: google.golang.org/grpc
- **ORM**: GORM
- **Mensajería**: RabbitMQ (amqp091-go)
- **Logger**: Zap
- **Swagger**: swaggo/swag
- **Configuración**: godotenv

## 📄 Licencia

MIT
