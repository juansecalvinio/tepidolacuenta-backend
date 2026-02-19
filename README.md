# TePidoLaCuenta - Backend API

Backend API para el sistema de solicitud de cuenta en restaurantes con notificaciones en tiempo real.

## Tabla de Contenidos

- [Stack Tecnologico](#stack-tecnologico)
- [Caracteristicas](#caracteristicas)
- [Arquitectura](#arquitectura)
- [Instalacion](#instalacion)
- [Configuracion](#configuracion)
- [Estructura del Proyecto](#estructura-del-proyecto)
- [API Reference](#api-reference)
  - [Authentication](#authentication)
  - [Restaurants](#restaurants)
  - [Branches](#branches)
  - [Tables](#tables)
  - [Requests](#requests)
  - [WebSocket](#websocket)
- [Ejemplos de Uso](#ejemplos-de-uso)
- [Docker](#docker)
- [Desarrollo](#desarrollo)
- [Seguridad](#seguridad)

---

## Stack Tecnologico

- **Language:** Go 1.25.3
- **Framework:** Gin (HTTP Router)
- **Database:** MongoDB Atlas
- **Authentication:** JWT (JSON Web Tokens) con `golang-jwt/jwt/v5`
- **Real-time:** WebSockets (Gorilla WebSocket)
- **Password Hashing:** bcrypt
- **Architecture:** Clean Architecture (domain, repository, usecase, handler)

## Caracteristicas

- Autenticacion JWT con expiracion de 24 horas
- CRUD completo de Restaurantes
- CRUD completo de Sucursales (Branches) por restaurante
- CRUD completo de Mesas con generacion automatica de QR por sucursal
- Creacion masiva de mesas (bulk create)
- Sistema de solicitudes de cuenta con validacion de QR (SHA256)
- WebSocket para notificaciones en tiempo real
- Validacion de ownership (usuarios solo acceden a sus recursos)
- CORS configurable con multiples origenes
- Clean Architecture modular con interfaces

---

## Arquitectura

El sistema sigue una jerarquia de recursos:

```
Usuario (User)
  └── Restaurante (Restaurant)
        └── Sucursal (Branch)
              └── Mesa (Table) → genera QR code
                    └── Solicitud (Request) → notifica via WebSocket
```

- Un **Usuario** puede tener multiples **Restaurantes**
- Un **Restaurante** puede tener multiples **Sucursales**
- Una **Sucursal** puede tener multiples **Mesas**
- Cada **Mesa** tiene un QR code unico que incluye restaurantId, branchId, tableId y tableNumber
- Los **Clientes** escanean el QR y crean una **Solicitud** publica (sin autenticacion)
- El **Restaurante** recibe la solicitud en tiempo real via **WebSocket**

---

## Instalacion

### Requisitos Previos

- Go 1.21 o superior
- MongoDB Atlas account (o MongoDB local)
- Git

### Pasos

1. **Clonar el repositorio**
```bash
git clone https://github.com/juansecalvinio/tepidolacuenta-backend.git
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

El servidor estara disponible en `http://localhost:8080`

5. **Verificar el health check**
```bash
curl http://localhost:8080/health
```

---

## Configuracion

### Variables de Entorno

| Variable | Descripcion | Ejemplo | Requerido |
|----------|-------------|---------|-----------|
| `MONGODB_URI` | URI de conexion a MongoDB | `mongodb+srv://...` | Si |
| `JWT_SECRET` | Clave secreta para firmar JWT tokens | `your_secret_key` | Si |
| `PORT` | Puerto del servidor | `8080` | No (default: 8080) |
| `GIN_MODE` | Modo de Gin (debug/release) | `debug` | No (default: debug) |
| `CORS_ALLOWED_ORIGINS` | Origenes permitidos para CORS (separados por coma) | `http://localhost:5173,https://app.com` | No (default: http://localhost:5173) |
| `FRONTEND_BASE_URL` | URL base del frontend para generar QR codes | `http://localhost:5173` | Si |

---

## Estructura del Proyecto

```
tepidolacuenta-backend/
├── cmd/
│   └── api/
│       └── main.go                        # Entry point
├── config/
│   └── config.go                          # Configuration management
├── internal/
│   ├── auth/                              # Auth module
│   │   ├── domain/
│   │   │   └── user.go                    # User entity & DTOs
│   │   ├── repository/
│   │   │   ├── repository.go              # Repository interface
│   │   │   └── mongodb.go                 # MongoDB implementation
│   │   ├── usecase/
│   │   │   └── auth_usecase.go            # Business logic
│   │   └── handler/
│   │       └── auth_handler.go            # HTTP handlers
│   ├── restaurant/                        # Restaurant module
│   │   ├── domain/
│   │   │   └── restaurant.go
│   │   ├── repository/
│   │   │   ├── repository.go
│   │   │   └── mongodb.go
│   │   ├── usecase/
│   │   │   └── restaurant_usecase.go
│   │   └── handler/
│   │       └── restaurant_handler.go
│   ├── branch/                            # Branch module (sucursales)
│   │   ├── domain/
│   │   │   └── branch.go
│   │   ├── repository/
│   │   │   ├── repository.go
│   │   │   └── mongodb.go
│   │   ├── usecase/
│   │   │   └── branch_usecase.go
│   │   └── handler/
│   │       └── branch_handler.go
│   ├── table/                             # Table module
│   │   ├── domain/
│   │   │   └── table.go
│   │   ├── repository/
│   │   │   ├── repository.go
│   │   │   └── mongodb.go
│   │   ├── usecase/
│   │   │   └── table_usecase.go
│   │   └── handler/
│   │       └── table_handler.go
│   ├── request/                           # Request module
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
│   │   └── mongodb.go                     # MongoDB connection
│   ├── middleware/
│   │   └── auth.go                        # JWT middleware
│   └── pkg/
│       ├── errors.go                      # Common errors
│       ├── jwt.go                         # JWT service
│       ├── qr.go                          # QR code generation & validation
│       ├── response.go                    # Response helpers
│       ├── validator.go                   # Input validation
│       └── websocket.go                   # WebSocket hub
├── .env.example                           # Environment template
├── go.mod                                 # Dependencies
├── go.sum                                 # Dependency checksums
└── README.md                              # This file
```

---

## API Reference

Base URL: `http://localhost:8080`

### Formato de Respuesta

Todas las respuestas siguen el mismo formato:

**Respuesta exitosa:**
```json
{
  "success": true,
  "message": "Descripcion del resultado",
  "data": { ... }
}
```

**Respuesta de error:**
```json
{
  "success": false,
  "message": "Descripcion del error",
  "error": "detalle tecnico del error"
}
```

---

### Health Check

**GET** `/health`

Verifica que el servidor este funcionando.

**Response:** `200 OK`
```json
{
  "status": "ok",
  "time": "2026-01-02T12:00:00Z"
}
```

---

### Authentication

#### Register User

**POST** `/api/v1/auth/register`

Registra un nuevo usuario en el sistema.

**Request Body:**
```json
{
  "email": "restaurant@example.com",
  "password": "securePassword123"
}
```

| Campo | Tipo | Requerido | Validacion |
|-------|------|-----------|------------|
| `email` | string | Si | Formato email valido |
| `password` | string | Si | Minimo 8 caracteres |

**Response:** `201 Created`
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "id": "64a7f8b3c12345678901234",
    "email": "restaurant@example.com",
    "createdAt": "2026-01-02T12:00:00Z",
    "updatedAt": "2026-01-02T12:00:00Z"
  }
}
```

**Errors:**
- `400 Bad Request` - Input invalido o email ya registrado

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

| Campo | Tipo | Requerido |
|-------|------|-----------|
| `email` | string | Si |
| `password` | string | Si |

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
      "createdAt": "2026-01-02T12:00:00Z",
      "updatedAt": "2026-01-02T12:00:00Z"
    }
  }
}
```

**Errors:**
- `401 Unauthorized` - Credenciales invalidas

---

### Restaurants

Todos los endpoints de restaurantes requieren autenticacion (`Authorization: Bearer {token}`).

Un restaurante representa la marca/negocio. Las sucursales fisicas se gestionan a traves de Branches.

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
  "name": "La Pizzeria"
}
```

| Campo | Tipo | Requerido | Validacion |
|-------|------|-----------|------------|
| `name` | string | Si | Min 3, Max 100 caracteres |

**Response:** `201 Created`
```json
{
  "success": true,
  "message": "Restaurant created successfully",
  "data": {
    "id": "64a7f9abc12345678901234",
    "userId": "64a7f8b3c12345678901234",
    "name": "La Pizzeria",
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
      "createdAt": "2026-01-02T12:05:00Z",
      "updatedAt": "2026-01-02T12:05:00Z"
    }
  ]
}
```

---

#### Get Restaurant by ID

**GET** `/api/v1/restaurants/{id}`

Obtiene un restaurante especifico por su ID.

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
    "createdAt": "2026-01-02T12:05:00Z",
    "updatedAt": "2026-01-02T12:05:00Z"
  }
}
```

**Errors:**
- `404 Not Found` - Restaurante no encontrado
- `401 Unauthorized` - El usuario no es dueno de este restaurante

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
  "name": "La Gran Pizzeria"
}
```

| Campo | Tipo | Requerido | Validacion |
|-------|------|-----------|------------|
| `name` | string | No | Min 3, Max 100 caracteres |

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Restaurant updated successfully",
  "data": {
    "id": "64a7f9abc12345678901234",
    "userId": "64a7f8b3c12345678901234",
    "name": "La Gran Pizzeria",
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

### Branches

Las sucursales (branches) representan las ubicaciones fisicas de un restaurante. Cada sucursal tiene su propia direccion, descripcion y estado activo/inactivo. Las mesas se crean dentro de una sucursal.

Todos los endpoints de sucursales requieren autenticacion (`Authorization: Bearer {token}`).

#### Create Branch

**POST** `/api/v1/branches`

Crea una nueva sucursal para un restaurante.

**Headers:**
```
Authorization: Bearer {token}
Content-Type: application/json
```

**Request Body:**
```json
{
  "restaurantId": "64a7f9abc12345678901234",
  "name": "Sucursal Palermo",
  "address": "Av. Santa Fe 1234, CABA",
  "description": "Sucursal con terraza al aire libre"
}
```

| Campo | Tipo | Requerido | Validacion |
|-------|------|-----------|------------|
| `restaurantId` | string | Si | ObjectID valido |
| `name` | string | Si | Min 3, Max 100 caracteres |
| `address` | string | Si | Max 200 caracteres |
| `description` | string | No | Max 500 caracteres |

**Response:** `201 Created`
```json
{
  "success": true,
  "message": "Branch created successfully",
  "data": {
    "id": "64a7fabcd1234567890abcd",
    "restaurantId": "64a7f9abc12345678901234",
    "name": "Sucursal Palermo",
    "address": "Av. Santa Fe 1234, CABA",
    "description": "Sucursal con terraza al aire libre",
    "isActive": true,
    "createdAt": "2026-01-02T12:07:00Z",
    "updatedAt": "2026-01-02T12:07:00Z"
  }
}
```

**Errors:**
- `400 Bad Request` - Input invalido
- `401 Unauthorized` - El usuario no es dueno del restaurante
- `404 Not Found` - Restaurante no encontrado

---

#### Get Branch by ID

**GET** `/api/v1/branches/{id}`

Obtiene una sucursal especifica por su ID.

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Branch retrieved successfully",
  "data": {
    "id": "64a7fabcd1234567890abcd",
    "restaurantId": "64a7f9abc12345678901234",
    "name": "Sucursal Palermo",
    "address": "Av. Santa Fe 1234, CABA",
    "description": "Sucursal con terraza al aire libre",
    "isActive": true,
    "createdAt": "2026-01-02T12:07:00Z",
    "updatedAt": "2026-01-02T12:07:00Z"
  }
}
```

**Errors:**
- `404 Not Found` - Sucursal no encontrada
- `401 Unauthorized` - El usuario no tiene acceso a esta sucursal

---

#### List Branches by Restaurant

**GET** `/api/v1/branches/restaurant/{restaurantId}`

Lista todas las sucursales de un restaurante.

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Branches retrieved successfully",
  "data": [
    {
      "id": "64a7fabcd1234567890abcd",
      "restaurantId": "64a7f9abc12345678901234",
      "name": "Sucursal Palermo",
      "address": "Av. Santa Fe 1234, CABA",
      "description": "Sucursal con terraza al aire libre",
      "isActive": true,
      "createdAt": "2026-01-02T12:07:00Z",
      "updatedAt": "2026-01-02T12:07:00Z"
    }
  ]
}
```

**Errors:**
- `404 Not Found` - Restaurante no encontrado
- `401 Unauthorized` - El usuario no tiene acceso a este restaurante

---

#### Update Branch

**PUT** `/api/v1/branches/{id}`

Actualiza una sucursal existente.

**Headers:**
```
Authorization: Bearer {token}
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "Sucursal Palermo Soho",
  "address": "Av. Santa Fe 1234, CABA",
  "description": "Sucursal renovada con terraza",
  "isActive": true
}
```

| Campo | Tipo | Requerido | Validacion |
|-------|------|-----------|------------|
| `name` | string | No | Min 3, Max 100 caracteres |
| `address` | string | No | Max 200 caracteres |
| `description` | string | No | Max 500 caracteres |
| `isActive` | boolean | No | true/false |

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Branch updated successfully",
  "data": {
    "id": "64a7fabcd1234567890abcd",
    "restaurantId": "64a7f9abc12345678901234",
    "name": "Sucursal Palermo Soho",
    "address": "Av. Santa Fe 1234, CABA",
    "description": "Sucursal renovada con terraza",
    "isActive": true,
    "createdAt": "2026-01-02T12:07:00Z",
    "updatedAt": "2026-01-02T12:15:00Z"
  }
}
```

---

#### Delete Branch

**DELETE** `/api/v1/branches/{id}`

Elimina una sucursal.

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Branch deleted successfully"
}
```

---

### Tables

Las mesas pertenecen a una sucursal (branch). Cada mesa tiene un numero unico dentro de su sucursal y un QR code generado automaticamente.

Todos los endpoints de mesas requieren autenticacion (`Authorization: Bearer {token}`).

#### Create Table

**POST** `/api/v1/tables`

Crea una nueva mesa con QR code generado automaticamente.

**Headers:**
```
Authorization: Bearer {token}
Content-Type: application/json
```

**Request Body:**
```json
{
  "branchId": "64a7fabcd1234567890abcd",
  "number": 5
}
```

| Campo | Tipo | Requerido | Validacion |
|-------|------|-----------|------------|
| `branchId` | string | Si | ObjectID valido |
| `number` | int | Si | Minimo 1 |

**Response:** `201 Created`
```json
{
  "success": true,
  "message": "Table created successfully",
  "data": {
    "id": "64a7fabc12345678901234",
    "branchId": "64a7fabcd1234567890abcd",
    "number": 5,
    "qrCode": "http://localhost:5173/request?r=64a7f9abc12345678901234&b=64a7fabcd1234567890abcd&t=64a7fabc12345678901234&n=5&h=GUyQvt7LbzYbdaX9",
    "isActive": true,
    "createdAt": "2026-01-02T12:15:00Z",
    "updatedAt": "2026-01-02T12:15:00Z"
  }
}
```

**Nota:** El QR code contiene: `r` (restaurantId), `b` (branchId), `t` (tableId), `n` (table number), `h` (hash de seguridad SHA256).

**Errors:**
- `400 Bad Request` - El numero de mesa ya existe para esta sucursal
- `401 Unauthorized` - El usuario no es dueno de la sucursal
- `404 Not Found` - Sucursal no encontrada

---

#### Bulk Create Tables

**POST** `/api/v1/tables/bulk`

Crea multiples mesas automaticamente para una sucursal. Las mesas se crean secuencialmente comenzando desde el siguiente numero disponible.

**Headers:**
```
Authorization: Bearer {token}
Content-Type: application/json
```

**Request Body:**
```json
{
  "branchId": "64a7fabcd1234567890abcd",
  "count": 10
}
```

| Campo | Tipo | Requerido | Validacion |
|-------|------|-----------|------------|
| `branchId` | string | Si | ObjectID valido |
| `count` | int | Si | Min 1, Max 100 |

**Response:** `201 Created`
```json
{
  "success": true,
  "message": "Tables created successfully",
  "data": [
    {
      "id": "64a7fabc12345678901234",
      "branchId": "64a7fabcd1234567890abcd",
      "number": 1,
      "qrCode": "http://localhost:5173/request?r=...&b=...&t=...&n=1&h=...",
      "isActive": true,
      "createdAt": "2026-01-02T12:15:00Z",
      "updatedAt": "2026-01-02T12:15:00Z"
    },
    {
      "id": "64a7fabc12345678901235",
      "branchId": "64a7fabcd1234567890abcd",
      "number": 2,
      "qrCode": "http://localhost:5173/request?r=...&b=...&t=...&n=2&h=...",
      "isActive": true,
      "createdAt": "2026-01-02T12:15:01Z",
      "updatedAt": "2026-01-02T12:15:01Z"
    }
  ]
}
```

**Notas:**
- Si ya existen mesas, la numeracion continua desde el numero mas alto + 1
- Todas las mesas se crean como activas (`isActive: true`)
- Cada mesa recibe su QR code unico

**Errors:**
- `400 Bad Request` - Datos de entrada invalidos
- `401 Unauthorized` - El usuario no es dueno de la sucursal
- `404 Not Found` - Sucursal no encontrada

---

#### Get Table by ID

**GET** `/api/v1/tables/{id}`

Obtiene una mesa especifica por su ID.

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
    "branchId": "64a7fabcd1234567890abcd",
    "number": 5,
    "qrCode": "http://localhost:5173/request?r=...&b=...&t=...&n=5&h=...",
    "isActive": true,
    "createdAt": "2026-01-02T12:15:00Z",
    "updatedAt": "2026-01-02T12:15:00Z"
  }
}
```

**Errors:**
- `404 Not Found` - Mesa no encontrada
- `401 Unauthorized` - El usuario no tiene acceso

---

#### List Tables by Branch

**GET** `/api/v1/tables/branch/{branchId}`

Lista todas las mesas de una sucursal.

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
      "branchId": "64a7fabcd1234567890abcd",
      "number": 1,
      "qrCode": "http://localhost:5173/request?r=...&b=...&t=...&n=1&h=...",
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

Actualiza una mesa. Si se cambia el numero, se regenera automaticamente el QR code.

**Headers:**
```
Authorization: Bearer {token}
Content-Type: application/json
```

**Request Body:**
```json
{
  "number": 6,
  "isActive": true
}
```

| Campo | Tipo | Requerido | Validacion |
|-------|------|-----------|------------|
| `number` | int | No | Minimo 1 |
| `isActive` | boolean | No | true/false |

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Table updated successfully",
  "data": {
    "id": "64a7fabc12345678901234",
    "branchId": "64a7fabcd1234567890abcd",
    "number": 6,
    "qrCode": "http://localhost:5173/request?r=...&b=...&t=...&n=6&h=...",
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

El sistema de solicitudes permite a los clientes pedir la cuenta escaneando el QR de la mesa. Las solicitudes incluyen informacion del restaurante, la sucursal y la mesa.

#### Create Request (Public)

**POST** `/api/v1/public/request-account`

**Endpoint publico** - No requiere autenticacion. Usado por clientes al escanear el QR.

**Request Body:**
```json
{
  "restaurantId": "64a7f9abc12345678901234",
  "branchId": "64a7fabcd1234567890abcd",
  "tableId": "64a7fabc12345678901234",
  "tableNumber": 5,
  "hash": "GUyQvt7LbzYbdaX9"
}
```

| Campo | Tipo | Requerido | Validacion |
|-------|------|-----------|------------|
| `restaurantId` | string | Si | ObjectID valido |
| `branchId` | string | Si | ObjectID valido |
| `tableId` | string | Si | ObjectID valido |
| `tableNumber` | int | Si | Minimo 1 |
| `hash` | string | Si | Hash de seguridad del QR |

**Validaciones que se realizan:**
1. Se valida el hash del QR code (SHA256)
2. Se verifica que el restaurante exista
3. Se verifica que la sucursal exista y pertenezca al restaurante
4. Se verifica que la sucursal este activa
5. Se verifica que la mesa exista y pertenezca a la sucursal
6. Se verifica que la mesa este activa

**Response:** `201 Created`
```json
{
  "success": true,
  "message": "Request created successfully",
  "data": {
    "id": "64a7fbcd12345678901234",
    "restaurantId": "64a7f9abc12345678901234",
    "branchId": "64a7fabcd1234567890abcd",
    "tableId": "64a7fabc12345678901234",
    "tableNumber": 5,
    "status": "pending",
    "createdAt": "2026-01-02T12:25:00Z",
    "updatedAt": "2026-01-02T12:25:00Z"
  }
}
```

**Nota:** Al crear un request, se envia automaticamente una notificacion WebSocket al restaurante.

**Errors:**
- `400 Bad Request` - QR invalido, mesa/sucursal inactiva, o datos invalidos
- `404 Not Found` - Restaurante, sucursal o mesa no encontrada

---

#### Get Request by ID

**GET** `/api/v1/requests/{id}`

Obtiene una solicitud especifica.

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
    "branchId": "64a7fabcd1234567890abcd",
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

Lista todas las solicitudes de un restaurante (todos los estados), ordenadas por fecha de creacion descendente.

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
      "branchId": "64a7fabcd1234567890abcd",
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

#### List Pending Requests

**GET** `/api/v1/requests/restaurant/{restaurantId}/pending`

Lista solo las solicitudes pendientes de un restaurante, ordenadas por fecha de creacion descendente.

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
      "branchId": "64a7fabcd1234567890abcd",
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
  "status": "attended"
}
```

**Valores permitidos para `status`:**

| Status | Descripcion |
|--------|-------------|
| `pending` | Pendiente (estado inicial) |
| `attended` | Atendida/Procesada |
| `cancelled` | Cancelada |

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Request status updated successfully",
  "data": {
    "id": "64a7fbcd12345678901234",
    "restaurantId": "64a7f9abc12345678901234",
    "branchId": "64a7fabcd1234567890abcd",
    "tableId": "64a7fabc12345678901234",
    "tableNumber": 5,
    "status": "attended",
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

Establece una conexion WebSocket para recibir notificaciones en tiempo real de nuevas solicitudes de cuenta.

**Parametros:**

| Parametro | Ubicacion | Requerido | Descripcion |
|-----------|-----------|-----------|-------------|
| `restaurantId` | path | Si | ID del restaurante |
| `token` | query | Si | JWT token del usuario |

**Ejemplo de conexion (JavaScript):**
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
  //   "branchId": "64a7fabcd1234567890abcd",
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
- El WebSocket envia mensajes JSON cuando se crea una nueva solicitud
- La conexion es especifica por restaurante (recibe solicitudes de todas las sucursales)
- Solo los usuarios autenticados pueden conectarse (token validado en el handler)
- El hub de WebSocket mantiene las conexiones activas y limpia automaticamente las desconectadas
- El servidor solo envia mensajes; no espera recibir mensajes del cliente

---

## Ejemplos de Uso

### Flujo completo: Registro > Restaurante > Sucursal > Mesa > Solicitud

```bash
# 1. Registrar usuario
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "restaurant@example.com",
    "password": "password123"
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
    "name": "La Pizzeria"
  }')

RESTAURANT_ID=$(echo $RESTAURANT | jq -r '.data.id')
echo "Restaurant ID: $RESTAURANT_ID"

# 4. Crear sucursal
BRANCH=$(curl -s -X POST http://localhost:8080/api/v1/branches \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{
    \"restaurantId\": \"$RESTAURANT_ID\",
    \"name\": \"Sucursal Palermo\",
    \"address\": \"Av. Corrientes 1234, CABA\",
    \"description\": \"Pizzeria con horno a lena\"
  }")

BRANCH_ID=$(echo $BRANCH | jq -r '.data.id')
echo "Branch ID: $BRANCH_ID"

# 5. Crear mesas en bulk (10 mesas)
curl -s -X POST http://localhost:8080/api/v1/tables/bulk \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{
    \"branchId\": \"$BRANCH_ID\",
    \"count\": 10
  }" | jq '.data[] | {number, qrCode}'

# 6. O crear una mesa individual
TABLE=$(curl -s -X POST http://localhost:8080/api/v1/tables \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{
    \"branchId\": \"$BRANCH_ID\",
    \"number\": 99
  }")

TABLE_ID=$(echo $TABLE | jq -r '.data.id')
QR_CODE=$(echo $TABLE | jq -r '.data.qrCode')
echo "QR Code: $QR_CODE"

# 7. Simular cliente escaneando QR y solicitando cuenta
# Extraer parametros del QR code
QR_HASH=$(echo "$QR_CODE" | grep -o 'h=[^&]*' | cut -d'=' -f2)

curl -X POST http://localhost:8080/api/v1/public/request-account \
  -H 'Content-Type: application/json' \
  -d "{
    \"restaurantId\": \"$RESTAURANT_ID\",
    \"branchId\": \"$BRANCH_ID\",
    \"tableId\": \"$TABLE_ID\",
    \"tableNumber\": 99,
    \"hash\": \"$QR_HASH\"
  }"

# 8. Obtener solicitudes pendientes
curl -s -X GET "http://localhost:8080/api/v1/requests/restaurant/$RESTAURANT_ID/pending" \
  -H "Authorization: Bearer $TOKEN" | jq
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

### Build para produccion

```bash
go build -o bin/server cmd/api/main.go && ./bin/server
```

---

## Seguridad

- Contrasenas hasheadas con bcrypt (cost factor: 10)
- Tokens JWT firmados con HS256 y expiracion de 24 horas
- Validacion de QR codes con hash SHA256
- Middleware de autenticacion en todas las rutas protegidas
- Validacion de ownership en todos los endpoints (usuario solo accede a sus recursos)
- Validacion de relaciones jerarquicas (branch pertenece a restaurant, table pertenece a branch)
- CORS configurable con multiples origenes
- Input validation en todos los endpoints (email, password, lengths, ObjectIDs)
- WebSocket con autenticacion via token en query parameter

---

## Codigos de Error HTTP

| Codigo | Descripcion |
|--------|-------------|
| `200` | OK - Solicitud exitosa |
| `201` | Created - Recurso creado exitosamente |
| `400` | Bad Request - Datos de entrada invalidos |
| `401` | Unauthorized - No autenticado, token invalido o sin permisos |
| `404` | Not Found - Recurso no encontrado |
| `500` | Internal Server Error - Error del servidor |

---

## Colecciones MongoDB

| Coleccion | Descripcion |
|-----------|-------------|
| `users` | Usuarios registrados (email + password hasheado) |
| `restaurants` | Restaurantes (marca/negocio, vinculado a user) |
| `branches` | Sucursales fisicas (vinculado a restaurant) |
| `tables` | Mesas con QR codes (vinculado a branch) |
| `requests` | Solicitudes de cuenta (vinculado a restaurant, branch y table) |

---

## License

MIT License - ver archivo LICENSE para mas detalles.

---

## Contacto

Para preguntas o soporte, contactar a: [j.s.calvinio@gmail.com]
