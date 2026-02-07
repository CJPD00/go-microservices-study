# Guía de Ejecución Local (sin Docker)

Esta guía explica cómo ejecutar el proyecto **go-micro** localmente sin usar Docker.

## Requisitos

- Go 1.23+
- PostgreSQL 14+
- RabbitMQ 3.x

## 1. Instalar PostgreSQL

### Windows (Chocolatey)
```powershell
choco install postgresql
```

### Windows (Instalador)
Descarga desde: https://www.postgresql.org/download/windows/

### Verificar instalación
```powershell
psql --version
```

## 2. Instalar RabbitMQ

### Windows (Chocolatey)
```powershell
choco install rabbitmq
```

### Windows (Instalador)
1. Instalar Erlang: https://www.erlang.org/downloads
2. Instalar RabbitMQ: https://www.rabbitmq.com/install-windows.html

### Iniciar servicio
```powershell
# Como servicio de Windows (automático)
rabbitmq-service start

# O manualmente
rabbitmq-server
```

### Verificar
- Accede a: http://localhost:15672
- Usuario: `guest` / Password: `guest`

## 3. Crear Bases de Datos

Conecta a PostgreSQL y ejecuta:

```sql
CREATE DATABASE users_db;
CREATE DATABASE orders_db;
```

Con línea de comandos:
```powershell
psql -U postgres -c "CREATE DATABASE users_db;"
psql -U postgres -c "CREATE DATABASE orders_db;"
```

## 4. Configurar Variables de Entorno

Copia el archivo de ejemplo:
```powershell
copy .env.example .env
```

Edita `.env` con tus credenciales:
```env
# Users Database
USERS_DB_HOST=localhost
USERS_DB_PORT=5432
USERS_DB_USER=postgres
USERS_DB_PASSWORD=tu_password_aqui
USERS_DB_NAME=users_db

# Orders Database
ORDERS_DB_HOST=localhost
ORDERS_DB_PORT=5432
ORDERS_DB_USER=postgres
ORDERS_DB_PASSWORD=tu_password_aqui
ORDERS_DB_NAME=orders_db

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# gRPC
USERS_GRPC_ADDR=localhost:50051
ORDERS_GRPC_ADDR=localhost:50052

# HTTP Ports
HTTP_PORT=8080
USERS_HTTP_PORT=8081
ORDERS_HTTP_PORT=8082

# Deshabilitar TLS para desarrollo local
TLS_ENABLED=false
GRPC_MTLS_ENABLED=false

# Logging
LOG_LEVEL=debug
```

## 5. Ejecutar los Servicios

Abre **3 terminales** en la carpeta del proyecto:

### Terminal 1: Users Service
```powershell
cd d:\github\go-micro
go run ./cmd/users
```

Salida esperada:
```
{"level":"info","service":"users-service","message":"starting users service"}
{"level":"info","service":"users-service","message":"connected to database"}
{"level":"info","service":"users-service","message":"HTTP server listening on :8081"}
{"level":"info","service":"users-service","message":"gRPC server listening on :50051"}
```

### Terminal 2: Orders Service
```powershell
cd d:\github\go-micro
go run ./cmd/orders
```

Salida esperada:
```
{"level":"info","service":"orders-service","message":"starting orders service"}
{"level":"info","service":"orders-service","message":"connected to database"}
{"level":"info","service":"orders-service","message":"connected to users service"}
{"level":"info","service":"orders-service","message":"HTTP server listening on :8082"}
{"level":"info","service":"orders-service","message":"gRPC server listening on :50052"}
```

### Terminal 3: Gateway
```powershell
cd d:\github\go-micro
go run ./cmd/gateway
```

Salida esperada:
```
{"level":"info","service":"gateway","message":"starting gateway service"}
{"level":"info","service":"gateway","message":"connected to backend services via gRPC"}
{"level":"info","service":"gateway","message":"HTTP server listening on http://localhost:8080"}
{"level":"info","service":"gateway","message":"Swagger UI: http://localhost:8080/swagger/index.html"}
```

## 6. Verificar que Todo Funciona

### Health Checks
```powershell
curl http://localhost:8080/health   # Gateway
curl http://localhost:8081/health   # Users
curl http://localhost:8082/health   # Orders
```

### Swagger UI
Abre en el navegador: http://localhost:8080/swagger/index.html

### RabbitMQ Management
Abre en el navegador: http://localhost:15672 (guest/guest)

## 7. Probar el Flujo Completo

### Crear un usuario
```powershell
curl -X POST http://localhost:8080/api/v1/users `
  -H "Content-Type: application/json" `
  -d '{"name": "Juan Garcia", "email": "juan@example.com"}'
```

Respuesta:
```json
{
  "data": {
    "id": 1,
    "name": "Juan Garcia",
    "email": "juan@example.com",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "trace_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Crear una orden
```powershell
curl -X POST http://localhost:8080/api/v1/orders `
  -H "Content-Type: application/json" `
  -d '{"user_id": 1, "total": 150.50}'
```

### Obtener datos
```powershell
curl http://localhost:8080/api/v1/users/1
curl http://localhost:8080/api/v1/orders/1
```

## 8. Ejecutar Tests

```powershell
go test -v ./internal/users/application/...
go test -v ./internal/orders/application/...

# O todos los tests
go test -v ./...
```

## Troubleshooting

### Error: "connection refused" en PostgreSQL
- Verifica que PostgreSQL esté corriendo
- Revisa las credenciales en `.env`

### Error: "connection refused" en RabbitMQ
- Verifica que el servicio esté iniciado
- Ejecuta: `rabbitmq-service start`

### Error: "address already in use"
- Otro proceso está usando el puerto
- Cambia el puerto en `.env` o cierra el proceso conflictivo

### Error: gRPC connection failed
- Asegúrate de iniciar Users antes que Orders
- Verifica que `USERS_GRPC_ADDR` esté correcto

## Orden de Inicio Recomendado

1. PostgreSQL (servicio de Windows)
2. RabbitMQ (servicio de Windows)
3. Users Service (terminal 1)
4. Orders Service (terminal 2)
5. Gateway (terminal 3)

## Puertos Utilizados

| Servicio | HTTP | gRPC |
|----------|------|------|
| Gateway | 8080 | - |
| Users | 8081 | 50051 |
| Orders | 8082 | 50052 |
| PostgreSQL | - | 5432 |
| RabbitMQ | 15672 (mgmt) | 5672 |
