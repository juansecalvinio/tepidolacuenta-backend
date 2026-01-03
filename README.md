# TePidoLaCuenta - Backend API

Backend API para el sistema de solicitud de cuenta en restaurantes con notificaciones en tiempo real.

## Tabla de Contenidos

- [Stack Tecnológico](#stack-tecnológico)
- [Características](#características)
- [Instalación](#instalación)
- [Configuración](#configuración)
- [Estructura del Proyecto](#estructura-del-proyecto)
- [API Reference](#api-reference)
  - [Authentication](#authentication)
  - [Restaurants](#restaurants)
  - [Tables](#tables)
  - [Requests](#requests)
  - [WebSocket](#websocket)
- [Ejemplos de Uso](#ejemplos-de-uso)
- [Docker](#docker)

---

## Stack Tecnológico

- **Language:** Go 1.25.3
- **Framework:** Gin (HTTP Router)
- **Database:** MongoDB Atlas
- **Authentication:** JWT (JSON Web Tokens)
- **Real-time:** WebSockets (Gorilla WebSocket)
- **Password Hashing:** bcrypt
- **Architecture:** Clean Architecture with modular structure

## Características

- ✅ Autenticación JWT
- ✅ CRUD completo de Restaurantes
- ✅ CRUD completo de Mesas con generación automática de QR
- ✅ Sistema de solicitudes de cuenta con validación de QR
- ✅ WebSocket para notificaciones en tiempo real
- ✅ Validación de ownership (usuarios solo acceden a sus recursos)
- ✅ CORS configurado
- ✅ Clean Architecture (domain, repository, usecase, handler)

---

## Instalación

### Requisitos Previos

- Go 1.21 o superior
- MongoDB Atlas account (o MongoDB local)
- Git

### Pasos

1. **Clonar el repositorio**
```bash
git clone https://github.com/yourusername/tepidolacuenta-backend.git
cd tepidolacuenta-backend
```

2. **Instalar dependencias**
```bash
go mod download
```

3. **Configurar variables de entorno**
```bash
cp .env.example .env
```

Editar `.env` con tus credenciales:
```env
MONGODB_URI=mongodb+srv://username:password@cluster.mongodb.net/tepidolacuenta?retryWrites=true&w=majority
JWT_SECRET=your_super_secret_jwt_key_change_this_in_production
PORT=8080
GIN_MODE=debug
CORS_ALLOWED_ORIGINS=http://localhost:5173
FRONTEND_BASE_URL=http://localhost:5173
```

4. **Ejecutar el servidor**
```bash
go run cmd/api/main.go
```

El servidor estará disponible en `http://localhost:8080`

5. **Verificar el health check**
```bash
curl http://localhost:8080/health
```

---

## Configuración

### Variables de Entorno

| Variable | Descripción | Ejemplo | Requerido |
|----------|-------------|---------|-----------|
| `MONGODB_URI` | URI de conexión a MongoDB | `mongodb+srv://...` | Sí |
| `JWT_SECRET` | Clave secreta para firmar JWT tokens | `your_secret_key` | Sí |
| `PORT` | Puerto del servidor | `8080` | No (default: 8080) |
| `GIN_MODE` | Modo de Gin (debug/release) | `debug` | No (default: debug) |
| `CORS_ALLOWED_ORIGINS` | Orígenes permitidos para CORS | `http://localhost:5173` | No |
| `FRONTEND_BASE_URL` | URL base del frontend para QR | `http://localhost:5173` | Sí |

---

## Estructura del Proyecto

```
tepidolacuenta-backend/
├── cmd/
│   └── api/
│       └── main.go                      # Entry point
├── config/
│   └── config.go                        # Configuration management
├── internal/
│   ├── auth/                           # Auth module
│   │   ├── domain/
│   │   │   └── user.go                 # User entity & DTOs
│   │   ├── repository/
│   │   │   ├── repository.go           # Repository interface
│   │   │   └── mongodb.go              # MongoDB implementation
│   │   ├── usecase/
│   │   │   └── auth_usecase.go         # Business logic
│   │   └── handler/
│   │       └── auth_handler.go         # HTTP handlers
│   ├── restaurant/                     # Restaurant module
│   │   ├── domain/
│   │   │   └── restaurant.go
│   │   ├── repository/
│   │   │   ├── repository.go
│   │   │   └── mongodb.go
│   │   ├── usecase/
│   │   │   └── restaurant_usecase.go
│   │   └── handler/
│   │       └── restaurant_handler.go
│   ├── table/                          # Table module
│   │   ├── domain/
│   │   │   └── table.go
│   │   ├── repository/
│   │   │   ├── repository.go
│   │   │   └── mongodb.go
│   │   ├── usecase/
│   │   │   └── table_usecase.go
│   │   └── handler/
│   │       └── table_handler.go
│   ├── request/                        # Request module
│   │   ├── domain/
│   │   │   └── request.go
│   │   ├── repository/
│   │   │   ├── repository.go
│   │   │   └── mongodb.go
│   │   ├── usecase/
│   │   │   └── request_usecase.go
│   │   └── handler/
│   │       └── request_handler.go
│   ├── database/
│   │   └── mongodb.go                  # MongoDB connection
│   ├── middleware/
│   │   └── auth.go                     # JWT middleware
│   └── pkg/
│       ├── errors.go                   # Common errors
│       ├── jwt.go                      # JWT service
│       ├── qr.go                       # QR code generation
│       ├── response.go                 # Response helpers
│       ├── validation.go               # Input validation
│       └── websocket.go                # WebSocket hub
├── .env.example                        # Environment template
├── go.mod                              # Dependencies
├── go.sum                              # Dependency checksums
└── README.md                           # This file
```

---

## API Reference

Base URL: `http://localhost:8080`

### Authentication

#### Register User

**POST** `/api/v1/auth/register`

Registra un nuevo usuario/restaurante en el sistema.

**Request Body:**
```json
{
  "email": "restaurant@example.com",
  "password": "securePassword123",
  "restaurantName": "La Pizzeria"
}
```

**Response:** `201 Created`
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "id": "64a7f8b3c12345678901234",
    "email": "restaurant@example.com",
    "restaurantName": "La Pizzeria",
    "createdAt": "2026-01-02T12:00:00Z",
    "updatedAt": "2026-01-02T12:00:00Z"
  }
}
```

**Errors:**
- `400 Bad Request` - Invalid input or email already exists

---

#### Login

**POST** `/api/v1/auth/login`

Autenticar usuario y obtener token JWT.

**Request Body:**
```json
{
  "email": "restaurant@example.com",
  "password": "securePassword123"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "64a7f8b3c12345678901234",
      "email": "restaurant@example.com",
      "restaurantName": "La Pizzeria",
      "createdAt": "2026-01-02T12:00:00Z",
      "updatedAt": "2026-01-02T12:00:00Z"
    }
  }
}
```

**Errors:**
- `401 Unauthorized` - Invalid credentials

---

### Restaurants

Todos los endpoints de restaurantes requieren autenticación (header `Authorization: Bearer {token}`).

#### Create Restaurant

**POST** `/api/v1/restaurants`

Crea un nuevo restaurante.

**Headers:**
```
Authorization: Bearer {token}
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "La Pizzeria",
  "address": "Av. Corrientes 1234, CABA",
  "phone": "+54 11 1234-5678",
  "description": "Pizzeria con horno a leña"
}
```

**Response:** `201 Created`
```json
{
  "success": true,
  "message": "Restaurant created successfully",
  "data": {
    "id": "64a7f9abc12345678901234",
    "userId": "64a7f8b3c12345678901234",
    "name": "La Pizzeria",
    "address": "Av. Corrientes 1234, CABA",
    "phone": "+54 11 1234-5678",
    "description": "Pizzeria con horno a leña",
    "createdAt": "2026-01-02T12:05:00Z",
    "updatedAt": "2026-01-02T12:05:00Z"
  }
}
```

---

#### List Restaurants

**GET** `/api/v1/restaurants`

Lista todos los restaurantes del usuario autenticado.

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Restaurants retrieved successfully",
  "data": [
    {
      "id": "64a7f9abc12345678901234",
      "userId": "64a7f8b3c12345678901234",
      "name": "La Pizzeria",
      "address": "Av. Corrientes 1234, CABA",
      "phone": "+54 11 1234-5678",
      "description": "Pizzeria con horno a leña",
      "createdAt": "2026-01-02T12:05:00Z",
      "updatedAt": "2026-01-02T12:05:00Z"
    }
  ]
}
```

---

#### Get Restaurant by ID

**GET** `/api/v1/restaurants/{id}`

Obtiene un restaurante específico por su ID.

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Restaurant retrieved successfully",
  "data": {
    "id": "64a7f9abc12345678901234",
    "userId": "64a7f8b3c12345678901234",
    "name": "La Pizzeria",
    "address": "Av. Corrientes 1234, CABA",
    "phone": "+54 11 1234-5678",
    "description": "Pizzeria con horno a leña",
    "createdAt": "2026-01-02T12:05:00Z",
    "updatedAt": "2026-01-02T12:05:00Z"
  }
}
```

**Errors:**
- `404 Not Found` - Restaurant not found
- `401 Unauthorized` - User doesn't own this restaurant

---

#### Update Restaurant

**PUT** `/api/v1/restaurants/{id}`

Actualiza un restaurante existente.

**Headers:**
```
Authorization: Bearer {token}
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "La Gran Pizzeria",
  "address": "Av. Corrientes 1234, CABA",
  "phone": "+54 11 1234-5678",
  "description": "La mejor pizzeria con horno a leña"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Restaurant updated successfully",
  "data": {
    "id": "64a7f9abc12345678901234",
    "userId": "64a7f8b3c12345678901234",
    "name": "La Gran Pizzeria",
    "address": "Av. Corrientes 1234, CABA",
    "phone": "+54 11 1234-5678",
    "description": "La mejor pizzeria con horno a leña",
    "createdAt": "2026-01-02T12:05:00Z",
    "updatedAt": "2026-01-02T12:10:00Z"
  }
}
```

---

#### Delete Restaurant

**DELETE** `/api/v1/restaurants/{id}`

Elimina un restaurante.

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Restaurant deleted successfully"
}
```

---

### Tables

Todos los endpoints de mesas requieren autenticación.

#### Create Table

**POST** `/api/v1/tables`

Crea una nueva mesa con QR code generado automáticamente.

**Headers:**
```
Authorization: Bearer {token}
Content-Type: application/json
```

**Request Body:**
```json
{
  "restaurantId": "64a7f9abc12345678901234",
  "number": 5,
  "capacity": 4
}
```

**Response:** `201 Created`
```json
{
  "success": true,
  "message": "Table created successfully",
  "data": {
    "id": "64a7fabc12345678901234",
    "restaurantId": "64a7f9abc12345678901234",
    "number": 5,
    "capacity": 4,
    "qrCode": "http://localhost:5173/request?r=64a7f9abc12345678901234&t=64a7fabc12345678901234&n=5&h=GUyQvt7LbzYbdaX9",
    "isActive": true,
    "createdAt": "2026-01-02T12:15:00Z",
    "updatedAt": "2026-01-02T12:15:00Z"
  }
}
```

**Errors:**
- `400 Bad Request` - Table number already exists for this restaurant
- `401 Unauthorized` - User doesn't own the restaurant

---

#### Get Table by ID

**GET** `/api/v1/tables/{id}`

Obtiene una mesa específica por su ID.

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Table retrieved successfully",
  "data": {
    "id": "64a7fabc12345678901234",
    "restaurantId": "64a7f9abc12345678901234",
    "number": 5,
    "capacity": 4,
    "qrCode": "http://localhost:5173/request?r=64a7f9abc12345678901234&t=64a7fabc12345678901234&n=5&h=GUyQvt7LbzYbdaX9",
    "isActive": true,
    "createdAt": "2026-01-02T12:15:00Z",
    "updatedAt": "2026-01-02T12:15:00Z"
  }
}
```

---

#### List Tables by Restaurant

**GET** `/api/v1/tables/restaurant/{restaurantId}`

Lista todas las mesas de un restaurante.

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Tables retrieved successfully",
  "data": [
    {
      "id": "64a7fabc12345678901234",
      "restaurantId": "64a7f9abc12345678901234",
      "number": 5,
      "capacity": 4,
      "qrCode": "http://localhost:5173/request?r=64a7f9abc12345678901234&t=64a7fabc12345678901234&n=5&h=GUyQvt7LbzYbdaX9",
      "isActive": true,
      "createdAt": "2026-01-02T12:15:00Z",
      "updatedAt": "2026-01-02T12:15:00Z"
    }
  ]
}
```

---

#### Update Table

**PUT** `/api/v1/tables/{id}`

Actualiza una mesa. Si se cambia el número, se regenera automáticamente el QR code.

**Headers:**
```
Authorization: Bearer {token}
Content-Type: application/json
```

**Request Body:**
```json
{
  "number": 6,
  "capacity": 6,
  "isActive": true
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Table updated successfully",
  "data": {
    "id": "64a7fabc12345678901234",
    "restaurantId": "64a7f9abc12345678901234",
    "number": 6,
    "capacity": 6,
    "qrCode": "http://localhost:5173/request?r=64a7f9abc12345678901234&t=64a7fabc12345678901234&n=6&h=XyZ123AbcDef4567",
    "isActive": true,
    "createdAt": "2026-01-02T12:15:00Z",
    "updatedAt": "2026-01-02T12:20:00Z"
  }
}
```

---

#### Delete Table

**DELETE** `/api/v1/tables/{id}`

Elimina una mesa.

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Table deleted successfully"
}
```

---

### Requests

El sistema de solicitudes permite a los clientes pedir la cuenta escaneando el QR de la mesa.

#### Create Request (Public)

**POST** `/api/v1/public/request-account`

**⚠️ Endpoint público** - No requiere autenticación. Usado por clientes al escanear QR.

**Request Body:**
```json
{
  "restaurantId": "64a7f9abc12345678901234",
  "tableId": "64a7fabc12345678901234",
  "tableNumber": 5,
  "hash": "GUyQvt7LbzYbdaX9"
}
```

**Response:** `201 Created`
```json
{
  "success": true,
  "message": "Request created successfully",
  "data": {
    "id": "64a7fbcd12345678901234",
    "restaurantId": "64a7f9abc12345678901234",
    "tableId": "64a7fabc12345678901234",
    "tableNumber": 5,
    "status": "pending",
    "createdAt": "2026-01-02T12:25:00Z",
    "updatedAt": "2026-01-02T12:25:00Z"
  }
}
```

**Errors:**
- `400 Bad Request` - Invalid QR code, table not active, or invalid data
- `404 Not Found` - Restaurant or table not found

**Nota:** Al crear un request, se envía automáticamente una notificación WebSocket al restaurante.

---

#### Get Request by ID

**GET** `/api/v1/requests/{id}`

Obtiene una solicitud específica.

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Request retrieved successfully",
  "data": {
    "id": "64a7fbcd12345678901234",
    "restaurantId": "64a7f9abc12345678901234",
    "tableId": "64a7fabc12345678901234",
    "tableNumber": 5,
    "status": "pending",
    "createdAt": "2026-01-02T12:25:00Z",
    "updatedAt": "2026-01-02T12:25:00Z"
  }
}
```

---

#### List All Requests by Restaurant

**GET** `/api/v1/requests/restaurant/{restaurantId}`

Lista todas las solicitudes de un restaurante (todas los estados).

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Requests retrieved successfully",
  "data": [
    {
      "id": "64a7fbcd12345678901234",
      "restaurantId": "64a7f9abc12345678901234",
      "tableId": "64a7fabc12345678901234",
      "tableNumber": 5,
      "status": "pending",
      "createdAt": "2026-01-02T12:25:00Z",
      "updatedAt": "2026-01-02T12:25:00Z"
    },
    {
      "id": "64a7fcde12345678901234",
      "restaurantId": "64a7f9abc12345678901234",
      "tableId": "64a7fabc12345678901234",
      "tableNumber": 5,
      "status": "processed",
      "createdAt": "2026-01-02T12:20:00Z",
      "updatedAt": "2026-01-02T12:22:00Z"
    }
  ]
}
```

---

#### List Pending Requests

**GET** `/api/v1/requests/restaurant/{restaurantId}/pending`

Lista solo las solicitudes pendientes de un restaurante.

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Pending requests retrieved successfully",
  "data": [
    {
      "id": "64a7fbcd12345678901234",
      "restaurantId": "64a7f9abc12345678901234",
      "tableId": "64a7fabc12345678901234",
      "tableNumber": 5,
      "status": "pending",
      "createdAt": "2026-01-02T12:25:00Z",
      "updatedAt": "2026-01-02T12:25:00Z"
    }
  ]
}
```

---

#### Update Request Status

**PUT** `/api/v1/requests/{id}/status`

Actualiza el estado de una solicitud.

**Headers:**
```
Authorization: Bearer {token}
Content-Type: application/json
```

**Request Body:**
```json
{
  "status": "processed"
}
```

**Valores permitidos para `status`:**
- `pending` - Pendiente
- `processed` - Procesada
- `cancelled` - Cancelada

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Request status updated successfully",
  "data": {
    "id": "64a7fbcd12345678901234",
    "restaurantId": "64a7f9abc12345678901234",
    "tableId": "64a7fabc12345678901234",
    "tableNumber": 5,
    "status": "processed",
    "createdAt": "2026-01-02T12:25:00Z",
    "updatedAt": "2026-01-02T12:30:00Z"
  }
}
```

---

#### Delete Request

**DELETE** `/api/v1/requests/{id}`

Elimina una solicitud.

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Request deleted successfully"
}
```

---

### WebSocket

#### Connect to WebSocket

**WS** `/api/v1/requests/ws/{restaurantId}`

Establece una conexión WebSocket para recibir notificaciones en tiempo real de nuevas solicitudes.

**Headers:**
```
Authorization: Bearer {token}
Upgrade: websocket
Connection: Upgrade
```

**Ejemplo de conexión (JavaScript):**
```javascript
const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...";
const restaurantId = "64a7f9abc12345678901234";

const ws = new WebSocket(
  `ws://localhost:8080/api/v1/requests/ws/${restaurantId}?token=${token}`
);

ws.onopen = () => {
  console.log("Connected to WebSocket");
};

ws.onmessage = (event) => {
  const request = JSON.parse(event.data);
  console.log("New request received:", request);
  // request = {
  //   "id": "64a7fbcd12345678901234",
  //   "restaurantId": "64a7f9abc12345678901234",
  //   "tableId": "64a7fabc12345678901234",
  //   "tableNumber": 5,
  //   "status": "pending",
  //   "createdAt": "2026-01-02T12:25:00Z",
  //   "updatedAt": "2026-01-02T12:25:00Z"
  // }
};

ws.onerror = (error) => {
  console.error("WebSocket error:", error);
};

ws.onclose = () => {
  console.log("WebSocket connection closed");
};
```

**Notas:**
- El WebSocket envía mensajes JSON cuando se crea una nueva solicitud
- La conexión es específica por restaurante
- Solo los usuarios autenticados y dueños del restaurante pueden conectarse
- El hub de WebSocket mantiene las conexiones activas y limpia automáticamente las desconectadas

---

## Ejemplos de Uso

### Flujo completo: Registro → Restaurante → Mesa → Solicitud

```bash
# 1. Registrar usuario
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "restaurant@example.com",
    "password": "password123",
    "restaurantName": "La Pizzeria"
  }'

# 2. Login y obtener token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "restaurant@example.com",
    "password": "password123"
  }' | jq -r '.data.token')

# 3. Crear restaurante
RESTAURANT=$(curl -s -X POST http://localhost:8080/api/v1/restaurants \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "La Pizzeria",
    "address": "Av. Corrientes 1234",
    "phone": "+54 11 1234-5678",
    "description": "Pizzeria con horno a leña"
  }')

RESTAURANT_ID=$(echo $RESTAURANT | jq -r '.data.id')

# 4. Crear mesa
TABLE=$(curl -s -X POST http://localhost:8080/api/v1/tables \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{
    \"restaurantId\": \"$RESTAURANT_ID\",
    \"number\": 5,
    \"capacity\": 4
  }")

echo $TABLE | jq '.data.qrCode'
# Output: "http://localhost:5173/request?r=...&t=...&n=5&h=..."

# 5. Simular cliente escaneando QR y solicitando cuenta
TABLE_ID=$(echo $TABLE | jq -r '.data.tableId')
QR_HASH=$(echo $TABLE | jq -r '.data.qrCode' | grep -o 'h=.*' | cut -d'=' -f2)

curl -X POST http://localhost:8080/api/v1/public/request-account \
  -H 'Content-Type: application/json' \
  -d "{
    \"restaurantId\": \"$RESTAURANT_ID\",
    \"tableId\": \"$TABLE_ID\",
    \"tableNumber\": 5,
    \"hash\": \"$QR_HASH\"
  }"

# 6. Obtener solicitudes pendientes
curl -X GET "http://localhost:8080/api/v1/requests/restaurant/$RESTAURANT_ID/pending" \
  -H "Authorization: Bearer $TOKEN"
```

---

## Docker

### Construir la imagen

```bash
docker build -t tepidolacuenta-backend:latest .
```

### Ejecutar el contenedor

```bash
docker run -p 8080:8080 --env-file .env tepidolacuenta-backend:latest
```

### Docker Compose (ejemplo)

```yaml
version: '3.8'

services:
  backend:
    build: .
    ports:
      - "8080:8080"
    environment:
      - MONGODB_URI=${MONGODB_URI}
      - JWT_SECRET=${JWT_SECRET}
      - PORT=8080
      - GIN_MODE=release
      - CORS_ALLOWED_ORIGINS=https://tepidolacuenta.com
      - FRONTEND_BASE_URL=https://tepidolacuenta.com
    restart: unless-stopped
```

---

## Desarrollo

### Ejecutar con auto-reload

```bash
# Instalar air
go install github.com/cosmtrek/air@latest

# Ejecutar con auto-reload
air
```

### Ejecutar tests

```bash
go test ./...
```

### Build para producción

```bash
go build -o bin/server cmd/api/main.go
./bin/server
```

---

## Seguridad

- ✅ Contraseñas hasheadas con bcrypt (cost factor: 10)
- ✅ Tokens JWT con expiración de 24 horas
- ✅ Validación de QR codes con hash SHA256
- ✅ Middleware de autenticación en rutas protegidas
- ✅ Validación de ownership en todos los endpoints
- ✅ CORS configurado
- ✅ Input validation en todos los endpoints

---

## Códigos de Error HTTP

| Código | Descripción |
|--------|-------------|
| `200` | OK - Solicitud exitosa |
| `201` | Created - Recurso creado exitosamente |
| `400` | Bad Request - Datos de entrada inválidos |
| `401` | Unauthorized - No autenticado o token inválido |
| `404` | Not Found - Recurso no encontrado |
| `500` | Internal Server Error - Error del servidor |

---

## License

MIT License - ver archivo LICENSE para más detalles.

---

## Contacto

Para preguntas o soporte, contactar a: [tu-email@example.com]
